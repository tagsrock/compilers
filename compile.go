package lllcserver

import (
	"fmt"
	"path"
)

// TODO: init
func init() {
	// read language config from  ~/.decerver
}

var Languages = map[string]LangConfig{
	"lll": LangConfig{
		URL:  "http://localhost:9999/compile",
		Path: path.Join(homeDir(), "cpp-ethereum/build/lllc/lllc"),
		Net:  true,
	},
}

func SetLanguagePath(lang, path string) error {
	l, ok := Languages[lang]
	if !ok {
		return UnknownLang(lang)
	}
	l.Path = path
	Languages[lang] = l
	return nil
}

func SetLanguageURL(lang, url string) error {
	l, ok := Languages[lang]
	if !ok {
		return UnknownLang(lang)
	}
	l.URL = url
	Languages[lang] = l
	return nil
}

func SetLanguageNet(lang string, net bool) error {
	l, ok := Languages[lang]
	if !ok {
		return UnknownLang(lang)
	}
	l.Net = net
	Languages[lang] = l
	return nil

}

var Compilers = assembleCompilers()

func assembleCompilers() map[string]Compiler {
	compilers := make(map[string]Compiler)
	for l, _ := range Languages {
		compilers[l], _ = NewCompiler(l)
	}
	return compilers
}

// A compiler interface adds extensions and replaces includes
type Compiler interface {
	Lang() string
	Ext(h string) string
	IncludeRegex() string           // regular expression string
	IncludeReplace(h string) string // new include stmt
}

type CompileClient struct {
	c    Compiler
	url  string
	path string
	net  bool
}

func (c *CompileClient) Lang() string {
	return c.c.Lang()
}

func (c *CompileClient) Ext(h string) string {
	return c.c.Ext(h)
}

func (c *CompileClient) IncludeRegex() string {
	return c.c.IncludeRegex()
}
func (c *CompileClient) IncludeReplace(h string) string {
	return c.c.IncludeReplace(h)
}

func UnknownLang(lang string) error {
	return fmt.Errorf("Unknown language %s", lang)
}

func NewCompileClient(lang string) (*CompileClient, error) {
	compiler, err := NewCompiler(lang)
	if err != nil {
		return nil, err
	}
	l := Languages[lang]
	cc := &CompileClient{
		c:    compiler,
		url:  l.URL,
		path: l.Path,
		net:  l.Net,
	}
	return cc, nil
}

func NewCompiler(lang string) (c Compiler, err error) {
	switch lang {
	case "lll":
		c = NewLLL()
	case "se", "serpent":
		err = UnknownLang(lang)
	case "sol", "solidity":
		err = UnknownLang(lang)
	}
	return
}

var LangConfigs map[string]LangConfig

type LangConfig struct {
	URL  string `json:"url"`
	Path string `json:"path"`
	Net  bool   `json:"net"`
}

func NewLLL() Compiler {
	return &LLLCompiler{}
}

type LLLCompiler struct {
}

func (c *LLLCompiler) Lang() string {
	return "lll"
}

func (c *LLLCompiler) Ext(h string) string {
	return h + "." + "lll"
}

func (c *LLLCompiler) IncludeReplace(h string) string {
	return `(include "` + h + `.lll")`
}

func (c *LLLCompiler) IncludeRegex() string {
	return `\(include "(.+?)"\)`
}
