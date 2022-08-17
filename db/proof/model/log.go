package model

// Log 删除和恢复的记录
type Log struct {
	ID        string
	ProofHash string
	Op        string
	Note      string
	Force     bool
	Address   string

	Height    int64
	Index     int64
	BlockTime int64
	BlockHash string
}

// LogID LogID
const LogID = "log"
