package swagger

// QMatch key
type QMatch struct {
	Key   string      `json:"key"`   // 字段名
	Value interface{} `json:"value"` // 值
	//SubQuery *Query      `json:"query"` // 子查询
}

// QFetch 获取
type QFetch struct {
	FetchSource bool     `json:"fetch_source"` // 是否获取
	Keys        []string `json:"keys"`         // 字段名列表
}

// QMultiMatch keys
type QMultiMatch struct {
	Keys  []string    `json:"keys"`  // 字段名列表
	Value interface{} `json:"value"` // 值
}

// QRange range
type QRange struct {
	Key    string      `json:"key"`   // 字段名
	RStart interface{} `json:"start"` // 大于等于
	REnd   interface{} `json:"end"`   // 小于等于
	GT     interface{} `json:"gt"`    // 大于
	LT     interface{} `json:"lt"`    // 小于
}

// QSort sort
type QSort struct {
	Key       string `json:"key"`       // 字段名
	Ascending bool   `json:"ascending"` // 是否递增
}

// QPage page
type QPage struct {
	Size   int `json:"size"`   // 大小
	Number int `json:"number"` // 当前页数
}

type QSize struct {
	Size int `json:"size"` // 大小
}

// Query queryPara
type Query struct {
	Page       *QPage         `json:"page"`        // 分页
	Size       *QSize         `json:"size"`        // 大小
	Sort       []*QSort       `json:"sort"`        // 排序
	Range      []*QRange      `json:"range"`       // 范围
	Match      []*QMatch      `json:"match"`       // 且匹配
	MatchOne   []*QMatch      `json:"match_one"`   // 或匹配
	MultiMatch []*QMultiMatch `json:"multi_match"` // 多字段匹配
	Filter     []*QMatch      `json:"filter"`      // 过滤
	Not        []*QMatch      `json:"not"`         // 非匹配
	Fetch      *QFetch        `json:"fetch"`       // 获取字段
}

type AAgg struct {
	Name           string `json:"name"`            // 聚合名
	Key            string `json:"key"`             // 聚合字段
	CollectionMode string `json:"collection_mode"` // depth_first/breadth_first
}

// ASize 大小
type ASize struct {
	Size int `json:"size"` // 大小
}

// AOrder 排序
type AOrder struct {
	Name string `json:"name"` // 名称
	Key  string `json:"key"`  // 字段名
	Asc  bool   `json:"asc"`  // 是否升序
}

// ARange 范围
type ARange struct {
	RStart interface{} `json:"start"` // 开始位置（大于等于）
	REnd   interface{} `json:"end"`   // 结束位置（小于等于）
}

// ARanges 范围
type ARanges struct {
	Key    string    `json:"key"`    // 字段名
	Ranges []*ARange `json:"ranges"` // 范围
}

// APipe 管道聚合：对其它聚合操作的输出以及关联指标进行聚合
type APipe struct {
	Order *AOrder `json:"order"` // 排序
}

// ASub 对聚合的嵌套操作
type ASub struct {
	Sum      []*AAgg `json:"sum"`      // 和
	Avg      []*AAgg `json:"avg"`      // 平均值
	Min      []*AAgg `json:"min"`      // 最小值
	Max      []*AAgg `json:"max"`      // 最大值
	Pipeline *APipe  `json:"pipeline"` // 管道聚合
	SubAgg   *Agg    `json:"sub_agg"`  // 子组合
}

// AMetric 指标聚合
type AMetric struct {
	Sum *AAgg `json:"sum"` // 和
	Avg *AAgg `json:"avg"` // 平均值
	Min *AAgg `json:"min"` // 最小值
	Max *AAgg `json:"max"` // 最大值
}

type Agg struct {
	Name   string   `json:"name"`   // 聚合名
	Metric *AMetric `json:"metric"` // 指标聚合
	Term   *AAgg    `json:"term"`   // 分桶聚合的Term聚合
	Ranges *ARanges `json:"ranges"` // 分桶聚合的Range聚合
	Subs   *ASub    `json:"sub"`    // 对聚合的嵌套操作
	Size   *ASize   `json:"size"`   // 大小
	Order  *AOrder  `json:"order"`  // 排序
}

type Search struct {
	Agg   *Agg   `json:"agg"`   // 聚合
	Query *Query `json:"query"` // 查询
}
