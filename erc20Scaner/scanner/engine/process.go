package engine

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"

	"strings"
	"time"

	chain33types "github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db/block"
	"github.com/33cn/externaldb/erc20Scaner/blockalign"
	"github.com/33cn/externaldb/erc20Scaner/config"
	"github.com/33cn/externaldb/erc20Scaner/database"
	"github.com/33cn/externaldb/erc20Scaner/erc20abi/generated"
	"github.com/33cn/externaldb/erc20Scaner/txparser"
	"github.com/33cn/externaldb/escli"
)

// normalizeAddress 规范化地址，统一转换为小写
// 以太坊地址是大小写不敏感的，统一转换为小写便于比较和查询
func normalizeAddress(address string) string {
	if address == "" {
		return address
	}
	// 保持0x前缀，将后面的字符转换为小写
	if strings.HasPrefix(address, "0x") || strings.HasPrefix(address, "0X") {
		return "0x" + strings.ToLower(address[2:])
	}
	return strings.ToLower(address)
}

// Process 业务处理模块
type Process struct {
	cli        *Client
	StartPoint uint64
	EndPoint   uint64
	EnableDB   bool
	DBDSN      string
	db         *database.DB
	NodeURL    string
	esClient   escli.ESClient // ES客户端，用于从ES读取区块

	SkipInlineBalanceUpdate bool // 为 true 时不扫块内联 balanceOf，依赖占位行 + balance_refresher

	// NoProgress 为 true 时不读/写 scan_progress 表（修复模式，方案 D）
	NoProgress bool

	// esResumeHeight：已处理完成的链上区块高度（用于 ES 模式 seq/height 对齐，仅 ES 路径使用）
	esResumeHeight uint64
}

// Init 初始化模块
func (p *Process) Init() {
	if p.cli == nil {
		p.cli = new(Client)
	}
	p.cli.ConnectEth(p.NodeURL)

	// 初始化数据库连接
	if p.EnableDB {
		var err error
		p.db, err = database.NewDB(p.DBDSN)
		if err != nil {
			log.Error("Failed to connect to database, database write will be disabled", "err", err)
			p.EnableDB = false
		} else {
			log.Info("Database connection established successfully")
			if p.NoProgress {
				log.Info("NoProgress mode: skipping scan_progress read, using configured start point",
					"startPoint", p.StartPoint)
			} else {
				// 读取上次处理的进度
				progress, err := p.db.GetScanProgress()
				if err != nil {
					log.Warn("Failed to get scan progress, will start from configured start point", "err", err, "startPoint", p.StartPoint)
				} else if progress != nil {
					// 如果存在进度记录，从上次处理的高度+1开始
					p.StartPoint = progress.LastBlockNumber + 1
					log.Info("Resuming scan from last processed block",
						"lastBlock", progress.LastBlockNumber,
						"startBlock", p.StartPoint)
				} else {
					log.Info("No previous progress found, starting from configured start point", "startPoint", p.StartPoint)
				}
			}
		}
	}
}

// esSeqHeightSmallDrift 当 |SyncSeq-height| < 该值时，高度与期望不一致也直接处理本轮（避免反复对齐）
const esSeqHeightSmallDrift = 100

// alignESScanStart 解决：进度存的是 block height，而 ES 文档 id 为 sync_seq（通常 seq>=height；回滚后 seq 会继续增加）。
// 探测一次 seq=lastH+1 的文档，取 d = SyncSeq - height，将下一条 ES 键设为 (lastH+1)+d，与链上下一块高度对齐。
func (p *Process) alignESScanStart() {
	if p.NoProgress || !p.EnableDB || p.db == nil {
		if p.StartPoint > 0 {
			p.esResumeHeight = p.StartPoint - 1
		}
		return
	}
	progress, err := p.db.GetScanProgress()
	if err != nil {
		log.Warn("ES align: get scan progress", "err", err)
		if p.StartPoint > 0 {
			p.esResumeHeight = p.StartPoint - 1
		}
		return
	}
	if progress == nil || progress.LastBlockNumber == 0 {
		if p.StartPoint > 0 {
			p.esResumeHeight = p.StartPoint - 1
		}
		return
	}
	lastH := progress.LastBlockNumber
	p.esResumeHeight = lastH

	probe, err := p.getBlockFromES(int64(lastH + 1))
	if err != nil {
		log.Warn("ES align: probe getBlockFromES failed", "err", err, "probeSeqAsHeight", lastH+1)
		return
	}
	if probe == nil {
		log.Warn("ES align: probe document missing", "probeSeqAsHeight", lastH+1)
		return
	}
	var detail chain33types.BlockDetail
	if err := chain33types.Decode(probe.BlockDetail, &detail); err != nil {
		log.Warn("ES align: decode probe BlockDetail", "err", err)
		return
	}
	H := int64(detail.Block.Height)
	S := int64(probe.SyncSeq)
	delta := S - H
	// 下一块链高 lastH+1 对应的 ES 键约为 (lastH+1) + delta
	p.StartPoint = uint64(int64(lastH+1) + delta)
	log.Info("ES scan start aligned",
		"lastBlockHeight", lastH,
		"probeSyncSeq", S,
		"probeHeight", H,
		"delta_seq_minus_height", delta,
		"nextESSeq", p.StartPoint)
}

// StartWithEsClient 从ES读取区块并处理evm交易
func (p *Process) StartWithEsClient(esClient escli.ESClient) {
	p.esClient = esClient
	p.alignESScanStart()
	for {
		// 检查是否到达结束点
		if p.EndPoint > 0 && p.StartPoint >= uint64(p.EndPoint) {
			if p.NoProgress {
				log.Info("scannerfix completed (ES mode)", "endBlock", p.EndPoint, "lastSeq", p.StartPoint)
				return
			}
			time.Sleep(time.Second)
			continue
		}

		// 从ES读取区块
		blockSeq, err := p.getBlockFromES(int64(p.StartPoint))
		if err != nil {
			log.Warn("Failed to get block from ES", "err", err, "seq", p.StartPoint)
			time.Sleep(time.Second)
			continue
		}

		if blockSeq == nil {
			// 区块不存在，等待
			time.Sleep(time.Second)
			continue
		}

		var detail chain33types.BlockDetail
		if err := chain33types.Decode(blockSeq.BlockDetail, &detail); err != nil {
			log.Error("Failed to decode BlockDetail from ES", "err", err, "seq", p.StartPoint)
			time.Sleep(time.Second)
			continue
		}
		have := uint64(detail.Block.Height)
		want := p.esResumeHeight + 1
		delta := int64(blockSeq.SyncSeq) - int64(have)
		absDelta := delta
		if absDelta < 0 {
			absDelta = -absDelta
		}
		if have != want {
			// 期望的链上下一块高度为 want，当前文档内高度为 have；用同一 d=SyncSeq-height 逼近 want 对应的 ES 键
			targetSeq := int64(want) + delta
			if targetSeq < 0 {
				log.Warn("ES seq align: negative targetSeq", "want", want, "have", have, "delta", delta)
				time.Sleep(time.Second)
				continue
			}
			if absDelta >= esSeqHeightSmallDrift {
				log.Warn("ES seq/height mismatch, realigning",
					"wantHeight", want, "haveHeight", have,
					"syncSeq", blockSeq.SyncSeq, "delta_seq_minus_height", delta,
					"nextESSeq", uint64(targetSeq))
				p.StartPoint = uint64(targetSeq)
				continue
			}
			// |delta| 较小时直接按当前块处理，一般区块量小、很快可追上
			log.Info("ES seq/height small drift, processing this seq anyway",
				"wantHeight", want, "haveHeight", have,
				"syncSeq", blockSeq.SyncSeq, "delta", delta)
		}

		err = p.parseBlockFromES(blockSeq)
		if err != nil {
			log.Error("Failed to parse block from ES", "err", err, "seq", p.StartPoint)
			time.Sleep(time.Second)
			continue
		}

		// 更新处理进度（存链上高度，非 ES 的 sync_seq）
		if !p.NoProgress && p.EnableDB && p.db != nil {
			blockTime := time.Unix(int64(detail.Block.BlockTime), 0)
			err = p.db.UpdateScanProgress(
				have,
				blockSeq.Hash,
				blockTime,
				0,
			)
			if err != nil {
				log.Error("Failed to update scan progress", "err", err, "height", have)
			}
		}
		p.esResumeHeight = have
		// 游标按 sync_seq 递增，避免把「高度」当 ES 文档 id
		p.StartPoint = uint64(int64(blockSeq.SyncSeq) + 1)
	}
}

