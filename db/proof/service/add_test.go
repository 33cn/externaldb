package service

import (
	"encoding/base64"
	"testing"

	dbcom "github.com/33cn/externaldb/db/common"
	proofconfig "github.com/33cn/externaldb/db/proof_config"
)

/*func Test_parseDataByVersion(t *testing.T) {
	convert := NewConvert("user.p.testproof.", "bty", "proof")
	c := convert.(*ProofConvert)
	c.RecordGen = NewRecordGen("p1", "p2", "l1", "l2", "t1", "t2")
	c.configDB = &ConfigDB{}
	c.proofDB = &myProofDB{}

	type args struct {
		height  int64
		title   string
		version string
		option  string
		data    string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			name: "test parse v3.0.0",
			args: args{
				height:  0,
				title:   "",
				version: "V3.0.0",
				option:  "[{\"key\":\"template\",\"value\":\"id123\"}]",
				data:    `{"ext":{"basehash":"null","prehash":"null","存证名称":"相册","存证类型":"相册"},"相册":{"照片描述":"","相册":["4bb42ecfe15a99ee586031b7a5a20a5d15c4364780f5ea14d278095d620619f5","b45f6c7728e9b59f9ad381f36c29945317ba646e49ad3b73511ec1731c22dff4"]}}`,
			},
			want:  "",
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := parseDataByVersion(c.proofDB, tt.args.height, tt.args.title, tt.args.version, tt.args.option, tt.args.data)
			t.Log("got record", "record", got)
			if got1 != tt.want1 {
				t.Errorf("parseDataByVersion() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}*/

