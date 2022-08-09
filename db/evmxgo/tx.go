package evmxgo

// Evmxgo Evmxgo
type Evmxgo struct {
	Symbol       string `json:"symbol,omitempty"`
	Amount       int64  `json:"amount,omitempty"`
	Introduction string `json:"introduction,omitempty"`
}

// TxOption TxOption
type TxOption struct {
	// evmxgo
	Symbol       string `json:"symbol,omitempty"`
	Address      string `json:"address,omitempty"`
	To           string `json:"to,omitempty"`
	ExecName     string `json:"exec_name,omitempty"`
	Amount       int64  `json:"amount,omitempty"`
	Name         string `json:"name,omitempty"`
	Introduction string `json:"introduction,omitempty"`
	Total        int64  `json:"total,omitempty"`
	Note         string `json:"note,omitempty"`
}
