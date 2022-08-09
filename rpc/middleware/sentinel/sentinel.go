package sentinel

import (
	l "github.com/33cn/chain33/common/log/log15"
	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/gin-gonic/gin"
	sentinelPlugin "github.com/sentinel-group/sentinel-go-adapters/gin"
)

var (
	log = l.New("module", "sentinel")
)

// StnSystem 系统限流
func StnSystem(triggerCount float64) gin.HandlerFunc {
	var count float64 = 200
	if triggerCount != 0 {
		count = triggerCount
	}
	log.Error("StnSystem", "count", count)
	if _, err := system.LoadRules([]*system.Rule{
		{
			MetricType:   system.InboundQPS,
			TriggerCount: count,
			Strategy:     system.BBR,
		},
	}); err != nil {
		log.Error("Sentinel", "err", err)
	}
	return sentinelPlugin.SentinelMiddleware()
}

// SentinelFlow 对应接口限流
// func SentinelFlow() gin.HandlerFunc {
//
// 	//Load sentinel rules
// 	if _, err := flow.LoadRules([]*flow.Rule{
// 		{
// 			Resource:         "POST:/v1/proof/VolunteerStats",
// 			Threshold:        100,
// 			RelationStrategy: flow.CurrentResource,
// 			ControlBehavior:  flow.Reject,
// 			StatIntervalInMs: 1000,
// 		},
// 	}); err != nil {
// 		log.Error("Sentinel:VolunteerStats", "err", err)
// 	}
//
// 	if _, err := flow.LoadRules([]*flow.Rule{
// 		{
// 			Resource:         "POST:/v1/proof/DonationStats",
// 			Threshold:        100,
// 			RelationStrategy: flow.CurrentResource,
// 			ControlBehavior:  flow.Reject,
// 			StatIntervalInMs: 1000,
// 		},
// 	}); err != nil {
// 		log.Error("Sentinel:DonationStats", "err", err)
// 	}
//
// 	return sentinelPlugin.SentinelMiddleware()
//
// }
