// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package bind

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

var bindTests = []struct {
	name     string
	contract string
	bytecode string
	abi      string
	imports  string
	tester   string
}{
	// Test that the binding is available in combined and separate forms too
	{
		`Empty`,
		`contract NilContract {}`,
		`606060405260068060106000396000f3606060405200`,
		`[]`,
		`"github.com/ethereum/go-ethereum/common"`,
		`
			if b, err := NewEmpty(common.Address{}, nil); b == nil || err != nil {
				t.Fatalf("combined binding (%v) nil or error (%v) not nil", b, nil)
			}
			if b, err := NewEmptyCaller(common.Address{}, nil); b == nil || err != nil {
				t.Fatalf("caller binding (%v) nil or error (%v) not nil", b, nil)
			}
			if b, err := NewEmptyTransactor(common.Address{}, nil); b == nil || err != nil {
				t.Fatalf("transactor binding (%v) nil or error (%v) not nil", b, nil)
			}
		`,
	},
	// Test that all the official sample contracts bind correctly
	{
		`Token`,
		`https://ethereum.org/token`,
		`60606040526040516107fd3803806107fd83398101604052805160805160a05160c051929391820192909101600160a060020a0333166000908152600360209081526040822086905581548551838052601f6002600019610100600186161502019093169290920482018390047f290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e56390810193919290918801908390106100e857805160ff19168380011785555b506101189291505b8082111561017157600081556001016100b4565b50506002805460ff19168317905550505050610658806101a56000396000f35b828001600101855582156100ac579182015b828111156100ac5782518260005055916020019190600101906100fa565b50508060016000509080519060200190828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061017557805160ff19168380011785555b506100c89291506100b4565b5090565b82800160010185558215610165579182015b8281111561016557825182600050559160200191906001019061018756606060405236156100775760e060020a600035046306fdde03811461007f57806323b872dd146100dc578063313ce5671461010e57806370a082311461011a57806395d89b4114610132578063a9059cbb1461018e578063cae9ca51146101bd578063dc3080f21461031c578063dd62ed3e14610341575b610365610002565b61036760008054602060026001831615610100026000190190921691909104601f810182900490910260809081016040526060828152929190828280156104eb5780601f106104c0576101008083540402835291602001916104eb565b6103d5600435602435604435600160a060020a038316600090815260036020526040812054829010156104f357610002565b6103e760025460ff1681565b6103d560043560036020526000908152604090205481565b610367600180546020600282841615610100026000190190921691909104601f810182900490910260809081016040526060828152929190828280156104eb5780601f106104c0576101008083540402835291602001916104eb565b610365600435602435600160a060020a033316600090815260036020526040902054819010156103f157610002565b60806020604435600481810135601f8101849004909302840160405260608381526103d5948235946024803595606494939101919081908382808284375094965050505050505060006000836004600050600033600160a060020a03168152602001908152602001600020600050600087600160a060020a031681526020019081526020016000206000508190555084905080600160a060020a0316638f4ffcb1338630876040518560e060020a0281526004018085600160a060020a0316815260200184815260200183600160a060020a03168152602001806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600f02600301f150905090810190601f1680156102f25780820380516001836020036101000a031916815260200191505b50955050505050506000604051808303816000876161da5a03f11561000257505050509392505050565b6005602090815260043560009081526040808220909252602435815220546103d59081565b60046020818152903560009081526040808220909252602435815220546103d59081565b005b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600f02600301f150905090810190601f1680156103c75780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b60408051918252519081900360200190f35b6060908152602090f35b600160a060020a03821660009081526040902054808201101561041357610002565b806003600050600033600160a060020a03168152602001908152602001600020600082828250540392505081905550806003600050600084600160a060020a0316815260200190815260200160002060008282825054019250508190555081600160a060020a031633600160a060020a03167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef836040518082815260200191505060405180910390a35050565b820191906000526020600020905b8154815290600101906020018083116104ce57829003601f168201915b505050505081565b600160a060020a03831681526040812054808301101561051257610002565b600160a060020a0380851680835260046020908152604080852033949094168086529382528085205492855260058252808520938552929052908220548301111561055c57610002565b816003600050600086600160a060020a03168152602001908152602001600020600082828250540392505081905550816003600050600085600160a060020a03168152602001908152602001600020600082828250540192505081905550816005600050600086600160a060020a03168152602001908152602001600020600050600033600160a060020a0316815260200190815260200160002060008282825054019250508190555082600160a060020a031633600160a060020a03167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040518082815260200191505060405180910390a3939250505056`,
		`[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"success","type":"bool"}],"type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[],"type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"},{"name":"_extraData","type":"bytes"}],"name":"approveAndCall","outputs":[{"name":"success","type":"bool"}],"type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"},{"name":"","type":"address"}],"name":"spentAllowance","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"},{"name":"","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"inputs":[{"name":"initialSupply","type":"uint256"},{"name":"tokenName","type":"string"},{"name":"decimalUnits","type":"uint8"},{"name":"tokenSymbol","type":"string"}],"type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`,
		`"github.com/ethereum/go-ethereum/common"`,
		`
			if b, err := NewToken(common.Address{}, nil); b == nil || err != nil {
				t.Fatalf("binding (%v) nil or error (%v) not nil", b, nil)
			}
		`,
	},
	{
		`Crowdsale`,
		`https://ethereum.org/crowdsale`,
		`606060408190526007805460ff1916905560a0806105a883396101006040529051608051915160c05160e05160008054600160a060020a03199081169095178155670de0b6b3a7640000958602600155603c9093024201600355930260045560058054909216909217905561052f90819061007990396000f36060604052361561006c5760e060020a600035046301cb3b20811461008257806329dcb0cf1461014457806338af3eed1461014d5780636e66f6e91461015f5780637a3a0e84146101715780637b3e5e7b1461017a578063a035b1fe14610183578063dc0d3dff1461018c575b61020060075460009060ff161561032357610002565b61020060035460009042106103205760025460015490106103cb576002548154600160a060020a0316908290606082818181858883f150915460025460408051600160a060020a039390931683526020830191909152818101869052517fe842aea7a5f1b01049d752008c53c52890b1a6daf660cf39e8eec506112bbdf6945090819003909201919050a15b60405160008054600160a060020a039081169230909116319082818181858883f150506007805460ff1916600117905550505050565b6103a160035481565b6103ab600054600160a060020a031681565b6103ab600554600160a060020a031681565b6103a160015481565b6103a160025481565b6103a160045481565b6103be60043560068054829081101561000257506000526002027ff652222313e28459528d920b65115c16c04f3efc82aaedc97be59f3f377c0d3f8101547ff652222313e28459528d920b65115c16c04f3efc82aaedc97be59f3f377c0d409190910154600160a060020a03919091169082565b005b505050815481101561000257906000526020600020906002020160005060008201518160000160006101000a815481600160a060020a030219169083021790555060208201518160010160005055905050806002600082828250540192505081905550600560009054906101000a9004600160a060020a0316600160a060020a031663a9059cbb3360046000505484046040518360e060020a0281526004018083600160a060020a03168152602001828152602001925050506000604051808303816000876161da5a03f11561000257505060408051600160a060020a03331681526020810184905260018183015290517fe842aea7a5f1b01049d752008c53c52890b1a6daf660cf39e8eec506112bbdf692509081900360600190a15b50565b5060a0604052336060908152346080819052600680546001810180835592939282908280158290116102025760020281600202836000526020600020918201910161020291905b8082111561039d57805473ffffffffffffffffffffffffffffffffffffffff19168155600060019190910190815561036a565b5090565b6060908152602090f35b600160a060020a03166060908152602090f35b6060918252608052604090f35b5b60065481101561010e576006805482908110156100025760009182526002027ff652222313e28459528d920b65115c16c04f3efc82aaedc97be59f3f377c0d3f0190600680549254600160a060020a0316928490811015610002576002027ff652222313e28459528d920b65115c16c04f3efc82aaedc97be59f3f377c0d40015460405190915082818181858883f19350505050507fe842aea7a5f1b01049d752008c53c52890b1a6daf660cf39e8eec506112bbdf660066000508281548110156100025760008290526002027ff652222313e28459528d920b65115c16c04f3efc82aaedc97be59f3f377c0d3f01548154600160a060020a039190911691908490811015610002576002027ff652222313e28459528d920b65115c16c04f3efc82aaedc97be59f3f377c0d40015460408051600160a060020a0394909416845260208401919091526000838201525191829003606001919050a16001016103cc56`,
		`[{"constant":false,"inputs":[],"name":"checkGoalReached","outputs":[],"type":"function"},{"constant":true,"inputs":[],"name":"deadline","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":true,"inputs":[],"name":"beneficiary","outputs":[{"name":"","type":"address"}],"type":"function"},{"constant":true,"inputs":[],"name":"tokenReward","outputs":[{"name":"","type":"address"}],"type":"function"},{"constant":true,"inputs":[],"name":"fundingGoal","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":true,"inputs":[],"name":"amountRaised","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":true,"inputs":[],"name":"price","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":true,"inputs":[{"name":"","type":"uint256"}],"name":"funders","outputs":[{"name":"addr","type":"address"},{"name":"amount","type":"uint256"}],"type":"function"},{"inputs":[{"name":"ifSuccessfulSendTo","type":"address"},{"name":"fundingGoalInEthers","type":"uint256"},{"name":"durationInMinutes","type":"uint256"},{"name":"etherCostOfEachToken","type":"uint256"},{"name":"addressOfTokenUsedAsReward","type":"address"}],"type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"name":"backer","type":"address"},{"indexed":false,"name":"amount","type":"uint256"},{"indexed":false,"name":"isContribution","type":"bool"}],"name":"FundTransfer","type":"event"}]`,
		`"github.com/ethereum/go-ethereum/common"`,
		`
			if b, err := NewCrowdsale(common.Address{}, nil); b == nil || err != nil {
				t.Fatalf("binding (%v) nil or error (%v) not nil", b, nil)
			}
		`,
	},
	// Test that named and anonymous inputs are handled correctly
	{
		`InputChecker`, ``, ``,
		`
			[
				{"type":"function","name":"noInput","constant":true,"inputs":[],"outputs":[]},
				{"type":"function","name":"namedInput","constant":true,"inputs":[{"name":"str","type":"string"}],"outputs":[]},
				{"type":"function","name":"anonInput","constant":true,"inputs":[{"name":"","type":"string"}],"outputs":[]},
				{"type":"function","name":"namedInputs","constant":true,"inputs":[{"name":"str1","type":"string"},{"name":"str2","type":"string"}],"outputs":[]},
				{"type":"function","name":"anonInputs","constant":true,"inputs":[{"name":"","type":"string"},{"name":"","type":"string"}],"outputs":[]},
				{"type":"function","name":"mixedInputs","constant":true,"inputs":[{"name":"","type":"string"},{"name":"str","type":"string"}],"outputs":[]}
			]
		`,
		`
			"fmt"

			"github.com/ethereum/go-ethereum/common"
		`,
		`if b, err := NewInputChecker(common.Address{}, nil); b == nil || err != nil {
			 t.Fatalf("binding (%v) nil or error (%v) not nil", b, nil)
		 } else if false { // Don't run, just compile and test types
			 var err error

			 err = b.NoInput(nil)
			 err = b.NamedInput(nil, "")
			 err = b.AnonInput(nil, "")
			 err = b.NamedInputs(nil, "", "")
			 err = b.AnonInputs(nil, "", "")
			 err = b.MixedInputs(nil, "", "")

			 fmt.Println(err)
		 }`,
	},
	// Test that named and anonymous outputs are handled correctly
	{
		`OutputChecker`, ``, ``,
		`
			[
				{"type":"function","name":"noOutput","constant":true,"inputs":[],"outputs":[]},
				{"type":"function","name":"namedOutput","constant":true,"inputs":[],"outputs":[{"name":"str","type":"string"}]},
				{"type":"function","name":"anonOutput","constant":true,"inputs":[],"outputs":[{"name":"","type":"string"}]},
				{"type":"function","name":"namedOutputs","constant":true,"inputs":[],"outputs":[{"name":"str1","type":"string"},{"name":"str2","type":"string"}]},
				{"type":"function","name":"collidingOutputs","constant":true,"inputs":[],"outputs":[{"name":"str","type":"string"},{"name":"Str","type":"string"}]},
				{"type":"function","name":"anonOutputs","constant":true,"inputs":[],"outputs":[{"name":"","type":"string"},{"name":"","type":"string"}]},
				{"type":"function","name":"mixedOutputs","constant":true,"inputs":[],"outputs":[{"name":"","type":"string"},{"name":"str","type":"string"}]}
			]
		`,
		`
			"fmt"

			"github.com/ethereum/go-ethereum/common"
		`,
		`if b, err := NewOutputChecker(common.Address{}, nil); b == nil || err != nil {
			 t.Fatalf("binding (%v) nil or error (%v) not nil", b, nil)
		 } else if false { // Don't run, just compile and test types
			 var str1, str2 string
			 var err error

			 err              = b.NoOutput(nil)
			 str1, err        = b.NamedOutput(nil)
			 str1, err        = b.AnonOutput(nil)
			 res, _          := b.NamedOutputs(nil)
			 str1, str2, err  = b.CollidingOutputs(nil)
			 str1, str2, err  = b.AnonOutputs(nil)
			 str1, str2, err  = b.MixedOutputs(nil)

			 fmt.Println(str1, str2, res.Str1, res.Str2, err)
		 }`,
	},
	// Tests that named, anonymous and indexed events are handled correctly
	{
		`EventChecker`, ``, ``,
		`
			[
				{"type":"event","name":"empty","inputs":[]},
				{"type":"event","name":"indexed","inputs":[{"name":"addr","type":"address","indexed":true},{"name":"num","type":"int256","indexed":true}]},
				{"type":"event","name":"mixed","inputs":[{"name":"addr","type":"address","indexed":true},{"name":"num","type":"int256"}]},
				{"type":"event","name":"anonymous","anonymous":true,"inputs":[]},
				{"type":"event","name":"dynamic","inputs":[{"name":"idxStr","type":"string","indexed":true},{"name":"idxDat","type":"bytes","indexed":true},{"name":"str","type":"string"},{"name":"dat","type":"bytes"}]}
			]
		`,
		`
			"fmt"
			"math/big"
			"reflect"

			"github.com/ethereum/go-ethereum/common"
		`,
		`if e, err := NewEventChecker(common.Address{}, nil); e == nil || err != nil {
			 t.Fatalf("binding (%v) nil or error (%v) not nil", e, nil)
		 } else if false { // Don't run, just compile and test types
			 var (
				 err  error
			   res  bool
				 str  string
				 dat  []byte
				 hash common.Hash
			 )
			 _, err = e.FilterEmpty(nil)
			 _, err = e.FilterIndexed(nil, []common.Address{}, []*big.Int{})

			 mit, err := e.FilterMixed(nil, []common.Address{})

			 res = mit.Next()  // Make sure the iterator has a Next method
			 err = mit.Error() // Make sure the iterator has an Error method
			 err = mit.Close() // Make sure the iterator has a Close method

			 fmt.Println(mit.Event.Raw.BlockHash) // Make sure the raw log is contained within the results
			 fmt.Println(mit.Event.Num)           // Make sure the unpacked non-indexed fields are present
			 fmt.Println(mit.Event.Addr)          // Make sure the reconstructed indexed fields are present

			 dit, err := e.FilterDynamic(nil, []string{}, [][]byte{})

			 str  = dit.Event.Str    // Make sure non-indexed strings retain their type
			 dat  = dit.Event.Dat    // Make sure non-indexed bytes retain their type
			 hash = dit.Event.IdxStr // Make sure indexed strings turn into hashes
			 hash = dit.Event.IdxDat // Make sure indexed bytes turn into hashes

			 sink := make(chan *EventCheckerMixed)
			 sub, err := e.WatchMixed(nil, sink, []common.Address{})
			 defer sub.Unsubscribe()

			 event := <-sink
			 fmt.Println(event.Raw.BlockHash) // Make sure the raw log is contained within the results
			 fmt.Println(event.Num)           // Make sure the unpacked non-indexed fields are present
			 fmt.Println(event.Addr)          // Make sure the reconstructed indexed fields are present

			 fmt.Println(res, str, dat, hash, err)
		 }
		 // Run a tiny reflection test to ensure disallowed methods don't appear
		 if _, ok := reflect.TypeOf(&EventChecker{}).MethodByName("FilterAnonymous"); ok {
		 	t.Errorf("binding has disallowed method (FilterAnonymous)")
		 }`,
	},
	// Test that contract interactions (deploy, transact and call) generate working code
	{
		`Interactor`,
		`
			contract Interactor {
				string public deployString;
				string public transactString;

				function Interactor(string str) {
				  deployString = str;
				}

				function transact(string str) {
				  transactString = str;
				}
			}
		`,
		`6060604052604051610328380380610328833981016040528051018060006000509080519060200190828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f10608d57805160ff19168380011785555b50607c9291505b8082111560ba57838155600101606b565b50505061026a806100be6000396000f35b828001600101855582156064579182015b828111156064578251826000505591602001919060010190609e565b509056606060405260e060020a60003504630d86a0e181146100315780636874e8091461008d578063d736c513146100ea575b005b610190600180546020600282841615610100026000190190921691909104601f810182900490910260809081016040526060828152929190828280156102295780601f106101fe57610100808354040283529160200191610229565b61019060008054602060026001831615610100026000190190921691909104601f810182900490910260809081016040526060828152929190828280156102295780601f106101fe57610100808354040283529160200191610229565b60206004803580820135601f81018490049093026080908101604052606084815261002f946024939192918401918190838280828437509496505050505050508060016000509080519060200190828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061023157805160ff19168380011785555b506102619291505b808211156102665760008155830161017d565b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600f02600301f150905090810190601f1680156101f05780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b820191906000526020600020905b81548152906001019060200180831161020c57829003601f168201915b505050505081565b82800160010185558215610175579182015b82811115610175578251826000505591602001919060010190610243565b505050565b509056`,
		`[{"constant":true,"inputs":[],"name":"transactString","outputs":[{"name":"","type":"string"}],"type":"function"},{"constant":true,"inputs":[],"name":"deployString","outputs":[{"name":"","type":"string"}],"type":"function"},{"constant":false,"inputs":[{"name":"str","type":"string"}],"name":"transact","outputs":[],"type":"function"},{"inputs":[{"name":"str","type":"string"}],"type":"constructor"}]`,
		`
			"math/big"

			"github.com/ethereum/go-ethereum/accounts/abi/bind"
			"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
			"github.com/ethereum/go-ethereum/core"
			"github.com/ethereum/go-ethereum/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey()
			auth := bind.NewKeyedTransactor(key)
			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)

			// Deploy an interaction tester contract and call a transaction on it
			_, _, interactor, err := DeployInteractor(auth, sim, "Deploy string")
			if err != nil {
				t.Fatalf("Failed to deploy interactor contract: %v", err)
			}
			if _, err := interactor.Transact(auth, "Transact string"); err != nil {
				t.Fatalf("Failed to transact with interactor contract: %v", err)
			}
			// Commit all pending transactions in the simulator and check the contract state
			sim.Commit()

			if str, err := interactor.DeployString(nil); err != nil {
				t.Fatalf("Failed to retrieve deploy string: %v", err)
			} else if str != "Deploy string" {
				t.Fatalf("Deploy string mismatch: have '%s', want 'Deploy string'", str)
			}
			if str, err := interactor.TransactString(nil); err != nil {
				t.Fatalf("Failed to retrieve transact string: %v", err)
			} else if str != "Transact string" {
				t.Fatalf("Transact string mismatch: have '%s', want 'Transact string'", str)
			}
		`,
	},
	// Tests that plain values can be properly returned and deserialized
	{
		`Getter`,
		`
			contract Getter {
				function getter() constant returns (string, int, bytes32) {
					return ("Hi", 1, sha3(""));
				}
			}
		`,
		`606060405260dc8060106000396000f3606060405260e060020a6000350463993a04b78114601a575b005b600060605260c0604052600260809081527f486900000000000000000000000000000000000000000000000000000000000060a05260017fc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a47060e0829052610100819052606060c0908152600261012081905281906101409060a09080838184600060046012f1505081517fffff000000000000000000000000000000000000000000000000000000000000169091525050604051610160819003945092505050f3`,
		`[{"constant":true,"inputs":[],"name":"getter","outputs":[{"name":"","type":"string"},{"name":"","type":"int256"},{"name":"","type":"bytes32"}],"type":"function"}]`,
		`
			"math/big"

			"github.com/ethereum/go-ethereum/accounts/abi/bind"
			"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
			"github.com/ethereum/go-ethereum/core"
			"github.com/ethereum/go-ethereum/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey()
			auth := bind.NewKeyedTransactor(key)
			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)

			// Deploy a tuple tester contract and execute a structured call on it
			_, _, getter, err := DeployGetter(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy getter contract: %v", err)
			}
			sim.Commit()

			if str, num, _, err := getter.Getter(nil); err != nil {
				t.Fatalf("Failed to call anonymous field retriever: %v", err)
			} else if str != "Hi" || num.Cmp(big.NewInt(1)) != 0 {
				t.Fatalf("Retrieved value mismatch: have %v/%v, want %v/%v", str, num, "Hi", 1)
			}
		`,
	},
	// Tests that tuples can be properly returned and deserialized
	{
		`Tupler`,
		`
			contract Tupler {
				function tuple() constant returns (string a, int b, bytes32 c) {
					return ("Hi", 1, sha3(""));
				}
			}
		`,
		`606060405260dc8060106000396000f3606060405260e060020a60003504633175aae28114601a575b005b600060605260c0604052600260809081527f486900000000000000000000000000000000000000000000000000000000000060a05260017fc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a47060e0829052610100819052606060c0908152600261012081905281906101409060a09080838184600060046012f1505081517fffff000000000000000000000000000000000000000000000000000000000000169091525050604051610160819003945092505050f3`,
		`[{"constant":true,"inputs":[],"name":"tuple","outputs":[{"name":"a","type":"string"},{"name":"b","type":"int256"},{"name":"c","type":"bytes32"}],"type":"function"}]`,
		`
			"math/big"

			"github.com/ethereum/go-ethereum/accounts/abi/bind"
			"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
			"github.com/ethereum/go-ethereum/core"
			"github.com/ethereum/go-ethereum/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey()
			auth := bind.NewKeyedTransactor(key)
			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)

			// Deploy a tuple tester contract and execute a structured call on it
			_, _, tupler, err := DeployTupler(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy tupler contract: %v", err)
			}
			sim.Commit()

			if res, err := tupler.Tuple(nil); err != nil {
				t.Fatalf("Failed to call structure retriever: %v", err)
			} else if res.A != "Hi" || res.B.Cmp(big.NewInt(1)) != 0 {
				t.Fatalf("Retrieved value mismatch: have %v/%v, want %v/%v", res.A, res.B, "Hi", 1)
			}
		`,
	},
	// Tests that arrays/slices can be properly returned and deserialized.
	// Only addresses are tested, remainder just compiled to keep the test small.
	{
		`Slicer`,
		`
			contract Slicer {
				function echoAddresses(address[] input) constant returns (address[] output) {
					return input;
				}
				function echoInts(int[] input) constant returns (int[] output) {
					return input;
				}
				function echoFancyInts(uint24[23] input) constant returns (uint24[23] output) {
					return input;
				}
				function echoBools(bool[] input) constant returns (bool[] output) {
					return input;
				}
			}
		`,
		`606060405261015c806100126000396000f3606060405260e060020a6000350463be1127a3811461003c578063d88becc014610092578063e15a3db71461003c578063f637e5891461003c575b005b604080516020600480358082013583810285810185019096528085526100ee959294602494909392850192829185019084908082843750949650505050505050604080516020810190915260009052805b919050565b604080516102e0818101909252610138916004916102e491839060179083908390808284375090955050505050506102e0604051908101604052806017905b60008152602001906001900390816100d15790505081905061008d565b60405180806020018281038252838181518152602001915080519060200190602002808383829060006004602084601f0104600f02600301f1509050019250505060405180910390f35b60405180826102e0808381846000600461015cf15090500191505060405180910390f3`,
		`[{"constant":true,"inputs":[{"name":"input","type":"address[]"}],"name":"echoAddresses","outputs":[{"name":"output","type":"address[]"}],"type":"function"},{"constant":true,"inputs":[{"name":"input","type":"uint24[23]"}],"name":"echoFancyInts","outputs":[{"name":"output","type":"uint24[23]"}],"type":"function"},{"constant":true,"inputs":[{"name":"input","type":"int256[]"}],"name":"echoInts","outputs":[{"name":"output","type":"int256[]"}],"type":"function"},{"constant":true,"inputs":[{"name":"input","type":"bool[]"}],"name":"echoBools","outputs":[{"name":"output","type":"bool[]"}],"type":"function"}]`,
		`
			"math/big"
			"reflect"

			"github.com/ethereum/go-ethereum/accounts/abi/bind"
			"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
			"github.com/ethereum/go-ethereum/common"
			"github.com/ethereum/go-ethereum/core"
			"github.com/ethereum/go-ethereum/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey()
			auth := bind.NewKeyedTransactor(key)
			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)

			// Deploy a slice tester contract and execute a n array call on it
			_, _, slicer, err := DeploySlicer(auth, sim)
			if err != nil {
					t.Fatalf("Failed to deploy slicer contract: %v", err)
			}
			sim.Commit()

			if out, err := slicer.EchoAddresses(nil, []common.Address{auth.From, common.Address{}}); err != nil {
					t.Fatalf("Failed to call slice echoer: %v", err)
			} else if !reflect.DeepEqual(out, []common.Address{auth.From, common.Address{}}) {
					t.Fatalf("Slice return mismatch: have %v, want %v", out, []common.Address{auth.From, common.Address{}})
			}
		`,
	},
	// Tests that anonymous default methods can be correctly invoked
	{
		`Defaulter`,
		`
			contract Defaulter {
				address public caller;

				function() {
					caller = msg.sender;
				}
			}
		`,
		`6060604052606a8060106000396000f360606040523615601d5760e060020a6000350463fc9c8d3981146040575b605e6000805473ffffffffffffffffffffffffffffffffffffffff191633179055565b606060005473ffffffffffffffffffffffffffffffffffffffff1681565b005b6060908152602090f3`,
		`[{"constant":true,"inputs":[],"name":"caller","outputs":[{"name":"","type":"address"}],"type":"function"}]`,
		`
			"math/big"

			"github.com/ethereum/go-ethereum/accounts/abi/bind"
			"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
			"github.com/ethereum/go-ethereum/core"
			"github.com/ethereum/go-ethereum/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey()
			auth := bind.NewKeyedTransactor(key)
			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)

			// Deploy a default method invoker contract and execute its default method
			_, _, defaulter, err := DeployDefaulter(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy defaulter contract: %v", err)
			}
			if _, err := (&DefaulterRaw{defaulter}).Transfer(auth); err != nil {
				t.Fatalf("Failed to invoke default method: %v", err)
			}
			sim.Commit()

			if caller, err := defaulter.Caller(nil); err != nil {
				t.Fatalf("Failed to call address retriever: %v", err)
			} else if (caller != auth.From) {
				t.Fatalf("Address mismatch: have %v, want %v", caller, auth.From)
			}
		`,
	},
	// Tests that non-existent contracts are reported as such (though only simulator test)
	{
		`NonExistent`,
		`
			contract NonExistent {
				function String() constant returns(string) {
					return "I don't exist";
				}
			}
		`,
		`6060604052609f8060106000396000f3606060405260e060020a6000350463f97a60058114601a575b005b600060605260c0604052600d60809081527f4920646f6e27742065786973740000000000000000000000000000000000000060a052602060c0908152600d60e081905281906101009060a09080838184600060046012f15050815172ffffffffffffffffffffffffffffffffffffff1916909152505060405161012081900392509050f3`,
		`[{"constant":true,"inputs":[],"name":"String","outputs":[{"name":"","type":"string"}],"type":"function"}]`,
		`
			"github.com/ethereum/go-ethereum/accounts/abi/bind"
			"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
			"github.com/ethereum/go-ethereum/common"
		`,
		`
			// Create a simulator and wrap a non-deployed contract
			sim := backends.NewSimulatedBackend(nil, uint64(10000000000))

			nonexistent, err := NewNonExistent(common.Address{}, sim)
			if err != nil {
				t.Fatalf("Failed to access non-existent contract: %v", err)
			}
			// Ensure that contract calls fail with the appropriate error
			if res, err := nonexistent.String(nil); err == nil {
				t.Fatalf("Call succeeded on non-existent contract: %v", res)
			} else if (err != bind.ErrNoCode) {
				t.Fatalf("Error mismatch: have %v, want %v", err, bind.ErrNoCode)
			}
		`,
	},
	// Tests that gas estimation works for contracts with weird gas mechanics too.
	{
		`FunkyGasPattern`,
		`
			contract FunkyGasPattern {
				string public field;

				function SetField(string value) {
					// This check will screw gas estimation! Good, good!
					if (msg.gas < 100000) {
						throw;
					}
					field = value;
				}
			}
		`,
		`606060405261021c806100126000396000f3606060405260e060020a600035046323fcf32a81146100265780634f28bf0e1461007b575b005b6040805160206004803580820135601f8101849004840285018401909552848452610024949193602493909291840191908190840183828082843750949650505050505050620186a05a101561014e57610002565b6100db60008054604080516020601f600260001961010060018816150201909516949094049384018190048102820181019092528281529291908301828280156102145780601f106101e957610100808354040283529160200191610214565b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600302600f01f150905090810190601f16801561013b5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b505050565b8060006000509080519060200190828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106101b557805160ff19168380011785555b506101499291505b808211156101e557600081556001016101a1565b82800160010185558215610199579182015b828111156101995782518260005055916020019190600101906101c7565b5090565b820191906000526020600020905b8154815290600101906020018083116101f757829003601f168201915b50505050508156`,
		`[{"constant":false,"inputs":[{"name":"value","type":"string"}],"name":"SetField","outputs":[],"type":"function"},{"constant":true,"inputs":[],"name":"field","outputs":[{"name":"","type":"string"}],"type":"function"}]`,
		`
			"math/big"

			"github.com/ethereum/go-ethereum/accounts/abi/bind"
			"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
			"github.com/ethereum/go-ethereum/core"
			"github.com/ethereum/go-ethereum/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey()
			auth := bind.NewKeyedTransactor(key)
			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)

			// Deploy a funky gas pattern contract
			_, _, limiter, err := DeployFunkyGasPattern(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy funky contract: %v", err)
			}
			sim.Commit()

			// Set the field with automatic estimation and check that it succeeds
			if _, err := limiter.SetField(auth, "automatic"); err != nil {
				t.Fatalf("Failed to call automatically gased transaction: %v", err)
			}
			sim.Commit()

			if field, _ := limiter.Field(nil); field != "automatic" {
				t.Fatalf("Field mismatch: have %v, want %v", field, "automatic")
			}
		`,
	},
	// Test that constant functions can be called from an (optional) specified address
	{
		`CallFrom`,
		`
			contract CallFrom {
				function callFrom() constant returns(address) {
					return msg.sender;
				}
			}
		`, `6060604052346000575b6086806100176000396000f300606060405263ffffffff60e060020a60003504166349f8e98281146022575b6000565b34600057602c6055565b6040805173ffffffffffffffffffffffffffffffffffffffff9092168252519081900360200190f35b335b905600a165627a7a72305820aef6b7685c0fa24ba6027e4870404a57df701473fe4107741805c19f5138417c0029`,
		`[{"constant":true,"inputs":[],"name":"callFrom","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"}]`,
		`
			"math/big"

			"github.com/ethereum/go-ethereum/accounts/abi/bind"
			"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
			"github.com/ethereum/go-ethereum/common"
			"github.com/ethereum/go-ethereum/core"
			"github.com/ethereum/go-ethereum/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey()
			auth := bind.NewKeyedTransactor(key)
			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)

			// Deploy a sender tester contract and execute a structured call on it
			_, _, callfrom, err := DeployCallFrom(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy sender contract: %v", err)
			}
			sim.Commit()

			if res, err := callfrom.CallFrom(nil); err != nil {
				t.Errorf("Failed to call constant function: %v", err)
			} else if res != (common.Address{}) {
				t.Errorf("Invalid address returned, want: %x, got: %x", (common.Address{}), res)
			}

			for _, addr := range []common.Address{common.Address{}, common.Address{1}, common.Address{2}} {
				if res, err := callfrom.CallFrom(&bind.CallOpts{From: addr}); err != nil {
					t.Fatalf("Failed to call constant function: %v", err)
				} else if res != addr {
					t.Fatalf("Invalid address returned, want: %x, got: %x", addr, res)
				}
			}
		`,
	},
	// Tests that methods and returns with underscores inside work correctly.
	{
		`Underscorer`,
		`
		contract Underscorer {
			function UnderscoredOutput() constant returns (int _int, string _string) {
				return (314, "pi");
			}
			function LowerLowerCollision() constant returns (int _res, int res) {
				return (1, 2);
			}
			function LowerUpperCollision() constant returns (int _res, int Res) {
				return (1, 2);
			}
			function UpperLowerCollision() constant returns (int _Res, int res) {
				return (1, 2);
			}
			function UpperUpperCollision() constant returns (int _Res, int Res) {
				return (1, 2);
			}
			function PurelyUnderscoredOutput() constant returns (int _, int res) {
				return (1, 2);
			}
			function AllPurelyUnderscoredOutput() constant returns (int _, int __) {
				return (1, 2);
			}
			function _under_scored_func() constant returns (int _int) {
				return 0;
			}
		}
		`, `6060604052341561000f57600080fd5b6103858061001e6000396000f30060606040526004361061008e576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806303a592131461009357806346546dbe146100c357806367e6633d146100ec5780639df4848514610181578063af7486ab146101b1578063b564b34d146101e1578063e02ab24d14610211578063e409ca4514610241575b600080fd5b341561009e57600080fd5b6100a6610271565b604051808381526020018281526020019250505060405180910390f35b34156100ce57600080fd5b6100d6610286565b6040518082815260200191505060405180910390f35b34156100f757600080fd5b6100ff61028e565b6040518083815260200180602001828103825283818151815260200191508051906020019080838360005b8381101561014557808201518184015260208101905061012a565b50505050905090810190601f1680156101725780820380516001836020036101000a031916815260200191505b50935050505060405180910390f35b341561018c57600080fd5b6101946102dc565b604051808381526020018281526020019250505060405180910390f35b34156101bc57600080fd5b6101c46102f1565b604051808381526020018281526020019250505060405180910390f35b34156101ec57600080fd5b6101f4610306565b604051808381526020018281526020019250505060405180910390f35b341561021c57600080fd5b61022461031b565b604051808381526020018281526020019250505060405180910390f35b341561024c57600080fd5b610254610330565b604051808381526020018281526020019250505060405180910390f35b60008060016002819150809050915091509091565b600080905090565b6000610298610345565b61013a8090506040805190810160405280600281526020017f7069000000000000000000000000000000000000000000000000000000000000815250915091509091565b60008060016002819150809050915091509091565b60008060016002819150809050915091509091565b60008060016002819150809050915091509091565b60008060016002819150809050915091509091565b60008060016002819150809050915091509091565b6020604051908101604052806000815250905600a165627a7a72305820d1a53d9de9d1e3d55cb3dc591900b63c4f1ded79114f7b79b332684840e186a40029`,
		`[{"constant":true,"inputs":[],"name":"LowerUpperCollision","outputs":[{"name":"_res","type":"int256"},{"name":"Res","type":"int256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"_under_scored_func","outputs":[{"name":"_int","type":"int256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"UnderscoredOutput","outputs":[{"name":"_int","type":"int256"},{"name":"_string","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"PurelyUnderscoredOutput","outputs":[{"name":"_","type":"int256"},{"name":"res","type":"int256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"UpperLowerCollision","outputs":[{"name":"_Res","type":"int256"},{"name":"res","type":"int256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"AllPurelyUnderscoredOutput","outputs":[{"name":"_","type":"int256"},{"name":"__","type":"int256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"UpperUpperCollision","outputs":[{"name":"_Res","type":"int256"},{"name":"Res","type":"int256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"LowerLowerCollision","outputs":[{"name":"_res","type":"int256"},{"name":"res","type":"int256"}],"payable":false,"stateMutability":"view","type":"function"}]`,
		`
			"fmt"
			"math/big"

			"github.com/ethereum/go-ethereum/accounts/abi/bind"
			"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
			"github.com/ethereum/go-ethereum/core"
			"github.com/ethereum/go-ethereum/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey()
			auth := bind.NewKeyedTransactor(key)
			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)

			// Deploy a underscorer tester contract and execute a structured call on it
			_, _, underscorer, err := DeployUnderscorer(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy underscorer contract: %v", err)
			}
			sim.Commit()

			// Verify that underscored return values correctly parse into structs
			if res, err := underscorer.UnderscoredOutput(nil); err != nil {
				t.Errorf("Failed to call constant function: %v", err)
			} else if res.Int.Cmp(big.NewInt(314)) != 0 || res.String != "pi" {
				t.Errorf("Invalid result, want: {314, \"pi\"}, got: %+v", res)
			}
			// Verify that underscored and non-underscored name collisions force tuple outputs
			var a, b *big.Int

			a, b, _ = underscorer.LowerLowerCollision(nil)
			a, b, _ = underscorer.LowerUpperCollision(nil)
			a, b, _ = underscorer.UpperLowerCollision(nil)
			a, b, _ = underscorer.UpperUpperCollision(nil)
			a, b, _ = underscorer.PurelyUnderscoredOutput(nil)
			a, b, _ = underscorer.AllPurelyUnderscoredOutput(nil)
			a, _ = underscorer.UnderScoredFunc(nil)

			fmt.Println(a, b, err)
		`,
	},
	// Tests that logs can be successfully filtered and decoded.
	{
		`Eventer`,
		`
			contract Eventer {
					event SimpleEvent (
					address indexed Addr,
					bytes32 indexed Id,
					bool    indexed Flag,
					uint    Value
				);
				function raiseSimpleEvent(address addr, bytes32 id, bool flag, uint value) {
					SimpleEvent(addr, id, flag, value);
				}

				event NodataEvent (
					uint   indexed Number,
					int16  indexed Short,
					uint32 indexed Long
				);
				function raiseNodataEvent(uint number, int16 short, uint32 long) {
					NodataEvent(number, short, long);
				}

				event DynamicEvent (
					string indexed IndexedString,
					bytes  indexed IndexedBytes,
					string NonIndexedString,
					bytes  NonIndexedBytes
				);
				function raiseDynamicEvent(string str, bytes blob) {
					DynamicEvent(str, blob, str, blob);
				}
			}
		`,
		`6060604052341561000f57600080fd5b61042c8061001e6000396000f300606060405260043610610057576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063528300ff1461005c578063630c31e2146100fc578063c7d116dd14610156575b600080fd5b341561006757600080fd5b6100fa600480803590602001908201803590602001908080601f0160208091040260200160405190810160405280939291908181526020018383808284378201915050505050509190803590602001908201803590602001908080601f01602080910402602001604051908101604052809392919081815260200183838082843782019150505050505091905050610194565b005b341561010757600080fd5b610154600480803573ffffffffffffffffffffffffffffffffffffffff16906020019091908035600019169060200190919080351515906020019091908035906020019091905050610367565b005b341561016157600080fd5b610192600480803590602001909190803560010b90602001909190803563ffffffff169060200190919050506103c3565b005b806040518082805190602001908083835b6020831015156101ca57805182526020820191506020810190506020830392506101a5565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040518091039020826040518082805190602001908083835b60208310151561022d5780518252602082019150602081019050602083039250610208565b6001836020036101000a03801982511681845116808217855250505050505090500191505060405180910390207f3281fd4f5e152dd3385df49104a3f633706e21c9e80672e88d3bcddf33101f008484604051808060200180602001838103835285818151815260200191508051906020019080838360005b838110156102c15780820151818401526020810190506102a6565b50505050905090810190601f1680156102ee5780820380516001836020036101000a031916815260200191505b50838103825284818151815260200191508051906020019080838360005b8381101561032757808201518184015260208101905061030c565b50505050905090810190601f1680156103545780820380516001836020036101000a031916815260200191505b5094505050505060405180910390a35050565b81151583600019168573ffffffffffffffffffffffffffffffffffffffff167f1f097de4289df643bd9c11011cc61367aa12983405c021056e706eb5ba1250c8846040518082815260200191505060405180910390a450505050565b8063ffffffff168260010b847f3ca7f3a77e5e6e15e781850bc82e32adfa378a2a609370db24b4d0fae10da2c960405160405180910390a45050505600a165627a7a72305820d1f8a8bbddbc5bb29f285891d6ae1eef8420c52afdc05e1573f6114d8e1714710029`,
		`[{"constant":false,"inputs":[{"name":"str","type":"string"},{"name":"blob","type":"bytes"}],"name":"raiseDynamicEvent","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"addr","type":"address"},{"name":"id","type":"bytes32"},{"name":"flag","type":"bool"},{"name":"value","type":"uint256"}],"name":"raiseSimpleEvent","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"number","type":"uint256"},{"name":"short","type":"int16"},{"name":"long","type":"uint32"}],"name":"raiseNodataEvent","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"anonymous":false,"inputs":[{"indexed":true,"name":"Addr","type":"address"},{"indexed":true,"name":"Id","type":"bytes32"},{"indexed":true,"name":"Flag","type":"bool"},{"indexed":false,"name":"Value","type":"uint256"}],"name":"SimpleEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"Number","type":"uint256"},{"indexed":true,"name":"Short","type":"int16"},{"indexed":true,"name":"Long","type":"uint32"}],"name":"NodataEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"IndexedString","type":"string"},{"indexed":true,"name":"IndexedBytes","type":"bytes"},{"indexed":false,"name":"NonIndexedString","type":"string"},{"indexed":false,"name":"NonIndexedBytes","type":"bytes"}],"name":"DynamicEvent","type":"event"}]`,
		`
			"math/big"
			"time"

			"github.com/ethereum/go-ethereum/accounts/abi/bind"
			"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
			"github.com/ethereum/go-ethereum/common"
			"github.com/ethereum/go-ethereum/core"
			"github.com/ethereum/go-ethereum/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey()
			auth := bind.NewKeyedTransactor(key)
			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)

			// Deploy an eventer contract
			_, _, eventer, err := DeployEventer(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy eventer contract: %v", err)
			}
			sim.Commit()

			// Inject a few events into the contract, gradually more in each block
			for i := 1; i <= 3; i++ {
				for j := 1; j <= i; j++ {
					if _, err := eventer.RaiseSimpleEvent(auth, common.Address{byte(j)}, [32]byte{byte(j)}, true, big.NewInt(int64(10*i+j))); err != nil {
						t.Fatalf("block %d, event %d: raise failed: %v", i, j, err)
					}
				}
				sim.Commit()
			}
			// Test filtering for certain events and ensure they can be found
			sit, err := eventer.FilterSimpleEvent(nil, []common.Address{common.Address{1}, common.Address{3}}, [][32]byte{{byte(1)}, {byte(2)}, {byte(3)}}, []bool{true})
			if err != nil {
				t.Fatalf("failed to filter for simple events: %v", err)
			}
			defer sit.Close()

			sit.Next()
			if sit.Event.Value.Uint64() != 11 || !sit.Event.Flag {
				t.Errorf("simple log content mismatch: have %v, want {11, true}", sit.Event)
			}
			sit.Next()
			if sit.Event.Value.Uint64() != 21 || !sit.Event.Flag {
				t.Errorf("simple log content mismatch: have %v, want {21, true}", sit.Event)
			}
			sit.Next()
			if sit.Event.Value.Uint64() != 31 || !sit.Event.Flag {
				t.Errorf("simple log content mismatch: have %v, want {31, true}", sit.Event)
			}
			sit.Next()
			if sit.Event.Value.Uint64() != 33 || !sit.Event.Flag {
				t.Errorf("simple log content mismatch: have %v, want {33, true}", sit.Event)
			}

			if sit.Next() {
				t.Errorf("unexpected simple event found: %+v", sit.Event)
			}
			if err = sit.Error(); err != nil {
				t.Fatalf("simple event iteration failed: %v", err)
			}
			// Test raising and filtering for an event with no data component
			if _, err := eventer.RaiseNodataEvent(auth, big.NewInt(314), 141, 271); err != nil {
				t.Fatalf("failed to raise nodata event: %v", err)
			}
			sim.Commit()

			nit, err := eventer.FilterNodataEvent(nil, []*big.Int{big.NewInt(314)}, []int16{140, 141, 142}, []uint32{271})
			if err != nil {
				t.Fatalf("failed to filter for nodata events: %v", err)
			}
			defer nit.Close()

			if !nit.Next() {
				t.Fatalf("nodata log not found: %v", nit.Error())
			}
			if nit.Event.Number.Uint64() != 314 {
				t.Errorf("nodata log content mismatch: have %v, want 314", nit.Event.Number)
			}
			if nit.Next() {
				t.Errorf("unexpected nodata event found: %+v", nit.Event)
			}
			if err = nit.Error(); err != nil {
				t.Fatalf("nodata event iteration failed: %v", err)
			}
			// Test raising and filtering for events with dynamic indexed components
			if _, err := eventer.RaiseDynamicEvent(auth, "Hello", []byte("World")); err != nil {
				t.Fatalf("failed to raise dynamic event: %v", err)
			}
			sim.Commit()

			dit, err := eventer.FilterDynamicEvent(nil, []string{"Hi", "Hello", "Bye"}, [][]byte{[]byte("World")})
			if err != nil {
				t.Fatalf("failed to filter for dynamic events: %v", err)
			}
			defer dit.Close()

			if !dit.Next() {
				t.Fatalf("dynamic log not found: %v", dit.Error())
			}
			if dit.Event.NonIndexedString != "Hello" || string(dit.Event.NonIndexedBytes) != "World" ||	dit.Event.IndexedString != common.HexToHash("0x06b3dfaec148fb1bb2b066f10ec285e7c9bf402ab32aa78a5d38e34566810cd2") || dit.Event.IndexedBytes != common.HexToHash("0xf2208c967df089f60420785795c0a9ba8896b0f6f1867fa7f1f12ad6f79c1a18") {
				t.Errorf("dynamic log content mismatch: have %v, want {'0x06b3dfaec148fb1bb2b066f10ec285e7c9bf402ab32aa78a5d38e34566810cd2, '0xf2208c967df089f60420785795c0a9ba8896b0f6f1867fa7f1f12ad6f79c1a18', 'Hello', 'World'}", dit.Event)
			}
			if dit.Next() {
				t.Errorf("unexpected dynamic event found: %+v", dit.Event)
			}
			if err = dit.Error(); err != nil {
				t.Fatalf("dynamic event iteration failed: %v", err)
			}
			// Test subscribing to an event and raising it afterwards
			ch := make(chan *EventerSimpleEvent, 16)
			sub, err := eventer.WatchSimpleEvent(nil, ch, nil, nil, nil)
			if err != nil {
				t.Fatalf("failed to subscribe to simple events: %v", err)
			}
			if _, err := eventer.RaiseSimpleEvent(auth, common.Address{255}, [32]byte{255}, true, big.NewInt(255)); err != nil {
				t.Fatalf("failed to raise subscribed simple event: %v", err)
			}
			sim.Commit()

			select {
			case event := <-ch:
				if event.Value.Uint64() != 255 {
					t.Errorf("simple log content mismatch: have %v, want 255", event)
				}
			case <-time.After(250 * time.Millisecond):
				t.Fatalf("subscribed simple event didn't arrive")
			}
			// Unsubscribe from the event and make sure we're not delivered more
			sub.Unsubscribe()

			if _, err := eventer.RaiseSimpleEvent(auth, common.Address{254}, [32]byte{254}, true, big.NewInt(254)); err != nil {
				t.Fatalf("failed to raise subscribed simple event: %v", err)
			}
			sim.Commit()

			select {
			case event := <-ch:
				t.Fatalf("unsubscribed simple event arrived: %v", event)
			case <-time.After(250 * time.Millisecond):
			}
		`,
	},
	{
		`DeeplyNestedArray`,
		`
			contract DeeplyNestedArray {
				uint64[3][4][5] public deepUint64Array;
				function storeDeepUintArray(uint64[3][4][5] arr) public {
					deepUint64Array = arr;
				}
				function retrieveDeepArray() public view returns (uint64[3][4][5]) {
					return deepUint64Array;
				}
			}
		`,
		`6060604052341561000f57600080fd5b6106438061001e6000396000f300606060405260043610610057576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063344248551461005c5780638ed4573a1461011457806398ed1856146101ab575b600080fd5b341561006757600080fd5b610112600480806107800190600580602002604051908101604052809291906000905b828210156101055783826101800201600480602002604051908101604052809291906000905b828210156100f25783826060020160038060200260405190810160405280929190826003602002808284378201915050505050815260200190600101906100b0565b505050508152602001906001019061008a565b5050505091905050610208565b005b341561011f57600080fd5b61012761021d565b604051808260056000925b8184101561019b578284602002015160046000925b8184101561018d5782846020020151600360200280838360005b8381101561017c578082015181840152602081019050610161565b505050509050019260010192610147565b925050509260010192610132565b9250505091505060405180910390f35b34156101b657600080fd5b6101de6004808035906020019091908035906020019091908035906020019091905050610309565b604051808267ffffffffffffffff1667ffffffffffffffff16815260200191505060405180910390f35b80600090600561021992919061035f565b5050565b6102256103b0565b6000600580602002604051908101604052809291906000905b8282101561030057838260040201600480602002604051908101604052809291906000905b828210156102ed578382016003806020026040519081016040528092919082600380156102d9576020028201916000905b82829054906101000a900467ffffffffffffffff1667ffffffffffffffff16815260200190600801906020826007010492830192600103820291508084116102945790505b505050505081526020019060010190610263565b505050508152602001906001019061023e565b50505050905090565b60008360058110151561031857fe5b600402018260048110151561032957fe5b018160038110151561033757fe5b6004918282040191900660080292509250509054906101000a900467ffffffffffffffff1681565b826005600402810192821561039f579160200282015b8281111561039e5782518290600461038e9291906103df565b5091602001919060040190610375565b5b5090506103ac919061042d565b5090565b610780604051908101604052806005905b6103c9610459565b8152602001906001900390816103c15790505090565b826004810192821561041c579160200282015b8281111561041b5782518290600361040b929190610488565b50916020019190600101906103f2565b5b5090506104299190610536565b5090565b61045691905b8082111561045257600081816104499190610562565b50600401610433565b5090565b90565b610180604051908101604052806004905b6104726105a7565b81526020019060019003908161046a5790505090565b82600380016004900481019282156105255791602002820160005b838211156104ef57835183826101000a81548167ffffffffffffffff021916908367ffffffffffffffff16021790555092602001926008016020816007010492830192600103026104a3565b80156105235782816101000a81549067ffffffffffffffff02191690556008016020816007010492830192600103026104ef565b505b50905061053291906105d9565b5090565b61055f91905b8082111561055b57600081816105529190610610565b5060010161053c565b5090565b90565b50600081816105719190610610565b50600101600081816105839190610610565b50600101600081816105959190610610565b5060010160006105a59190610610565b565b6060604051908101604052806003905b600067ffffffffffffffff168152602001906001900390816105b75790505090565b61060d91905b8082111561060957600081816101000a81549067ffffffffffffffff0219169055506001016105df565b5090565b90565b50600090555600a165627a7a7230582087e5a43f6965ab6ef7a4ff056ab80ed78fd8c15cff57715a1bf34ec76a93661c0029`,
		`[{"constant":false,"inputs":[{"name":"arr","type":"uint64[3][4][5]"}],"name":"storeDeepUintArray","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"retrieveDeepArray","outputs":[{"name":"","type":"uint64[3][4][5]"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"uint256"},{"name":"","type":"uint256"},{"name":"","type":"uint256"}],"name":"deepUint64Array","outputs":[{"name":"","type":"uint64"}],"payable":false,"stateMutability":"view","type":"function"}]`,
		`
			"math/big"

			"github.com/ethereum/go-ethereum/accounts/abi/bind"
			"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
			"github.com/ethereum/go-ethereum/core"
			"github.com/ethereum/go-ethereum/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey()
			auth := bind.NewKeyedTransactor(key)
			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)

			//deploy the test contract
			_, _, testContract, err := DeployDeeplyNestedArray(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy test contract: %v", err)
			}

			// Finish deploy.
			sim.Commit()

			//Create coordinate-filled array, for testing purposes.
			testArr := [5][4][3]uint64{}
			for i := 0; i < 5; i++ {
				testArr[i] = [4][3]uint64{}
				for j := 0; j < 4; j++ {
					testArr[i][j] = [3]uint64{}
					for k := 0; k < 3; k++ {
						//pack the coordinates, each array value will be unique, and can be validated easily.
						testArr[i][j][k] = uint64(i) << 16 | uint64(j) << 8 | uint64(k)
					}
				}
			}

			if _, err := testContract.StoreDeepUintArray(&bind.TransactOpts{
				From: auth.From,
				Signer: auth.Signer,
			}, testArr); err != nil {
				t.Fatalf("Failed to store nested array in test contract: %v", err)
			}

			sim.Commit()

			retrievedArr, err := testContract.RetrieveDeepArray(&bind.CallOpts{
				From: auth.From,
				Pending: false,
			})
			if err != nil {
				t.Fatalf("Failed to retrieve nested array from test contract: %v", err)
			}

			//quick check to see if contents were copied
			// (See accounts/abi/unpack_test.go for more extensive testing)
			if retrievedArr[4][3][2] != testArr[4][3][2] {
				t.Fatalf("Retrieved value does not match expected value! got: %d, expected: %d. %v", retrievedArr[4][3][2], testArr[4][3][2], err)
			}
		`,
	},
}