func TestDataParser_ParseDataByVersion(t *testing.T) {
	convert := NewConvert("user.p.testproof.", "bty", "proof")
	c := convert.(*ProofConvert)
	c.RecordGen = NewRecordGen("p1", "p2", "l1", "l2", "t1", "t2", "u1", "u2")
	c.configDB = &proofconfig.None{}
	c.proofDB = &myProofDB{}

	dataP := &dataParseWithTemplate{
		version: "V3.0.0",
		option:  "[{\"key\":\"template\",\"value\":\"id123\"}]",
		title:   "",
		height:  0,
		data:    `{"ext":{"basehash":"null","prehash":"null","存证名称":"相册","存证类型":"相册"},"相册":{"照片描述":"","相册":["4bb42ecfe15a99ee586031b7a5a20a5d15c4364780f5ea14d278095d620619f5","b45f6c7728e9b59f9ad381f36c29945317ba646e49ad3b73511ec1731c22dff4"]}}`,
		db:      c.proofDB,
	}
	tests := []struct {
		name  string
		want  string
		dataP *dataParseWithTemplate
		want1 bool
	}{
		{
			name:  "test parse v3.0.0",
			want:  "",
			dataP: dataP,
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.dataP.parseDataByVersion()
			//got, got1 := parseDataByVersion(c.proofDB, tt.args.height, tt.args.title, tt.args.version, tt.args.option, tt.args.data)
			t.Log("got record", "record", got)
			if got1 != tt.want1 {
				t.Errorf("parseDataByVersion() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_parseProofV3(t *testing.T) {
	jsonstr := "{\"ext\":{\"basehash\":\"null\",\"prehash\":\"null\",\"存证名称\":\"相册\",\"存证类型\":\"相册\"},\"相册\":{\"照片描述\":\"\",\"相册\":[\"4bb42ecfe15a99ee586031b7a5a20a5d15c4364780f5ea14d278095d620619f5\",\"b45f6c7728e9b59f9ad381f36c29945317ba646e49ad3b73511ec1731c22dff4\"]}}"
	enJSONStr, _ := dbcom.PublicEncrypt(rsaPubKey, jsonstr)
	//base编码
	xx := dbcom.EncodeToString([]byte(enJSONStr))

	optionStr := `[{"key":"encrypt","value":"rsa"},{"key":"template","value":"id123"}]`
	enOptionStr := dbcom.EncodeToString([]byte(optionStr))

	notestr := "{\"userName\":\"18701358726\",\"userIcon\":\"\",\"evidenceName\":\"公益捐赠yyh\",\"stepName\":\"\",\"version\":1}"
	enNoteStr := dbcom.EncodeToString([]byte(notestr))

	extstr := "{\"basehash\":\"1\",\"prehash\":\"2\"}"
	enextstr := dbcom.EncodeToString([]byte(extstr))

	payload := "{\"version\":\"V3.0.0\",\"option\":\"" + enOptionStr + "\",\"data\":\"" + xx + "\",\"note\":\"" + enNoteStr + "\",\"ext\":\"" + enextstr + "\"}"

	t.Log("Test_Proof", "payload", payload)

	convert := NewConvert("user.p.testproof.", "bty", "proof")
	c := convert.(*ProofConvert)
	c.RecordGen = NewRecordGen("p1", "p2", "l1", "l2", "t1", "t2", "u1", "u2")
	c.configDB = &proofconfig.None{}
	c.proofDB = &myProofDB{}

	mapinfo, _ := parseProof(c.proofDB, "", int64(1), []byte(payload))
	t.Log(mapinfo)
	for k, v := range mapinfo {
		t.Log(k, v)
	}
}

func Test_parseProofV1(t *testing.T) {
	jsonstr := `[{"label":"相册","key":"","type":3,"data":[{"data":[{"type":"image","format":"hash","value":"4bb42ecfe15a99ee586031b7a5a20a5d15c4364780f5ea14d278095d620619f5"},{"type":"image","format":"hash","value":"b45f6c7728e9b59f9ad381f36c29945317ba646e49ad3b73511ec1731c22dff4"}],"type":1,"key":"","label":"相册"},{"data":{"type":"text","format":"string","value":""},"type":0,"key":"","label":"照片描述"}]},{"label":"ext","key":"","type":3,"data":[{"data":{"type":"text","format":"string","value":"相册"},"type":0,"key":"存证名称","label":"存证名称"},{"data":{"type":"text","format":"string","value":"null"},"type":0,"key":"basehash","label":"basehash"},{"data":{"type":"text","format":"string","value":"null"},"type":0,"key":"prehash","label":"prehash"},{"data":{"type":"text","format":"string","value":"相册"},"type":0,"key":"存证类型","label":"存证类型"}]}]`
	//base编码
	xx := base64.StdEncoding.EncodeToString([]byte(jsonstr))

	optionStr := `{"encrypt":"rsa"}`
	enOptionStr := dbcom.EncodeToString([]byte(optionStr))

	notestr := "{\"userName\":\"18701358726\",\"userIcon\":\"\",\"evidenceName\":\"公益捐赠yyh\",\"stepName\":\"\",\"version\":1}"
	enNoteStr := dbcom.EncodeToString([]byte(notestr))

	// extstr := "{\"basehash\":\"1\",\"prehash\":\"2\"}"
	// enextstr := dbcom.EncodeToString([]byte(extstr))

	payload := "{\"version\":\"1.0.0\",\"option\":\"" + enOptionStr + "\",\"data\":\"" + xx + "\",\"note\":\"" + enNoteStr + "\"}"

	t.Log("Test_Proof", "payload", payload)

	convert := NewConvert("user.p.testproof.", "bty", "proof")
	c := convert.(*ProofConvert)
	c.RecordGen = NewRecordGen("p1", "p2", "l1", "l2", "t1", "t2", "u1", "u2")
	c.configDB = &proofconfig.None{}
	c.proofDB = &myProofDB{}

	mapinfo, _ := parseProof(c.proofDB, "", int64(1), []byte(payload))
	t.Log(mapinfo)
	for k, v := range mapinfo {
		t.Log(k, v)
	}
}
