package converts

import (
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/stat"
)

var (
	execConverts = make(map[string]CreateFunc)
	statConverts = make(map[string]StatCreateFunc)
)

// CreateFunc create ExecConvert func point
type CreateFunc func(title, symbol string, supports []string) db.ExecConvert

// StatCreateFunc create stat
type StatCreateFunc func(title, symbol string, genesis, coin int64) stat.Stat

// Register Register
func Register(name string, create CreateFunc) {
	execConverts[name] = create
}

// Load find exec
func Load(name string) (driver CreateFunc, err error) {
	c, ok := execConverts[name]
	if !ok {
		return nil, nil
	}

	return c, nil
}

// RegisterStat Register
func RegisterStat(name string, create StatCreateFunc) {
	statConverts[name] = create
}

// LoadStat find exec
func LoadStat(name string) (driver StatCreateFunc, err error) {
	c, ok := statConverts[name]
	if !ok {
		return nil, nil
	}

	return c, nil
}