// Start 启动模块
func (p *Process) Start() {
	for {
		blockNum, err := p.cli.BlockNum()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		if blockNum < p.StartPoint {
			time.Sleep(time.Second)
			continue
		}
		if p.EndPoint > 0 && blockNum >= p.EndPoint {
			if p.NoProgress {
				log.Info("scannerfix completed (node mode)", "endBlock", p.EndPoint, "lastNum", p.StartPoint)
				return
			}
			time.Sleep(time.Second)
			continue
		}

		block, err := p.cli.BlockByNumber(p.StartPoint)
		if err != nil {
			log.Error("Failed to get block by number", "err", err, "block", p.StartPoint)
			continue
		}
		err = p.ParaseBlock(block)
		if err != nil {
			log.Error("Failed to parse block", "err", err, "block", p.StartPoint)
			continue
		}

		// 更新处理进度
		if !p.NoProgress && p.EnableDB && p.db != nil {
			blockTime := time.Unix(int64(block.Time()), 0)
			err = p.db.UpdateScanProgress(
				block.NumberU64(),
				block.Hash().Hex(),
				blockTime,
				0, // processed_tx_count 可以根据需要统计
			)
			if err != nil {
				log.Error("Failed to update scan progress", "err", err, "block", p.StartPoint)
			}
		}

		p.StartPoint++
	}

}

func (p *Process) Close() error {
	if p.cli != nil {
		p.cli.CloseConnect()
	}
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// BlockByNumber delegates to Client.BlockByNumber for testing/inspection.
func (p *Process) BlockByNumber(number uint64) (*types.Block, error) {
	return p.cli.BlockByNumber(number)
}

// getBlockFromES 从ES获取区块
func (p *Process) getBlockFromES(seqNum int64) (*block.Seq, error) {
	if p.esClient == nil {
		return nil, fmt.Errorf("esClient is nil")
	}

	id := fmt.Sprintf("%d", seqNum)
	result, err := p.esClient.Get(block.StatusDB, block.StatusDB, id)
	if err != nil {
		return nil, err
	}

	var seq block.Seq
	err = json.Unmarshal([]byte(*result), &seq)
	if err != nil {
		return nil, err
	}

	return &seq, nil
}

// isEvmExecer 检查执行器是否是evm
func isEvmExecer(execer string) bool {
	return execer == "evm" || strings.HasSuffix(execer, ".evm")
}

// parseBlockFromES 解析从ES获取的区块
func (p *Process) parseBlockFromES(blockSeq *block.Seq) error {
	if blockSeq == nil {
		return fmt.Errorf("blockSeq is nil")
	}

	// 解码BlockDetail
	var detail chain33types.BlockDetail
	err := chain33types.Decode(blockSeq.BlockDetail, &detail)
	if err != nil {
		return fmt.Errorf("decode BlockDetail failed: %w", err)
	}

	log.Info("Processing block from ES",
		"seq", blockSeq.SyncSeq,
		"type", blockSeq.Type,
		"height", detail.Block.Height,
		"txCount", len(detail.Block.Txs))

	// 遍历交易，查找evm交易
	evmtxs := make([]int, 0)
	for i, tx := range detail.Block.Txs {
		// 检查执行器是否是evm
		if !isEvmExecer(string(tx.Execer)) {
			continue
		}
		evmtxs = append(evmtxs, i)
	}

	if len(evmtxs) == 0 {
		return nil
	}

	//block := detail.Block
	//var block *types.Block
	block, err := p.cli.BlockByNumber(uint64(detail.Block.Height))
	if err != nil {
		log.Error("Failed to get block by number", "err", err, "height", detail.Block.Height)
		return err
	}
	seqN := len(detail.Block.Txs)
	ethN := block.Transactions().Len()

	var aligned []*types.Transaction
	if ethN < seqN {
		expects, errExpect := buildSlotEthExpectsFromBlockDetail(&detail)
		if errExpect != nil {
			log.Error("Failed to build chain33 slot expects for nonce alignment", "err", errExpect, "height", detail.Block.Height)
			return fmt.Errorf("parseBlockFromES height=%d: eth tx count %d < chain33 %d, nonce alignment: %w",
				detail.Block.Height, ethN, seqN, errExpect)
		}
		aligned, err = blockalign.AlignEthTxsByNonce(block, expects)
		if err != nil {
			log.Error("Failed to align eth txs by nonce", "err", err, "height", detail.Block.Height)
			return fmt.Errorf("parseBlockFromES height=%d: AlignEthTxsByNonce: %w", detail.Block.Height, err)
		}
		log.Info("Aligned eth txs by nonce (eth body count < chain33 seq)",
			"height", detail.Block.Height,
			"chain33TxCount", seqN,
			"ethBlockTxCount", ethN,
		)
	} else if ethN > seqN {
		aligned, err = blockalign.AlignEthTxsWithSeqCount(block, seqN, p.cli)
		if err != nil {
			log.Error("Failed to align eth txs with chain33 block", "err", err, "height", detail.Block.Height)
			return nil
		}
	} else {
		// ethN == seqN: 两边交易数一致，直接使用 block 的交易列表
		aligned = make([]*types.Transaction, ethN)
		for i := 0; i < ethN; i++ {
			aligned[i] = block.Transactions()[i]
		}
	}
	if block.Transactions().Len() != len(detail.Block.Txs) {
		log.Warn("chain33 seq tx count differs from eth_getBlock tx count",
			"seq-seq", blockSeq.SyncSeq,
			"seq-type", blockSeq.Type,
			"seq-hash", blockSeq.Hash,
			"blockHash", block.Hash().Hex(),
			"seq-height", detail.Block.Height,
			"seq-txCount", len(detail.Block.Txs),
			"ethBlockTxCount", block.Transactions().Len(),
		)
	}

	log.Debug("Processing block transactions",
		"startPoint", p.StartPoint,
		"height", detail.Block.Height,
		"txCount", len(aligned))
	for _, idx := range evmtxs {
		if idx < 0 || idx >= len(aligned) {
			log.Warn("parseBlockFromES: EVM tx index out of range, skip",
				"height", detail.Block.Height,
				"evmTxIndex", idx,
				"alignedLen", len(aligned))
			continue
		}
		if aligned[idx] == nil {
			var c33TxHash string
			if idx < len(detail.Block.Txs) {
				c33TxHash = hexutil.Encode(detail.Block.Txs[idx].Hash())
			}
			log.Warn("parseBlockFromES: missing ethereum tx body for EVM slot, skip",
				"height", detail.Block.Height,
				"evmTxIndex", idx,
				"chain33TxHash", c33TxHash)
			continue
		}
		err := p.processTransactionWithReceipt(aligned[idx], block)
		if err != nil {
			// 处理失败只打日志，不返回错误，避免上层对同一块反复重试形成死循环
			log.Error("Failed to process transaction",
				"err", err,
				"txHash", aligned[idx].Hash().Hex(),
				"block", block.NumberU64())
		}
	}

	return nil
}

// handleContractCreation 处理合约创建
func (p *Process) handleContractCreation(tx *types.Transaction, receipt *types.Receipt, block *types.Block) error {
	if receipt == nil || receipt.ContractAddress == (common.Address{}) {
		return fmt.Errorf("invalid receipt or contract address")
	}

	// 检查是否为ERC20合约
	isERC20, err := p.checkERC20BySelector(&receipt.ContractAddress)
	if err != nil {
		log.Warn("Failed to check ERC20 selector, saving as UNKNOWN contract",
			"err", err,
			"contract", receipt.ContractAddress.Hex())
		// 检查失败，保存为UNKNOWN类型
		return p.saveNonERC20ContractToDB(receipt, block, tx, "UNKNOWN")
	}

	if !isERC20 {
		// 不是ERC20合约，保存为UNKNOWN类型
		log.Info("Contract is not ERC20, saving as UNKNOWN",
			"contract", receipt.ContractAddress.Hex())
		return p.saveNonERC20ContractToDB(receipt, block, tx, "UNKNOWN")
	}

	// 是ERC20合约，获取合约信息
	decimals, err := p.unPackageAbi("decimals", &receipt.ContractAddress)
	if err != nil {
		return err
	}
	cname, err := p.unPackageAbi("name", &receipt.ContractAddress)
	if err != nil {
		return err
	}
	symbol, err := p.unPackageAbi("symbol", &receipt.ContractAddress)
	if err != nil {
		return err
	}
	supply, err := p.unPackageAbi("totalSupply", &receipt.ContractAddress)
	if err != nil {
		return err
	}

	ut := time.Unix(int64(block.Time()), 0)
	cst, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return err
	}

	// 使用 Debug 级别输出详细的合约信息
	log.Debug("Contract creation details",
		"address", receipt.ContractAddress.Hex(),
		"name", fmt.Sprintf("%v", cname),
		"symbol", fmt.Sprintf("%v", symbol),
		"totalSupply", fmt.Sprintf("%v", supply),
		"decimals", fmt.Sprintf("%v", decimals),
		"deployTime", ut.In(cst).Format("2006-01-02 15:04:05"),
		"block", block.NumberU64())

	// 写入数据库
	if p.EnableDB {
		err = p.saveContractToDB(receipt, block, tx, cname, symbol, decimals, supply)
		if err != nil {
			log.Error("Failed to save contract to database",
				"err", err,
				"contract", receipt.ContractAddress.Hex())
		} else {
			log.Info("Contract saved to database",
				"contract", receipt.ContractAddress.Hex(),
				"name", fmt.Sprintf("%v", cname),
				"symbol", fmt.Sprintf("%v", symbol))
		}
	}

	return nil
}

