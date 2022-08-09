package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Trim(t *testing.T) {
	s := "/xx///"
	s1 := "xx///"
	s2 := "///xx"
	s3 := "xx"
	assert.Equal(t, "xx", withoutSlash(s))
	assert.Equal(t, "xx", withoutSlash(s1))
	assert.Equal(t, "xx", withoutSlash(s2))
	assert.Equal(t, "xx", withoutSlash(s3))
}
