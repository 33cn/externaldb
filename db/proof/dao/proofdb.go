package dao

import (
	"encoding/json"

	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/proof/model"
	"github.com/33cn/externaldb/db/proof/proofdb"
)

var _ proofdb.IProofDB = &ProofDB{}

// ProofDB ProofDB
type ProofDB struct {
	db            db.WrapDB
	dbX, tableX   string
	ldbX, ltableX string
	tdbX, ttableX string
	udbX, utableX string
}

// NewProofDB NewProofDB
func NewProofDB(db db.WrapDB) *ProofDB {
	d := &ProofDB{
		db:      db,
		dbX:     proofdb.ProofDBX,
		tableX:  proofdb.ProofTableX,
		ldbX:    proofdb.LogDBX,
		ltableX: proofdb.LogTableX,
		tdbX:    proofdb.TemplateDBX,
		ttableX: proofdb.TemplateTableX,
		udbX:    proofdb.ProofUpdateDBX,
		utableX: proofdb.ProofUpdateTableX,
	}
	return d
}

// GetProof impl
func (db1 *ProofDB) GetProof(id string) (*model.Proof, error) {
	m1, err := db1.db.Get(db1.dbX, db1.tableX, id)
	if err != nil {
		return nil, err
	}
	var p model.Proof
	err = json.Unmarshal(*m1, &p.Proof)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ListProof 找出id增量存证 basehash = id
func (db1 *ProofDB) ListProof(id string) ([]*model.Proof, error) {
	kv := &db.ListKV{Key: "basehash", Value: id}
	m1, err := db1.db.List(db1.dbX, db1.tableX, []*db.ListKV{kv})
	if err != nil {
		return nil, err
	}
	proofs := make([]*model.Proof, 0)
	for _, r := range m1 {
		var p model.Proof
		err = json.Unmarshal(*r, &p.Proof)
		if err != nil {
			return nil, err
		}
		proofs = append(proofs, &p)
	}
	return proofs, nil
}

// GetProofLog impl
func (db1 *ProofDB) GetProofLog(id string) (*model.Log, error) {
	m1, err := db1.db.Get(db1.ldbX, db1.ltableX, id)
	if err != nil {
		return nil, err
	}
	var p model.Log
	err = json.Unmarshal(*m1, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetProof impl
func (db1 *ProofDB) GetTemplate(id string) (*model.Template, error) {
	m1, err := db1.db.Get(db1.tdbX, db1.ttableX, id)
	if err != nil {
		return nil, err
	}
	var t model.Template
	err = json.Unmarshal(*m1, &t.Template)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetProofLog impl
func (db1 *ProofDB) GetProofUpdateRecord(id string, version int) (*model.Proof, error) {
	kv1 := &db.ListKV{Key: "update_hash", Value: id}
	kv2 := &db.ListKV{Key: "update_version", Value: version}
	m1, err := db1.db.List(db1.udbX, db1.utableX, []*db.ListKV{kv1, kv2})
	if err != nil {
		return nil, err
	}
	proofs := make([]*model.Proof, 0)
	for _, r := range m1 {
		var p model.Proof
		err = json.Unmarshal(*r, &p.Proof)
		if err != nil {
			return nil, err
		}
		proofs = append(proofs, &p)
	}
	return proofs[0], nil
}
