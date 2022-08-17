package sync

import (
	"errors"
	"fmt"
	"testing"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db/block"
	"github.com/33cn/externaldb/store"
)

var testSeqsArray = []*types.BlockSeqs{}

func buildTestData() []*types.BlockSeqs {
	// 创世, 1空, 2回滚1, 测试流程
	//receipt0 := []*types.ReceiptData{}
	//txs0 := []*types.Transaction{}
	//block0 := makeBlock(0, txs0, receipt0)
	//seq0 := makeSeq(0, 0, 1, block0)
	//testSeqsArray = append(testSeqsArray, seq0)

	receipt1 := []*types.ReceiptData{}
	txs1 := []*types.Transaction{}
	block1 := makeBlock(1, txs1, receipt1)
	seq1 := makeSeq(1, 1, 1, block1)
	testSeqsArray = append(testSeqsArray, seq1)

	seq2 := makeSeq(2, 1, 2, block1)
	testSeqsArray = append(testSeqsArray, seq2)

	// 下面测试合约 proof_config 测试
	// 构造数据 db/proof_config/config_test.go
	txs3string := []string{
		"0x0a17757365722e702e7465737470726f6f662e636f6e6669671287017b226f7022203a2022616464222c20226f7267616e697a6174696f6e223a202273797374656d222c2022726f6c65223a20226d616e61676572222c20226164647265737322203a2022314b534264313748375a4b3869543337614a7a744642323258477773505464774534222c202022616464726573735f6e6f7465223a202273312d6131227d1a6e08011221023c712ab96682fdc1d7a9101f860bf07750b719221f01a56e0e0d61547650f2551a4730450221009b48d5a5b5f72a10ab0a24ca3a13728038d9b435c3946739f39a524a25fb9711022029ad606d1656733f02b505d3eae3e881167cfef54748a55e7dc5c418ddc2765c2080c2d72f30ad9bf4f390c7c5dc373a22314e55646b684b4d4e33636378685564773964755458576e58686667784645796b47",
		"0x0a17757365722e702e7465737470726f6f662e636f6e666967129f017b226f7022203a2022616464222c20226f7267616e697a6174696f6e223a20226f726731222c2022726f6c65223a20226d616e61676572222c20226164647265737322203a2022314a524e6a64457170344c4a356671796355426d396179434b536565736b674d4b52222c20226f7267616e697a6174696f6e5f6e6f7465223a20226f31222c2022616464726573735f6e6f7465223a20226f312d6131227d1a6e08011221023c712ab96682fdc1d7a9101f860bf07750b719221f01a56e0e0d61547650f2551a473045022100b065a44bed226cae63f0aabc98b68fa620abda9206e11f88d513a487f2c8cbbc02206425c899f990fdd0cb81d26cc1de66e00311ba24b216cb9210d7680fccd1df2a2080c2d72f30e8c4b8a0908db5a91c3a22314e55646b684b4d4e33636378685564773964755458576e58686667784645796b47",
		"0x0a17757365722e702e7465737470726f6f662e636f6e666967129e017b226f7022203a2022616464222c20226f7267616e697a6174696f6e223a20226f726731222c2022726f6c65223a20226d656d626572222c20226164647265737322203a2022314e4c4850456362545757787855336447555a426861796a7243484433707358376b222c20226f7267616e697a6174696f6e5f6e6f7465223a20226f31222c2022616464726573735f6e6f7465223a20226f312d6131227d1a6e08011221023c712ab96682fdc1d7a9101f860bf07750b719221f01a56e0e0d61547650f2551a473045022100883fcea1c638aba3098637a6f720a26f169190ffb0393dc9a0b63621ef59b7ee02204f887dc06424dd4065b7dff4a5b86b05d0fa3a54cf9f648a1e048a4d327b52752080c2d72f3090dcd6a192f896c66d3a22314e55646b684b4d4e33636378685564773964755458576e58686667784645796b47",
	}
	receipt3 := []*types.ReceiptData{{}} // &types.ReceiptData{},
	// &types.ReceiptData{},

	blockConfig := makeBlock(1, decodeTxs(txs3string), receipt3)
	seq3 := makeSeq(3, 1, 1, blockConfig)
	testSeqsArray = append(testSeqsArray, seq3)

	txs4string := []string{
		"0x0a17757365722e702e7465737470726f6f662e636f6e666967125d7b226f7022203a202264656c657465222c20226f7267616e697a6174696f6e223a202273797374656d222c20226164647265737322203a2022314b534264313748375a4b3869543337614a7a744642323258477773505464774534227d1a6d08011221023c712ab96682fdc1d7a9101f860bf07750b719221f01a56e0e0d61547650f2551a46304402206594b669433fc757976e7b998704fa19e1fc63be4274d2b9daa5d4bee847fec8022073681460c050c9ab9049c747ab6719e2b7136e541dc49daa674670234c5144c32080c2d72f30dffab3f8b299a2bd103a22314e55646b684b4d4e33636378685564773964755458576e58686667784645796b47",
		"0x0a17757365722e702e7465737470726f6f662e636f6e66696712a1017b226f7022203a202264656c657465222c20226f7267616e697a6174696f6e223a20226f726731222c2022726f6c65223a20226d656d626572222c20226164647265737322203a2022314e4c4850456362545757787855336447555a426861796a7243484433707358376b222c20226f7267616e697a6174696f6e5f6e6f7465223a20226f31222c2022616464726573735f6e6f7465223a20226f312d6131227d1a6e080112210387bdf56091888507bfaf12034fbdb52ff412471edae0d11cff8a066695ec5e7f1a473045022100b4f7a7c1c51f4f85515f1e0c2fae85887d4f0e791e63130fc72553709b734ce1022004b7ce69c6f260e131ab84478637098b91df5bcdd14df0de011596064dd39b9b2080c2d72f3094afaa86aab7f8cb223a22314e55646b684b4d4e33636378685564773964755458576e58686667784645796b47",
		"0x0a17757365722e702e7465737470726f6f662e636f6e666967125b7b226f7022203a202264656c657465222c20226f7267616e697a6174696f6e223a20226f726731222c20226164647265737322203a2022314a524e6a64457170344c4a356671796355426d396179434b536565736b674d4b52227d1a6d08011221023c712ab96682fdc1d7a9101f860bf07750b719221f01a56e0e0d61547650f2551a4630440220453489c6093e7c2c504fb5d0852f73dd779131fc23c459dced2b1b65a01f5bdb022041f49bcc6399f7ca80af385d0a5f1f25f71c375901fafd9f2ac4bab4756f8fde2080c2d72f308bfda7be93aab5a3433a22314e55646b684b4d4e33636378685564773964755458576e58686667784645796b47",
	}
	receipt4 := []*types.ReceiptData{{}} // &types.ReceiptData{},
	// &types.ReceiptData{},

	blockConfig4 := makeBlock(1, decodeTxs(txs4string), receipt4)
	seq4 := makeSeq(4, 2, 1, blockConfig4)
	testSeqsArray = append(testSeqsArray, seq4)

	txs5string := []string{
		"0x0a17757365722e702e7465737470726f6f662e636f6e666967129e017b226f7022203a2022616464222c20226f7267616e697a6174696f6e223a20226f726731222c2022726f6c65223a20226d656d626572222c20226164647265737322203a2022314e4c4850456362545757787855336447555a426861796a7243484433707358376b222c20226f7267616e697a6174696f6e5f6e6f7465223a20226f31222c2022616464726573735f6e6f7465223a20226f312d6131227d1a6d08011221023c712ab96682fdc1d7a9101f860bf07750b719221f01a56e0e0d61547650f2551a46304402207d41150b02f188738af93acec9896e5a175851a60633d7b853cf010816e0883b02204dd809f662d5439d26e1fbbb95383818d38f0496e63b6fada32f10dbaf8ee7ac2080c2d72f30ddb480fed988e19d7d3a22314e55646b684b4d4e33636378685564773964755458576e58686667784645796b47",
	}
	receipt5 := []*types.ReceiptData{{}}
	blockConfig5 := makeBlock(1, decodeTxs(txs5string), receipt5)
	seq5 := makeSeq(5, 3, 1, blockConfig5)
	testSeqsArray = append(testSeqsArray, seq5)
	return testSeqsArray
	// 下面测试合约
	// TODO
	// 增加测试数据后, 可能同时需要修改测试检测的脚本来检测数据是否和预期一致
}

