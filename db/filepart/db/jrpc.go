package db

import (
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
)

// jRPCDB chain json rpc client
type jRPCDB struct {
	client *jsonclient.JSONClient
}

// NewJRpcDB new Chain json rpc db
func NewJRpcDB(client *jsonclient.JSONClient) DB {
	return &jRPCDB{client: client}
}

// Get file part by tx hash
func (d jRPCDB) Get(hash string) (*FilePart, error) {
	var resp rpctypes.TransactionDetail
	err := d.client.Call("Chain33.QueryTransaction", rpctypes.QueryParm{Hash: hash}, &resp)
	if err != nil {
		log.Error("JRPCClient client.Call", "err", err)
		return nil, err
	}
	if resp.Tx == nil {
		return nil, db.ErrDBNotFound
	}
	pl, err := common.HexToBytes(resp.Tx.RawPayload)
	if err != nil {
		log.Error("JRPCClient.Get common.HexToBytes", "err", err)
		return nil, err
	}
	ans := FilePart{
		Data:   string(pl),
		TxHash: resp.Tx.Hash,
	}
	return &ans, nil
}

// Set impl DB
func (d *jRPCDB) Set(*RecordFilePart) error {
	return db.ErrDBInvalidOperation
}

// Clean impl DB
func (d *jRPCDB) Clean() error {
	return db.ErrDBInvalidOperation
}
