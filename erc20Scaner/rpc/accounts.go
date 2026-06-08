package main

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/33cn/externaldb/erc20Scaner/database"
)

// handleAccountsRouter 处理 /evmapi/accounts 路径
func handleAccountsRouter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/evmapi/accounts")
	if path == "" || path == "/" {
		writeError(w, http.StatusBadRequest, "Account address is required, e.g. /evmapi/accounts/{address}/erc20-balances")
		return
	}
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")

	// GET /evmapi/accounts/{address}/erc20-balances
	if len(parts) == 2 && parts[1] == "erc20-balances" {
		handleAccountERC20Balances(w, r, parts[0])
		return
	}

	// GET /evmapi/accounts/{address}/approvals
	if len(parts) == 2 && parts[1] == "approvals" {
		handleAccountApprovals(w, r, parts[0])
		return
	}

	writeError(w, http.StatusNotFound, "Invalid path")
}

// handleAccountERC20Balances 按用户地址查询其持有的 ERC20 列表及余额（来自 token_balances + contracts）
// GET /evmapi/accounts/{address}/erc20-balances?page=1&size=20&min_balance=0
func handleAccountERC20Balances(w http.ResponseWriter, r *http.Request, holderAddress string) {
	if !strings.HasPrefix(strings.ToLower(holderAddress), "0x") || len(holderAddress) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid account address format")
		return
	}
	holderAddress = normalizeAddress(holderAddress)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if size < 1 || size > 100 {
		size = 20
	}
	minBalanceStr := r.URL.Query().Get("min_balance")

	offset := (page - 1) * size
	rows, total, err := db.ListERC20BalancesByHolderAddress(holderAddress, minBalanceStr, size, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
		return
	}

	tokens := make([]AccountERC20TokenBalance, 0, len(rows))
	for _, row := range rows {
		tokens = append(tokens, accountERC20TokenBalanceFromRow(row))
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "Success",
		Data: map[string]interface{}{
			"address": holderAddress,
			"tokens":  tokens,
			"page":    page,
			"size":    size,
			"total":   total,
		},
	})
}

func accountERC20TokenBalanceFromRow(row database.AccountERC20BalanceRow) AccountERC20TokenBalance {
	bal := row.Balance
	if bal == nil {
		bal = big.NewInt(0)
	}
	return AccountERC20TokenBalance{
		ContractAddress:  normalizeAddress(row.ContractAddress),
		Name:             row.ContractName,
		Symbol:           row.ContractSymbol,
		Decimals:         row.Decimals,
		Balance:          bal.String(),
		BalanceFormatted: formatTokenAmount(bal, row.Decimals),
		LastTxHash:       row.LastTxHash,
		LastTxBlock:      row.LastTxBlock,
		LastUpdated:      row.LastUpdatedAt,
	}
}

// handleAccountApprovals 按 owner 地址查询所有授权记录
// GET /evmapi/accounts/{address}/approvals?page=1&size=20&min_amount=0&contract_address=0x...
func handleAccountApprovals(w http.ResponseWriter, r *http.Request, ownerAddress string) {
	if !strings.HasPrefix(strings.ToLower(ownerAddress), "0x") || len(ownerAddress) != 42 {
		writeError(w, http.StatusBadRequest, "Invalid account address format")
		return
	}
	ownerAddress = normalizeAddress(ownerAddress)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if size < 1 || size > 100 {
		size = 20
	}
	minAmountStr := r.URL.Query().Get("min_amount")
	contractFilter := r.URL.Query().Get("contract_address")
	if contractFilter != "" {
		contractFilter = normalizeAddress(contractFilter)
	}

	offset := (page - 1) * size
	rows, total, err := db.ListAllowancesByOwner(ownerAddress, minAmountStr, contractFilter, size, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
		return
	}

	approvals := make([]ApprovalInfo, 0, len(rows))
	for _, row := range rows {
		amt := row.Amount
		if amt == nil {
			amt = big.NewInt(0)
		}
		approvals = append(approvals, ApprovalInfo{
			ContractAddress: normalizeAddress(row.ContractAddress),
			ContractName:    row.ContractName,
			ContractSymbol:  row.ContractSymbol,
			Decimals:        row.Decimals,
			Spender:         normalizeAddress(row.Spender),
			Amount:          amt.String(),
			AmountFormatted: formatTokenAmount(amt, row.Decimals),
			LastTxHash:      row.LastTxHash,
			LastBlockNumber: row.LastBlockNumber,
			LastUpdated:     row.LastUpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Code:    0,
		Message: "Success",
		Data: map[string]interface{}{
			"address":   ownerAddress,
			"approvals": approvals,
			"page":      page,
			"size":      size,
			"total":     total,
		},
	})
}
