package pos33

import (
	"encoding/json"
	"fmt"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/pkg/errors"
	pty "github.com/yccproject/ycc/plugin/dapp/pos33/types"
	"github.com/33cn/externaldb/converts"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/account"
	"github.com/33cn/externaldb/db/transaction"
	"github.com/33cn/externaldb/util"
)

var (
	TicketReward int64
	ChainHost    string
)

const (
	pos33TicketStatusOpen  = "open"
	pos33TicketStatusMiner = "miner"
	pos33TicketStatusVoter = "voter"
	pos33TicketStatusClose = "close"
)

// BlockInfo ticket/ticket/Id
type BlockInfo struct {
	Height int64 `json:"height"`
	Ts     int64 `json:"ts"`
}

// Ticket Pos33Ticket
type Ticket struct {
	Owner   string    `json:"owner"`
	Miner   string    `json:"miner"`
	Status  string    `json:"status"`
	Account int64     `json:"account"`
	OpenAt  BlockInfo `json:"open_at"`
	MinerAt BlockInfo `json:"miner_at"`
	CloseAt BlockInfo `json:"close_at"`
}

type pos33TicketUpdate struct {
	Miner   string     `json:"miner"`
	Status  string     `json:"status"`
	Account int64      `json:"account"`
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

var log = l.New("module", "db.pos33")

func init() {
	converts.Register("pos33", NewConvert)
}

// NewConvert NewConvert
func NewConvert(paraTitle, symbol string, supports []string) db.ExecConvert {
	e := &pos33TicketConvert{symbol: symbol, title: paraTitle}

	return e
}

// SetCoinsReward block ticket reward
func SetCoinsReward(precision, reward float64) {
	TicketReward = int64(precision * reward)
}

func SetHost(host string) {
	ChainHost = host
}

// InitDB init db
func (t *pos33TicketConvert) InitDB(cli db.DBCreator) error {
	var err error

	err = account.InitDB(cli)
	if err != nil {
		return err
	}

	err = transaction.InitDB(cli)
	if err != nil {
		return err
	}

	err = util.InitIndex(cli, Pos33TicketDBX, Pos33TicketTableX, Pos33TicketMapping)
	if err != nil {
		return err
	}

	err = util.InitIndex(cli, Pos33TicketBindDBX, Pos33TicketBindTableX, BindMapping)
	if err != nil {
		return err
	}

	return nil
}

type pos33TicketConvert struct {
	symbol string
	title  string

	env *db.TxEnv

	tx      *types.Transaction
	receipt *types.ReceiptData
	block   *db.Block
	owner   string

	accountIDYcc    account.Account // for fee
	accountIDTicket account.Account // for other
}

func (t *pos33TicketConvert) positionID() string {
	return fmt.Sprintf("%s:%d.%d", "pos33", t.block.Height, t.block.Index)
}

func (t *pos33TicketConvert) ConvertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	t.env = env
	tx := env.Block.Block.Txs[env.TxIndex]
	receipt := env.Block.Receipts[env.TxIndex]
	t.tx = tx
	t.receipt = receipt
	t.block = db.SetupBlock(env, tx.From(), common.ToHex(tx.Hash()))

	t.accountIDYcc = account.Account{
		AssetSymbol: t.symbol,
		AssetExec:   account.ExecCoinsxX,
		HeightIndex: db.HeightIndex(t.block.Height, t.block.Index),
		Height:      t.block.Height,
		BlockTime:   t.env.Block.Block.BlockTime,
	}
	t.accountIDTicket = account.Account{
		AssetSymbol: t.symbol,
		AssetExec:   account.ExecCoinsxX,
		HeightIndex: db.HeightIndex(t.block.Height, t.block.Index),
		Height:      t.block.Height,
		BlockTime:   t.env.Block.Block.BlockTime,
	}

	var records []db.Record
	txRec := transaction.ConvertTransaction(t.env)
	if env.TxIndex == 0 {
		addr, err := t.GetMinerAddr()
		if err != nil {
			log.Error("ConvertTx", "t.GetMinerAddr err", err)
		}
		txRec.AddrRecord = addr
	}
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

func (t *pos33TicketConvert) convertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	receipt := env.Block.Receipts[env.TxIndex]
	tx := env.Block.Block.Txs[env.TxIndex]
	if receipt.Ty != types.ExecOk {
		return nil, nil
	}

	var action pty.Pos33TicketAction
	err := types.Decode(tx.Payload, &action)
	if err != nil {
		return nil, errors.Wrap(err, "decode tx action")
	}

	switch action.Ty {
	case pty.Pos33TicketActionGenesis:
		log.Debug("Convert", "action", "create")
		return t.convertGenesis(&action, op)
	case pty.Pos33TicketActionOpen:
		log.Debug("Convert", "action", "open")
		return t.convertOpen(&action, op)
	case pty.Pos33TicketActionClose:
		log.Debug("Convert", "action", "close")
		return t.convertClose(&action, op)
	case pty.Pos33TicketActionMiner:
		log.Debug("Convert", "action", "miner")
		return t.convertMiner(&action, op)
	case pty.Pos33TicketActionBind:
		log.Debug("Convert", "action", "bind")
		return t.convertBind(&action, op)
	case pty.Pos33ActionEntrust:
		log.Debug("Convert", "action", "entrust")
		return t.convertBind(&action, op)
	case pty.Pos33ActionMigrate:
		log.Debug("Convert", "action", "migrate")
		return t.convertBind(&action, op)
	case pty.Pos33ActionBlsBind:
		log.Debug("Convert", "action", "bls bind")
		return t.convertBind(&action, op)
	case pty.Pos33ActionMinerFeeRate:
		log.Debug("Convert", "action", "set miner fee rate")
		return t.convertBind(&action, op)
	case pty.Pos33ActionWithdrawReward:
		log.Debug("Convert", "action", "withdraw reward")
		return t.convertBind(&action, op)
	}

	return nil, nil
}

