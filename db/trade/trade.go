package trade

import (
	"github.com/33cn/chain33/common"
	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/converts"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/account"
	"github.com/33cn/externaldb/db/transaction"
	"github.com/33cn/externaldb/util"
	pty "github.com/33cn/plugin/plugin/dapp/trade/types"
	"github.com/pkg/errors"
)

// status
const (
	StatusCreated = "created"
	StatusRevoked = "revoked"
	StatusDone    = "done"
)

// tx type
const (
	TxSellMarketX = "sell_market"
	TxSellLimitX  = "sell_limit"
	TxSellRevokeX = "sell_revoke"
	TxBuyMarketX  = "buy_market"
	TxBuyLimitX   = "buy_limit"
	TxBuyRevokeX  = "buy_revoke"
)

// db
const (
	TradeDBX      = "trade"
	TradeTxDBX    = "trade_tx"
	TradeAssetDBX = "trade_asset"
	DefaultType   = "_doc"
)

// Asset in trade
type Asset struct {
	AssetSymbol string `json:"asset_symbol"`
	AssetExec   string `json:"asset_exec"`
	PriceSymbol string `json:"price_symbol"`
	PriceExec   string `json:"price_exec"`
}

// Order Order
type Order struct {
	*db.Block

	AssetSymbol       string `json:"asset_symbol"`
	AssetExec         string `json:"asset_exec"`
	Owner             string `json:"owner"`
	AmountPerBoardlot int64  `json:"boardlot_amount"`
	MinBoardlot       int64  `json:"min_boardlot"`
	PricePerBoardlot  int64  `json:"boardlot_price"`
	TotalBoardlot     int64  `json:"total_boardlot"`

	TradedBoardlot int64  `json:"traded_boardlot"`
	IsSellOrder    bool   `json:"is_sell"`
	IsFinished     bool   `json:"is_finished"`
	Status         string `json:"status"`

	BuyID  string `json:"buy_id"`
	SellID string `json:"sell_id"`

	PriceSymbol string `json:"price_symbol"`
	PriceExec   string `json:"price_exec"`
}

// Tx trade tx
type Tx struct {
	*db.Block
	TxType  string `json:"tx_type"`
	Success bool   `json:"success"`

	// option
	SellMarket *pty.TradeForSellMarket `json:"sell_market,omitempty"`
	SellLimit  *pty.TradeForSell       `json:"sell_limit,omitempt"`
	SellRevoke *pty.TradeForRevokeSell `json:"sell_revoke,omitempty"`
	BuyMarket  *pty.TradeForBuy        `json:"buy_market,omitempty"`
	BuyLimit   *pty.TradeForBuyLimit   `json:"buy_limit,omitempty"`
	BuyRevoke  *pty.TradeForRevokeBuy  `json:"buy_revoke,omitempty"`
}

var log = l.New("module", "db.trade")

type tradeConvert struct {
	env    *db.TxEnv
	title  string
	symbol string

	tx      *types.Transaction
	receipt *types.ReceiptData
	block   *db.Block

	accountIDBty   account.Account // for fee bty
	accountIDPrice account.Account // for price
	accountIDAsset account.Account // for asset
	boardlotCount  int64
}

func init() {
	converts.Register("trade", NewConvert)
}

// NewConvert NewConvert
func NewConvert(paraTitle, symbol string, supports []string) db.ExecConvert {
	e := &tradeConvert{symbol: symbol, title: paraTitle}

	return e
}

// InitDB init db
func (t *tradeConvert) InitDB(cli db.DBCreator) error {
	var err error

	err = account.InitDB(cli)
	if err != nil {
		return err
	}

	err = transaction.InitDB(cli)
	if err != nil {
		return err
	}

	err = util.InitIndex(cli, TradeDBX, TradeDBX, OrderMapping)
	if err != nil {
		return err
	}

	err = util.InitIndex(cli, TradeAssetDBX, TradeAssetDBX, AssetMapping)
	if err != nil {
		return err
	}

	return nil
}

func (t *tradeConvert) positionID() string {
	return util.PositionID("trade", t.block.Height, t.block.Index)
}

