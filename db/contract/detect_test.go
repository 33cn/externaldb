package contract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testChainJRpcHost = "http://183.134.99.140:8901"
	testParaName      = "user.p.testproofv2."
)

func TestDetectClient_Detect20(t *testing.T) {
	client := NewDetectClient(testChainJRpcHost, testParaName)
	erc20 := &Contract{
		Address: "16fYTUZgHX4Tj7cUAY3tHMubJfeHiTpZ9h",
		Creator: "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY",
	}
	client.Detect(erc20)
	assert.Equal(t, ERC20, erc20.ContractType)
	t.Log(erc20.Symbol, erc20.Name)
}

func TestDetectClient_Detect721(t *testing.T) {
	client := NewDetectClient(testChainJRpcHost, testParaName)

	erc721 := &Contract{
		Address: "1DJmxndBVTgSX4Cei2NabJfbNZcq8rGDSN",
		Creator: "133AfuMYQXRxc45JGUb1jLk1M1W4ka39L1",
	}
	client.Detect(erc721)
	assert.Equal(t, ERC721, erc721.ContractType)
	t.Log(erc721.Symbol, erc721.Name)
}

func TestDetectClient_Detect1155(t *testing.T) {
	client := NewDetectClient(testChainJRpcHost, testParaName)

	erc1155 := &Contract{
		Address: "1NY7MMMha2wajBaoGLanxn68eqwnXKCWVC",
		Creator: "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY",
	}
	client.Detect(erc1155)
	assert.Equal(t, ERC1155, erc1155.ContractType)
	t.Log(erc1155.URI)
}
