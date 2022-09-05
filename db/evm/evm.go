package evm

import (
	"strings"

	"github.com/33cn/externaldb/db"
	pabi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	pcom "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
)

var funcConverts = make(map[string]FuncConvert)
var eventHandlers = make(map[string]EventHandle)
var initDBs = make([]InitDB, 0)

type FuncConvert func(*Convert, int, EVM) ([]db.Record, error)
type InitDB func(cli db.DBCreator) error

type EventHandle func(*Convert, int, EVM) ([]db.Record, error)

func RegisterFuncConvert(name string, convert FuncConvert) {
	name = UpperFirstLetter(name)
	if _, ok := funcConverts[name]; ok {
		panic("double register function convert:" + name)
	}
	funcConverts[name] = convert
}

func GetFuncConvert(name string) (FuncConvert, bool) {
	name = UpperFirstLetter(name)
	fc, ok := funcConverts[name]
	return fc, ok
}

func UpperFirstLetter(name string) string {
	if len(name) > 0 {
		name = strings.ToTitle(name[:1]) + name[1:]
	}
	return name
}

func RegisterEventHandle(name string, convert EventHandle) {
	if _, ok := eventHandlers[name]; ok {
		panic("double register function convert:" + name)
	}
	eventHandlers[name] = convert
}

func GetEventHandle(name string) (EventHandle, bool) {
	fc, ok := eventHandlers[name]
	return fc, ok
}

func RegisterInitDB(initDB InitDB) {
	initDBs = append(initDBs, initDB)
}

func UnpackEvent(data []byte, topics []pcom.Hash, abi *pabi.ABI) (name string, ans map[string]interface{}, err error) {
	if len(topics) <= 0 || abi == nil {
		return
	}
	ans = make(map[string]interface{})
	// non index arguments
	evt, arg, err := UnpackEventData(data, topics[0], abi)
	if err != nil {
		return
	}
	name = evt.Name
	nonIndexed := evt.Inputs.NonIndexed()
	for i, v := range arg {
		ans[nonIndexed[i].Name] = v
		if nonIndexed[i].Type.T == pabi.AddressTy {
			if addr, ok := v.(pcom.Hash160Address); ok {
				ans[nonIndexed[i].Name] = db.Hash160AddressToString(addr)
			}
		}
	}

	// index arguments
	as := make(pabi.Arguments, 0)
	for _, a := range evt.Inputs {
		if a.Indexed {
			as = append(as, a)
		}
	}
	err = pabi.ParseTopicsIntoMap(ans, as, topics[1:])
	if err != nil {
		return
	}
	for k, v := range ans {
		if addr, ok := v.(pcom.Hash160Address); ok {
			ans[k] = db.Hash160AddressToString(addr)
		}
	}
	return
}

// UnpackEventData use abi and topic unpack event data
func UnpackEventData(data []byte, topic pcom.Hash, abi *pabi.ABI) (event *pabi.Event, args []interface{}, err error) {
	event, err = abi.EventByID(topic)
	if err != nil {
		return
	}
	args, err = event.Inputs.UnpackValues(data)
	return
}

func UnpackParam(data []byte, abi *pabi.ABI) (args map[string]interface{}, err error) {
	method, arg, err := UnpackParamData(data, abi)
	if err != nil {
		return
	}
	args = make(map[string]interface{})
	args["call_func_name"] = method.Name
	nonIndexed := method.Inputs.NonIndexed()
	for i, v := range arg {
		args[method.Inputs[i].Name] = v
		if nonIndexed[i].Type.T == pabi.AddressTy {
			if addr, ok := v.(pcom.Hash160Address); ok {
				args[nonIndexed[i].Name] = db.Hash160AddressToString(addr)
			}
		}
	}
	return
}

// UnpackParamData 通过abi解析调用方法和输入参数
func UnpackParamData(data []byte, abi *pabi.ABI) (method *pabi.Method, args []interface{}, err error) {
	method, err = abi.MethodByID(data[:4])
	if err != nil {
		return
	}
	args, err = method.Inputs.UnpackValues(data[4:])
	return
}