func (t *tradeConvert) setupEnv(env *db.TxEnv) {
	t.env = env
	tx := env.Block.Block.Txs[env.TxIndex]
	receipt := env.Block.Receipts[env.TxIndex]
	t.tx = tx
	t.receipt = receipt
	t.block = db.SetupBlock(env, util.AddressConvert(tx.From()), common.ToHex(tx.Hash()))

	t.accountIDBty = account.Account{
		AssetSymbol: t.symbol,
		AssetExec:   account.ExecCoinsX,
		HeightIndex: db.HeightIndex(t.block.Height, t.block.Index),
		Height:      t.block.Height,
		BlockTime:   t.env.Block.Block.BlockTime,
	}
}

// ConveretTx impl
func (t *tradeConvert) ConvertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	t.setupEnv(env)

	var action pty.Trade
	err := types.Decode(t.tx.Payload, &action)
	if err != nil {
		return nil, errors.Wrap(err, "decode tx action")
	}

	var records []db.Record
	switch action.Ty {
	case pty.TradeSellLimit:
		log.Debug("Convert", "action", "TradeSellLimit")
		records, err = t.convertSellLimit(action.GetSellLimit(), op)
	case pty.TradeSellMarket:
		log.Debug("Convert", "action", "TradeSellMarket")
		records, err = t.convertSellMarket(action.GetSellMarket(), op)
	case pty.TradeRevokeSell:
		log.Debug("Convert", "action", "TradeRevokeSell")
		records, err = t.convertSellRevoke(action.GetRevokeSell(), op)
	case pty.TradeBuyLimit:
		log.Debug("Convert", "action", "TradeBuyLimit")
		records, err = t.convertBuyLimit(action.GetBuyLimit(), op)
	case pty.TradeBuyMarket:
		log.Debug("Convert", "action", "TradeBuyMarket")
		records, err = t.convertBuyMarket(action.GetBuyMarket(), op)
	case pty.TradeRevokeBuy:
		log.Debug("Convert", "action", "TradeRevokeBuy")
		records, err = t.convertBuyRevoke(action.GetRevokeBuy(), op)
	}

	tx := transaction.ConvertTransaction(env)
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(tx.Hash),
		Op:   db.NewOp(op),
		Tx:   tx,
	}
	records = append(records, &txRecord)

	return records, err

}

func newTradeAsset(rs []db.Record, asset *account.Account, price *account.Account) []db.Record {
	if asset.AssetExec != "" && asset.AssetSymbol != "" {
		r := &dbAsset{
			IKey: newAssetKey(asset.AssetExec + "." + asset.AssetSymbol),
			Op:   db.NewOp(db.OpAdd),
			asset: &Asset{
				AssetExec:   asset.AssetExec,
				AssetSymbol: asset.AssetSymbol,
				PriceExec:   price.AssetExec,
				PriceSymbol: price.AssetSymbol,
			},
		}
		rs = append(rs, r)
	}

	return rs
}

// 1 order, 2 account
func (t *tradeConvert) convertSellLimit(action *pty.TradeForSell, op int) ([]db.Record, error) {
	rs := make([]db.Record, 0)
	txRecord := Tx{
		Block:     t.block,
		TxType:    TxSellLimitX,
		Success:   (t.receipt.Ty == types.ExecOk),
		SellLimit: action,
	}
	r := &dbTxOrder{IKey: newTxKey(t.block.TxHash), Op: db.NewOp(op), tx: &txRecord}
	rs = append(rs, r)

	if t.receipt.Ty != types.ExecOk {
		return rs, nil
	}

	rs2, err := t.convertSellLimitLog(op)
	if err != nil {
		return rs, err
	}
	rs = append(rs, rs2...)

	rs = newTradeAsset(rs, &t.accountIDAsset, &t.accountIDPrice)

	return rs, nil
}

// 1 order, 2 account
func (t *tradeConvert) convertSellMarket(action *pty.TradeForSellMarket, op int) ([]db.Record, error) {
	rs := make([]db.Record, 0)
	txRecord := Tx{
		Block:      t.block,
		TxType:     TxSellMarketX,
		Success:    (t.receipt.Ty == types.ExecOk),
		SellMarket: action,
	}
	r := &dbTxOrder{IKey: newTxKey(t.block.TxHash), Op: db.NewOp(op), tx: &txRecord}
	rs = append(rs, r)

	if t.receipt.Ty != types.ExecOk {
		return rs, nil
	}

	t.boardlotCount = action.BoardlotCnt

	rs2, err := t.convertSellMarketLog(op)
	if err != nil {
		return rs, err
	}
	rs = append(rs, rs2...)

	return rs, nil
}

