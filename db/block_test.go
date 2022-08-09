package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcParaTitle(t *testing.T) {
	assert.Equal(t, CalcParaTitle("user.p.xx."), "user.p.xx.")
	assert.Equal(t, CalcParaTitle("bityuan"), "")
}
