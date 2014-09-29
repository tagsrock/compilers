package lllcserver

import (
	"net/http"
	"log"
	"encoding/json"
	"io/ioutil"
    "path"
    "os"
    "os/exec"
    "bytes"
    "fmt"
	"github.com/ethereum/eth-go/ethutil"
	"github.com/ethereum/eth-go/ethcrypto"
)

/*
    To use:
        HTTP json post to /compile with {"code":"(lll ... )"}
        response is simply the compiled byte code
        uses arrays so we can pass multiple scripts at once
*/

// must have LLL compiler installed!
var PathToLLL = "/root/cpp-ethereum/build/lllc/lllc"

// request object
type Request struct{
	Code []string `json:"code"` // array of lll scripts to compile
}

// response object
type Response struct{
    Bytecode [][]byte `json:"bytecode"` // array of bytecode scripts to return
    Error []string    `json:"error"` // an error for each script
}

// read in request body (should be pure lll code)
// compile lll, build response object, write
func CompileHandler(w http.ResponseWriter, r *http.Request){

    // read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil{
        log.Println("err on read http request body", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}

    // unmarshall body into req struct
	var req Request
	err = json.Unmarshal(body, &req)
	if err != nil{
		log.Println("err on json unmarshal", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}
   
    resp := Response{[][]byte{}, []string{}}
    
    // loop through the scripts, save each to drive, compile, return bytecode and error 
    for _, c := range req.Code{
        // take sha3 of request object to get tmp filename
        filename := path.Join("tmp", ethutil.Bytes2Hex(ethcrypto.Sha3Bin([]byte(c))) + ".lll")

        // lllc requires a file to read
        // check if filename already exists
        if _, err = os.Stat(filename); err != nil{
            ioutil.WriteFile(filename, []byte(c),0644)
        }

        compiled, err := CompileLLLWrapper(filename)	
        if err != nil{
           resp.Error = append(resp.Error, err.Error())
        } else{
           resp.Error = append(resp.Error, "")
        }
        resp.Bytecode = append(resp.Bytecode, compiled)
    }

    respJ, err := json.Marshal(resp)
    if err != nil{
        fmt.Println("failed to marshal", err)
        return
    }
    w.Write(respJ)
}

// wrapper to lllc cli
func CompileLLLWrapper(filename string) ([]byte, error){
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

func StartServer(){
	mux := http.NewServeMux()
	mux.HandleFunc("/compile", CompileHandler)
	err := http.ListenAndServe(":9999", mux)
	if err != nil{
		log.Println("error starting server:", err)
	}
}
