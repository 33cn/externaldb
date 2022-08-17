package proof

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserJSON(t *testing.T) {
	testKV := "{\"key\":\"testkv_key\",\"data\":{\"value\":\"testkv_value\",\"type\":\"text\",\"format\":\"string\"}}"
	testData := "{\"key\":\"testkv_datakey\",\"data\":{\"value\":\"2006-01-02 15:04:05\",\"type\":\"date\",\"format\":\"rfc\"}}"
	testTimestamp := int64(1136214245)

	testArray := "{\"key\":\"testkv_arraykey\",\"data\":[{\"value\":\"arr01\",\"type\":\"text\",\"format\":\"string\"}, {\"value\":\"arr02\",\"type\":\"text\",\"format\":\"string\"}, {\"value\":\"arr03\",\"type\":\"text\",\"format\":\"string\"}]}"

	testObject := "{\"key\":\"testkv_objectkey\",\"data\":[             {\"key\": \"objk01\", \"data\" : {\"value\":\"objv01\",\"type\":\"text\",\"format\":\"string\"}},                                             {\"key\": \"objk02\", \"data\" : {\"value\":\"objv02\",\"type\":\"text\",\"format\":\"string\"}},                                              {\"key\": \"objk03\", \"data\" : {\"value\":\"objv03\",\"type\":\"text\",\"format\":\"string\"}}] }"

	testInput := "[" + testKV + ", " + testData + "," + testArray + "," + testObject + "]"

	result := ParserJSON(testInput)

	assert.Equal(t, "testkv_value", result["testkv_key"].(string))
	assert.Equal(t, testTimestamp, result["testkv_datakey"].(int64))

	arrayResult := result["testkv_arraykey"].([]interface{})
	assert.Equal(t, "arr01", arrayResult[0].(string))
	assert.Equal(t, "arr02", arrayResult[1].(string))
	assert.Equal(t, "arr03", arrayResult[2].(string))

	// 无嵌套, 解析后, 内部的key会变成外部
	//objResult := result["testkv_objectkey"].(map[string]interface{})
	assert.Equal(t, "objv01", result["objk01"].(string))
	assert.Equal(t, "objv02", result["objk02"].(string))
	assert.Equal(t, "objv03", result["objk03"].(string))

}

func TestParseObjToKV(t *testing.T) {
	cases := []struct {
		key   string
		input map[string]interface{}
	}{
		{
			key: "key1",
			input: map[string]interface{}{
				"value":  int32(3),
				"format": "int32",
				"type":   "number",
			},
		},
		{
			key: "key2",
			input: map[string]interface{}{
				"value":  float32(3.3),
				"format": "float32",
				"type":   "number",
			},
		},
		{
			key: "key3",
			input: map[string]interface{}{
				"value":  "3.3",
				"format": "string",
				"type":   "text",
			},
		},
	}
	for _, c := range cases {
		v, isKV := parserObjToKV(c.key, c.input)
		assert.Equal(t, true, isKV)

		result := v.(map[string]interface{})
		assert.Equal(t, c.input["value"], result[c.key])
	}

}
