// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/externaldb/proto"
	"github.com/33cn/externaldb/util"
	"github.com/33cn/externaldb/version"
	tml "github.com/BurntSushi/toml"
	"github.com/rs/cors"
)

var (
	log        = l.New("module", "main")
	configPath = flag.String("f", "externaldb.toml", "configfile")
)

//HTTPConn http连接
type HTTPConn struct {
	in  io.Reader
	out io.Writer
}

func (c *HTTPConn) Read(p []byte) (n int, err error)  { return c.in.Read(p) }
func (c *HTTPConn) Write(d []byte) (n int, err error) { return c.out.Write(d) }

//Close 关闭连接
func (c *HTTPConn) Close() error { return nil }

func main() {
	log.Info("jrpc", "version", version.GetVersion())
	d, _ := os.Getwd()
	log.Debug("current dir:", "dir", d)
	os.Chdir(pwd())
	flag.Parse()
	cfg := InitCfg(*configPath)
	util.SetupLog(cfg.Rpc.GetJrpcName(), "debug")
	log.Info("load config", "cfgPath", *configPath, "wl", cfg.Rpc.WhiteList, "host", cfg.Rpc.JrpcHost, "titles", cfg.Chain)
	//初始化白名单
	//返回whitelist["0.0.0.0"] = true
	whitelist := InitWhiteList(cfg)

	supports := make(map[string]*rpc.Server)
	chain := cfg.Chain
	if chain != nil {
		// 功能对象注册
		server := rpc.NewServer()

		convDB := &DBRead{Host: cfg.ConvertEs.Host, Title: chain.Title, Prefix: cfg.ConvertEs.Prefix, Symbol: chain.Symbol, Version: cfg.EsVersion, Username: cfg.ConvertEs.User, Password: cfg.ConvertEs.Pwd}
		syncDB := &DBRead{Host: cfg.SyncEs.Host, Title: chain.Title, Prefix: cfg.SyncEs.Prefix, Symbol: chain.Symbol, Version: cfg.EsVersion, Username: cfg.SyncEs.User, Password: cfg.SyncEs.Pwd}
		//account server
		shower := MinerAccount{DBRead: convDB}
		server.Register(&shower)

		//trade server
		trade := Trade{DBRead: convDB}
		server.Register(&trade)

		//token server
		token := Token{DBRead: convDB}
		InitCache(&token)
		server.Register(&token)

		//tx server
		tx := Tx{DBRead: convDB, ChainGrpc: chain.GrpcHost}
		server.Register(&tx)

		//block stat server
		bs := BlockStat{DBRead: convDB}
		server.Register(&bs)

		ms := MultiSig{DBRead: convDB}
		server.Register(&ms)

		//account server
		acc := Account{DBRead: convDB}
		server.Register(&acc)

		//evm server
		evm := EVM{DBRead: convDB}
		server.Register(&evm)

		//comm server
		comm := Comm{convDB: convDB, syncDB: syncDB, ConvertID: cfg.Convert.AppName, Version: cfg.EsVersion}
		server.Register(&comm)

		//block server
		block := Block{DBRead: convDB}
		server.Register(&block)

		evm1 := Evm{DBRead: convDB, ChainGrpc: chain.GrpcHost}
		server.Register(&evm1)

		// TODO support more
		// ... server

		supports[chain.Title] = server
	}
	// HTTP注册
	var handler http.Handler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			//fmt.Println(r.URL, r.Header, r.Body)
			//r.RemoteAddr=[::1]:50617
			if !checkWhitlist(strings.Split(r.RemoteAddr, ":")[0], whitelist) {
				log.Error("HandlerFunc", "peer not whitelist", r.RemoteAddr)
				w.Write([]byte(`{"errcode":"-1","result":null,"msg":"reject"}`))
				return
			}
			//Path返回相对路径/hello/x/x
			//withoutSlash取出左侧的“/”
			path := withoutSlash(r.URL.Path)
			if s1, ok := supports[path]; ok {
				//jsonrpc方式支持跨语言调用。
				serverCodec := jsonrpc.NewServerCodec(&HTTPConn{in: r.Body, out: w})
				w.Header().Set("Content-type", "application/json")
				w.WriteHeader(200)

				err := s1.ServeRequest(serverCodec)
				if err != nil {
					log.Debug("HandlerFunc", "Error while serving JSON request: %v", err)
					return
				}
			} else {
				log.Error("not support title", "t", r.URL.Path)
			}
		})

	//co := cors.New(cors.Options{
	//    AllowedOrigins: []string{"http://foo.com"},
	//    Debug: true,
	//})
	//跨域请求
	co := cors.New(cors.Options{})
	handler = co.Handler(handler)

	server := &http.Server{
		Addr:    cfg.Rpc.JrpcHost,
		Handler: handler,
	}

	log.Info("HTTPService，启动服务", "ip端口", server.Addr)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Error("HTTPService, 已关闭", "ip端口", server.Addr)
			} else {
				log.Error("HTTPService, 启动失败", "ip端口", server.Addr, "err", err)
				os.Exit(1)
			}
		}
	}()

	gracefulShutdown(context.Background(), server)
}

func withoutSlash(s string) string {
	return strings.Trim(s, "/")
}

//InitCfg 初始化cfg
func InitCfg(path string) *proto.ConfigNew {
	var cfg proto.ConfigNew
	if _, err := tml.DecodeFile(path, &cfg); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	fmt.Println(&cfg)
	return &cfg
}

//InitWhiteList 初始化白名单
func InitWhiteList(cfg *proto.ConfigNew) map[string]bool {
	whitelist := map[string]bool{}
	if len(cfg.Rpc.WhiteList) == 1 && cfg.Rpc.WhiteList[0] == "*" {
		whitelist["0.0.0.0"] = true
		return whitelist
	}

	for _, addr := range cfg.Rpc.WhiteList {
		log.Debug("initWhitelist", "addr", addr)
		whitelist[addr] = true
	}
	return whitelist
}

func pwd() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return dir
}

func checkWhitlist(addr string, whitlist map[string]bool) bool {
	if _, ok := whitlist["0.0.0.0"]; ok {
		return true
	}

	if _, ok := whitlist[addr]; ok {
		return true
	}
	return false
}

func gracefulShutdown(ctx context.Context, server *http.Server) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// nolint
	select {
	case <-sigs:
		timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		// 使用3秒超时的context，再使用Shutdown，
		// 当程序收到停止信号时，会执行Shutdown函数，
		//   1.停止新http链接建立；
		//   2.等服务内所有http链接正常返回或者超过超时时间, 两个条件满足一个时候，结束http服务。
		err := server.Shutdown(timeoutCtx)

		// 等待3秒，等服务内其他功能执行
		// 如果http服务结束3秒内，其他功能仍然不能在3秒内执行完毕，可以适当的延迟等待时间
		time.Sleep(3 * time.Second)
		log.Info("server.Shutdown", "err", err)
	}
}
