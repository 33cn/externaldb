package db

import (
	"context"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db"
)

// gRPCDB chain grpc client
type gRPCDB struct {
	client types.Chain33Client
}

// NewGRpcDB new grpc db
func NewGRpcDB(client types.Chain33Client) DB {
	return &gRPCDB{client: client}
}

// Get file part by tx hash
func (d *gRPCDB) Get(hash string) (*FilePart, error) {
	buf, err := common.FromHex(hash)
	if err != nil {
		return nil, err
	}
	param := &types.ReqHash{Hash: buf}
	resp, err := d.client.QueryTransaction(context.Background(), param)
	if err != nil {
		log.Error("gRPCDB.Get QueryTransaction", "err", err)
		return nil, err
	}
	if resp.Tx == nil {
		return nil, db.ErrDBNotFound
	}
	ans := FilePart{
		Data:   string(resp.Tx.Payload),
		TxHash: common.ToHex(resp.Tx.Hash()),
	}
	return &ans, nil
}

// Set impl DB
func (d *gRPCDB) Set(*RecordFilePart) error {
	return db.ErrDBInvalidOperation
}

// Clean impl DB
func (d *gRPCDB) Clean() error {
	return db.ErrDBInvalidOperation
}
