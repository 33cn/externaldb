package swagger

// Status 服务状态详细信息
type Status struct {
	Server *ServerStatus `json:"server"` // 服务状态
	Chain  *ChainStatus  `json:"chain"`  // 链状态
	ES     *EsStatus     `json:"es"`     // ElasticSearch状态
}

// ServerStatus 当前服务状态信息
type ServerStatus struct {
	Version string `json:"version"`  // 版本
	SyncSeq int64  `json:"sync_seq"` // 同步序列高度
	ConvSeq int64  `json:"conv_seq"` // 转换序列高度
	Title   string `json:"title"`    // 标题
	Coin    string `json:"coin"`     // 币名称
}

// ChainStatus 区块链状态信息
type ChainStatus struct {
	Status  string            `json:"status"`   // 状态
	PushSeq int64             `json:"push_seq"` // 推送高度
	Coin    string            `json:"coin"`     // 主代币信息
	Version *ChainVersionInfo `json:"version"`  // 版本
}

// ChainVersionInfo 区块链版本信息
type ChainVersionInfo struct {
	Title   string `json:"title"`   // 区块链名，该节点 chain33.toml 中配置的 title 值
	App     string `json:"app"`     // 应用 app 的版本
	Chain33 string `json:"chain33"` // 版本信息，版本号-GitCommit（前八个字符）
	LocalDB string `json:"localDb"` // localdb 版本号
}

