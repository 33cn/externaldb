package evm

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/33cn/chain33/common"
	pabi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	pcom "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ecom "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/33cn/externaldb/db"
)

var DefaultNFTABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"values\",\"type\":\"uint256[]\"}],\"name\":\"TransferBatch\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"TransferSingle\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"URI\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"goodsID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"int32\",\"name\":\"goodsType\",\"type\":\"int32\"},{\"indexed\":false,\"internalType\":\"int64\",\"name\":\"publishTime\",\"type\":\"int64\"},{\"indexed\":false,\"internalType\":\"string[]\",\"name\":\"hash\",\"type\":\"string[]\"},{\"indexed\":false,\"internalType\":\"string[]\",\"name\":\"source\",\"type\":\"string[]\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"publisher\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"labelID\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"remark\",\"type\":\"string\"}],\"name\":\"addNewGoodsResult\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"goodsID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"balance\",\"type\":\"uint256\"}],\"indexed\":false,\"internalType\":\"struct ProcessingFactory.AccountBalance[]\",\"name\":\"balanceList\",\"type\":\"tuple[]\"}],\"name\":\"balanceResult\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"amounts\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"batchTransferResult\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"amounts\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"BatchTransferWithEvent\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"int32\",\"name\":\"goodsType\",\"type\":\"int32\"},{\"internalType\":\"int64\",\"name\":\"publishTime\",\"type\":\"int64\"},{\"internalType\":\"string[]\",\"name\":\"hash\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"source\",\"type\":\"string[]\"},{\"internalType\":\"string\",\"name\":\"publisher\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"labelID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"remark\",\"type\":\"string\"}],\"name\":\"addNewGoods\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"goodsID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"int32\",\"name\":\"goodsType\",\"type\":\"int32\"},{\"internalType\":\"int64\",\"name\":\"publishTime\",\"type\":\"int64\"},{\"internalType\":\"string[]\",\"name\":\"hash\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"source\",\"type\":\"string[]\"},{\"internalType\":\"string\",\"name\":\"publisher\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"labelID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"remark\",\"type\":\"string\"}],\"name\":\"addNewGoodsAssignGoodsID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"accounts\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"}],\"name\":\"balanceOfBatch\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"goodsID\",\"type\":\"uint256\"}],\"name\":\"getGoodsAttribute\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"goodsID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"int32\",\"name\":\"goodsType\",\"type\":\"int32\"},{\"internalType\":\"int64\",\"name\":\"publishTime\",\"type\":\"int64\"},{\"internalType\":\"string\",\"name\":\"publisher\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"labelID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"remark\",\"type\":\"string\"}],\"internalType\":\"struct ProcessingFactory.Goods\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"goodsID\",\"type\":\"uint256\"}],\"name\":\"getGoodsHash\",\"outputs\":[{\"internalType\":\"string[]\",\"name\":\"\",\"type\":\"string[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMaxGoodsID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"goodsID\",\"type\":\"uint256\"}],\"name\":\"getsourceHash\",\"outputs\":[{\"internalType\":\"string[]\",\"name\":\"\",\"type\":\"string[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"goodsHash\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"maxGoodsId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"amounts\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeBatchTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"sourceHash\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"uri\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

func TestUnpackEvent(t *testing.T) {
	// 0x3aae0a63ecafcb0c4796293d9d4d45ab2fab401e4242ca2a3673abb94aa0bb03
	topics := []string{"0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62",
		"0x000000000000000000000000a9f7f8d22ad5b914b85283fc9db1f2264fa09862",
		"0x0000000000000000000000000000000000000000000000000000000000000000",
		"0x000000000000000000000000a9f7f8d22ad5b914b85283fc9db1f2264fa09862"}
	data := "0x000000000000000000000000000000000000000000000000000000000000000d0000000000000000000000000000000000000000000000000000000000000002"
	testUnpackEvent(t, data, topics)

	t.Log("")
	topics = []string{"0x66f8dce34b1b463d11787b0740c48f78f1623e4626f69387af0cd47ebed58f2d"}
	data = "0x000000000000000000000000a9f7f8d22ad5b914b85283fc9db1f2264fa09862000000000000000000000000000000000000000000000000000000000000000d00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000006178f4ca00000000000000000000000000000000000000000000000000000000000001600000000000000000000000000000000000000000000000000000000000000220000000000000000000000000000000000000000000000000000000000000028000000000000000000000000000000000000000000000000000000000000002e0000000000000000000000000000000000000000000000000000000000000032000000000000000000000000000000000000000000000000000000000000003600000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000423078653266643937343338373963633036363331636565666463636236323465343365643366316561326330663866366631666664383634363566376338396436300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000021e69dade5b79ee5a48de69d82e7be8ee7a791e68a80e69c89e99990e585ace58fb800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ce6b58be8af95e6b58be8af950000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000e534c43303030312d4654303031330000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	testUnpackEvent(t, data, topics)
}

