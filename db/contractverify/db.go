package contractverify

import (
	"encoding/hex"
	"encoding/json"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/util"
)

const (
	TableName = "contract_verify"
	KeyPrefix = TableName + "-"
)

// DB operate FilePart
type DB interface {
	Get(hash string) (*ContractVerify, error)
	Set(r *Record) error
}

type ContractVerify struct {
	ContractBinHash string `json:"contract_bin_hash"`
	ContractBin     string `json:"contract_bin"`
	ContractAbi     string `json:"contract_abi"`
	ContractType    string `json:"contract_type"`
	CompileType     string `json:"compile_type"`
	CompileVersion  string `json:"compile_version"`
}

func (c *ContractVerify) Key() string {
	return AddKeyPrefix(c.GetContractBinHash())
}

func (c *ContractVerify) GetContractBinHash() string {
	if c.ContractBinHash == "" {
		c.ContractBinHash = hex.EncodeToString(crypto.Sha256([]byte(c.ContractBin)))
	}
	return c.ContractBinHash
}

// Record es record
type Record struct {
	*db.IKey
	*db.Op
	value *ContractVerify
}

// Value value
func (r *Record) Value() []byte {
	v, _ := json.Marshal(r.value)
	return v
}

// NewRecord new RecordFilePart
func NewRecord(op int, value *ContractVerify) *Record {
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
	return &esDB{client: client}
}

// EsDB elasticsearch
type esDB struct {
	client db.WrapDB
}

// Get FilePart
func (d *esDB) Get(hash string) (*ContractVerify, error) {
	buf, err := d.client.Get(TableName, TableName, AddKeyPrefix(hash))
	if err != nil {
		return nil, err
	}
	var ans ContractVerify
	err = json.Unmarshal(*buf, &ans)
	return &ans, err
}

// Set a file record
func (d *esDB) Set(r *Record) error {
	return d.client.Set(TableName, TableName, r.value.Key(), r)
}
