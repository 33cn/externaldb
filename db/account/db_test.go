package account

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_tojons(t *testing.T) {
	acc := &Account{Detall: &Detall{Frozen: 1}, AssetSymbol: "b"}
	v, _ := json.Marshal(acc)
	//t.Error(string(v))
	assert.Equal(t, `{"address":"","exec":"","frozen":1,"balance":0,"total":0,"type":"","height_index":0,"asset_symbol":"b","asset_exec":""}`, string(v))
}

func Test_AddExecAddress(t *testing.T) {
	execes := []string{"ticket", "trade"}
	addressBty := []string{"16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp", "1BXvgjmBw1aBgmGn1hjfGyRkmN3krWpFP4"}
	// title = "user.p.para."
	addressPara := []string{"1EZrEKPPC36SLRoLQBwLDjzcheiLRZJg49", "12bihjzbaYWjcpDiiy9SuAWeqNksQdiN13"}
	Init("bityuan", execes)
	for _, a := range addressBty {
		_, found := execAddress[a]
		assert.True(t, found)
	}
	Init("user.p.para.", execes)
	for _, a := range addressPara {
		_, found := execAddress[a]
		assert.True(t, found)
	}
}
