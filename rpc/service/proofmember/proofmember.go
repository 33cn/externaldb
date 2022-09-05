// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proofmember

import (
	"encoding/json"
	"net/http"

	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/querypara"

	proofconfig "github.com/33cn/externaldb/db/proof_config"
	"github.com/33cn/externaldb/rpc/middleware/protocol"
	rpcutils "github.com/33cn/externaldb/rpc/utils"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/xuperchain/log15"
)

// InitRouter 初始化proofrpc接口的router路由表
func InitRouter(router gin.IRouter, db *rpcutils.DBRead) {
	m := ProofMember{DBRead: db}
	o := ProofOrganization{DBRead: db}

	v1 := router.Group("/v1")
	v1.POST("/proofmember/List", m.GinList)
	v1.POST("/proofmember/Count", m.GinCount)
	v1.POST("/proofmember/Gets", m.GinGets)
	v1.POST("/prooforganization/List", o.GinList)
	v1.POST("/prooforganization/Count", o.GinCount)
	v1.POST("/prooforganization/Gets", o.GinGets)
}

func decodeProofMember(x *json.RawMessage) (interface{}, error) {
	var tempx proofconfig.Member
	err := json.Unmarshal(*x, &tempx)
	if err != nil {
		log.Info("decodeProof-Unmarshal", "err", err)
		return *x, err
	}
	return tempx, nil
}

func decodeProofOrganization(x *json.RawMessage) (interface{}, error) {
	var tempx proofconfig.Organization
	err := json.Unmarshal(*x, &tempx)
	if err != nil {
		log.Info("decodeProof-Unmarshal", "err", err)
		return *x, err
	}
	return tempx, nil
}

// ProofMember Proof Member
type ProofMember struct {
	*rpcutils.DBRead
}

// GinList 参数: 分页, organization, 输出: 列表)
// @Summary 分页列出指定范围的用户
// @Description get proofmember by organization
// @Tags proofmember
// @Router /v1/proofmember/List [post]
// @Accept json
// @Param input body swagger.ClientRequest{params=[]swagger.Query} true "INPUT"
// @Produce json
// @Success 200 {object} swagger.ServerResponse{result=[]swagger.Member}
// @Failure 400 {object} swagger.ServerResponse{error=string}
func (t *ProofMember) GinList(c *gin.Context) {
	q, err := protocol.ParserESclient(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	var out interface{}
	err = t.List(q, &out)
	protocol.SetResult(c, out, err)
}

// GinCount 获得指定范围的用户的数量
// @Summary 获得指定范围的用户的数量
// @Description count proofmember by organization
// @Tags proofmember
// @Router /v1/proofmember/Count [post]
// @Accept json
// @Param input body swagger.ClientRequest{params=[]swagger.Query} true "INPUT"
// @Produce json
// @Success 200 {object} swagger.ServerResponse{result=int64}
// @Failure 400 {object} swagger.ServerResponse{error=string}
func (t *ProofMember) GinCount(c *gin.Context) {
	q, err := protocol.ParserESclient(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	var out interface{}
	err = t.Count(q, &out)
	protocol.SetResult(c, out, err)
}

// GinGets 获得指定地址的用户
// @Summary 获得指定地址的用户
// @Description proofmember by addresses
// @Tags proofmember
// @Router /v1/proofmember/Gets [post]
// @Accept json
// @Param input body swagger.ClientRequest{params=[]rpcutils.Addresses} true "INPUT"
// @Produce json
// @Success 200 {object} swagger.ServerResponse{result=[]swagger.Member}
// @Failure 400 {object} swagger.ServerResponse{error=string}
func (t *ProofMember) GinGets(c *gin.Context) {
	q, err := protocol.GetRequest(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}
	var req rpcutils.Addresses
	err = json.Unmarshal(*q.Params[0], &req)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	var out interface{}
	err = t.Gets(&req, &out)
	protocol.SetResult(c, out, err)
}

// List 参数: 分页, organization, 输出: 列表
func (t *ProofMember) List(q *querypara.Query, out *interface{}) error {
	if q == nil {
		return errors.Wrapf(rpcutils.ErrBadParam, "empty queryPara input")
	}
	// 默认分页一次取10个数据
	if q.Page == nil {
		q.Page = &querypara.QPage{
			Number: 1,
			Size:   10,
		}
	}

	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}

	r, err := cli.Search(proofconfig.DBX, proofconfig.TableX, q, decodeProofMember)
	if err != nil || r == nil {
		return err
	}
	*out = r
	return nil
}

// Count 参数: organization, 输出: N
func (t *ProofMember) Count(q *querypara.Query, out *interface{}) error {
	if q == nil {
		return errors.Wrapf(rpcutils.ErrBadParam, "empty queryPara input")
	}

	var err error
	*out, err = t.DBRead.Count(proofconfig.DBX, proofconfig.TableX, q)
	return err
}

// Members rpc return
type Members struct {
	Member proofconfig.Member `json:"member"`
}

// Gets gets
func (t *ProofMember) Gets(req *rpcutils.Addresses, out *interface{}) error {
	if req == nil || len(req.Address) == 0 {
		log.Debug("1")
		return nil
	}
	ids := make([]string, 0)
	for _, hash := range req.Address {
		ids = append(ids, proofconfig.MemberID(hash))
	}

	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}
	resp, err := cli.MGet(proofconfig.DBX, proofconfig.TableX, ids, decodeProofMember)
	if err != nil {
		return err
	}
	*out = resp

	return nil
}

