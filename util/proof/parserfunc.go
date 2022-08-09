package proof

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	l "github.com/xuperchain/log15"
)

var log = l.New("module", "db.proof")

var (
	isKVpair    = true
	isNotKVpair = false
)

// InputMapInfo InputMapInfo
type InputMapInfo struct {
	Key      string
	InputObj map[string]interface{}
}

var (
	timeLayout = "2006-01-02 15:04:05"
	//type formt
	parserFunc = make(map[string]CreateFunc)
	// IsParserToKV 目前支持解析成嵌套模式，或者非嵌套模式的kv模式，默认支持非嵌套的kv模式
	IsParserToKV = true
)

// CreateFunc create parse func
type CreateFunc func(value interface{}, format string, typ string) interface{}

/*
复合的项值出现在json中, 可以使json的 列表 (数组)和结构 (嵌的结构)
单独个项值类型举例
1. 数字. number int64,int32,float32
2. 文本. text. string, http, json, base64
3. 时间. date. rfc-3339 iso-8601 utc
	1. rfc-3339 2006-08-14 02:34:56-06:00
	2. iso-8601 2006-08-14T02:34:56-0600
	3. utc 1584945631
4. 图片. image. 如jpg, png, url, hash
	1. jpg xxxxxxxxxx(图片内容)
	2. url http://xxxxx/1.jpg
	3. hash 0x1234
5. 文件. file. pdf, url, hash 等
	1. pdf xxxxxx(文件内容)
	2. url http://xxxxx/1.pdf
	3. hash 0x1122
*/

//数据type以及format常量的定义
const (
	//Number Format定义
	Format_int     = "int"
	Format_int64   = "int64"
	Format_int32   = "int32"
	Format_float32 = "float32"
	Format_float64 = "float64"

	//Text Format定义
	Format_string = "string"
	Format_http   = "http"
	Format_json   = "json"
	Format_base64 = "base64"

	//Date Format定义
	Format_rfc = "rfc"
	Format_iso = "iso"
	Format_utc = "utc"

	//Image Format定义
	Format_jpg  = "jpg"
	Format_png  = "png"
	Format_url  = "url"
	Format_hash = "hash"

	//File Format定义
	Format_pdf = "pdf"
)

//启动时默认注册解析数据的接口
func init() {
	registerParseFunc(Format_int, parserNumber)
	registerParseFunc(Format_int32, parserNumber)
	registerParseFunc(Format_int64, parserNumber)
	registerParseFunc(Format_float32, parserNumber)
	registerParseFunc(Format_float64, parserNumber)

	registerParseFunc(Format_string, parserToStr)
	registerParseFunc(Format_http, parserToStr)
	registerParseFunc(Format_json, parserToStr)
	registerParseFunc(Format_base64, parserToStr)
	registerParseFunc(Format_url, parserToStr)
	registerParseFunc(Format_hash, parserToStr)

	registerParseFunc(Format_rfc, parserDate)
	registerParseFunc(Format_iso, parserDate)

}

// register 注册对应数据类型的解析函数
func registerParseFunc(formt string, create CreateFunc) {
	parserFunc[formt] = create
}

// load 加载对应数据类型的解析函数
func loadParserFunc(formt string) CreateFunc {
	call, ok := parserFunc[formt]
	if ok {
		return call
	}
	return nil
}

// ParserJSON 将json字符串数据解析成kv的map格式方便通过指定key搜索
//目前只支持格式[{k:v},{k:v},{k:v}...]格式的jsonstr
func ParserJSON(jsonstr string) map[string]interface{} {

	//log.Info("parserJson-begin", "jsonstr", jsonstr)

	var payloadinf []interface{}
	err := json.Unmarshal([]byte(jsonstr), &payloadinf)
	if err != nil {
		log.Error("parserJson", "err", err)
		return nil
	}

	mapinfo := make(map[string]interface{})

	//将数据分发到多个chan并行处理
	in := distributeData(payloadinf)
	parser1 := parserObjData(in)
	parser2 := parserObjData(in)
	parser3 := parserObjData(in)
	parser4 := parserObjData(in)

	//合并解析之后的数据到一个chan
	for info := range mergeMapInfo(parser1, parser2, parser3, parser4) {
		for k, v := range info {
			mapinfo[k] = v
		}
	}
	//log.Info("parserJson---end", "mapinfo", mapinfo)

	return mapinfo
}

