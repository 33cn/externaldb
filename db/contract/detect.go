package contract

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/proto"
	"github.com/33cn/go-kit/convert"
	pabi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	pcom "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

const (
	MethodChain33Version = "Chain33.Version"
	MethodChain33Query   = "Chain33.Query"
	Default20Abi         = "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol_\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"
	Default721Abi        = "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol_\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"approved\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"
	Default1155Abi       = "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"uri_\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"values\",\"type\":\"uint256[]\"}],\"name\":\"TransferBatch\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"TransferSingle\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"URI\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"accounts\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"}],\"name\":\"balanceOfBatch\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"amounts\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeBatchTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"uri\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"
	ERC20                = "ERC20"
	ERC721               = "ERC721"
	ERC1155              = "ERC1155"
	ExecEVM              = "evm"
	FuncNameEstimateGas  = "EstimateGas"
	FuncNameQuery        = "Query"
)

// EstimateGasReq 估算部署交易或者调用交易需要的 gas 请求
type EstimateGasReq struct {
	Tx   string `json:"tx"`   // 部署合约交易或者调用合约交易的序列化后的字符串
	From string `json:"from"` // 合约交易调用者地址
}

// EstimateGasResp 估算部署交易或者调用交易需要的 gas 响应
type EstimateGasResp struct {
	Gas string `json:"gas"` // 估算需要的 gas 数值
}

var Detector *DetectClient
var DefaultABIs map[string]*pabi.ABI

func init() {
	DefaultABIs = make(map[string]*pabi.ABI)
	abi20, err := pabi.JSON(strings.NewReader(Default20Abi))
	if err != nil {
		panic(err)
	}
	DefaultABIs[ERC20] = &abi20

	abi721, err := pabi.JSON(strings.NewReader(Default721Abi))
	if err != nil {
		panic(err)
	}
	DefaultABIs[ERC721] = &abi721

	abi1155, err := pabi.JSON(strings.NewReader(Default1155Abi))
	if err != nil {
		panic(err)
	}
	DefaultABIs[ERC1155] = &abi1155

}

func InitDetector(cfg *proto.ConfigNew) {
	Detector = NewDetectClient(cfg.Chain.Host, cfg.Chain.Title)
}

type DetectClient struct {
	chainVersion *types.VersionInfo
	host         string
	paraName     string
}

func NewDetectClient(host, paraName string) *DetectClient {
	var client DetectClient
	client.host = host
	client.paraName = paraName
	return &client
}

func (c *DetectClient) detect(contract *Contract, standard string, standardAbi string) (bool, error) {
	contractABI, err := pabi.JSON(strings.NewReader(standardAbi))
	if err != nil {
		return false, err
	}
	for name, method := range contractABI.Methods {
		args := make([]interface{}, 0)
		for _, input := range method.Inputs {
			args = append(args, GetDefaultValue(method, &input.Type, contract.Creator))
		}
		data, err := contractABI.Pack(name, args...)
		if err != nil {
			return false, err
		}
		tx := c.CreateCallTx(contract.Address, data)
		var resp evmtypes.EstimateEVMGasResp
		var req evmtypes.EstimateEVMGasReq
		req.Tx = tx
		req.From = contract.Creator
		reqData, _ := json.Marshal(&req)
		ctx := jsonclient.NewRPCCtx(c.host, MethodChain33Query, &rpctypes.Query4Jrpc{
			Execer:   c.paraName + ExecEVM,
			FuncName: FuncNameEstimateGas,
			Payload:  reqData,
		}, &resp)
		if _, err := ctx.RunResult(); err != nil {
			if strings.Contains(err.Error(), standard) {
				continue
			}
			if strings.Contains(err.Error(), "E(MISSING)RC") {
				continue
			}
			return false, fmt.Errorf("detect %s func:%s, err:%s", standard, name, err.Error())
		}
	}
	contract.ContractType = standard
	c.GetContractName(&contractABI, contract)
	c.GetContractSymbol(&contractABI, contract)
	c.GetContractURI(&contractABI, contract)
	return true, nil
}

func (c *DetectClient) GetContractName(contractABI *pabi.ABI, contract *Contract) {
	if !(contract.ContractType == ERC721 || contract.ContractType == ERC20) {
		return
	}

	resp, err := c.ContractCall(contract, contractABI, "name")
	if err != nil {
		log.Error("GetContractSymbol ContractCall", "err", err)
		return
	}
	contract.Name = convert.ToString(resp[""])
}

