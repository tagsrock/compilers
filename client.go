package lllcserver

import (
	"net/http"
	"log"
    "os"
    "path"
    "regexp"
    "strings"
	"encoding/json"
	"encoding/hex"
    "crypto/sha256"
	"io/ioutil"
    "bytes"
    "fmt"
)

var URL = "http://localhost:9999/compile"
var TMP = path.Join(homeDir(), ".lllc")
var null  = CheckMakeDir(TMP)

func replaceIncludes(code []byte, dir string, req Request, included map[string][]byte) ([]byte, map[string]bool){
    // find includes, load those as well
    r, _ :=  regexp.Compile(`\(include "(.+?)"\)`)
    // replace all includes with hash of included lll
    //  make sure to return hashes of includes so we can cache check them too
    includeHashes := make(map[string]bool)
    // do it recursively
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
        this_dir := path.Dir(p)
        incl_code, these_incl_hashes := replaceIncludes(incl_code, this_dir, req, included)
        for ih, _ := range these_incl_hashes{
            includeHashes[ih] = true
        }
        // compute hash
        hash := sha256.Sum256(incl_code)
        h := hex.EncodeToString(hash[:])
        includeHashes[h] = true
        req.Includes[h] = incl_code
        ret := []byte(`(include "`+h+`.lll")`)
        included[name] = ret
        return ret
    })
    return ret, includeHashes
}

// takes a list of lll scripts 
// returns a response object (contains list of compiled bytecodes and errors if any)
func CompileLLLClient(filenames []string, literal bool) (*Response, error){
    // cached bytecode
    hashmap := make(map[int][]byte) // map indices to hashes

    // empty request obj
    req := Request{
        Scripts: [][]byte{},
        Includes: make(map[string][]byte),
    }
   
    included := make(map[string][]byte)

    var err error
    for i, f := range filenames{
        var code []byte
        if !literal{
            code, err = ioutil.ReadFile(f) 
            if err != nil{
                log.Println("failed to read file", err)
                return nil, err
            }
        } else{
            code = []byte(f)
        }
        dir := path.Dir(f)
        // replace includes with hash of included contents and add those contents to Includes (recursive)
        code, includes := replaceIncludes(code, dir, req, included)

        // go through all includes, check if they have changed
        cached := true
        for k, _ := range includes{
            f := path.Join(TMP, k+".lll")
            if _, err := os.Stat(f); err != nil{
                cached = false
                // save empty file named hash of include so we can check
                // whether includes have changed
                ioutil.WriteFile(f, []byte{}, 0644)
            }
        }

        // check if the main script has been cached
        hash := sha256.Sum256(code)
        hashmap[i] = hash[:]
        filename := path.Join(TMP, hex.EncodeToString(hash[:])+".lll")
        _, scriptErr := os.Stat(filename)

        // if an include has changed or the script has not been cached, append the code
        // else, append nil
        if !cached || scriptErr != nil{
            req.Scripts = append(req.Scripts, code)
        } else{
            req.Scripts = append(req.Scripts, nil)
        }
    }
    /*
    for _, v := range req.Scripts{
        fmt.Println("code;", string(v))
    }
    for _, v := range req.Includes{
        fmt.Println("req includes;", string(v))
    }*/

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

    // check if we should compile locally instead of firing off to server
    if !strings.Contains(URL, "http://") && !strings.Contains(URL, "NETCALL"){
        fmt.Println("compiling locally...")
        respJ = compileServerCore(req)
    } else {
        fmt.Println("compiling remotely...")
        if strings.Contains(URL, "NETCALL"){
            URL = "http://lllc.projectdouglas.org/compile"
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
            ioutil.WriteFile(f, b, 0644)
        }
    }
    //fmt.Println(respJ.Bytecode[0])

    return &respJ, nil
}  

// compile just one file
// but resolve "includes"
func Compile(filename string, literal bool) ([]byte, error){
    r, err := CompileLLLClient([]string{filename}, literal)
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

func RunClient(tocompile []string, literal bool){
    _,  err := CompileLLLClient(tocompile, literal) 
    if err != nil{
        fmt.Println("shucks", err)
        os.Exit(0)
    }
    /*
    for i, c := range r.Bytecode{
        if r.Error[i] != ""{
            log.Println("script", i, "\tcompilation failed:", r.Error[i])
        } else{
            log.Println("script", i, "\tcompilation successful", hex.EncodeToString(c))
        }
    }*/
}


func CheckMakeDir(dir string) int{
   _, err := os.Stat(dir)
   if err != nil{
       err := os.Mkdir(dir, 0777)  //wtf!
       if err != nil{
            fmt.Println("Could not make directory. Exiting", err)
            os.Exit(0)
       }
   }
   return 0
}
