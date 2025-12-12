package database

// 使用示例
//
// 1. 安装MySQL驱动:
//    go get github.com/go-sql-driver/mysql
//
// 2. 初始化数据库:
//    mysql -u root -p < schema.sql
//
// 3. 使用示例:
//
//    import (
//        "github.com/33cn/externaldb/erc20Scaner/database"
//    )
//
//    // 连接数据库
//    dsn := "user:password@tcp(localhost:3306)/token_scanner?charset=utf8mb4&parseTime=True&loc=Local"
//    db, err := database.NewDB(dsn)
//    if err != nil {
//        log.Fatal(err)
//    }
//    defer db.Close()
//
//    // 查询函数签名
//    funcSig, err := db.GetFunctionSignature("a9059cbb") // transfer函数
//    if err == nil {
//        fmt.Printf("Function: %s, Standard: %s\n", funcSig.FunctionName, funcSig.StandardType)
//    }
//
//    // 获取ERC20标准的所有函数签名
//    erc20Funcs, err := db.GetFunctionSignaturesByStandard("ERC20")
//    if err == nil {
//        fmt.Printf("Found %d ERC20 functions\n", len(erc20Funcs))
//    }
//
//    // 保存合约信息
//    contract := &database.Contract{
//        ContractAddress:   "0x...",
//        ContractName:      "Tether USD",
//        ContractSymbol:    "USDT",
//        ContractType:      "ERC20",  // 合约类型：ERC20/ERC721/ERC1155等
//        Decimals:          6,
//        TotalSupply:       big.NewInt(1000000000000),
//        DeployTxHash:     "0x...",
//        DeployBlockNumber: 12345678,
//        DeployBlockTime:  time.Now(),
//        DeployerAddress:   "0x...",
//        VerificationStatus: 1,
//        VerifiedFunctions: `["18160ddd","70a08231","a9059cbb"]`, // JSON格式的函数选择器列表
//    }
//    err = db.SaveContract(contract)
//
//    // 保存交易信息
//    tx := &database.Transaction{
//        TxHash:          "0x...",
//        BlockNumber:     12345679,
//        BlockTime:       time.Now(),
//        FromAddress:     "0x...",
//        ToAddress:       "0x...",
//        ContractAddress: "0x...",
//        FuncSelector:    "a9059cbb",
//        FuncName:        "transfer",
//        Value:           big.NewInt(1000000),
//        GasUsed:         21000,
//        Status:          1,
//    }
//    err = db.SaveTransaction(tx)
//
//    // 更新余额
//    err = db.UpdateTokenBalance(
//        "0x...",  // address
//        "0x...",  // contract address
//        big.NewInt(1000000),  // balance
//        "0x...",  // last tx hash
//        12345679, // last tx block number
//    )
//
//    // 保存事件
//    event := &database.Event{
//        TxHash:          "0x...",
//        BlockNumber:     12345679,
//        BlockTime:       time.Now(),
//        LogIndex:        0,
//        ContractAddress: "0x...",
//        EventName:       "Transfer",
//        FromAddress:     "0x...",
//        ToAddress:       "0x...",
//        Value:           big.NewInt(1000000),
//        Topic0:          "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
//        Topic1:          "0x000000000000000000000000...",
//        Topic2:          "0x000000000000000000000000...",
//    }
//    err = db.SaveEvent(event)
//
//    // 查询操作
//    contract, err := db.GetContractByAddress("0x...")
//    balances, err := db.GetTokenBalancesByAddress("0x...")
//    events, err := db.GetEventsByContract("0x...", 100)
