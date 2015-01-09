package lllcserver

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

// must have compiler installed!
func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

var ServerTmp = ".tmp"
var null2 = CheckMakeDir(ServerTmp)

// compile request object
type Request struct {
	ScriptName string            `json:name"`
	Language   string            `json:"language"`
	Script     []byte            `json:"script"`   // source code file bytes
	Includes   map[string][]byte `json:"includes"` // filename => source code file bytes
}

// compile response object
type Response struct {
	Bytecode []byte `json:"bytecode"`
	Error    string `json:"error"`
}

// read in request body (should be pure lll code)
// compile lll, build response object, write
func CompileHandler(w http.ResponseWriter, r *http.Request) {
	resp := compileResponse(w, r)
	if resp == nil {
		return
	}
	respJ, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("failed to marshal", err)
		return
	}
	w.Write(respJ)
}

// convenience wrapper for javascript frontend
func CompileHandlerJs(w http.ResponseWriter, r *http.Request) {
	resp := compileResponse(w, r)
	if resp == nil {
		return
	}
	code := resp.Bytecode
	hexx := hex.EncodeToString(code)
	w.Write([]byte(fmt.Sprintf(`{"bytecode": "%s"}`, hexx)))
}

func compileResponse(w http.ResponseWriter, r *http.Request) *Response {
	// read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("err on read http request body", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	// unmarshall body into req struct
	req := new(Request)
	err = json.Unmarshal(body, req)
	if err != nil {
		log.Println("err on json unmarshal", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	resp := compileServerCore(req)
	return resp
}

// core compile functionality. used by the server and locally to mimic the server
func compileServerCore(req *Request) *Response {

	var name string
	lang := req.Language
	compiler := Compilers[lang]

	c := req.Script
	if c == nil || len(c) == 0 {
		name = "NULLCACHED"
	} else {
		// take sha2 of request object to get tmp filename
		hash := sha256.Sum256([]byte(c))
		filename := path.Join(ServerTmp, compiler.Ext(hex.EncodeToString(hash[:])))
		name = filename

		// lllc requires a file to read
		// check if filename already exists. if not, write
		if _, err := os.Stat(filename); err != nil {
			ioutil.WriteFile(filename, c, 0644)
		}
	}

	// loop through includes, also save to drive
	for k, v := range req.Includes {
		filename := path.Join(ServerTmp, compiler.Ext(k))
		if _, err := os.Stat(filename); err != nil {
			ioutil.WriteFile(filename, v, 0644)
		}
	}
	var resp *Response
	//compile scripts, return bytecode and error
	if name == "NULLCACHED" {

		resp = NewResponse([]byte("NULLCACHED"), "")
	} else {
		var e string
		compiled, err := CompileWrapper(name, lang)
		if err != nil {
			e = err.Error()
		} else {
			e = ""
		}
		resp = NewResponse(compiled, e)
	}

	return resp
}

// wrapper to cli
func CompileWrapper(filename string, lang string) ([]byte, error) {
	// we need to be in the same dir as the files for sake of includes
	cur, _ := os.Getwd()
	dir := path.Dir(filename)
	dir, _ = filepath.Abs(dir)
	filename = path.Base(filename)

	os.Chdir(dir)
	cmd := exec.Command(Languages[lang].Path, filename)
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

	b, err := hex.DecodeString(outstr)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func StartServer(addr string) {
	//martini.Env = martini.Prod
	srv := martini.Classic()
	// Static files
	srv.Use(martini.Static("./web"))

	srv.Post("/compile", CompileHandler)
	srv.Post("/compile2", CompileHandlerJs)

	srv.RunOnAddr(addr)

}
