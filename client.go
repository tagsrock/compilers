package lllcserver

import (
	"net/http"
	"log"
	"encoding/json"
	"io/ioutil"
    "bytes"
)

var URL = "http://localhost:9999/compile"

// takes a list of lll scripts (source code, not filenames)
// returns a response object (contains list of compiled bytecodes and errors if any)
func CompileLLLClient(filenames []string) (*Response, error){
    // empty request obj
    req := Request{[]string{}}
    
    for _, f := range filenames{
        code, err  := ioutil.ReadFile(f) 
        if err != nil{                                                        
            log.Println("failed to read file", err)                           
            return nil, err
        }                                                                     
        req.Code = append(req.Code, string(code))
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

func RunClient(tocompile []string){
    r, _ := CompileLLLClient(tocompile) 
    for i, c := range r.Bytecode{
        if r.Error[i] != ""{
            log.Println("script", i, "\tcompileated failed:", r.Error[i])
        } else{
            log.Println("script", i, "\tcompilation successful", c)
        }
    }
}
