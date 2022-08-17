package db

import (
	"fmt"
	"strings"

	"github.com/33cn/chain33/common/address"
	pcom "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	lru "github.com/hashicorp/golang-lru"
)

var (
	DefaultAddrID int32
	execAddrCache *lru.Cache
)

const (
	NormalAddressID = 0
	EthAddressID    = 2
)

func init() {
	SetAddrID("btc")
}

func SetAddrID(addrDriver string) {
	var err error
	execAddrCache, err = lru.New(10240)
	if err != nil {
		panic(err)
	}

	switch addrDriver {
	case "":
		DefaultAddrID = NormalAddressID
	case "eth":
		DefaultAddrID = EthAddressID
	case "btc":
		DefaultAddrID = NormalAddressID
	default:
		panic("not support addrDriver" + fmt.Sprint(addrDriver))
	}
}

// CalcParaTitle 计算exec的前缀
func CalcParaTitle(title string) string {
	if strings.HasPrefix(title, "user.p.") {
		return title
	}
	return ""
}

// Hash160AddressToString 将Hash160格式的地址转换为string格式
func Hash160AddressToString(addr pcom.Hash160Address) string {
	if DefaultAddrID == EthAddressID {
		return addr.String()
	}
	return addr.ToAddress().String()
}

//ExecAddress 获得执行器地址
func ExecAddress(name string) string {
	if value, ok := execAddrCache.Get(name); ok {
		return value.(string)
	}
	addr, _ := address.GetExecAddress(name, DefaultAddrID)
	execAddrCache.Add(name, addr)
	return addr
}