// 1 order, 2 account
func (t *tradeConvert) convertSellRevoke(action *pty.TradeForRevokeSell, op int) ([]db.Record, error) {
	rs := make([]db.Record, 0)
	txRecord := Tx{
		Block:      t.block,
		TxType:     TxSellRevokeX,
		Success:    (t.receipt.Ty == types.ExecOk),
		SellRevoke: action,
	}
	r := &dbTxOrder{IKey: newTxKey(t.block.TxHash), Op: db.NewOp(op), tx: &txRecord}
	rs = append(rs, r)

	if t.receipt.Ty != types.ExecOk {
		return rs, nil
	}

	rs2, err := t.convertSellRevokeLog(op)
	if err != nil {
		return rs, err
	}

	rs = append(rs, rs2...)
	return rs, nil
}

// buy tx
// 1 order, 2 account
func (t *tradeConvert) convertBuyLimit(action *pty.TradeForBuyLimit, op int) ([]db.Record, error) {
	rs := make([]db.Record, 0)
	txRecord := Tx{
		Block:    t.block,
		TxType:   TxBuyLimitX,
		Success:  (t.receipt.Ty == types.ExecOk),
		BuyLimit: action,
	}
	r := &dbTxOrder{IKey: newTxKey(t.block.TxHash), Op: db.NewOp(op), tx: &txRecord}
	rs = append(rs, r)

	if t.receipt.Ty != types.ExecOk {
		return rs, nil
	}

	rs2, err := t.convertBuyLimitLog(op)
	if err != nil {
		return rs, err
	}
	rs = append(rs, rs2...)

	rs = newTradeAsset(rs, &t.accountIDAsset, &t.accountIDPrice)

	return rs, nil
}

// 1 order, 2 account
func (t *tradeConvert) convertBuyMarket(action *pty.TradeForBuy, op int) ([]db.Record, error) {
	rs := make([]db.Record, 0)
	txRecord := Tx{
		Block:     t.block,
		TxType:    TxBuyMarketX,
		Success:   (t.receipt.Ty == types.ExecOk),
		BuyMarket: action,
	}
	r := &dbTxOrder{IKey: newTxKey(t.block.TxHash), Op: db.NewOp(op), tx: &txRecord}
	rs = append(rs, r)

	if t.receipt.Ty != types.ExecOk {
		return rs, nil
	}

	t.boardlotCount = action.BoardlotCnt

	rs2, err := t.convertBuyMarketLog(op)
	if err != nil {
		return rs, err
	}

	rs = append(rs, rs2...)
	return rs, nil
}

// 1 order, 2 account
func (t *tradeConvert) convertBuyRevoke(action *pty.TradeForRevokeBuy, op int) ([]db.Record, error) {
	rs := make([]db.Record, 0)
	txRecord := Tx{
		Block:     t.block,
		TxType:    TxBuyRevokeX,
		Success:   (t.receipt.Ty == types.ExecOk),
		BuyRevoke: action,
	}
	r := &dbTxOrder{IKey: newTxKey(t.block.TxHash), Op: db.NewOp(op), tx: &txRecord}
	rs = append(rs, r)

	if t.receipt.Ty != types.ExecOk {
		return rs, nil
	}

	rs2, err := t.convertBuyRevokeLog(op)
	if err != nil {
		return rs, err
	}

	rs = append(rs, rs2...)
	return rs, nil
}

type dbTxOrder struct {
	*db.IKey
	*db.Op

	tx *Tx
}

func newTxKey(id string) *db.IKey {
	return db.NewIKey(TradeTxDBX, TradeTxDBX, id)
}

type dbOrder struct {
	*db.IKey
	*db.Op

	order *Order
}

func newOrderKey(id string) *db.IKey {
	return db.NewIKey(TradeDBX, TradeDBX, id)
}

type dbAsset struct {
	*db.IKey
	*db.Op

	asset *Asset
}

func newAssetKey(id string) *db.IKey {
	return db.NewIKey(TradeAssetDBX, TradeAssetDBX, id)
}
