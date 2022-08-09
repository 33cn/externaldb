package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_configV165(t *testing.T) {
	title := "bityuan"
	symbol := "bbbbty"
	cfg := `log=aaa
`
	expact := `Title="bityuan"
CoinSymbol="bbbbty"
log=aaa
`

	result := configV165(title, symbol, cfg)
	assert.Equal(t, expact, result)
}
