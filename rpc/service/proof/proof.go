// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proof

import (
	"encoding/json"
	"net/http"
	"time"

	l "github.com/33cn/chain33/common/log/log15"
	"github.com/gin-gonic/gin"

	"github.com/33cn/externaldb/db/proof/proofdb"
	"github.com/33cn/externaldb/db/proof/service"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/querypara"
	"github.com/33cn/externaldb/rpc/middleware/protocol"
	rpcutils "github.com/33cn/externaldb/rpc/utils"
)

var (
	log             = l.New("module", "proof")
	StatisDonation  = "Donation"
	StatisVolunteer = "Volunteer"
	StatisItem      = 10
)

func decodeProof(x *json.RawMessage) (interface{}, error) {
	tempx := make(map[string]interface{})
	err := json.Unmarshal([]byte(*x), &tempx)
	if err != nil {
		log.Info("decodeProof-Unmarshal", "err", err)
		return *x, err
	}
	return tempx, nil
}

// InitRouter 初始化proofrpc接口的router路由表
func InitRouter(router gin.IRouter, dbread *rpcutils.DBRead) {
	pr := Proof{DBRead: dbread}

	pr.statCache = &rpcutils.BaseStatsItem{MaxStatsItems: rpcutils.MaxStatsItems}

	v1 := router.Group("/v1")
	v1.POST("/proof/List", pr.ListProof)
	v1.POST("/proof/ListUpdateProof", pr.ListUpdateProof)
	v1.POST("/proof/ListUpdateRecord", pr.ListUpdateRecord)
	v1.POST("/proof/Count", pr.CountProof)
	v1.POST("/proof/CountByTime", pr.CountByTime)
	v1.POST("/proof/Show", pr.ShowProof)
	v1.POST("/proof/Gets", pr.Gets)
	v1.POST("/proof/GetProofs", pr.GetProofs)
	v1.POST("/proof/GetTemplates", pr.GetTemplates)
	v1.POST("/proof/FetchSource", pr.FetchSource)
	v1.POST("/proof/VolunteerStats", pr.VolunteerStats)
	v1.POST("/proof/DonationStats", pr.DonationStats)
	v1.POST("/proof/TotalStats", pr.TotalStats)
	v1.POST("/proof/QueryStatsInfo", pr.QueryStatsInfo)
}

// Proof Proof
type Proof struct {
	*rpcutils.DBRead
	statCache *rpcutils.BaseStatsItem
}

// ListProof 获取存证列表
// @Summary 获取存证列表
// @Description list proof of organization/sender
// @Tags Proof
// @Produce json
// @Param input body swagger.ClientRequest{params=[]swagger.Query} true "INPUT"
// @Success 200 {object} swagger.ListProofResult
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/proof/List [post]
func (p *Proof) ListProof(c *gin.Context) {
	protocol.List(c, proofdb.ProofDBX, proofdb.ProofTableX, decodeProof, p.DBRead)
}

// ListUpdateProof 获取最新存证列表
// @Summary 获取最新存证列表
// @Description list proof of organization/sender
// @Tags Proof
// @Produce json
// @Param input body swagger.ClientRequest{params=[]swagger.Query} true "INPUT"
// @Success 200 {object} swagger.ListProofResult
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/proof/ListUpdateProof [post]
func (p *Proof) ListUpdateProof(c *gin.Context) {
	var err error
	reselt := make([]interface{}, 0)
	q, err := protocol.ParserESclient(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}
	// 默认分页一次取10个数据
	if q.Page == nil {
		q.Page = &querypara.QPage{
			Number: 1,
			Size:   10,
		}
	}
	// 查询原始存证：update_hash=="null"或者不存在update_hash
	matchOne := []*querypara.QMatch{
		{Key: "update_hash", Value: "null"},
		{SubQuery: &querypara.Query{
			Not: []*querypara.QMatch{
				{Key: "update_hash"},
			},
		}},
	}
	match := &querypara.QMatch{
		SubQuery: &querypara.Query{
			MatchOne: matchOne,
		},
	}
	q.Match = append(q.Match, match)

	cli, err := escli.NewESShortConnect(p.Host, p.Prefix, p.Version, p.Username, p.Password)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}

	//查询增量存证
	var majors []interface{}
	majors, err = cli.Search(proofdb.ProofDBX, proofdb.ProofTableX, q, decodeProof)
	if err != nil || majors == nil {
		//特殊处理一下没有数据时sort功能报错的场景,取消sort再尝试调用一次
		if q.Sort != nil {
			q.Sort = nil
			majors, err = cli.Search(proofdb.ProofDBX, proofdb.ProofTableX, q, decodeProof)
		} else {
			protocol.SetError(c, http.StatusInternalServerError, err)
			return
		}
	}

	//查询到update proof的最新版本，也就是最新状态
	for _, v := range majors {
		proof := v.(map[string]interface{})
		if _, ok := proof["update_version"]; !ok || proof["update_version"].(float64) == 0 {
			//没有设置更新存证
			reselt = append(reselt, proof)
		} else {
			q2 := &querypara.Query{
				Match: []*querypara.QMatch{
					{Key: "update_hash", Value: proof["proof_tx_hash"]},
					{Key: "update_version", Value: proof["update_version"]},
				},
			}
			// 查询最新的存证
			new, err := cli.Search(proofdb.ProofDBX, proofdb.ProofTableX, q2, decodeProof)
			if err != nil || new == nil {
				protocol.SetError(c, http.StatusInternalServerError, err)
				return
			}
			reselt = append(reselt, new[0])
		}
	}

	protocol.SetResult(c, reselt, err)
}

