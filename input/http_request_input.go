package input

import (
	"io"
	"net/http"
	"time"

	"rulecraft/plugin"
)

// HTTPRequestInput HTTP 请求插件。
type HTTPRequestInput struct {
	url     string
	timeout int // 超时秒数
	lastResult map[string]interface{}
}

func NewHTTPRequestInput() *HTTPRequestInput {
	return &HTTPRequestInput{
		timeout:    10,
		lastResult: make(map[string]interface{}),
	}
}

func (h *HTTPRequestInput) ID() string        { return "http_request" }
func (h *HTTPRequestInput) Name() string      { return "HTTP 请求传感器" }
func (h *HTTPRequestInput) IsAvailable() bool { return true }

// Configure 设置请求参数（通过 API 调用）。
func (h *HTTPRequestInput) Configure(url string, timeout int) {
	h.url = url
	if timeout > 0 {
		h.timeout = timeout
	}
}

func (h *HTTPRequestInput) Collect(ctx *plugin.SystemContext) error {
	if h.url == "" {
		return nil
	}
	client := &http.Client{Timeout: time.Duration(h.timeout) * time.Second}
	resp, err := client.Get(h.url)
	if err != nil {
		ctx.SetState("http_request.error", err.Error())
		ctx.SetState("http_request.status_code", 0)
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	ctx.SetState("http_request.status_code", resp.StatusCode)
	ctx.SetState("http_request.body", bodyStr)
	ctx.SetState("http_request.url", h.url)
	ctx.SetState("http_request.timestamp", time.Now().Unix())
	// 自动解析返回数据
	parseIncomingData(ctx, "http_request", bodyStr, "auto")
	return nil
}
