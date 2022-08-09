package pos33

import (
	"encoding/json"

	"github.com/33cn/externaldb/db"
)

// ticket db
const (
	Pos33TicketDBX        = "pos33_ticket"
	Pos33TicketTableX     = "pos33_ticket"
	Pos33TicketBindDBX    = "pos33_ticket_bind"
	Pos33TicketBindTableX = "pos33_ticket_bind"
	DefaultType           = "_doc"
)

type dbTicket struct {
	*db.IKey
	*db.Op

	minerAddress, status string
	account              int64
	open, miner, close   *BlockInfo
	owner                string
}

func (r *dbTicket) Value() []byte {
	if r.Op.OpType() == db.OpDel {
		return nil
	}

	if r.Op.OpType() == db.OpAdd {
		t := Ticket{
			Miner:   r.minerAddress,
			Status:  r.status,
			Account: r.account,
			OpenAt:  *r.open,
			Owner:   r.owner,
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
	t := pos33TicketUpdate{
		Miner:   r.minerAddress,
		Status:  r.status,
		Account: r.account,
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

func newTicketKey(id string) *db.IKey {
	return db.NewIKey(Pos33TicketDBX, Pos33TicketTableX, id)
}

func newBindKey(id string) *db.IKey {
	return db.NewIKey(Pos33TicketBindDBX, Pos33TicketBindTableX, id)
}
