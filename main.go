package main

import (
	"net/http"
	"log"
	"encoding/json"
	"io/ioutil"
    "path"
    "os/exec"
    "bytes"
    "fmt"
	"github.com/ethereum/eth-go/ethutil"
	"github.com/ethereum/eth-go/ethcrypto"
)

/*
    To use:
        HTTP json post to /compile with {"code":"(lll ... )"}
        response is simple the compiled byte code

    TODO:
        better response (return errors, too)
        allow requests with multiple code bodies at once ... compression even?
*/

// must have LLL compiler installed!
var PathToLLL = "/root/cpp-ethereum/build/lllc/lllc"

// request object
type Data struct{
	Code string `json:"code"`
}

// read in request body (should be pure lll code)
func CompileHandler(w http.ResponseWriter, r *http.Request){
	body, err := ioutil.ReadAll(r.Body)
	if err != nil{
		return
	}
	var code Data
	log.Println("body:", string(body))
	err = json.Unmarshal(body, &code)
	if err != nil{
		log.Println("err on json unmarshal", err)
	}
	log.Println("unmarshal:", code)
    // take sha3 of request object to get tmp filename
    filename := path.Join("tmp", ethutil.Bytes2Hex(ethcrypto.Sha3Bin(body)) + ".lll")
    // lllc requires to read from process
	ioutil.WriteFile(filename, ethutil.Hex2Bytes(code.Code), 0644)

	compiled, err := CompileLLL(filename)	
    log.Println("compiled", err)
    w.Write(compiled)
}

// wrapper to lllc cli
func CompileLLL(filename string) ([]byte, error){
    cmd := exec.Command(PathToLLL, filename)
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        fmt.Println("Couldn't compile!!", err)
        return nil, err
    }
    outstr := out.String()
    // get rid of new lines at the end
    for l:=len(outstr);outstr[l-1] == '\n';l--{
        outstr = outstr[:l-1]
    }
    fmt.Println("script hex", outstr)
    return ethutil.Hex2Bytes(outstr), nil
}


func main(){
	mux := http.NewServeMux()
	mux.HandleFunc("/compile", CompileHandler)
	err := http.ListenAndServe(":9999", mux)
	if err != nil{
		log.Println("error starting server:", err)
	}
}
