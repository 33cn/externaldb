package rpcutils

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

var (
	DonateFlag    = 1
	VolunteerFlag = 2
	MaxStatsItems = 5
)

type CallBackFunc func(req interface{}) (interface{}, error)

//统计项组件的方法集
//StatsItem
type StatsItem interface {
	IsMatch(falgType int, req interface{}) bool
	PrintStatsItemJSON() string
	StatsComm
}

type StatsComm interface {
	SetLastSeq(lastseq int64)
	GetLastSeq() int64
	SetTimeout(timeout *time.Timer)
	GetTimeout() *time.Timer
	StopTimeOut()
	SetIsRefreshing(IsRefreshing bool)
	GetIsRefreshing() bool
	GetInterval() int64
	GetLastCallTime() int64
	Update(rep interface{})
	GetRep() interface{}
	GetReq() interface{}
	PrintStatusJSON() string
	StatsHandle(req interface{}) (interface{}, error)
}

// BaseStatisItem 基础统计项组件
type BaseStatsItem struct {
	ChildStatsItems []StatsItem // 子统计项组件列表
	MaxStatsItems   int
	sync.RWMutex
}

// AddStatisItem 添加新的统计项
func (b *BaseStatsItem) AddStatsItem(s StatsItem) {
	b.Lock()
	defer b.Unlock()

	b.ChildStatsItems = append(b.ChildStatsItems, s)

	// 超过最大缓存的统计项时需要删除最旧没有被调用的统计项
	if len(b.ChildStatsItems) > b.MaxStatsItems {
		b.removeOldStatsItem()
	}

}

// GetStatisItem 获取已经缓存的统计项信息
func (b *BaseStatsItem) GetStatsItem() []StatsItem {
	b.RLock()
	defer b.RUnlock()
	return b.ChildStatsItems
}

// removeStatsItem 删除老的统计项
func (b *BaseStatsItem) removeOldStatsItem() {

	var lastCallTime int64
	index := 0
	for k, childStatsItem := range b.ChildStatsItems {
		callTime := childStatsItem.GetLastCallTime()
		if callTime < lastCallTime {
			lastCallTime = callTime
			index = k
		}
	}
	b.ChildStatsItems[index].StopTimeOut()
	b.ChildStatsItems = append(b.ChildStatsItems[:index], b.ChildStatsItems[index+1:]...)

}

// 定义一个缓存慈善全款统计信息的结构，底层定时刷新到cache中，前端直接访问cache中的即可
// Req 用于统计的条件
// Rep 统计的结果
// Interval 定时统计的时间间隔
// LastSeq   定时刷新时的最新区块的seq，如果seq没有增加就不去刷新直接使用cache缓存的统计信息
// LastCallTime 此缓存数据被调用的最新时间，用于缓存项超过最大值时删除最旧没有被调用的统计项，防止缓存数据太多
// IsRefreshing 标记此项统计是否正在进行，如果正在统计进行时不再触发统计，等待下个周期再刷新统计
type StatsCommInfo struct {
	Req          interface{}
	Rep          interface{}
	Interval     int64
	LastSeq      int64
	LastCallTime int64
	IsRefreshing bool
	timeout      *time.Timer
	sync.RWMutex
	CallBack CallBackFunc
}

//NewStatsComm 创建一个新的统计公共实例
func NewStatsComm(req, rep interface{}, interval int64, lastSeq int64, callBack CallBackFunc) StatsComm {
	return &StatsCommInfo{
		Req:          req,
		Rep:          rep,
		Interval:     interval,
		LastSeq:      lastSeq,
		CallBack:     callBack,
		LastCallTime: time.Now().Unix(),
	}
}
func (bi *StatsCommInfo) SetLastSeq(lastseq int64) {
	bi.Lock()
	defer bi.Unlock()
	bi.LastSeq = lastseq
}
func (bi *StatsCommInfo) GetLastSeq() int64 {
	bi.RLock()
	defer bi.RUnlock()
	return bi.LastSeq
}

func (bi *StatsCommInfo) SetTimeout(timeout *time.Timer) {
	bi.Lock()
	defer bi.Unlock()
	bi.timeout = timeout
}
func (bi *StatsCommInfo) GetTimeout() *time.Timer {
	bi.RLock()
	defer bi.RUnlock()
	return bi.timeout
}

func (bi *StatsCommInfo) StopTimeOut() {
	bi.Lock()
	defer bi.Unlock()
	if bi.timeout != nil {
		bi.timeout.Stop()
	}
}
func (bi *StatsCommInfo) SetIsRefreshing(IsRefreshing bool) {
	bi.Lock()
	defer bi.Unlock()
	bi.IsRefreshing = IsRefreshing
}

func (bi *StatsCommInfo) GetIsRefreshing() bool {
	bi.RLock()
	defer bi.RUnlock()
	return bi.IsRefreshing
}
func (bi *StatsCommInfo) GetInterval() int64 {
	bi.RLock()
	defer bi.RUnlock()
	return bi.Interval
}
func (bi *StatsCommInfo) GetLastCallTime() int64 {
	bi.RLock()
	defer bi.RUnlock()
	return bi.LastCallTime
}

