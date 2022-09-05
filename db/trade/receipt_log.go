package trade

// 只能通过日志顺序判断 帐号receipt对应的资产类型
// 平行链或交易组少 log fee
// sell limit/revoke： log fee，exec frozen/active asset, trade log
// buy limit/revoke: log fee, exec frozen/active coins, trade log
// buy markert: log fee, coins exec tranfer * 2, asset exec transfer * 2, trade log * 2
// sell markert: log fee, coins exec tranfer * 2, asset exec transfer * 2, trade log * 2

import (
	"encoding/hex"
	"strconv"

	"github.com/33cn/chain33/types"
	"github.com/pkg/errors"

	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/account"
	pty "github.com/33cn/plugin/plugin/dapp/trade/types"
)

func (t *tradeConvert) convertSellLimitLog(op int) ([]db.Record, error) {
	// sell limit/revoke： log fee， frozen/active asset, trade log
	//logs := []int{types.TyLogFee, types.TyLogExecFrozen, pty.TyLogTradeSellLimit}
	var records []db.Record
	var err error
	cnt := len(t.receipt.Logs)
	log.Debug("convertSellLimitLog", "log_cnt", cnt)
	if cnt < 2 {
		return nil, errors.Wrapf(errors.New("LossLogs"), "sell_limit e:2~3.a:%d", cnt)
	}

	sell, err := t.TyLogTradeSellLimit(t.receipt.Logs[cnt-1].Log, op)
	if err != nil {
		return nil, errors.WithMessage(err, "sell_limit log[-1]")
	}
	records = append(records, sell)

	r1, err := account.RecordHelper(t.receipt.Logs[cnt-2], op, t.accountIDAsset)
	if err != nil {
		return nil, errors.WithMessage(err, "sell_limit: log[-2]")
	}
	records = append(records, r1...)

	if cnt == 3 {
		fee, err := account.RecordHelper(t.receipt.Logs[0], op, t.accountIDBty)
		if err != nil {
			return nil, errors.WithMessage(err, "sell_limit: log-fee")
		}
		records = append(records, fee...)
	}
	return records, nil
}

func (t *tradeConvert) convertSellRevokeLog(op int) ([]db.Record, error) {
	// sell limit/revoke： log fee， frozen/active asset, trade log
	//logs := []int{types.TyLogFee, types.TyLogExecActive, pty.TyLogTradeSellRevoke}
	var records []db.Record
	var err error
	cnt := len(t.receipt.Logs)
	log.Debug("convertSellRevokeLog", "log_cnt", cnt)
	if cnt < 2 {
		return nil, errors.Wrapf(errors.New("LossLogs"), "sell_revoke e:2~3.a:%d", cnt)
	}

	sell, err := t.TyLogTradeSellRevoke(t.receipt.Logs[cnt-1].Log, op)
	if err != nil {
		return nil, errors.WithMessage(err, "sell_revoke log[-1]")
	}
	records = append(records, sell)

	r1, err := account.RecordHelper(t.receipt.Logs[cnt-2], op, t.accountIDAsset)
	if err != nil {
		return nil, errors.WithMessage(err, "sell_revoke: log[-2]")
	}
	records = append(records, r1...)

	if cnt == 3 {
		fee, err := account.RecordHelper(t.receipt.Logs[0], op, t.accountIDBty)
		if err != nil {
			return nil, errors.WithMessage(err, "sell_revoke: log-fee")
		}
		records = append(records, fee...)
	}
	return records, nil
}

func (t *tradeConvert) convertBuyRevokeLog(op int) ([]db.Record, error) {
	// buy limit/revoke: log fee, frozen/active price, trade log
	//logs := []int{types.TyLogFee, types.TyLogExecActive, pty.TyLogTradeBuyRevoke}
	var records []db.Record
	var err error
	cnt := len(t.receipt.Logs)
	log.Debug("convertBuyRevokeLog", "log_cnt", cnt)
	if cnt < 2 {
		return nil, errors.Wrapf(errors.New("LossLogs"), "buy_revoke e:2~3.a:%d", cnt)
	}
	buy, err := t.TyLogTradeBuyRevoke(t.receipt.Logs[cnt-1].Log, op)
	if err != nil {
		return nil, errors.WithMessage(err, "buy_revoke log[-1]")
	}
	records = append(records, buy)

	r1, err := account.RecordHelper(t.receipt.Logs[cnt-2], op, t.accountIDPrice)
	if err != nil {
		return nil, errors.WithMessage(err, "buy_revoke: log[-2]")
	}
	records = append(records, r1...)

	if cnt == 3 {
		fee, err := account.RecordHelper(t.receipt.Logs[0], op, t.accountIDBty)
		if err != nil {
			return nil, errors.WithMessage(err, "buy_revoke: log-fee")
		}
		records = append(records, fee...)
	}
	return records, nil
}

