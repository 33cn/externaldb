package main

import (
	"testing"

	"github.com/33cn/externaldb/db/block"
	"github.com/33cn/externaldb/escli/querypara"
)

func getLastSeq(prefix, id string) int64 {
	syncSeq := GetLastSeq("http://172.16.101.87:9200", prefix, id, 7, "elastic", "elastic")
	return syncSeq
}

func TestGetLastSeq(t *testing.T) {
	lastSeq := LastSeq{}
	lastSeq.LastSyncSeq = getLastSeq("v12seq01_", block.SyncSeq)
	lastSeq.LastConvertSeq = getLastSeq("v12db02_", "convert_bty3")

	t.Log(lastSeq)
}

func TestNormalizeAddrInQuery_Nil(t *testing.T) {
	// nil query 不应 panic
	normalizeAddrInQuery(nil)
}

func TestNormalizeAddrInQuery_LowercaseAddr(t *testing.T) {
	// 大写混合的 0x 地址应转为小写
	q := &querypara.Query{
		Match: []*querypara.QMatch{
			{Key: "from", Value: "0xAbCdEf1234567890AbCdEf1234567890AbCdEf12"},
			{Key: "to", Value: "0xAbCDeFAbCDeFAbCDeFAbCDeFAbCDeFAbCDeFAbCDeF"},
		},
	}
	normalizeAddrInQuery(q)

	from := q.Match[0].Value.(string)
	to := q.Match[1].Value.(string)

	expectedFrom := "0xabcdef1234567890abcdef1234567890abcdef12"
	expectedTo := "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdef"

	if from != expectedFrom {
		t.Errorf("from: got %q, want %q", from, expectedFrom)
	}
	if to != expectedTo {
		t.Errorf("to: got %q, want %q", to, expectedTo)
	}
}

func TestNormalizeAddrInQuery_AlreadyLowercase(t *testing.T) {
	// 已小写的地址不变
	q := &querypara.Query{
		Match: []*querypara.QMatch{
			{Key: "from", Value: "0xabcdef1234567890abcdef1234567890abcdef12"},
		},
	}
	normalizeAddrInQuery(q)
	if q.Match[0].Value.(string) != "0xabcdef1234567890abcdef1234567890abcdef12" {
		t.Error("already lowercase addr should not change")
	}
}

func TestNormalizeAddrInQuery_NonAddrKey(t *testing.T) {
	// 非地址 key 的 0x 值不应被修改（如 tx hash）
	q := &querypara.Query{
		Match: []*querypara.QMatch{
			{Key: "hash", Value: "0xAbCdEf1234567890AbCdEf1234567890AbCdEf1234567890AbCdEf1234567890"},
			{Key: "height", Value: "0x1A"},
		},
	}
	normalizeAddrInQuery(q)

	if q.Match[0].Value.(string) != "0xAbCdEf1234567890AbCdEf1234567890AbCdEf1234567890AbCdEf1234567890" {
		t.Error("non-addr key 'hash' should not be lowercased")
	}
	if q.Match[1].Value.(string) != "0x1A" {
		t.Error("non-addr key 'height' should not be lowercased")
	}
}

func TestNormalizeAddrInQuery_NonHexValue(t *testing.T) {
	// 非 0x 开头的地址值不应被修改（如 chain33 base58 地址）
	q := &querypara.Query{
		Match: []*querypara.QMatch{
			{Key: "from", Value: "1Q5QcUaDXET3RJ3UBurMZzF3gGHyjnFQEa"},
			{Key: "to", Value: "19ozyoUGPAQ9spsFiz9CJfnUCFeszpaFuF"},
		},
	}
	normalizeAddrInQuery(q)

	if q.Match[0].Value.(string) != "1Q5QcUaDXET3RJ3UBurMZzF3gGHyjnFQEa" {
		t.Error("non-0x addr should not be lowercased")
	}
	if q.Match[1].Value.(string) != "19ozyoUGPAQ9spsFiz9CJfnUCFeszpaFuF" {
		t.Error("non-0x addr should not be lowercased")
	}
}

