package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
	"time"

	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/proto"
)

type config struct {
	address string
}

var log = l.New("module", "main")

var cfg = config{
	address: "0.0.0.0:8801",
}

type Reply struct {
	IsOk bool   `json:"isOK"`
	Msg  string `json:"msg"`
}

type Sequence struct {
	Hash     string `json:"Hash"`
	Type     int64  `json:"Type"`
	Sequence int64  `json:"sequence"`
	Height   int64  `json:"height"`
}

// ReplyAddCallback Reply AddCallback
type ReplyAddCallback struct {
	Reply
	Seqs []*Sequence `json:"seqs"`
}

type callbackServerImpl struct {
}

func (c *callbackServerImpl) AddSeqCallBack(in *proto.BlockSeqCB, result *interface{}) error {
	if in.GetName() == "" {
		return fmt.Errorf("input must with name")
	}

	go tryToPushBlocks(in)

	*result = &ReplyAddCallback{
		Reply: Reply{
			IsOk: true,
			Msg:  "",
		},
	}

	return nil
}

func tryToPushBlocks(in *proto.BlockSeqCB) {
	time.Sleep(5 * time.Second) // 有一定的延时, 让对端服务准备好
	var client = &http.Client{}

	buildTestData()
	for i, seqs := range testSeqsArray {
		postdata, err := types.PBToJSON(seqs)
		if err != nil {
			panic("PBToJSON block failed")
		}
		err = postData(client, in, postdata, int64(i))
		if err != nil {
			panic("post block failed")
		}
	}
}

func postData(client *http.Client, cb *proto.BlockSeqCB, postdata []byte, seq int64) (err error) {
	//post data in body
	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	if _, err = g.Write(postdata); err != nil {
		log.Error("postData write", "cb.name", cb.Name, "err", err)
		return err
	}
	if err = g.Close(); err != nil {
		log.Error("postData write close", "cb.name", cb.Name, "err", err)
		return err
	}

	req, err := http.NewRequest("POST", cb.URL, &buf)
	if err != nil {
		log.Error("postData http request", "cb.name", cb.Name, "err", err, "url", cb.URL)
		return err
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Content-Encoding", "gzip")
	resp, err := client.Do(req)
	if err != nil {
		log.Error("postData http client.Do", "cb.name", cb.Name, "err", err, "url", cb.URL)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("postData ioutil.ReadAll", "cb.name", cb.Name, "err", err, "url", cb.URL)
		return err
	}
	if string(body) != "ok" && string(body) != "OK" {
		log.Error("postData fail", "cb.name", cb.Name, "body", string(body))
		return types.ErrPushSeqPostData
	}
	log.Debug("postData success", "cb.name", cb.Name, "updateSeq", seq)
	return nil
}

func main() {
	listener, err := net.Listen("tcp", cfg.address)
	if err != nil {
		log.Error("Listen", "err", err, "address", cfg.address)
		return
	}

	server := rpc.NewServer()
	err = server.RegisterName("Chain33", &callbackServerImpl{})
	if err != nil {
		log.Error("Listen", "err", err, "address", cfg.address)
		return
	}
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("JSONRPCServer", "RemoteAddr", r.RemoteAddr)

		if r.URL.Path != "/" {
			writeError(w, r, 0, fmt.Sprintf(`URL.PATH must be '/': %s`, r.URL.Path))
			return
		}

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			writeError(w, r, 0, "Can't get request body!")
			return
		}

		//格式做一个检查
		client, err := parseJSONRpcParams(data)
		if err != nil {
			err = fmt.Errorf(`invalid json request err:%s`, err.Error())
			log.Debug("JSONRPCServer", "request", string(data), "parseErr", err)
			writeError(w, r, 0, err.Error())
			return
		}

		if client.Method != "Chain33.AddSeqCallBack" {
			err = fmt.Errorf(`Only support Chain33.AddSeqCallBack:  %s called `, client.Method)
			log.Debug("JSONRPCServer", "request", string(data), "parseErr", err)
			writeError(w, r, 0, err.Error())
			return
		}

		serverCodec := jsonrpc.NewServerCodec(&HTTPConn{in: ioutil.NopCloser(bytes.NewReader(data)), out: w, r: r})
		w.Header().Set("Content-type", "application/json")
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
		}
		w.WriteHeader(200)
		err = server.ServeRequest(serverCodec)
		if err != nil {
			log.Debug("Error while serving JSON request: %v", err)
			return
		}
	})
	http.Serve(listener, handler)
}

// URL : /
// mothed: "Chain33.AddSeqCallBack"
// params:  types.BlockSeqCB{
//		Name:   name,
//		URL:    url,
//		Encode: encode,
//	}

type clientRequest struct {
	Method string         `json:"method"`
	Params [1]interface{} `json:"params"`
	ID     uint64         `json:"id"`
}

func parseJSONRpcParams(data []byte) (*clientRequest, error) {
	var req clientRequest
	err := json.Unmarshal(data, &req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// HTTPConn adapt HTTP connection to ReadWriteCloser
type HTTPConn struct {
	r   *http.Request
	in  io.Reader
	out io.Writer
}

// Read rewrite the read of http
func (c *HTTPConn) Read(p []byte) (n int, err error) { return c.in.Read(p) }

// Write rewrite the write of http
func (c *HTTPConn) Write(d []byte) (n int, err error) { //添加支持gzip 发送
	if strings.Contains(c.r.Header.Get("Accept-Encoding"), "gzip") {
		gw := gzip.NewWriter(c.out)
		defer gw.Close()
		return gw.Write(d)
	}
	return c.out.Write(d)
}

// Close rewrite the close of http
func (c *HTTPConn) Close() error { return nil }

type serverResponse struct {
	ID     uint64      `json:"id"`
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
}

func writeError(w http.ResponseWriter, r *http.Request, id uint64, errstr string) {
	w.Header().Set("Content-type", "application/json")
	//错误的请求也返回 200
	w.WriteHeader(200)
	resp, err := json.Marshal(&serverResponse{id, nil, errstr})
	if err != nil {
		log.Debug("json marshal error, nerver happen")
		return
	}
	_, err = w.Write(resp)
	if err != nil {
		log.Debug("Write", "err", err)
		return
	}
}
