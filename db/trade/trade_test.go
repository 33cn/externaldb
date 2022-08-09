package trade

//func TestTx(t *testing.T) {
//	tx := &Tx{
//		Block: &db.Block{
//			Height:      1,
//			Ts:          2222,
//			BlockHash:   "common.ToHex",
//			Index:       2,
//			Send:        "tx.From",
//			TxHash:      "Hash",
//			HeightIndex: db.HeightIndex(1, 2),
//		},
//		TxType:  TxSellLimitX,
//		Success: true,
//	}
//	v, _ := json.Marshal(tx)
//	t.Log("tx", "v", string(v))
//	t.Error("test")
//}
//
//func TestOrder(t *testing.T) {
//	order := &Order{
//		Block: &db.Block{
//			Height:      1,
//			Ts:          2222,
//			BlockHash:   "common.ToHex",
//			Index:       2,
//			Send:        "tx.From",
//			TxHash:      "Hash",
//			HeightIndex: db.HeightIndex(1, 2222),
//		},
//		AssetSymbol:       "Test",
//		AssetExec:         "token",
//		Owner:             "linjing-address",
//		AmountPerBoardlot: 1,
//		MinBoardlot:       30,
//		PricePerBoardlot:  2,
//		TotalBoardlot:     100,
//
//		TradedBoardlot: 30,
//		IsSellOrder:    true,
//		IsFinished:     false,
//		Status:         StatusCreated,
//
//		BuyID:  "buy-id",
//		SellID: "SellID",
//	}
//	v, _ := json.Marshal(order)
//	t.Log("order", "v", string(v))
//	t.Error("test")
//}
//
//func TestAddr(t *testing.T) {
//	t.Error(address.ExecAddress("user.p.sakurachain." + "trade"))
//}
