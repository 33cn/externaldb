package transaction

import (
	"strings"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/util"
)

// GetTransactionRecord get tx record
func GetTransactionRecord(env *db.TxEnv, op int) db.Record {
	tx := ConvertTransaction(env)
	txRecord := TxRecord{
		IKey: NewTransactionKey(tx.Hash),
		Op:   db.NewOp(op),
		Tx:   tx,
	}
	return &txRecord
}

// ConvertTransaction ConvertTransaction
func ConvertTransaction(env *db.TxEnv) *Transaction {
	tx := env.Block.Block.Txs[env.TxIndex]
	receipt := env.Block.Receipts[env.TxIndex]
	amount, err := tx.Amount()
	if err != nil {
		amount = 0
	}
	tx2 := Transaction{
		HeightIndex: db.HeightIndex(env.Block.Block.Height, env.TxIndex),
		Block: &Block{
			Height:    env.Block.Block.Height,
			BlockTime: env.Block.Block.BlockTime,
			BlockHash: env.BlockHash,
		},

		Success:    receipt.Ty == types.ExecOk,
		Index:      env.TxIndex,
		Hash:       common.ToHex(tx.Hash()),
		From:       tx.From(),
		To:         tx.GetRealToAddr(),
		Amount:     amount,
		Fee:        tx.Fee,
		Execer:     string(tx.Execer),
		ActionName: tx.ActionName(),
		GroupCount: int64(tx.GroupCount),
		IsWithdraw: tx.IsWithdraw(string(tx.GetExecer())),
		Assets:     genAssets(tx),
		Next:       common.ToHex(tx.Next),
		IsPara:     strings.HasPrefix(string(tx.Execer), "user.p."),
	}

	if string(tx.Execer) == "none" && receipt.Ty == types.ExecPack {
		tx2.Success = true
	}

	return &tx2
}

// InitDB init db
func InitDB(cli db.DBCreator) error {
	return util.InitIndex(cli, TransactionX, TransactionX, TxRecordMapping)
}

func genAssets(tx *types.Transaction) []Asset {
	assets, err := tx.Assets()
	if err != nil {
		return nil
	}
	output := make([]Asset, 0)
	for _, a := range assets {
		output = append(output, Asset{Exec: a.Exec, Symbol: a.Symbol, Amount: a.Amount})
	}
	return output
}