func (t *tradeConvert) convertBuyLimitLog(op int) ([]db.Record, error) {
	// buy limit/revoke: log fee, frozen/active price, trade log
	//logs := []int{types.TyLogFee, types.TyLogExecFrozen, pty.TyLogTradeBuyLimit}
	var records []db.Record
	var err error
	cnt := len(t.receipt.Logs)
	log.Debug("convertBuyLimitLog", "log_cnt", cnt)
	if cnt < 2 {
		return nil, errors.Wrapf(errors.New("LossLogs"), "buy_limit e:2~3.a:%d", cnt)
	}

	buy, err := t.TyLogTradeBuyLimit(t.receipt.Logs[cnt-1].Log, op)
	if err != nil {
		return nil, errors.WithMessage(err, "buy_limit log[-1]")
	}
	records = append(records, buy)

	r1, err := account.RecordHelper(t.receipt.Logs[cnt-2], op, t.accountIDPrice)
	if err != nil {
		return nil, errors.WithMessage(err, "buy_limit: log[-2]")
	}
	records = append(records, r1...)

	if cnt == 3 {
		fee, err := account.RecordHelper(t.receipt.Logs[0], op, t.accountIDBty)
		if err != nil {
			return nil, errors.WithMessage(err, "buy_limit: log-fee")
		}
		records = append(records, fee...)
	}
	return records, nil
}

func (t *tradeConvert) convertBuyMarketLog(op int) ([]db.Record, error) {
	// buy markert: log fee, price tranfer * 2, asset transfer * 2, trade log * 2
	//logs := []int{types.TyLogFee,
	// types.TyLogExecTransfer, types.TyLogExecTransfer, // price
	// types.TyLogExecTransfer, types.TyLogExecTransfer, // asset
	// pty.TyLogTradeSellLimit, pty.TyLogTradeBuyMarket}
	var records []db.Record
	var err error
	cnt := len(t.receipt.Logs)
	log.Debug("convertBuyMarketLog", "log_cnt", cnt)
	if cnt < 6 {
		return nil, errors.Wrapf(errors.New("LossLogs"), "buy_market e:6~7.a:%d", cnt)
	}

	buy, err := t.TyLogTradeBuyMarket(t.receipt.Logs[cnt-1].Log, op)
	if err != nil {
		return nil, errors.WithMessage(err, "buy_market log[-1]")
	}
	records = append(records, buy)

	sell, err := t.TyLogTradeSellLimit(t.receipt.Logs[cnt-2].Log, op)
	if err != nil {
		return nil, errors.WithMessage(err, "buy_market log[-2]")
	}
	records = append(records, sell)

	r3, err := account.RecordHelper(t.receipt.Logs[cnt-3], op, t.accountIDAsset)
	if err != nil {
		return nil, errors.WithMessage(err, "buy_market: log[-3]")
	}
	records = append(records, r3...)

	r4, err := account.RecordHelper(t.receipt.Logs[cnt-4], op, t.accountIDAsset)
	if err != nil {
		return nil, errors.WithMessage(err, "buy_market: log[-4]")
	}
	records = append(records, r4...)

	r5, err := account.RecordHelper(t.receipt.Logs[cnt-5], op, t.accountIDPrice)
	if err != nil {
		return nil, errors.WithMessage(err, "buy_market: log[-5]")
	}
	records = append(records, r5...)

	r6, err := account.RecordHelper(t.receipt.Logs[cnt-6], op, t.accountIDPrice)
	if err != nil {
		return nil, errors.WithMessage(err, "buy_market: log[-6]")
	}
	records = append(records, r6...)

	if cnt == 7 {
		fee, err := account.RecordHelper(t.receipt.Logs[0], op, t.accountIDBty)
		if err != nil {
			return nil, errors.WithMessage(err, "buy_market: log-fee")
		}
		records = append(records, fee...)
	}

	return records, nil
}