func (t *pos33TicketConvert) convertCommon(op int) ([]db.Record, error) {
	var items []db.Record
	var err error
	for _, l := range t.receipt.Logs {
		var record db.Record
		switch l.Ty {
		case pty.TyLogNewPos33Ticket:
			record, err = t.LogNewPos33TicketConvert(l.Log, op)
		case pty.TyLogClosePos33Ticket:
			record, err = t.LogClosePos33TicketConvert(l.Log, op)
		case pty.TyLogPos33TicketBind:
			record, err = t.LogBindPos33TicketConvert(l.Log, op)
		case pty.TyLogMinerPos33Ticket:
			record, err = t.LogMinerPos33TicketConvert(l.Log, op)
		case pty.TyLogVoterPos33Ticket:
			record, err = t.LogVoterPos33TicketConvert(l.Log, op)
		default:
			accountDetail, err2 := account.AssetLogConvert(l.Ty, l.Log, op)
			if err2 != nil {
				log.Info("convertTX", "position", t.positionID(), "logType", l.Ty, "err", err)
				continue
			}

			acc := t.accountIDYcc
			acc.Detall = accountDetail
			record = &account.Record{Acc: acc, IKey: account.NewAccountKey(acc.Key()), Op: db.NewOp(db.OpAdd)}
			items = append(items, record)

			rrecord := &account.Record{Acc: acc, IKey: account.NewAccountRecordKey(acc.RecordKey()), Op: db.NewOp(op)}
			items = append(items, rrecord)
		}
		items = append(items, record)
	}

	return items, err
}

func (t *pos33TicketConvert) convertGenesis(action *pty.Pos33TicketAction, op int) ([]db.Record, error) {
	g := action.GetGenesis()
	t.owner = g.GetReturnAddress()
	return t.convertCommon(op)
}

func (t *pos33TicketConvert) convertOpen(action *pty.Pos33TicketAction, op int) ([]db.Record, error) {
	g := action.GetTopen()
	t.owner = g.GetReturnAddress()
	return t.convertCommon(op)
}

func (t *pos33TicketConvert) convertClose(action *pty.Pos33TicketAction, op int) ([]db.Record, error) {
	return t.convertCommon(op)
}

func (t *pos33TicketConvert) convertMiner(action *pty.Pos33TicketAction, op int) ([]db.Record, error) {
	return t.convertCommon(op)
}

func (t *pos33TicketConvert) convertBind(action *pty.Pos33TicketAction, op int) ([]db.Record, error) {
	return t.convertCommon(op)
}

func (t *pos33TicketConvert) convertEntrust(action *pty.Pos33TicketAction, op int) ([]db.Record, error) {
	return t.convertCommon(op)
}

func (t *pos33TicketConvert) convertMigrate(action *pty.Pos33TicketAction, op int) ([]db.Record, error) {
	return t.convertCommon(op)
}

func (t *pos33TicketConvert) convertBlsBind(action *pty.Pos33TicketAction, op int) ([]db.Record, error) {
	return t.convertCommon(op)
}

func (t *pos33TicketConvert) convertMinerFeeRate(action *pty.Pos33TicketAction, op int) ([]db.Record, error) {
	return t.convertCommon(op)
}

func (t *pos33TicketConvert) convertWithdrawReward(action *pty.Pos33TicketAction, op int) ([]db.Record, error) {
	return t.convertCommon(op)
}