func testUnpackEvent(t *testing.T, data string, topics []string) {
	eData := pcom.FromHex(data)
	var hashs []pcom.Hash
	for _, topic := range topics {
		hashs = append(hashs, pcom.BytesToHash(pcom.FromHex(topic)))
	}
	contractABI, err := pabi.JSON(strings.NewReader(DefaultNFTABI))
	assert.Nil(t, err)
	name, args, err := UnpackEvent(eData, hashs, &contractABI)
	assert.Nil(t, err)
	t.Log("eventName:", name)
	printMap(t, args)
}

func printMap(t *testing.T, m map[string]interface{}) {
	for k, v := range m {
		if m, ok := v.(map[string]interface{}); ok {
			t.Log(k + ":")
			printMap(t, m)
		} else if s, ok := v.([]map[string]interface{}); ok {
			t.Log(k + ":")
			for _, m := range s {
				printMap(t, m)
			}
		} else if s, ok := v.([]interface{}); ok {
			t.Log(k + ":")
			for _, m := range s {
				t.Log(m)
			}
		} else {
			t.Log(k, v, reflect.TypeOf(v))
		}
	}
	if balanceList, ok := m["balanceList"]; ok {
		var res = make([]AccountBalance, 0)
		infoByte, _ := json.Marshal(balanceList)
		_ = json.Unmarshal(infoByte, &res)
		for _, b := range res {
			t.Log("Account", db.Hash160AddressToString(b.Account))
			t.Log("GoodsID", b.GoodsID)
			t.Log("Balance", b.Balance)
		}
	}
}

type AccountBalance struct {
	Account pcom.Hash160Address `json:"account"`
	GoodsID int64               `json:"goodsID"`
	Balance int64               `json:"balance"`
}

func TestUnpackParam(t *testing.T) {
	// http://parallel.bityuan.com/tradeHash?hash=0x575307f7eadce7844ab8981219ed56fc9ec1f8b36949b01552423a1b15328ad9
	// http://121.40.18.70:8885/llq/prove?hash=0x575307f7eadce7844ab8981219ed56fc9ec1f8b36949b01552423a1b15328ad9
	param := "0x033a3727000000000000000000000000a9f7f8d22ad5b914b85283fc9db1f2264fa09862000000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000006177f00a00000000000000000000000000000000000000000000000000000000000001600000000000000000000000000000000000000000000000000000000000000220000000000000000000000000000000000000000000000000000000000000028000000000000000000000000000000000000000000000000000000000000002e0000000000000000000000000000000000000000000000000000000000000032000000000000000000000000000000000000000000000000000000000000003600000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000423078326231616462386435316130663739333963653032343332303463353532636637633634616230616233333732663064313330343530303830323432336666360000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000021e69dade5b79ee5a48de69d82e7be8ee7a791e68a80e69c89e99990e585ace58fb8000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000009e697a0e68980e8b0930000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f534c43303030312d4e46543030313200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	eParam, err := common.FromHex(param)
	assert.Nil(t, err)
	contractABI, err := pabi.JSON(strings.NewReader(DefaultNFTABI))
	assert.Nil(t, err)
	args, err := UnpackParam(eParam, &contractABI)
	assert.Nil(t, err)
	for k, v := range args {
		t.Log(k, v)
	}
}

