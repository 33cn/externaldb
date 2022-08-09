package proof

//解析json字符串并保存原有的嵌套关系
//parserArray解析数组，支持格式[{v},{k:v},{k:v}...]
func parserArrayToNestKV(array []interface{}) []interface{} {

	//log.Info("parserArray-begin", "array", array)

	var mapinfo []interface{}

	for _, value := range array {
		switch value := value.(type) {
		case map[string]interface{}: // 对象数组模式
			inf := parserObjToNestKV(false, value)
			if inf != nil {
				mapinfo = append(mapinfo, inf)
			}
		default:
		}
	}
	//log.Info("parserArray-end", "mapinfo", mapinfo)

	return mapinfo
}

//解析object对象数据
//主要分 data obj 和 k:v对obj。
func parserObjToNestKV(isFirstLayer bool, inputobj map[string]interface{}) interface{} {
	//log.Info("parserObj-begin", "isFirstLayer", isFirstLayer, "inputobj", inputobj)

	mapinfo := make(map[string]interface{})

	var keyinfo = ""
	if inputobj["key"] != nil {
		keyinfo = (inputobj["key"]).(string)
	}

	//log.Info("parserObj", "keyinfo", keyinfo)

	//obj是一个完整的key：data结构
	valuestr := ""
	formatstr := ""
	typestr := ""
	for key, value := range inputobj {
		switch value := value.(type) {
		case string: // 单个data obj类型
			if key == "key" {
				keyinfo = value
			} else if key == "value" {
				valuestr = value
			} else if key == "format" {
				formatstr = value
			} else if key == "type" {
				typestr = value
			}
			//解析data对象中的具体数据
			if formatstr != "" && valuestr != "" && typestr != "" {
				parserFunc := loadParserFunc(formatstr)
				if parserFunc != nil {
					kv := parserFunc(valuestr, formatstr, typestr)
					if !isFirstLayer && keyinfo != "" && kv != nil {
						mapinfo[keyinfo] = kv
						return mapinfo
					} else if kv != nil {
						return kv
					}
				}
			}
		case []interface{}: //对象中嵌套数组类型
			inf := parserArrayToNestKV(value)
			if !isFirstLayer && keyinfo != "" && inf != nil {
				mapinfo[keyinfo] = inf
				return mapinfo
			} else if inf != nil {
				return inf
			}

		case map[string]interface{}: //对象中嵌套对象模式
			inf := parserObjToNestKV(false, value)
			if !isFirstLayer && keyinfo != "" && inf != nil {
				mapinfo[keyinfo] = inf
				return mapinfo
			} else if inf != nil {
				return inf
			}

		default:
			log.Error("parserObj type error!", "key", key, "value", value)
		}
	}

	return nil
}