func (t *pos33TicketConvert) LogNewPos33TicketConvert(v []byte, op int) (db.Record, error) {
	var l pty.ReceiptPos33Deposit
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "pos33Ticket Decode new ticket log")
	}

	t2 := &dbTicket{IKey: newTicketKey(l.Addr), Op: db.NewOp(op), owner: t.owner}
	if op == db.SeqTypeDel {
		return t2, nil
	}

	open := BlockInfo{
		Height: t.block.Height,
		Ts:     t.block.Ts,
	}
	t2.minerAddress, t2.status, t2.account = l.Addr, pos33TicketStatusOpen, l.Count
	t2.open = &open

	return t2, nil
}

func (t *pos33TicketConvert) LogClosePos33TicketConvert(v []byte, op int) (db.Record, error) {
	var l pty.ReceiptPos33Deposit
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "pos33Ticket Decode Close ticket log")
	}
	t2 := &dbTicket{IKey: newTicketKey(l.Addr), Op: db.NewOp(db.OpUpdate)}
	t2.minerAddress, t2.status, t2.account = l.Addr, pos33TicketStatusClose, l.Count

	if op == db.SeqTypeDel {
		t2.close = &BlockInfo{}
		t2.status = pos33TicketStatusOpen
		return t2, nil
	}

	t2.close = &BlockInfo{
		Height: t.block.Height,
		Ts:     t.block.Ts,
	}
	return t2, nil
}

func (t *pos33TicketConvert) LogMinerPos33TicketConvert(v []byte, op int) (db.Record, error) {
	var l pty.ReceiptPos33Miner
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "pos33Ticket Decode Miner ticket log")
	}

	t2 := &dbTicket{IKey: newTicketKey(l.Addr), Op: db.NewOp(db.OpUpdate)}
	t2.minerAddress, t2.status, t2.account = l.Addr, pos33TicketStatusVoter, l.Reward

	if op == db.SeqTypeDel {
		t2.miner = &BlockInfo{}
		t2.status = pos33TicketStatusOpen
		return t2, nil
	}

	t2.miner = &BlockInfo{
		Height: t.block.Height,
		Ts:     t.block.Ts,
	}

	return t2, nil
}

func (t *pos33TicketConvert) LogVoterPos33TicketConvert(v []byte, op int) (db.Record, error) {
	var l pty.ReceiptPos33Miner
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "pos33Ticket Decode Voter ticket log")
	}

	t2 := &dbTicket{IKey: newTicketKey(l.Addr), Op: db.NewOp(db.OpUpdate)}
	t2.minerAddress, t2.status, t2.account = l.Addr, pos33TicketStatusMiner, l.Reward

	if op == db.SeqTypeDel {
		t2.miner = &BlockInfo{}
		t2.status = pos33TicketStatusOpen
		return t2, nil
	}

	t2.miner = &BlockInfo{
		Height: t.block.Height,
		Ts:     t.block.Ts,
	}

	return t2, nil
}

func (t *pos33TicketConvert) LogBindPos33TicketConvert(v []byte, op int) (db.Record, error) {
	var l pty.ReceiptPos33TicketBind
	err := types.Decode(v, &l)
	if err != nil {
		return nil, errors.Wrap(err, "pos33Ticket Decode bind ticket log")
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

func (t *pos33TicketConvert) GetMinerAddr() (*transaction.AddrRecord, error) {
	var addrRecord transaction.AddrRecord

	// 获得投票地址
	var action pty.Pos33TicketAction
	err := types.Decode(t.tx.Payload, &action)
	if err != nil {
		log.Error("decode tx action", "err", err)
		return nil, err
	}

	blsList := action.GetMiner().BlsPkList
	for _, bls := range blsList {
		addr := address.PubKeyToAddr(pty.EthAddrID, bls)

		para := types.ReqAddr{Addr: addr}
		reqData, _ := json.Marshal(&para)
		res := types.ReplyString{}

		ctx := jsonclient.NewRPCCtx(ChainHost, "Chain33.Query", &rpctypes.Query4Jrpc{
			Execer:   "pos33",
			FuncName: "Pos33BlsAddr",
			Payload:  reqData,
		}, &res)
		_, err = ctx.RunResult()
		if err != nil {
			log.Error("Chain33.Query Pos33BlsAddr", "err", err)
			return nil, err
		}
		addrRecord.VoterAddr = append(addrRecord.VoterAddr, res.Data)
	}
	// 获得打包地址
	addrRecord.MakerAddr = []string{t.tx.From()}

	return &addrRecord, nil
}
