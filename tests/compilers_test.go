package compilersTest

import (
	"testing"
	//"bytes"
	//"encoding/hex"
	//"path"
	"reflect"
	"net/http/httptest"
	"net/http"

	net "github.com/eris-ltd/eris-compilers/network"
	"github.com/eris-ltd/eris-compilers/util"

	"github.com/eris-ltd/common/go/common"
)


func TestRequestCreation(t *testing.T) {
	var err error
	contractCode := `contract c {
    function f() {
        uint8[5] memory foo3 = [1, 1, 1, 1, 1];
    }
}`
	var testMap = map[string]*util.IncludedFiles {
		"13db7b5ea4e589c03c4b09b692723247c4029ab59047957940b06e1611be66ba.sol": {
			ObjectNames: []string{"c"},
			Script: []byte(contractCode),
		},
	}

	req, err := util.CreateRequest("simpleContract.sol", "", false)
	if (err != nil) {
		t.Fatal(err)
	}
	if req.Libraries != "" {
		t.Errorf("Expected empty libraries, got ", req.Libraries)
	}
	if req.Language != "sol" {
		t.Errorf("Expected Solidity file, got ", req.Language)
	}
	if req.Optimize != false {
		t.Errorf("Expected false optimize, got true")
	}
	if !reflect.DeepEqual(req.Includes, testMap) {
		t.Errorf("Got incorrect Includes map")
	}

}

func TestServerSingle(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(net.CompileHandler))
	defer testServer.Close()
	
	simpleByteCode := "6060604052609e8060106000396000f360606040526000357c01000000000000000000000000000000000000000000000000000000009004806326121ff0146037576035565b005b604260048050506044565b005b60a0604051908101604052806005905b600081526020019060019003908160545790505060a06040519081016040528060018152602001600181526020016001815260200160018152602001600181526020015090505b5056"
	simpleJson := `[{"constant":false,"inputs":[],"name":"f","outputs":[],"type":"function"}]`
	util.ClearCache(common.SolcScratchPath)
	t.Log(testServer.URL)
	resp, err := net.BeginCompile(testServer.URL, "simpleContract.sol", false, "")
	if err != nil {
		t.Fatal(err)
	}
	simpleContractResponse := resp.Objects[0]
	if simpleContractResponse.Objectname != "c" {
		t.Errorf("Incorrect object name, expected c, got: ", simpleContractResponse.Objectname)
	}
	if simpleContractResponse.Bytecode != simpleByteCode {
		t.Errorf("Incorrect bytecode, expected %v, got %v ", simpleByteCode, simpleContractResponse.Bytecode)
	}
	if simpleContractResponse.ABI != simpleJson {
		t.Errorf("Incorrect abi, expected %v, got %v ", simpleJson, simpleContractResponse.ABI)
	}
	util.ClearCache(common.SolcScratchPath)
}

