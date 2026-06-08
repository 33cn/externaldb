// Package txparser provides shared ERC20 event parsing used by both
// scanner (daemon) and txinspect (debug CLI).
package txparser

import (
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/33cn/externaldb/erc20Scaner/erc20abi/generated"
)

// ParsedTransfer represents a raw ERC20 Transfer event extracted from a receipt log.
type ParsedTransfer struct {
	TokenAddress common.Address
	From         common.Address
	To           common.Address
	Value        *big.Int
	LogIndex     uint
}

// ParsedApproval represents a raw ERC20 Approval event extracted from a receipt log.
type ParsedApproval struct {
	TokenAddress common.Address
	Owner        common.Address
	Spender      common.Address
	Value        *big.Int
	LogIndex     uint
}

// NormalizeAddress lowercases an Ethereum address for storage/comparison.
func NormalizeAddress(address string) string {
	if address == "" {
		return address
	}
	if strings.HasPrefix(address, "0x") || strings.HasPrefix(address, "0X") {
		return "0x" + strings.ToLower(address[2:])
	}
	return strings.ToLower(address)
}

var (
	eventIDsOnce    sync.Once
	transferEventID common.Hash
	approvalEventID common.Hash
	eventIDsErr     error
)

func initEventIDs() {
	parsedAbi, err := abi.JSON(strings.NewReader(generated.ERC20ABI))
	if err != nil {
		eventIDsErr = fmt.Errorf("parse ERC20 ABI: %w", err)
		return
	}
	tev := parsedAbi.Events["Transfer"]
	if tev.ID == (common.Hash{}) {
		eventIDsErr = fmt.Errorf("Transfer event not found in ERC20 ABI")
		return
	}
	transferEventID = tev.ID

	aev := parsedAbi.Events["Approval"]
	if aev.ID != (common.Hash{}) {
		approvalEventID = aev.ID
	}
}

// TransferEventID returns the keccak256 topic0 for ERC20 Transfer event.
func TransferEventID() common.Hash {
	eventIDsOnce.Do(initEventIDs)
	return transferEventID
}

// ApprovalEventID returns the keccak256 topic0 for ERC20 Approval event, or
// common.Hash{} if the ABI does not contain the event.
func ApprovalEventID() common.Hash {
	eventIDsOnce.Do(initEventIDs)
	return approvalEventID
}

// ParseReceiptLogs scans receipt logs and returns all ERC20 Transfer and
// Approval events found.  Callers enrich results with token info and handle
// DB persistence or display independently.
func ParseReceiptLogs(logs []*types.Log) ([]ParsedTransfer, []ParsedApproval, error) {
	eventIDsOnce.Do(initEventIDs)
	if eventIDsErr != nil {
		return nil, nil, eventIDsErr
	}

	var transfers []ParsedTransfer
	var approvals []ParsedApproval

	for logIndex, lg := range logs {
		if lg.Removed || len(lg.Topics) < 3 {
			continue
		}

		if lg.Topics[0] == transferEventID {
			// Transfer(address indexed from, address indexed to, uint256 value)
			if len(lg.Data) < 32 {
				continue
			}
			transfers = append(transfers, ParsedTransfer{
				TokenAddress: lg.Address,
				From:         common.BytesToAddress(lg.Topics[1].Bytes()),
				To:           common.BytesToAddress(lg.Topics[2].Bytes()),
				Value:        new(big.Int).SetBytes(lg.Data[:32]),
				LogIndex:     uint(logIndex),
			})
		} else if approvalEventID != (common.Hash{}) && lg.Topics[0] == approvalEventID {
			// Approval(address indexed owner, address indexed spender, uint256 value)
			if len(lg.Data) < 32 {
				continue
			}
			approvals = append(approvals, ParsedApproval{
				TokenAddress: lg.Address,
				Owner:        common.BytesToAddress(lg.Topics[1].Bytes()),
				Spender:      common.BytesToAddress(lg.Topics[2].Bytes()),
				Value:        new(big.Int).SetBytes(lg.Data[:32]),
				LogIndex:     uint(logIndex),
			})
		}
	}

	return transfers, approvals, nil
}
