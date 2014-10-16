package lllcserver

import (
	"github.com/go-martini/martini"
	"net/http"
	"log"
    "strings"
	"encoding/json"
    "encoding/hex"
	"io/ioutil"
    "path"
    "path/filepath"
    "os"
    "os/exec"
	"os/user"
    "bytes"
    "fmt"
    "crypto/sha256"
)

/*
    To use:
        HTTP json post to /compile with {"code":"(lll ... )"}
        response is simply the compiled byte code
        uses arrays so we can pass multiple scripts at once
*/

// must have LLL compiler installed!
func homeDir() string{
	usr, err := user.Current()
	if err != nil{
		log.Fatal(err)
	}
	return usr.HomeDir
}

var PathToLLL = path.Join(homeDir(), "cpp-ethereum/build/lllc/lllc")

// request object
// includes are named but scripts are nameless
type Request struct{
	Scripts [][]byte `json:"scripts"` // array of scripts (lll ascii bytes)
    Includes map[string][]byte `json:"includes"` // filename => lll ascii bytes
}

// response object
type Response struct{
    Bytecode [][]byte `json:"bytecode"` // array of bytecode scripts to return
    Error []string    `json:"error"` // an error for each script
}

// convenience wrapper for javascript frontend
func CompileHandler2(w http.ResponseWriter, r *http.Request){
    resp := compileResponse(w, r)
    if resp == nil{
        return 
    }
    code := resp.Bytecode[0]
    hexx := hex.EncodeToString(code)
    w.Write([]byte(fmt.Sprintf(`{"bytecode": "%s"}`, hexx)))
}

// read in request body (should be pure lll code)
// compile lll, build response object, write
func CompileHandler(w http.ResponseWriter, r *http.Request){
    resp := compileResponse(w, r)
    if resp == nil{
        return 
    }
    respJ, err := json.Marshal(resp)
    if err != nil{
        fmt.Println("failed to marshal", err)
        return
    }
    w.Write(respJ)
}

func compileResponse(w http.ResponseWriter, r *http.Request) *Response{
    // read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil{
        log.Println("err on read http request body", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return nil
	}

    // unmarshall body into req struct
	var req Request
	err = json.Unmarshal(body, &req)
	if err != nil{
		log.Println("err on json unmarshal", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return nil
	}
    fmt.Println("request:", req)
   
    resp := Response{[][]byte{}, []string{}}

    names := []string{}
    
    // loop through the scripts, save each to drive
    for _, c := range req.Scripts{
        if c == nil || len(c) == 0{
            names = append(names, "NULLCACHED")
            continue
        }
        // take sha2 of request object to get tmp filename
        hash := sha256.Sum256([]byte(c))
        filename := path.Join("tmp", hex.EncodeToString(hash[:]) + ".lll")
        names = append(names, filename)

        // lllc requires a file to read
        // check if filename already exists. if not, write
        if _, err = os.Stat(filename); err != nil{
            ioutil.WriteFile(filename, c, 0644)
        }
    }
    // loop through includes, also save to drive
    for k, v := range req.Includes{
        filename := path.Join("tmp", k+".lll")
        if _, err = os.Stat(filename); err != nil{
            ioutil.WriteFile(filename, v, 0644)
        }
    }

    //compile scripts, return bytecode and error 
    for _, c := range names{
        fmt.Println("name:", c)
        if c == "NULLCACHED"{
            resp.Error = append(resp.Error, "")
            resp.Bytecode = append(resp.Bytecode, []byte("NULLCACHED"))
            continue
        }
        compiled, err := CompileLLLWrapper(c)
        if err != nil{
           resp.Error = append(resp.Error, err.Error())
        } else{
           resp.Error = append(resp.Error, "")
        }
        resp.Bytecode = append(resp.Bytecode, compiled)
    }
    return &resp
}

// wrapper to lllc cli
func CompileLLLWrapper(filename string) ([]byte, error){
    // we need to be in the same dir as the files for sake of includes
    cur, _ := os.Getwd()
    dir := path.Dir(filename)
    dir, _ = filepath.Abs(dir)
    filename = path.Base(filename)

    os.Chdir(dir)
    cmd := exec.Command(PathToLLL, filename)
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        fmt.Println("Couldn't compile!!", err)
        os.Chdir(cur)
        return nil, err
    }
    os.Chdir(cur)

    outstr := out.String()
    // get rid of new lines at the end
    outstr = strings.TrimRight(outstr, "\n")
    //for l:=len(outstr);outstr[l-1] == '\n';l--{
        //outstr = outstr[:l-1]
    //}
    fmt.Println("script hex", outstr)
    b, err := hex.DecodeString(outstr)
    if err != nil{
        return nil, err
    }
    return b, nil
}

func StartServer(addr string){
	//martini.Env = martini.Prod
	srv := martini.Classic()
	// Static files
	srv.Use(martini.Static("./web"))
	
	srv.Post("/compile", CompileHandler)
	srv.Post("/compile2", CompileHandler2)

	srv.RunOnAddr(addr)
	
}
