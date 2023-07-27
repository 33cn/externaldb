package evm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/33cn/externaldb/util"
	"github.com/33cn/go-kit/convert"
)

const EmptyAddr = "1111111111111111111114oLvT2"
const EmptyAddrEth = "0x0000000000000000000000000000000000000000"

var ErrNoEvents = errors.New("no events")
var ErrNoEvent = errors.New("no event")

// EVM evm执行记录
type EVM map[string]interface{}

// Key for token index id
func (e EVM) Key() string {
	return fmt.Sprintf("evm-%s", e["evm_tx_hash"])
}

// GetEvent return event
func (e EVM) GetEvent(name string) (map[string]interface{}, error) {
	events, err := e.GetEvents()
	if err != nil {
		log.Error("GetEvent e.GetEvents", "err", err)
		return nil, err
	}
	if m, ok := events[name].(map[string]interface{}); ok {
		return m, nil
	}
	// 首字母转大写再试一次
	name = strings.ToUpper(name[:1]) + name[1:]
	if m, ok := events[name].(map[string]interface{}); ok {
		return m, nil
	}
	return nil, ErrNoEvent
}

// GetEvents return events
func (e EVM) GetEvents() (map[string]interface{}, error) {
	if m, ok := e["evm_events_map"].(map[string]interface{}); ok {
		return m, nil
	}
	events := make(map[string]interface{})
	eventsBuf, ok := e["evm_events"].(string)
	if !ok {
		return nil, ErrNoEvents
	}
	der := json.NewDecoder(bytes.NewReader([]byte(eventsBuf)))
	// 用number类型替换float64
	der.UseNumber()
	if err := der.Decode(&events); err != nil {
		return nil, err
	}
	e["evm_events_map"] = events
	return events, nil
}

func (e EVM) ToJSON() ([]byte, error) {
	delete(e, "evm_events_map")
	return json.Marshal(e)
}

type Token struct {
	TokenID      string `json:"token_id"`
	Owner        string `json:"owner"`
	TokenType    string `json:"token_type"`
	Amount       int64  `json:"amount"`
	ContractAddr string `json:"contract_addr"`

	PublishAddress     string `json:"publish_address"`
	PublishTxHash      string `json:"publish_tx_hash"`
	PublishHeight      int64  `json:"publish_height"`
	PublishHeightIndex int64  `json:"publish_height_index"`
	PublishBlockHash   string `json:"publish_block_hash"`
	PublishBlockTime   int64  `json:"publish_block_time"`
}

func (t *Token) Key() string {
	return TokenID(t.ContractAddr, t.Owner, t.TokenID)
}

func (t *Token) LoadBlockData(data map[string]interface{}) {
	t.PublishBlockHash = convert.ToString(data["evm_block_hash"])
	t.PublishBlockTime = convert.ToInt64(data["evm_block_time"])
	t.PublishHeight = convert.ToInt64(data["evm_height"])
	t.PublishHeightIndex = convert.ToInt64(data["evm_height_index"])
	t.PublishTxHash = convert.ToString(data["evm_tx_hash"])
	t.ContractAddr = util.AddressConvert(convert.ToString(data["contract_addr"]))
}

type Transfer struct {
	TokenID      string `json:"token_id"`
	From         string `json:"from"`
	To           string `json:"to"`
	Operator     string `json:"operator"`
	Amount       int64  `json:"amount"`
	TokenType    string `json:"token_type"`
	ContractAddr string `json:"contract_addr"`

	TxHash      string `json:"tx_hash"`
	Height      int64  `json:"height"`
	HeightIndex int64  `json:"height_index"`
	BlockHash   string `json:"block_hash"`
	BlockTime   int64  `json:"block_time"`
}

func (t *Transfer) IsMint() bool {
	return t.From == EmptyAddr || t.From == EmptyAddrEth
}

func (t *Transfer) Key() string {
	return TransferID(t.TxHash, t.TokenID)
}

func TransferID(txHash, TokenID string) string {
	return fmt.Sprintf("transfer-%s-%s", txHash, TokenID)
}

func (t *Transfer) LoadBlockData(data map[string]interface{}) {
	t.BlockHash = convert.ToString(data["evm_block_hash"])
	t.BlockTime = convert.ToInt64(data["evm_block_time"])
	t.Height = convert.ToInt64(data["evm_height"])
	t.HeightIndex = convert.ToInt64(data["evm_height_index"])
	t.TxHash = convert.ToString(data["evm_tx_hash"])
	t.ContractAddr = convert.ToString(data["contract_addr"])
}

// GetNewToken 获取新增的token
func (t *Transfer) GetNewToken() *Token {
	return &Token{
		TokenID:            t.TokenID,
		Owner:              util.AddressConvert(t.To),
		TokenType:          t.TokenType,
		Amount:             t.Amount,
		ContractAddr:       t.ContractAddr,
		PublishAddress:     t.Operator,
		PublishTxHash:      t.TxHash,
		PublishHeight:      t.Height,
		PublishHeightIndex: t.HeightIndex,
		PublishBlockHash:   t.BlockHash,
		PublishBlockTime:   t.BlockTime,
	}
}

// TxOption Evm TxOption
type TxOption struct {
	ContractAddr string                 `json:"contract_addr,omitempty"`
	ContractType string                 `json:"contract_type,omitempty"`
	CallFunName  string                 `json:"call_fun_name,omitempty"`
	TokenID      []string               `json:"token_id,omitempty"`
	Transfers    []*Transfer            `json:"transfers,omitempty"`
	Events       map[string]interface{} `json:"events,omitempty"`
	Params       map[string]interface{} `json:"params,omitempty"`
}