// ListUpdateRecord 获取存证更新记录的列表
// @Summary 获取存证更新记录的列表
// @Description list update proof record of organization/sender
// @Tags Proof
// @Produce json
// @Param input body swagger.ClientRequest{params=[]swagger.Query} true "INPUT"
// @Success 200 {object} swagger.ListProofResult
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/proof/ListUpdateRecord [post]
func (p *Proof) ListUpdateRecord(c *gin.Context) {
	q, err := protocol.ParserESclient(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}
	// 默认分页一次取10个数据
	if q.Page == nil {
		q.Page = &querypara.QPage{
			Number: 1,
			Size:   10,
		}
	}

	cli, err := escli.NewESShortConnect(p.Host, p.Prefix, p.Version, p.Username, p.Password)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}

	r, err := cli.Search(proofdb.ProofUpdateDBX, proofdb.ProofUpdateTableX, q, decodeProof)
	if err != nil || r == nil {
		//特殊处理一下没有数据时sort功能报错的场景,取消sort再尝试调用一次
		if q.Sort != nil {
			q.Sort = nil
			r2, err2 := cli.Search(proofdb.ProofUpdateDBX, proofdb.ProofUpdateTableX, q, decodeProof)
			if err2 == nil && r2 != nil {
				protocol.SetResult(c, r2, err2)
				return
			}
		}

		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}

	protocol.SetResult(c, r, err)
}

// CountProof 获取存证数量
// @Summary 获取存证数量
// @Description get proof count
// @Tags Proof
// @Produce json
// @Param input body swagger.ClientRequest{params=[]swagger.Query} true "INPUT"
// @Success 200 {object} swagger.ServerResponse{result=int64}
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/proof/Count [post]
func (p *Proof) CountProof(c *gin.Context) {
	protocol.Count(c, proofdb.ProofDBX, proofdb.ProofTableX, p.DBRead)
}

// CountByTime 根据年/月/日对存证的数量进行统计
// @Summary 根据年/月/日对存证的数量进行统计
// @Description get proof count with time
// @Tags Proof
// @Produce json
// @Param input body swagger.ClientRequest{params=[]rpcutils.CountByTime} true "INPUT"
// @Success 200 {object} swagger.ServerResponse // todo
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/proof/CountByTime [post]
func (p *Proof) CountByTime(c *gin.Context) {
	q, err := protocol.GetRequest(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	var req rpcutils.CountByTime
	err = json.Unmarshal(*q.Params[0], &req)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}

	cli, _ := escli.NewESShortConnect(p.Host, p.Prefix, p.Version, p.Username, p.Password)

	var matches []*querypara.QMatch
	for _, match := range req.Match {
		matches = append(matches, &querypara.QMatch{Key: match.Key, Value: match.Value})
	}
	var ranges []*querypara.ARange
	for _, r := range req.Ranges.Ranges {
		ranges = append(ranges, &querypara.ARange{RStart: r.RStart, REnd: r.REnd})
	}

	query := querypara.Query{
		Match: matches,
	}
	agg := querypara.Agg{
		Name: "count",
		Ranges: &querypara.ARanges{
			Key:    req.Ranges.Key,
			Ranges: ranges,
		},
	}

	search := querypara.Search{
		Query: &query,
		Agg:   &agg,
	}

	result, err := cli.Agg(proofdb.ProofDBX, proofdb.ProofTableX, &search)
	r, found := result.Range("count")
	if !found {
		log.Error("CountByTime not found")
	}

	protocol.SetResult(c, r, err)
}