func (t *tradeConvert) convertSellMarketLog(op int) ([]db.Record, error) {
	// sell markert: log fee, price exec tranfer * 2, asset exec transfer * 2, trade log * 2
	//logs := []int{types.TyLogFee,
	// types.TyLogExecTransfer, types.TyLogExecTransfer, // price
	// types.TyLogExecTransfer, types.TyLogExecTransfer, // asset
	// pty.TyLogTradeBuyLimit, pty.TyLogTradeSellMarket}
	var records []db.Record
	var err error
	cnt := len(t.receipt.Logs)
	log.Debug("convertSellMarketLog", "log_cnt", cnt)
	if cnt < 6 {
		return nil, errors.Wrapf(errors.New("LossLogs"), "sell_market e:6~7.a:%d", cnt)
	}

	buy, err := t.TyLogTradeSellMarket(t.receipt.Logs[cnt-1].Log, op)
	if err != nil {
		return nil, errors.WithMessage(err, "sell_market log[-1]")
	}
	records = append(records, buy)

	sell, err := t.TyLogTradeBuyLimit(t.receipt.Logs[cnt-2].Log, op)
	if err != nil {
		return nil, errors.WithMessage(err, "sell_market log[-2]")
	}
	records = append(records, sell)

	r3, err := account.RecordHelper(t.receipt.Logs[cnt-3], op, t.accountIDAsset)
	if err != nil {
		return nil, errors.WithMessage(err, "sell_market: log[-3]")
	}
	records = append(records, r3...)

	r4, err := account.RecordHelper(t.receipt.Logs[cnt-4], op, t.accountIDAsset)
	if err != nil {
		return nil, errors.WithMessage(err, "sell_market: log[-4]")
	}
	records = append(records, r4...)

	r5, err := account.RecordHelper(t.receipt.Logs[cnt-5], op, t.accountIDPrice)
	if err != nil {
		return nil, errors.WithMessage(err, "sell_market: log[-5]")
	}
	records = append(records, r5...)

	r6, err := account.RecordHelper(t.receipt.Logs[cnt-6], op, t.accountIDPrice)
	if err != nil {
		return nil, errors.WithMessage(err, "sell_market: log[-6]")
	}
	records = append(records, r6...)

	if cnt == 7 {
		fee, err := account.RecordHelper(t.receipt.Logs[0], op, t.accountIDBty)
		if err != nil {
			return nil, errors.WithMessage(err, "sell_market: log-fee")
		}
		records = append(records, fee...)
	}
	return records, nil
}

// convert trade logs and setup asset

func (t *tradeConvert) TyLogTradeSellLimit(v []byte, op int) (db.Record, error) {
	var l pty.ReceiptTradeSellLimit
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "trade Decode TyLogTradeSellLimit log")
	}
	t.setupAsset(l.Base.AssetExec, l.Base.TokenSymbol)
	t.setupPrice(l.Base.PriceExec, l.Base.PriceSymbol)

	if l.Base.TotalBoardlot == 0 && op == db.SeqTypeDel {
		t2 := &dbOrder{IKey: newOrderKey(l.Base.SellID), Op: db.NewOp(op)}
		return t2, nil
	}

	t2 := &dbOrder{IKey: newOrderKey(l.Base.SellID), Op: db.NewOp(db.OpAdd)}

	order, err := t.createSellOrder(l.Base)
	if err != nil {
		return nil, err
	}
	order.IsFinished = l.Base.TotalBoardlot == l.Base.SoldBoardlot
	order.Status = StatusCreated

	if order.IsFinished {
		order.Status = StatusDone
	}
	if op == db.SeqTypeDel {
		order.Status = StatusCreated
		order.IsFinished = false
		order.TradedBoardlot -= t.boardlotCount
	}

	t2.order = order

	return t2, nil
}