func TestUnpackParaByEthereum(t *testing.T) {
	// http://parallel.bityuan.com/tradeHash?hash=0x575307f7eadce7844ab8981219ed56fc9ec1f8b36949b01552423a1b15328ad9
	// http://121.40.18.70:8885/llq/prove?hash=0x575307f7eadce7844ab8981219ed56fc9ec1f8b36949b01552423a1b15328ad9
	var param = "0x033a3727000000000000000000000000a9f7f8d22ad5b914b85283fc9db1f2264fa09862000000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000006177f00a00000000000000000000000000000000000000000000000000000000000001600000000000000000000000000000000000000000000000000000000000000220000000000000000000000000000000000000000000000000000000000000028000000000000000000000000000000000000000000000000000000000000002e0000000000000000000000000000000000000000000000000000000000000032000000000000000000000000000000000000000000000000000000000000003600000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000423078326231616462386435316130663739333963653032343332303463353532636637633634616230616233333732663064313330343530303830323432336666360000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000021e69dade5b79ee5a48de69d82e7be8ee7a791e68a80e69c89e99990e585ace58fb8000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000009e697a0e68980e8b0930000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f534c43303030312d4e46543030313200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	mp := make(map[string]interface{})
	eventData, err := common.FromHex(param)
	assert.Nil(t, err)
	err = UnpackParaByEthereum(eventData, mp)
	assert.Nil(t, err)
	for k, v := range mp {
		t.Log(k, v)
	}
}

func TestUnpackEventByEthereum(t *testing.T) {
	topic := "0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62"
	var edata = "0x0000000000000000000000000000000000000000000000000000017ca20beb9b00000000000000000000000000000000000000000000000000000000000022b8"
	mp2 := make(map[string]interface{})
	eData, err := common.FromHex(edata)
	err = UnpackEventByEthereum(topic, eData, mp2)
	assert.Nil(t, err)
	for k, v := range mp2 {
		t.Log(k, v)
	}
}

func UnpackParaByEthereum(data []byte, m map[string]interface{}) error {
	contractABI, err := abi.JSON(strings.NewReader(DefaultNFTABI))
	if err != nil {
		log.Error("parseParam: abi.JSON", "err", err)
		return err
	}
	method, err := contractABI.MethodById(data[:4])
	if err != nil {
		log.Error("parseParam: contractABI.MethodById", "err", err)
		return err
	}
	m["call_func_name"] = method.Name
	arg, err := method.Inputs.UnpackValues(data[4:])
	if err != nil {
		log.Error("parseParam: method.Inputs.Unpack", "err", err)
		return err
	}
	nonIndexed := method.Inputs.NonIndexed()
	for i, v := range arg {
		m[nonIndexed[i].Name] = v
		if nonIndexed[i].Type.T == abi.AddressTy {
			if addr, ok := v.(ecom.Address); ok {
				m[nonIndexed[i].Name] = pcom.BytesToAddress(addr.Bytes()).String()
			}
		}
	}
	return nil
}

func UnpackEventByEthereum(topic string, data []byte, m map[string]interface{}) error {
	contractABI, err := abi.JSON(strings.NewReader(DefaultNFTABI))
	if err != nil {
		log.Error("parseParam: abi.JSON", "err", err)
		return err
	}
	event, err := contractABI.EventByID(ecom.HexToHash(topic))
	if err != nil {
		log.Error("parseParam: contractABI.EventByID", "err", err)
		return err
	}
	m["event_name"] = event.Name
	arg, err := event.Inputs.UnpackValues(data)
	if err != nil {
		log.Error("parseParam: method.Inputs.Unpack", "err", err)
		return err
	}
	nonIndexed := event.Inputs.NonIndexed()
	for i, v := range arg {
		fmt.Println("deb", event.Inputs[i].Name, event.Inputs[i].Type.T, v)
		m[nonIndexed[i].Name] = v
		if event.Inputs[i].Type.T == abi.AddressTy {
			if addr, ok := v.(ecom.Address); ok {
				m[nonIndexed[i].Name] = pcom.BytesToAddress(addr.Bytes()).String()
			}
		}
	}
	return nil
}
