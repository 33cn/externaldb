package proofdb

import (
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/proof/model"
)

// IProofDB ProofDB
type IProofDB interface {
	GetProof(id string) (*model.Proof, error)
	ListProof(baseHash string) ([]*model.Proof, error)
	GetProofLog(id string) (*model.Log, error)
	GetTemplate(id string) (*model.Template, error)
	GetProofUpdateRecord(updatehash string, version int) (*model.Proof, error)
}

// IProofRecord 抽象记录, 用于输出, 不直接保存数据库
type IProofRecord interface {
	Log(l *model.Log, id string, op int) db.Record
	Proof(p *model.Proof, id string, op int) db.Record
	Template(t *model.Template, id string, op int) db.Record
	ProofUpdateRecord(p *model.Proof, id string, op int) db.Record
}

// default asset db
const (
	ProofDBX          = "proof"
	ProofTableX       = "proof"
	LogDBX            = "proof_log"
	LogTableX         = "proof_log"
	TemplateDBX       = "proof_template"
	TemplateTableX    = "proof_template"
	ProofUpdateDBX    = "proof_update"
	ProofUpdateTableX = "proof_update"
	DefaultType       = "_doc"
)