func (t *tradeConvert) TyLogTradeSellMarket(v []byte, op int) (db.Record, error) {
	var l pty.ReceiptSellMarket
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "trade Decode TyLogTradeSellLimit log")
	}
	t.setupAsset(l.Base.AssetExec, l.Base.TokenSymbol)
	t.setupPrice(l.Base.PriceExec, l.Base.PriceSymbol)

	sellID := l.Base.SellID
	if sellID == "" {
		sellID = "mavl-trade-sell-" + hex.EncodeToString(t.tx.Hash())
	}

	if op == db.SeqTypeDel {
		t2 := &dbOrder{IKey: newOrderKey(sellID), Op: db.NewOp(op)}
		return t2, nil
	}

	t2 := &dbOrder{IKey: newOrderKey(sellID), Op: db.NewOp(db.OpAdd)}

	order, err := t.createSellOrder(l.Base)
	if err != nil {
		return nil, err
	}
	order.IsFinished = true
	order.Status = StatusDone

	t2.order = order

	return t2, nil
}

func (t *tradeConvert) TyLogTradeSellRevoke(v []byte, op int) (db.Record, error) {
	var l pty.ReceiptTradeSellRevoke
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "trade Decode TyLogTradeSellRevoke log")
	}
	t.setupAsset(l.Base.AssetExec, l.Base.TokenSymbol)
	t.setupPrice(l.Base.PriceExec, l.Base.PriceSymbol)

	t2 := &dbOrder{IKey: newOrderKey(l.Base.SellID), Op: db.NewOp(op)}
	order, err := t.createSellOrder(l.Base)
	if err != nil {
		return nil, err
	}
	order.IsFinished = true
	order.Status = StatusRevoked

	if op == db.SeqTypeDel {
		order.Status = StatusCreated
		order.IsFinished = false
	}
	t2.order = order

	return t2, nil
}

// buy log
func (t *tradeConvert) TyLogTradeBuyLimit(v []byte, op int) (db.Record, error) {
	var l pty.ReceiptTradeBuyLimit
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "trade Decode TyLogTradeBuyLimit log")
	}
	t.setupAsset(l.Base.AssetExec, l.Base.TokenSymbol)
	t.setupPrice(l.Base.PriceExec, l.Base.PriceSymbol)

	if l.Base.TotalBoardlot == 0 && op == db.SeqTypeDel {
		t2 := &dbOrder{IKey: newOrderKey(l.Base.BuyID), Op: db.NewOp(op)}
		return t2, nil
	}
	t2 := &dbOrder{IKey: newOrderKey(l.Base.BuyID), Op: db.NewOp(db.OpAdd)}

	order, err := t.createBuyOrder(l.Base)
	if err != nil {
		return nil, err
	}

	order.IsFinished = l.Base.TotalBoardlot == l.Base.BoughtBoardlot
	order.Status = StatusCreated

	if order.IsFinished {
		order.Status = StatusDone
	}
	if op == db.SeqTypeDel {
		order.IsFinished = false
		order.Status = StatusCreated
		order.TradedBoardlot -= t.boardlotCount
	}
	t2.order = order

	return t2, nil
}

func (t *tradeConvert) TyLogTradeBuyMarket(v []byte, op int) (db.Record, error) {
	var l pty.ReceiptTradeBuyMarket
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "trade Decode TyLogTradeBuyMarket log")
	}
	t.setupAsset(l.Base.AssetExec, l.Base.TokenSymbol)
	t.setupPrice(l.Base.PriceExec, l.Base.PriceSymbol)

	buyID := l.Base.BuyID
	if l.Base.BuyID == "" {
		buyID = "mavl-trade-buy-" + hex.EncodeToString(t.tx.Hash())
	}

	log.Debug("buy market", "log", &l, "v", l.Base, "op", op, "l.Base.BuyID", buyID)
	if op == db.SeqTypeDel {
		t2 := &dbOrder{IKey: newOrderKey(buyID), Op: db.NewOp(op)}
		return t2, nil
	}
	log.Debug("buy market", "log", &l, "v", l.Base)
	t2 := &dbOrder{IKey: newOrderKey(buyID), Op: db.NewOp(db.OpAdd)}

	order, err := t.createBuyOrder(l.Base)
	if err != nil {
		return nil, err
	}
	order.Status = StatusDone
	order.IsFinished = true

	t2.order = order
	log.Debug("buy market", "log", &l, "t2", *t2.order)
	return t2, nil
}

