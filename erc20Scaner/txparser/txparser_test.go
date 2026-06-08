package txparser

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestIsUnlimitedApproval(t *testing.T) {
	if !IsUnlimitedApproval(MaxUint256) {
		t.Error("MaxUint256 should be unlimited")
	}
	if IsUnlimitedApproval(nil) {
		t.Error("nil should not be unlimited")
	}
	if IsUnlimitedApproval(big.NewInt(0)) {
		t.Error("0 should not be unlimited")
	}
	if IsUnlimitedApproval(big.NewInt(1000000000000000000)) {
		t.Error("10^18 should not be unlimited")
	}
}

func TestNormalizeAddress(t *testing.T) {
	tests := []struct{ in, want string }{
		{"0xABC", "0xabc"},
		{"0xAbC123", "0xabc123"},
		{"0XABC", "0xabc"},
		{"", ""},
		{"0x", "0x"},
	}
	for _, tt := range tests {
		got := NormalizeAddress(tt.in)
		if got != tt.want {
			t.Errorf("NormalizeAddress(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestEventIDs(t *testing.T) {
	expectedTransfer := "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	if got := TransferEventID().Hex(); got != expectedTransfer {
		t.Errorf("TransferEventID = %s, want %s", got, expectedTransfer)
	}
	expectedApproval := "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"
	if got := ApprovalEventID().Hex(); got != expectedApproval {
		t.Errorf("ApprovalEventID = %s, want %s", got, expectedApproval)
	}
}

func TestParseReceiptLogs(t *testing.T) {
	transferHash := TransferEventID()
	approvalHash := ApprovalEventID()

	logs := []*types.Log{
		// Transfer event
		{
			Address: common.HexToAddress("0xtoken1"),
			Topics: []common.Hash{
				transferHash,
				common.BytesToHash(common.HexToAddress("0xfrom1").Bytes()),
				common.BytesToHash(common.HexToAddress("0xto1").Bytes()),
			},
			Data:   common.Hex2Bytes("0000000000000000000000000000000000000000000000000de0b6b3a7640000"),
			Removed: false,
		},
		// Approval event
		{
			Address: common.HexToAddress("0xtoken2"),
			Topics: []common.Hash{
				approvalHash,
				common.BytesToHash(common.HexToAddress("0xowner").Bytes()),
				common.BytesToHash(common.HexToAddress("0xspender").Bytes()),
			},
			Data:   common.Hex2Bytes("0000000000000000000000000000000000000000000000000de0b6b3a7640000"),
			Removed: false,
		},
		// Removed log (should be skipped)
		{
			Address: common.HexToAddress("0xbad"),
			Topics:  []common.Hash{transferHash, {}, {}},
			Data:    common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000001"),
			Removed: true,
		},
	}

	transfers, approvals, err := ParseReceiptLogs(logs)
	if err != nil {
		t.Fatalf("ParseReceiptLogs: %v", err)
	}
	if len(transfers) != 1 {
		t.Errorf("expected 1 transfer, got %d", len(transfers))
	}
	if len(approvals) != 1 {
		t.Errorf("expected 1 approval, got %d", len(approvals))
	}

	if transfers[0].LogIndex != 0 {
		t.Errorf("transfer LogIndex = %d, want 0", transfers[0].LogIndex)
	}
	if approvals[0].LogIndex != 1 {
		t.Errorf("approval LogIndex = %d, want 1", approvals[0].LogIndex)
	}

	expectedVal := big.NewInt(1000000000000000000) // 10^18
	if transfers[0].Value.Cmp(expectedVal) != 0 {
		t.Errorf("transfer value = %s, want %s", transfers[0].Value.String(), expectedVal.String())
	}
}

func TestMaxUint256(t *testing.T) {
	// Verify MaxUint256 = 2^256 - 1
	one := big.NewInt(1)
	maxPlusOne := new(big.Int).Add(MaxUint256, one)
	expected := new(big.Int).Lsh(one, 256)
	if maxPlusOne.Cmp(expected) != 0 {
		t.Error("MaxUint256 + 1 != 2^256")
	}
	// Verify it's not negative
	if MaxUint256.Sign() <= 0 {
		t.Error("MaxUint256 should be positive")
	}
}
