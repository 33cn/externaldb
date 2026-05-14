package txinspect

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/33cn/externaldb/erc20Scaner/erc20abi/generated"
)

// Analyzer uses the same RPC and ABI calls as scanner Process.
type Analyzer struct {
	c *Client
}

// NewAnalyzer wraps a connected Client.
func NewAnalyzer(c *Client) *Analyzer {
	return &Analyzer{c: c}
}

// Report is the full inspection output (scanner-equivalent fields + raw logs).
type Report struct {
	TxHash string `json:"txHash"`

	ScannerWouldProcess bool   `json:"scannerWouldProcess"`
	ScannerSkipReason   string `json:"scannerSkipReason,omitempty"`

	Transaction *TxSummary   `json:"transaction"`
	Receipt     *ReceiptView `json:"receipt"`
	Block       *BlockView   `json:"block"`

	// Same paths as scanner
	ContractCreation *ContractCreationView `json:"contractCreation,omitempty"`
	ERC20Path        *ERC20PathView        `json:"erc20TransferPath,omitempty"`

	AllLogs []LogView `json:"allLogs"`
}

// TxSummary mirrors fields scanner uses / persists.
type TxSummary struct {
	From            string `json:"from"`
	To              string `json:"to,omitempty"`
	ContractAddress string `json:"contractAddress,omitempty"` // receipt for create
	Nonce           uint64 `json:"nonce"`
	Gas             uint64 `json:"gasLimit"`
	GasPrice        string `json:"gasPrice,omitempty"`
	GasTipCap       string `json:"maxPriorityFeePerGas,omitempty"`
	GasFeeCap       string `json:"maxFeePerGas,omitempty"`
	ValueWei        string `json:"valueWei"`
	ChainID         string `json:"chainId,omitempty"`
	Type            uint8  `json:"type"`
	DataHex         string `json:"dataHex"`
	FuncSelector    string `json:"funcSelector,omitempty"` // first 4 bytes if any
	IsContractCall  bool   `json:"isContractCall"`
}

// ReceiptView fields used by scanner / DB.
type ReceiptView struct {
	Status            uint64 `json:"status"`
	CumulativeGasUsed uint64 `json:"cumulativeGasUsed"`
	GasUsed           uint64 `json:"gasUsed"`
	TransactionIndex  uint64 `json:"transactionIndex"`
	BlockNumber       uint64 `json:"blockNumber"`
	BlockHash         string `json:"blockHash"`
	TxHash            string `json:"transactionHash"`
	ContractAddress   string `json:"contractAddress,omitempty"`
	LogsCount         int    `json:"logsCount"`
}

// BlockView minimal context scanner logs.
type BlockView struct {
	Number   uint64 `json:"number"`
	Hash     string `json:"hash"`
	TimeUnix uint64 `json:"timeUnix"`
	TimeCST  string `json:"timeAsiaShanghai"`
	TxIndex  int    `json:"txIndexInBlock"`
}

// LogView full log row (scanner saveEvent uses subset).
type LogView struct {
	Index           uint     `json:"logIndex"`
	Address         string   `json:"address"`
	Topics          []string `json:"topics"`
	Data            string   `json:"data"`
	Removed         bool     `json:"removed"`
	BlockNumber     uint64   `json:"blockNumber"`
	TxHash          string   `json:"transactionHash"`
	TxIndex         uint     `json:"transactionIndex"`
	BlockHash       string   `json:"blockHash"`
	IsTransferEvent bool     `json:"isERC20TransferShape"`
}

// ContractCreationView matches handleContractCreation / save paths.
type ContractCreationView struct {
	DeployedAddress string           `json:"deployedAddress"`
	ERC20           ERC20CheckResult `json:"erc20Check"`
	// If ERC20, same ABI reads as scanner
	Name        string `json:"name,omitempty"`
	Symbol      string `json:"symbol,omitempty"`
	Decimals    string `json:"decimals,omitempty"`
	TotalSupply string `json:"totalSupply,omitempty"`
	// DB-shaped row (scanner saveContractToDB / saveNonERC20ContractToDB)
	DBContractPreview *ContractDBPreview `json:"dbContractRowPreview,omitempty"`
}

