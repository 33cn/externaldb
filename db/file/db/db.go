package db

import (
	"strings"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db"
	fpdb "github.com/33cn/externaldb/db/filepart/db"
	fsdb "github.com/33cn/externaldb/db/filesummary/db"
)

type DB interface {
	Get(hash string) (*File, error)
}

type esGrpcDB struct {
	fpEsGrpcDB fpdb.DB
	fsEsDB     fsdb.DB
}

func NewEsGrpcDB(esCli db.WrapDB, grCli types.Chain33Client) DB {
	return &esGrpcDB{
		fpEsGrpcDB: fpdb.NewEsGrpcDB(esCli, grCli),
		fsEsDB:     fsdb.NewEsDB(esCli),
	}
}

func (d *esGrpcDB) Get(hash string) (*File, error) {
	sum, err := d.fsEsDB.Get(hash)
	if err != nil {
		return nil, err
	}

	if sum.FileBlacklistFlag {
		return nil, fsdb.ErrFileBlacklisted
	}

	hashes := sum.PartHashs
	content := strings.Builder{}
	for i := 0; i < len(hashes); i += common.Sha256Len {
		ph := hashes[i : i+common.Sha256Len]
		fp, err := d.fpEsGrpcDB.Get(common.ToHex(ph))
		if err != nil {
			return nil, err
		}
		content.WriteString(fp.Data)
	}
	return NewFile(sum, content.String()), nil
}