// ParaseBlock 解析区块
func (p *Process) ParaseBlock(block *types.Block) error {

	if block == nil {
		time.Sleep(time.Second)
		return fmt.Errorf("nil block")
	}

	txs := block.Transactions()

	log.Debug("Processing block transactions",
		"startPoint", p.StartPoint,
		"height", block.NumberU64(),
		"txCount", len(txs))
	for _, tx := range txs {
		if tx == nil {
			log.Warn("Skipping nil transaction slot (node omitted tx body)",
				"block", block.NumberU64())
			continue
		}
		err := p.processTransactionWithReceipt(tx, block)
		if err != nil {
			// 处理失败不影响其他交易的处理
			log.Error("Failed to process transaction",
				"err", err,
				"txHash", tx.Hash().Hex(),
				"block", block.NumberU64())
		}
	}
	return nil
}

// processTransactionWithReceipt 获取receipt并处理交易
func (p *Process) processTransactionWithReceipt(tx *types.Transaction, block *types.Block) error {
	// 获取交易receipt
	receipt, err := p.cli.TxReceipt(tx.Hash())
	if err != nil {
		return fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	if receipt == nil {
		return fmt.Errorf("receipt is nil for tx: %s", tx.Hash().Hex())
	}

	// 检查交易状态
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("transaction failed, status: %d, tx: %s", receipt.Status, tx.Hash().Hex())
	}

	// 根据交易类型处理
	if tx.To() == nil {
		// 合约创建交易
		return p.processContractCreation(tx, receipt, block)
	} else {
		// 合约调用交易（token transfer or contract call）
		return p.parseERC20Transfer(tx, receipt, block)
	}
}

// processContractCreation 处理合约创建交易
func (p *Process) processContractCreation(tx *types.Transaction, receipt *types.Receipt, block *types.Block) error {
	if receipt.ContractAddress == (common.Address{}) {
		return fmt.Errorf("contract address is empty")
	}

	// 使用统一的处理函数，handleContractCreation 内部会判断是否为ERC20并保存
	return p.handleContractCreation(tx, receipt, block)
}

func (p *Process) unPackageAbi(methodName string, cAddress *common.Address) (interface{}, error) {
	parsedAbi, err := abi.JSON(strings.NewReader(generated.ERC20ABI))
	if err != nil {
		panic(err)
	}
	abidata, err := parsedAbi.Pack(methodName)
	if err != nil {
		return nil, err
	}
	//通过节点查询返回evm查询结果
	var calldata ethereum.CallMsg
	calldata.Data = abidata
	calldata.To = cAddress
	callResult, err := p.cli.CallContract(calldata, nil)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = parsedAbi.UnpackIntoInterface(&result, methodName, callResult)
	if err != nil {
		return nil, err
	}

	return result, nil

}