// ContractDBPreview mirrors database.Contract fields scanner writes.
type ContractDBPreview struct {
	ContractAddress    string `json:"contract_address"`
	ContractName       string `json:"contract_name"`
	ContractSymbol     string `json:"contract_symbol"`
	ContractType       string `json:"contract_type"`
	Decimals           uint8  `json:"decimals"`
	TotalSupply        string `json:"total_supply,omitempty"`
	DeployTxHash       string `json:"deploy_tx_hash"`
	DeployBlockNumber  uint64 `json:"deploy_block_number"`
	DeployBlockTime    string `json:"deploy_block_time"`
	DeployerAddress    string `json:"deployer_address"`
	VerificationStatus int8   `json:"verification_status"`
	VerifiedFunctions  string `json:"verified_functions"`
}

// ERC20PathView matches parseERC20Transfer + saveTransactionToDB + saveEventToDB + balances.
type ERC20PathView struct {
	CalldataFuncSelector string `json:"calldataFuncSelector"`
	IsDirectTransferCall bool   `json:"calldataIsTransferOrTransferFrom"`

	Transfers []TransferInspect `json:"transfers"`

	// Per distinct token contract: same classification scanner uses inside loop
	PerTokenContract []TokenContractCallView `json:"perTokenContract"`

	// Single DB transaction row preview (scanner calls save once with last loop values — we document both)
	DBTransactionPreview *TransactionDBPreview `json:"dbTransactionRowPreview,omitempty"`

	ScannerNote string `json:"scannerNote,omitempty"`
}

type TokenContractCallView struct {
	TokenAddress      string `json:"tokenAddress"`
	IsDirectCall      bool   `json:"isDirectCall"`
	FuncName          string `json:"funcName"`
	FinalFuncSelector string `json:"finalFuncSelector"`
	TxToEqualsToken   bool   `json:"txToEqualsToken"`
}

type TransferInspect struct {
	LogIndex uint `json:"logIndex"`

	TokenAddress string    `json:"tokenAddress"`
	TokenInfo    TokenInfo `json:"tokenInfo"`

	From string `json:"from"`
	To   string `json:"to"`

	ValueRaw       string `json:"valueRaw"`
	ValueFormatted string `json:"valueFormatted"`

	// Event row preview (scanner saveEventToDB)
	DBEventPreview *EventDBPreview `json:"dbEventRowPreview,omitempty"`

	FromBalance *string `json:"fromBalanceCurrent,omitempty"`
	ToBalance   *string `json:"toBalanceCurrent,omitempty"`
	BalanceErr  string  `json:"balanceQueryError,omitempty"`
}

// TransactionDBPreview mirrors database.Transaction.
type TransactionDBPreview struct {
	TxHash          string `json:"tx_hash"`
	BlockNumber     uint64 `json:"block_number"`
	BlockHash       string `json:"block_hash"`
	BlockTime       string `json:"block_time"`
	TxIndex         uint   `json:"tx_index"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	ContractAddress string `json:"contract_address"`
	FuncSelector    string `json:"func_selector"`
	FuncName        string `json:"func_name"`
	Value           string `json:"value,omitempty"`
	GasLimit        uint64 `json:"gas_limit"`
	GasUsed         uint64 `json:"gas_used"`
	GasPrice        string `json:"gas_price,omitempty"`
	TxFee           string `json:"tx_fee,omitempty"`
	Status          int8   `json:"status"`
	TxData          string `json:"tx_data"`
}

// EventDBPreview mirrors database.Event (subset scanner sets).
type EventDBPreview struct {
	TxHash          string `json:"tx_hash"`
	BlockNumber     uint64 `json:"block_number"`
	BlockTime       string `json:"block_time"`
	LogIndex        uint   `json:"log_index"`
	ContractAddress string `json:"contract_address"`
	EventName       string `json:"event_name"`
	EventSignature  string `json:"event_signature"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Value           string `json:"value"`
	Topic0          string `json:"topic0"`
	Topic1          string `json:"topic1"`
	Topic2          string `json:"topic2"`
}

var transferTopic0 = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

