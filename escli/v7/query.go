package v7

import (
	elasticV7 "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"

	"github.com/33cn/externaldb/escli/querypara"
)

// QueryFunc 需要根据需求调整 match_phase term match queryPara
var QueryFunc = "match_phase"

func queryShouldHelper(q *elasticV7.BoolQuery, s *querypara.QMatch) *elasticV7.BoolQuery {
	if s.Value == nil {
		qi := elasticV7.NewExistsQuery(s.Key)
		return q.Should(qi)
	}
	if QueryFunc == "match_phase" {
		qi := elasticV7.NewMatchPhraseQuery(s.Key, s.Value)
		return q.Should(qi)
	} else if QueryFunc == "term" {
		qi := elasticV7.NewTermQuery(s.Key, s.Value)
		return q.Should(qi)
	}
	qi := elasticV7.NewMatchQuery(s.Key, s.Value)
	return q.Should(qi)
}

func queryMustHelper(q *elasticV7.BoolQuery, s *querypara.QMatch) *elasticV7.BoolQuery {
	if s.Value == nil {
		qi := elasticV7.NewExistsQuery(s.Key)
		return q.Must(qi)
	}
	if QueryFunc == "match_phase" {
		qi := elasticV7.NewMatchPhraseQuery(s.Key, s.Value)
		return q.Must(qi)
	} else if QueryFunc == "term" {
		qi := elasticV7.NewTermQuery(s.Key, s.Value)
		return q.Must(qi)
	}
	qi := elasticV7.NewMatchQuery(s.Key, s.Value)
	return q.Must(qi)
}

func queryMustNotHelper(q *elasticV7.BoolQuery, s *querypara.QMatch) *elasticV7.BoolQuery {
	if s.Value == nil {
		qi := elasticV7.NewExistsQuery(s.Key)
		return q.MustNot(qi)
	}
	if QueryFunc == "match_phase" {
		qi := elasticV7.NewMatchPhraseQuery(s.Key, s.Value)
		return q.MustNot(qi)
	} else if QueryFunc == "term" {
		qi := elasticV7.NewTermQuery(s.Key, s.Value)
		return q.MustNot(qi)
	}
	qi := elasticV7.NewMatchQuery(s.Key, s.Value)
	return q.MustNot(qi)
}

func queryFilterHelper(q *elasticV7.BoolQuery, s *querypara.QMatch) *elasticV7.BoolQuery {
	if s.Value == nil {
		qi := elasticV7.NewExistsQuery(s.Key)
		return q.Filter(qi)
	}
	if QueryFunc == "match_phase" {
		qi := elasticV7.NewMatchPhraseQuery(s.Key, s.Value)
		return q.Filter(qi)
	} else if QueryFunc == "term" {
		qi := elasticV7.NewTermQuery(s.Key, s.Value)
		return q.Filter(qi)
	}
	qi := elasticV7.NewMatchQuery(s.Key, s.Value)
	return q.Filter(qi)
}

// ShouldMatch 最少匹配一项
func addShouldMatch(query *querypara.Query, q *elasticV7.BoolQuery) *elasticV7.BoolQuery {
	if query.MatchOne != nil && len(query.MatchOne) > 0 {
		for _, s := range query.MatchOne {
			if s.SubQuery != nil {
				q.Should(getQuery(s.SubQuery))
				continue
			}
			q = queryShouldHelper(q, s)
		}
		matchCount := 1
		q = q.MinimumNumberShouldMatch(matchCount)
	}
	return q
}

func addMultiMatch(query *querypara.Query, q *elasticV7.BoolQuery) *elasticV7.BoolQuery {
	if query.MultiMatch != nil && len(query.MultiMatch) > 0 {
		for _, s := range query.MultiMatch {
			qi := elasticV7.NewMultiMatchQuery(s.Value, s.Keys...)
			q = q.Should(qi)
		}
		matchCount := 1
		if query.Match != nil {
			matchCount += len(query.Match)
		}
		q = q.MinimumNumberShouldMatch(matchCount)
	}
	return q
}

func addMatch(query *querypara.Query, q *elasticV7.BoolQuery) *elasticV7.BoolQuery {
	if query.Match != nil && len(query.Match) > 0 {
		for _, s := range query.Match {
			if s.SubQuery != nil {
				q.Must(getQuery(s.SubQuery))
				continue
			}
			q = queryMustHelper(q, s)
		}
	}
	return q
}

func addMatchNot(query *querypara.Query, q *elasticV7.BoolQuery) *elasticV7.BoolQuery {
	if query.Not != nil && len(query.Not) > 0 {
		for _, s := range query.Not {
			if s.SubQuery != nil {
				q.MustNot(getQuery(s.SubQuery))
				continue
			}
			q = queryMustNotHelper(q, s)
		}
	}
	return q
}

func addFilterMatch(query *querypara.Query, q *elasticV7.BoolQuery) *elasticV7.BoolQuery {
	if query.Filter != nil && len(query.Filter) > 0 {
		for _, s := range query.Filter {
			if s.SubQuery != nil {
				q.Filter(getQuery(s.SubQuery))
				continue
			}
			q = queryFilterHelper(q, s)
		}
	}
	return q
}

func setRange(query *querypara.Query, q *elasticV7.BoolQuery) *elasticV7.BoolQuery {
	if query.Range != nil && len(query.Range) > 0 {
		for _, s := range query.Range {
			qi := elasticV7.NewRangeQuery(s.Key)
			if s.RStart != nil {
				qi = qi.From(s.RStart)
			}
			if s.REnd != nil {
				qi = qi.To(s.REnd)
			}
			if s.GT != nil {
				qi = qi.Gt(s.GT)
			}
			if s.LT != nil {
				qi = qi.Lt(s.LT)
			}
			q = q.Filter(qi)
		}
	}
	return q
}

func getQuery(query *querypara.Query) *elasticV7.BoolQuery {
	if query == nil {
		return nil
	}
	if query.Range != nil || query.Match != nil || query.MatchOne != nil || query.MultiMatch != nil || query.Filter != nil || query.Not != nil {
		q := elasticV7.NewBoolQuery()
		q = setRange(query, q)
		q = addMatch(query, q)
		q = addMultiMatch(query, q)
		q = addShouldMatch(query, q)
		q = addFilterMatch(query, q)
		q = addMatchNot(query, q)
		return q
	}
	return nil
}

func Query(search *elasticV7.SearchService, query *querypara.Query) (*elasticV7.SearchService, error) {
	if query == nil {
		return nil, errors.Errorf("search bad queryPara input")
	}

	q := getQuery(query)
	if q != nil {
		search = search.Query(q)
	}

	if query.Page != nil {
		if query.Page.Size <= 0 || query.Page.Number <= 0 {
			return nil, errors.Errorf("search bad queryPara page input")
		}
		search = search.Size(query.Page.Size).From(query.Page.Size * (query.Page.Number - 1))
	}
	if query.Sort != nil && len(query.Sort) > 0 {
		for _, s := range query.Sort {
			search = search.Sort(s.Key, s.Ascending)
		}
	}
	if query.Size != nil {
		search = search.Size(query.Size.Size)
	}
	if query.Fetch != nil && len(query.Fetch.Keys) > 0 {
		fsc := elasticV7.NewFetchSourceContext(query.Fetch.FetchSource).Include(query.Fetch.Keys...)
		search = search.FetchSourceContext(fsc)
	}

	return search, nil
}
