package sync

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpcTypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/proto"
	"github.com/33cn/externaldb/util/health"
	"github.com/rs/cors"
	"google.golang.org/grpc"
)

// SaveSeq 保存seq
type SaveSeq func(seq *types.BlockSeqs) error

// SeqSaver 保存seq
type SeqSaver interface {
	Save(seq *types.BlockSeqs) error
}

// Receiver 从节点接受推送
// 1. 注册/重新激活推送
// 2. 接受推送: (BlockSeqs, pb/json Receiver 处理, gzip http 外部处理)
//     数据协议: *types.BlockSeqs < pb/json < gzip < http
type Receiver interface {
	Register() error
	ReceiveLoop(s SeqSaver) // 内部启动http
	// Receive(data []byte, format string) (seq *types.BlockSeqs, err error)
}

// pusher 节点推送的相关信息
type pusher struct {
	pushServer string

	// 基本信息
	name, url, encode string

	// 开始推送点
	height, seq int64
	blockHash   string
}

// push format
const (
	PushFormatJSON    = "json"
	PushFormatPb      = "pb"
	defaultPushFormat = PushFormatJSON
)

// CreateReceiver ...
func CreateReceiver(cfg *proto.ConfigNew, s SeqSaver) (Receiver, error) {
	err := checkPushFormat(cfg.Sync.PushFormat)
	if err != nil {
		log.Error("checkPushFormat failed", "err", err.Error())
		return nil, err
	}

	// 在新版本的协议中, 默认行为从注册推送时当前节点高度进行同步
	// 所以导致无法配置成从0开始同步
	// 实际上如果指定高度进行同步, 应该会在一个比较大的数值, 不会配置成1
	// 故在配置成1 时, 程序主动去获得高度 0, 1 的的区块信息
	if cfg.Sync.StartHeight == 0 && cfg.Sync.StartSeq == 0 {
		cfg.Sync.StartHeight = 1
		cfg.Sync.StartSeq = 1
		cfg.Sync.StartBlockHash = ""

		seqs, err := getBlocks(cfg.Chain.GrpcHost, 0, 1)
		if err != nil {
			log.Error("getBlocks 0-1 failed", "err", err.Error())
			return nil, err
		}
		err = s.Save(seqs)
		if err != nil {
			log.Error("getBlocks 0-1 failed", "err", err.Error())
			return nil, err
		}
	}

	p := pusher{
		pushServer: cfg.Chain.Host,
		name:       cfg.Sync.PushName,
		url:        "http://" + cfg.Sync.PushHost,
		encode:     cfg.Sync.PushFormat,

		height:    cfg.Sync.StartHeight,
		seq:       cfg.Sync.StartSeq,
		blockHash: cfg.Sync.StartBlockHash,
	}

	if cfg.Sync.StartHeight != 0 && cfg.Sync.StartBlockHash == "" {
		hash, err := getHash(cfg.Chain.Host, cfg.Sync.StartHeight)
		if err != nil {
			log.Error("getHash failed", "err", err.Error())
			return nil, err
		}
		p.blockHash = hash
	}

	// 兼容老版本
	if p.height == 0 && p.blockHash == "" {
		p.seq = 0
	}
	return NewReceiver(&p, cfg.Sync.PushBind)
}

func getBlocks(host string, heightStart, heightEnd int64) (*types.BlockSeqs, error) {
	var seqs types.BlockSeqs
	conn, err := grpc.Dial(host, grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*100)))
	if err != nil {
		panic(err)
	}

	client := types.NewChain33Client(conn)
	for i := heightStart; i <= heightEnd; i++ {
		seq, err := client.GetBlockBySeq(context.Background(), &types.Int64{Data: i})
		if err != nil {
			return nil, err
		}
		seqs.Seqs = append(seqs.Seqs, seq)
	}
	return &seqs, err
}

func getHash(host string, hight int64) (string, error) {
	var res rpcTypes.ReplyHash
	params := types.ReqInt{
		Height: hight,
	}

	ctx := jsonclient.NewRPCCtx(host, "Chain33.GetBlockHash", &params, &res)
	_, err := ctx.RunResult()
	return res.Hash, err
}

func NewReceiver(p *pusher, pushBind string) (Receiver, error) {
	v2 := &receiverV2{p: p, bindAddr: pushBind}
	v1 := &receiverV1{p: p, bindAddr: pushBind}
	var err1, err2 error
	if err2 = v2.Register(); err2 == nil {
		return v2, nil
	}
	if err1 = v1.Register(); err1 == nil {
		return v1, nil
	}
	log.Error("Register: receiverV1 err:", "err", err1)
	log.Error("Register: receiverV2 err:", "err", err2)
	return nil, fmt.Errorf("v1:%s, v2:%s", err1, err2)
}

type receiverV1 struct {
	p        *pusher
	bindAddr string
}

