package account

import (
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/util"
)

// InitDB init account db
func InitDB(cli db.DBCreator) error {
	err := util.InitIndex(cli, AccountRecordDBX, AccountRecordTableX, AccountMapping)
	if err != nil {
		return err
	}

	return util.InitIndex(cli, DBX, TableX, AccountMapping)
}

// RecordHelper create account record
func RecordHelper(log *types.ReceiptLog, op int, acc Account) ([]db.Record, error) {
	var records []db.Record

	accountDetail, err := AssetLogConvert(log.Ty, log.Log, op)
	if err != nil {
		return nil, err
	}
	acc.Detall = accountDetail
	record := &Record{Acc: acc, IKey: NewAccountKey(acc.Key()), Op: db.NewOp(db.OpAdd)}
	records = append(records, record)

	rrecord := &Record{Acc: acc, IKey: NewAccountRecordKey(acc.RecordKey()), Op: db.NewOp(op)}
	records = append(records, rrecord)
	return records, nil
}