func decodeTxs(ss []string) (txs []*types.Transaction) {
	for _, s := range ss {
		bs, _ := common.FromHex(s)
		var tx types.Transaction
		_ = types.Decode(bs, &tx)
		txs = append(txs, &tx)
	}
	return
}

// 用需要测试的合约交易构造测试的block
func makeBlock(height int64, t []*types.Transaction, r []*types.ReceiptData) *types.BlockDetail {
	detail := &types.BlockDetail{
		Block: &types.Block{
			Version:    1,
			ParentHash: dummyHash("blockhash", height-1),
			TxHash:     dummyHash("txhash", height),
			StateHash:  dummyHash("statehash", height),
			Height:     height,
			BlockTime:  height*5 + 1,
			Difficulty: 10000,
			MainHash:   dummyHash("mainhash", height),
			MainHeight: height,
			Signature:  nil, //*Signature
			Txs:        t,
		},
		Receipts:       r,
		PrevStatusHash: dummyHash("statehash", height-1),
	}
	return detail

}
func dummyHash(key string, height int64) []byte {
	if height < 0 {
		return []byte("")
	}
	return []byte(fmt.Sprintf("%s-%04d", key, height))
}

// 构造测试数据 sewq
// op: 1 add, 2 delete
func makeSeq(seq int64, height int64, op int64, block *types.BlockDetail) *types.BlockSeqs {
	seqs := []*types.BlockSeq{
		{
			Num: seq,
			Seq: &types.BlockSequence{
				Type: op,
				Hash: block.Block.MainHash,
			},
			Detail: block,
		},
	}
	return &types.BlockSeqs{Seqs: seqs}
}

