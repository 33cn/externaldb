package transaction

// Block info
type Block struct {
	Height    int64  `json:"height"`
	BlockTime int64  `json:"block_time"`
	BlockHash string `json:"block_hash"`
}

// Asset asset
type Asset struct {
	Exec   string `json:"exec"`
	Symbol string `json:"symbol"`
	Amount int64  `json:"amount"`
}

// AddrRecord 投票账户和打包账户的记录
type AddrRecord struct {
	VoterAddr []string `json:"voter_addr"`
	MakerAddr []string `json:"maker_addr"`
}

// Transaction 记录通用的项
type Transaction struct {
	// as key
	HeightIndex int64 `json:"height_index"`
	*Block
	Success    bool        `json:"success"`
	Index      int64       `json:"index"`
	Hash       string      `json:"hash"`
	From       string      `json:"from"`
	To         string      `json:"to"`
	Execer     string      `json:"execer"`
	Amount     int64       `json:"amount"`
	Fee        int64       `json:"fee"`
	ActionName string      `json:"action_name"`
	GroupCount int64       `json:"group_count"`
	IsWithdraw bool        `json:"is_withdraw"`
	Options    interface{} `json:"options"`
	Assets     []Asset     `json:"assets"`
	Next       string      `json:"next"`
	IsPara     bool        `json:"is_para"`
	*AddrRecord
}
