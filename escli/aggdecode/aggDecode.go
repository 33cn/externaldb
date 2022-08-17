package aggdecode

import (
	"encoding/json"

	"github.com/33cn/chain33/common/log/log15"
	elasticV6 "github.com/olivere/elastic"
	elasticV7 "github.com/olivere/elastic/v7"
)

var (
	log = log15.New("module", "aggDecode")
)

type AggregationBucketKeyItem struct {
	Aggregations map[string]json.RawMessage
	Key          interface{}
	DocCount     int64
	Version      int32
}

func DisPoint(m map[string]*json.RawMessage) map[string]json.RawMessage {
	value := make(map[string]json.RawMessage)
	for k, v := range m {
		value[k] = *v
	}
	return value
}

func (a AggregationBucketKeyItem) Term(name string) ([]*AggregationBucketKeyItem, bool) {
	var bucket []*AggregationBucketKeyItem

	if raw, found := a.Aggregations[name]; found {
		if raw == nil {
			log.Info("get aggregations about term failed", "raw", "nil")
			return bucket, true
		}

		switch a.Version {
		case 6:
			agg := new(elasticV6.AggregationBucketKeyItems)
			if err := json.Unmarshal(raw, agg); err == nil {
				for _, item := range agg.Buckets {
					b := &AggregationBucketKeyItem{
						Aggregations: DisPoint(item.Aggregations),
						Key:          item.Key,
						Version:      6,
					}
					bucket = append(bucket, b)
				}
				return bucket, true
			}

		case 7:
			agg := new(elasticV7.AggregationBucketKeyItems)
			if err := json.Unmarshal(raw, agg); err == nil {
				for _, item := range agg.Buckets {
					b := &AggregationBucketKeyItem{
						Aggregations: item.Aggregations,
						Key:          item.Key,
						Version:      7,
					}
					bucket = append(bucket, b)
				}
				return bucket, true
			}
		}
	}
	return nil, false
}

func (a AggregationBucketKeyItem) Sum(name string) (float64, bool) {
	if raw, found := a.Aggregations[name]; found {
		if raw == nil {
			log.Info("get aggregations about sum failed", "raw", "nil")
			return 0, true
		}

		switch a.Version {
		case 6:
			sumagg := new(elasticV6.AggregationValueMetric)
			if err := json.Unmarshal(raw, sumagg); err == nil {
				return *sumagg.Value, true
			}

		case 7:
			sumagg := new(elasticV7.AggregationValueMetric)
			if err := json.Unmarshal(raw, sumagg); err == nil {
				return *sumagg.Value, true
			}
		}
	}
	return -1, false
}

func (a AggregationBucketKeyItem) Range(name string) ([]*AggregationBucketKeyItem, bool) {
	var bucket []*AggregationBucketKeyItem

	if raw, found := a.Aggregations[name]; found {
		if raw == nil {
			log.Info("get aggregations about range failed", "raw", "nil")
			return nil, true
		}

		switch a.Version {
		case 6:
			items := new(elasticV6.AggregationBucketRangeItems)
			if err := json.Unmarshal(raw, items); err == nil {
				for _, item := range items.Buckets {
					b := &AggregationBucketKeyItem{
						Aggregations: DisPoint(item.Aggregations),
						Key:          item.Key,
						DocCount:     item.DocCount,
						Version:      6,
					}
					bucket = append(bucket, b)
				}
				return bucket, true
			}

		case 7:
			items := new(elasticV7.AggregationBucketRangeItems)
			if err := json.Unmarshal(raw, items); err == nil {
				for _, item := range items.Buckets {
					b := &AggregationBucketKeyItem{
						Aggregations: item.Aggregations,
						Key:          item.Key,
						DocCount:     item.DocCount,
						Version:      7,
					}
					bucket = append(bucket, b)
				}
				return bucket, true
			}
		}
	}
	return nil, false
}