// Inspect loads tx, receipt, block and builds a Report.
func (a *Analyzer) Inspect(txHash common.Hash) (*Report, error) {
	tx, pending, err := a.c.TxByHash(txHash)
	if err != nil {
		return nil, fmt.Errorf("TransactionByHash: %w", err)
	}
	if pending {
		return nil, fmt.Errorf("transaction is still pending")
	}

	receipt, err := a.c.TxReceipt(txHash)
	if err != nil {
		return nil, fmt.Errorf("TransactionReceipt: %w", err)
	}
	if receipt == nil {
		return nil, fmt.Errorf("nil receipt")
	}

	block, _, err := a.c.BlockByNumber(receipt.BlockNumber.Uint64())
	if err != nil {
		return nil, fmt.Errorf("BlockByNumber %d: %w", receipt.BlockNumber.Uint64(), err)
	}

	txIdx, ok := txIndexInBlock(block, txHash)
	if !ok {
		return nil, fmt.Errorf("tx not found in block %d (hash mismatch?)", receipt.BlockNumber.Uint64())
	}

	return a.buildReport(tx, receipt, block, txIdx)
}

func blockView(block *types.Block, txIdx int) *BlockView {
	bt := time.Unix(int64(block.Time()), 0)
	cst, _ := time.LoadLocation("Asia/Shanghai")
	return &BlockView{
		Number:   block.NumberU64(),
		Hash:     block.Hash().Hex(),
		TimeUnix: block.Time(),
		TimeCST:  bt.In(cst).Format("2006-01-02 15:04:05"),
		TxIndex:  txIdx,
	}
}

func txIndexInBlock(block *types.Block, h common.Hash) (int, bool) {
	for i, t := range block.Transactions() {
		if t == nil {
			continue
		}
		if t.Hash() == h {
			return i, true
		}
	}
	return 0, false
}

func topicsHex(ts []common.Hash) []string {
	out := make([]string, len(ts))
	for i, t := range ts {
		out[i] = t.Hex()
	}
	return out
}

func transferEventID() (common.Hash, error) {
	parsedAbi, err := abi.JSON(strings.NewReader(generated.ERC20ABI))
	if err != nil {
		return common.Hash{}, err
	}
	ev := parsedAbi.Events["Transfer"]
	if ev.ID == (common.Hash{}) {
		return common.Hash{}, fmt.Errorf("transfer event not in ABI")
	}
	return ev.ID, nil
}

func (a *Analyzer) buildContractCreationView(receipt *types.Receipt, block *types.Block, tx *types.Transaction, blockTime time.Time, cst *time.Location) *ContractCreationView {
	addr := receipt.ContractAddress
	cv := &ContractCreationView{
		DeployedAddress: addr.Hex(),
		ERC20:           a.checkERC20BySelector(&addr),
	}

	var deployer string
	if tx.ChainId() != nil {
		signer := types.NewEIP155Signer(tx.ChainId())
		if s, err := types.Sender(signer, tx); err == nil {
			deployer = normalizeAddress(s.Hex())
		}
	}

	tStr := blockTime.In(cst).Format("2006-01-02 15:04:05")

	if !cv.ERC20.IsERC20 {
		cv.DBContractPreview = &ContractDBPreview{
			ContractAddress:    normalizeAddress(addr.Hex()),
			ContractName:       "Unknown",
			ContractSymbol:     "UNKNOWN",
			ContractType:       "UNKNOWN",
			Decimals:           0,
			TotalSupply:        "0",
			DeployTxHash:       tx.Hash().Hex(),
			DeployBlockNumber:  block.NumberU64(),
			DeployBlockTime:    tStr,
			DeployerAddress:    deployer,
			VerificationStatus: 0,
			VerifiedFunctions:  "[]",
		}
		return cv
	}

	cname, _ := a.unPackageAbi("name", &addr)
	symbol, _ := a.unPackageAbi("symbol", &addr)
	decimals, _ := a.unPackageAbi("decimals", &addr)
	supply, _ := a.unPackageAbi("totalSupply", &addr)

	cv.Name = fmt.Sprintf("%v", cname)
	cv.Symbol = fmt.Sprintf("%v", symbol)
	cv.Decimals = fmt.Sprintf("%v", decimals)
	cv.TotalSupply = fmt.Sprintf("%v", supply)

	decVal := uint8(18)
	if decimals != nil {
		switch v := decimals.(type) {
		case uint8:
			decVal = v
		case uint64:
			decVal = uint8(v)
		case *big.Int:
			decVal = uint8(v.Uint64())
		}
	}
	var totalSupply *big.Int
	if supply != nil {
		switch v := supply.(type) {
		case *big.Int:
			totalSupply = v
		case uint64:
			totalSupply = big.NewInt(int64(v))
		}
	}
	supStr := ""
	if totalSupply != nil {
		supStr = totalSupply.String()
	}

	verifiedFuncs := `["18160ddd","70a08231","a9059cbb","23b872dd","095ea7b3","dd62ed3e"]`
	cv.DBContractPreview = &ContractDBPreview{
		ContractAddress:    normalizeAddress(addr.Hex()),
		ContractName:       cv.Name,
		ContractSymbol:     cv.Symbol,
		ContractType:       "ERC20",
		Decimals:           decVal,
		TotalSupply:        supStr,
		DeployTxHash:       tx.Hash().Hex(),
		DeployBlockNumber:  block.NumberU64(),
		DeployBlockTime:    tStr,
		DeployerAddress:    deployer,
		VerificationStatus: 1,
		VerifiedFunctions:  verifiedFuncs,
	}
	return cv
}