// ShowProof 获得指定hash的存证信息
// @Summary 获得指定hash的存证信息
// @Description get proof by txhash
// @Tags Proof
// @Produce json
// @Param input body swagger.ClientRequest{params=[]swagger.Query} true "INPUT"
// @Success 200 {object} swagger.ListProofResult
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/proof/Show [post]
func (p *Proof) ShowProof(c *gin.Context) {
	q, err := protocol.ParserESclient(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	cli, err := escli.NewESShortConnect(p.Host, p.Prefix, p.Version, p.Username, p.Password)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}
	r, err := cli.Search(proofdb.ProofDBX, proofdb.ProofTableX, q, decodeProof)
	if err != nil || r == nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}
	protocol.SetResult(c, r, err)
}

// Gets 获取多个指定hash的存证信息
// Deprecated: Use Proof.GetProofs instead.
// @Summary 获取多个指定hash的存证信息
// @Description get proof by hashes
// @Tags Proof
// @Produce json
// @Param input body swagger.ClientRequest{params=[]rpcutils.Hashes} true "INPUT"
// @Success 200 {object} swagger.ListProofResult
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/proof/Gets [post]
func (p *Proof) Gets(c *gin.Context) {
	q, err := protocol.GetRequest(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	var req rpcutils.Hashes
	err = json.Unmarshal(*q.Params[0], &req)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}

	if len(req.Hash) == 0 {
		protocol.SetError(c, http.StatusBadRequest, rpcutils.ErrBadParam)
		return
	}

	ids := make([]string, 0)
	for _, hash := range req.Hash {
		ids = append(ids, service.ProofID(hash))
	}

	cli, err := escli.NewESShortConnect(p.Host, p.Prefix, p.Version, p.Username, p.Password)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}
	resp, err := cli.MGet(proofdb.ProofDBX, proofdb.ProofTableX, ids, decodeProof)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}
	protocol.SetResult(c, resp, nil)
}

// GetProofs 获取多个指定hash的存证信息
// @Summary 获取多个指定hash的存证信息
// @Description get proof by hashes
// @Tags Proof
// @Produce json
// @Param input body swagger.ClientRequest{params=[]rpcutils.Hashes} true "INPUT"
// @Success 200 {object} swagger.ListProofResult
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/proof/GetProofs [post]
func (p *Proof) GetProofs(c *gin.Context) {
	q, err := protocol.GetRequest(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	var req rpcutils.Hashes
	err = json.Unmarshal([]byte(*q.Params[0]), &req)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}

	if len(req.Hash) == 0 {
		protocol.SetError(c, http.StatusBadRequest, rpcutils.ErrBadParam)
		return
	}

	ids := make([]string, 0)
	for _, hash := range req.Hash {
		ids = append(ids, service.ProofID(hash))
	}

	cli, err := escli.NewESShortConnect(p.Host, p.Prefix, p.Version, p.Username, p.Password)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}
	resp, err := cli.MGet(proofdb.ProofDBX, proofdb.ProofTableX, ids, decodeProof)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}
	protocol.SetResult(c, resp, nil)
}

// GetTemplates 获取多个指定hash的存证模板
// @Summary 获取多个指定hash的存证模板
// @Description get proof template by hashes
// @Tags Proof
// @Produce json
// @Param input body swagger.ClientRequest{params=[]rpcutils.Hashes} true "INPUT"
// @Success 200 {object} swagger.ServerResponse{result=[]swagger.Template}
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/proof/GetTemplates [post]
func (p *Proof) GetTemplates(c *gin.Context) {
	q, err := protocol.GetRequest(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	var req rpcutils.Hashes
	err = json.Unmarshal([]byte(*q.Params[0]), &req)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}

	if len(req.Hash) == 0 {
		protocol.SetError(c, http.StatusBadRequest, rpcutils.ErrBadParam)
		return
	}

	ids := make([]string, 0)
	for _, hash := range req.Hash {
		ids = append(ids, service.TemplateID(hash))
	}

	cli, err := escli.NewESShortConnect(p.Host, p.Prefix, p.Version, p.Username, p.Password)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}
	resp, err := cli.MGet(proofdb.TemplateDBX, proofdb.TemplateTableX, ids, decodeProof)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}
	protocol.SetResult(c, resp, nil)
}

