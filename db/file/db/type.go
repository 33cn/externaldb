package db

import (
	fsdb "github.com/33cn/externaldb/db/filesummary/db"
)

// File 文件
type File struct {
	Data     string `json:"file_data"`
	TxHash   string `json:"file_sum_tx_hash"`
	FileType string `json:"file_type"`
	FileHash string `json:"file_hash"`
	FileSize int64  `json:"file_size"`
}

// NewFile new File
func NewFile(sum *fsdb.FileSummary, content string) *File {
	return &File{
		Data:     content,
		TxHash:   sum.TxHash,
		FileType: sum.FileType,
		FileHash: sum.FileHash,
		FileSize: sum.FileSize,
	}
}
