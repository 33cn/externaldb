package db

import (
	"github.com/33cn/chain33/types"
	l "github.com/xuperchain/log15"
	"github.com/33cn/externaldb/db"
)

var log = l.New("module", "db.file_part")

// DB operate FilePart
type DB interface {
	Get(hash string) (*FilePart, error)
	Set(r *RecordFilePart) error
	Clean() error
}

// esGRPCDB es and grpc db
type esGRPCDB struct {
	es, grpc DB
}

// NewEsGrpcDB new
func NewEsGrpcDB(esCli db.WrapDB, grCli types.Chain33Client) DB {
	return &esGRPCDB{
		grpc: NewGRpcDB(grCli),
		es:   NewEsDB(esCli),
	}
}

// Get file part by tx hash
func (d *esGRPCDB) Get(hash string) (*FilePart, error) {
	fp, err := d.es.Get(hash)
	switch err {
	case nil:
	case db.ErrDBNotFound:
		fp, err = d.grpc.Get(hash)
		if err != nil {
			return nil, err
		}
		go func() {
			err := d.Set(NewRecordFilePart(db.OpAdd, fp))
			if err != nil {
				log.Error("esGRPCDB.es.Set", "err", err)
			}
		}()
	default:
		return nil, err
	}
	return fp, nil
}

func (d *esGRPCDB) Set(r *RecordFilePart) error {
	return d.es.Set(r)
}

func (d *esGRPCDB) Clean() error {
	return d.es.Clean()
}