// Register 注册/重新激活推送
func (r *receiverV1) Register() error {
	regMothed := "Chain33.AddSeqCallBack"
	p := r.p
	params := proto.BlockSeqCB{
		Name:   p.name,
		URL:    p.url,
		Encode: p.encode,
	}

	var res rpcTypes.Reply
	ctx := jsonclient.NewRPCCtx(p.pushServer, regMothed, &params, &res)
	_, err := ctx.RunResult()
	return err
}

// ReceiveLoop receive seqs
func (r *receiverV1) ReceiveLoop(s SeqSaver) {
	handler := func(req []byte) error {
		return handleRequest(req, r.p.encode, s.Save)
	}
	startHTTPService(r.bindAddr, "*", handler)
}

type receiverV2 struct {
	p        *pusher
	bindAddr string
}

// Register 注册/重新激活推送
func (r *receiverV2) Register() error {
	regMothed := "Chain33.AddPushSubscribe"
	p := r.p
	params := proto.PushSubscribeReq{
		Name:   p.name,
		URL:    p.url,
		Encode: p.encode,
		// start push point
		LastBlockHash: p.blockHash,
		LastHeight:    p.height,
		LastSequence:  p.seq,
	}

	var res proto.ReplySubscribePushV2
	ctx := jsonclient.NewRPCCtx(p.pushServer, regMothed, &params, &res)
	_, err := ctx.RunResult()
	log.Info("receiverV2.Register", "req", &res, "err", err)
	if !res.IsOk {
		return errors.New(res.Msg)
	}
	return err
}

// ReceiveLoop receive seqs
func (r *receiverV2) ReceiveLoop(s SeqSaver) {
	handler := func(req []byte) error {
		return handleRequest(req, r.p.encode, s.Save)
	}
	startHTTPService(r.bindAddr, "*", handler)
}

func startHTTPService(url string, clientHost string, handlerReq func([]byte) error) {
	var handler http.Handler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			//fmt.Println(r.URL, r.Header, r.Body)
			beg := types.Now()
			defer func() {
				log.Info("handler", "cost", types.Since(beg))
			}()

			client := strings.Split(r.RemoteAddr, ":")[0]
			if !checkClient(client, clientHost) {
				log.Error("HandlerFunc", "client", r.RemoteAddr, "expect", clientHost)
				w.Write([]byte(`{"errcode":"-1","result":null,"msg":"reject"}`))
				return
			}

			if r.URL.Path == "/" {
				w.Header().Set("Content-type", "application/json")
				w.WriteHeader(200)
				if len(r.Header["Content-Encoding"]) >= 1 && r.Header["Content-Encoding"][0] == "gzip" {
					gr, err := gzip.NewReader(r.Body)
					body, err := ioutil.ReadAll(gr)
					if err != nil {
						log.Debug("Error while serving JSON request: %v", err)
						return
					}

					err = handlerReq(body)
					if err == nil {
						w.Write([]byte("OK"))
					} else {
						w.Write([]byte(err.Error()))
					}
				}
			} else if r.URL.Path == "/v1/health" {
				w.Header().Set("Content-type", "application/json")
				w.WriteHeader(200)
				body, _ := json.Marshal(health.GetHealth())
				w.Write(body)
			}
		})

	co := cors.New(cors.Options{})
	handler = co.Handler(handler)

	srv := &http.Server{
		Addr:    url,
		Handler: handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Error("startHTTPService, 监听推送服务已关闭", "ip端口", url)
			} else {
				log.Error("startHTTPService, 启动服务监听链推送, 失败", "ip端口", url, "err", err)
			}
		}
	}()

	log.Info("startHTTPService, 启动服务监听链推送，成功", "ip端口", url)
	gracefulShutdown(context.Background(), srv)
}

func handleRequest(body []byte, format string, saveSeq SaveSeq) error {
	beg := types.Now()
	defer func() {
		log.Info("handleRequest", "cost", types.Since(beg))
	}()
	var req types.BlockSeqs
	var err error
	if format == "json" {
		err = types.JSONToPB(body, &req)
	} else {
		err = types.Decode(body, &req)
	}
	if err != nil {
		log.Info("handleRequest", "JSONToPB", err, "req", &req)
		return err
	}
	log.Info("handleRequest", "p1", "JSONToPB", "cost", types.Since(beg))

	err = saveSeq(&req)
	log.Info("response", "err", err)
	return err
}

func checkPushFormat(pushFormat string) error {
	if pushFormat == "" {
		pushFormat = defaultPushFormat
	}
	if pushFormat != PushFormatJSON && pushFormat != PushFormatPb {
		log.Error("init config push format support json/pb", "push-format", pushFormat)
		return errors.New("bad config")
	}
	return nil
}

func checkClient(addr string, expectClient string) bool {
	if expectClient == "0.0.0.0" || expectClient == "*" {
		return true
	}
	return addr == expectClient
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
