package lllcserver

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"
)

var TMP = path.Join(homeDir(), ".lllc")
var null = CheckMakeDir(TMP)

// filename is either a filename or literal code
func resolveCode(filename string, literal bool) (code []byte, err error) {
	if !literal {
		code, err = ioutil.ReadFile(filename)
	} else {
		code = []byte(filename)
	}
	return
}

//
func (c *CompileClient) compileRequest(req *Request) (respJ *Response, err error) {
	if c.net {
		fmt.Println("compiling remotely...")
		respJ, err = requestResponse(req)
	} else {
		fmt.Println("compiling locally...")
		respJ = compileServerCore(req)
	}
	return
}

// Takes a dir and some code, replaces all includes, checks cache, compiles, caches
func (c *CompileClient) Compile(dir string, code []byte) (*Response, error) {
	// replace includes with hash of included contents and add those contents to Includes (recursive)
	var includes = make(map[string][]byte)
	var err error
	code, err = c.replaceIncludes(code, dir, includes)
	if err != nil {
		return nil, err
	}

	// go through all includes, check if they have changed
	hash, cached := c.checkCached(code, includes)

	// if everything is cached, no need for request
	if cached {
		return c.cachedResponse(hash)
	}
	req := NewRequest(code, includes, c.Lang())

	// response struct (returned)
	respJ, err := c.compileRequest(req)
	if err != nil {
		return nil, err
	}
	// fill in cached values, cache new values
	if err := c.cacheFile(respJ.Bytecode, hash); err != nil {
		return nil, err
	}

	return respJ, nil
}

// create a new compiler for the language and compile the code
func compile(code []byte, lang, dir string) ([]byte, error) {
	c, err := NewCompileClient(lang)
	if err != nil {
		return nil, err
	}
	r, err := c.Compile(dir, code)
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

// Compile a file and resolve includes
func Compile(filename string) ([]byte, error) {
	lang, err := langFromFile(filename)
	if err != nil {
		return nil, err
	}

	fmt.Println("lang:", lang)

	literal := strings.HasSuffix(filename, Compilers[lang].Ext(""))
	code, err := resolveCode(filename, literal)
	if err != nil {
		return nil, err

	}
	dir := path.Dir(filename)
	return compile(code, lang, dir)
}

// Compile a literal piece of code
func CompileLiteral(code []byte, lang string) ([]byte, error) {
	return compile(code, lang, "something-intelligent")
}
