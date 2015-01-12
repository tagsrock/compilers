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
	"strings"
)

// Find all matches to the include regex
// Replace filenames with hashes
func (c *CompileClient) replaceIncludes(code []byte, dir string, includes map[string][]byte) ([]byte, error) {
	// find includes, load those as well
	r, _ := regexp.Compile(c.IncludeRegex())
	// replace all includes with hash of included lll
	//  make sure to return hashes of includes so we can cache check them too
	// do it recursively
	ret := r.ReplaceAllFunc(code, func(s []byte) []byte {
		s, err := c.includeReplacer(r, s, dir, includes)
		if err != nil {
			fmt.Println("ERR!:", err)
			// panic (catch)
		}
		return s
	})
	return ret, nil
}

// read the included file, hash it; if we already have it, return include replacement
// if we don't, run replaceIncludes on it (recursive)
// modifies the "includes" map
func (c *CompileClient) includeReplacer(r *regexp.Regexp, s []byte, dir string, included map[string][]byte) ([]byte, error) {
	m := r.FindSubmatch(s)
	match := m[1]
	// load the file
	p := path.Join(dir, string(match))
	incl_code, err := ioutil.ReadFile(p)
	if err != nil {
		logger.Errorln("failed to read include file", err)
		return nil, fmt.Errorf("Failed to read include file: %s", err.Error())
	}

	// compute hash
	hash := sha256.Sum256(incl_code)
	h := hex.EncodeToString(hash[:])
	ret := []byte(c.IncludeReplace(h))
	// if we've already loaded this, return the replacement
	// and move on
	if _, ok := included[h]; ok {
		return ret, nil
	}

	// recursively replace the includes for this file
	this_dir := path.Dir(p)
	incl_code, err = c.replaceIncludes(incl_code, this_dir, included)
	if err != nil {
		return nil, err
	}
	included[h] = incl_code
	return ret, nil
}

func (c *CompileClient) checkCacheIncludes(includes map[string][]byte) bool {
	cached := true
	for k, _ := range includes {
		f := path.Join(ClientCache, c.Ext(k))
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
	fname := path.Join(ClientCache, c.Ext(hexHash))
	_, scriptErr := os.Stat(fname)

	// if an include has changed or the script has not been cached, append the code
	// else, append nil
	if !cachedIncludes || scriptErr != nil {
		return hexHash, false
	}
	return hexHash, true
}

func (c *CompileClient) cachedResponse(hash string) (*Response, error) {
	f := path.Join(ClientCache, c.Ext(hash))
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return NewResponse(b, ""), nil
}

func (c *CompileClient) cacheFile(b []byte, hash string) error {
	f := path.Join(ClientCache, c.Ext(hash))
	if b != nil {
		if err := ioutil.WriteFile(f, b, 0644); err != nil {
			return err
		}
	}
	return nil
}

func LangFromFile(filename string) (string, error) {
	ext := path.Ext(filename)
	ext = strings.Trim(ext, ".")
	if _, ok := Languages[ext]; ok {
		return ext, nil
	}
	for l, lc := range Languages {
		for _, e := range lc.Extensions {
			if ext == e {
				return l, nil
			}
		}
	}
	return "", UnknownLang(ext)
}

func isLiteral(f, lang string) bool {
	if strings.HasSuffix(f, Compilers[lang].Ext("")) {
		return false
	}

	for _, lc := range Languages {
		for _, e := range lc.Extensions {
			if strings.HasSuffix(f, e) {
				return false
			}
		}
	}
	return true
}

func ClearCaches() error {
	if err := ClearServerCache(); err != nil {
		return err
	}
	return ClearClientCache()
}

func clearDir(dir string) error {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range fs {
		n := f.Name()
		if err := os.Remove(path.Join(dir, n)); err != nil {
			return err
		}
	}
	return nil
}

func ClearServerCache() error {
	return clearDir(ServerCache)
}

func ClearClientCache() error {
	return clearDir(ClientCache)
}

func CheckMakeDir(dir string) int {
	_, err := os.Stat(dir)
	if err != nil {
		err := os.Mkdir(dir, 0777) //wtf!
		if err != nil {
			logger.Errorln("Could not make directory. Exiting", err)
			os.Exit(0)
		}
	}
	return 0
}

type Logger struct {
}

func (l *Logger) Errorln(s ...interface{}) {
	if DebugMode > 0 {
		log.Println(s...)
	}
}

func (l *Logger) Warnln(s ...interface{}) {
	if DebugMode > 1 {
		log.Println(s...)
	}
}

func (l *Logger) Infoln(s ...interface{}) {
	if DebugMode > 2 {
		log.Println(s...)
	}
}

func (l *Logger) Debugln(s ...interface{}) {
	if DebugMode > 3 {
		log.Println(s...)
	}
}
