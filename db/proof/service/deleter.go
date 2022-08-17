package service

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/proof/api"
	"github.com/33cn/externaldb/db/proof/model"
	"github.com/33cn/externaldb/db/proof/proofdb"
)

type deleter struct {
	env    *db.TxEnv
	arg    *api.DeleteProof
	txHash string
}

func newDeleter(env *db.TxEnv, hash string, arg *api.DeleteProof) *deleter {
	return &deleter{env: env, txHash: hash, arg: arg}
}

func (d *deleter) del(p *model.Proof, ps []*model.Proof, op int, gen proofdb.IProofRecord) ([]db.Record, error) {
	tx := d.env.Block.Block.Txs[d.env.TxIndex]
	from := tx.From()
	hash := common.ToHex(tx.Hash())

	var records []db.Record

	mp := newProof(p)
	ret := mp.del(op, hash, d.arg.Note, d.arg.Force)
	if !ret {
		// 强制删除, 不能回滚
		return nil, nil
	}

	op2 := db.OpUpdate
	if d.arg.Force {
		op2 = db.OpDel
	}
	id := p.Proof["proof_id"].(string)
	r1 := gen.Proof((*model.Proof)(mp), id, op2)
	records = append(records, r1)

	for _, p2 := range ps {
		mp := newProof(p2)
		mp.del(op, hash, d.arg.Note, d.arg.Force)
		id := p2.Proof["proof_id"].(string)
		r2 := gen.Proof((*model.Proof)(mp), id, op2)
		records = append(records, r2)
	}

	ml := newProofLog(hash, from)
	ml.setBlock(d.env.BlockHash, d.env.Block.Block.BlockTime, d.env.Block.Block.Height, d.env.TxIndex)
	ml.del(op, p, d.arg)
	op3 := db.OpAdd
	if op == db.SeqTypeDel {
		op3 = db.OpDel
	}
	r3 := gen.Log((*model.Log)(ml), hash, op3)
	records = append(records, r3)

	return records, nil
}