// ProofOrganization Proof Organization
type ProofOrganization struct {
	*rpcutils.DBRead
}

// GinList 分页列出指定范围的组织
// @Summary 分页列出指定范围的组织
// @Description get proof organization
// @Tags prooforganization
// @Router /v1/prooforganization/List [post]
// @Accept json
// @Param input body swagger.ClientRequest{params=[]swagger.Query} true "INPUT"
// @Produce json
// @Success 200 {object} swagger.ServerResponse{result=[]swagger.Organization}
// @Failure 400 {object} swagger.ServerResponse{error=string}
func (t *ProofOrganization) GinList(c *gin.Context) {
	q, err := protocol.ParserESclient(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	var out interface{}
	err = t.List(q, &out)
	protocol.SetResult(c, out, err)
}

// GinCount 获得指定范围的组织的数量
// @Summary 获得指定范围的组织的数量
// @Description get proof organization count
// @Tags prooforganization
// @Router /v1/prooforganization/Count [post]
// @Accept json
// @Param input body swagger.ClientRequest{params=[]swagger.Query} true "INPUT"
// @Produce json
// @Success 200 {object} swagger.ServerResponse{result=int64}
// @Failure 400 {object} swagger.ServerResponse{error=string}
func (t *ProofOrganization) GinCount(c *gin.Context) {
	q, err := protocol.ParserESclient(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	var out interface{}
	err = t.Count(q, &out)
	protocol.SetResult(c, out, err)
}

// GinGets 获得指定的组织的信息
// @Summary 获得指定的组织的信息
// @Description get proof organization info
// @Tags prooforganization
// @Router /v1/prooforganization/Gets [post]
// @Accept json
// @Param input body swagger.ClientRequest{params=[]rpcutils.Organizations} true "INPUT"
// @Produce json
// @Success 200 {object} swagger.ServerResponse{result=[]swagger.Organization}
// @Failure 400 {object} swagger.ServerResponse{error=string}
func (t *ProofOrganization) GinGets(c *gin.Context) {
	q, err := protocol.GetRequest(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}
	var req rpcutils.Organizations
	err = json.Unmarshal(*q.Params[0], &req)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	var out interface{}
	err = t.Gets(&req, &out)
	protocol.SetResult(c, out, err)
}

// List 参数: 分页
func (t *ProofOrganization) List(q *querypara.Query, out *interface{}) error {
	if q == nil {
		return errors.Wrapf(rpcutils.ErrBadParam, "empty queryPara input")
	}
	// 默认分页一次取10个数据
	if q.Page == nil {
		q.Page = &querypara.QPage{
			Number: 1,
			Size:   10,
		}
	}

	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}

	r, err := cli.Search(proofconfig.OrgDBX, proofconfig.OrgTableX, q, decodeProofOrganization)
	if err != nil || r == nil {
		return err
	}
	*out = r
	return nil
}

// Count 参数: Nil 输出: 列表
func (t *ProofOrganization) Count(q *querypara.Query, out *interface{}) error {
	if q == nil {
		return errors.Wrapf(rpcutils.ErrBadParam, "empty queryPara input")
	}

	var err error
	*out, err = t.DBRead.Count(proofconfig.OrgDBX, proofconfig.OrgTableX, q)
	return err
}

// Gets gets by id
func (t *ProofOrganization) Gets(req *rpcutils.Organizations, out *interface{}) error {
	if req == nil || len(req.Organization) == 0 {
		log.Debug("1")
		return nil
	}
	ids := make([]string, 0)
	for _, org := range req.Organization {
		ids = append(ids, proofconfig.OrganizationID(org))
	}

	cli, err := escli.NewESShortConnect(t.Host, t.Prefix, t.Version, t.Username, t.Password)
	if err != nil {
		return err
	}
	resp, err := cli.MGet(proofconfig.OrgDBX, proofconfig.OrgTableX, ids, decodeProofOrganization)
	if err != nil {
		return err
	}
	*out = resp

	return nil
}