// VolunteerStats 获取志愿者的分布图按照省/单位
// @Summary 获取志愿者的分布图按照省/单位
// @Description get volunteer statistics
// @Tags Proof
// @Produce json
// @Param input body swagger.ClientRequest{params=[]rpcutils.VolunteerStats} true "INPUT"
// @Success 200 {object} swagger.VolunteerStatsResult
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/proof/VolunteerStats [post]
func (p *Proof) VolunteerStats(c *gin.Context) {
	q, err := protocol.GetRequest(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	var req rpcutils.VolunteerStats
	err = json.Unmarshal([]byte(*q.Params[0]), &req)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}

	//入参校验
	if &req == nil || req.Match == nil || req.TermsAgg == nil || req.SubTermsAgg == nil || req.SubSumAgg == nil {
		protocol.SetError(c, http.StatusInternalServerError, rpcutils.ErrBadParam)
		return
	}
	if p.statCache != nil && len(p.statCache.ChildStatsItems) != 0 {
		for _, stat := range p.statCache.ChildStatsItems {
			if stat != nil && stat.IsMatch(rpcutils.VolunteerFlag, req) {
				protocol.SetResult(c, stat.GetRep(), nil)
				return
			}
		}
	}

	//第一次查询需要将此统计信息添加到statCache中
	rep, err := p.volunteerStats(req)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}
	lastSeq := p.LastConvertSeq()

	newStat := &rpcutils.VolunteerStatsItem{}
	newStat.StatsComm = rpcutils.NewStatsComm(req, rep, 300, lastSeq, p.volunteerStats)

	p.statCache.AddStatsItem(newStat)
	log.Debug("VolunteerStats:add statinfo:", "Req.Match:", newStat.PrintStatsItemJSON(), "lastSeq:", lastSeq, "Interval:", newStat.GetInterval())

	p.refreshStat(newStat)
	protocol.SetResult(c, rep, nil)

}

