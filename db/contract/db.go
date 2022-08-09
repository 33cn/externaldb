package contract

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/33cn/chain33/common/crypto"
	l "github.com/33cn/chain33/common/log/log15"
	pabi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	lru "github.com/hashicorp/golang-lru"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/contractverify"
	"github.com/33cn/externaldb/util"
	"github.com/33cn/go-kit/convert"
)

const (
	TableName = "contract"
	KeyPrefix = TableName + "-"
)

var log = l.New("module", "db.contract")

// DB operate FilePart
type DB interface {
	Get(hash string) (*Contract, error)
	Set(r *Record) error
	GetABI(addr string) (*pabi.ABI, error)
	GetContract(addr string) (*Contract, error)
	UpdateCache(contract *Contract)
}

type Contract struct {
	Address           string                         `json:"contract_address"`
	Creator           string                         `json:"creator"`
	DeployBlockHash   string                         `json:"deploy_block_hash"`
	DeployBlockTime   int64                          `json:"deploy_block_time"`
	DeployHeight      int64                          `json:"deploy_height"`
	DeployHeightIndex int64                          `json:"deploy_height_index"`
	DeployTxHash      string                         `json:"deploy_tx_hash"`    // 部署的交易hash
	ContractBinHash   string                         `json:"contract_bin_hash"` // bin文件直接sha256处理
	ContractBin       string                         `json:"contract_bin"`
	ContractAbi       string                         `json:"contract_abi"`
	ContractType      string                         `json:"contract_type"`
	TxCount           int64                          `json:"tx_count"`
	PublishCount      int64                          `json:"publish_count"`
	ParsedAbi         *pabi.ABI                      `json:"-"`
	ParsedAbiErr      error                          `json:"-"`
	ContractVerify    *contractverify.ContractVerify `json:"-"`
	Name              string                         `json:"name"`
	Symbol            string                         `json:"symbol"`
	URI               string                         `json:"uri"`
}

func NewContractByMap(data map[string]interface{}) *Contract {
	var c Contract
	c.Address = convert.ToString(data["contract_addr"])
	c.Creator = convert.ToString(data["contract_creator"])
	c.DeployBlockHash = convert.ToString(data["evm_block_hash"])
	c.DeployBlockTime = convert.ToInt64(data["evm_block_time"])
	c.DeployHeight = convert.ToInt64(data["evm_height"])
	c.DeployHeightIndex = convert.ToInt64(data["evm_height_index"])
	c.DeployTxHash = convert.ToString(data["evm_tx_hash"])
	c.ContractBin = convert.ToString(data["contract_bin"])
	c.ContractAbi = convert.ToString(data["contract_abi"])
	c.ContractVerify = &contractverify.ContractVerify{
		ContractBin: c.ContractBin,
		ContractAbi: c.ContractAbi,
	}
	return &c
}

func (c *Contract) Key() string {
	return AddKeyPrefix(c.Address)
}

func (c *Contract) SetContractBinHash(bin string) {
	c.ContractBinHash = hex.EncodeToString(crypto.Sha256([]byte(bin)))
}

func (c *Contract) GetContractBinHash() string {
	if c.ContractBinHash == "" {
		c.ContractBinHash = hex.EncodeToString(crypto.Sha256([]byte(c.ContractBin)))
	}
	return c.ContractBinHash
}

// Record es record
type Record struct {
	*db.IKey
	*db.Op
	value *Contract
}

// Value value
func (r *Record) Value() []byte {
	v, _ := json.Marshal(r.value)
	return v
}

func (r *Record) Source() interface{} {
	return r.value
}

// NewRecord new RecordFilePart
func NewRecord(op int, value *Contract) *Record {
	return &Record{
		IKey:  db.NewIKey(TableName, TableName, value.Key()),
		Op:    db.NewOp(op),
		value: value,
	}
}

// AddKeyPrefix 添加索引前缀
func AddKeyPrefix(hash string) string {
	return KeyPrefix + hash
}

// InitESDB 创建db (插件满足接口 db.ExecConvert)
func InitESDB(cli db.DBCreator) error {
	return util.InitIndex(cli, TableName, TableName, Mapping)
}

// NewEsDB new es DB
func NewEsDB(client db.WrapDB) DB {
	cd := &esDB{}
	cd.db = client
	cd.cache, _ = lru.New(1024)
	return cd
}

// EsDB elasticsearch
type esDB struct {
	cache *lru.Cache
	db    db.WrapDB
}

// Set a file record
func (d *esDB) Set(r *Record) error {
	return d.db.Set(TableName, TableName, r.value.Key(), r)
}

// Get 加载合约，优先从缓存加载
func (d *esDB) Get(addr string) (*Contract, error) {
	if c, ok := d.cache.Get(addr); ok {
		if ct, ok := c.(*Contract); ok {
			return ct, nil
		}
		log.Error("contract cache error", "cache", c)
		d.cache.Remove(addr)
	}
	var c = &Contract{}
	data, err := d.db.Get(TableName, TableName, AddKeyPrefix(addr))
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(*data, c); err != nil {
		return nil, err
	}
	if c.ContractAbi == "" {
		vfyDb := contractverify.NewEsDB(d.db)
		vfy, err := vfyDb.Get(c.GetContractBinHash())
		if err == nil && vfy.ContractAbi != "" {
			c.ContractAbi = vfy.ContractAbi
			c.ContractVerify = vfy
		}
	}
	if c.ContractAbi != "" {
		contractABI, err := pabi.JSON(strings.NewReader(c.ContractAbi))
		if err != nil {
			log.Error("unpackValues: abi.JSON", "err", err)
			return nil, err
		}
		c.ParsedAbiErr = err
		c.ParsedAbi = &contractABI
	} else {
		c.ParsedAbiErr = fmt.Errorf("no abi in contract %s", addr)
	}
	d.cache.Add(addr, c)
	return c, nil
}

// GetABI 通过合约地址获取abi
func (d *esDB) GetABI(addr string) (*pabi.ABI, error) {
	cc, err := d.Get(addr)
	if err != nil {
		return nil, err
	}
	return cc.ParsedAbi, cc.ParsedAbiErr
}

func (d *esDB) GetContract(addr string) (*Contract, error) {
	var c = &Contract{}
	data, err := d.db.Get(TableName, TableName, AddKeyPrefix(addr))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(*data, &c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (d *esDB) UpdateCache(c *Contract) {
	if c == nil {
		return
	}
	if c.ContractAbi == "" {
		vfyDb := contractverify.NewEsDB(d.db)
		vfy, err := vfyDb.Get(c.GetContractBinHash())
		if err == nil && vfy.ContractAbi != "" {
			c.ContractAbi = vfy.ContractAbi
			c.ContractVerify = vfy
		}
	}
	d.cache.Add(c.Address, c)
	if c.ParsedAbi != nil {
		return
	}
	if c.ContractAbi != "" {
		contractABI, err := pabi.JSON(strings.NewReader(c.ContractAbi))
		if err != nil {
			log.Error("unpackValues: abi.JSON", "err", err)
			return
		}
		c.ParsedAbiErr = err
		c.ParsedAbi = &contractABI
	} else {
		c.ParsedAbiErr = fmt.Errorf("no abi in contract %s", c.Address)
	}
}
