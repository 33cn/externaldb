package db

import (
	"encoding/json"
	"fmt"

	proofconfig "github.com/33cn/externaldb/db/proof_config"
)

type Transfer struct {
	From         string  `json:"from"`
	FromName     string  `json:"from_name"`
	FromRealName string  `json:"from_real_name"`
	FromUserName string  `json:"from_user_name"`
	To           string  `json:"to"`
	ToName       string  `json:"to_name"`
	ToRealName   string  `json:"to_real_name"`
	ToUserName   string  `json:"to_user_name"`
	IDs          []int64 `json:"ids"`
	Amounts      []int64 `json:"amounts"`
	Data         []byte  `json:"data"`
	EvmState
}

func NewTransfer(event, info map[string]interface{}, confDb proofconfig.ConfigDB) (*Transfer, error) {
	t := Transfer{}
	buf, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buf, &t)
	if err != nil {
		return nil, err
	}
	from, err := confDb.GetMember(t.From)
	if err == nil {
		t.FromUserName = from.GetUserName()
		t.FromName = from.GetUserName()
		if from.GetUserRealName() != "" {
			t.FromRealName = from.GetUserRealName()
			t.FromName = from.GetUserRealName()
		}
	}
	to, err := confDb.GetMember(t.To)
	if err == nil {
		t.ToUserName = to.GetUserName()
		t.ToName = to.GetUserName()
		if to.GetUserRealName() != "" {
			t.ToRealName = to.GetUserRealName()
			t.ToName = to.GetUserRealName()
		}
	}
	t.GetEvmState(info)
	return &t, nil
}

// Key for transfer index id
func (transfer *Transfer) Key() string {
	return fmt.Sprintf("%s-%v", "transfer", transfer.TxHash)
}