// Tests that packages generated by the binder can be successfully compiled and
// the requested tester run against it.
func TestBindings(t *testing.T) {
	if val, ok := os.LookupEnv("SKIP_KNOWN_FAIL"); ok && val == "1" {
		t.Skip("unit test is broken: conditional test skipping activated")
	}
	// Skip the test if no Go command can be found
	gocmd := runtime.GOROOT() + "/bin/go"
	if !common.FileExist(gocmd) {
		t.Skip("go sdk not found for testing")
	}
	// Create a temporary workspace for the test suite
	ws, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create temporary workspace: %v", err)
	}
	defer os.RemoveAll(ws)

	pkg := filepath.Join(ws, "bindtest")
	if err = os.MkdirAll(pkg, 0700); err != nil {
		t.Fatalf("failed to create package: %v", err)
	}
	// Generate the test suite for all the contracts
	for i, tt := range bindTests {
		// Generate the binding and create a Go source file in the workspace
		bind, err := Bind([]string{tt.name}, []string{tt.abi}, []string{tt.bytecode}, []string{ /*tt.runtimecodes*/ }, "bindtest", LangGo)
		if err != nil {
			t.Fatalf("test %d: failed to generate binding: %v", i, err)
		}
		if err = ioutil.WriteFile(filepath.Join(pkg, strings.ToLower(tt.name)+".go"), []byte(bind), 0600); err != nil {
			t.Fatalf("test %d: failed to write binding: %v", i, err)
		}
		// Generate the test file with the injected test code
		code := fmt.Sprintf(`
			package bindtest

			import (
				"testing"
				%s
			)

			func Test%s(t *testing.T) {
				%s
			}
		`, tt.imports, tt.name, tt.tester)
		if err := ioutil.WriteFile(filepath.Join(pkg, strings.ToLower(tt.name)+"_test.go"), []byte(code), 0600); err != nil {
			t.Fatalf("test %d: failed to write tests: %v", i, err)
		}
	}
	// Test the entire package and report any failures
	cmd := exec.Command(gocmd, "test", "-v", "-count", "1")
	cmd.Dir = pkg
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binding test: %v\n%s", err, out)
	}
}
