package token

// Token Token
type Token struct {
	Name         string `json:"name,omitempty"`
	Symbol       string `json:"symbol,omitempty"`
	Amount       int64  `json:"amount,omitempty"`
	Owner        string `json:"owner,omitempty"`
	Creator      string `json:"creator,omitempty"`
	Introduction string `json:"introduction,omitempty"`
	// Price 发行该token愿意承担的费用
	Price int64 `json:"price,omitempty"`
	// Category: 0.普通类别 1.可增发燃烧
	Category int64 `json:"category,omitempty"`
	// Status: 0.预创建 1.完成创建 2.撤销
	Status        int64 `json:"status,omitempty"`
	PrepareHeight int64 `json:"prepare_height,omitempty"`
	CreateHeight  int64 `json:"create_height,omitempty"`
	RevokeHeight  int64 `json:"revoke_height,omitempty"`
}

// TxOption TxOption
type TxOption struct {
	//token
	Symbol       string `json:"symbol,omitempty"`
	To           string `json:"to,omitempty"`
	ExecName     string `json:"exec_name,omitempty"`
	Amount       int64  `json:"amount,omitempty"`
	Name         string `json:"name,omitempty"`
	Introduction string `json:"introduction,omitempty"`
	Total        int64  `json:"total,omitempty"`
	Price        int64  `json:"price,omitempty"`
	Owner        string `json:"owner,omitempty"`
	Category     int64  `json:"category,omitempty"`
	Note         string `json:"note,omitempty"`
}
