// Copyright 2019 The Energi Core Authors
// This file is part of the Energi Core library.
//
// The Energi Core library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Energi Core library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Energi Core library. If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"io"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"

	"github.com/shengdoushi/base58"
	"golang.org/x/crypto/ripemd160"

	energi_abi "energi.world/core/gen3/energi/abi"
	energi_params "energi.world/core/gen3/energi/params"
)

const (
	base54PrivateKeyLen int    = 52
	privateKeyLen       int    = 32
	migrationGas        uint64 = 100000
)

type MigrationAPI struct {
	backend Backend
}

func NewMigrationAPI(b Backend) *MigrationAPI {
	return &MigrationAPI{b}
}

type Gen2Coin struct {
	ItemID   uint64
	RawOwner common.Address
	Owner    string
	Amount   *big.Int
}

type Gen2Key struct {
	RawOwner common.Address
	Key      *ecdsa.PrivateKey
}

func (m *MigrationAPI) ListGen2Coins() (coins []Gen2Coin) {
	log.Info("Preparing a coin list")

	mgrt_contract, err := energi_abi.NewGen2MigrationCaller(
		energi_params.Energi_MigrationContract, m.backend.(bind.ContractCaller))
	if err != nil {
		log.Error("Failed to create contract face", "err", err)
		return []Gen2Coin{}
	}

	call_opts := &bind.CallOpts{}
	bigItems, err := mgrt_contract.ItemCount(call_opts)
	if err != nil {
		log.Error("Failed to get coin count", "err", err)
		return []Gen2Coin{}
	}

	items := bigItems.Int64()
	coins = make([]Gen2Coin, 0, items)

	prefix := byte(33)
	if m.backend.ChainConfig().ChainID.Int64() == 49797 {
		prefix = byte(127)
	}

	for i := int64(0); i < items; i++ {
		res, err := mgrt_contract.Coins(call_opts, big.NewInt(i))
		if err != nil {
			log.Error("Failed to get coin info", "err", err)
			return []Gen2Coin{}
		}

		owner := make([]byte, 25)
		owner[0] = prefix
		copy(owner[1:], res.Owner[:])
		ownerhash := sha256.Sum256(owner[:21])
		ownerhash = sha256.Sum256(ownerhash[:])
		copy(owner[21:], ownerhash[:4])

		coins = append(coins, Gen2Coin{
			ItemID:   uint64(i),
			RawOwner: common.BytesToAddress(res.Owner[:]),
			Owner:    base58.Encode(owner, base58.BitcoinAlphabet),
			Amount:   res.Amount,
		})
	}

	return
}

func (m *MigrationAPI) SearchGen2Coins(
	owners []string,
	include_empty bool,
) (coins []Gen2Coin) {
	rawOwners := make([]common.Address, len(owners))
	for i, o := range owners {
		ro, err := base58.Decode(o, base58.BitcoinAlphabet)
		if err != nil || len(ro) < 20 {
			log.Error("Failed to decode owner", "err", err, "owner", o)
			continue
		}
		rawOwners[i] = common.BytesToAddress(ro[1 : len(ro)-4])
	}
	return m.searchGen2Coins(rawOwners, m.ListGen2Coins(), include_empty)
}

func (m *MigrationAPI) SearchRawGen2Coins(
	rawOwners []common.Address,
	include_empty bool,
) (coins []Gen2Coin) {
	return m.searchGen2Coins(rawOwners, m.ListGen2Coins(), include_empty)
}

func (m *MigrationAPI) searchGen2Coins(
	owners []common.Address,
	all_coins []Gen2Coin,
	include_empty bool,
) (coins []Gen2Coin) {
	coins = make([]Gen2Coin, 0, len(owners))

	owners_map := make(map[common.Address]bool)
	for _, o := range owners {
		owners_map[o] = true
	}

	for _, c := range all_coins {
		if _, ok := owners_map[c.RawOwner]; ok {
			if include_empty || c.Amount.Cmp(common.Big0) > 0 {
				coins = append(coins, c)
			}
		}
	}

	return coins
}

func (m *MigrationAPI) loadGen2Dump(file string) (keys []Gen2Key, err error) {
	f, err := os.Open(file)
	if err != nil {
		log.Error("Failed to open dump file", "err", err)
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Error("Failed to stat file", "err", err)
		return nil, err
	}

	buf := make([]byte, fi.Size())
	len, err := io.ReadFull(f, buf)
	if err != nil {
		log.Error("Failed to read file", "err", err)
		return nil, err
	}

	return m.parseGen2Dump(string(buf[:len])), nil
}

func (m *MigrationAPI) parseGen2Dump(data string) (keys []Gen2Key) {
	lines := strings.Split(data, "\n")
	keys = make([]Gen2Key, 0, len(lines))

	for i, l := range lines {
		lp := strings.Split(l, " ")
		if len(lp) < 3 || lp[0] == "#" {
			continue
		}

		key, err := m.parseGen2Key(lp[0])
		if err != nil {
			log.Error("Failed to parse key", "err", err, "line", i)
			continue
		}

		keys = append(keys, *key)
	}

	return
}