func (c *DetectClient) GetContractSymbol(contractABI *pabi.ABI, contract *Contract) {
	if !(contract.ContractType == ERC721 || contract.ContractType == ERC20) {
		return
	}

	resp, err := c.ContractCall(contract, contractABI, "symbol")
	if err != nil {
		log.Error("GetContractSymbol ContractCall", "err", err)
		return
	}
	contract.Symbol = convert.ToString(resp[""])
}

func (c *DetectClient) GetContractURI(contractABI *pabi.ABI, contract *Contract) {
	if contract.ContractType != ERC1155 {
		return
	}
	resp, err := c.ContractCall(contract, contractABI, "uri", big.NewInt(0))
	if err != nil {
		log.Error("GetContractSymbol ContractCall", "err", err)
		return
	}
	contract.URI = convert.ToString(resp[""])
}

func (c *DetectClient) ContractCall(contract *Contract, contractABI *pabi.ABI, method string, args ...interface{}) (map[string]interface{}, error) {
	data, err := contractABI.Pack(method, args...)
	if err != nil {
		log.Error("ContractCall contractABI.Pack", "method", method, "err", err)
		return nil, err
	}

	var req evmtypes.EvmQueryReq
	var resp evmtypes.EvmQueryResp
	req.Input = hex.EncodeToString(data)
	req.Address = contract.Address
	reqData, _ := json.Marshal(&req)
	ctx := jsonclient.NewRPCCtx(c.host, MethodChain33Query, &rpctypes.Query4Jrpc{
		Execer:   c.paraName + ExecEVM,
		FuncName: FuncNameQuery,
		Payload:  reqData,
	}, &resp)

	if _, err := ctx.RunResult(); err != nil {
		log.Error("ContractCall client.Query", "method", method, "err", err)
		return nil, err
	}

	res := make(map[string]interface{})
	buf, err := common.FromHex(resp.RawData)
	if err != nil {
		log.Error("ContractCall client.Query", "method", method, "err", err)
		return nil, err
	}

	err = contractABI.UnpackIntoMap(res, method, buf)
	if err != nil {
		log.Error("ContractCall UnpackIntoMap", "method", method, "err", err)
		return nil, err
	}
	return res, nil
}

func (c *DetectClient) Detect(contract *Contract) {
	ok, err := c.detect(contract, ERC20, Default20Abi)
	if err == nil && ok {
		return
	}
	ok, err = c.detect(contract, ERC721, Default721Abi)
	if err == nil && ok {
		return
	}
	ok, err = c.detect(contract, ERC1155, Default1155Abi)
}

func (c *DetectClient) GetChainID() int32 {
	if c.chainVersion != nil {
		return c.chainVersion.ChainID
	}
	var ver types.VersionInfo
	ctx := jsonclient.NewRPCCtx(c.host, MethodChain33Version, nil, &ver)
	if _, err := ctx.RunResult(); err != nil {
		log.Error("Health, get Chain33.Version", "err", err)
	}
	c.chainVersion = &ver
	return c.chainVersion.ChainID
}

func (c *DetectClient) CreateCallTx(contractAddr string, data []byte) string {
	exec := c.paraName + evmtypes.ExecutorName
	toAddr := db.ExecAddress(exec)
	action := evmtypes.EVMContractAction{
		GasLimit:     0,
		GasPrice:     0,
		Code:         nil,
		Para:         data,
		Alias:        "",
		Note:         "",
		ContractAddr: contractAddr,
	}
	tx := &types.Transaction{Execer: []byte(exec), Payload: types.Encode(&action), Fee: 0, To: toAddr}
	random := rand.New(rand.NewSource(time.Now().UnixNano())) // nolint
	tx.Nonce = random.Int63()
	tx.ChainID = c.GetChainID()
	txHex := types.Encode(tx)
	return hex.EncodeToString(txHex)
}

func GetDefaultValue(method pabi.Method, t *pabi.Type, defaultAddr string) interface{} {
	switch t.T {
	case pabi.IntTy, pabi.UintTy:
		return big.NewInt(1)
	case pabi.SliceTy:
		switch t.Elem.T {
		case pabi.IntTy, pabi.UintTy:
			return []*big.Int{big.NewInt(1)}
		case pabi.AddressTy:
			return []pcom.Hash160Address{pcom.StringToAddress(defaultAddr).ToHash160()}
		default:
			panic(t.String())
		}
	case pabi.StringTy:
		return "string"
	case pabi.AddressTy:
		return pcom.StringToAddress(defaultAddr).ToHash160()
	case pabi.BoolTy:
		return true
	case pabi.BytesTy:
		return []byte{1}
	case pabi.FixedBytesTy:
		var ans [4]byte
		copy(ans[:], method.ID)
		return ans
	default:
		panic(t.String())
	}
}