// volunteerStats  志愿者统计
func (p *Proof) volunteerStats(reqinfo interface{}) (interface{}, error) {
	req := reqinfo.(rpcutils.VolunteerStats)

	cli, err := escli.NewESShortConnect(p.Host, p.Prefix, p.Version, p.Username, p.Password)
	if err != nil {
		return nil, err
	}

	var q querypara.Query
	for _, match := range req.Match {
		q.Filter = append(q.Filter, &querypara.QMatch{Key: match.Key, Value: match.Value})
	}

	firstField := req.TermsAgg.Key + ".keyword"

	secondField := req.SubTermsAgg.Key + ".keyword"
	subSumAggKey := req.SubSumAgg.Key

	sub := &querypara.Agg{
		Name: "subagg",
		Size: &querypara.ASize{Size: 100000},
		Term: &querypara.AAgg{Key: secondField},
		Subs: &querypara.ASub{
			Sum: []*querypara.AAgg{{Name: "totalSum", Key: subSumAggKey}},
			Pipeline: &querypara.APipe{
				Order: &querypara.AOrder{Name: "totalOrder", Key: "totalSum", Asc: false},
			},
		},
	}

	a := &querypara.Agg{
		Name: "agg",
		Size: &querypara.ASize{Size: 100000},
		Term: &querypara.AAgg{Key: firstField, CollectionMode: "breadth_first"},
		Subs: &querypara.ASub{SubAgg: sub},
	}
	s := &querypara.Search{
		Query: &q,
		Agg:   a,
	}

	result, err := cli.Agg(proofdb.ProofTableX, proofdb.ProofTableX, s)
	if err != nil {
		log.Error("VolunteerStatistics", "err", err.Error())
		return nil, err
	}
	results, found := result.Term("agg")
	if !found {
		log.Error("VolunteerStatistics", "load total failed", "not found")
		return nil, rpcutils.ErrNotFound
	}

	volunteerStates := rpcutils.RepVolunteerStat{}
	var total int64
	//TermsAgg
	var termsAgges []*rpcutils.RepTermsAgg
	for _, b := range results {

		highagg, found := b.Term("subagg")
		if !found {
			log.Error("VolunteerStatistics.high", "load total failed", "not found")
			continue
		}

		//SubTermsAgg
		var subTermsAgges []*rpcutils.RepSubTermsAgg
		var pCount int64
		for _, highbucket := range highagg {
			//acli2, _ := escli.NewAggDecode(p.Version, highbucket)
			//sumagg, found := acli2.Sum("totalSum")
			sumagg, found := highbucket.Sum("totalSum")
			if !found {
				log.Error("VolunteerStatistics.Sum", "load total sum failed", "not found")
				continue
			}
			total += int64(sumagg)
			pCount += int64(sumagg)

			subTermsAgg := rpcutils.RepSubTermsAgg{}
			subTermsAgg.SubTermsAggKey = highbucket.Key.(string)
			subTermsAgg.Count = int64(sumagg)

			subTermsAgges = append(subTermsAgges, &subTermsAgg)
		}
		termsAgg := rpcutils.RepTermsAgg{}
		termsAgg.Count = pCount
		termsAgg.TermsAggKey = b.Key.(string)
		termsAgg.SubTermsAgges = append(termsAgg.SubTermsAgges, subTermsAgges...)

		termsAgges = append(termsAgges, &termsAgg)
	}
	volunteerStates.Count = total
	volunteerStates.TermsAgges = append(volunteerStates.TermsAgges, termsAgges...)
	return volunteerStates, nil
	//protocol.SetResult(c, volunteerStates, nil)
}

func decodeFetchSource(x *json.RawMessage) (interface{}, error) {
	return string(*x), nil
}

// FetchSource 获取满足条件的数据的指定字段的值
// @Summary 获取满足条件的数据的指定字段的值
// @Description get specified fields of match
// @Tags Proof
// @Produce json
// @Param input body swagger.ClientRequest{params=[]rpcutils.SpecifiedFields} true "INPUT"
// @Success 200 {object} swagger.ServerResponse{result=[]string}
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/proof/FetchSource [post]
func (p *Proof) FetchSource(c *gin.Context) {
	q, err := protocol.GetRequest(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	var req rpcutils.SpecifiedFields
	err = json.Unmarshal([]byte(*q.Params[0]), &req)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}

	cli, err := escli.NewESShortConnect(p.Host, p.Prefix, p.Version, p.Username, p.Password)
	if err != nil {
		log.Error("FetchSource:Elasticsearch connect", "err", err)
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}

	var query querypara.Query
	for _, match := range req.Match {
		query.Filter = append(query.Filter, &querypara.QMatch{Key: match.Key, Value: match.Value})
	}
	query.Fetch = &querypara.QFetch{FetchSource: true, Keys: req.Fields}
	query.Size = &querypara.QSize{Size: req.Count}

	if req.Sort != nil && len(req.Sort) > 0 {
		for _, s := range req.Sort {
			query.Sort = append(query.Sort, &querypara.QSort{Key: s.Key, Ascending: s.Ascending})
		}
	}

	result, err := cli.Search(proofdb.ProofTableX, proofdb.ProofTableX, &query, decodeFetchSource)
	if err != nil {
		log.Error("VolunteerStatistics", "err", err.Error())
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}

	protocol.SetResult(c, result, nil)
}

