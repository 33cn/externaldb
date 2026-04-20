package txinspect

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
)

// rpcErrFromRaw tolerates non-standard JSON-RPC error shapes (e.g. Chain33 / some
// forks return "error" as a string, or "" on success). go-ethereum's rpc.Client
// only accepts error objects and fails unmarshaling otherwise.
func rpcErrFromRaw(raw json.RawMessage) error {
	b := bytes.TrimSpace(raw)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		if s == "" {
			return nil
		}
		return errors.New(s)
	}
	var obj struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(b, &obj); err == nil {
		if obj.Message != "" || obj.Code != 0 {
			if len(obj.Data) > 0 && !bytes.Equal(bytes.TrimSpace(obj.Data), []byte("null")) {
				return fmt.Errorf("json-rpc error %d: %s (data=%s)", obj.Code, obj.Message, string(obj.Data))
			}
			return fmt.Errorf("json-rpc error %d: %s", obj.Code, obj.Message)
		}
	}
	return fmt.Errorf("json-rpc error: %s", strings.TrimSpace(string(b)))
}

func (c *Client) callRPC(ctx context.Context, method string, params []interface{}, result interface{}) error {
	if c.nodeURL == "" {
		return errors.New("node URL is empty")
	}
	id := atomic.AddInt64(&c.rpcReqID, 1)
	reqBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	})
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.nodeURL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http %s: %s", resp.Status, truncateBytes(respBody, 512))
	}

	var env struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Result  json.RawMessage `json:"result"`
		Error   json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(respBody, &env); err != nil {
		return fmt.Errorf("decode json-rpc envelope: %w (body=%s)", err, truncateBytes(respBody, 512))
	}
	if err := rpcErrFromRaw(env.Error); err != nil {
		return err
	}
	if result == nil {
		return nil
	}
	res := bytes.TrimSpace(env.Result)
	if len(res) == 0 || bytes.Equal(res, []byte("null")) {
		// Unmarshal null into result (e.g. **T or *T becomes nil where supported)
		return json.Unmarshal([]byte("null"), result)
	}
	return json.Unmarshal(env.Result, result)
}

func truncateBytes(b []byte, max int) string {
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "..."
}
