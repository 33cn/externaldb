package db

import "errors"

// DB operate FileSummary
type DB interface {
	Get(hash string) (*FileSummary, error)
}

var (
	ErrInvalidOperate     = errors.New("ErrInvalidOperate")
	ErrFileBlacklisted    = errors.New("ErrFileBlacklisted")
	ErrFileNotBlacklisted = errors.New("ErrFileNotBlacklisted")
)
