package lllcserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

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

// New Request object from script and map of include files
func NewRequest(script []byte, includes map[string][]byte, lang string) *Request {
	if includes == nil {
		includes = make(map[string][]byte)
	}
	req := &Request{
		Script:   script,
		Includes: includes,
		Language: lang,
	}
	return req
}

// New response object from bytecode and an error
func NewResponse(bytecode []byte, err string) *Response {
	return &Response{
		Bytecode: bytecode,
		Error:    err,
	}
}

//
func requestResponse(req *Request) (*Response, error) {
	lang := req.Language
	URL := Languages[lang].URL
	logger.Infoln("lang/url for request:", lang, URL)
	// make request
	reqJ, err := json.Marshal(req)
	if err != nil {
		logger.Errorln("failed to marshal req obj", err)
		return nil, err
	}
	httpreq, err := http.NewRequest("POST", URL, bytes.NewBuffer(reqJ))
	httpreq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpreq)
	if err != nil {
		logger.Errorln("failed!", err)
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
		logger.Errorln("failed to unmarshal", err)
		return nil, err
	}
	return respJ, nil
}

func printRequest(req *Request) {
	fmt.Println("SCRIPT:", string(req.Script))
	for k, v := range req.Includes {
		fmt.Println("include:", k)
		fmt.Println("SCRIPT:", string(v))
	}
}
