package compilers

// individual contract items
type SolcItem struct {
	Bin string `json:"bin"`
	Abi string `json:"abi"`
}

// full solc response object
type SolcResponse struct {
	Contracts map[string]SolcItem `json:"contracts"`
	Version   string              `json:"version"` // json encoded
}
