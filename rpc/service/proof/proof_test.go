package proof

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/33cn/externaldb/db/proof/proofdb"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/escli/querypara"
	rpcutils "github.com/33cn/externaldb/rpc/utils"
)

func TestProof_DonationStats(t *testing.T) {
	proof := Proof{DBRead: &rpcutils.DBRead{Host: "http://172.16.101.87:9200", Title: "", Symbol: "bty", Prefix: "v12db02_", Version: 6}}
	req := rpcutils.DonationStats{
		Match:     []*rpcutils.QMatch{{Key: "存证类型", Value: "公益捐赠"}},
		SubSumAgg: &rpcutils.QMatchKey{Key: "捐赠数量"},
		TermsAgg:  &rpcutils.QMatchKey{Key: "捐赠人"},
	}
	Stats, err := proof.donationStats(req)
	if err != nil {
		// 测试ES没有开启
		return
	}
	assert.Nil(t, err)
	t.Log("Stats", Stats)
}

func TestProof_VolunteerStats(t *testing.T) {
	proof := Proof{DBRead: &rpcutils.DBRead{Host: "http://172.16.101.87:9200", Title: "", Symbol: "bty", Prefix: "v12db02_", Version: 6}}
	req := rpcutils.VolunteerStats{
		Match:       []*rpcutils.QMatch{{Key: "存证类型", Value: "志愿者登记"}},
		SubSumAgg:   &rpcutils.QMatchKey{Key: "志愿者数量"},
		SubTermsAgg: &rpcutils.QMatchKey{Key: "所在单位"},
		TermsAgg:    &rpcutils.QMatchKey{Key: "所在省份"},
	}
	proofs, err := proof.volunteerStats(req)
	if err != nil {
		// 测试ES没有开启
		return
	}
	assert.Nil(t, err)
	t.Log("proofs", proofs)
}

func TestProof_TotalStats(t *testing.T) {
	p := Proof{DBRead: &rpcutils.DBRead{Host: "http://172.16.101.87:9200", Title: "", Symbol: "bty", Prefix: "v12db02_", Version: 6}}
	req := &rpcutils.TotalStats{
		Match:  []*rpcutils.QMatch{{Key: "存证类型", Value: "志愿者登记"}},
		SumAgg: &rpcutils.QMatchKey{Key: "志愿者数量"},
	}

	cli, err := escli.NewESShortConnect(p.Host, p.Prefix, p.Version, p.Username, p.Password)
	if err != nil {
		log.Error("TotalStats:Elasticsearch connect", "err", err)
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
		return
	}
	sumagg, found := result.Sum("sumstats")

	assert.Nil(t, err)
	t.Log("TotalStats", sumagg, found)
}

func TestProof_FetchSource(t *testing.T) {
	p := Proof{DBRead: &rpcutils.DBRead{Host: "http://172.16.101.87:9200", Title: "", Symbol: "bty", Prefix: "v12db02_", Version: 6}}
	req := &rpcutils.SpecifiedFields{
		Match:  []*rpcutils.QMatch{{Key: "捐赠平台", Value: "大连"}},
		Count:  10,
		Fields: []string{"捐赠人", "存证名称"},
	}

	cli, err := escli.NewESShortConnect(p.Host, p.Prefix, p.Version, p.Username, p.Password)
	if err != nil {
		log.Error("FetchSource:Elasticsearch connect", "err", err)
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
		return
	}
	assert.Nil(t, err)
	t.Log("FetchSource", result)
}

func TestProof_ListUpdateProof(t *testing.T) {
	p := Proof{DBRead: &rpcutils.DBRead{Host: "http://172.16.101.87:9200", Title: "", Symbol: "bty", Prefix: "v2db02_", Version: 6}}
	var err error
	reselt := make([]interface{}, 0)
	if err != nil {
		t.Log(err)
		return
	}
	var q querypara.Query
	// 默认分页一次取10个数据
	if q.Page == nil {
		q.Page = &querypara.QPage{
			Number: 1,
			Size:   10,
		}
	}
	q.MatchOne = []*querypara.QMatch{
		{Key: "proof_tx_hash", Value: "SyName\":\"1111"},
		{Key: "proof_note", Value: "SyName\":\"1111"},
	}
	q.Match = []*querypara.QMatch{
		{Key: "basehash", Value: "null"},
		{Key: "proof_deleted_flag", Value: false},
	}
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
		return
	}

	//查询增量存证
	var majors []interface{}
	majors, err = cli.Search(proofdb.ProofDBX, proofdb.ProofTableX, &q, decodeProof)
	if err != nil || majors == nil {
		//特殊处理一下没有数据时sort功能报错的场景,取消sort再尝试调用一次
		if q.Sort != nil {
			q.Sort = nil
			majors, err = cli.Search(proofdb.ProofDBX, proofdb.ProofTableX, &q, decodeProof)
		} else {
			return
		}
	}

	//查询到update proof的最新版本，也就是最新状态
	for _, proof := range majors {
		proof := proof.(map[string]interface{})
		if _, ok := proof["update_version"]; !ok || proof["update_version"].(float64) == 0 {
			//没有设置更新存证
			reselt = append(reselt, proof)
		} else {
			q2 := &querypara.Query{
				Match: []*querypara.QMatch{
					{Key: "update_hash", Value: proof["proof_tx_hash"]},
					{Key: "update_version", Value: proof["update_version"]},
				},
				Page: &querypara.QPage{
					Size:   1,
					Number: 1,
				},
			}
			// 查询最新的存证
			new, err := cli.Search(proofdb.ProofDBX, proofdb.ProofTableX, q2, decodeProof)
			if err != nil || new == nil {
				return
			}
			reselt = append(reselt, new[0])
		}
	}
	for _, i2 := range reselt {
		t.Log(2, i2)
	}
}

func TestProof_CountByTime(t *testing.T) {
	p := Proof{DBRead: &rpcutils.DBRead{Host: "http://172.16.101.87:9200", Title: "", Symbol: "bty", Prefix: "v2db02_", Version: 6}}

	cli, _ := escli.NewESShortConnect(p.Host, p.Prefix, p.Version, p.Username, p.Password)

	req := &rpcutils.CountByTime{
		Ranges: &rpcutils.QRanges{
			Key:    "proof_block_time",
			Ranges: []*rpcutils.Range{{RStart: 1632098305, REnd: 1633998305}, {RStart: 1633998305, REnd: 1634627709}},
		},
	}

	var ranges []*querypara.ARange
	for _, r := range req.Ranges.Ranges {
		ranges = append(ranges, &querypara.ARange{RStart: r.RStart, REnd: r.REnd})
	}

	ag := &querypara.Agg{
		Name: "name",
		Ranges: &querypara.ARanges{
			Key:    "proof_block_time",
			Ranges: ranges,
		},
	}
	qu := &querypara.Query{
		Match: []*querypara.QMatch{
			{Key: "proof_delete_flag", Value: false},
		},
	}

	s := querypara.Search{
		Agg:   ag,
		Query: qu,
	}

	result, _ := cli.Agg(proofdb.ProofDBX, proofdb.ProofTableX, &s)
	r, found := result.Range("name")
	if !found {
		log.Error("CountByTime", "not found")
	}
	for i, item := range r {
		t.Log(i, item)
	}
}
