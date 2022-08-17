package ticket

import (
	"encoding/json"
	"fmt"

	"github.com/33cn/chain33/common"
	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/ticket/types"
	"github.com/pkg/errors"
	"github.com/33cn/externaldb/converts"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/account"
	"github.com/33cn/externaldb/db/transaction"
	"github.com/33cn/externaldb/util"
)

// 地址余额： 是否有多余的币，没有参与挖矿, 通过帐号信息可以看到
// 票的变化： 挖矿是否正常进行： 指定时间内，是否存在miner_at.ts > ts0 状态的 -> 查看矿机更合理， 可以更快的反映情况
//          挖矿收益： 统计指定时间内，挖矿成功的数量

const (
	ticketStatusOpen  = "open"
	ticketStatusMiner = "miner"
	ticketStatusClose = "close"
)

// BlockInfo ticket/ticket/Id
type BlockInfo struct {
	Height int64 `json:"height"`
	Ts     int64 `json:"ts"`
}

// Ticket Ticket
type Ticket struct {
	ID      string    `json:"id"`
	Owner   string    `json:"owner"`
	Miner   string    `json:"miner"`
	Status  string    `json:"status"`
	OpenAt  BlockInfo `json:"open_at"`
	MinerAt BlockInfo `json:"miner_at"`
	CloseAt BlockInfo `json:"close_at"`
}

type ticketUpdate struct {
	ID      string     `json:"id"`
	Miner   string     `json:"miner"`
	Status  string     `json:"status"`
	OpenAt  *BlockInfo `json:"open_at,omitempty"`
	MinerAt *BlockInfo `json:"miner_at,omitempty"`
	CloseAt *BlockInfo `json:"close_at,omitempty"`
}

type bind struct {
	OldMiner      string `json:"old_miner"`
	NewMiner      string `json:"new_miner"`
	ReturnAddress string `json:"return_address"`
	Height        int64  `json:"height"`
	Ts            int64  `json:"ts"`
}

const (
	//log for ticket

	//TyLogNewTicket new ticket log type
	TyLogNewTicket = 111
	// TyLogCloseTicket close ticket log type
	TyLogCloseTicket = 112
	// TyLogMinerTicket miner ticket log type
	TyLogMinerTicket = 113
	// TyLogTicketBind bind ticket log type
	TyLogTicketBind = 114
)

var log = l.New("module", "db.ticket")

func init() {
	converts.Register("ticket", NewConvert)
}

// NewConvert NewConvert
func NewConvert(paraTitle, symbol string, supports []string) db.ExecConvert {
	e := &ticketConvert{symbol: symbol, title: paraTitle}

	return e
}

// InitDB init db
func (t *ticketConvert) InitDB(cli db.DBCreator) error {
	var err error

	err = account.InitDB(cli)
	if err != nil {
		return err
	}

	err = transaction.InitDB(cli)
	if err != nil {
		return err
	}

	err = util.InitIndex(cli, TicketDBX, TicketTableX, TicketMapping)
	if err != nil {
		return err
	}

	err = util.InitIndex(cli, TicketBindDBX, TicketBindTableX, BindMapping)
	if err != nil {
		return err
	}

	return nil
}

type ticketConvert struct {
	symbol string
	title  string

	env *db.TxEnv

	tx      *types.Transaction
	receipt *types.ReceiptData
	block   *db.Block
	owner   string

	accountIDBty    account.Account // for fee
	accountIDTicket account.Account // for other
}

func (t *ticketConvert) positionID() string {
	return fmt.Sprintf("%s:%d.%d", "ticket", t.block.Height, t.block.Index)
}

func (t *ticketConvert) ConvertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	t.env = env
	tx := env.Block.Block.Txs[env.TxIndex]
	receipt := env.Block.Receipts[env.TxIndex]
	t.tx = tx
	t.receipt = receipt
	t.block = db.SetupBlock(env, tx.From(), common.ToHex(tx.Hash()))

	t.accountIDBty = account.Account{
		AssetSymbol: t.symbol,
		AssetExec:   account.ExecCoinsX,
		HeightIndex: db.HeightIndex(t.block.Height, t.block.Index),
		Height:      t.block.Height,
		BlockTime:   t.env.Block.Block.BlockTime,
	}
	t.accountIDTicket = account.Account{
		AssetSymbol: t.symbol,
		AssetExec:   account.ExecCoinsX,
		HeightIndex: db.HeightIndex(t.block.Height, t.block.Index),
		Height:      t.block.Height,
		BlockTime:   t.env.Block.Block.BlockTime,
	}

	var records []db.Record
	txRec := transaction.ConvertTransaction(t.env)
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(txRec.Hash),
		Op:   db.NewOp(op),
		Tx:   txRec,
	}
	records = append(records, &txRecord)

	rs, err := t.convertTx(env, op)
	if err != nil {
		return records, err
	}
	if rs != nil {
		records = append(records, rs...)
	}
	return records, nil
}

