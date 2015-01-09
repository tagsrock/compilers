package lllcserver

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
)

func (c *CompileClient) replaceIncludes(code []byte, dir string, includes map[string][]byte) ([]byte, error) {
	// find includes, load those as well
	r, _ := regexp.Compile(c.IncludeRegex())
	// replace all includes with hash of included lll
	//  make sure to return hashes of includes so we can cache check them too
	// do it recursively
	ret := r.ReplaceAllFunc(code, func(s []byte) []byte {
		s, err := c.includeReplacer(r, s, dir, includes)
		if err != nil {
			// panic (catch)
		}
		return s
	})
	return ret, nil
}

func (c *CompileClient) includeReplacer(r *regexp.Regexp, s []byte, dir string, included map[string][]byte) ([]byte, error) {
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
	incl_code, err = c.replaceIncludes(incl_code, this_dir, included)
	if err != nil {
		return nil, err
	}
	// compute hash
	hash := sha256.Sum256(incl_code)
	h := hex.EncodeToString(hash[:])
	included[h] = incl_code
	ret := []byte(c.IncludeReplace(h))
	return ret, nil
}

func (c *CompileClient) checkCacheIncludes(includes map[string][]byte) bool {
	cached := true
	for k, _ := range includes {
		f := path.Join(TMP, c.Ext(k))
		if _, err := os.Stat(f); err != nil {
			cached = false
			// save empty file named hash of include so we can check
			// whether includes have changed
			ioutil.WriteFile(f, []byte{}, 0644)
		}
	}
	return cached
}

func (c *CompileClient) checkCached(code []byte, includes map[string][]byte) (string, bool) {
	cachedIncludes := c.checkCacheIncludes(includes)

	// check if the main script has been cached
	hash := sha256.Sum256(code)
	hexHash := hex.EncodeToString(hash[:])
	fname := path.Join(TMP, c.Ext(hexHash))
	_, scriptErr := os.Stat(fname)

	// if an include has changed or the script has not been cached, append the code
	// else, append nil
	if !cachedIncludes || scriptErr != nil {
		return hexHash, false
	}
	return hexHash, true
}

func (c *CompileClient) cachedResponse(hash string) (*Response, error) {
	f := path.Join(TMP, c.Ext(hash))
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return NewResponse(b, ""), nil
}

func (c *CompileClient) cacheFile(b []byte, hash string) error {
	f := path.Join(TMP, c.Ext(hash))
	if b != nil {
		if err := ioutil.WriteFile(f, b, 0644); err != nil {
			return err
		}
	}
	return nil
}

func langFromFile(filename string) (string, error) {
	ext := path.Ext(filename)
	if _, ok := Languages[ext]; !ok {
		return "", UnknownLang(ext)
	}

	return "", nil
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
