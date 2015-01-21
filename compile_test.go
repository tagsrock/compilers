package lllcserver

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
)

func init() {
	ClearCaches()
}

func testContract(t *testing.T, file string) {
	our_code, err := Compile(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(our_code) == 0 {
		t.Fatal(fmt.Errorf("Output is empty!"))
	}

	lang, _ := LangFromFile(file)
	truth_code, err := CompileWrapper(file, lang)
	if err != nil {
		t.Fatal(err)
	}
	if len(truth_code) == 0 {
		t.Fatal(fmt.Errorf("Output is empty!"))
	}
	N := 100
	printCodeTop("us", our_code, N)
	printCodeTop("them", truth_code, N)
	if bytes.Compare(our_code, truth_code) != 0 {
		t.Fatal(err)
	}
}

func printCodeTop(s string, code []byte, n int) {
	fmt.Println("length:", len(code))

	if len(code) > n {
		code = code[:n]
	}
	fmt.Printf("%s\t %s\n", s, hex.EncodeToString(code))
}

func TestLLLClientLocal(t *testing.T) {
	ClearCaches()
	SetLanguageNet("lll", false)
	testContract(t, "tests/namereg.lll")
	// Note: can't test more complex ones against the native compiler
	// since it doesnt handle paths in the includes...
	//testContract(t, path.Join(utils.ErisLtd, "eris-std-lib", "DTT", "tests", "stdarraytest.lll"))
}

func TestLLLClientRemote(t *testing.T) {
	ClearCaches()
	SetLanguageNet("lll", false)
	testContract(t, "tests/namereg.lll")
	ClearCaches()
	SetLanguageNet("lll", true)
	testContract(t, "tests/namereg.lll")
	ClearCaches()
}

func TestSerpentClientLocal(t *testing.T) {
	ClearCaches()
	SetLanguageNet("se", false)
	testContract(t, "tests/test.se")
}

func TestSerpentClientRemote(t *testing.T) {
	ClearCaches()
	SetLanguageNet("se", false)
	testContract(t, "tests/test.se")
	ClearCaches()
	SetLanguageNet("se", true)
	testContract(t, "tests/test.se")
	ClearCaches()
}
