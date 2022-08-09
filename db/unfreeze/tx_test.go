package unfreeze

//import (
//	"encoding/json"
//	"testing"
//)
//
//func TestFmt(t *testing.T) {
//	b := &Block{
//		Hash:   "hash1",
//		Height: 6,
//		Ts:     15000000,
//		Index:  2,
//	}
//	c := &ActionCreate{
//		StartTime:       15000000,
//		AssetExec:       "coins",
//		AssetSymbol:     "bty",
//		TotalCount:      6 * 1e8,
//		Means:           "fix_amount",
//		FixAmountOption: &FixAmount{Period: 3600, Amount: 1e7},
//	}
//	w := &ActionWithdraw{
//		Amount: 2 * 1e7,
//	}
//	t1 := &ActionTerminate{
//		AmountBack: 57 * 1e7,
//		AmountLeft: 1 * 1e7,
//	}
//	tx := UnfreezeTx{
//		BlockInfo:   b,
//		Creator:     "me",
//		Beneficiary: "you",
//		UnfreezeID:  "the-id",
//		Success:     true,
//
//		ActionType: ActionTypeCreate,
//		Create:     c,
//	}
//	v, _ := json.Marshal(&tx)
//	t.Log(string(v))
//
//	tx.ActionType = ActionTypeWithdraw
//	tx.Create = nil
//	tx.Withdraw = w
//	v, _ = json.Marshal(&tx)
//	t.Log(string(v))
//
//	tx.ActionType = ActionTypeTerminate
//	tx.Withdraw = nil
//	tx.Terminate = t1
//	v, _ = json.Marshal(&tx)
//	t.Log(string(v))
//
//}

/*
 fmt json:

{
   "creator" : "me",
   "create" : {
      "means" : "fix_amount",
      "fix_amount" : {
         "period" : 3600,
         "amount" : 10000000
      },
      "total_count" : 600000000,
      "start_time" : 15000000,
      "asset_symbol" : "bty",
      "asset_exec" : "coins"
   },
   "unfreeze_id" : "the-id",
   "block" : {
      "hash" : "hash1",
      "height" : 6,
      "ts" : 15000000,
      "index" : 2
   },
   "action_type" : 1,
   "beneficiary" : "you",
   "success" : true
}

{
   "block" : {
      "index" : 2,
      "height" : 6,
      "ts" : 15000000,
      "hash" : "hash1"
   },
   "beneficiary" : "you",
   "unfreeze_id" : "the-id",
   "creator" : "me",
   "action_type" : 2,
   "success" : true,
   "withdraw" : {
      "amount" : 20000000
   }
}

{
   "beneficiary" : "you",
   "terminate" : {
      "amount_left" : 10000000,
      "amount_back" : 570000000
   },
   "success" : true,
   "block" : {
      "hash" : "hash1",
      "index" : 2,
      "height" : 6,
      "ts" : 15000000
   },
   "unfreeze_id" : "the-id",
   "action_type" : 3,
   "creator" : "me"
}

*/
