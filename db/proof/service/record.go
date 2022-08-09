package service

import (
	"encoding/json"

	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/proof/model"
)

// NewLogKey 为LogDBX数据的LogTableX表创建一个唯一的id标识log-hash
func (g *RecordGen) newLogKey(id string) *db.IKey {
	return db.NewIKey(g.logDB, g.logTable, id)
}

// NewProofKey 为ProofDBX数据的ProofTableX表创建一个唯一的id标识proof-hash
func (g *RecordGen) newProofKey(id string) *db.IKey {
	return db.NewIKey(g.proofDB, g.proofTable, id)
}

// NewTemplateKey 为TemplateDBX数据的TemplateTableX表创建一个唯一的id标识template-hash
func (g *RecordGen) newTemplateKey(id string) *db.IKey {
	return db.NewIKey(g.TemplateDB, g.TemplateTable, id)
}

// NewProofUpdateKey 为ProofUpdateDBX数据的ProofUpdateTableX表创建一个唯一的id标识proofUpdate-hash
func (g *RecordGen) newProofUpdateRecordKey(id string) *db.IKey {
	return db.NewIKey(g.ProofUpdateDB, g.ProofUpdateTable, id)
}

// LogRecord 用于db 记录
type LogRecord struct {
	*db.IKey
	*db.Op
	Log *model.Log
}

// Value impl
func (r *LogRecord) Value() []byte {
	v, _ := json.Marshal(r.Log)
	return v
}

// ProofRecord 用于db 记录
type ProofRecord struct {
	*db.IKey
	*db.Op
	Proof *model.Proof
}

// Value impl
func (r *ProofRecord) Value() []byte {
	v, _ := json.Marshal(r.Proof.Proof)
	return v
}

// TemplateRecord 用于db 记录
type TemplateRecord struct {
	*db.IKey
	*db.Op
	Template *model.Template
}

// Value impl
func (r *TemplateRecord) Value() []byte {
	v, _ := json.Marshal(r.Template.Template)
	return v
}

// ProofUpdateRecord 用于db 记录
type ProofUpdateRecord struct {
	*db.IKey
	*db.Op
	Proof *model.Proof
}

// Value impl
func (r *ProofUpdateRecord) Value() []byte {
	v, _ := json.Marshal(r.Proof.Proof)
	return v
}

// RecordGen gen record
type RecordGen struct {
	proofDB          string
	proofTable       string
	logDB            string
	logTable         string
	TemplateDB       string
	TemplateTable    string
	ProofUpdateDB    string
	ProofUpdateTable string
}

// NewRecordGen NewRecordGen
func NewRecordGen(pdb, ptable, ldb, ltable, tdb, ttable, udb, utable string) *RecordGen {
	return &RecordGen{
		proofDB:          pdb,
		proofTable:       ptable,
		logDB:            ldb,
		logTable:         ltable,
		TemplateDB:       tdb,
		TemplateTable:    ttable,
		ProofUpdateDB:    udb,
		ProofUpdateTable: utable,
	}
}

// Log gen log record
func (g *RecordGen) Log(log *model.Log, id string, op int) db.Record {
	record := LogRecord{
		IKey: g.newLogKey(id),
		Op:   db.NewOp(op),
		Log:  log,
	}

	return &record
}

// Proof gen proof record
func (g *RecordGen) Proof(proof *model.Proof, id string, op int) db.Record {
	ProofRecord := ProofRecord{
		IKey:  g.newProofKey(id),
		Op:    db.NewOp(op),
		Proof: proof,
	}

	return &ProofRecord
}

// Template gen template record
func (g *RecordGen) Template(template *model.Template, id string, op int) db.Record {
	TemplateRecord := TemplateRecord{
		IKey:     g.newTemplateKey(id),
		Op:       db.NewOp(op),
		Template: template,
	}

	return &TemplateRecord
}

// ProofUpdate gen proof update record
func (g *RecordGen) ProofUpdateRecord(proof *model.Proof, id string, op int) db.Record {
	ProofUpdateRecord := ProofUpdateRecord{
		IKey:  g.newProofUpdateRecordKey(id),
		Op:    db.NewOp(op),
		Proof: proof,
	}

	return &ProofUpdateRecord
}
