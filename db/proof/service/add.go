package service

import (
	"encoding/json"
	"fmt"

	"github.com/33cn/externaldb/db/proof/proofdb"
	proofUtil "github.com/33cn/externaldb/util/proof"

	dbcom "github.com/33cn/externaldb/db/common"
	"github.com/33cn/externaldb/db/proof/api"
	parse "github.com/33cn/proofparse"
)

// 按格式解析存证
func parseProof(db proofdb.IProofDB, title string, height int64, input []byte) (map[string]interface{}, error) {
	var payload api.ProofInfo
	err := json.Unmarshal(input, &payload)
	if err != nil {
		log.Error("AddProof:payload:Unmarshal", "err", err)
		return nil, ErrParams
	}

	//payload.Note base64格式解码,可以不填写
	note, err := dbcom.DecodeString(payload.Note)
	if err != nil {
		log.Error("AddProof:DecodeString:payload.Note", "err", err)
		return nil, ErrDecode
	}

	//payload.Data base64格式解码
	data, err := dbcom.DecodeString(payload.Data)
	if err != nil {
		log.Error("AddProof:DecodeString:payload.Data", "err", err)
		return nil, ErrDecode
	}

	//payload.Option base64格式解码
	option, err := dbcom.DecodeString(payload.Option)
	if err != nil {
		log.Error("AddProof:DecodeString:payload.Option", "err", err)
		return nil, ErrDecode
	}

	// 根据version以及option解析存证信息
	//proofData, parseOk := parseDataByVersion(db, height, title, payload.Version, string(option), string(data))
	var dataP DataParser
	log.Debug("AddProof:DecodeString:payload.version", "version", payload.Version)

	//格式化version
	payload.Version, err = parse.FormatVersion(payload.Version)
	if err != nil {
		log.Error("parse FormatVersion failed", "err", err)
	}

	switch payload.Version {
	case parse.Version1:
		dataP = &dataParseV1{
			version: payload.Version,
			option:  string(option),
			data:    string(data),
			title:   title,
			height:  height,
		}
	case parse.Version2:
		return nil, fmt.Errorf("version is custom , version:" + parse.Version2)
	case parse.Version3, parse.Version4:
		dataP = &dataParseWithTemplate{
			version: payload.Version,
			option:  string(option),
			data:    string(data),
			title:   title,
			height:  height,
			db:      db,
		}
	default:
		panic("not support version" + payload.Version)
	}

	proofData, parseOk := dataP.parseDataByVersion()

	//填写展开后的信息
	mapinfo := make(map[string]interface{})

	mapinfo["proof_original"] = string(data)
	mapinfo["proof_data"] = proofData
	mapinfo["proof_note"] = string(note)

	//若设置了note，则对note值进行解析
	if note != nil {
		err = json.Unmarshal(note, &mapinfo)
		if err != nil {
			log.Error("Unmarshal note failed", "err", err)
		}
	}

	//尝试解析payload中的jsonstr
	if parseOk {
		parseInfo := proofUtil.ParserJSON(proofData)
		for k, v := range parseInfo {
			mapinfo[k] = v
		}
	}

	//若设置了ext，则对ext的值进行解析
	//payload.Ext base64格式解码
	if payload.Ext != "" {
		ext, err := dbcom.DecodeString(payload.Ext)
		if err != nil {
			log.Error("AddProof:DecodeString:payload.Ext", "err", err)
			return nil, ErrDecode
		}

		err = json.Unmarshal(ext, &mapinfo)
		if err != nil {
			log.Error("Unmarshal ext failed", "err", err)
		}
	}

	return mapinfo, nil
}

type dataParseV1 struct {
	version string
	option  string
	data    string
	title   string
	height  int64
}

type dataParseWithTemplate struct {
	version string
	option  string
	data    string
	title   string
	height  int64
	db      proofdb.IProofDB
}

// type dataParseV3 dataParseWithTemplate
// type dataParseV4 dataParseWithTemplate

type DataParser interface {
	parseDataByVersion() (string, bool)
}

