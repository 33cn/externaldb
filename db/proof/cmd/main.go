package main

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/33cn/externaldb/db"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	dbcom "github.com/33cn/externaldb/db/common"
	"github.com/33cn/externaldb/db/proof/api"
)

// 预计做下面功能
// 创建存证交易
// 创建存证增量交易
// 创建存证删除交易
// 创建存证恢复交易
func main() {

	basehash := "0xd9c440b8ecfe7dece4897b8f4d56535b9178ebaaf2b77de8d48505091e25d43e"
	appendPayload := map[string]interface{}{
		"basehash":  basehash,
		"Increment": "增量1",
	}
	bs, _ := json.Marshal(appendPayload)
	_ = bs

	j1 := "{\"key\":\"Increment\",\"data\":{\"value\":\"增量1\",\"type\":\"text\",\"format\":\"string\"}}"
	j2 := "{\"key\":\"basehash\",\"data\":{\"value\":\"0xd9c440b8ecfe7dece4897b8f4d56535b9178ebaaf2b77de8d48505091e25d43e\",\"type\":\"text\",\"format\":\"string\"}}"
	jsonstr := "[" + j1 + "," + j2 + "]"

	note := "{\"note\": \"test append proof\"}"
	exec := "user.p.testcsproof.proof"
	createProofTx(exec, Priv, jsonstr, note)

	exec = "user.p.testcsproof.proof_delete"
	delPayload := "{\"id\": \"0xd9c440b8ecfe7dece4897b8f4d56535b9178ebaaf2b77de8d48505091e25d43e\",\"note\":\"test delete\", \"force\":false}"
	createDeleteProofTx(exec, Priv, delPayload)

	exec = "user.p.testcsproof.proof_recover"
	recoverPayload := "{\"id\": \"0xd9c440b8ecfe7dece4897b8f4d56535b9178ebaaf2b77de8d48505091e25d43e\",\"note\":\"test recover\"}"
	createRecoverProofTx(exec, Priv, recoverPayload)

}

func createRecoverProofTx(exec, privStr string, payload string) {
	tx := &types.Transaction{Execer: []byte(exec), Payload: []byte(payload)}
	tx.To = db.ExecAddress(exec)
	tx.Fee = 100000
	tx.Nonce = rand.Int63() // nolint
	priv := HexToPrivkey(Priv)
	if priv != nil {
		tx.Sign(types.SECP256K1, priv)
	}
	bbbb := types.Encode(tx)
	b64 := common.ToHex(bbbb)
	fmt.Printf("\n createRecoverProofTx:\n")

	fmt.Printf("%s\n", b64)
}

func createDeleteProofTx(exec, privStr string, payload string) {
	tx := &types.Transaction{Execer: []byte(exec), Payload: []byte(payload)}
	tx.To = db.ExecAddress(exec)
	tx.Fee = 100000
	tx.Nonce = rand.Int63() // nolint
	priv := HexToPrivkey(Priv)
	if priv != nil {
		tx.Sign(types.SECP256K1, priv)
	}
	bbbb := types.Encode(tx)
	b64 := common.ToHex(bbbb)
	fmt.Printf("\n createDeleteProofTx:\n")

	fmt.Printf("%s\n", b64)
}

func createProofTx(exec, privStr string, data string, note string) {
	v1 := api.ProofInfo{
		Version: "1.0.0",
		Data:    dbcom.EncodeToString([]byte(data)), // base64 encode
		Option:  "",
		Note:    dbcom.EncodeToString([]byte(note)),
	}
	// fmt.Print(v1)
	bs2, _ := json.Marshal(&v1)

	tx := &types.Transaction{Execer: []byte(exec), Payload: bs2}
	tx.To = db.ExecAddress(exec)
	tx.Fee = 100000
	tx.Nonce = rand.Int63() // nolint
	priv := HexToPrivkey(privStr)
	if priv != nil {
		tx.Sign(types.SECP256K1, priv)
	}
	bbbb := types.Encode(tx)
	b64 := common.ToHex(bbbb)
	fmt.Printf("\n createProofTx:\n")

	fmt.Printf("%s\n", b64)
}

// // TODO 需要由普通的结构变成, 复制的带格式说明的结构
// func createIncrementProofTx(basehash, exec, privStr string, payload, note string) {
// 	incrementPayload := make(map[string]interface{})
// 	json.Unmarshal([]byte(payload), &incrementPayload)
//
// 	incrementPayload["basehash"] = basehash
// 	//fmt.Print(incrementPayload)
// 	bs, _ := json.Marshal(incrementPayload)
// 	createProofTx(exec, privStr, string(bs), note)
// }

// Priv 测试环境管理员 addr:133AfuMYQXRxc45JGUb1jLk1M1W4ka39L1
var Priv = "0x85c6c95bcb41779f1d197e686d26b228a523fa36b77cfed79edb59b8853b569b"

// HexToPrivkey ： convert hex string to private key
func HexToPrivkey(key string) crypto.PrivKey {
	cr, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	if err != nil {
		panic(err)
	}
	bkey, err := common.FromHex(key)
	if err != nil {
		panic(err)
	}
	priv, err := cr.PrivKeyFromBytes(bkey)
	if err != nil {
		panic(err)
	}
	return priv
}