// unPackageAbiWithArgs 带参数的ABI调用
func (p *Process) unPackageAbiWithArgs(methodName string, cAddress *common.Address, args ...interface{}) (interface{}, error) {
	parsedAbi, err := abi.JSON(strings.NewReader(generated.ERC20ABI))
	if err != nil {
		panic(err)
	}
	abidata, err := parsedAbi.Pack(methodName, args...)
	if err != nil {
		return nil, err
	}
	//通过节点查询返回evm查询结果
	var calldata ethereum.CallMsg
	calldata.Data = abidata
	calldata.To = cAddress
	callResult, err := p.cli.CallContract(calldata, nil)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = parsedAbi.UnpackIntoInterface(&result, methodName, callResult)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// checkERC20BySelector 通过检查函数选择器验证合约是否为ERC20合约
// 该方法检查合约字节码中是否包含ERC20标准函数的函数选择器
func (p *Process) checkERC20BySelector(contractAddress *common.Address) (bool, error) {
	// 获取合约字节码
	code, err := p.cli.CodeAt(*contractAddress, nil)
	if err != nil {
		return false, fmt.Errorf("failed to get contract code: %w", err)
	}

	// 如果合约代码为空，不是合约地址
	if len(code) == 0 {
		return false, nil
	}

	// 将字节码转换为十六进制字符串以便搜索
	codeHex := hex.EncodeToString(code)

	// ERC20标准必须实现的6个核心函数选择器（ERC20标准要求）
	// 这些是ERC20接口必须实现的函数，从ERC20FuncSigs中获取
	requiredCoreSelectors := []string{
		"18160ddd", // totalSupply() - 总供应量
		"70a08231", // balanceOf(address) - 余额查询
		"a9059cbb", // transfer(address,uint256) - 转账
		"23b872dd", // transferFrom(address,address,uint256) - 授权转账
		"095ea7b3", // approve(address,uint256) - 授权
		"dd62ed3e", // allowance(address,address) - 查询授权额度
	}

	// ERC20可选但常见的函数选择器（大多数ERC20代币都实现）
	// 这些函数虽然不是ERC20标准强制要求，但绝大多数代币都会实现
	optionalSelectors := []string{
		"06fdde03", // name() - 代币名称
		"95d89b41", // symbol() - 代币符号
		"313ce567", // decimals() - 小数位数
	}

	// ERC20扩展函数选择器（OpenZeppelin等库提供的安全函数）
	// 这些是增强安全性的函数，不是所有ERC20代币都实现
	extensionSelectors := []string{
		"39509351", // increaseAllowance(address,uint256) - 增加授权
		"a457c2d7", // decreaseAllowance(address,uint256) - 减少授权
	}
	// 注意：以上所有选择器都对应generated.ERC20FuncSigs中定义的函数

	// 检查核心函数选择器（必须实现）
	coreFoundCount := 0
	for _, selector := range requiredCoreSelectors {
		if strings.Contains(codeHex, selector) {
			coreFoundCount++
		}
	}

	// 检查可选函数选择器
	optionalFoundCount := 0
	for _, selector := range optionalSelectors {
		if strings.Contains(codeHex, selector) {
			optionalFoundCount++
		}
	}

	// 检查扩展函数选择器
	extensionFoundCount := 0
	for _, selector := range extensionSelectors {
		if strings.Contains(codeHex, selector) {
			extensionFoundCount++
		}
	}

	// ERC20标准要求至少实现6个核心函数
	// 为了更严格的验证，我们要求至少找到4个核心函数选择器
	// 这样可以确保合约基本符合ERC20标准
	if coreFoundCount >= 4 {
		// 进一步验证：尝试调用totalSupply和balanceOf函数，确保合约确实实现了ERC20接口
		// 这两个是最核心的view函数，必须能够成功调用
		_, err := p.unPackageAbi("totalSupply", contractAddress)
		if err != nil {
			return false, fmt.Errorf("totalSupply call failed: %w", err)
		}

		// 尝试调用balanceOf（使用零地址作为测试）
		zeroAddr := common.HexToAddress("0x0000000000000000000000000000000000000000")
		_, err = p.unPackageAbiWithArgs("balanceOf", contractAddress, zeroAddr)
		if err != nil {
			// balanceOf调用失败不影响验证，因为可能只是参数问题
			// 但至少totalSupply应该能成功
		}

		// 输出验证信息（Debug级别）
		log.Debug("ERC20 validation result",
			"core", fmt.Sprintf("%d/%d", coreFoundCount, len(requiredCoreSelectors)),
			"optional", fmt.Sprintf("%d/%d", optionalFoundCount, len(optionalSelectors)),
			"extension", fmt.Sprintf("%d/%d", extensionFoundCount, len(extensionSelectors)),
			"contract", contractAddress.Hex())

		return true, nil
	}

	return false, fmt.Errorf("insufficient ERC20 core functions found: %d/%d required", coreFoundCount, len(requiredCoreSelectors))
}

// parseERC20Transfer 解析ERC20 token transfer交易
func (p *Process) parseERC20Transfer(tx *types.Transaction, receipt *types.Receipt, block *types.Block) error {
	// 检查交易输入数据，判断是否是ERC20调用
	txData := tx.Data()
	if len(txData) < 4 {
		// 不是合约调用，可能是普通ETH转账
		return nil
	}

	// 获取函数选择器（前4个字节）
	funcSelector := hex.EncodeToString(txData[:4])

	// ERC20 transfer相关函数选择器
	transferSelector := "a9059cbb"     // transfer(address,uint256)
	transferFromSelector := "23b872dd" // transferFrom(address,address,uint256)

	// 检查是否是ERC20 transfer调用（用于日志输出）
	isDirectTransfer := funcSelector == transferSelector || funcSelector == transferFromSelector
	_ = isDirectTransfer // 可用于后续扩展，比如区分直接调用和间接触发

	// 使用共享解析器解析 Transfer 和 Approval 事件
	rawTransfers, rawApprovals, err := txparser.ParseReceiptLogs(receipt.Logs)
	if err != nil {
		return fmt.Errorf("failed to parse receipt logs: %w", err)
	}

	type TransferWithLogIndex struct {
		TransferInfo
		LogIndex uint
	}
	var transfers []TransferWithLogIndex
	for _, rt := range rawTransfers {
		tokenInfo, err := p.getTokenInfo(&rt.TokenAddress)
		if err != nil {
			tokenInfo = &TokenInfo{
				Address:  rt.TokenAddress.Hex(),
				Symbol:   "UNKNOWN",
				Decimals: 18,
				Name:     "Unknown Token",
			}
		}
		transfers = append(transfers, TransferWithLogIndex{
			TransferInfo: TransferInfo{
				TokenAddress: rt.TokenAddress,
				From:         rt.From,
				To:           rt.To,
				Value:        rt.Value,
				TokenInfo:    *tokenInfo,
			},
			LogIndex: rt.LogIndex,
		})
	}
	approvals := rawApprovals

	// 如果没有找到Transfer事件，可能不是ERC20转账
	if len(transfers) == 0 {
		return nil
	}

	// 输出详细的transfer信息
	ut := time.Unix(int64(block.Time()), 0)
	cst, _ := time.LoadLocation("Asia/Shanghai")

	// 尝试从交易签名中恢复发送者地址
	var fromAddr common.Address
	if tx.ChainId() != nil {
		signer := types.NewEIP155Signer(tx.ChainId())
		if sender, err := types.Sender(signer, tx); err == nil {
			fromAddr = sender
		}
	}

	// 使用 Debug 级别输出详细的交易信息
	log.Debug("Processing ERC20 transfer transaction",
		"txHash", tx.Hash().Hex(),
		"block", block.NumberU64(),
		"blockTime", ut.In(cst).Format("2006-01-02 15:04:05"),
		"from", func() string {
			if fromAddr != (common.Address{}) {
				return fromAddr.Hex()
			}
			return ""
		}(),
		"to", func() string {
			if tx.To() != nil {
				return tx.To().Hex()
			}
			return ""
		}(),
		"gasUsed", receipt.GasUsed,
		"transferCount", len(transfers))

	// 输出每个Transfer事件的详细信息
	for i, transfer := range transfers {
		// 计算实际金额（考虑decimals）
		decimals := big.NewInt(int64(transfer.TokenInfo.Decimals))
		divisor := new(big.Int).Exp(big.NewInt(10), decimals, nil)
		actualValue := new(big.Float).Quo(new(big.Float).SetInt(transfer.Value), new(big.Float).SetInt(divisor))

		log.Debug("Transfer event",
			"index", i+1,
			"tokenAddress", transfer.TokenAddress.Hex(),
			"tokenName", transfer.TokenInfo.Name,
			"tokenSymbol", transfer.TokenInfo.Symbol,
			"tokenDecimals", transfer.TokenInfo.Decimals,
			"from", transfer.From.Hex(),
			"to", transfer.To.Hex(),
			"valueRaw", transfer.Value.String(),
			"valueFormatted", actualValue.Text('f', int(transfer.TokenInfo.Decimals))+" "+transfer.TokenInfo.Symbol)
	}

	// 写入数据库
	if p.EnableDB {
		// 收集所有不同的ERC20合约地址
		contractAddresses := make(map[common.Address]bool)
		for _, transfer := range transfers {
			contractAddresses[transfer.TokenAddress] = true
		}

		var funcName string
		var finalFuncSelector string
		var isDirectCall bool

		// 为每个不同的ERC20合约分别处理
		for contractAddress := range contractAddresses {
			// 检查合约是否已经记录到数据库, 如果没有需要通过合约地址查询合约信息并检查是否为erc20合约
			// 如果是erc20合约, 需要在数据库中记录合约信息
			// 规范化地址为小写，确保大小写不敏感查询
			_, err := p.db.GetContractByAddress(normalizeAddress(contractAddress.Hex()))
			if err != nil {
				// 合约不在数据库中，需要查询并保存
				log.Info("Contract not found in database, querying contract info", "contract", contractAddress.Hex())

				// 验证是否为ERC20合约
				isERC20, err := p.checkERC20BySelector(&contractAddress)
				if err != nil {
					log.Warn("Failed to verify ERC20 contract, saving as UNKNOWN",
						"err", err,
						"contract", contractAddress.Hex())
					// 验证失败，保存为UNKNOWN类型
					err = p.saveNonERC20ContractInfoFromAddress(&contractAddress, block, "UNKNOWN")
					if err != nil {
						log.Error("Failed to save non-ERC20 contract info",
							"err", err,
							"contract", contractAddress.Hex())
					} else {
						log.Info("Non-ERC20 contract saved to database", "contract", contractAddress.Hex())
					}
					// 继续处理transfer事件
					continue
				} else if !isERC20 {
					log.Info("Contract is not ERC20, but has Transfer event, saving as UNKNOWN",
						"contract", contractAddress.Hex())
					// 虽然不是ERC20，但既然有Transfer事件，保存为UNKNOWN类型
					err = p.saveNonERC20ContractInfoFromAddress(&contractAddress, block, "UNKNOWN")
					if err != nil {
						log.Error("Failed to save non-ERC20 contract info",
							"err", err,
							"contract", contractAddress.Hex())
					} else {
						log.Info("Non-ERC20 contract saved to database", "contract", contractAddress.Hex())
					}
					// 继续处理transfer事件
					continue
				} else {
					// 是ERC20合约，获取合约信息并保存
					log.Info("Contract is ERC20, fetching contract info", "contract", contractAddress.Hex())
					err = p.saveContractInfoFromAddress(&contractAddress, block)
					if err != nil {
						log.Error("Failed to save contract info",
							"err", err,
							"contract", contractAddress.Hex())
						// 即使保存失败，也继续处理transfer事件
					} else {
						log.Info("Contract info saved to database", "contract", contractAddress.Hex())
					}
				}
			} else {
				log.Debug("Contract already exists in database", "contract", contractAddress.Hex())
			}

			// 判断是直接调用还是嵌套调用
			// 如果交易的to地址等于ERC20合约地址，且函数选择器是transfer/transferFrom，则是直接调用
			// 否则是嵌套调用（其他合约调用了ERC20合约）
			if tx.To() != nil && *tx.To() == contractAddress {
				if funcSelector == transferSelector || funcSelector == transferFromSelector {
					isDirectCall = true
				}
			}

			// 确定函数名称和选择器

			if isDirectCall {
				// 直接调用，使用交易的函数选择器
				funcName = "transfer"
				if funcSelector == transferFromSelector {
					funcName = "transferFrom"
				}
				finalFuncSelector = funcSelector
			} else {
				// 嵌套调用，标记为通过其他合约调用
				funcName = "transfer"             // 嵌套调用时，可能是transfer或transferFrom，统一标记为transfer
				finalFuncSelector = "nested_call" // 使用特殊标记表示嵌套调用
			}

		}

		// 保存交易信息（为每个ERC20合约分别保存）
		err = p.saveTransactionToDB(tx, receipt, block, *tx.To(), finalFuncSelector, funcName)
		if err != nil {
			log.Error("Failed to save transaction to database",
				"err", err,
				"txHash", tx.Hash().Hex(),
				"contract", (*tx.To()).Hex(),
				"block", block.NumberU64())
		} else {
			log.Debug("Transaction saved to database",
				"txHash", tx.Hash().Hex(),
				"contract", (*tx.To()).Hex(),
				"block", block.NumberU64(),
				"funcName", funcName,
				"isDirectCall", isDirectCall)
		}

		// 保存每个Transfer事件并更新余额
		for _, transfer := range transfers {
			// 保存事件（使用实际的log index）
			err := p.saveEventToDB(&transfer.TransferInfo, block, tx.Hash(), transfer.LogIndex)
			if err != nil {
				log.Error("Failed to save event to database",
					"err", err,
					"txHash", tx.Hash().Hex(),
					"logIndex", transfer.LogIndex,
					"block", block.NumberU64())
			} else {
				log.Debug("Event saved to database",
					"txHash", tx.Hash().Hex(),
					"logIndex", transfer.LogIndex,
					"from", transfer.From.Hex(),
					"to", transfer.To.Hex(),
					"value", transfer.Value.String(),
					"token", transfer.TokenAddress.Hex())
			}

			if p.SkipInlineBalanceUpdate {
				if err := p.touchBalanceRowInDB(transfer.From, transfer.TokenAddress, tx.Hash(), block.NumberU64()); err != nil {
					log.Error("Failed to touch from balance row",
						"err", err,
						"address", transfer.From.Hex(),
						"token", transfer.TokenAddress.Hex())
				}
				if err := p.touchBalanceRowInDB(transfer.To, transfer.TokenAddress, tx.Hash(), block.NumberU64()); err != nil {
					log.Error("Failed to touch to balance row",
						"err", err,
						"address", transfer.To.Hex(),
						"token", transfer.TokenAddress.Hex())
				}
			} else {
				// 更新发送者余额（减少）
				fromBalance, err := p.getBalanceFromContract(&transfer.TokenAddress, &transfer.From)
				if err == nil {
					err = p.updateBalanceInDB(transfer.From, transfer.TokenAddress, fromBalance, tx.Hash(), block.NumberU64())
					if err != nil {
						log.Error("Failed to update from balance",
							"err", err,
							"address", transfer.From.Hex(),
							"token", transfer.TokenAddress.Hex())
					}
				}

				// 更新接收者余额（增加）
				toBalance, err := p.getBalanceFromContract(&transfer.TokenAddress, &transfer.To)
				if err == nil {
					err = p.updateBalanceInDB(transfer.To, transfer.TokenAddress, toBalance, tx.Hash(), block.NumberU64())
					if err != nil {
						log.Error("Failed to update to balance",
							"err", err,
							"address", transfer.To.Hex(),
							"token", transfer.TokenAddress.Hex())
					}
				}
			}
		}
	}

	// 处理 Approval 事件：保存到 events 表和 token_allowances 表
	if p.EnableDB {
		for _, approval := range approvals {
			// 保存 Approval 事件
			err := p.saveApprovalEventToDB(&approval, block, tx.Hash())
			if err != nil {
				log.Error("Failed to save approval event",
					"err", err,
					"txHash", tx.Hash().Hex(),
					"owner", approval.Owner.Hex(),
					"spender", approval.Spender.Hex(),
					"token", approval.TokenAddress.Hex())
			}

			// 更新 token_allowances
			err = p.updateAllowanceInDB(
				approval.Owner,
				approval.Spender,
				approval.TokenAddress,
				approval.Value,
				tx.Hash(),
				block.NumberU64(),
			)
			if err != nil {
				log.Error("Failed to update allowance",
					"err", err,
					"owner", approval.Owner.Hex(),
					"spender", approval.Spender.Hex(),
					"token", approval.TokenAddress.Hex())
			} else {
				log.Debug("Allowance updated",
					"owner", approval.Owner.Hex(),
					"spender", approval.Spender.Hex(),
					"token", approval.TokenAddress.Hex(),
					"amount", approval.Value.String())
			}
		}
	}

	return nil
}

// saveApprovalEventToDB 保存 Approval 事件到 events 表
func (p *Process) saveApprovalEventToDB(approval *txparser.ParsedApproval, block *types.Block, txHash common.Hash) error {
	if p.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Approval事件的topic0签名
	approvalEventSig := common.HexToHash("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925")

	dbEvent := &database.Event{
		TxHash:          txHash.Hex(),
		BlockNumber:     block.NumberU64(),
		BlockTime:       time.Unix(int64(block.Time()), 0),
		LogIndex:        approval.LogIndex,
		ContractAddress: normalizeAddress(approval.TokenAddress.Hex()),
		EventName:       "Approval",
		EventSignature:  approvalEventSig.Hex(),
		OwnerAddress:    normalizeAddress(approval.Owner.Hex()),
		SpenderAddress:  normalizeAddress(approval.Spender.Hex()),
		Amount:          approval.Value,
		Topic0:          approvalEventSig.Hex(),
		Topic1:          common.BytesToHash(approval.Owner.Bytes()).Hex(),
		Topic2:          common.BytesToHash(approval.Spender.Bytes()).Hex(),
	}

	return p.db.SaveEvent(dbEvent)
}

// updateAllowanceInDB 更新授权额度到 token_allowances 表
// amount=0 时删除记录（授权已撤销）
func (p *Process) updateAllowanceInDB(owner, spender, contractAddress common.Address, amount *big.Int, txHash common.Hash, blockNumber uint64) error {
	if p.db == nil {
		return fmt.Errorf("database not initialized")
	}

	zero := big.NewInt(0)
	if amount.Cmp(zero) == 0 {
		// amount=0 表示取消授权，删除记录
		return p.db.DeleteAllowance(
			normalizeAddress(owner.Hex()),
			normalizeAddress(spender.Hex()),
			normalizeAddress(contractAddress.Hex()),
		)
	}

	return p.db.UpsertAllowance(
		normalizeAddress(owner.Hex()),
		normalizeAddress(spender.Hex()),
		normalizeAddress(contractAddress.Hex()),
		amount,
		txHash.Hex(),
		blockNumber,
	)
}

// getBalanceFromContract 从合约获取地址余额
func (p *Process) getBalanceFromContract(contractAddress, address *common.Address) (*big.Int, error) {
	result, err := p.unPackageAbiWithArgs("balanceOf", contractAddress, *address)
	if err != nil {
		return nil, err
	}

	var balance *big.Int
	switch v := result.(type) {
	case *big.Int:
		balance = v
	case []interface{}:
		if len(v) > 0 {
			if b, ok := v[0].(*big.Int); ok {
				balance = b
			}
		}
	}

	if balance == nil {
		return big.NewInt(0), nil
	}
	return balance, nil
}

// TransferInfo 转账信息
type TransferInfo struct {
	TokenAddress common.Address
	From         common.Address
	To           common.Address
	Value        *big.Int
	TokenInfo    TokenInfo
}

// TokenInfo 代币信息
type TokenInfo struct {
	Address  string
	Name     string
	Symbol   string
	Decimals uint8
}

// getTokenInfo 获取代币信息
func (p *Process) getTokenInfo(tokenAddress *common.Address) (*TokenInfo, error) {
	info := &TokenInfo{
		Address: tokenAddress.Hex(),
	}

	// 获取代币名称
	name, err := p.unPackageAbi("name", tokenAddress)
	if err == nil {
		info.Name = fmt.Sprintf("%v", name)
	}

	// 获取代币符号
	symbol, err := p.unPackageAbi("symbol", tokenAddress)
	if err == nil {
		info.Symbol = fmt.Sprintf("%v", symbol)
	}

	// 获取小数位数
	decimals, err := p.unPackageAbi("decimals", tokenAddress)
	if err == nil {
		// decimals通常返回uint8
		switch v := decimals.(type) {
		case uint8:
			info.Decimals = v
		case uint64:
			info.Decimals = uint8(v)
		case *big.Int:
			info.Decimals = uint8(v.Uint64())
		default:
			info.Decimals = 18 // 默认值
		}
	} else {
		info.Decimals = 18 // 默认值
	}

	return info, nil
}

// saveContractToDB 保存合约信息到数据库
func (p *Process) saveContractToDB(receipt *types.Receipt, block *types.Block, tx *types.Transaction, name, symbol interface{}, decimals, supply interface{}) error {
	if p.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// 转换数据类型
	var decimalsVal uint8 = 18
	if decimals != nil {
		switch v := decimals.(type) {
		case uint8:
			decimalsVal = v
		case uint64:
			decimalsVal = uint8(v)
		case *big.Int:
			decimalsVal = uint8(v.Uint64())
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

	// 获取已验证的函数选择器列表
	verifiedFuncs := []string{"18160ddd", "70a08231", "a9059cbb", "23b872dd", "095ea7b3", "dd62ed3e"}
	verifiedFuncsJSON, _ := json.Marshal(verifiedFuncs)

	// 尝试获取发送者地址
	var deployerAddr string
	if tx.ChainId() != nil {
		signer := types.NewEIP155Signer(tx.ChainId())
		if sender, err := types.Sender(signer, tx); err == nil {
			deployerAddr = sender.Hex()
		}
	}

	contract := &database.Contract{
		ContractAddress:    normalizeAddress(receipt.ContractAddress.Hex()),
		ContractName:       fmt.Sprintf("%v", name),
		ContractSymbol:     fmt.Sprintf("%v", symbol),
		ContractType:       "ERC20",
		Decimals:           decimalsVal,
		TotalSupply:        totalSupply,
		DeployTxHash:       tx.Hash().Hex(),
		DeployBlockNumber:  block.NumberU64(),
		DeployBlockTime:    time.Unix(int64(block.Time()), 0),
		DeployerAddress:    normalizeAddress(deployerAddr),
		VerificationStatus: 1,
		VerifiedFunctions:  string(verifiedFuncsJSON),
	}

	return p.db.SaveContract(contract)
}

// saveContractInfoFromAddress 从合约地址获取合约信息并保存到数据库
// 用于处理已存在的合约（非部署交易中发现）
func (p *Process) saveContractInfoFromAddress(contractAddress *common.Address, block *types.Block) error {
	if p.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// 获取合约信息
	cname, err := p.unPackageAbi("name", contractAddress)
	if err != nil {
		log.Warn("Failed to get contract name, using default",
			"err", err,
			"contract", contractAddress.Hex())
		cname = "Unknown"
	}

	symbol, err := p.unPackageAbi("symbol", contractAddress)
	if err != nil {
		log.Warn("Failed to get contract symbol, using default",
			"err", err,
			"contract", contractAddress.Hex())
		symbol = "UNKNOWN"
	}

	decimals, err := p.unPackageAbi("decimals", contractAddress)
	if err != nil {
		log.Warn("Failed to get contract decimals, using default",
			"err", err,
			"contract", contractAddress.Hex())
		decimals = uint8(18) // 默认值
	}

	supply, err := p.unPackageAbi("totalSupply", contractAddress)
	if err != nil {
		log.Warn("Failed to get contract totalSupply, using default",
			"err", err,
			"contract", contractAddress.Hex())
		supply = big.NewInt(0)
	}

	// 转换数据类型
	var decimalsVal uint8 = 18
	if decimals != nil {
		switch v := decimals.(type) {
		case uint8:
			decimalsVal = v
		case uint64:
			decimalsVal = uint8(v)
		case *big.Int:
			decimalsVal = uint8(v.Uint64())
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
	} else {
		totalSupply = big.NewInt(0)
	}

	// 获取已验证的函数选择器列表
	verifiedFuncs := []string{"18160ddd", "70a08231", "a9059cbb", "23b872dd", "095ea7b3", "dd62ed3e"}
	verifiedFuncsJSON, _ := json.Marshal(verifiedFuncs)

	// 对于已存在的合约，我们没有部署交易信息，使用默认值
	contract := &database.Contract{
		ContractAddress:    normalizeAddress(contractAddress.Hex()),
		ContractName:       fmt.Sprintf("%v", cname),
		ContractSymbol:     fmt.Sprintf("%v", symbol),
		ContractType:       "ERC20",
		Decimals:           decimalsVal,
		TotalSupply:        totalSupply,
		DeployTxHash:       "",                                // 未知，因为不是从部署交易中发现的
		DeployBlockNumber:  0,                                 // 未知
		DeployBlockTime:    time.Unix(int64(block.Time()), 0), // 使用当前区块时间作为参考
		DeployerAddress:    "",                                // 未知
		VerificationStatus: 1,
		VerifiedFunctions:  string(verifiedFuncsJSON),
	}

	return p.db.SaveContract(contract)
}

// saveNonERC20ContractToDB 保存非ERC20合约信息到数据库（从部署交易）
func (p *Process) saveNonERC20ContractToDB(receipt *types.Receipt, block *types.Block, tx *types.Transaction, contractType string) error {
	if p.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// 尝试获取发送者地址
	var deployerAddr string
	if tx.ChainId() != nil {
		signer := types.NewEIP155Signer(tx.ChainId())
		if sender, err := types.Sender(signer, tx); err == nil {
			deployerAddr = sender.Hex()
		}
	}

	contract := &database.Contract{
		ContractAddress:    normalizeAddress(receipt.ContractAddress.Hex()),
		ContractName:       "Unknown",
		ContractSymbol:     "UNKNOWN",
		ContractType:       contractType,
		Decimals:           0, // 非ERC20合约没有decimals
		TotalSupply:        big.NewInt(0),
		DeployTxHash:       tx.Hash().Hex(),
		DeployBlockNumber:  block.NumberU64(),
		DeployBlockTime:    time.Unix(int64(block.Time()), 0),
		DeployerAddress:    normalizeAddress(deployerAddr),
		VerificationStatus: 0,    // 未验证
		VerifiedFunctions:  "[]", // 空数组
	}

	return p.db.SaveContract(contract)
}

// saveNonERC20ContractInfoFromAddress 保存非ERC20合约信息到数据库（从地址查询）
func (p *Process) saveNonERC20ContractInfoFromAddress(contractAddress *common.Address, block *types.Block, contractType string) error {
	if p.db == nil {
		return fmt.Errorf("database not initialized")
	}

	contract := &database.Contract{
		ContractAddress:    normalizeAddress(contractAddress.Hex()),
		ContractName:       "Unknown",
		ContractSymbol:     "UNKNOWN",
		ContractType:       contractType,
		Decimals:           0, // 非ERC20合约没有decimals
		TotalSupply:        big.NewInt(0),
		DeployTxHash:       "",                                // 未知，因为不是从部署交易中发现的
		DeployBlockNumber:  0,                                 // 未知
		DeployBlockTime:    time.Unix(int64(block.Time()), 0), // 使用当前区块时间作为参考
		DeployerAddress:    "",                                // 未知
		VerificationStatus: 0,                                 // 未验证
		VerifiedFunctions:  "[]",                              // 空数组
	}

	return p.db.SaveContract(contract)
}

// saveTransactionToDB 保存交易信息到数据库
func (p *Process) saveTransactionToDB(tx *types.Transaction, receipt *types.Receipt, block *types.Block, contractAddress common.Address, funcSelector, funcName string) error {
	if p.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// 尝试获取发送者地址
	var fromAddr string
	if tx.ChainId() != nil {
		signer := types.NewEIP155Signer(tx.ChainId())
		if sender, err := types.Sender(signer, tx); err == nil {
			fromAddr = sender.Hex()
		}
	}

	var toAddr string
	if tx.To() != nil {
		toAddr = tx.To().Hex()
	}

	// 解析交易数据获取函数选择器和参数
	txData := tx.Data()
	var value *big.Int
	if len(txData) >= 4 {
		// 如果是transfer函数，解析value参数
		if funcSelector == "a9059cbb" && len(txData) >= 68 {
			// transfer(address,uint256) - 参数从第4字节开始
			value = new(big.Int).SetBytes(txData[36:68])
		} else if funcSelector == "23b872dd" && len(txData) >= 100 {
			// transferFrom(address,address,uint256) - value在最后32字节
			value = new(big.Int).SetBytes(txData[68:100])
		}
	}

	gasPrice := tx.GasPrice()
	var txFee *big.Int
	if gasPrice != nil && receipt.GasUsed > 0 {
		txFee = new(big.Int).Mul(gasPrice, big.NewInt(int64(receipt.GasUsed)))
	}

	// 打印 funcSelector 用于调试
	log.Debug("Saving transaction",
		"txHash", tx.Hash().Hex(),
		"funcSelector", funcSelector,
		"funcSelectorLen", len(funcSelector),
		"funcName", funcName,
		"contract", contractAddress.Hex())

	// 确保 funcSelector 不超过数据库字段长度限制
	// 注意：数据库字段已修改为 VARCHAR(50)，但为了兼容性，这里仍然检查
	finalFuncSelector := funcSelector
	if len(funcSelector) > 50 {
		log.Warn("funcSelector too long, truncating",
			"original", funcSelector,
			"length", len(funcSelector),
			"txHash", tx.Hash().Hex())
		// 截断到50个字符
		finalFuncSelector = funcSelector[:50]
	}

	dbTx := &database.Transaction{
		TxHash:          tx.Hash().Hex(),
		BlockNumber:     block.NumberU64(),
		BlockHash:       block.Hash().Hex(),
		BlockTime:       time.Unix(int64(block.Time()), 0),
		TxIndex:         uint(receipt.TransactionIndex),
		FromAddress:     normalizeAddress(fromAddr),
		ToAddress:       normalizeAddress(toAddr),
		ContractAddress: normalizeAddress(contractAddress.Hex()),
		FuncSelector:    finalFuncSelector,
		FuncName:        funcName,
		Value:           value,
		GasLimit:        tx.Gas(),
		GasUsed:         receipt.GasUsed,
		GasPrice:        gasPrice,
		TxFee:           txFee,
		Status:          int8(receipt.Status),
		TxData:          hex.EncodeToString(txData),
	}

	return p.db.SaveTransaction(dbTx)
}

// saveEventToDB 保存事件信息到数据库
func (p *Process) saveEventToDB(transfer *TransferInfo, block *types.Block, txHash common.Hash, logIndex uint) error {
	if p.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Transfer事件的topic0签名
	transferEventSig := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	dbEvent := &database.Event{
		TxHash:          txHash.Hex(),
		BlockNumber:     block.NumberU64(),
		BlockTime:       time.Unix(int64(block.Time()), 0),
		LogIndex:        logIndex,
		ContractAddress: normalizeAddress(transfer.TokenAddress.Hex()),
		EventName:       "Transfer",
		EventSignature:  transferEventSig.Hex(),
		FromAddress:     normalizeAddress(transfer.From.Hex()),
		ToAddress:       normalizeAddress(transfer.To.Hex()),
		Value:           transfer.Value,
		Topic0:          transferEventSig.Hex(),
		Topic1:          common.BytesToHash(transfer.From.Bytes()).Hex(),
		Topic2:          common.BytesToHash(transfer.To.Bytes()).Hex(),
	}

	return p.db.SaveEvent(dbEvent)
}

// updateBalanceInDB 更新地址余额到数据库
func (p *Process) updateBalanceInDB(address, contractAddress common.Address, balance *big.Int, txHash common.Hash, blockNumber uint64) error {
	if p.db == nil {
		return fmt.Errorf("database not initialized")
	}

	return p.db.UpdateTokenBalance(
		normalizeAddress(address.Hex()),
		normalizeAddress(contractAddress.Hex()),
		balance,
		txHash.Hex(),
		blockNumber,
	)
}

// touchBalanceRowInDB 在关闭扫块内联 balanceOf 时调用：若无行则插入 balance=0 与 last_tx_*；若已有行则只更新 last_tx_*（不覆盖 balance），真实余额由 runBalanceRefresher 写入。
func (p *Process) touchBalanceRowInDB(address, contractAddress common.Address, txHash common.Hash, blockNumber uint64) error {
	if p.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return p.db.UpsertTokenBalanceMetadata(
		normalizeAddress(address.Hex()),
		normalizeAddress(contractAddress.Hex()),
		txHash.Hex(),
		blockNumber,
	)
}

// runBalanceRefresher 定时从链上 balanceOf 刷新 token_balances（仅更新 balance 与 last_updated_at）
func (p *Process) RunBalanceRefresher(ctx context.Context, br config.BalanceRefresherConfig) {
	interval, minAge, err := br.ParseBalanceRefresherDurations()
	if err != nil {
		log.Error("balance_refresher: invalid duration config", "err", err)
		return
	}
	batch := br.BatchSize
	if batch <= 0 {
		batch = 50
	}
	concurrency := br.Concurrency
	if concurrency <= 0 {
		concurrency = 5
	}
	log.Info("balance_refresher started",
		"interval", interval.String(),
		"batch_size", batch,
		"concurrency", concurrency,
		"min_age", minAge.String())

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Info("balance_refresher stopped")
			return
		case <-ticker.C:
			p.refreshBalanceBatch(batch, minAge, concurrency)
		}
	}
}

func (p *Process) refreshBalanceBatch(batchSize int, minAge time.Duration, concurrency int) {
	if p.db == nil {
		return
	}
	rows, err := p.db.ListTokenBalanceRowsForRefresh(batchSize, minAge)
	if err != nil {
		log.Error("balance_refresher: list rows", "err", err)
		return
	}
	if len(rows) == 0 {
		return
	}

	var ok, fail atomic.Int64
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for _, row := range rows {
		row := row
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			addr := common.HexToAddress(row.Address)
			contract := common.HexToAddress(row.ContractAddress)
			bal, err := p.getBalanceFromContract(&contract, &addr)
			if err != nil {
				fail.Add(1)
				log.Debug("balance_refresher: balanceOf failed",
					"err", err, "address", row.Address, "contract", row.ContractAddress)
				return
			}
			err = p.db.UpdateTokenBalanceFromChain(
				normalizeAddress(row.Address),
				normalizeAddress(row.ContractAddress),
				bal,
			)
			if err != nil {
				fail.Add(1)
				log.Error("balance_refresher: update db",
					"err", err, "address", row.Address, "contract", row.ContractAddress)
				return
			}
			ok.Add(1)
		}()
	}
	wg.Wait()
	log.Info("balance_refresher tick",
		"rows", len(rows),
		"ok", ok.Load(),
		"fail", fail.Load())
}
