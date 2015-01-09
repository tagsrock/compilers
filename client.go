package lllcserver

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
)

var URL = "http://localhost:9999/compile"
var TMP = path.Join(homeDir(), ".lllc")
var null = CheckMakeDir(TMP)

func replaceIncludes(code []byte, dir string, includes map[string][]byte) ([]byte, error) {
	// find includes, load those as well
	r, _ := regexp.Compile(`\(include "(.+?)"\)`)
	// replace all includes with hash of included lll
	//  make sure to return hashes of includes so we can cache check them too
	// do it recursively
	ret := r.ReplaceAllFunc(code, func(s []byte) []byte {
		s, err := includeReplacer(r, s, dir, includes)
		if err != nil {
			// panic (catch)
		}
		return s
	})
	return ret, nil
}

func includeReplacer(r *regexp.Regexp, s []byte, dir string, included map[string][]byte) ([]byte, error) {
	m := r.FindSubmatch(s)
	match := m[1]
	name := path.Base(string(match))
	// if we've already loaded this, move on
	if v, ok := included[name]; ok {
		return v, nil
	}
	// load the file
	p := path.Join(dir, string(match))
	incl_code, err := ioutil.ReadFile(p)
	if err != nil {
		log.Println("failed to read include file", err)
		return nil, fmt.Errorf("Failed to read include file: %s", err.Error())
	}
	this_dir := path.Dir(p)
	incl_code, err = replaceIncludes(incl_code, this_dir, included)
	if err != nil {
		return nil, err
	}
	// compute hash
	hash := sha256.Sum256(incl_code)
	h := hex.EncodeToString(hash[:])
	included[h] = incl_code
	ret := []byte(`(include "` + h + `.lll")`)
	return ret, nil
}

func checkCacheIncludes(includes map[string][]byte) bool {
	cached := true
	for k, _ := range includes {
		f := path.Join(TMP, k+".lll")
		if _, err := os.Stat(f); err != nil {
			cached = false
			// save empty file named hash of include so we can check
			// whether includes have changed
			ioutil.WriteFile(f, []byte{}, 0644)
		}
	}
	return cached
}

func checkCached(code []byte, includes map[string][]byte) (string, bool) {
	cachedIncludes := checkCacheIncludes(includes)

	// check if the main script has been cached
	hash := sha256.Sum256(code)
	hexHash := hex.EncodeToString(hash[:])
	fname := path.Join(TMP, hexHash+".lll")
	_, scriptErr := os.Stat(fname)

	// if an include has changed or the script has not been cached, append the code
	// else, append nil
	if !cachedIncludes || scriptErr != nil {
		return hexHash, false
	}
	return hexHash, true
}

func NewRequest(script []byte, includes map[string][]byte) *Request {
	if includes == nil {
		includes = make(map[string][]byte)
	}
	req := &Request{
		Script:   script,
		Includes: includes,
	}
	return req
}

func cachedResponse(hash string) (*Response, error) {
	// fill in cached values,
	f := path.Join(TMP, hash+".lll")
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return NewResponse(b, ""), nil
}

func NewResponse(bytecode []byte, err string) *Response {
	return &Response{
		Bytecode: bytecode,
		Error:    err,
	}
}

func resolveCode(filename string, literal bool) (code []byte, err error) {
	if !literal {
		code, err = ioutil.ReadFile(filename)
	} else {
		code = []byte(filename)
	}
	return
}

func requestResponse(req *Request) (*Response, error) {
	if strings.Contains(URL, "NETCALL") {
		URL = "http://lllc.projectdouglas.org/compile"
	}

	// make request
	reqJ, err := json.Marshal(req)
	if err != nil {
		log.Println("failed to marshal req obj", err)
		return nil, err
	}
	httpreq, err := http.NewRequest("POST", URL, bytes.NewBuffer(reqJ))
	httpreq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpreq)
	if err != nil {
		log.Println("failed!", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 300 {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	respJ := new(Response)
	// read in response body
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, respJ)
	if err != nil {
		fmt.Println("failed to unmarshal", err)
		return nil, err
	}
	return respJ, nil
}

// takes a list of lll scripts
// returns a response object (contains list of compiled bytecodes and errors if any)
func CompileLLLClient(filename string, literal bool) (*Response, error) {
	code, err := resolveCode(filename, literal)
	if err != nil {
		return nil, err

	}
	dir := path.Dir(filename)
	// replace includes with hash of included contents and add those contents to Includes (recursive)
	var includes = make(map[string][]byte)
	code, err = replaceIncludes(code, dir, includes)
	if err != nil {
		return nil, err
	}

	// go through all includes, check if they have changed
	hash, cached := checkCached(code, includes)

	// if everything is cached, no need for request
	if cached {
		return cachedResponse(hash)
	}
	req := NewRequest(code, includes)

	// response struct (returned)
	respJ, err := compileRequest(req)
	if err != nil {
		return nil, err
	}
	// fill in cached values, cache new values
	if err := fillSaveCache(respJ, hash); err != nil {
		return nil, err
	}

	return respJ, nil
}

func compileRequest(req *Request) (respJ *Response, err error) {
	// check if we should compile locally instead of firing off to server
	if !strings.Contains(URL, "http://") && !strings.Contains(URL, "NETCALL") {
		fmt.Println("compiling locally...")
		respJ = compileServerCore(req)
	} else {
		fmt.Println("compiling remotely...")
		if respJ, err = requestResponse(req); err != nil {
			return
		}
	}
	return
}

func fillSaveCache(respJ *Response, hash string) error {
	var err error
	b := respJ.Bytecode
	f := path.Join(TMP, hash+".lll")
	if string(b) == "NULLCACHED" {
		respJ.Bytecode, err = ioutil.ReadFile(f)
		if err != nil {
			fmt.Println("read fil:", err)
			return err
		}
	} else if b != nil {
		if err := ioutil.WriteFile(f, b, 0644); err != nil {
			return err
		}
	}
	return nil
}

// compile just one file
// but resolve "includes"
func Compile(filename string, literal bool) ([]byte, error) {
	r, err := CompileLLLClient(filename, literal)
	if err != nil {
		return nil, err
	}
	b := r.Bytecode
	if r.Error != "" {
		err = fmt.Errorf(r.Error)
	} else {
		err = nil
	}
	return b, err
}

func RunClient(tocompile string, literal bool) {
	_, err := CompileLLLClient(tocompile, literal)
	if err != nil {
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

func CheckMakeDir(dir string) int {
	_, err := os.Stat(dir)
	if err != nil {
		err := os.Mkdir(dir, 0777) //wtf!
		if err != nil {
			fmt.Println("Could not make directory. Exiting", err)
			os.Exit(0)
		}
	}
	return 0
}
