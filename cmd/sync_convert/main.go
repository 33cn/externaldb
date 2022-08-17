// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package main 数据导出到外部数据库
// 使用 blockchain push seq 的接口
// AddSeqCallBack + postData
package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	l "github.com/33cn/chain33/common/log/log15"
	tml "github.com/BurntSushi/toml"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/proto"
	"github.com/33cn/externaldb/util"
	"github.com/33cn/externaldb/util/cli/sync"
	"github.com/33cn/externaldb/version"
)

var (
	log        = l.New("module", "main")
	configPath = flag.String("f", "externaldb.toml", "config file")
)

func main() {
	// 处理工作路径
	log.Info("sync_convert", "version", version.GetVersion())
	d1, _ := os.Getwd()
	os.Chdir(pwd())
	d, _ := os.Getwd()
	log.Debug("work dir:", "work-dir", d, "start-cmd-dir", d1)

	// 加载配置文件
	flag.Parse()
	cfg := InitCfgNew(*configPath)

	// 初始化配置
	util.SetupLog(cfg.Sync.GetName(), "debug")
	util.InitMapSet(cfg.EsVersion, cfg.EsIndex)
	log.Info("init config", "config", cfg)
	log.Info("load config", "cfgPath", *configPath, "h1", cfg.Chain.Host, "h2", "me", cfg.Sync.PushHost, "me-bind", cfg.Sync.PushBind)
	db.SetVersion(cfg.EsVersion)

	log.Debug("started ")

	// 连接ES
	EsWrite, err := escli.NewESLongConnect(cfg.ConvertEs.Host, cfg.ConvertEs.Prefix, cfg.EsVersion, cfg.ConvertEs.User, cfg.ConvertEs.Pwd)
	if err != nil {
		log.Error("ES Connect failed", "err", err.Error())
		log.Error("ES 连接失败，请确保ES服务正常开启，ES配置正确，网络通常",
			"当前ES配置为 host ", cfg.ConvertEs.Host, " Prefix ", cfg.ConvertEs.Prefix, " EsVersion ", cfg.EsVersion,
			" User ", cfg.ConvertEs.User, " Pwd ", cfg.ConvertEs.Pwd)
		return
	}

	// 自检ES各项配置，如果存在问题，则开始自动修复，修复成功，继续，修复不成功，退出
	err = util.ESCheckAndRepair(cfg, EsWrite)
	if err != nil {
		log.Error("ESCheckAndRepair failed", "err", err.Error())
		log.Error("ES 自检修复失败，请手动执行proof-init脚本，初始化ES数据")
		return
	}

	// TODO db.LastSeqDB 待确定
	err = sync.InitLastSyncSeqCache(EsWrite, db.LastSeqDB, cfg.Sync.StartSeq)
	if err != nil {
		log.Error("InitLastSyncSeqCache failed", "err", err.Error())
		log.Error("初始化 区块解析进度参数 last_seq 失败，请确保ES服务正常且 配置文件参数sync.startSeq 参数大于或等于0")
		return
	}
	// 初始化设置 convert流程的 ES 服务是否设置批量提交
	sync.InitConvertEsBulk(cfg.ConvertEs.Bulk)

	// 创建服务实例
	receiver, err := sync.CreateReceiverConvert(cfg, EsWrite)
	if err != nil {
		log.Error("CreateReceiver failed", "err", err.Error())
		log.Error("创建sync服务失败，请检查配置文件各项参数")
		return
	}

	// 加载解析插件
	err = receiver.RecoverStats()
	if err != nil {
		log.Error("RecoverStats failed", "err", err.Error())
		log.Error("注册加载解析插件失败，请检查配置文件各项参数 convert.stat")
		return
	}

	// 注册推送
	err = receiver.Register()
	if err != nil {
		log.Error("Register client failed", "err", err.Error())
		log.Error("注册推送失败，请确保链服务开启, 请检查配置 sync里的各项参数， 以及网络，请确保网络能够正常访问", "sync里的参数", cfg.Sync, "区块链节点参数", cfg.Chain)
		if strings.Contains(err.Error(), "ErrPushNotSupport") {
			log.Error("注册推送失败, 链节点可能未开启支持推送， 请检查链节点配置", "blockchain.enablePushSubscribe", "true", "blockchain.enableReduseLocaldb", "false")
		}
		return
	}

	// 启动http服务监听推送
	receiver.ReceiveLoop()
}

//InitCfgNew 初始化cfg
func InitCfgNew(path string) *proto.ConfigNew {
	var cfg proto.ConfigNew
	if _, err := tml.DecodeFile(path, &cfg); err != nil {
		log.Info("init config failed", "err", err)
		os.Exit(0)
	}
	return &cfg
}

func pwd() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return dir
}
