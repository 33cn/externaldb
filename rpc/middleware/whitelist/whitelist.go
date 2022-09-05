package whitelist

import (
	"net/http"
	"strings"

	l "github.com/33cn/chain33/common/log/log15"
	rpcutils "github.com/33cn/externaldb/rpc/utils"
	"github.com/gin-gonic/gin"
)

var (
	log = l.New("module", "middleware.whitelist")
)

type WhiteList struct {
	whitelist map[string]bool
}

func New(addresses []string) *WhiteList {
	w := WhiteList{
		whitelist: make(map[string]bool),
	}
	if len(addresses) == 1 && addresses[0] == "*" {
		w.whitelist["0.0.0.0"] = true
		return &w
	}

	for _, addr := range addresses {
		log.Debug("initWhitelist", "addr", addr)
		w.whitelist[addr] = true
	}
	return &w
}

func (w *WhiteList) GinHandler(c *gin.Context) {
	ip := strings.Split(c.Request.RemoteAddr, ":")[0]
	ok := w.Check(ip)
	if !ok {
		log.Error("checkWhiteList", "addr", ip, "org-addr", c.Request.RemoteAddr)
		c.AbortWithStatusJSON(http.StatusForbidden, rpcutils.ServerResponse{
			Error: "address not in whitelist",
		})
	}
}

func (w *WhiteList) Check(addr string) bool {
	return checkWhitlist(addr, w.whitelist)
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
