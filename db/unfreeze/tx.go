package unfreeze

import "github.com/33cn/externaldb/db"

// 1. unfreeze tx

// unfreeze ActionType
const (
	ActionTypeCreate    = "create"
	ActionTypeWithdraw  = "withdraw"
	ActionTypeTerminate = "terminate"
)

// Tx tx
type Tx struct {
	BlockInfo *db.Block `json:"block"`

	// tx info
	Creator     string `json:"creator"`
	Beneficiary string `json:"beneficiary"`
	UnfreezeID  string `json:"unfreeze_id"`
	Success     bool   `json:"success"`
	// tx action
	ActionType string `json:"action_type"`

	// action detail
	Create    *ActionCreate    `json:"create,omitempty"`
	Withdraw  *ActionWithdraw  `json:"withdraw,omitempty"`
	Terminate *ActionTerminate `json:"terminate,omitempty"`
}

// ActionTerminate impl
type ActionTerminate struct {
	AmountBack int64 `json:"amount_back"`
	AmountLeft int64 `json:"amount_left"`
}

// ActionWithdraw impl
type ActionWithdraw struct {
	Amount int64 `json:"amount"`
}

// ActionCreate impl
type ActionCreate struct {
	StartTime             int64           `json:"start_time"`
	AssetExec             string          `json:"asset_exec"`
	AssetSymbol           string          `json:"asset_symbol"`
	TotalCount            int64           `json:"total_count"`
	Means                 string          `json:"means"`
	FixAmountOption       *FixAmount      `json:"fix_amount,omitempty"`
	LeftProportionOptioin *LeftProportion `json:"left_proportion,omitempty"`
}

// FixAmount option
type FixAmount struct {
	Period int64 `json:"period"`
	Amount int64 ` json:"amount"`
}

// LeftProportion option
type LeftProportion struct {
	Period        int64 `json:"period"`
	TenThousandth int64 `json:"tenThousandth"`
}
