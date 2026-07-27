package output

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"rulecraft/plugin"
)

// HTTPGetOutput 发送 HTTP GET 请求。
type HTTPGetOutput struct {
	client *http.Client
}

// NewHTTPGetOutput 创建 HTTP GET 输出插件。
func NewHTTPGetOutput() *HTTPGetOutput {
	return &HTTPGetOutput{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *HTTPGetOutput) ID() string        { return "http_get" }
func (h *HTTPGetOutput) Name() string      { return "HTTP GET 请求" }
func (h *HTTPGetOutput) IsAvailable() bool { return true }
func (h *HTTPGetOutput) Reset(map[string]interface{}) error { return plugin.ErrResetNotSupported }

// Execute 发送 HTTP GET 请求。
// params:
//
//	url (string, 必填): 请求 URL
//	headers (object, 可选): 自定义请求头，JSON 对象格式
//	timeout (number, 可选): 超时秒数，默认 10
func (h *HTTPGetOutput) Execute(params map[string]interface{}) error {
	url, _ := params["url"].(string)
	if url == "" {
		return fmt.Errorf("http_get: url is required")
	}

	// 可选超时
	if timeout, ok := params["timeout"]; ok {
		if t, ok := timeout.(float64); ok && t > 0 {
			h.client.Timeout = time.Duration(t) * time.Second
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("http_get: create request failed: %w", err)
	}

	// 自定义请求头
	if headers, ok := params["headers"]; ok {
		if hdrMap, ok := headers.(map[string]interface{}); ok {
			for k, v := range hdrMap {
				req.Header.Set(k, fmt.Sprintf("%v", v))
			}
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("http_get: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("http_get: %s returned %d: %s", url, resp.StatusCode, string(body))
	}
	return nil
}
