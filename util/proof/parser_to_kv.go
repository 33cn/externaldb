package proof

//解析json字符串最底层的kv对关系
//parserArray解析数组，支持格式[{k:v},{k:v}...]
func parserArrayToKV(array []interface{}) ([]interface{}, bool) {

	//log.Info("parserArrayToKV-begin", "array", array)

	var mapinfo []interface{}
	iskv := true
	for _, value := range array {
		switch value := value.(type) {
		case map[string]interface{}:
			inf, iskvpair := parserObjToKV("", value)
			if inf != nil {
				mapinfo = append(mapinfo, inf)
				iskv = iskvpair
			}
		default:
		}
	}
	//log.Info("parserArrayToKV-end", "mapinfo", mapinfo, "iskv", iskv)

	return mapinfo, iskv
}

//解析object对象数据: 需要区分 data obj 和 k:v对obj。
// 1. keyvalue: $key, inputobj: $data : { type: , format:, value: }
// 2. keyvalue: "",   inputobj: $kv:   {key:, type: lable:,  data: {} }
func parserObjToKV(keyvalue string, inputobj map[string]interface{}) (interface{}, bool) {
	//log.Info("parserObjToKV-begin", "keyvalue", keyvalue, "inputobj", inputobj)

	mapinfo := make(map[string]interface{})
	var keyinfo = keyvalue
	if inputobj["key"] != nil {
		keyinfo = (inputobj["key"]).(string)
	}

	//obj是一个完整的key：data结构
	var valuestr interface{}
	formatstr := ""
	typestr := ""
	for key, value := range inputobj {
		log.Info("parse", "key", key, "value", value)
		switch value := value.(type) {
		case string: // 单个data obj类型
			if key == "key" {
				keyinfo = value
			} else if key == "format" {
				formatstr = value
			} else if key == "type" {
				typestr = value
			} else if key == "value" {
				valuestr = value
			}
			//解析data对象中的具体数据
			if formatstr != "" && valuestr != nil && typestr != "" {
				parserfunc := loadParserFunc(formatstr)
				if parserfunc != nil {
					kv := parserfunc(valuestr, formatstr, typestr)
					if keyinfo != "" && kv != nil {
						mapinfo[keyinfo] = kv
						return mapinfo, isKVpair
					} else if kv != nil {
						return kv, isNotKVpair
					}
				}
			}
		case int, int32, int64, float32, float64: //单个data obj类型. 其他项都是string, 只有value 是才有可能是 数字
			if key != "value" {
				continue
			}
			if key == "value" {
				valuestr = value
			}
			//解析data对象中的具体数据
			if formatstr != "" && valuestr != nil && typestr != "" {
				parserfunc := loadParserFunc(formatstr)
				if parserfunc != nil {
					kv := parserfunc(valuestr, formatstr, typestr)
					if keyinfo != "" && kv != nil {
						mapinfo[keyinfo] = kv
						return mapinfo, isKVpair
					} else if kv != nil {
						return kv, isNotKVpair
					}
				}
			}
		case []interface{}: //对象中嵌套数组类型
			inf, iskv := parserArrayToKV(value)
			if keyinfo != "" && inf != nil && !iskv {
				mapinfo[keyinfo] = inf
				return mapinfo, isKVpair
			} else if inf != nil && iskv {
				for _, kv := range inf {
					for k, v := range kv.(map[string]interface{}) {
						mapinfo[k] = v
					}
				}
				return mapinfo, isKVpair
			}

		case map[string]interface{}: //对象中嵌套对象模式
			inf, iskv := parserObjToKV(keyinfo, value)
			if keyinfo != "" && inf != nil && !iskv {
				mapinfo[keyinfo] = inf
				return mapinfo, isKVpair
			} else if inf != nil && iskv { //retur k:v pair
				return inf, isKVpair
			}

		default:
			log.Error("parserObjToKV type error!", "key", key, "value", value)
		}
	}

	return nil, isNotKVpair
}
