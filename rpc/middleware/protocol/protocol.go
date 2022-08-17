package protocol

import (
	"encoding/json"

	"github.com/33cn/externaldb/escli/querypara"

	"net/http"

	l "github.com/33cn/chain33/common/log/log15"
	"github.com/gin-gonic/gin"
	rpcutils "github.com/33cn/externaldb/rpc/utils"
)

const (
	keyRequest = "jsonRequest"
	keyResult  = "objResult"
	keyError   = "serverError"
	keyCode    = "serverErrorCode"
)

var (
	log = l.New("module", "ginrpc")
)

func RequestParse(c *gin.Context) {
	var req rpcutils.ClientRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Error("bind json", "err", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, rpcutils.ServerResponse{
			Error: "bad request",
		})
		return
	}
	c.Set(keyRequest, req)
	c.Next()

	pcode, exist := c.Get(keyCode)
	if exist {
		perr, _ := c.Get(keyError)
		log.Error("Get error result", "err", perr)
		c.AbortWithStatusJSON(pcode.(int), rpcutils.ServerResponse{
			ID:    req.ID,
			Error: perr,
		})
		return
	}

	result, exist := c.Get(keyResult)
	if !exist {
		log.Error("Get result")
		c.AbortWithStatusJSON(http.StatusInternalServerError, rpcutils.ServerResponse{
			ID:    req.ID,
			Error: "implement error",
		})
		return
	}
	var resp rpcutils.ServerResponse
	resp.ID = req.ID
	resp.Result = result
	c.JSON(http.StatusOK, resp)
}

func GetRequest(c *gin.Context) (*rpcutils.ClientRequest, error) {
	v, exist := c.Get(keyRequest)
	if !exist {
		return nil, rpcutils.ErrBadParam
	}
	q := v.(rpcutils.ClientRequest)
	return &q, nil
}

func SetResult(c *gin.Context, result interface{}, err error) {
	if err != nil {
		c.Set(keyCode, http.StatusInternalServerError)
		c.Set(keyError, err.Error())
	}
	c.Set(keyResult, result)
}

func SetError(c *gin.Context, code int, err error) {
	c.Set(keyCode, code)
	if err != nil {
		c.Set(keyError, err.Error())
	}
}

//ParserESclient  解析es的query结构体
func ParserESclient(c *gin.Context) (*querypara.Query, error) {
	var q querypara.Query
	req, err := GetRequest(c)
	if err != nil {
		return nil, err
	}
	// 输入 -d {}, 即没有参数时的情况
	if req.Params[0] == nil {
		return &q, nil
	}
	err = json.Unmarshal([]byte(*req.Params[0]), &q)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// List 查询列表
func List(c *gin.Context, index, typ string, decoder func(x *json.RawMessage) (interface{}, error), db *rpcutils.DBRead) {
	q, err := ParserESclient(c)
	if err != nil {
		SetError(c, http.StatusBadRequest, err)
		return
	}
	r, err := db.List(index, typ, q, decoder)
	SetResult(c, r, err)
}

// Count 统计数量
func Count(c *gin.Context, index, typ string, db *rpcutils.DBRead) {
	q, err := ParserESclient(c)
	if err != nil {
		SetError(c, http.StatusBadRequest, err)
		return
	}
	count, err := db.Count(index, typ, q)
	SetResult(c, count, err)
}
