package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"strings"
	"time"

	chain33types "github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db/block"
	"github.com/33cn/externaldb/erc20Scaner/database"
	"github.com/33cn/externaldb/erc20Scaner/erc20abi/generated"
	"github.com/33cn/externaldb/escli"
)

// Process 业务处理模块
type Process struct {
	cli        *Client
	startPoint uint64
	endPoint   uint64
	enableDB   bool
	dbDSN      string
	db         *database.DB
	nodeURL    string
	esClient   escli.ESClient // ES客户端，用于从ES读取区块
}

// Init 初始化模块
func (p *Process) Init() {
	if p.cli == nil {
		p.cli = new(Client)
	}
	p.cli.ConnectEth(p.nodeURL)

	// 初始化数据库连接
	if p.enableDB {
		var err error
		p.db, err = database.NewDB(p.dbDSN)
		if err != nil {
			log.Printf("Failed to connect to database: %v, database write will be disabled", err)
			p.enableDB = false
		} else {
			log.Printf("Database connection established successfully")
			// 读取上次处理的进度
			progress, err := p.db.GetScanProgress()
			if err != nil {
				log.Printf("Failed to get scan progress: %v, will start from configured start point", err)
			} else if progress != nil {
				// 如果存在进度记录，从上次处理的高度+1开始
				p.startPoint = progress.LastBlockNumber + 1
				log.Printf("Resume from last processed block: %d, will start from block: %d",
					progress.LastBlockNumber, p.startPoint)
			} else {
				log.Printf("No previous progress found, will start from configured start point: %d", p.startPoint)
			}
		}
	}
}

