package compilersTest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"reflect"
	"strings"
	"testing"

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
	var testMap = map[string]*util.IncludedFiles{
		"13db7b5ea4e589c03c4b09b692723247c4029ab59047957940b06e1611be66ba.sol": {
			ObjectNames: []string{"c"},
			Script:      []byte(contractCode),
		},
	}

	req, err := util.CreateRequest("simpleContract.sol", "", false)
	if err != nil {
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

	expectedSolcResponse := util.BlankSolcResponse()

	actualOutput, err := exec.Command("solc", "--combined-json", "bin,abi", "simpleContract.sol").Output()
	if err != nil {
		t.Fatal(err)
	}
	output := strings.TrimSpace(string(actualOutput))
	err = json.Unmarshal([]byte(output), expectedSolcResponse)

	respItemArray := make([]util.ResponseItem, 0)

	for contract, item := range expectedSolcResponse.Contracts {
		respItem := util.ResponseItem{
			Objectname: strings.TrimSpace(contract),
			Bytecode:   strings.TrimSpace(item.Bin),
			ABI:        strings.TrimSpace(item.Abi),
		}
		respItemArray = append(respItemArray, respItem)
	}
	expectedResponse := &util.Response{
		Objects: respItemArray,
		Error:   "",
	}
	util.ClearCache(common.SolcScratchPath)
	t.Log(testServer.URL)
	resp, err := net.BeginCompile(testServer.URL, "simpleContract.sol", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expectedResponse, resp) {
		t.Errorf("Got incorrect response, expected %v, \n\n got %v", expectedResponse, resp)
	}

	util.ClearCache(common.SolcScratchPath)
}

func TestServerMulti(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(net.CompileHandler))
	defer testServer.Close()
	util.ClearCache(common.SolcScratchPath)
	expectedSolcResponse := util.BlankSolcResponse()

	actualOutput, err := exec.Command("solc", "--combined-json", "bin,abi", "contractImport1.sol").Output()
	if err != nil {
		t.Fatal(err)
	}
	output := strings.TrimSpace(string(actualOutput))
	err = json.Unmarshal([]byte(output), expectedSolcResponse)

	respItemArray := make([]util.ResponseItem, 0)

	for contract, item := range expectedSolcResponse.Contracts {
		respItem := util.ResponseItem{
			Objectname: strings.TrimSpace(contract),
			Bytecode:   strings.TrimSpace(item.Bin),
			ABI:        strings.TrimSpace(item.Abi),
		}
		respItemArray = append(respItemArray, respItem)
	}
	expectedResponse := &util.Response{
		Objects: respItemArray,
		Error:   "",
	}
	util.ClearCache(common.SolcScratchPath)
	t.Log(testServer.URL)
	resp, err := net.BeginCompile(testServer.URL, "contractImport1.sol", false, "")
	if err != nil {
		t.Fatal(err)
	}
	allClear := true
	for _, object := range expectedResponse.Objects {
		if !contains(resp.Objects, object) {
			allClear = false
		}
	}
	if !allClear {
		t.Errorf("Got incorrect response, expected %v, \n\n got %v", expectedResponse, resp)
	}
	util.ClearCache(common.SolcScratchPath)
}

func TestLocalMulti(t *testing.T) {
	util.ClearCache(common.SolcScratchPath)
	expectedSolcResponse := util.BlankSolcResponse()

	actualOutput, err := exec.Command("solc", "--combined-json", "bin,abi", "contractImport1.sol").Output()
	if err != nil {
		t.Fatal(err)
	}
	output := strings.TrimSpace(string(actualOutput))
	err = json.Unmarshal([]byte(output), expectedSolcResponse)

	respItemArray := make([]util.ResponseItem, 0)

	for contract, item := range expectedSolcResponse.Contracts {
		respItem := util.ResponseItem{
			Objectname: strings.TrimSpace(contract),
			Bytecode:   strings.TrimSpace(item.Bin),
			ABI:        strings.TrimSpace(item.Abi),
		}
		respItemArray = append(respItemArray, respItem)
	}
	expectedResponse := &util.Response{
		Objects: respItemArray,
		Error:   "",
	}
	util.ClearCache(common.SolcScratchPath)
	resp, err := net.BeginCompile("", "contractImport1.sol", false, "")
	if err != nil {
		t.Fatal(err)
	}
	allClear := true
	for _, object := range expectedResponse.Objects {
		if !contains(resp.Objects, object) {
			allClear = false
		}
	}
	if !allClear {
		t.Errorf("Got incorrect response, expected %v, \n\n got %v", expectedResponse, resp)
	}
	util.ClearCache(common.SolcScratchPath)
}

func TestLocalSingle(t *testing.T) {
	util.ClearCache(common.SolcScratchPath)
	expectedSolcResponse := util.BlankSolcResponse()

	actualOutput, err := exec.Command("solc", "--combined-json", "bin,abi", "simpleContract.sol").Output()
	if err != nil {
		t.Fatal(err)
	}
	output := strings.TrimSpace(string(actualOutput))
	err = json.Unmarshal([]byte(output), expectedSolcResponse)

	respItemArray := make([]util.ResponseItem, 0)

	for contract, item := range expectedSolcResponse.Contracts {
		respItem := util.ResponseItem{
			Objectname: strings.TrimSpace(contract),
			Bytecode:   strings.TrimSpace(item.Bin),
			ABI:        strings.TrimSpace(item.Abi),
		}
		respItemArray = append(respItemArray, respItem)
	}
	expectedResponse := &util.Response{
		Objects: respItemArray,
		Error:   "",
	}
	util.ClearCache(common.SolcScratchPath)
	resp, err := net.BeginCompile("", "simpleContract.sol", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expectedResponse, resp) {
		t.Errorf("Got incorrect response, expected %v, \n\n got %v", expectedResponse, resp)
	}
	util.ClearCache(common.SolcScratchPath)
}

func TestFaultyContract(t *testing.T) {
	util.ClearCache(common.SolcScratchPath)
	var expectedSolcResponse util.Response

	actualOutput, err := exec.Command("solc", "--combined-json", "bin,abi", "faultyContract.sol").CombinedOutput()
	err = json.Unmarshal(actualOutput, expectedSolcResponse)
	t.Log(expectedSolcResponse.Error)
	resp, err := net.BeginCompile("", "faultyContract.sol", false, "")
	t.Log(resp.Error)
	if err != nil {
		if expectedSolcResponse.Error != resp.Error {
			t.Errorf("Expected %v got %v", expectedSolcResponse.Error, resp.Error)
		}
	}
	output := strings.TrimSpace(string(actualOutput))
	err = json.Unmarshal([]byte(output), expectedSolcResponse)
}

func contains(s []util.ResponseItem, e util.ResponseItem) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