//将数据分发到chan
func distributeData(payloadinf []interface{}) <-chan InputMapInfo {

	out := make(chan InputMapInfo)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("distributeData panic error", "err", r)
			}
		}()
		defer close(out)
		for _, obj := range payloadinf {
			var key string
			objInfo := obj.(map[string]interface{})
			if objInfo["key"] != nil {
				key = objInfo["key"].(string)
			}
			out <- InputMapInfo{InputObj: objInfo, Key: key}
		}
	}()
	return out
}

//将数据解析成map[string]interface{}类型
func parserObjData(inCh <-chan InputMapInfo) <-chan map[string]interface{} {

	out := make(chan map[string]interface{})

	go func() {
		defer close(out)
		for input := range inCh {
			if IsParserToKV {
				kvinfo, iskvpair := parserObjToKV("", input.InputObj)
				if iskvpair {
					out <- kvinfo.(map[string]interface{})
				}
			} else {
				inf := parserObjToNestKV(true, input.InputObj)
				if inf != nil && input.Key != "" {
					mapinfo := make(map[string]interface{})
					mapinfo[input.Key] = inf
					out <- mapinfo
				}
			}

		}
	}()
	return out

}

//将解析收后的数据从多个in chan合并到一个 out chan
func mergeMapInfo(mapinfo ...<-chan map[string]interface{}) <-chan map[string]interface{} {

	out := make(chan map[string]interface{}, 4)

	var wg sync.WaitGroup

	collect := func(in <-chan map[string]interface{}) {
		defer wg.Done()
		for info := range in {
			out <- info
		}
	}

	wg.Add(len(mapinfo))

	for _, c := range mapinfo {
		go collect(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

//解析number接口函数
func parserNumber(value interface{}, format string, typ string) interface{} {
	switch value := value.(type) {
	case string:
		return parserNumberFormString(value, format, typ)
	default: // int32, int64, float32, ...
		return value
	}
}
func parserNumberFormString(value string, format string, typ string) interface{} {

	if format == Format_int32 {
		num, err := strconv.ParseInt(value, 10, 32)
		if err == nil {
			return num
		}
	} else if format == Format_int64 {
		num, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return num
		}
	} else if format == Format_float64 || format == Format_float32 {
		num, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return num
		}
	} else {
		num, err := strconv.Atoi(value)
		if err == nil {
			return num
		}
	}
	return nil
}

//解析Text接口函数
//Text Format定义
//Format_string = "string"
//Format_http   = "http"
//Format_json   = "json"
//Format_base64 = "base64"
//Format_url  = "url"
//Format_hash = "hash"
func parserToStr(value0 interface{}, format string, typ string) interface{} {
	value, ok := value0.(string)
	if !ok {
		return nil
	}
	switch format {
	case Format_string, Format_http, Format_json, Format_base64, Format_url, Format_hash:
		return value
	default:
	}
	return nil
}

//Date Format定义
//Format_rfc = "rfc"
//Format_iso = "iso"
//Format_utc      = "utc"
//解析时间类型的数据
func parserDate(value0 interface{}, format string, typ string) interface{} {
	value, ok := value0.(string)
	if !ok {
		return nil
	}
	if format == Format_rfc {
		times, _ := time.Parse(timeLayout, value)
		timeUnix := times.Unix()
		return timeUnix
	} else if format == Format_iso {
		return value
	} else if format == Format_utc {
		return value
	}
	return nil
}

// test func
// func testTimeConvert() {
// 	// 获取时间，该时间带有时区等信息，获取的为当前地区所用时区的时间//2020-03-31T19:39:55+0800
// 	timeNow := time.Now()
// 	// 获取时间戳 //1585654795
// 	unix := time.Now().Unix()
// 	// 获取UTC时区的时间//2020-03-31T11:39:55+0000
// 	utcTime := time.Now().UTC()
// 	// time.Unix的第二个参数传递0或10结果一样，因为都不大于1e9//"2020-03-31 19:39:55"
// 	timeStr := time.Unix(unix, 0).Format(timeLayout)
//
// 	log.Info("testTimeConvert", "timeNow", timeNow, "unix", unix, "utcTime", utcTime, "timeStr", timeStr)
// }