func (t *ticketConvert) convertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	receipt := env.Block.Receipts[env.TxIndex]
	tx := env.Block.Block.Txs[env.TxIndex]
	if receipt.Ty != types.ExecOk {
		return nil, nil
	}

	var action pty.TicketAction
	err := types.Decode(tx.Payload, &action)
	if err != nil {
		return nil, errors.Wrap(err, "decode tx action")
	}

	switch action.Ty {
	case pty.TicketActionGenesis:
		log.Debug("Convert", "action", "create")
		return t.convertGenesis(&action, op)
	case pty.TicketActionOpen:
		log.Debug("Convert", "action", "open")
		return t.convertOpen(&action, op)
	case pty.TicketActionClose:
		log.Debug("Convert", "action", "close")
		return t.convertClose(&action, op)
	case pty.TicketActionMiner:
		log.Debug("Convert", "action", "miner")
		return t.convertMiner(&action, op)
	case pty.TicketActionBind:
		log.Debug("Convert", "action", "bind")
		return t.convertBind(&action, op)
	}
	return nil, nil

}

func (t *ticketConvert) convertCommon(op int) ([]db.Record, error) {
	var items []db.Record
	var err error
	for _, l := range t.receipt.Logs {
		var record db.Record
		switch l.Ty {
		case TyLogNewTicket:
			record, err = t.LogNewTicketConvert(l.Log, op)
		case TyLogCloseTicket:
			record, err = t.LogCloseTicketConvert(l.Log, op)
		case TyLogTicketBind:
			record, err = t.LogBindTicketConvert(l.Log, op)
		case TyLogMinerTicket:
			record, err = t.LogMinerTicketConvert(l.Log, op)
		default:
			accountDetail, err2 := account.AssetLogConvert(l.Ty, l.Log, op)
			if err2 != nil {
				log.Info("convertTX", "position", t.positionID(), "logType", l.Ty, "err", err)
				continue
			}
			acc := t.accountIDBty
			acc.Detall = accountDetail
			record = &account.Record{Acc: acc, IKey: account.NewAccountKey(acc.Key()), Op: db.NewOp(db.OpAdd)}

			rrecord := &account.Record{Acc: acc, IKey: account.NewAccountRecordKey(acc.RecordKey()), Op: db.NewOp(op)}
			items = append(items, rrecord)
		}
		items = append(items, record)
	}
	return items, err
}

func (t *ticketConvert) convertGenesis(action *pty.TicketAction, op int) ([]db.Record, error) {
	g := action.GetGenesis()
	t.owner = g.GetReturnAddress()
	return t.convertCommon(op)
}

func (t *ticketConvert) convertOpen(action *pty.TicketAction, op int) ([]db.Record, error) {
	g := action.GetTopen()
	t.owner = g.GetReturnAddress()
	return t.convertCommon(op)
}

func (t *ticketConvert) convertClose(action *pty.TicketAction, op int) ([]db.Record, error) {
	return t.convertCommon(op)
}

func (t *ticketConvert) convertMiner(action *pty.TicketAction, op int) ([]db.Record, error) {
	return t.convertCommon(op)
}

func (t *ticketConvert) convertBind(action *pty.TicketAction, op int) ([]db.Record, error) {
	return t.convertCommon(op)
}

// ticket db
const (
	TicketDBX        = "ticket"
	TicketTableX     = "ticket"
	TicketBindDBX    = "ticket_bind"
	TicketBindTableX = "ticket"
	TicketSeqDBX     = "ticket_seq"
	TicketSeqTableX  = "ticket"
	TicketLastSeqX   = "last_seq"
	DefaultType      = "_doc"
)

