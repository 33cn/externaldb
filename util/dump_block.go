package util

import (
	"encoding/hex"
	"fmt"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db"
)

// GenEnv gen env for test
func GenEnv(hexBlock string, index int64) (*db.TxEnv, error) {
	bs, err := hex.DecodeString(hexBlock)
	if err != nil {
		return nil, err
	}
	env := db.TxEnv{
		TxIndex: index,
		Block:   &types.BlockDetail{},
	}

	err = types.Decode(bs, env.Block)
	env.BlockHash = common.ToHex(env.Block.Block.HashByForkHeight(3000000))
	return &env, err
}

// DumpEnv dump txs for unit test
func DumpEnv(env *db.TxEnv, check bool) (string, int64) {
	blockByte := types.Encode(env.Block)
	s := hex.EncodeToString(blockByte)

	fmt.Printf("\tblockByte := \"%s\"\n", s)
	fmt.Printf("\tindex := int64(%d)\n", env.TxIndex)
	fmt.Printf("\tenv, err := util.GenEnv(blockByte, index)\n")

	if check {
		env2, err := GenEnv(hex.EncodeToString(blockByte), env.TxIndex)
		if err != nil {
			panic(err)
		}
		s2, index2 := DumpEnv(env2, false)
		if s != s2 || index2 != env.TxIndex {
			panic("not the same")
		}
	}
	return s, env.TxIndex
}
