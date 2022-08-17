package rpcutils

import "github.com/pkg/errors"

var (
	ErrBadParam   = errors.New("Bad Param")
	ErrNotFound   = errors.New("Not Found")
	ErrTypeAsset  = errors.New("Type Asset failed")
	ErrSearchSize = errors.New("Search Return Size not match")
)
