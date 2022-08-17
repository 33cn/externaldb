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

	l "github.com/33cn/chain33/common/log/log15"
	tml "github.com/BurntSushi/toml"

	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/proto"
	"github.com/33cn/externaldb/store/syncseq"
	"github.com/33cn/externaldb/util"
	"github.com/33cn/externaldb/util/cli/sync"
	"github.com/33cn/externaldb/version"
)

var (
	log        = l.New("module", "main")
	configPath = flag.String("f", "externaldb.toml", "config file")
)

func main() {
	log.Info("sync", "version", version.GetVersion())
	d1, _ := os.Getwd()
	os.Chdir(pwd())
	d, _ := os.Getwd()
	log.Debug("work dir:", "work-dir", d, "start-cmd-dir", d1)

	flag.Parse()
	cfg := InitCfgNew(*configPath)

	util.SetupLog(cfg.Sync.GetName(), "debug")
	util.InitMapSet(cfg.EsVersion, cfg.EsIndex)
	log.Info("init config", "config", cfg)
	log.Info("load config", "cfgPath", *configPath, "h1", cfg.Chain.Host, "h2", cfg.SyncEs.Host, "me", cfg.Sync.PushHost, "me-bind", cfg.Sync.PushBind)
	db.SetVersion(cfg.EsVersion)

	log.Debug("started ")

	seqNumStore, seqStore, err := syncseq.NewSeqStore(cfg)
	if err != nil {
		log.Error("NewSeqStore failed", "err", err.Error())
		return
	}

	proc1, err := sync.NewSeqsProc(seqNumStore, seqStore, cfg.Chain)
	if err != nil {
		log.Error("NewSeqsProc failed", "err", err.Error())
		return
	}

	go proc1.Proc(cfg.Sync.StartSeq)

	receiver, err := sync.CreateReceiver(cfg)
	if err != nil {
		log.Error("Register failed", "err", err.Error())
		return
	}
	receiver.ReceiveLoop(proc1)
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