func TestServerMulti(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(net.CompileHandler))
	defer testServer.Close()
	util.ClearCache(common.SolcScratchPath)
	var expectedResponseArray = []util.ResponseItem {
		{
			"bar", 
			"606060405260d88060106000396000f360606040526000357c01000000000000000000000000000000000000000000000000000000009004806394aef022146037576035565b005b60426004805050605f565b604051808381526020018281526020019250505060405180910390f35b600060006060604051908101604052806001815260200160028152602001600381526020015060006000506000820151816000016000505560208201518160010160005055604082015181600201600050559050506000600050600001600050546000600050600101600050549150915060d4565b909156",
			`[{"constant":false,"inputs":[],"name":"getVariables","outputs":[{"name":"","type":"uint256"},{"name":"","type":"uint256"}],"type":"function"}]`,
		},
		{
			"c",
			"6060604052604051606080610179833981016040528080519060200190919080519060200190919080519060200190919050505b82600060006101000a81548173ffffffffffffffffffffffffffffffffffffffff0219169083021790555081600160005081905550806002600050819055505b50505060f6806100836000396000f360606040526000357c0100000000000000000000000000000000000000000000000000000000900480634f2be91f146037576035565b005b604260048050506044565b005b600060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663a5f3c23b600160005054600260005054604051837c010000000000000000000000000000000000000000000000000000000002815260040180838152602001828152602001925050506020604051808303816000876161da5a03f115600257505050604051805190602001506001600050819055505b56",
			`[{"constant":false,"inputs":[],"name":"add","outputs":[],"type":"function"},{"inputs":[{"name":"Addr","type":"address"},{"name":"a","type":"int256"},{"name":"b","type":"int256"}],"type":"constructor"}]`,
		},
		{
			"importedContract",
			"6060604052610199806100126000396000f360606040526000357c01000000000000000000000000000000000000000000000000000000009004806348d597e51461004f578063a5f3c23b14610084578063b93ea812146100b95761004d565b005b61006e60048080359060200190919080359060200190919050506100ee565b6040518082815260200191505060405180910390f35b6100a36004808035906020019091908035906020019091905050610175565b6040518082815260200191505060405180910390f35b6100d86004808035906020019091908035906020019091905050610187565b6040518082815260200191505060405180910390f35b60006000828173ffffffffffffffffffffffffffffffffffffffff16636446afde86604051827c0100000000000000000000000000000000000000000000000000000000028152600401808281526020019150506020604051808303816000876161da5a03f115610002575050506040518051906020015001915061016e565b5092915050565b60008183019050610181565b92915050565b60008183039050610193565b9291505056",
			`[{"constant":false,"inputs":[{"name":"a","type":"uint256"},{"name":"b","type":"uint256"}],"name":"addFromMapping","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":false,"inputs":[{"name":"a","type":"int256"},{"name":"b","type":"int256"}],"name":"add","outputs":[{"name":"","type":"int256"}],"type":"function"},{"constant":false,"inputs":[{"name":"a","type":"int256"},{"name":"b","type":"int256"}],"name":"subtract","outputs":[{"name":"","type":"int256"}],"type":"function"}]`,
		},
		{
			"map",
			"606060405260888060106000396000f360606040526000357c0100000000000000000000000000000000000000000000000000000000900480636446afde146037576035565b005b604b60048080359060200190919050506061565b6040518082815260200191505060405180910390f35b6000600060005060008381526020019081526020016000206000505490506083565b91905056",
			`[{"constant":false,"inputs":[{"name":"a","type":"uint256"}],"name":"getMappingElement","outputs":[{"name":"","type":"uint256"}],"type":"function"}]`,
		},
	}
	resp, err := net.BeginCompile(testServer.URL, "contractImport1.sol", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(resp.Objects, expectedResponseArray) {
		t.Errorf("Incorrect output from imported contracts")
	}
	util.ClearCache(common.SolcScratchPath)
}