// TotalStats 获取满足条件的数据的指定字段的总值
// @Summary 获取满足条件的数据的指定字段的总值
// @Description get sum of match
// @Tags Proof
// @Produce json
// @Param input body swagger.ClientRequest{params=[]rpcutils.TotalStats} true "INPUT"
// @Success 200 {object} rpcutils.ServerResponse{result=float64}
// @Failure 400 {object} rpcutils.ServerResponse
// @Router /v1/proof/TotalStats [post]
func (p *Proof) TotalStats(c *gin.Context) {
	q, err := protocol.GetRequest(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	var req rpcutils.TotalStats
	err = json.Unmarshal([]byte(*q.Params[0]), &req)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}
	if req.SumAgg == nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	cli, err := escli.NewESShortConnect(p.Host, p.Prefix, p.Version, p.Username, p.Password)
	if err != nil {
		log.Error("TotalStats:Elasticsearch connect", "err", err)
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}

	var query querypara.Query
	for _, match := range req.Match {
		query.Filter = append(query.Filter, &querypara.QMatch{Key: match.Key, Value: match.Value})
	}
	a := &querypara.Agg{
		Name: "sumstats",
		Metric: &querypara.AMetric{
			Sum: &querypara.AAgg{Key: req.SumAgg.Key},
		},
	}
	s := &querypara.Search{
		Query: &query,
		Agg:   a,
	}

	result, err := cli.Agg(proofdb.ProofTableX, proofdb.ProofTableX, s)
	if err != nil {
		log.Error("TotalStats", "err", err.Error())
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}

	sumagg, found := result.Sum("sumstats")
	if !found {
		log.Error("TotalStats", "load stat terms failed", "not found")
		protocol.SetError(c, http.StatusInternalServerError, rpcutils.ErrNotFound)
		return
	}
	protocol.SetResult(c, sumagg, nil)
}

// DonationStats 获取捐款排名信息
// @Summary 获取捐款排名信息
// @Description get donation stats
// @Tags Proof
// @Produce json
// @Param input body swagger.ClientRequest{params=[]rpcutils.DonationStats} true "INPUT"
// @Success 200 {object} swagger.ServerResponse{result=swagger.DonationStats}
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/proof/DonationStats [post]
func (p *Proof) DonationStats(c *gin.Context) {
	q, err := protocol.GetRequest(c)
	if err != nil {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	var req rpcutils.DonationStats
	err = json.Unmarshal([]byte(*q.Params[0]), &req)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}

	if req.TermsAgg == nil || req.SubSumAgg == nil || len(req.Match) == 0 {
		protocol.SetError(c, http.StatusBadRequest, err)
		return
	}

	if p.statCache != nil && len(p.statCache.ChildStatsItems) != 0 {
		for _, stat := range p.statCache.ChildStatsItems {
			if stat != nil && stat.IsMatch(rpcutils.DonateFlag, req) {
				protocol.SetResult(c, stat.GetRep(), nil)
				return
			}
		}
	}

	//第一次查询需要将此统计信息添加到statCache中
	rep, err := p.donationStats(req)
	if err != nil {
		protocol.SetError(c, http.StatusInternalServerError, err)
		return
	}
	lastSeq := p.LastConvertSeq()

	newStat := &rpcutils.DonateStatsItem{}
	newStat.StatsComm = rpcutils.NewStatsComm(req, rep, 300, lastSeq, p.donationStats)

	p.statCache.AddStatsItem(newStat)
	log.Debug("DonationStats:add statinfo:", "Req.Match:", newStat.PrintStatsItemJSON(), "lastSeq:", lastSeq, "Interval:", newStat.GetInterval())

	p.refreshStat(newStat)
	protocol.SetResult(c, rep, nil)

}

// QueryStatsInfo 获取统计项信息
// @Summary 获取统计项信息
// @Description get donation stats info
// @Tags Proof
// @Produce json
// @Param input body swagger.ClientRequestNil true "INPUT"
// @Success 200 {object} swagger.ServerResponse{result=[]string}
// @Failure 400 {object} swagger.ServerResponse{error=string}
// @Router /v1/proof/QueryStatsInfo [post]
func (p *Proof) QueryStatsInfo(c *gin.Context) {
	var stats []string

	if p.statCache != nil && len(p.statCache.ChildStatsItems) != 0 {
		for _, stat := range p.statCache.GetStatsItem() {
			if stat != nil {
				stats = append(stats, stat.PrintStatsItemJSON()+","+stat.PrintStatusJSON())
			}
		}
	}
	protocol.SetResult(c, stats, nil)
}

