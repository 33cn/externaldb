package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/rpc/service/evm"
	"github.com/33cn/externaldb/rpc/service/file"

	l "github.com/33cn/chain33/common/log/log15"
	tml "github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"

	cors "github.com/rs/cors/wrapper/gin"
	"github.com/33cn/externaldb/proto"
	"github.com/33cn/externaldb/rpc/docs"
	"github.com/33cn/externaldb/rpc/middleware/protocol"
	"github.com/33cn/externaldb/rpc/middleware/sentinel"
	"github.com/33cn/externaldb/rpc/middleware/whitelist"
	"github.com/33cn/externaldb/rpc/service/comm"
	"github.com/33cn/externaldb/rpc/service/proof"
	"github.com/33cn/externaldb/rpc/service/proofmember"
	rpcutils "github.com/33cn/externaldb/rpc/utils"
	"github.com/33cn/externaldb/util"
	"github.com/33cn/externaldb/version"
)

var (
	log            = l.New("module", "ginrpc")
	configPath     = flag.String("f", "externaldb-test.toml", "configfile")
	swaggerHandler gin.HandlerFunc
)

// @title externaldb-rpc
// @version 1.0
// @description This is a rpc service for proof system
// @termsOfService https://chain.33.cn/

// @contract.name API Support
// @contract.url https://chain.33.cn/
// @contract.email xx@33.cn

// @host localhost:
// @BasePath /
func main() {
	log.Info("jrpc", "version", version.GetVersion())
	d, _ := os.Getwd()
	log.Debug("current dir:", "dir", d)
	os.Chdir(rpcutils.Pwd())
	flag.Parse()

	cfg := InitCfgNew(*configPath)
	util.SetupLog(cfg.Rpc.GetName(), "debug")
	chain := cfg.Chain
	dbread := &rpcutils.DBRead{Host: cfg.ConvertEs.Host, Title: chain.Title, Prefix: cfg.ConvertEs.Prefix, Symbol: chain.Symbol, ID: cfg.Convert.AppName, Version: cfg.EsVersion, Username: cfg.ConvertEs.User, Password: cfg.ConvertEs.Pwd}

	route := gin.Default()
	route.Use(LoggerToFile())
	route.Use(cors.Default())

	route.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "welcome to externaldb.rpc")
	})

	if cfg.Rpc.EnableSwagger {
		docs.SwaggerInfo.Host = cfg.Rpc.SwaggerHost
		route.GET("/docs/*any", swaggerHandler)
		route.GET("/swagger/*any", swaggerHandler)
	}
	wl := whitelist.New(cfg.Rpc.WhiteList)
	route.Use(wl.GinHandler)
	file.InitRouter(route, dbread, cfg)
	group := route.Group("/")
	group.Use(protocol.RequestParse)

	//系统流控中间件初始化
	group.Use(sentinel.StnSystem(cfg.Rpc.GetTriggerCount()))

	//初始化提供服务模块的路由表
	proof.InitRouter(group, dbread)
	proofmember.InitRouter(group, dbread)
	comm.InitRouter(group, cfg)
	evm.InitRouter(group, dbread)

	server := &http.Server{
		Addr:    cfg.Rpc.Host,
		Handler: route,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Error("HTTPService, 监听推送服务已关闭", "ip端口", server.Addr)
			} else {
				log.Error("HTTPService, 启动服务监听链推送, 失败", "ip端口", server.Addr, "err", err)
			}
		}
	}()

	log.Info("HTTPService, 启动服务监听链推送，成功", "ip端口", server.Addr)
	gracefulShutdown(context.Background(), server)
}

// //InitCfg 初始化cfg
// func InitCfg(path string) *proto.JRPCConfig {
// 	var cfg proto.JRPCConfig
// 	if _, err := tml.DecodeFile(path, &cfg); err != nil {
// 		fmt.Println(err)
// 		os.Exit(0)
// 	}
// 	fmt.Println(&cfg)
// 	return &cfg
// }

// InitCfgNew 初始化cfg
func InitCfgNew(path string) *proto.ConfigNew {
	var cfg proto.ConfigNew
	if _, err := tml.DecodeFile(path, &cfg); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	db.SetVersion(cfg.EsVersion)
	fmt.Println(&cfg)
	return &cfg
}

// 实现一个logger的中间件
func LoggerToFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		// 处理请求
		c.Next()
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		reqMethod := c.Request.Method
		reqURI := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		// 日志格式
		log.Info("GinRPC", "statusCode", statusCode, "latencyTime", latencyTime, "clientIP", clientIP, "reqMethod", reqMethod, "reqURI", reqURI)
	}

}

// // ParserClientRequest：实现一个解析请求数据的中间件。
// // 将body中的数据解析到context的inputparam字段中
// func ParserClientRequest() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var req rpcutils.ClientRequest
// 		if err := c.ShouldBindJSON(&req); err == nil {
// 			c.Set("inputparam", req)
// 		} else {
// 			var res = rpcutils.ServerResponse{Result: nil, Error: nil, ID: 0}
// 			res.Error = err.Error()
// 			c.JSON(http.StatusBadRequest, res)
// 		}
// 	}
// }

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
