package lllcserver

import (
	"net/http"
	"log"
    "path"
    "regexp"
	"encoding/json"
	"encoding/hex"
    "crypto/sha256"
	"io/ioutil"
    "bytes"
    "fmt"
)

var URL = "http://localhost:9999/compile"

// takes a list of lll scripts (source code, not filenames)
// returns a response object (contains list of compiled bytecodes and errors if any)
func CompileLLLClient(filenames []string) (*Response, error){
    // empty request obj
    req := Request{
        Scripts: [][]byte{},
        Includes: make(map[string][]byte),
    }
   
    included := make(map[string][]byte)
    
    for _, f := range filenames{
        code, err  := ioutil.ReadFile(f) 
        if err != nil{
            log.Println("failed to read file", err)
            return nil, err
        }
        dir := path.Dir(f)

        // find includes, load those as well
        r, _ :=  regexp.Compile(`\(include (.+?)\)`)
        // replace all includes with hash of included lll
        code = r.ReplaceAllFunc(code, func(s []byte)[]byte{
            m := r.FindSubmatch(s)
            match := m[1]
            name := path.Base(string(match))
            // if we've already loaded this, move on
            if v, ok := included[name]; ok{
                return v
            }
            // load the file
            p := path.Join(dir, string(match))
            code, err  := ioutil.ReadFile(p) 
            // TODO: how to make this return nil, err up a call
            if err != nil{
                log.Println("failed to read include file", err)
                return nil
            }
            // compute hash
            hash := sha256.Sum256(code)
            h := hex.EncodeToString(hash[:])
            req.Includes[h] = code
            ret := []byte(`(include "`+h+`.lll")`)
            included[name] = ret
            return ret
        })
        req.Scripts = append(req.Scripts, code)
    }
    reqJ, err := json.Marshal(req)
    if err != nil{
        log.Println("failed to marshal req obj", err)
        return nil, err
    }
    
    httpreq, err := http.NewRequest("POST", URL, bytes.NewBuffer(reqJ))       
    httpreq.Header.Set("Content-Type", "application/json")                    
    
    client := &http.Client{}                                              
    resp, err := client.Do(httpreq)
    if err != nil{
        log.Println("failed!", err)                                       
        return nil, err
    }   
    defer resp.Body.Close()                                               
    // read in response body
    body, err := ioutil.ReadAll(resp.Body)
    var respJ Response
    err = json.Unmarshal(body, &respJ)                                    
    if err != nil{
        return nil , err
    }   

    return &respJ, nil
}  

// compile just one file
// but resolve "includes"
func Compile(filename string) ([]byte, error){
    r, err := CompileLLLClient([]string{filename})
    if err != nil{
        return nil, err
    }
    b := r.Bytecode[0]
    if r.Error[0] != ""{
        err = fmt.Errorf(r.Error[0])
    } else {
        err = nil
    }
    return b, err
}

func RunClient(tocompile []string){
    r, _ := CompileLLLClient(tocompile) 
    for i, c := range r.Bytecode{
        if r.Error[i] != ""{
            log.Println("script", i, "\tcompilation failed:", r.Error[i])
        } else{
            log.Println("script", i, "\tcompilation successful", c)
        }
    }
}