func (m *MigrationAPI) parseGen2Key(tkey string) (*Gen2Key, error) {
	if len(tkey) != base54PrivateKeyLen {
		return nil, errors.New("Invalid private key length")
	}

	rkey, err := base58.Decode(tkey, base58.BitcoinAlphabet)
	if err != nil {
		return nil, err
	}

	// There is prefix + key + [magic +] checksum
	key_obj, err := crypto.ToECDSA(rkey[1 : 1+privateKeyLen])
	if err != nil {
		return nil, err
	}

	var owner common.Address

	basehash := sha256.Sum256(crypto.CompressPubkey(&key_obj.PublicKey))
	ripemd := ripemd160.New()
	ripemd.Write(basehash[:])
	owner.SetBytes(ripemd.Sum(nil))

	return &Gen2Key{
		RawOwner: owner,
		Key:      key_obj,
	}, nil
}

func (m *MigrationAPI) ClaimGen2CoinsDirect(
	password string,
	dst common.Address,
	tkey string,
) error {
	key, err := m.parseGen2Key(tkey)
	if err != nil {
		log.Error("Failed to parse key", "err", err)
		return err
	}

	coins := m.SearchRawGen2Coins([]common.Address{key.RawOwner}, false)

	if len(coins) != 1 {
		log.Error("Unable to find coins")
		return errors.New("No coins found")
	}

	err = m.claimGen2Coins(password, dst, &coins[0], key)
	if err != nil {
		log.Error("Failed to claim", "err", err)
		return err
	}

	return nil
}

func (m *MigrationAPI) ClaimGen2CoinsCombined(
	password string,
	dst common.Address,
	file string,
) error {
	keys, err := m.loadGen2Dump(file)
	if err != nil {
		return err
	}

	raw_owners := make([]common.Address, len(keys))
	owner2key := make(map[common.Address]*Gen2Key, len(keys))
	for i, k := range keys {
		raw_owners[i] = k.RawOwner
		owner2key[k.RawOwner] = &keys[i]
	}

	coins := m.SearchRawGen2Coins(raw_owners, false)

	for _, c := range coins {
		err = m.claimGen2Coins(password, dst, &c, owner2key[c.RawOwner])
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MigrationAPI) ClaimGen2CoinsImport(password string, file string) error {
	keys, err := m.loadGen2Dump(file)
	if err != nil {
		return err
	}

	raw_owners := make([]common.Address, len(keys))
	owner2key := make(map[common.Address]*Gen2Key, len(keys))
	for i, k := range keys {
		raw_owners[i] = k.RawOwner
		owner2key[k.RawOwner] = &keys[i]
	}

	coins := m.SearchRawGen2Coins(raw_owners, false)
	am := m.backend.AccountManager()
	ks := am.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	for _, c := range coins {
		key := owner2key[c.RawOwner]
		dst := crypto.PubkeyToAddress(key.Key.PublicKey)

		//----
		sink := make(chan accounts.WalletEvent)
		evtsub := am.Subscribe(sink)
		defer evtsub.Unsubscribe()

		if _, err := ks.ImportECDSA(key.Key, password); err != nil {
			log.Warn("Failed to import private key", "err", err)
			// Most likely key exists
		} else {
			select {
			case <-sink:
			}
		}

		evtsub.Unsubscribe()
		//----

		err = m.claimGen2Coins(password, dst, &c, key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MigrationAPI) claimGen2Coins(
	password string,
	dst common.Address,
	coin *Gen2Coin,
	key *Gen2Key,
) error {
	account := accounts.Account{Address: dst}
	wallet, err := m.backend.AccountManager().Find(accounts.Account{Address: dst})
	if err != nil {
		return err
	}

	mgrt_contract_obj, err := energi_abi.NewGen2Migration(
		energi_params.Energi_MigrationContract, m.backend.(bind.ContractBackend))
	if err != nil {
		return err
	}

	mgrt_contract := energi_abi.Gen2MigrationSession{
		Contract: mgrt_contract_obj,
		CallOpts: bind.CallOpts{
			From: dst,
		},
		TransactOpts: bind.TransactOpts{
			From: dst,
			Signer: func(
				signer types.Signer,
				addr common.Address,
				tx *types.Transaction,
			) (*types.Transaction, error) {
				return wallet.SignTxWithPassphrase(
					account, password, tx, m.backend.ChainConfig().ChainID)
			},
			Value:    common.Big0,
			GasPrice: common.Big0,
			GasLimit: migrationGas,
		},
	}

	hts, err := mgrt_contract.HashToSign(dst)
	if err != nil {
		return err
	}

	sig, err := crypto.Sign(hts[:], key.Key)
	if err != nil {
		return err
	}

	if len(sig) != 65 {
		return errors.New("Wrong signature size")
	}

	item := new(big.Int).SetUint64(coin.ItemID)
	r := [32]byte{}
	copy(r[:], sig[:32])
	s := [32]byte{}
	copy(s[:], sig[32:64])
	v := uint8(sig[64])

	amt, err := mgrt_contract.VerifyClaim(item, dst, v, r, s)
	if err != nil {
		return err
	}

	if amt.Cmp(common.Big0) == 0 {
		log.Warn("Already claimed", "coins", coin.Owner)
		return nil
	}

	tx, err := mgrt_contract.Claim(item, dst, v, r, s)
	log.Info("Sent migration transaction", "tx", tx.Hash(), "coins", coin.Owner)

	return err
}