func (a *Analyzer) buildERC20Path(tx *types.Transaction, receipt *types.Receipt, block *types.Block, transferEventID common.Hash, fromAddr common.Address, blockTime time.Time, cst *time.Location) *ERC20PathView {
	txData := tx.Data()
	if len(txData) < 4 {
		return nil
	}
	funcSelector := hex.EncodeToString(txData[:4])
	transferSelector := "a9059cbb"
	transferFromSelector := "23b872dd"
	isDirectTransfer := funcSelector == transferSelector || funcSelector == transferFromSelector

	type transferWithIdx struct {
		token   common.Address
		from    common.Address
		to      common.Address
		value   *big.Int
		logIdx  uint
		tokInfo TokenInfo
	}
	var transfers []transferWithIdx

	for logIndex, lg := range receipt.Logs {
		if lg.Removed || len(lg.Topics) < 3 || lg.Topics[0] != transferEventID {
			continue
		}
		from := common.BytesToAddress(lg.Topics[1].Bytes())
		to := common.BytesToAddress(lg.Topics[2].Bytes())
		if len(lg.Data) < 32 {
			continue
		}
		value := new(big.Int).SetBytes(lg.Data[:32])
		tokenAddr := lg.Address
		ti, err := a.getTokenInfo(&tokenAddr)
		if err != nil {
			ti = &TokenInfo{Address: tokenAddr.Hex(), Symbol: "UNKNOWN", Decimals: 18, Name: "Unknown Token"}
		}
		transfers = append(transfers, transferWithIdx{
			token:   tokenAddr,
			from:    from,
			to:      to,
			value:   value,
			logIdx:  uint(logIndex),
			tokInfo: *ti,
		})
	}

	if len(transfers) == 0 {
		return nil
	}

	out := &ERC20PathView{
		CalldataFuncSelector: funcSelector,
		IsDirectTransferCall: isDirectTransfer,
	}

	contractAddresses := make(map[common.Address]struct{})
	for _, t := range transfers {
		contractAddresses[t.token] = struct{}{}
	}

	var lastName, lastSel string
	var lastDirect bool
	for ca := range contractAddresses {
		isDirect := false
		if tx.To() != nil && *tx.To() == ca {
			if funcSelector == transferSelector || funcSelector == transferFromSelector {
				isDirect = true
			}
		}
		fn := "transfer"
		fs := "nested_call"
		if isDirect {
			fs = funcSelector
			if funcSelector == transferFromSelector {
				fn = "transferFrom"
			} else {
				fn = "transfer"
			}
		}
		lastName, lastSel, lastDirect = fn, fs, isDirect
		out.PerTokenContract = append(out.PerTokenContract, TokenContractCallView{
			TokenAddress:      ca.Hex(),
			IsDirectCall:      isDirect,
			FuncName:          fn,
			FinalFuncSelector: fs,
			TxToEqualsToken:   tx.To() != nil && *tx.To() == ca,
		})
	}

	// Note: scanner overwrites loop locals; last token wins for saveTransactionToDB
	out.ScannerNote = "saveTransactionToDB in scanner uses funcName/funcSelector from the last iteration over distinct token contracts in the map (Go map iteration order varies)."

	gasPrice := tx.GasPrice()
	var txFee *big.Int
	if gasPrice != nil && receipt.GasUsed > 0 {
		txFee = new(big.Int).Mul(gasPrice, big.NewInt(int64(receipt.GasUsed)))
	}

	var calldataValue *big.Int
	if funcSelector == "a9059cbb" && len(txData) >= 68 {
		calldataValue = new(big.Int).SetBytes(txData[36:68])
	} else if funcSelector == "23b872dd" && len(txData) >= 100 {
		calldataValue = new(big.Int).SetBytes(txData[68:100])
	}

	fromStr := ""
	if fromAddr != (common.Address{}) {
		fromStr = normalizeAddress(fromAddr.Hex())
	}
	toStr := ""
	if tx.To() != nil {
		toStr = normalizeAddress(tx.To().Hex())
	}
	valStr := ""
	if calldataValue != nil {
		valStr = calldataValue.String()
	}
	gpStr := ""
	if gasPrice != nil {
		gpStr = gasPrice.String()
	}
	feeStr := ""
	if txFee != nil {
		feeStr = txFee.String()
	}

	bt := blockTime.In(cst).Format("2006-01-02 15:04:05")
	out.DBTransactionPreview = &TransactionDBPreview{
		TxHash:          tx.Hash().Hex(),
		BlockNumber:     block.NumberU64(),
		BlockHash:       block.Hash().Hex(),
		BlockTime:       bt,
		TxIndex:         uint(receipt.TransactionIndex),
		FromAddress:     fromStr,
		ToAddress:       toStr,
		ContractAddress: normalizeAddress(tx.To().Hex()),
		FuncSelector:    lastSel,
		FuncName:        lastName,
		Value:           valStr,
		GasLimit:        tx.Gas(),
		GasUsed:         receipt.GasUsed,
		GasPrice:        gpStr,
		TxFee:           feeStr,
		Status:          int8(receipt.Status),
		TxData:          hex.EncodeToString(txData),
	}
	_ = lastDirect // documented via per-token rows

	topic0 := transferTopic0.Hex()
	for _, tr := range transfers {
		dec := big.NewInt(int64(tr.tokInfo.Decimals))
		div := new(big.Int).Exp(big.NewInt(10), dec, nil)
		formatted := new(big.Float).Quo(new(big.Float).SetInt(tr.value), new(big.Float).SetInt(div))

		ti := tr.tokInfo
		ev := &EventDBPreview{
			TxHash:          tx.Hash().Hex(),
			BlockNumber:     block.NumberU64(),
			BlockTime:       bt,
			LogIndex:        tr.logIdx,
			ContractAddress: normalizeAddress(tr.token.Hex()),
			EventName:       "Transfer",
			EventSignature:  topic0,
			FromAddress:     normalizeAddress(tr.from.Hex()),
			ToAddress:       normalizeAddress(tr.to.Hex()),
			Value:           tr.value.String(),
			Topic0:          topic0,
			Topic1:          common.BytesToHash(tr.from.Bytes()).Hex(),
			Topic2:          common.BytesToHash(tr.to.Bytes()).Hex(),
		}

		fromBal, err1 := a.getBalanceFromContract(&tr.token, &tr.from)
		toBal, err2 := a.getBalanceFromContract(&tr.token, &tr.to)
		var fromS, toS *string
		var balErr string
		if err1 == nil {
			s := fromBal.String()
			fromS = &s
		} else {
			balErr += fmt.Sprintf("from balanceOf: %v; ", err1)
		}
		if err2 == nil {
			s := toBal.String()
			toS = &s
		} else {
			balErr += fmt.Sprintf("to balanceOf: %v", err2)
		}

		out.Transfers = append(out.Transfers, TransferInspect{
			LogIndex:       tr.logIdx,
			TokenAddress:   tr.token.Hex(),
			TokenInfo:      ti,
			From:           tr.from.Hex(),
			To:             tr.to.Hex(),
			ValueRaw:       tr.value.String(),
			ValueFormatted: formatted.Text('f', int(ti.Decimals)) + " " + ti.Symbol,
			DBEventPreview: ev,
			FromBalance:    fromS,
			ToBalance:      toS,
			BalanceErr:     strings.TrimSpace(balErr),
		})
	}

	return out
}
