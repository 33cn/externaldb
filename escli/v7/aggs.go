package v7

import (
	elasticV7 "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"

	"github.com/33cn/externaldb/escli/querypara"
)

func subSum(agg *querypara.Agg, a *elasticV7.TermsAggregation) *elasticV7.TermsAggregation {
	if agg.Subs.Sum != nil && len(agg.Subs.Sum) > 0 {
		for _, s := range agg.Subs.Sum {
			a = a.SubAggregation(s.Name, elasticV7.NewSumAggregation().Field(s.Key))
			continue
		}
	}
	return a
}

func subAvg(agg *querypara.Agg, a *elasticV7.TermsAggregation) *elasticV7.TermsAggregation {
	if agg.Subs.Avg != nil && len(agg.Subs.Avg) > 0 {
		for _, avg := range agg.Subs.Avg {
			a = a.SubAggregation(avg.Name, elasticV7.NewAvgAggregation().Field(avg.Key))
			continue
		}
	}
	return a
}

func subMin(query *querypara.Agg, a *elasticV7.TermsAggregation) *elasticV7.TermsAggregation {
	if query.Subs.Min != nil && len(query.Subs.Min) > 0 {
		for _, s := range query.Subs.Min {
			a = a.SubAggregation(s.Name, elasticV7.NewMinAggregation().Field(s.Key))
			continue
		}
	}
	return a
}

func subMax(agg *querypara.Agg, a *elasticV7.TermsAggregation) *elasticV7.TermsAggregation {
	if agg.Subs.Max != nil && len(agg.Subs.Max) > 0 {
		for _, s := range agg.Subs.Max {
			a = a.SubAggregation(s.Name, elasticV7.NewMaxAggregation().Field(s.Key))
			continue
		}
	}
	return a
}

func subPipeline(agg *querypara.Agg, a *elasticV7.TermsAggregation) *elasticV7.TermsAggregation {
	if agg.Subs.Pipeline != nil {
		if agg.Subs.Pipeline.Order != nil {
			a = a.SubAggregation(agg.Subs.Pipeline.Order.Name, elasticV7.NewBucketSortAggregation().Sort(agg.Subs.Pipeline.Order.Key, agg.Subs.Pipeline.Order.Asc))
		}
	}
	return a
}

func subAgg(agg *querypara.Agg, a *elasticV7.TermsAggregation) *elasticV7.TermsAggregation {
	if agg.Subs.SubAgg != nil {
		s := getTermAgg(agg.Subs.SubAgg)
		a = a.SubAggregation(agg.Subs.SubAgg.Name, s)
		return a
	}
	a = subSum(agg, a)
	a = subAvg(agg, a)
	a = subMin(agg, a)
	a = subMax(agg, a)
	a = subPipeline(agg, a)
	return a
}

func getTermAgg(agg *querypara.Agg) *elasticV7.TermsAggregation {
	if agg == nil {
		return nil
	}

	if agg.Term != nil {
		a := elasticV7.NewTermsAggregation().Field(agg.Term.Key)
		if agg.Term.CollectionMode != "" {
			a = a.CollectionMode(agg.Term.CollectionMode)
		}

		if agg.Subs != nil {
			a = subAgg(agg, a)
		}

		if agg.Size != nil {
			a = a.Size(agg.Size.Size)
		}
		if agg.Order != nil {
			a = a.OrderByAggregation(agg.Order.Key, agg.Order.Asc)
		}
		return a
	}
	return nil
}

func getMetricAgg(agg *querypara.Agg) elasticV7.Aggregation {
	if agg == nil {
		return nil
	}

	if agg.Metric.Sum != nil {
		a := elasticV7.NewSumAggregation().Field(agg.Metric.Sum.Key)
		return a
	}
	if agg.Metric.Avg != nil {
		a := elasticV7.NewAvgAggregation().Field(agg.Metric.Avg.Key)
		return a
	}
	if agg.Metric.Max != nil {
		a := elasticV7.NewMaxAggregation().Field(agg.Metric.Max.Key)
		return a
	}
	if agg.Metric.Min != nil {
		a := elasticV7.NewMinAggregation().Field(agg.Metric.Min.Key)
		return a
	}

	return nil
}

func getRangeAgg(agg *querypara.Agg) *elasticV7.RangeAggregation {
	if agg.Ranges == nil {
		return nil
	}

	a := elasticV7.NewRangeAggregation().Field(agg.Ranges.Key)

	for _, aRange := range agg.Ranges.Ranges {
		if aRange.RStart == nil && aRange.REnd != nil {
			a = a.AddUnboundedFrom(aRange.REnd)
		} else if aRange.RStart != nil && aRange.REnd != nil {
			a = a.AddRange(aRange.RStart, aRange.REnd)
		} else if aRange.RStart != nil && aRange.REnd == nil {
			a = a.AddUnboundedTo(aRange.RStart)
		}
	}

	return a
}

func getCardinalityAgg(agg *querypara.Agg) *elasticV7.CardinalityAggregation {
	if agg.Cardinality == "" {
		return nil
	}

	a := elasticV7.NewCardinalityAggregation().Field(agg.Cardinality)

	return a
}

// Aggregation Aggregation
func Aggregation(search *elasticV7.SearchService, aggs *querypara.Agg) (*elasticV7.SearchService, error) {
	if aggs.Name == "" {
		return nil, errors.Errorf("have not set agg key")
	}

	if aggs.Metric != nil {
		agg := getMetricAgg(aggs)
		if agg != nil {
			search = search.Aggregation(aggs.Name, agg)
		}

		return search, nil
	} else if aggs.Ranges != nil {
		agg := getRangeAgg(aggs)
		if agg != nil {
			search = search.Aggregation(aggs.Name, agg)
		}

		return search, nil
	} else if aggs.Cardinality != "" {
		agg := getCardinalityAgg(aggs)
		if agg != nil {
			search = search.Aggregation(aggs.Name, agg)
		}

		return search, nil
	} else {
		agg := getTermAgg(aggs)
		if agg != nil {
			search = search.Aggregation(aggs.Name, agg)
		}

		return search, nil
	}

}