func (t *tradeConvert) TyLogTradeBuyRevoke(v []byte, op int) (db.Record, error) {
	var l pty.ReceiptTradeBuyRevoke
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "trade Decode TyLogTradeBuyRevoke log")
	}
	t.setupAsset(l.Base.AssetExec, l.Base.TokenSymbol)
	t.setupPrice(l.Base.PriceExec, l.Base.PriceSymbol)

	t2 := &dbOrder{IKey: newOrderKey(l.Base.BuyID), Op: db.NewOp(db.OpAdd)}

	order, err := t.createBuyOrder(l.Base)
	if err != nil {
		return nil, err
	}
	order.Status = StatusRevoked
	order.IsFinished = true
	if op == db.SeqTypeDel {
		order.Status = StatusCreated
		order.IsFinished = false
	}
	t2.order = order

	return t2, nil
}

func (t *tradeConvert) createBuyOrder(base *pty.ReceiptBuyBase) (*Order, error) {
	price, amount, err := priceAmount(base.PricePerBoardlot, base.AmountPerBoardlot)
	if err != nil {
		return nil, err
	}
	order := &Order{
		Block:             t.block,
		AssetSymbol:       base.TokenSymbol,
		AssetExec:         base.AssetExec,
		Owner:             base.Owner,
		AmountPerBoardlot: amount,
		MinBoardlot:       base.MinBoardlot,
		PricePerBoardlot:  price,
		TotalBoardlot:     base.TotalBoardlot,

		TradedBoardlot: base.BoughtBoardlot,
		IsSellOrder:    false,

		IsFinished: true,
		Status:     StatusRevoked,

		BuyID:       base.BuyID,
		SellID:      base.SellID,
		PriceSymbol: base.PriceSymbol,
		PriceExec:   base.PriceExec,
	}
	if order.AssetExec == "" {
		order.AssetExec = "token"
	}
	if order.PriceExec == "" {
		order.PriceExec = "coins"
		order.PriceSymbol = t.symbol
	}
	return order, nil
}

func (t *tradeConvert) createSellOrder(base *pty.ReceiptSellBase) (*Order, error) {
	price, amount, err := priceAmount(base.PricePerBoardlot, base.AmountPerBoardlot)
	if err != nil {
		return nil, err
	}
	order := &Order{
		Block:             t.block,
		AssetSymbol:       base.TokenSymbol,
		AssetExec:         base.AssetExec,
		Owner:             base.Owner,
		AmountPerBoardlot: amount,
		MinBoardlot:       base.MinBoardlot,
		PricePerBoardlot:  price,
		TotalBoardlot:     base.TotalBoardlot,

		TradedBoardlot: base.SoldBoardlot,
		IsSellOrder:    true,
		IsFinished:     true,
		Status:         StatusRevoked,

		BuyID:  base.BuyID,
		SellID: base.SellID,

		PriceSymbol: base.PriceSymbol,
		PriceExec:   base.PriceExec,
	}
	if order.AssetExec == "" {
		order.AssetExec = "token"
	}
	if order.PriceExec == "" {
		order.PriceExec = "coins"
		order.PriceSymbol = t.symbol
	}
	return order, nil
}

func priceAmount(price, amount string) (int64, int64, error) {
	p1, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return 0, 0, errors.Wrap(err, "price: get float from string failed")
	}
	p2 := int64(p1 * 1e8)
	a1, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return 0, 0, errors.Wrap(err, "amount: get float from string failed")
	}
	a2 := int64(a1 * 1e8)
	return p2, a2, nil
}

func (t *tradeConvert) setupAsset(exec, symbol string) {
	if exec == "" {
		exec = "token"
	}
	t.accountIDAsset = account.Account{
		AssetSymbol: symbol,
		AssetExec:   exec,
		HeightIndex: db.HeightIndex(t.block.Height, t.block.Index),
		Height:      t.block.Height,
		BlockTime:   t.env.Block.Block.BlockTime,
	}
}

func (t *tradeConvert) setupPrice(exec, symbol string) {
	if exec == "" {
		exec = "coins"
		symbol = t.symbol
	}
	t.accountIDPrice = account.Account{
		AssetSymbol: symbol,
		AssetExec:   exec,
		HeightIndex: db.HeightIndex(t.block.Height, t.block.Index),
		Height:      t.block.Height,
		BlockTime:   t.env.Block.Block.BlockTime,
	}
}
