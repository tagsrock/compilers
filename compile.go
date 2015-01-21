package lllcserver

import (
	"encoding/json"
	"fmt"
	"github.com/eris-ltd/epm-go/utils"
	"io/ioutil"
	"os"
	"path"
)

var DefaultUrl = "http://162.218.65.211:8090/compile"

// A compiler interface adds filename extensions, replaces includes, and executes a compiler program
type Compiler interface {
	Lang() string
	Ext(h string) string
	IncludeRegex() string                      // regular expression string
	IncludeReplace(h string) string            // new include stmt
	CompileCmd(file string) (string, []string) // command line string to execute
}

// language configuration struct
type LangConfig struct {
	URL        string   `json:"url"`
	Path       string   `json:"path"`
	Net        bool     `json:"net"`
	Extensions []string `json:"extensions"`
}

// global variable mapping languages to their configs
var Languages = map[string]LangConfig{
	"lll": LangConfig{
		URL:        DefaultUrl,
		Path:       path.Join(homeDir(), "cpp-ethereum/build/lllc/lllc"),
		Net:        true,
		Extensions: []string{"lll", "def"},
	},

	"se": LangConfig{
		URL:        DefaultUrl,
		Path:       "/usr/local/bin/serpent",
		Net:        true,
		Extensions: []string{"se"},
	},
}

func init() {
	utils.InitDataDir(ClientCache)
	utils.InitDataDir(ServerCache)

	// read language config from  ~/.decerver
	// if it doesnt exist yet, do nothing
	if _, err := os.Stat(utils.Languages); err != nil {
		return
	}
	f := path.Join(utils.Languages, "config.json")
	err := checkConfig(f)
	if err != nil {
		logger.Errorln(err)
		logger.Errorln("resorting to default language settings")
		return
	}

}

func checkConfig(f string) error {
	if _, err := os.Stat(f); err != nil {
		err := utils.WriteJson(&Languages, f)
		if err != nil {
			return fmt.Errorf("Could not write default config to %s: %s", f, err.Error())
		}
	}

	b, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	c := new(map[string]LangConfig)
	err = json.Unmarshal(b, c)
	if err != nil {
		return err
	}

	Languages = *c
	return nil
}

// Set the languages compiler path
func SetLanguagePath(lang, path string) error {
	l, ok := Languages[lang]
	if !ok {
		return UnknownLang(lang)
	}
	l.Path = path
	Languages[lang] = l
	return nil
}

// Set the languages url
func SetLanguageURL(lang, url string) error {
	l, ok := Languages[lang]
	if !ok {
		return UnknownLang(lang)
	}
	l.URL = url
	Languages[lang] = l
	return nil
}

// Set whether the language should use the remote server or compile locally
func SetLanguageNet(lang string, net bool) error {
	l, ok := Languages[lang]
	if !ok {
		return UnknownLang(lang)
	}
	l.Net = net
	Languages[lang] = l
	return nil

}

// gloval variable mapping languages to their compiler interfaces
var Compilers = assembleCompilers()

func assembleCompilers() map[string]Compiler {
	compilers := make(map[string]Compiler)
	for l, _ := range Languages {
		compilers[l], _ = NewCompiler(l)
	}
	return compilers
}

// Main client struct to wrap a compiler interface and its configuration data
type CompileClient struct {
	c    Compiler
	url  string
	path string
	net  bool
}

// Return the language name
func (c *CompileClient) Lang() string {
	return c.c.Lang()
}

// Return the language's main extension
func (c *CompileClient) Ext(h string) string {
	return c.c.Ext(h)
}

// Return the regex string to match include statements
func (c *CompileClient) IncludeRegex() string {
	return c.c.IncludeRegex()
}

// Return the string to replace matched regex expressions
func (c *CompileClient) IncludeReplace(h string) string {
	return c.c.IncludeReplace(h)
}

// Unknown language error
func UnknownLang(lang string) error {
	return fmt.Errorf("Unknown language %s", lang)
}

// Create a new compile client
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

// Create a new compiler interface for a given language
func NewCompiler(lang string) (c Compiler, err error) {
	switch lang {
	case "lll":
		c = NewLLL()
	case "se", "serpent":
		c = NewSerpent()
	case "sol", "solidity":
		err = UnknownLang(lang)
	}
	return
}

// New LLL compiler
func NewLLL() Compiler {
	return &LLLCompiler{Languages["lll"].Path}
}

type LLLCompiler struct {
	path string
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

func (c *LLLCompiler) CompileCmd(f string) (string, []string) {
	return c.path, []string{f}
}

// New Serpent compiler
func NewSerpent() Compiler {
	return &SerpentCompiler{Languages["se"].Path}
}

type SerpentCompiler struct {
	path string
}

func (c *SerpentCompiler) Lang() string {
	return "se"
}

func (c *SerpentCompiler) Ext(h string) string {
	return h + "." + "se"
}

// TODO
func (c *SerpentCompiler) IncludeReplace(h string) string {
	return `(include "` + h + `.lll")`
}

// TODO
func (c *SerpentCompiler) IncludeRegex() string {
	return `\(include "(.+?)"\)`
}

func (c *SerpentCompiler) CompileCmd(f string) (string, []string) {
	return c.path, []string{"compile", f}
}
