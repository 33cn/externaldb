package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_isChainTx(t *testing.T) {
	cases := []struct {
		para     string
		exec     string
		is       bool
		realExec string
	}{
		{"", "token", true, "token"},
		{"user.p.gd.", "user.p.gd.token", true, "token"},

		{"user.p.gd.", "user.p.gd.", false, "token"},
		{"", "user.p.gd.token", false, "token"},

		{"bityuan", "token", true, "token"},
		{"local", "user.p.gd.token", false, "token"},
	}

	for _, c := range cases {
		t.Log(c)
		is, exec := isChainTx(c.para, c.exec)
		assert.Equal(t, c.is, is)
		if c.is {
			assert.Equal(t, c.realExec, exec)
		}
	}
}