func TestNormalizeAddrInQuery_AllFields(t *testing.T) {
	// Match/MatchOne/Filter/Not 各字段都需处理
	q := &querypara.Query{
		Match:    []*querypara.QMatch{{Key: "from", Value: "0xAAA"}},
		MatchOne: []*querypara.QMatch{{Key: "to", Value: "0xBBB"}},
		Filter:   []*querypara.QMatch{{Key: "contract_addr", Value: "0xCCC"}},
		Not:      []*querypara.QMatch{{Key: "owner_addr", Value: "0xDDD"}},
	}
	normalizeAddrInQuery(q)

	if q.Match[0].Value.(string) != "0xaaa" {
		t.Error("Match from not lowercased")
	}
	if q.MatchOne[0].Value.(string) != "0xbbb" {
		t.Error("MatchOne to not lowercased")
	}
	if q.Filter[0].Value.(string) != "0xccc" {
		t.Error("Filter contract_addr not lowercased")
	}
	if q.Not[0].Value.(string) != "0xddd" {
		t.Error("Not owner_addr not lowercased")
	}
}

func TestNormalizeAddrInQuery_NilMatch(t *testing.T) {
	// nil QMatch 元素不应 panic
	q := &querypara.Query{
		Match: []*querypara.QMatch{
			nil,
			{Key: "from", Value: "0xAAA"},
			nil,
		},
	}
	// 不应 panic
	normalizeAddrInQuery(q)

	if q.Match[1].Value.(string) != "0xaaa" {
		t.Error("addr after nil element not lowercased")
	}
}

func TestNormalizeAddrInQuery_EmptyQuery(t *testing.T) {
	// 空 query 不应 panic
	q := &querypara.Query{}
	normalizeAddrInQuery(q)
}

func TestNormalizeAddrInQuery_AllAddrKeys(t *testing.T) {
	// 验证所有地址 key 都会处理
	q := &querypara.Query{}
	keys := []string{"from", "to", "contract_addr", "contract_address", "owner_addr", "owner_address", "creator"}
	for _, k := range keys {
		q.Match = append(q.Match, &querypara.QMatch{Key: k, Value: "0x" + k})
	}
	normalizeAddrInQuery(q)

	for i, m := range q.Match {
		if s, ok := m.Value.(string); ok && s != "0x"+keys[i] {
			t.Errorf("key %q: got %q, want %q", keys[i], s, "0x"+keys[i])
		}
	}
}

func TestNormalizeAddrInQuery_NonStringValue(t *testing.T) {
	// 非 string 类型的 value 不应 panic
	q := &querypara.Query{
		Match: []*querypara.QMatch{
			{Key: "from", Value: 123},
			{Key: "to", Value: nil},
		},
	}
	// 不应 panic
	normalizeAddrInQuery(q)
}

func TestNormalizeAddrInQuery_SubQuery(t *testing.T) {
	// SubQuery 中的地址也需要处理
	q := &querypara.Query{
		Match: []*querypara.QMatch{
			{Key: "from", Value: "0xAAA"},
			{Key: "to", Value: "0xBBB", SubQuery: &querypara.Query{
				Match: []*querypara.QMatch{
					{Key: "contract_addr", Value: "0xCCC"},
				},
			}},
		},
	}
	normalizeAddrInQuery(q)

	if q.Match[0].Value.(string) != "0xaaa" {
		t.Error("top-level from not lowercased")
	}
	if q.Match[1].Value.(string) != "0xbbb" {
		t.Error("top-level to not lowercased")
	}
	// SubQuery 中的地址也应处理
	if q.Match[1].SubQuery.Match[0].Value.(string) != "0xccc" {
		t.Error("SubQuery contract_addr not lowercased")
	}
}