func newTicketKey(id string) *db.IKey {
	return db.NewIKey(TicketDBX, TicketTableX, id)
}

func newBindKey(id string) *db.IKey {
	return db.NewIKey(TicketBindDBX, TicketBindTableX, id)
}

func (t *ticketConvert) LogNewTicketConvert(v []byte, op int) (db.Record, error) {
	var l pty.ReceiptTicket
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "ticket Decode new ticket log")
	}

	t2 := &dbTicket{IKey: newTicketKey(l.TicketId), Op: db.NewOp(op), owner: t.owner}
	if op == db.SeqTypeDel {
		return t2, nil
	}

	open := BlockInfo{
		Height: t.block.Height,
		Ts:     t.block.Ts,
	}
	t2.id, t2.minerAddress, t2.status = l.TicketId, l.Addr, ticketStatusOpen
	t2.open = &open

	return t2, nil
}

func (t *ticketConvert) LogCloseTicketConvert(v []byte, op int) (db.Record, error) {
	var l pty.ReceiptTicket
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "ticket Decode Close ticket log")
	}
	t2 := &dbTicket{IKey: newTicketKey(l.TicketId), Op: db.NewOp(db.OpUpdate)}
	t2.id, t2.minerAddress, t2.status = l.TicketId, l.Addr, ticketStatusClose

	if op == db.SeqTypeDel {
		t2.close = &BlockInfo{}
		t2.status = ticketStatusOpen
		if l.PrevStatus == 2 {
			t2.status = ticketStatusMiner
		}
		return t2, nil
	}

	t2.close = &BlockInfo{
		Height: t.block.Height,
		Ts:     t.block.Ts,
	}
	return t2, nil
}

func (t *ticketConvert) LogMinerTicketConvert(v []byte, op int) (db.Record, error) {
	var l pty.ReceiptTicket
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "ticket Decode Miner ticket log")
	}

	t2 := &dbTicket{IKey: newTicketKey(l.TicketId), Op: db.NewOp(db.OpUpdate)}
	t2.id, t2.minerAddress, t2.status = l.TicketId, l.Addr, ticketStatusMiner

	if op == db.SeqTypeDel {
		t2.miner = &BlockInfo{}
		t2.status = ticketStatusOpen
		return t2, nil
	}

	t2.miner = &BlockInfo{
		Height: t.block.Height,
		Ts:     t.block.Ts,
	}

	return t2, nil
}

func (t *ticketConvert) LogBindTicketConvert(v []byte, op int) (db.Record, error) {
	var l pty.ReceiptTicketBind
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "ticket Decode bind ticket log")
	}

	b := &dbBind{IKey: newBindKey(l.ReturnAddress),
		Op: db.NewOp(db.OpAdd),
		current: bind{
			OldMiner:      l.OldMinerAddress,
			NewMiner:      l.NewMinerAddress,
			ReturnAddress: l.ReturnAddress,
			Height:        t.block.Height,
			Ts:            t.block.Ts,
		},
	}

	if op == db.SeqTypeDel {
		b.current.NewMiner = b.current.OldMiner
	}

	return b, nil

}

type dbTicket struct {
	*db.IKey
	*db.Op

	id, minerAddress, status string
	open, miner, close       *BlockInfo
	owner                    string
}

func (r *dbTicket) Value() []byte {
	if r.Op.OpType() == db.OpDel {
		return nil
	}

	if r.Op.OpType() == db.OpAdd {
		t := Ticket{
			ID:     r.id,
			Miner:  r.minerAddress,
			Status: r.status,
			OpenAt: *r.open,
			Owner:  r.owner,
		}
		if r.close != nil {
			t.CloseAt = *r.close
		}
		if r.miner != nil {
			t.MinerAt = *r.miner
		}
		v, _ := json.Marshal(t)
		return v
	}
	t := ticketUpdate{
		ID:      r.id,
		Miner:   r.minerAddress,
		Status:  r.status,
		OpenAt:  r.open,
		MinerAt: r.miner,
		CloseAt: r.close,
	}
	v, _ := json.Marshal(t)
	return v
}

type dbBind struct {
	*db.IKey
	*db.Op
	current bind
}

func (r *dbBind) Value() []byte {
	v, _ := json.Marshal(r.current)
	return v
}
