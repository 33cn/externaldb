package service

import (
	"fmt"

	"github.com/33cn/externaldb/db/proof/api"
	"github.com/33cn/externaldb/db/proof/model"
)

type proofLog model.Log

func newProofLog(id string, from string) *proofLog {
	l := &model.Log{
		Address: from,
		ID:      id,
	}
	return (*proofLog)(l)
}

func (l *proofLog) setBlock(hash string, time int64, height, index int64) {
	l.BlockHash = hash
	l.BlockTime = time
	l.Height = height
	l.Index = index
}

func (l *proofLog) del(op int, p *model.Proof, arg *api.DeleteProof) {
	l.Op = "delete"
	l.Note = arg.Note
	l.Force = arg.Force
	l.ProofHash = arg.ID
}

func (l *proofLog) recover(op int, p *model.Proof, arg *api.RecoverProof) {
	l.Op = "recover"
	l.Note = arg.Note
	l.Force = false
	l.ProofHash = arg.ID
}

// LogID log-hash
func LogID(hash string) string {
	return fmt.Sprintf("%s-%s", model.LogID, hash)
}
