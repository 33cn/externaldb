package txinspect

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/33cn/externaldb/erc20Scaner/erc20abi/generated"
)

func normalizeAddress(address string) string {
	if address == "" {
		return address
	}
	if strings.HasPrefix(address, "0x") || strings.HasPrefix(address, "0X") {
		return "0x" + strings.ToLower(address[2:])
	}
	return strings.ToLower(address)
}

func (a *Analyzer) unPackageAbi(methodName string, cAddress *common.Address) (interface{}, error) {
	parsedAbi, err := abi.JSON(strings.NewReader(generated.ERC20ABI))
	if err != nil {
		return nil, err
	}
	abidata, err := parsedAbi.Pack(methodName)
	if err != nil {
		return nil, err
	}
	var calldata ethereum.CallMsg
	calldata.Data = abidata
	calldata.To = cAddress
	callResult, err := a.c.CallContract(calldata, nil)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = parsedAbi.UnpackIntoInterface(&result, methodName, callResult)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Analyzer) unPackageAbiWithArgs(methodName string, cAddress *common.Address, args ...interface{}) (interface{}, error) {
	parsedAbi, err := abi.JSON(strings.NewReader(generated.ERC20ABI))
	if err != nil {
		return nil, err
	}
	abidata, err := parsedAbi.Pack(methodName, args...)
	if err != nil {
		return nil, err
	}
	var calldata ethereum.CallMsg
	calldata.Data = abidata
	calldata.To = cAddress
	callResult, err := a.c.CallContract(calldata, nil)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = parsedAbi.UnpackIntoInterface(&result, methodName, callResult)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// TokenInfo matches scanner's TokenInfo.
type TokenInfo struct {
	Address  string `json:"address"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals uint8  `json:"decimals"`
}

func (a *Analyzer) getTokenInfo(tokenAddress *common.Address) (*TokenInfo, error) {
	info := &TokenInfo{Address: tokenAddress.Hex()}
	name, err := a.unPackageAbi("name", tokenAddress)
	if err == nil {
		info.Name = fmt.Sprintf("%v", name)
	}
	symbol, err := a.unPackageAbi("symbol", tokenAddress)
	if err == nil {
		info.Symbol = fmt.Sprintf("%v", symbol)
	}
	decimals, err := a.unPackageAbi("decimals", tokenAddress)
	if err == nil {
		switch v := decimals.(type) {
		case uint8:
			info.Decimals = v
		case uint64:
			info.Decimals = uint8(v)
		case *big.Int:
			info.Decimals = uint8(v.Uint64())
		default:
			info.Decimals = 18
		}
	} else {
		info.Decimals = 18
	}
	return info, nil
}

func (a *Analyzer) getBalanceFromContract(contractAddress, holder *common.Address) (*big.Int, error) {
	result, err := a.unPackageAbiWithArgs("balanceOf", contractAddress, *holder)
	if err != nil {
		return nil, err
	}
	var balance *big.Int
	switch v := result.(type) {
	case *big.Int:
		balance = v
	case []interface{}:
		if len(v) > 0 {
			if b, ok := v[0].(*big.Int); ok {
				balance = b
			}
		}
	}
	if balance == nil {
		return big.NewInt(0), nil
	}
	return balance, nil
}