type EsStatus struct {
	Status      string          `json:"status"`       // 状态
	NodesCount  NodesCount      `json:"_nodes"`       // 节点数量信息
	ClusterName string          `json:"cluster_name"` // 集群名
	Nodes       map[string]Node `json:"nodes"`        // 节点信息
}
type NodesCount struct {
	Total      int `json:"total"`      // 总计
	Successful int `json:"successful"` // 正常数量
	Failed     int `json:"failed"`     // 不正常数量
}
type Attributes struct {
	MlMachineMemory string `json:"ml.machine_memory"` //
	MlMaxOpenJobs   string `json:"ml.max_open_jobs"`  //
	MlEnabled       string `json:"ml.enabled"`        //
}
type Docs struct {
	Count   int `json:"count"`   //
	Deleted int `json:"deleted"` //
}
type Store struct {
	SizeInBytes int `json:"size_in_bytes"` //
}
type Indexing struct {
	IndexTotal           int  `json:"index_total"`             //
	IndexTimeInMillis    int  `json:"index_time_in_millis"`    //
	IndexCurrent         int  `json:"index_current"`           //
	IndexFailed          int  `json:"index_failed"`            //
	DeleteTotal          int  `json:"delete_total"`            //
	DeleteTimeInMillis   int  `json:"delete_time_in_millis"`   //
	DeleteCurrent        int  `json:"delete_current"`          //
	NoopUpdateTotal      int  `json:"noop_update_total"`       //
	IsThrottled          bool `json:"is_throttled"`            //
	ThrottleTimeInMillis int  `json:"throttle_time_in_millis"` //
}
type Get struct {
	Total               int `json:"total"`                  //
	TimeInMillis        int `json:"time_in_millis"`         //
	ExistsTotal         int `json:"exists_total"`           //
	ExistsTimeInMillis  int `json:"exists_time_in_millis"`  //
	MissingTotal        int `json:"missing_total"`          //
	MissingTimeInMillis int `json:"missing_time_in_millis"` //
	Current             int `json:"current"`                //
}
type ISearch struct {
	OpenContexts        int `json:"open_contexts"`          //
	QueryTotal          int `json:"query_total"`            //
	QueryTimeInMillis   int `json:"query_time_in_millis"`   //
	QueryCurrent        int `json:"query_current"`          //
	FetchTotal          int `json:"fetch_total"`            //
	FetchTimeInMillis   int `json:"fetch_time_in_millis"`   //
	FetchCurrent        int `json:"fetch_current"`          //
	ScrollTotal         int `json:"scroll_total"`           //
	ScrollTimeInMillis  int `json:"scroll_time_in_millis"`  //
	ScrollCurrent       int `json:"scroll_current"`         //
	SuggestTotal        int `json:"suggest_total"`          //
	SuggestTimeInMillis int `json:"suggest_time_in_millis"` //
	SuggestCurrent      int `json:"suggest_current"`        //
}
type Merges struct {
	Current                    int   `json:"current"`                        //
	CurrentDocs                int   `json:"current_docs"`                   //
	CurrentSizeInBytes         int   `json:"current_size_in_bytes"`          //
	Total                      int   `json:"total"`                          //
	TotalTimeInMillis          int   `json:"total_time_in_millis"`           //
	TotalDocs                  int   `json:"total_docs"`                     //
	TotalSizeInBytes           int64 `json:"total_size_in_bytes"`            //
	TotalStoppedTimeInMillis   int   `json:"total_stopped_time_in_millis"`   //
	TotalThrottledTimeInMillis int   `json:"total_throttled_time_in_millis"` //
	TotalAutoThrottleInBytes   int64 `json:"total_auto_throttle_in_bytes"`   //
}
type Refresh struct {
	Total             int `json:"total"`                //
	TotalTimeInMillis int `json:"total_time_in_millis"` //
	Listeners         int `json:"listeners"`            //
}
type Flush struct {
	Total             int `json:"total"`                //
	TotalTimeInMillis int `json:"total_time_in_millis"` //
}
type Warmer struct {
	Current           int `json:"current"`              //
	Total             int `json:"total"`                //
	TotalTimeInMillis int `json:"total_time_in_millis"` //
}
type QueryCache struct {
	MemorySizeInBytes int `json:"memory_size_in_bytes"` //
	TotalCount        int `json:"total_count"`          //
	HitCount          int `json:"hit_count"`            //
	MissCount         int `json:"miss_count"`           //
	CacheSize         int `json:"cache_size"`           //
	CacheCount        int `json:"cache_count"`          //
	Evictions         int `json:"evictions"`            //
}
type Fielddata struct {
	MemorySizeInBytes int `json:"memory_size_in_bytes"` //
	Evictions         int `json:"evictions"`            //
}
type Completion struct {
	SizeInBytes int `json:"size_in_bytes"` //
}
type FileSizes struct {
}
type Segments struct {
	Count                     int       `json:"count"`                         //
	MemoryInBytes             int       `json:"memory_in_bytes"`               //
	TermsMemoryInBytes        int       `json:"terms_memory_in_bytes"`         //
	StoredFieldsMemoryInBytes int       `json:"stored_fields_memory_in_bytes"` //
	TermVectorsMemoryInBytes  int       `json:"term_vectors_memory_in_bytes"`  //
	NormsMemoryInBytes        int       `json:"norms_memory_in_bytes"`         //
	PointsMemoryInBytes       int       `json:"points_memory_in_bytes"`        //
	DocValuesMemoryInBytes    int       `json:"doc_values_memory_in_bytes"`    //
	IndexWriterMemoryInBytes  int       `json:"index_writer_memory_in_bytes"`  //
	VersionMapMemoryInBytes   int       `json:"version_map_memory_in_bytes"`   //
	FixedBitSetMemoryInBytes  int       `json:"fixed_bit_set_memory_in_bytes"` //
	MaxUnsafeAutoIDTimestamp  int       `json:"max_unsafe_auto_id_timestamp"`  //
	FileSizes                 FileSizes `json:"file_sizes"`                    //
}
type Translog struct {
	Operations             int `json:"operations"`                //
	SizeInBytes            int `json:"size_in_bytes"`             //
	UncommittedOperations  int `json:"uncommitted_operations"`    //
	UncommittedSizeInBytes int `json:"uncommitted_size_in_bytes"` //
}
type RequestCache struct {
	MemorySizeInBytes int `json:"memory_size_in_bytes"` //
	Evictions         int `json:"evictions"`            //
	HitCount          int `json:"hit_count"`            //
	MissCount         int `json:"miss_count"`           //
}
type Recovery struct {
	CurrentAsSource      int `json:"current_as_source"`       //
	CurrentAsTarget      int `json:"current_as_target"`       //
	ThrottleTimeInMillis int `json:"throttle_time_in_millis"` //
}
type Indices struct {
	Docs         Docs         `json:"docs"`          //
	Store        Store        `json:"store"`         //
	Indexing     Indexing     `json:"indexing"`      //
	Get          Get          `json:"get"`           //
	Search       ISearch      `json:"search"`        //
	Merges       Merges       `json:"merges"`        //
	Refresh      Refresh      `json:"refresh"`       //
	Flush        Flush        `json:"flush"`         //
	Warmer       Warmer       `json:"warmer"`        //
	QueryCache   QueryCache   `json:"query_cache"`   //
	Fielddata    Fielddata    `json:"fielddata"`     //
	Completion   Completion   `json:"completion"`    //
	Segments     Segments     `json:"segments"`      //
	Translog     Translog     `json:"translog"`      //
	RequestCache RequestCache `json:"request_cache"` //
	Recovery     Recovery     `json:"recovery"`      //
}
type LoadAverage struct {
	OneM  float64 `json:"1m"`  //
	FiveM float64 `json:"5m"`  //
	One5M float64 `json:"15m"` //
}
type CPU struct {
	Percent     int         `json:"percent"`      //
	LoadAverage LoadAverage `json:"load_average"` //
}
type OsMem struct {
	TotalInBytes int64 `json:"total_in_bytes"` //
	FreeInBytes  int64 `json:"free_in_bytes"`  //
	UsedInBytes  int64 `json:"used_in_bytes"`  //
	FreePercent  int   `json:"free_percent"`   //
	UsedPercent  int   `json:"used_percent"`   //
}
type Swap struct {
	TotalInBytes int `json:"total_in_bytes"` //
	FreeInBytes  int `json:"free_in_bytes"`  //
	UsedInBytes  int `json:"used_in_bytes"`  //
}
type Os struct {
	Timestamp int64 `json:"timestamp"` //
	CPU       CPU   `json:"cpu"`       //
	OsMem     OsMem `json:"mem"`       //
	Swap      Swap  `json:"swap"`      //
}
type PoolsYoung struct {
	UsedInBytes     int `json:"used_in_bytes"`      //
	MaxInBytes      int `json:"max_in_bytes"`       //
	PeakUsedInBytes int `json:"peak_used_in_bytes"` //
	PeakMaxInBytes  int `json:"peak_max_in_bytes"`  //
}
type Survivor struct {
	UsedInBytes     int `json:"used_in_bytes"`      //
	MaxInBytes      int `json:"max_in_bytes"`       //
	PeakUsedInBytes int `json:"peak_used_in_bytes"` //
	PeakMaxInBytes  int `json:"peak_max_in_bytes"`  //
}
type PoolsOld struct {
	UsedInBytes     int `json:"used_in_bytes"`      //
	MaxInBytes      int `json:"max_in_bytes"`       //
	PeakUsedInBytes int `json:"peak_used_in_bytes"` //
	PeakMaxInBytes  int `json:"peak_max_in_bytes"`  //
}
type Pools struct {
	PoolsYoung PoolsYoung `json:"young"`    //
	Survivor   Survivor   `json:"survivor"` //
	PoolsOld   PoolsOld   `json:"old"`      //
}
type JvmMem struct {
	HeapUsedInBytes         int   `json:"heap_used_in_bytes"`          //
	HeapUsedPercent         int   `json:"heap_used_percent"`           //
	HeapCommittedInBytes    int   `json:"heap_committed_in_bytes"`     //
	HeapMaxInBytes          int   `json:"heap_max_in_bytes"`           //
	NonHeapUsedInBytes      int   `json:"non_heap_used_in_bytes"`      //
	NonHeapCommittedInBytes int   `json:"non_heap_committed_in_bytes"` //
	Pools                   Pools `json:"pools"`                       //
}
type Threads struct {
	Count     int `json:"count"`      //
	PeakCount int `json:"peak_count"` //
}
type GcYoung struct {
	CollectionCount        int `json:"collection_count"`          //
	CollectionTimeInMillis int `json:"collection_time_in_millis"` //
}
type GcOld struct {
	CollectionCount        int `json:"collection_count"`          //
	CollectionTimeInMillis int `json:"collection_time_in_millis"` //
}
type Collectors struct {
	Young GcYoung `json:"young"` //
	Old   GcOld   `json:"old"`   //
}
type Gc struct {
	Collectors Collectors `json:"collectors"` //
}
type Direct struct {
	Count                int `json:"count"`                   //
	UsedInBytes          int `json:"used_in_bytes"`           //
	TotalCapacityInBytes int `json:"total_capacity_in_bytes"` //
}
type Mapped struct {
	Count                int `json:"count"`                   //
	UsedInBytes          int `json:"used_in_bytes"`           //
	TotalCapacityInBytes int `json:"total_capacity_in_bytes"` //
}
type BufferPools struct {
	Direct Direct `json:"direct"` //
	Mapped Mapped `json:"mapped"` //
}
type Classes struct {
	CurrentLoadedCount int `json:"current_loaded_count"` //
	TotalLoadedCount   int `json:"total_loaded_count"`   //
	TotalUnloadedCount int `json:"total_unloaded_count"` //
}
type Jvm struct {
	Timestamp      int64       `json:"timestamp"`        //
	UptimeInMillis int         `json:"uptime_in_millis"` //
	Mem            JvmMem      `json:"mem"`              //
	Threads        Threads     `json:"threads"`          //
	Gc             Gc          `json:"gc"`               //
	BufferPools    BufferPools `json:"buffer_pools"`     //
	Classes        Classes     `json:"classes"`          //
}
type HTTP struct {
	CurrentOpen int `json:"current_open"` //
	TotalOpened int `json:"total_opened"` //
}
type Node struct {
	Timestamp        int64      `json:"timestamp"`         //
	Name             string     `json:"name"`              //
	TransportAddress string     `json:"transport_address"` //
	Host             string     `json:"host"`              //
	IP               string     `json:"ip"`                //
	Roles            []string   `json:"roles"`             //
	Attributes       Attributes `json:"attributes"`        //
	Indices          Indices    `json:"indices"`           // 索引
	Os               Os         `json:"os"`                // 系统
	Jvm              Jvm        `json:"jvm"`               // java虚拟机
	HTTP             HTTP       `json:"http"`              //
}
