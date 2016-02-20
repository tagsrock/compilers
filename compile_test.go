package compilers

import (
	"bytes"
	"encoding/hex"
	"fmt"
	// "path"
	"testing"
)

func init() {
	ClearCaches()
}

// test the result of compiling through the lllc pipeline vs giving it to the wrapper
func testContract(t *testing.T, file string) {
	our_resp := Compile(file, "")
	if our_resp.Error != "" {
		t.Fatal(fmt.Errorf(our_resp.Error))
	}
	if len(our_resp.Objects) == 0 {
		t.Fatal(fmt.Errorf("Output is empty!"))
	}

	lang, _ := LangFromFile(file)
	truth_resp := CompileWrapper(file, lang, []string{}, "")
	if truth_resp.Error != "" {
		t.Fatal(fmt.Errorf(truth_resp.Error))
	}
	if len(truth_resp.Objects) == 0 {
		t.Fatal(fmt.Errorf("Output is empty!"))
	}
	if len(our_resp.Objects) != len(truth_resp.Objects) {
		t.Fatal(fmt.Errorf("Number of compiled objects differ!"))
	}
	for i, r := range our_resp.Objects {
		N := 100
		printCodeTop("us", r.Bytecode, N)
		printCodeTop("them", truth_resp.Objects[i].Bytecode, N)
		if bytes.Compare(r.Bytecode, truth_resp.Objects[i].Bytecode) != 0 {
			t.Fatal(fmt.Errorf("Difference of %d", bytes.Compare(r.Bytecode, truth_resp.Objects[i].Bytecode)))
		}
		if r.ABI != truth_resp.Objects[i].ABI {
			t.Fatal(fmt.Errorf("ABI results don't match:", r.ABI, truth_resp.Objects[i].ABI))
		}
	}
}

func testLocalRemote(t *testing.T, lang, filename string) {
	ClearCaches()
	SetLanguageNet(lang, false)
	testContract(t, filename)
	ClearCaches()
	SetLanguageNet(lang, true)
	testContract(t, filename)
	ClearCaches()
}

func TestLLLClientLocal(t *testing.T) {
	ClearCaches()
	SetLanguageNet("lll", false)
	// testContract(t, "tests/namereg.lll")
	// Note: can't test more complex ones against the native compiler
	// since it doesnt handle paths in the includes...
	//testContract(t, path.Join(utils.ErisLtd, "eris-std-lib", "DTT", "tests", "stdarraytest.lll"))
}

func TestLLLClientRemote(t *testing.T) {
	// testLocalRemote(t, "lll", "tests/namereg.lll")
}

func TestSerpentClientLocal(t *testing.T) {
	ClearCaches()
	SetLanguageNet("se", false)
	// testContract(t, "tests/test.se")
}

func TestSerpentClientRemote(t *testing.T) {
	// testLocalRemote(t, "se", "tests/test.se")
	// testLocalRemote(t, "se", path.Join(homeDir(), "serpent", "examples", "schellingcoin", "schellingcoin.se"))
}

func printCodeTop(s string, code []byte, n int) {
	fmt.Println("length:", len(code))
	if len(code) > n {
		code = code[:n]
	}
	fmt.Printf("%s\t %s\n", s, hex.EncodeToString(code))
}