//refreshStat 定时刷新对应项的统计信息
func (p *Proof) refreshStat(stats rpcutils.StatsItem) {

	if stats.GetTimeout() != nil {
		stats.GetTimeout().Reset(time.Second * time.Duration(stats.GetInterval()))
		return
	}

	timeout := time.AfterFunc(time.Second*time.Duration(stats.GetInterval()), func() {
		lastConvertSeq := p.LastConvertSeq()
		if stats.GetIsRefreshing() || lastConvertSeq <= stats.GetLastSeq() {
			log.Debug("refreshStat:IsRefreshing", "Req.Match:", stats.PrintStatsItemJSON(), "Interval:", stats.GetInterval(), "LastSeq:", stats.GetLastSeq(), "lastConvertSeq:", lastConvertSeq, "IsRefreshing:", stats.GetIsRefreshing())
		} else {
			stats.SetIsRefreshing(true)
			rep, err := stats.StatsHandle(stats.GetReq())
			if err != nil {
				log.Error("refreshStat:fail!", "Req.Match:", stats.PrintStatsItemJSON(), "Interval:", stats.GetInterval(), "LastSeq:", stats.GetLastSeq(), "err", err)
			} else {
				stats.Update(rep)
				stats.SetLastSeq(lastConvertSeq)
				log.Debug("refreshStat:success!", "Req.Match:", stats.PrintStatsItemJSON(), "Interval:", stats.GetInterval(), "LastSeq:", stats.GetLastSeq())
			}
			stats.SetIsRefreshing(false)
		}
		stats.GetTimeout().Reset(time.Second * time.Duration(stats.GetInterval()))
	})
	stats.SetTimeout(timeout)
}

//donationStats 慈善捐款统计接口
func (p *Proof) donationStats(reqinfo interface{}) (interface{}, error) {
	req := reqinfo.(rpcutils.DonationStats)

	cli, err := escli.NewESShortConnect(p.Host, p.Prefix, p.Version, p.Username, p.Password)
	if err != nil {
		return nil, err
	}
	countq := querypara.Query{}
	matches := []*querypara.QMatch{}
	var q querypara.Query
	for _, match := range req.Match {
		q.Filter = append(q.Filter, &querypara.QMatch{Key: match.Key, Value: match.Value})
		matches = append(matches, &querypara.QMatch{Key: match.Key, Value: match.Value})
	}
	// 首先通过match条件获取数据个数用于桶的统计
	countq.Match = matches
	totalcount, err := p.Count(proofdb.ProofDBX, proofdb.ProofTableX, &countq)
	if err != nil || totalcount <= 0 {
		totalcount = 1000000
	}
	log.Debug("DonationStats:Elasticsearch totalcount", "totalcount", totalcount)

	field := req.TermsAgg.Key + ".keyword"

	a := &querypara.Agg{
		Name: "stat",
		Size: &querypara.ASize{Size: int(totalcount)},
		Term: &querypara.AAgg{Key: field, CollectionMode: "breadth_first"},
		Subs: &querypara.ASub{
			Sum: []*querypara.AAgg{{Name: "totalSum", Key: req.SubSumAgg.Key}},
			Pipeline: &querypara.APipe{
				Order: &querypara.AOrder{Name: "totalOrder", Key: "totalSum", Asc: false},
			},
		},
	}
	s := &querypara.Search{
		Query: &q,
		Agg:   a,
	}

	result, err := cli.Agg(proofdb.ProofTableX, proofdb.ProofTableX, s)
	if err != nil {
		log.Error("DonationStatOrder", "err", err.Error())
		return nil, err
	}

	totalAgg, found := result.Term("stat")
	if !found {
		log.Error("DonationStatOrder", "load stat terms failed", "not found")
		return nil, rpcutils.ErrNotFound
	}

	//用于统计返回的数量，目前只返回数量最多的前10000
	retcount := 0
	donationstats := rpcutils.RepDonationStates{}
	for index, b := range totalAgg {
		if b == nil {
			log.Error("DonationStatOrder.totalAgg.Buckets is nil!", "index", index)
			continue
		}

		sumagg, found := b.Sum("totalSum")
		if !found {
			log.Error("DonationStatOrder.Sum", "load total sum failed", "not found")
			continue
		}
		item := rpcutils.RepDonationStat{Name: b.Key.(string), Total: sumagg, Count: b.DocCount}
		donationstats.Itemes = append(donationstats.Itemes, &item)

		retcount++
		if retcount >= 10000 {
			log.Debug("DonationStatOrder", "retcount", retcount)
			break
		}
	}
	if len(donationstats.Itemes) != 0 {
		return &donationstats, nil
	}
	return nil, rpcutils.ErrNotFound
}
