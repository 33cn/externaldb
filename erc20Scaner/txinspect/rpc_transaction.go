package txinspect

// rpcTransaction mirrors go-ethereum ethclient.rpcTransaction for JSON decoding.

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type rpcTransaction struct {
	tx *types.Transaction
	txExtraInfo
}

type txExtraInfo struct {
	BlockNumber *string         `json:"blockNumber,omitempty"`
	BlockHash   *common.Hash    `json:"blockHash,omitempty"`
	From        *common.Address `json:"from,omitempty"`
}

func (tx *rpcTransaction) UnmarshalJSON(msg []byte) error {
	if err := json.Unmarshal(msg, &tx.tx); err != nil {
		return err
	}
	return json.Unmarshal(msg, &tx.txExtraInfo)
}

func decodeTransactionByHashResult(raw json.RawMessage) (*types.Transaction, bool, error) {
	r := bytesTrimSpace(raw)
	if len(r) == 0 || string(r) == "null" {
		return nil, false, nil
	}
	var j rpcTransaction
	if err := json.Unmarshal(raw, &j); err != nil {
		return nil, false, err
	}
	if j.tx == nil {
		return nil, false, nil
	}
	if _, r, _ := j.tx.RawSignatureValues(); r == nil {
		return nil, false, fmt.Errorf("server returned transaction without signature")
	}
	if j.From != nil && j.BlockHash != nil {
		setSenderFromServer(j.tx, *j.From, *j.BlockHash)
	}
	isPending := j.BlockNumber == nil
	return j.tx, isPending, nil
}

func bytesTrimSpace(b json.RawMessage) []byte {
	// avoid importing bytes for one helper
	i, j := 0, len(b)
	for i < j && (b[i] == ' ' || b[i] == '\t' || b[i] == '\n' || b[i] == '\r') {
		i++
	}
	for j > i && (b[j-1] == ' ' || b[j-1] == '\t' || b[j-1] == '\n' || b[j-1] == '\r') {
		j--
	}
	return b[i:j]
}
