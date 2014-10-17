package lllcserver

import (
	"net/http"
	"log"
    "os"
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
var TMP = path.Join(homeDir(), ".lllc")

func replaceIncludes(code []byte, dir string, req Request, included map[string][]byte) []byte{
    // find includes, load those as well
    r, _ :=  regexp.Compile(`\(include "(.+?)"\)`)
    // replace all includes with hash of included lll
    ret := r.ReplaceAllFunc(code, func(s []byte)[]byte{
        m := r.FindSubmatch(s)
        match := m[1]
        name := path.Base(string(match))
        // if we've already loaded this, move on
        if v, ok := included[name]; ok{
            return v
        }
        // load the file
        p := path.Join(dir, string(match))
        incl_code, err  := ioutil.ReadFile(p) 
        // TODO: how to make this return nil, err up a call
        if err != nil{
            log.Println("failed to read include file", err)
            return nil
        }
        incl_code = replaceIncludes(incl_code, dir, req, included)
        // compute hash
        hash := sha256.Sum256(incl_code)
        h := hex.EncodeToString(hash[:])
        req.Includes[h] = incl_code
        ret := []byte(`(include "`+h+`.lll")`)
        included[name] = ret
        return ret
    })
    return ret
}

// takes a list of lll scripts (source code, not filenames)
// returns a response object (contains list of compiled bytecodes and errors if any)
func CompileLLLClient(filenames []string) (*Response, error){
    // cached bytecode
    hashmap := make(map[int][]byte) // map indices to hashes

    // empty request obj
    req := Request{
        Scripts: [][]byte{},
        Includes: make(map[string][]byte),
    }
   
    included := make(map[string][]byte)
    
    for i, f := range filenames{
        code, err  := ioutil.ReadFile(f) 
        if err != nil{
            log.Println("failed to read file", err)
            return nil, err
        }
        dir := path.Dir(f)
        // replace includes with hash of included contents and add those contents to Includes (recursive)
        code = replaceIncludes(code, dir, req, included)

        // if the file is cached, append nil
        hash := sha256.Sum256(code)
        hashmap[i] = hash[:]
        filename := path.Join(TMP, hex.EncodeToString(hash[:])+".lll")
        _, err = os.Stat(filename)
        if err  != nil{
            req.Scripts = append(req.Scripts, code)
        } else{
            req.Scripts = append(req.Scripts, nil)
        }
    }


    // response struct (returned)
    var respJ Response

    // if everything is cached, no need for request
    if len(req.Scripts) == 0 || len(req.Scripts) == 1 && len(req.Scripts[0]) == 0{
        fmt.Println("have all files locally")
        respJ := Response{
            Bytecode : make([][]byte, len(hashmap)),
            Error : make([]string, len(hashmap)),
        }
        // fill in cached values,
        for i, h := range hashmap{
            f := path.Join(TMP, hex.EncodeToString(h)+".lll")
            b, err := ioutil.ReadFile(f)
            if err != nil{
                fmt.Println("read fil:", err)
                return nil, err
            }
            respJ.Bytecode[i] = b
        }
        return &respJ, nil
    }

    // make request
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

    if resp.StatusCode > 300{
        return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
    }

    // read in response body
    body, err := ioutil.ReadAll(resp.Body)
    err = json.Unmarshal(body, &respJ)                                    
    if err != nil{
        fmt.Println("failed to unmarshal", err)
        return nil , err
    }   

    // fill in cached values, cache new values
    for i, b := range respJ.Bytecode{
        f := path.Join(TMP, hex.EncodeToString(hashmap[i])+".lll")
        if string(b) == "NULLCACHED"{
            respJ.Bytecode[i], err = ioutil.ReadFile(f)
            if err != nil{
                fmt.Println("read fil:", err)
            }
        } else if b != nil{
            fmt.Println("sacving byte code:", b)
            ioutil.WriteFile(f, b, 0644)
        }
    }
    //fmt.Println(respJ.Bytecode[0])

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
    r, err := CompileLLLClient(tocompile) 
    if err != nil{
        fmt.Println("shucks", err)
        os.Exit(0)
    }
    for i, c := range r.Bytecode{
        if r.Error[i] != ""{
            log.Println("script", i, "\tcompilation failed:", r.Error[i])
        } else{
            log.Println("script", i, "\tcompilation successful", hex.EncodeToString(c))
        }
    }
}
