// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package main 导出各种合约数据到外部数据库
package main

import (
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/33cn/externaldb/util/cli/convert"

	l "github.com/33cn/chain33/common/log/log15"
	_ "github.com/33cn/externaldb/db/account"
	_ "github.com/33cn/externaldb/db/coins"
	_ "github.com/33cn/externaldb/db/coinsx"
	_ "github.com/33cn/externaldb/db/evm"
	_ "github.com/33cn/externaldb/db/evm/erc1155"
	_ "github.com/33cn/externaldb/db/evm/erc721"
	_ "github.com/33cn/externaldb/db/evm/nft"
	_ "github.com/33cn/externaldb/db/evmxgo"
	_ "github.com/33cn/externaldb/db/filepart"
	_ "github.com/33cn/externaldb/db/filesummary"
	_ "github.com/33cn/externaldb/db/multisig"
	_ "github.com/33cn/externaldb/db/pos33"
	_ "github.com/33cn/externaldb/db/proof"
	_ "github.com/33cn/externaldb/db/proof_config"
	_ "github.com/33cn/externaldb/db/ticket"
	_ "github.com/33cn/externaldb/db/token"
	_ "github.com/33cn/externaldb/db/trade"
	_ "github.com/33cn/externaldb/db/unfreeze"
	"github.com/33cn/externaldb/proto"
	"github.com/33cn/externaldb/util"
	"github.com/33cn/externaldb/version"
	tml "github.com/BurntSushi/toml"

	_ "github.com/33cn/externaldb/stat/block"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin/dapp/init"
	_ "github.com/33cn/plugin/plugin/store/init"
)

var (
	log        = l.New("module", "main")
	configPath = flag.String("f", "externaldb.toml", "config file")
	chainPath  = flag.String("c", "", "chain33 or para node config file")
)

// 对比不同的合约有几个不同的点： 流程一样， 但流程的部分实现不一样
// 1. config
// 2. initIndex
// 3. 具体 tx 如何处理, getCurrentSeq
// 4. 用配置，生成生成交易处理的工具

func main() {
	log.Info("convert", "version", version.GetVersion())
	d, _ := os.Getwd()
	log.Debug("current dir:", "dir", d)
	_ = os.Chdir(pwd())

	flag.Parse()

	cfg := InitCfgNew(*configPath)
	title := cfg.Chain.Title
	if title == "" {
		title = "bityuan"
	}
	symbol := cfg.Chain.Symbol
	if symbol == "" {
		symbol = "bty"
	}
	util.InitChain33(title, symbol, *chainPath)
	util.SetupLog(cfg.Convert.GetAppName(), "debug")
	util.InitMapSet(cfg.EsVersion, cfg.EsIndex)

	log.Info("init config", "config", cfg)
	log.Info("load config", "cfgPath", *configPath, "read_es", cfg.SyncEs.Host, "write_es", cfg.ConvertEs.Host)
	convert.InitDB(cfg)
	// 启动服务
	convertService := convert.NewConvertService(cfg)
	log.Info("main   ", "init", "convertService")
	go convertService.Start()

	gracefulCLoseConvert()
}

// SetDataConfig data config rule
// 配置在common项中的， 对每个合约生效
// func SetDataConfig(cfg *proto.AppConfig) {
// 	foundCommon := false
// 	commonIndex := -1
// 	for i, data := range cfg.Data {
// 		if data.Exec == common.NameX {
// 			foundCommon = true
// 			commonIndex = i
// 			break
// 		}
// 	}
// 	if !foundCommon {
// 		return
// 	}
// 	commonCfg := cfg.Data[commonIndex]
// 	if len(commonCfg.Generate) == 0 {
// 		return
// 	}
// 	for _, data := range cfg.Data {
// 		if data.Exec != common.NameX {
// 			data.Generate = append(data.Generate, commonCfg.Generate...)
// 		}
// 	}
// }

// InitCfgNew 初始化cfg
func InitCfgNew(path string) *proto.ConfigNew {
	var cfg proto.ConfigNew
	if _, err := tml.DecodeFile(path, &cfg); err != nil {
		log.Info("init config failed", "err", err)
		os.Exit(0)
	}

	// SetDataConfig(&cfg)
	return &cfg
}

func pwd() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return dir
}

func gracefulCLoseConvert() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// nolint
	select {
	case <-sigs:
		util.ConvertServerStatus.CloseServer()
		time.Sleep(5 * time.Second)
		log.Info("convert服务关闭")
	}
}