func TestParseSeqs(t *testing.T) {
	rollbackSeq = []int64{3, 4}
	data := buildTestData()
	for i, i2 := range data {
		count, start, seqsMap, err := parseSeqs(i2)
		t.Log(i, count, start, seqsMap)
		if err != nil {
			//t.Log(i, count, start, seqsMap)
			t.Error(err)
		}
	}
}

func dealBlocks(seqs *types.BlockSeqs, d *store.SeqNum) error {
	beg := types.Now()
	defer func() {
		log.Info("dealBlocks", "cost", types.Since(beg))
	}()
	count, start, seqsMap, err := parseSeqs(seqs)
	if err != nil {
		log.Error("deal request", "parseSeqs", err)
		return err
	}
	log.Info("dealBlocks", "p1", "parseSeqs", "cost", types.Since(beg))

	// 在app 端保存成功， 但回复ok时，程序挂掉, 记录日志
	if start <= d.Number {
		log.Error("deal request", "seq", start, "current_seq", d.Number)
	}
	if start+int64(count-1) <= d.Number {
		return nil
	}

	// 一般在配置有误时发生， 在同一个节点上配置相同的推送名
	if start > d.Number+1 {
		log.Error("deal request", "seq", start, "current_seq", d.Number)
		return errors.New("bad seq")
	}

	number := d.Number
	//records := make([]db.Record, 0)
	for i := 0; i < count; i++ {
		if start+int64(i) <= d.Number {
			continue
		}

		number++
		if seq, ok := seqsMap[int64(i)+start]; ok {
			blockSeq := &block.Seq{
				SyncSeq:     int(number),
				From:        d.From,
				Number:      int(seq.Num),
				Hash:        common.ToHex(seq.Seq.Hash),
				Type:        int(seq.Seq.Type),
				BlockDetail: types.Encode(seq.Detail),
			}
			log.Info("blockSeq", "blockSeq", blockSeq)
			//seqRecord := block.NewSeqRecord(blockSeq)
			//records = append(records, seqRecord)
		} else {
			continue
		}
	}
	//lastSeq := block.NewLastRecord(number)
	log.Info("dealBlocks", "p2", "genRecord", "cost", types.Since(beg))
	//log.Info("newLastRecord", "seq", lastSeq.ID())

	//// save
	//err = seqStore.SaveSeqs(records)
	//if err == nil {
	//	err = seqNumStore.UpdateLastSeq(lastSeq)
	//}

	if err == nil {
		d.Number = number
	}
	log.Info("dealBlocks", "p3", "callSaveDB", "cost", types.Since(beg))

	return err
}

func TestDealBlocks(t *testing.T) {
	d := &store.SeqNum{
		Number: 0,
	}
	rollbackSeq = []int64{3, 4}
	data := buildTestData()
	for _, i2 := range data {
		err := dealBlocks(i2, d)
		if err != nil {
			t.Error(nil)
		}
	}
}
