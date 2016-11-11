package util

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	//log "github.com/eris-ltd/eris-logger"
)

// check/cache all includes, hash the code, return whether or not there was a full cache hit
func CheckCached(includes map[string]*IncludedFiles, lang string) bool {
	cached := true
	for name, metadata := range includes {
		hashPath := path.Join(Languages[lang].CacheDir, name)
		if _, scriptErr := os.Stat(hashPath); os.IsNotExist(scriptErr) {
			cached = false
			break
		}
		for _, object := range metadata.ObjectNames {
			objectAbi := path.Join(hashPath, object+"-abi")
			objectBin := path.Join(hashPath, object+"-bin")
			if _, abiErr := os.Stat(objectAbi); abiErr != nil {
				cached = false
				break
			}
			if _, binErr := os.Stat(objectBin); binErr != nil {
				cached = false
				break
			}
		}
		if cached == false {
			break
		}
	}

	return cached
}

// return cached byte code as a response
func CachedResponse(includes map[string]*IncludedFiles, lang string) (*Response, error) {

	var resp *Response
	var respItemArray []ResponseItem
	for name, metadata := range includes {
		dir := path.Join(Languages[lang].CacheDir, name)
		for _, object := range metadata.ObjectNames {
			bin, err := ioutil.ReadFile(path.Join(dir, object+"-bin"))
			if err != nil {
				return nil, err
			}
			abi, err := ioutil.ReadFile(path.Join(dir, object+"-abi"))
			if err != nil {
				return nil, err
			}

			respItem := ResponseItem{
				Objectname: object,
				Bytecode:   string(bin),
				ABI:        string(abi),
			}
			respItemArray = append(respItemArray, respItem)
		}
	}
	resp = &Response{
		Objects: respItemArray,
		Error:   "",
	}

	return resp, nil
}

func (resp Response) CacheNewResponse(req Request) {
	objects := resp.Objects
	//log.Debug(objects)
	cacheLocation := Languages[req.Language].CacheDir
	cur, _ := os.Getwd()
	os.Chdir(cacheLocation)
	defer func() {
		os.Chdir(cur)
	}()
	for fileDir, metadata := range req.Includes {
		dir := path.Join(cacheLocation, strings.TrimRight(fileDir, "."+req.Language))
		os.MkdirAll(dir, 0700)
		objectNames := metadata.ObjectNames
		for _, name := range objectNames {
			for _, object := range objects {
				if object.Objectname == name {
					//log.WithField("=>", resp.Objects).Debug("Response objects over the loop")
					CacheResult(dir, object.Objectname, object.Bytecode, object.ABI)
					break
				}
			}
		}
	}
}

// cache ABI and Binary to
func CacheResult(cacheLocation string, object string, binary string, abi string) error {
	os.Chdir(cacheLocation)
	ioutil.WriteFile(object+"-bin", []byte(binary), 0644)
	ioutil.WriteFile(object+"-abi", []byte(abi), 0644)
	return nil
}
