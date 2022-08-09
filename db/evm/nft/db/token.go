package db

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	pcom "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/go-kit/convert"
)

type Token struct {
	Owner            string   `json:"owner,omitempty"`
	GoodsID          int64    `json:"goods_id"`
	Amount           int64    `json:"amount"`
	GoodsType        int32    `json:"goods_type"`
	PublishTime      int64    `json:"publish_time"`
	Hash             []string `json:"trace_hash"`
	Source           []string `json:"source_hash,omitempty"`
	Publisher        string   `json:"publisher"`
	PublisherAddress string   `json:"publisher_address"`
	Name             string   `json:"name"`
	LabelID          string   `json:"label_id"`
	Remark           string   `json:"remark,omitempty"`
	FileURL          string   `json:"file_url"`
	FileType         string   `json:"file_type"`
	Operator         string   `json:"operator"`               // 操作者
	BatchNumber      string   `json:"batch_number,omitempty"` // 批次号
	Image            string   `json:"image,omitempty"`
	IsUsed           bool     `json:"is_used"`                  // 使用标志
	UseTime          int64    `json:"use_time,omitempty"`       // 使用时间
	UserAddr         string   `json:"user_addr,omitempty"`      // 使用人地址
	UserName         string   `json:"user_name,omitempty"`      // 使用人名称
	UserRealName     string   `json:"user_real_name,omitempty"` // 使用人真实名称
	UseCode          string   `json:"use_code,omitempty"`       // 使用的验证码
	TokenType        string   `json:"token_type"`
	EvmState
}

func (t *Token) Process() {
	if t.Remark != "" {
		if buf, err := base64.StdEncoding.DecodeString(t.Remark); err == nil {
			var rm TokenRemark
			if err := json.Unmarshal(buf, &rm); err == nil {
				t.FileURL = rm.FileURL
				t.FileType = rm.FileType
			}
		}
	}
	if t.FileURL == "" && t.Image != "" {
		if buf, err := base64.StdEncoding.DecodeString(t.Image); err == nil {
			var rm TokenRemark
			if err := json.Unmarshal(buf, &rm); err == nil {
				t.FileURL = rm.FileURL
				t.FileType = rm.FileType
			}
		}
	}
	if t.PublishTime == 0 {
		t.PublishTime = t.BlockTime
	}
}

type EvmState struct {
	ContractAddr string `json:"contract_addr,omitempty"`
	TxHash       string `json:"evm_tx_hash,omitempty"`
	HeightIndex  int64  `json:"evm_height_index,omitempty"`
	Height       int64  `json:"evm_height,omitempty"`
	BlockHash    string `json:"evm_block_hash,omitempty"`
	BlockTime    int64  `json:"evm_block_time,omitempty"`
}

func (e *EvmState) GetEvmState(m map[string]interface{}) {
	e.ContractAddr = convert.ToString(m["contract_addr"])
	e.TxHash = convert.ToString(m["evm_tx_hash"])
	e.HeightIndex = convert.ToInt64(m["evm_height_index"])
	e.Height = convert.ToInt64(m["evm_height"])
	e.BlockHash = convert.ToString(m["evm_block_hash"])
	e.BlockTime = convert.ToInt64(m["evm_block_time"])
}

func (t *Token) Key() string {
	return TokenID(t.ContractAddr, t.Owner, t.GoodsID)
}

func TokenID(contractAddr, ownerAddr string, goodsID int64) string {
	return fmt.Sprintf("%s-%s-%s-%v", "nft", contractAddr, ownerAddr, goodsID)
}

// TokenRemark 合约备注信息
type TokenRemark struct {
	FileType string `json:"file_type"`
	FileURL  string `json:"file_url"`
}

// AddNewGoodsResult 发行event结构
type AddNewGoodsResult struct {
	Owner            string   `json:"owner"`
	GoodsID          int64    `json:"goodsID"`
	Amount           int64    `json:"amount"`
	GoodsType        int32    `json:"goodsType"`
	PublishTime      int64    `json:"publishTime"`
	Hash             []string `json:"hash"`
	Source           []string `json:"source"`
	Publisher        string   `json:"publisher"`
	Name             string   `json:"name"`
	LabelID          string   `json:"labelID"`
	Remark           string   `json:"remark"`
	Operator         string   `json:"operator"`
	PublisherAddress string   `json:"publisherAddress"`
	BatchNumber      string   `json:"batchNumber"`
	Image            string   `json:"image"`
	TraceHash        string   `json:"traceHash"`
	IsUsed           bool     `json:"isUsed"`
}

