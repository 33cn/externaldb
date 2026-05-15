package blockalign

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/trie"
)

func mustKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	k, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func signedLegacy(t *testing.T, key *ecdsa.PrivateKey, chainID int64, nonce uint64, to common.Address) *types.Transaction {
	t.Helper()
	tx := types.NewTransaction(nonce, to, big.NewInt(0), 21000, big.NewInt(1e9), nil)
	signed, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(chainID)), key)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func TestAlignEthTxsByNonce_skipsNonEvmSlots(t *testing.T) {
	key := mustKey(t)
	from := crypto.PubkeyToAddress(key.PublicKey)
	to := common.HexToAddress("0x0000000000000000000000000000000000000001")
	chainID := int64(1)

	txA := signedLegacy(t, key, chainID, 10, to)
	txB := signedLegacy(t, key, chainID, 11, to)

	block := types.NewBlock(
		&types.Header{Number: big.NewInt(1)},
		[]*types.Transaction{txA, txB},
		nil,
		nil,
		trie.NewStackTrie(nil),
	)

	expects := []SlotEthExpect{
		{NeedEth: false},
		{NeedEth: true, From: from.Hex(), Nonce: 10},
		{NeedEth: true, From: from.Hex(), Nonce: 11},
	}
	out, err := AlignEthTxsByNonce(block, expects)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 3 {
		t.Fatalf("len=%d", len(out))
	}
	if out[0] != nil {
		t.Fatalf("slot0 want nil non-evm, got %v", out[0])
	}
	if out[1] == nil || out[1].Hash() != txA.Hash() {
		t.Fatalf("slot1 want txA")
	}
	if out[2] == nil || out[2].Hash() != txB.Hash() {
		t.Fatalf("slot2 want txB")
	}
}

func TestAlignEthTxsByNonce_missingBodyIsNil(t *testing.T) {
	key := mustKey(t)
	from := crypto.PubkeyToAddress(key.PublicKey)
	to := common.HexToAddress("0x0000000000000000000000000000000000000001")
	chainID := int64(1)

	txB := signedLegacy(t, key, chainID, 11, to)
	block := types.NewBlock(
		&types.Header{Number: big.NewInt(2)},
		[]*types.Transaction{txB},
		nil,
		nil,
		trie.NewStackTrie(nil),
	)

	expects := []SlotEthExpect{
		{NeedEth: true, From: from.Hex(), Nonce: 10},
		{NeedEth: true, From: from.Hex(), Nonce: 11},
	}
	out, err := AlignEthTxsByNonce(block, expects)
	if err != nil {
		t.Fatal(err)
	}
	if out[0] != nil {
		t.Fatalf("missing nonce 10 should be nil")
	}
	if out[1] == nil || out[1].Hash() != txB.Hash() {
		t.Fatalf("want txB at index 1")
	}
}

// 区块体内连续两笔相同 nonce、不同发送方时，应用 expect.From 选中第二笔。
func TestAlignEthTxsByNonce_consecutiveSameNonceUsesFromTiebreaker(t *testing.T) {
	keyA := mustKey(t)
	keyB := mustKey(t)
	fromA := crypto.PubkeyToAddress(keyA.PublicKey)
	fromB := crypto.PubkeyToAddress(keyB.PublicKey)
	to := common.HexToAddress("0x0000000000000000000000000000000000000002")
	chainID := int64(1)
	nonce := uint64(7)

	txA := signedLegacy(t, keyA, chainID, nonce, to)
	txB := signedLegacy(t, keyB, chainID, nonce, to)

	block := types.NewBlock(
		&types.Header{Number: big.NewInt(3)},
		[]*types.Transaction{txA, txB},
		nil,
		nil,
		trie.NewStackTrie(nil),
	)

	expects := []SlotEthExpect{
		{NeedEth: true, From: fromB.Hex(), Nonce: nonce},
	}
	out, err := AlignEthTxsByNonce(block, expects)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0] == nil || out[0].Hash() != txB.Hash() {
		t.Fatalf("want txB (fromB), got %v", out[0])
	}
	// expect.From 与第一笔一致则取第一笔
	expects2 := []SlotEthExpect{
		{NeedEth: true, From: fromA.Hex(), Nonce: nonce},
	}
	out2, err := AlignEthTxsByNonce(block, expects2)
	if err != nil {
		t.Fatal(err)
	}
	if len(out2) != 1 || out2[0] == nil || out2[0].Hash() != txA.Hash() {
		t.Fatalf("want txA (fromA), got %v", out2[0])
	}
}
