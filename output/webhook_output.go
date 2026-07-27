package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"rulecraft/plugin"
)

// WebhookOutput 发送 HTTP 请求（纯 Go net/http，无需平台层）。
type WebhookOutput struct {
	client *http.Client
}

// NewWebhookOutput 创建 Webhook 输出插件。
func NewWebhookOutput() *WebhookOutput {
	return &WebhookOutput{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WebhookOutput) ID() string        { return "webhook" }
func (w *WebhookOutput) Name() string      { return "Webhook 请求" }
func (w *WebhookOutput) IsAvailable() bool { return true }

// Execute 发送 HTTP 请求。
// params:
//
//	url (string, 必填): 请求 URL
//	method (string, 可选): GET / POST / PUT，默认 POST
//	body (object, 可选): JSON 请求体
//	headers (map[string]string, 可选): 自定义请求头
func (w *WebhookOutput) Execute(params map[string]interface{}) error {
	url, _ := params["url"].(string)
	if url == "" {
		return fmt.Errorf("webhook: url is required")
	}

	method, _ := params["method"].(string)
	if method == "" {
		method = "POST"
	}

	var bodyReader io.Reader
	if body, ok := params["body"]; ok {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("webhook: marshal body failed: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("webhook: create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 自定义请求头
	if headers, ok := params["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if s, ok := v.(string); ok {
				req.Header.Set(k, s)
			}
		}
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook: server returned %d", resp.StatusCode)
	}

	return nil
}

// Reset 不支持 Reset。
func (w *WebhookOutput) Reset(params map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