// parseDataByVersion:通过version以及option解析存证data数据
// 默认版本中只支持加密功能,获取加密类型并解密
func (dP *dataParseV1) parseDataByVersion() (string, bool) {
	log.Info("parseDataByVersion", "version", dP.version, "option", dP.option)
	opt := make(map[string]interface{})
	if err := json.Unmarshal([]byte(dP.option), &opt); err != nil {
		return dP.data, true
	}
	encrypType := opt[dbcom.Encrypt]
	if encrypType == nil || encrypType.(string) == "" {
		return dP.data, true
	}
	return decryptProofData(dP.height, dP.title, encrypType.(string), dP.data)
}

func (dP *dataParseWithTemplate) parseDataByVersion() (string, bool) {
	log.Info("parseDataByVersion", "version", dP.version, "option", dP.option)
	opts := make([]map[string]interface{}, 0)
	if err := json.Unmarshal([]byte(dP.option), &opts); err != nil {
		return dP.data, true
	}
	for _, opt := range opts {
		switch opt["key"] {
		case dbcom.Template:
			tmpTx := opt["value"]
			log.Debug("tmpTx is", "tmpTx", tmpTx)
			if tmpTx == nil || tmpTx.(string) == "" {
				return dP.data, true
			}
			comleteData, parseOk := combineProofData(dP.db, tmpTx.(string), dP.version, dP.data)
			if comleteData != dP.data {
				dP.data = comleteData
			} else {
				return comleteData, parseOk
			}
		case dbcom.Encrypt:
			encrypType := opt["value"]
			log.Debug("encrypType is", "encrypType", encrypType)
			if encrypType == nil || encrypType.(string) == "" {
				return dP.data, true
			}
			return decryptProofData(dP.height, dP.title, encrypType.(string), dP.data)
		default:
			log.Debug("undefined option type", "type", opt["key"])
			return dP.data, true
		}
	}
	return dP.data, true
}

//combineProofData 将存证数据组合起来成为一条完整的数据
func combineProofData(db proofdb.IProofDB, tmpTx, version, data string) (string, bool) {
	//find template content
	template, err := db.GetTemplate(TemplateID(tmpTx))
	if err != nil {
		log.Error("GetTemplate failed", "err", err, "template id", TemplateID(tmpTx))
		return data, true
	}
	tmp := template.Template["template_data"]
	log.Info("convert complete data by template", "tmp", tmp, "data", data, "version", version)

	p := parse.NewProof("", tmp.(string), data, version)
	err = p.ContentToComleteData()
	if err != nil {
		log.Debug("ContentToComleteData failed", "err", err)
		return data, true
	}

	data = p.ComleteData
	return data, true
}

//decryptProofData 解密proof的加密信息
func decryptProofData(height int64, title string, encrypType string, data string) (string, bool) {
	privKey := getPrivetKey(title, height, encrypType)
	if encrypType == dbcom.RsaCrypto {
		decryptData, err := dbcom.PrivateDecrypt(privKey, data)
		if err != nil {
			return data, false
		}
		return decryptData, true
	}
	return data, false
}

//getPrivetKey  获取对应title在对应高度上加密算法的私钥，查询数据库或者从配置获取
func getPrivetKey(title string, height int64, encrypType string) string {
	var defaultRsaPrivKey = "0x3078020100300b06092a864886f70d01010104663064020100021100ce8fee27c53555fc9bd4fd40e96fc5990203010001021100a418f3b9e4915a9cc648719de8926f81020900e2e6f864dd1ca567020900e90d368c9e1a5cff020900cd53674996d13a57020900ca445d83cdf4b3a1020863997caaaf1c973a"

	return defaultRsaPrivKey
}

// 按格式解析模板
func parseTemplate(title string, height int64, input []byte) (map[string]interface{}, error) {
	var payload api.TemplateInfo
	err := json.Unmarshal(input, &payload)
	if err != nil {
		log.Error("AddProof:payload:Unmarshal", "err", err)
		return nil, ErrParams
	}

	//填写展开后的信息
	mapinfo := make(map[string]interface{})

	mapinfo["template_name"] = payload.Name
	mapinfo["template_data"] = payload.Data
	return mapinfo, nil
}
