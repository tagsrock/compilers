package compilers

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/eris-ltd/eris-compilers/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

// Client cache location in eris tree
var ClientCache = path.Join(common.LllcScratchPath, "client")

// filename is either a filename or literal code
func resolveCode(filename string, literal bool) (code []byte, err error) {
	if !literal {
		code, err = ioutil.ReadFile(filename)
	} else {
		code = []byte(filename)
	}
	logger.Debugf("Code that is read =>\t%s\n", code)
	return
}

// send compile request to server or compile directly
func (c *CompileClient) compileRequest(req *Request) (respJ *Response, err error) {
	if c.config.Net {
		logger.Infof("Compiling code remotely =>\t%s\n", c.config.URL)
		respJ, err = requestResponse(req)
	} else {
		logger.Infoln("Compiling code locally.")
		respJ = compileServerCore(req)
	}
	return
}

// Takes a dir and some code, replaces all includes, checks cache, compiles, caches
func (c *CompileClient) Compile(dir string, code []byte) (*Response, error) {
	// replace includes with hash of included contents and add those contents to Includes (recursive)
	var includes = make(map[string][]byte)     // hashes to code
	var includeNames = make(map[string]string) //hashes before replace to hashes after
	var err error
	// logger.Debugf("Before parsing includes =>\n\n%s", string(code))
	code, err = c.replaceIncludes(code, dir, includes, includeNames)
	if err != nil {
		return nil, err
	}
	// logger.Debugf("After parsing includes =>\t\t%s\n\n%s", includes, string(code))

	// go through all includes, check if they have changed
	hash, cached := c.checkCached(code, includes)
	logger.Debugf("Files [Hash, Cached?] =>\t%s:%v\n", hash, cached)

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

	if respJ.Error == "" {
		// fill in cached values, cache new values
		if err := c.cacheFile(respJ.Bytecode, hash); err != nil {
			return nil, err
		}
		if err := c.cacheFile([]byte(respJ.ABI), hash+"-abi"); err != nil {
			return nil, err
		}
	}

	return respJ, nil
}

// create a new compiler for the language and compile the code
func compile(code []byte, lang, dir string) ([]byte, string, error) {
	c, err := NewCompileClient(lang)
	if err != nil {
		return nil, "", err
	}
	r, err := c.Compile(dir, code)
	if err != nil {
		return nil, "", err
	}
	b := r.Bytecode
	if r.Error != "" {
		err = fmt.Errorf(r.Error)
	} else {
		err = nil
	}
	return b, r.ABI, err
}

// Compile a file and resolve includes
func Compile(filename string) ([]byte, string, error) {
	lang, err := LangFromFile(filename)
	if err != nil {
		return nil, "", err
	}

	logger.Infof("Language to use =>\t\t%s\n", lang)

	code, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, "", err

	}
	dir := path.Dir(filename)
	return compile(code, lang, dir)
}

// Compile a literal piece of code
func CompileLiteral(code string, lang string) ([]byte, string, error) {
	return compile([]byte(code), lang, common.LllcScratchPath)
}
