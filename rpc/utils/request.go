package rpcutils

import (
	"encoding/json"
)

// Addresses Addresses
type Addresses struct {
	Address []string `json:"address"` // 地址列表
}

// Organizations Organization
type Organizations struct {
	Organization []string `json:"organization"` // 组织列表
}

// Hashes 交易hash数组
// swagger:parameters Hashes
type Hashes struct {
	Hash []string `json:"hash"` // 哈希列表
}

type ServerResponse struct {
	ID     uint64      `json:"id"`     // 请求标识
	Result interface{} `json:"result"` // 返回结果
	Error  interface{} `json:"error"`  // 错误描述
}

type ClientRequest struct {
	Method string              `json:"method"` // 请求方法
	Params [1]*json.RawMessage `json:"params"` // 请求参数
	ID     uint64              `json:"id"`     // 请求标识
}

type CountByTime struct {
	Match  []*QMatch `json:"match"`  // 匹配条件
	Ranges *QRanges  `json:"ranges"` // 范围
}

// SpecifiedFields 用于获取指定字段的值，例如获取某个省份某个单位志愿者的姓名
type SpecifiedFields struct {
	Match  []*QMatch `json:"match"`  // 匹配条件
	Sort   []*QSort  `json:"sort"`   // 排序
	Count  int       `json:"count"`  // 总量
	Fields []string  `json:"fields"` // 字段列表
}

// TotalStats 用于统计指定字段值的总和
type TotalStats struct {
	Match  []*QMatch  `json:"match"`  // 匹配条件
	SumAgg *QMatchKey `json:"sumAgg"` // 聚合字段
}

// DonationStats DonationStats 用于统计指定字段的捐款并排名
type DonationStats struct {
	Match     []*QMatch  `json:"match"`     // 匹配条件
	TermsAgg  *QMatchKey `json:"termsAgg"`  // 聚合字段
	SubSumAgg *QMatchKey `json:"subSumAgg"` // 子聚合字段
}

// RepDonationStat 捐款排名统计结果：返回捐款人名以及捐款额度
type RepDonationStat struct {
	Name  string  `json:"name"`  // 名字
	Total float64 `json:"total"` // 总和
	Count int64   `json:"count"` // 个数
}

type RepDonationStates struct {
	Itemes []*RepDonationStat `json:"itemes"` // 项
}

// VolunteerStats 用于统计志愿者分布图：按照省分桶，再按照单位分桶
type VolunteerStats struct {
	Match       []*QMatch  `json:"match"`       // 匹配条件
	TermsAgg    *QMatchKey `json:"termsAgg"`    // 聚合字段
	SubTermsAgg *QMatchKey `json:"subTermsAgg"` // 子组合字段
	SubSumAgg   *QMatchKey `json:"subSumAgg"`   // 子统计字段
}

// RepVolunteerStat 志愿者分布图统计
type RepVolunteerStat struct {
	TermsAgges []*RepTermsAgg `json:"termsAgges"` // 聚合信息
	Count      int64          `json:"count"`      // 计数
}

// RepTermsAgg 省份信息:
type RepTermsAgg struct {
	Count         int64             `json:"count"`         // 计数
	TermsAggKey   string            `json:"termsAggKey"`   // 聚合字段
	SubTermsAgges []*RepSubTermsAgg `json:"subTermsAgges"` // 子聚合字段
}

// RepSubTermsAgg 单位
type RepSubTermsAgg struct {
	SubTermsAggKey string `json:"subTermsAggKey"` // 子聚合字段
	Count          int64  `json:"count"`          // 计数
}

// RepLastSeq 最新的seq序列号
type RepLastSeq struct {
	LastSyncSeq    int64 `json:"lastSyncSeq"`    // 最新同步区块高度
	LastConvertSeq int64 `json:"lastConvertSeq"` // 最新解析区块高度
}

// QMatchKey key
type QMatchKey struct {
	Key string `json:"key"` // 字段名
}

// swagger:parameters QMatch
type QMatch struct {
	Key   string      `json:"key"`   // 字段名
	Value interface{} `json:"value"` // 值
}

// swagger:parameters QPage
type QPage struct {
	Size   int `json:"size"`   // 大小
	Number int `json:"number"` // 当前页数
}

// swagger:parameters QSort
type QSort struct {
	Key       string `json:"key"`       // 字段名
	Ascending bool   `json:"ascending"` // 是否升序
}

// swagger:parameters QRange
type QRange struct {
	Key    string      `json:"key"`   // 字段名
	RStart interface{} `json:"start"` // 开始位置（大于等于）
	REnd   interface{} `json:"end"`   // 结束位置（小于等于）
	GT     interface{} `json:"gt"`    // 大于
	LT     interface{} `json:"lt"`    // 小于
}

// swagger:parameters QRanges
type QRanges struct {
	Key    string   `json:"key"`    // 字段名
	Ranges []*Range `json:"ranges"` // 范围
}

type Range struct {
	RStart interface{} `json:"start"` // 开始位置（大于等于）
	REnd   interface{} `json:"end"`   // 结束位置（小于等于）
}
