package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// AddressConvert 对地址进行转化
// eth 格式地址, 都转化成小写格式
func TestAddressConvert(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{"1", "1"},
		{"0xAA", "0xaa"},
		{"0xA21410b54AdB3B3CCEb789607520c0c8D7A603Ef", "0xa21410b54adb3b3cceb789607520c0c8d7a603ef"},
		{"143apHcTTVN8nhP8JaZ27ZYVh4Zy5VVbDc", "143apHcTTVN8nhP8JaZ27ZYVh4Zy5VVbDc"},
	}

	for _, c := range cases {
		r := AddressConvert(c.input)
		assert.Equal(t, c.output, r)
	}

}
