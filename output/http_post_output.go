package output

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"rulecraft/plugin"
)

// HTTPPostOutput 发送 HTTP POST 请求。
type HTTPPostOutput struct {
	client *http.Client
}

// NewHTTPPostOutput 创建 HTTP POST 输出插件。
func NewHTTPPostOutput() *HTTPPostOutput {
	return &HTTPPostOutput{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *HTTPPostOutput) ID() string        { return "http_post" }
func (h *HTTPPostOutput) Name() string      { return "HTTP POST 请求" }
func (h *HTTPPostOutput) IsAvailable() bool { return true }
func (h *HTTPPostOutput) Reset(map[string]interface{}) error { return plugin.ErrResetNotSupported }

// Execute 发送 HTTP POST 请求。
// params:
//
//	url (string, 必填): 请求 URL
//	body (string, 可选): 请求体内容
//	content_type (string, 可选): Content-Type，默认 application/json
//	headers (object, 可选): 自定义请求头，JSON 对象格式
//	timeout (number, 可选): 超时秒数，默认 10
func (h *HTTPPostOutput) Execute(params map[string]interface{}) error {
	url, _ := params["url"].(string)
	if url == "" {
		return fmt.Errorf("http_post: url is required")
	}

	// 可选超时
	if timeout, ok := params["timeout"]; ok {
		if t, ok := timeout.(float64); ok && t > 0 {
			h.client.Timeout = time.Duration(t) * time.Second
		}
	}

	// 请求体
	var bodyReader io.Reader
	if body, ok := params["body"]; ok {
		bodyStr := fmt.Sprintf("%v", body)
		bodyReader = bytes.NewReader([]byte(bodyStr))
	}

	req, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		return fmt.Errorf("http_post: create request failed: %w", err)
	}

	// Content-Type
	contentType, _ := params["content_type"].(string)
	if contentType == "" {
		contentType = "application/json"
	}
	req.Header.Set("Content-Type", contentType)

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
		return fmt.Errorf("http_post: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("http_post: %s returned %d: %s", url, resp.StatusCode, string(body))
	}
	return nil
}
