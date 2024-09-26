package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"encoding/json"

	"github.com/33cn/chain33/common"
	"github.com/stretchr/testify/assert"
)

type testAbiSaveRequst struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

func TestSaveEvmAbi2(t *testing.T) {
	url := "https://mainnet.bityuan.com/btydata/bityuan"
	address := LuckyPackageContract
	abi := LuckyPakcageAbi
	doSaveEvmAbi(t, url, address, abi)
}

func TestSaveEvmAbi22(t *testing.T) {
	url := "https://mainnet.bityuan.com/btydata/bityuan"
	address := LuckyPackageContract2
	abi := LuckyPakcageAbi2
	doSaveEvmAbi(t, url, address, abi)
}

func TestSaveEvmAbi23(t *testing.T) {
	url := "https://mainnet.bityuan.com/btydata/bityuan"
	address := LuckyPackageContract3
	abi := LuckyPakcageAbi3
	doSaveEvmAbi(t, url, address, abi)
}

func doSaveEvmAbi(t *testing.T, url string, address string, abi string) {

	param := SaveAbiRequest{
		Address: address,
		Abi:     common.ToHex([]byte(abi)),
	}
	req := testAbiSaveRequst{
		Method: "Evm.SaveAbi",
		Params: []interface{}{param},
	}

	bs, err := json.Marshal(req)
	assert.Nil(t, err)
	assert.Equal(t, "", string(bs))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bs))
	assert.Nil(t, err)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, "", string(body))
}

func TestSaveEvmAbi(t *testing.T) {
	url := "http://183.134.99.137:9993/testdex"
	param := SaveAbiRequest{
		Address: "0x639874a1978065ea394a444f032400655ed55e7b",
		Abi:     common.ToHex([]byte(testAbi)),
	}
	req := testAbiSaveRequst{
		Method: "Evm.SaveAbi",
		Params: []interface{}{param},
	}

	bs, err := json.Marshal(req)
	assert.Nil(t, err)
	assert.Equal(t, "", string(bs))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bs))
	assert.Nil(t, err)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, "", string(body))
}

func TestRpcParseDeployTx(t *testing.T) {
	txHash := "0xcbcd45ffaf7e84338f72f1909e3afc8dcdd7b4b4292fa45739442214d196d073"
	testRpcParseTx(t, txHash)
}

func TestRpcParseTransferTx(t *testing.T) {
	txHash := "0xd189267a13d9101f75d639d58ba1ab8a6fffee129f8868afdd5b40db7f1d4635"
	testRpcParseTx(t, txHash)
}

func TestRpcParseEvmCallTx(t *testing.T) {
	txHash := "0xcc4d63d76c6f31b63e90d46dfba08b421da8aa23648054107c5f54b4cf95f1c5"
	testRpcParseTx(t, txHash)
}

func testRpcParseTx(t *testing.T, txHash string) {
	url := "http://183.134.99.137:9993/testdex"

	param := ParseEvmTxRequest{
		TxHash: txHash,
	}
	req := testAbiSaveRequst{
		Method: "Evm.ParseTx",
		Params: []interface{}{param},
	}

	bs, err := json.Marshal(req)
	assert.Nil(t, err)
	assert.Equal(t, "", string(bs))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bs))
	assert.Nil(t, err)
	if err != nil {
		assert.Nil(t, err.Error())
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, "", string(body))
}
