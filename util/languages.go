package util

import (
	"github.com/eris-ltd/common/go/common"
)

const (
	SOLIDITY = "sol"
	SERPENT  = "se"
	LLL      = "lll"
)

type LangConfig struct {
	CacheDir     string   `json:"cache"`
	IncludeRegex string   `json:"regex"`
	CompileCmd   []string `json:"cmd"`
}

// todo: add indexes for where to find certain parts in submatches (quotes, filenames, etc.)
// Global variable mapping languages to their configs
var Languages = map[string]LangConfig{
	LLL: {
		CacheDir:     common.LllcScratchPath,
		IncludeRegex: `\(include "(.+?)"\)`,
		CompileCmd: []string{
			"lllc",
			"_",
		},
	},
	SERPENT: {
		CacheDir:     common.SerpScratchPath,
		IncludeRegex: `create\(("|')(.+?)("|')\)`,
		CompileCmd: []string{
			"serpent",
			"mk_contract_info_decl",
			"_",
		},
	},
	SOLIDITY: {
		CacheDir:     common.SolcScratchPath,
		IncludeRegex: `import (.+?)??("|')(.+?)("|')(as)?(.+)?;`,
		CompileCmd: []string{
			"solc",
			"--combined-json", "bin,abi",
			"_",
		},
	},
}

// individual contract items
type SolcItem struct {
	Bin string `json:"bin"`
	Abi string `json:"abi"`
}

// full solc response object
type SolcResponse struct {
	Contracts map[string]*SolcItem `mapstructure:"contracts" json:"contracts"`
	Version   string               `mapstructure:"version" json:"version"` // json encoded
}

func BlankSolcItem() *SolcItem {
	return &SolcItem{}
}

func BlankSolcResponse() *SolcResponse {
	return &SolcResponse{
		Version:   "",
		Contracts: make(map[string]*SolcItem),
	}
}