//StatsHandle 比较需要统计的信息和现有cache统计信息是否匹配
func (bi *StatsCommInfo) StatsHandle(req interface{}) (interface{}, error) {
	bi.RLock()
	defer bi.RUnlock()
	return bi.CallBack(req)
}

//Update 定时更新对应的统计信息
func (bi *StatsCommInfo) Update(rep interface{}) {
	bi.Lock()
	defer bi.Unlock()
	bi.Rep = rep
}

//GetRep 获取缓存的返回信息
func (bi *StatsCommInfo) GetRep() interface{} {
	bi.RLock()
	defer bi.RUnlock()
	bi.LastCallTime = time.Now().Unix()
	return bi.Rep
}

//GetReq 获取缓存的请求信息
func (bi *StatsCommInfo) GetReq() interface{} {
	bi.RLock()
	defer bi.RUnlock()
	return bi.Req
}

type StatsItemStatus struct {
	Interval     int64 `json:"interval"`
	LastSeq      int64 `json:"lastSeq"`
	LastCallTime int64 `json:"lastCallTime"`
	IsRefreshing bool  `json:"isRefreshing"`
}

func (bi *StatsCommInfo) PrintStatusJSON() string {

	status := &StatsItemStatus{
		Interval:     bi.Interval,
		LastSeq:      bi.LastSeq,
		LastCallTime: bi.LastCallTime,
		IsRefreshing: bi.IsRefreshing,
	}
	buf, err := json.MarshalIndent(status, "", " ") //格式化编码
	if err != nil {
		fmt.Println("err = ", err)
		return ""
	}
	return string(buf)

}

//StatsItemInfo具体统计项的信息
type DonateStatsItem struct {
	PrintJSON string
	StatsComm
}

//IsMatch 比较需要统计的信息和现有cache统计信息是否匹配
func (bi *DonateStatsItem) IsMatch(falgType int, reqinfo interface{}) bool {
	if falgType != bi.FlagType() {
		return false
	}
	req := reqinfo.(DonationStats)

	sreq := bi.GetReq().(DonationStats)

	if sreq.TermsAgg.Key != req.TermsAgg.Key || sreq.SubSumAgg.Key != req.SubSumAgg.Key {
		return false
	}
	if len(sreq.Match) != len(req.Match) {
		return false
	}
	matchNum := 0
	for _, match := range sreq.Match {
		exist := false
		for _, inmatch := range req.Match {
			if match.Key == inmatch.Key && match.Value == inmatch.Value {
				exist = true
				matchNum++
				break
			}
		}
		if !exist {
			return false
		}
	}
	return matchNum == len(req.Match)
}

//PrintJson  通过json格式打印字符串
func (bi *DonateStatsItem) PrintStatsItemJSON() string {
	if len(bi.PrintJSON) != 0 {
		return bi.PrintJSON
	}

	req := bi.GetReq().(DonationStats)
	buf, err := json.MarshalIndent(req, "", " ") //格式化编码
	if err != nil {
		return ""
	}
	bi.PrintJSON = string(buf)
	return bi.PrintJSON
}

//FalgType   返回统计项的类型
func (bi *DonateStatsItem) FlagType() int {
	return DonateFlag
}

//VolunteerStatsItem 具体统计项的信息
type VolunteerStatsItem struct {
	PrintJSON string
	StatsComm
}

//IsMatch 比较需要统计的信息和现有cache统计信息是否匹配
func (bi *VolunteerStatsItem) IsMatch(falgType int, reqinfo interface{}) bool {
	if falgType != bi.FlagType() {
		return false
	}
	req := reqinfo.(VolunteerStats)

	sreq := bi.GetReq().(VolunteerStats)

	if sreq.TermsAgg.Key != req.TermsAgg.Key || sreq.SubSumAgg.Key != req.SubSumAgg.Key || sreq.SubTermsAgg.Key != req.SubTermsAgg.Key {
		return false
	}
	if len(sreq.Match) != len(req.Match) {
		return false
	}
	matchNum := 0
	for _, match := range sreq.Match {
		exist := false
		for _, inmatch := range req.Match {
			if match.Key == inmatch.Key && match.Value == inmatch.Value {
				exist = true
				matchNum++
				break
			}
		}
		if !exist {
			return false
		}
	}
	return matchNum == len(req.Match)
}

//PrintJson  通过json格式打印字符串
func (bi *VolunteerStatsItem) PrintStatsItemJSON() string {
	if len(bi.PrintJSON) != 0 {
		return bi.PrintJSON
	}

	req := bi.GetReq().(VolunteerStats)
	buf, err := json.MarshalIndent(req, "", " ") //格式化编码
	if err != nil {
		return ""
	}

	bi.PrintJSON = string(buf)

	return bi.PrintJSON
}

//FalgType   返回统计项的类型
func (bi *VolunteerStatsItem) FlagType() int {
	return VolunteerFlag
}
