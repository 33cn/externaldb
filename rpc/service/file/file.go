package file

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"path/filepath"
	"time"

	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/rpc/grpcclient"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db"
	fdb "github.com/33cn/externaldb/db/file/db"
	fpdb "github.com/33cn/externaldb/db/filepart/db"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/proto"
	rpcutils "github.com/33cn/externaldb/rpc/utils"
	"github.com/gin-gonic/gin"
)

var log = l.New("module", "file")
var ErrFileNotBlock = errors.New("文件同步中")

// File handler
type File struct {
	*rpcutils.DBRead
	cfg *proto.ConfigNew
}

// InitRouter 初始化proofrpc接口的router路由表
func InitRouter(router gin.IRouter, dbread *rpcutils.DBRead, cfg *proto.ConfigNew) {
	f := File{DBRead: dbread, cfg: cfg}

	v1 := router.Group("/v1")
	v1.GET("/file/*name", f.GetFile)
	v1.GET("/file-clean-cache", f.CleanCache)
}

func (f *File) NewESShortConnect() (escli.ESClient, error) {
	return escli.NewESShortConnect(f.Host, f.Prefix, f.Version, f.Username, f.Password)
}

func (f *File) NewChainGRPCConnect() (types.Chain33Client, error) {
	return grpcclient.NewMainChainClient(types.NewChain33Config(types.GetDefaultCfgstring()), f.cfg.Chain.GetGrpcHost())
}

// GetFile 获取上链文件
// @Summary 获取上链文件
// @Description get file
// @Tags File
// @Produce json
// @Param hash query string true "文件哈希"
// @Param name path string true "文件名称"
// @Success 200 {array} byte
// @Failure 500 {string} string
// @Router /v1/file/{hash} [get]
func (f *File) GetFile(c *gin.Context) {
	ec, err := f.NewESShortConnect()
	if err != nil {
		log.Error("GetFile:NewESShortConnect ", "err", err)
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	gc, err := f.NewChainGRPCConnect()
	if err != nil {
		log.Error("GetChainFile:NewChainConnect ", "err", err)
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	hash := c.Query("hash")
	name := c.Param("name")
	if hash == "" {
		hash = name
	}
	cli := fdb.NewEsGrpcDB(ec, gc)
	hash = filepath.Base(hash)
	file, err := cli.Get(hash)
	if err != nil {
		if err == db.ErrDBNotFound {
			err = ErrFileNotBlock
		}
		log.Error("GetFileFromDB:fcli.Get(hash)", "err", err)
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	f.ServerFile(c, file, name)
}

func (f *File) ServerFile(c *gin.Context, file *fdb.File, name string) {
	content := file.Data
	data, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		log.Error("GetFile:base64.DecodeString", "err", err)
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	if name == "" {
		name = file.FileHash
	}
	http.ServeContent(c.Writer, c.Request, name, time.Now(), bytes.NewReader(data))
}

// CleanCache 清理文件缓存
// @Summary 清理文件缓存
// @Description clean file cache
// @Tags File
// @Produce json
// @Success 200 {string} string "ok"
// @Failure 500 {string} string
// @Router /v1/file-clean-cache [get]
func (f *File) CleanCache(c *gin.Context) {
	ec, err := f.NewESShortConnect()
	if err != nil {
		log.Error("GetFile:NewESShortConnect ", "err", err)
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	err = fpdb.NewEsDB(ec).Clean()
	if err != nil {
		log.Error("GetFile:NewESShortConnect ", "err", err)
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Error(c.Writer, "ok", http.StatusOK)
}
