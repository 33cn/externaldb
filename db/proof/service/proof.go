package service

import (
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/proof/model"
)

type proof model.Proof

func newProof(p *model.Proof) *proof {
	return (*proof)(p)
}

func (p *proof) del(op int, hash, note string, force bool) bool {
	if op == db.SeqTypeAdd {
		p.Proof["proof_deleted"] = hash
		p.Proof["proof_deleted_note"] = note
		p.Proof["proof_deleted_flag"] = true
	} else {
		p.Proof["proof_deleted"] = ""
		p.Proof["proof_deleted_note"] = ""
		p.Proof["proof_deleted_flag"] = false
	}
	if force && op == db.SeqTypeDel {
		return false
	}
	return true
}

func (p *proof) recover(op int) {
	if op == db.SeqTypeDel {
		p.Proof["proof_deleted_flag"] = true
	} else {
		p.Proof["proof_deleted_flag"] = false
	}
}