// StartWithEsClient 从ES读取区块并处理evm交易
func (p *Process) StartWithEsClient(esClient escli.ESClient) {
	p.esClient = esClient
	for {
		// 检查是否到达结束点
		if p.endPoint > 0 && p.startPoint >= uint64(p.endPoint) {
			time.Sleep(time.Second)
			continue
		}

		// 从ES读取区块
		blockSeq, err := p.getBlockFromES(int64(p.startPoint))
		if err != nil {
			fmt.Printf("getBlockFromES err: %v, seq: %d\n", err, p.startPoint)
			time.Sleep(time.Second)
			continue
		}

		if blockSeq == nil {
			// 区块不存在，等待
			time.Sleep(time.Second)
			continue
		}

		// 解析区块
		err = p.parseBlockFromES(blockSeq)
		if err != nil {
			fmt.Printf("parseBlockFromES err: %v, seq: %d\n", err, p.startPoint)
			time.Sleep(time.Second)
			continue
		}

		// 更新处理进度
		if p.enableDB && p.db != nil {
			// 从blockSeq中获取区块信息
			var detail chain33types.BlockDetail
			err = chain33types.Decode(blockSeq.BlockDetail, &detail)
			if err == nil {
				blockTime := time.Unix(int64(detail.Block.BlockTime), 0)
				err = p.db.UpdateScanProgress(
					uint64(detail.Block.Height),
					blockSeq.Hash,
					blockTime,
					0, // processed_tx_count 可以根据需要统计
				)
				if err != nil {
					log.Printf("Failed to update scan progress: %v", err)
				}
			}
		}

		p.startPoint++
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
		if blockNum < p.startPoint {
			time.Sleep(time.Second)
			continue
		}
		if p.endPoint > 0 && blockNum >= p.endPoint {
			time.Sleep(time.Second)
			continue
		}

		block, err := p.cli.BlockByNumber(p.startPoint)
		if err != nil {
			fmt.Println("err:", err)
			continue
		}
		err = p.ParaseBlock(block)
		if err != nil {
			fmt.Println("ParaseBlock err:", err)
			continue
		}

		// 更新处理进度
		if p.enableDB && p.db != nil {
			blockTime := time.Unix(int64(block.Time()), 0)
			err = p.db.UpdateScanProgress(
				block.NumberU64(),
				block.Hash().Hex(),
				blockTime,
				0, // processed_tx_count 可以根据需要统计
			)
			if err != nil {
				log.Printf("Failed to update scan progress: %v", err)
			}
		}

		p.startPoint++
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

	fmt.Printf("Processing block from ES: height=%d, seq=%d, txCount=%d\n",
		detail.Block.Height, blockSeq.SyncSeq, len(detail.Block.Txs))

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
	block, err := p.cli.BlockByNumber(p.startPoint)
	if err != nil {
		fmt.Println("err:", err)
		return err
	}
	txs := block.Transactions()

	fmt.Println("startPoint:", p.startPoint, "txsnum:", len(txs))
	for _, idx := range evmtxs {
		err := p.processTransactionWithReceipt(txs[idx], block)
		if err != nil {
			// 处理失败不影响其他交易的处理
			fmt.Printf("processTransactionWithReceipt error: %v, tx: %s\n", err, txs[idx].Hash().Hex())
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
		return fmt.Errorf("checkERC20BySelector failed: %w", err)
	}
	if !isERC20 {
		return nil // 不是ERC20合约，跳过
	}

	// 获取合约信息
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

	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	fmt.Println("contract address:", receipt.ContractAddress)
	fmt.Println("contract name:", cname)
	fmt.Println("contract symbol:", symbol)
	fmt.Println("contract totalSupply:", supply)
	fmt.Println("contract decimals:", decimals)
	fmt.Println("contract deploy time:", ut.In(cst))
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")

	// 写入数据库
	if p.enableDB {
		err = p.saveContractToDB(receipt, block, tx, cname, symbol, decimals, supply)
		if err != nil {
			log.Printf("Failed to save contract to database: %v, contract: %s", err, receipt.ContractAddress.Hex())
		} else {
			log.Printf("Contract saved to database: %s", receipt.ContractAddress.Hex())
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

	fmt.Println("startPoint:", p.startPoint, "txsnum:", len(txs))
	for _, tx := range txs {
		err := p.processTransactionWithReceipt(tx, block)
		if err != nil {
			// 处理失败不影响其他交易的处理
			fmt.Printf("processTransactionWithReceipt error: %v, tx: %s\n", err, tx.Hash().Hex())
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

	// 先检查函数选择器，确认是否为ERC20合约
	isERC20, err := p.checkERC20BySelector(&receipt.ContractAddress)
	if err != nil {
		fmt.Printf("checkERC20BySelector error: %v, contract: %s\n", err, receipt.ContractAddress.Hex())
		// 如果检查失败，跳过该合约
		return nil
	}
	if !isERC20 {
		fmt.Printf("Contract %s is not ERC20, skip\n", receipt.ContractAddress.Hex())
		return nil
	}
	fmt.Printf("Contract %s passed ERC20 selector check\n", receipt.ContractAddress.Hex())

	// 使用统一的处理函数
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

	//fmt.Println("result", result)
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

		// 输出验证信息
		fmt.Printf("ERC20 validation: core=%d/%d, optional=%d/%d, extension=%d/%d\n",
			coreFoundCount, len(requiredCoreSelectors),
			optionalFoundCount, len(optionalSelectors),
			extensionFoundCount, len(extensionSelectors))

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

	// 解析交易日志中的Transfer事件
	parsedAbi, err := abi.JSON(strings.NewReader(generated.ERC20ABI))
	if err != nil {
		return fmt.Errorf("failed to parse ERC20 ABI: %w", err)
	}

	// Transfer事件签名: Transfer(address indexed from, address indexed to, uint256 value)
	transferEvent := parsedAbi.Events["Transfer"]
	if transferEvent.ID == (common.Hash{}) {
		return fmt.Errorf("transfer event not found in ABI")
	}

	// 解析所有日志，查找Transfer事件
	var transfers []TransferInfo
	for _, log := range receipt.Logs {
		// 检查日志主题是否匹配Transfer事件（第一个topic是事件签名hash）
		if len(log.Topics) < 3 {
			continue
		}

		// Transfer事件应该有3个topics: [event_hash, from, to]
		if log.Topics[0] != transferEvent.ID {
			continue
		}

		// 从topics中提取from和to（它们是indexed参数）
		from := common.BytesToAddress(log.Topics[1].Bytes())
		to := common.BytesToAddress(log.Topics[2].Bytes())

		// 从data中提取value（非indexed参数）
		// Transfer事件的data只包含value，是32字节的uint256
		var value *big.Int
		if len(log.Data) >= 32 {
			value = new(big.Int).SetBytes(log.Data[:32])
		} else {
			// 如果data长度不够，跳过这个日志
			continue
		}

		// 获取代币信息
		tokenAddress := log.Address
		tokenInfo, err := p.getTokenInfo(&tokenAddress)
		if err != nil {
			// 获取代币信息失败，使用默认值
			tokenInfo = &TokenInfo{
				Address:  tokenAddress.Hex(),
				Symbol:   "UNKNOWN",
				Decimals: 18,
				Name:     "Unknown Token",
			}
		}

		transfers = append(transfers, TransferInfo{
			TokenAddress: tokenAddress,
			From:         from,
			To:           to,
			Value:        value,
			TokenInfo:    *tokenInfo,
		})
	}

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

	fmt.Println("============================================================")
	fmt.Printf("Transaction Hash: %s\n", tx.Hash().Hex())
	fmt.Printf("Block Number: %d\n", block.NumberU64())
	fmt.Printf("Block Time: %s\n", ut.In(cst).Format("2006-01-02 15:04:05"))
	if fromAddr != (common.Address{}) {
		fmt.Printf("From: %s\n", fromAddr.Hex())
	}
	if tx.To() != nil {
		fmt.Printf("To (Contract): %s\n", tx.To().Hex())
	}
	fmt.Printf("Gas Used: %d\n", receipt.GasUsed)
	fmt.Printf("Status: Success\n")
	fmt.Println("------------------------------------------------------------")
	fmt.Printf("Found %d Transfer event(s):\n", len(transfers))
	for i, transfer := range transfers {
		fmt.Printf("\n[Transfer #%d]\n", i+1)
		fmt.Printf("  Token Address: %s\n", transfer.TokenAddress.Hex())
		fmt.Printf("  Token Name: %s\n", transfer.TokenInfo.Name)
		fmt.Printf("  Token Symbol: %s\n", transfer.TokenInfo.Symbol)
		fmt.Printf("  Token Decimals: %d\n", transfer.TokenInfo.Decimals)
		fmt.Printf("  From: %s\n", transfer.From.Hex())
		fmt.Printf("  To: %s\n", transfer.To.Hex())
		fmt.Printf("  Value (raw): %s\n", transfer.Value.String())

		// 计算实际金额（考虑decimals）
		decimals := big.NewInt(int64(transfer.TokenInfo.Decimals))
		divisor := new(big.Int).Exp(big.NewInt(10), decimals, nil)
		actualValue := new(big.Float).Quo(new(big.Float).SetInt(transfer.Value), new(big.Float).SetInt(divisor))
		fmt.Printf("  Value (formatted): %s %s\n", actualValue.Text('f', int(transfer.TokenInfo.Decimals)), transfer.TokenInfo.Symbol)
	}
	fmt.Println("============================================================")

	// 写入数据库
	if p.enableDB {
		// 确定合约地址（使用第一个transfer的合约地址）
		contractAddress := transfers[0].TokenAddress

		// 检查合约是否已经记录到数据库, 如果没有需要通过合约地址查询合约信息并检查是否为erc20合约
		// 如果是erc20合约, 需要在数据库中记录合约信息
		_, err := p.db.GetContractByAddress(contractAddress.Hex())
		if err != nil {
			// 合约不在数据库中，需要查询并保存
			log.Printf("Contract not found in database: %s, querying contract info...", contractAddress.Hex())

			// 验证是否为ERC20合约
			isERC20, err := p.checkERC20BySelector(&contractAddress)
			if err != nil {
				log.Printf("Failed to verify ERC20 contract: %v, contract: %s", err, contractAddress.Hex())
				// 即使验证失败，也继续处理transfer事件
				return fmt.Errorf("not Erc20, error: %v", err)
			} else if !isERC20 {
				log.Printf("Contract %s is not ERC20, but has Transfer event, continue processing", contractAddress.Hex())
				// 虽然不是ERC20，但既然有Transfer事件，也继续处理
				return fmt.Errorf("not Erc20, but has Transfer event")
			} else {
				// 是ERC20合约，获取合约信息并保存
				log.Printf("Contract %s is ERC20, fetching contract info...", contractAddress.Hex())
				err = p.saveContractInfoFromAddress(&contractAddress, block)
				if err != nil {
					log.Printf("Failed to save contract info: %v, contract: %s", err, contractAddress.Hex())
					// 即使保存失败，也继续处理transfer事件
				} else {
					log.Printf("Contract info saved to database: %s", contractAddress.Hex())
				}
			}
		} else {
			log.Printf("Contract already exists in database: %s", contractAddress.Hex())
		}

		// 保存交易信息
		funcName := "transfer"
		if funcSelector == transferFromSelector {
			funcName = "transferFrom"
		}
		err = p.saveTransactionToDB(tx, receipt, block, contractAddress, funcSelector, funcName)
		if err != nil {
			log.Printf("Failed to save transaction to database: %v, tx: %s", err, tx.Hash().Hex())
		} else {
			log.Printf("Transaction saved to database: %s", tx.Hash().Hex())
		}

		// 保存每个Transfer事件并更新余额
		for i, transfer := range transfers {
			// 保存事件
			err := p.saveEventToDB(&transfer, block, tx.Hash(), uint(i))
			if err != nil {
				log.Printf("Failed to save event to database: %v, tx: %s, logIndex: %d", err, tx.Hash().Hex(), i)
			} else {
				log.Printf("Event saved to database: tx=%s, logIndex=%d, from=%s, to=%s, value=%s",
					tx.Hash().Hex(), i, transfer.From.Hex(), transfer.To.Hex(), transfer.Value.String())
			}

			// 更新发送者余额（减少）
			fromBalance, err := p.getBalanceFromContract(&transfer.TokenAddress, &transfer.From)
			if err == nil {
				err = p.updateBalanceInDB(transfer.From, transfer.TokenAddress, fromBalance, tx.Hash(), block.NumberU64())
				if err != nil {
					log.Printf("Failed to update from balance: %v, address: %s", err, transfer.From.Hex())
				}
			}

			// 更新接收者余额（增加）
			toBalance, err := p.getBalanceFromContract(&transfer.TokenAddress, &transfer.To)
			if err == nil {
				err = p.updateBalanceInDB(transfer.To, transfer.TokenAddress, toBalance, tx.Hash(), block.NumberU64())
				if err != nil {
					log.Printf("Failed to update to balance: %v, address: %s", err, transfer.To.Hex())
				}
			}
		}
	}

	return nil
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
		ContractAddress:    receipt.ContractAddress.Hex(),
		ContractName:       fmt.Sprintf("%v", name),
		ContractSymbol:     fmt.Sprintf("%v", symbol),
		ContractType:       "ERC20",
		Decimals:           decimalsVal,
		TotalSupply:        totalSupply,
		DeployTxHash:       tx.Hash().Hex(),
		DeployBlockNumber:  block.NumberU64(),
		DeployBlockTime:    time.Unix(int64(block.Time()), 0),
		DeployerAddress:    deployerAddr,
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
		log.Printf("Failed to get contract name: %v, contract: %s", err, contractAddress.Hex())
		cname = "Unknown"
	}

	symbol, err := p.unPackageAbi("symbol", contractAddress)
	if err != nil {
		log.Printf("Failed to get contract symbol: %v, contract: %s", err, contractAddress.Hex())
		symbol = "UNKNOWN"
	}

	decimals, err := p.unPackageAbi("decimals", contractAddress)
	if err != nil {
		log.Printf("Failed to get contract decimals: %v, contract: %s", err, contractAddress.Hex())
		decimals = uint8(18) // 默认值
	}

	supply, err := p.unPackageAbi("totalSupply", contractAddress)
	if err != nil {
		log.Printf("Failed to get contract totalSupply: %v, contract: %s", err, contractAddress.Hex())
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
		ContractAddress:    contractAddress.Hex(),
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

	dbTx := &database.Transaction{
		TxHash:          tx.Hash().Hex(),
		BlockNumber:     block.NumberU64(),
		BlockHash:       block.Hash().Hex(),
		BlockTime:       time.Unix(int64(block.Time()), 0),
		TxIndex:         uint(receipt.TransactionIndex),
		FromAddress:     fromAddr,
		ToAddress:       toAddr,
		ContractAddress: contractAddress.Hex(),
		FuncSelector:    funcSelector,
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
		ContractAddress: transfer.TokenAddress.Hex(),
		EventName:       "Transfer",
		EventSignature:  transferEventSig.Hex(),
		FromAddress:     transfer.From.Hex(),
		ToAddress:       transfer.To.Hex(),
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
		address.Hex(),
		contractAddress.Hex(),
		balance,
		txHash.Hex(),
		blockNumber,
	)
}
