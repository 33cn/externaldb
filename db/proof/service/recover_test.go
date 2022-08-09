package service

import (
	"testing"

	proofconfig "github.com/33cn/externaldb/db/proof_config"

	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/proof/model"
)

type myRecoverDB struct {
}

func (c *myRecoverDB) GetProof(id string) (*model.Proof, error) {
	p := &model.Proof{
		Proof: make(map[string]interface{}),
	}
	p.Proof["proof_organization"] = "orgamization1"
	p.Proof["proof_sender"] = "address1"
	p.Proof["proof_deleted"] = ""
	p.Proof["proof_deleted_note"] = ""
	p.Proof["proof_tx_hash"] = "id123"
	p.Proof["basehash"] = "null"
	p.Proof["proof_deleted_flag"] = true
	p.Proof["proof_id"] = "proof-id123"
	return p, nil
}

func (c *myRecoverDB) ListProof(id string) ([]*model.Proof, error) {
	p1 := &model.Proof{
		Proof: make(map[string]interface{}),
	}
	p1.Proof["proof_organization"] = "orgamization1"
	p1.Proof["proof_sender"] = "address1"
	p1.Proof["proof_deleted"] = ""
	p1.Proof["proof_deleted_note"] = ""
	p1.Proof["proof_tx_hash"] = "id124"
	p1.Proof["proof_id"] = "proof-id124"
	p1.Proof["basehash"] = "id123"
	p1.Proof["proof_deleted_flag"] = true

	p2 := &model.Proof{
		Proof: make(map[string]interface{}),
	}
	p2.Proof["proof_organization"] = "orgamization1"
	p2.Proof["proof_sender"] = "address1"
	p2.Proof["proof_deleted"] = ""
	p2.Proof["proof_deleted_note"] = ""
	p2.Proof["proof_tx_hash"] = "id125"
	p2.Proof["basehash"] = "id123"
	p2.Proof["proof_id"] = "proof-id125"
	p2.Proof["proof_deleted_flag"] = true
	return []*model.Proof{p1, p2}, nil
}

func (c *myRecoverDB) GetProofLog(id string) (*model.Log, error) {
	return &model.Log{
		ID: id,
	}, nil
}

func (c *myRecoverDB) GetTemplate(id string) (*model.Template, error) {
	p := &model.Template{
		Template: make(map[string]interface{}),
	}
	p.Template["proof_organization"] = "orgamization1"
	p.Template["proof_sender"] = "address1"
	p.Template["proof_deleted"] = ""
	p.Template["proof_deleted_note"] = ""
	p.Template["proof_tx_hash"] = "id123"
	p.Template["basehash"] = "null"
	p.Template["proof_deleted_flag"] = true
	p.Template["proof_id"] = "proof-id123"
	return p, nil
}

func (c *myRecoverDB) GetProofUpdateRecord(updatehash string, version int) (*model.Proof, error) {
	p := &model.Proof{
		Proof: make(map[string]interface{}),
	}
	return p, nil
}

func Test_RecoverProof(t *testing.T) {
	delPayload := "{\"id\": \"id123\",\"note\":\"test recover\"}"
	tx := &types.Transaction{Execer: []byte("user.p.sanhe.proof_recover"), Payload: []byte(delPayload)}
	tx.To = db.ExecAddress("user.p.sanhe.proof_recover")
	priv := HexToPrivkey(Priv)
	if priv != nil {
		tx.Sign(types.SECP256K1, priv)
	}

	//构建区块
	newblock := &types.Block{}
	newblock.Height = 2
	newblock.BlockTime = types.Now().Unix()
	newblock.ParentHash = zeroHash[:]
	newblock.Txs = []*types.Transaction{tx}
	newblock.TxHash = zeroHash[:]

	env := db.TxEnv{
		TxIndex: 0,
		Block:   &types.BlockDetail{},
	}
	env.Block.Block = newblock
	rec := types.ReceiptData{Ty: types.ExecPack}
	env.Block.Receipts = append(env.Block.Receipts, &rec)

	convert := NewConvert("user.p.sanhe.", "bty", "proof")
	c := convert.(*ProofConvert)
	c.RecordGen = NewRecordGen("p1", "p2", "l1", "l2", "t1", "t2", "u1", "u2")
	c.configDB = &proofconfig.None{}
	c.proofDB = &myRecoverDB{}

	records, err := convert.ConvertTx(&env, db.SeqTypeAdd)

	assert.Nil(t, err)
	assert.Equal(t, 5, len(records))
	for _, r := range records {
		t.Log("Test_DeleteProof", "jsonstr", string(r.Value()))
	}
	r1 := records[0].(*ProofRecord)
	r2 := records[1].(*ProofRecord)
	r3 := records[2].(*ProofRecord)
	l1 := records[3].(*LogRecord)

	assert.Equal(t, false, r1.Proof.Proof["proof_deleted_flag"].(bool))
	assert.Equal(t, false, r2.Proof.Proof["proof_deleted_flag"].(bool))
	assert.Equal(t, false, r3.Proof.Proof["proof_deleted_flag"].(bool))
	assert.Equal(t, "recover", l1.Log.Op)
	assert.Equal(t, db.OpUpdate, r1.Op.OpType())
	assert.Equal(t, db.OpUpdate, r2.Op.OpType())
	assert.Equal(t, db.OpUpdate, r3.Op.OpType())

	records2, err2 := convert.ConvertTx(&env, db.SeqTypeDel)
	assert.Nil(t, err2)
	assert.Equal(t, 5, len(records2))
	for _, r := range records2 {
		t.Log("Test_DeleteProof", "jsonstr", string(r.Value()))
	}
	r11 := records2[0].(*ProofRecord)
	r12 := records2[1].(*ProofRecord)
	r13 := records2[2].(*ProofRecord)
	l11 := records2[3].(*LogRecord)

	assert.Equal(t, true, r11.Proof.Proof["proof_deleted_flag"].(bool))
	assert.Equal(t, true, r12.Proof.Proof["proof_deleted_flag"].(bool))
	assert.Equal(t, true, r13.Proof.Proof["proof_deleted_flag"].(bool))
	assert.Equal(t, db.OpDel, l11.Op.OpType())
	assert.Equal(t, "recover", l1.Log.Op)
	assert.Equal(t, db.OpUpdate, r11.Op.OpType())
	assert.Equal(t, db.OpUpdate, r12.Op.OpType())
	assert.Equal(t, db.OpUpdate, r13.Op.OpType())
}
