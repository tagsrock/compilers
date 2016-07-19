package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/eris-ltd/eris-logger"
)

var (
	ext string
)

// Find all matches to the include regex
// Replace filenames with hashes
func (c *Compiler) replaceIncludes(code []byte, dir string, includes map[string]*IncludedFiles) ([]byte, error) {
	// find includes, load those as well
	regexPattern := c.IncludeRegex()
	var regExpression *regexp.Regexp
	var err error
	if regExpression, err = regexp.Compile(regexPattern); err != nil {
		return nil, err
	}
	OriginObjectNames, err := extractObjectNames(code)
	if err != nil {
		return nil, err
	}
	// replace all includes with hash of included imports
	// make sure to return hashes of includes so we can cache check them too
	// do it recursively
	code = regExpression.ReplaceAllFunc(code, func(s []byte) []byte {
		log.WithField("=>", string(s)).Debug("Include Replacer result")
		s, err := c.includeReplacer(regExpression, s, dir, includes)
		if err != nil {
			log.Error("ERR!:", err)
		}
		return s
	})

	originHash := sha256.Sum256(code)
	origin := hex.EncodeToString(originHash[:])
	origin += "." + c.lang
	
	includeFile := &IncludedFiles{
		ObjectNames: OriginObjectNames,
		Script:      code,
	}

	includes[origin] = includeFile

	return code, nil
}

// read the included file, hash it; if we already have it, return include replacement
// if we don't, run replaceIncludes on it (recursive)
// modifies the "includes" map
func (c *Compiler) includeReplacer(r *regexp.Regexp, originCode []byte, dir string, included map[string]*IncludedFiles) ([]byte, error) {
	// regex look for strings that would match the import statement
	m := r.FindStringSubmatch(string(originCode))
	match := m[3]
	log.WithField("=>", match).Debug("Match")
	// load the file
	newFilePath := path.Join(dir, match)
	incl_code, err := ioutil.ReadFile(newFilePath)
	if err != nil {
		log.Errorln("failed to read include file", err)
		return nil, fmt.Errorf("Failed to read include file: %s", err.Error())
	}

	// take hash before replacing includes to see if we've already parsed this file
	hash := sha256.Sum256(incl_code)
	includeHash := hex.EncodeToString(hash[:])
	log.Debug("This is hash of included code", includeHash)
	if _, ok := included[includeHash]; ok {
		//then replace
		fullReplacement := strings.SplitAfter(m[0], m[2])
		fullReplacement[1] = includeHash + "." + c.lang + "\""
		ret := strings.Join(fullReplacement, "")
		return []byte(ret), nil
	}

	// recursively replace the includes for this file
	this_dir := path.Dir(newFilePath)
	incl_code, err = c.replaceIncludes(incl_code, this_dir, included)
	if err != nil {
		return nil, err
	}

	// compute hash
	hash = sha256.Sum256(incl_code)
	h := hex.EncodeToString(hash[:])

	//Starting with full regex string,
	//Split strings from the quotation mark and then,
	//assuming 3 array cells, replace the middle one.
	fullReplacement := strings.SplitAfter(m[0], m[2])
	fullReplacement[1] = h + "." + c.lang + m[4]
	ret := []byte(strings.Join(fullReplacement, ""))
	return ret, nil
}

// clear a directory of its contents
func ClearCache(dir string) error {
	d, err := os.Open(dir)
    if err != nil {
        return err
    }
    defer d.Close()
    names, err := d.Readdirnames(-1)
    if err != nil {
        return err
    }
    for _, name := range names {
        err = os.RemoveAll(filepath.Join(dir, name))
        if err != nil {
            return err
        }
    }
	return nil
}

// Get language from filename extension
func LangFromFile(filename string) (string, error) {
	ext := path.Ext(filename)
	ext = strings.Trim(ext, ".")
	if _, ok := Languages[ext]; ok {
		return ext, nil
	}
	return "", UnknownLang(ext)
}

// Return the regex string to match include statements
func (c *Compiler) IncludeRegex() string {
	return c.config.IncludeRegex
}

func extractObjectNames(script []byte) ([]string, error) {
	regExpression, err := regexp.Compile("(contract|library) (.+?) (is)?(.+?)?({)")
	if err != nil {
		return nil, err
	}
	objectNamesList := regExpression.FindAllSubmatch(script, -1)
	var objects []string
	for _, objectNames := range objectNamesList {
		objects = append(objects, string(objectNames[2]))
	}
	return objects, nil
}

// Unknown language error
func UnknownLang(lang string) error {
	return fmt.Errorf("Unknown language %s", lang)
}

func createTemporaryFile(name string, code []byte) (*os.File, error) {
	file, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	_, err = file.Write(code)
	if err != nil {
		return nil, err
	}
	if err = file.Close(); err != nil {
		return nil, err
	}
	return file, nil
}

func PrintResponse(resp Response) {
	for _, r := range resp.Objects {
		log.WithFields(log.Fields{
			"name": r.Objectname,
			"bin":  r.Bytecode,
			"abi":  r.ABI,
		}).Warn("Response")
	}
}