func NewAddNewGoodsResult(m map[string]interface{}) (*AddNewGoodsResult, error) {
	var r AddNewGoodsResult
	buf, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(buf, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (r AddNewGoodsResult) GetToken(info map[string]interface{}) *Token {
	var t = Token{
		Owner:            r.Owner,
		GoodsID:          r.GoodsID,
		Amount:           r.Amount,
		GoodsType:        r.GoodsType,
		Publisher:        r.Publisher,
		PublishTime:      r.PublishTime,
		PublisherAddress: r.PublisherAddress,
		Hash:             r.Hash,
		Source:           r.Source,
		Name:             r.Name,
		LabelID:          r.LabelID,
		Remark:           r.Remark,
		Operator:         r.Operator,
		BatchNumber:      r.BatchNumber,
		Image:            r.Image,
		IsUsed:           r.IsUsed,
	}
	t.GetEvmState(info)
	if r.TraceHash != "" {
		t.Hash = append(t.Hash, r.TraceHash)
	}
	t.Process()
	return &t
}

type BatchAddNewGoodsResult struct {
	Owner       string   `json:"owner"`
	GoodsID     []int64  `json:"goodsIDs"`
	Amount      []int64  `json:"amounts"`
	GoodsType   int32    `json:"goodsType"`
	PublishTime int64    `json:"publishTime"`
	Hash        []string `json:"hash"`
	Source      []string `json:"source"`
	Publisher   string   `json:"publisher"`
	Name        []string `json:"names"`
	LabelID     []string `json:"labelIDs"`
	Remark      []string `json:"remarks"`
	Results     []*AddNewGoodsResult
}

func NewBatchAddNewGoodsResult(m map[string]interface{}) (*BatchAddNewGoodsResult, error) {
	var r BatchAddNewGoodsResult
	buf, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(buf, &r); err != nil {
		return nil, err
	}
	for i := range r.GoodsID {
		res := AddNewGoodsResult{
			Owner:       r.Owner,
			GoodsType:   r.GoodsType,
			PublishTime: r.PublishTime,
			Publisher:   r.Publisher,
		}
		if len(r.GoodsID) > i {
			res.GoodsID = r.GoodsID[i]
		}
		if len(r.Amount) > i {
			res.Amount = r.Amount[i]
		}
		res.Hash = r.Hash
		res.Source = r.Source
		if len(r.Name) > i {
			res.Name = r.Name[i]
		}
		if len(r.LabelID) > i {
			res.LabelID = r.LabelID[i]
		}
		if len(r.Remark) > i {
			res.Remark = r.Remark[i]
		}
		r.Results = append(r.Results, &res)
	}
	return &r, nil
}

// MintGoodsResult 批量发行event结构，由于使用时是数组形式，所以地址类型没有被解析成字符串
type MintGoodsResult struct {
	AddNewGoodsResult
	Owner            *pcom.Hash160Address `json:"owner"`
	Operator         *pcom.Hash160Address `json:"operator"`
	PublisherAddress *pcom.Hash160Address `json:"publisherAddress"`
}

func NewMintGoodsResults(m interface{}) ([]MintGoodsResult, error) {
	res := make([]MintGoodsResult, 0)
	buf, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(buf, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (r MintGoodsResult) GetToken(info map[string]interface{}) *Token {
	t := r.AddNewGoodsResult.GetToken(info)
	if r.Owner != nil {
		t.Owner = db.Hash160AddressToString(*r.Owner)
	}
	if r.Operator != nil {
		t.Operator = db.Hash160AddressToString(*r.Operator)
	}
	if r.PublisherAddress != nil {
		t.PublisherAddress = db.Hash160AddressToString(*r.PublisherAddress)
	}
	if t.Owner == "" {
		t.Owner = t.PublisherAddress
	}
	return t
}

// GoodsUsedResult 货物领取event结构
type GoodsUsedResult struct {
	From    string `json:"from"`
	To      string `json:"to"`
	GoodsID int64  `json:"id"`
	Code    string `json:"code"`
	IsUsed  bool   `json:"isUsed"`
}

func NewGoodsUsedResult(m map[string]interface{}) (*GoodsUsedResult, error) {
	var r GoodsUsedResult
	buf, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(buf, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
