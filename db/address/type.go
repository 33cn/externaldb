package address

const (
	AccountPersonage = "personage"
	AccountContract  = "contract"
)

// Address 地址
type Address struct {
	Address          string `json:"address"`
	TxCount          int64  `json:"tx_count"`
	EvmTransferCount int64  `json:"evm_transfer_count"`
	AddrType         string `json:"addr_type"`
}

func (c *Address) Key() string {
	return AddKeyPrefix(c.Address)
}