func TestLocalMulti(t *testing.T) {
	util.ClearCache(common.SolcScratchPath)
	var expectedResponseArray = []util.ResponseItem {
		{
			"bar", 
			"606060405260d88060106000396000f360606040526000357c01000000000000000000000000000000000000000000000000000000009004806394aef022146037576035565b005b60426004805050605f565b604051808381526020018281526020019250505060405180910390f35b600060006060604051908101604052806001815260200160028152602001600381526020015060006000506000820151816000016000505560208201518160010160005055604082015181600201600050559050506000600050600001600050546000600050600101600050549150915060d4565b909156",
			`[{"constant":false,"inputs":[],"name":"getVariables","outputs":[{"name":"","type":"uint256"},{"name":"","type":"uint256"}],"type":"function"}]`,
		},
		{
			"c",
			"6060604052604051606080610179833981016040528080519060200190919080519060200190919080519060200190919050505b82600060006101000a81548173ffffffffffffffffffffffffffffffffffffffff0219169083021790555081600160005081905550806002600050819055505b50505060f6806100836000396000f360606040526000357c0100000000000000000000000000000000000000000000000000000000900480634f2be91f146037576035565b005b604260048050506044565b005b600060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663a5f3c23b600160005054600260005054604051837c010000000000000000000000000000000000000000000000000000000002815260040180838152602001828152602001925050506020604051808303816000876161da5a03f115600257505050604051805190602001506001600050819055505b56",
			`[{"constant":false,"inputs":[],"name":"add","outputs":[],"type":"function"},{"inputs":[{"name":"Addr","type":"address"},{"name":"a","type":"int256"},{"name":"b","type":"int256"}],"type":"constructor"}]`,
		},
		{
			"importedContract",
			"6060604052610199806100126000396000f360606040526000357c01000000000000000000000000000000000000000000000000000000009004806348d597e51461004f578063a5f3c23b14610084578063b93ea812146100b95761004d565b005b61006e60048080359060200190919080359060200190919050506100ee565b6040518082815260200191505060405180910390f35b6100a36004808035906020019091908035906020019091905050610175565b6040518082815260200191505060405180910390f35b6100d86004808035906020019091908035906020019091905050610187565b6040518082815260200191505060405180910390f35b60006000828173ffffffffffffffffffffffffffffffffffffffff16636446afde86604051827c0100000000000000000000000000000000000000000000000000000000028152600401808281526020019150506020604051808303816000876161da5a03f115610002575050506040518051906020015001915061016e565b5092915050565b60008183019050610181565b92915050565b60008183039050610193565b9291505056",
			`[{"constant":false,"inputs":[{"name":"a","type":"uint256"},{"name":"b","type":"uint256"}],"name":"addFromMapping","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":false,"inputs":[{"name":"a","type":"int256"},{"name":"b","type":"int256"}],"name":"add","outputs":[{"name":"","type":"int256"}],"type":"function"},{"constant":false,"inputs":[{"name":"a","type":"int256"},{"name":"b","type":"int256"}],"name":"subtract","outputs":[{"name":"","type":"int256"}],"type":"function"}]`,
		},
		{
			"map",
			"606060405260888060106000396000f360606040526000357c0100000000000000000000000000000000000000000000000000000000900480636446afde146037576035565b005b604b60048080359060200190919050506061565b6040518082815260200191505060405180910390f35b6000600060005060008381526020019081526020016000206000505490506083565b91905056",
			`[{"constant":false,"inputs":[{"name":"a","type":"uint256"}],"name":"getMappingElement","outputs":[{"name":"","type":"uint256"}],"type":"function"}]`,
		},
	}
	resp, err := net.BeginCompile("", "contractImport1.sol", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(resp.Objects, expectedResponseArray) {
		t.Errorf("Incorrect output from imported contracts")
	}
	util.ClearCache(common.SolcScratchPath)
}

func TestLocalSingle(t *testing.T) {	
	simpleByteCode := "6060604052609e8060106000396000f360606040526000357c01000000000000000000000000000000000000000000000000000000009004806326121ff0146037576035565b005b604260048050506044565b005b60a0604051908101604052806005905b600081526020019060019003908160545790505060a06040519081016040528060018152602001600181526020016001815260200160018152602001600181526020015090505b5056"
	simpleJson := `[{"constant":false,"inputs":[],"name":"f","outputs":[],"type":"function"}]`
	util.ClearCache(common.SolcScratchPath)
	resp, err := net.BeginCompile("", "simpleContract.sol", false, "")
	if err != nil {
		t.Fatal(err)
	}
	simpleContractResponse := resp.Objects[0]
	if simpleContractResponse.Objectname != "c" {
		t.Errorf("Incorrect object name, expected c, got: ", simpleContractResponse.Objectname)
	}
	if simpleContractResponse.Bytecode != simpleByteCode {
		t.Errorf("Incorrect bytecode, expected %v, got %v ", simpleByteCode, simpleContractResponse.Bytecode)
	}
	if simpleContractResponse.ABI != simpleJson {
		t.Errorf("Incorrect abi, expected %v, got %v ", simpleJson, simpleContractResponse.ABI)
	}
	util.ClearCache(common.SolcScratchPath)
}