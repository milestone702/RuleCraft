package input

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"rulecraft/plugin"
)

// HTTPInInput HTTP 入站插件 — 监听 HTTP POST 请求，支持数据解析。
type HTTPInInput struct {
	mu        sync.Mutex
	server    *http.Server
	started   bool
	port      int
	authKey   string
	parseMode string // "raw", "json", "kv", "auto"
	lastBody  string
	lastPath  string
	lastTime  int64
}

func NewHTTPInInput() *HTTPInInput {
	return &HTTPInInput{port: 19531, parseMode: "auto"}
}

func (h *HTTPInInput) ID() string        { return "http_in" }
func (h *HTTPInInput) Name() string      { return "HTTP 入站传感器" }
func (h *HTTPInInput) IsAvailable() bool { return true }

// Configure 设置监听端口和鉴权密钥。
func (h *HTTPInInput) Configure(port int, authKey string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if port > 0 && port != h.port {
		h.stop()
		h.port = port
	}
	h.authKey = authKey
}

func (h *HTTPInInput) Collect(ctx *plugin.SystemContext) error {
	h.mu.Lock()
	if !h.started {
		h.started = true
		h.start()
	}
	body := h.lastBody
	path := h.lastPath
	ts := h.lastTime
	h.mu.Unlock()

	if body != "" {
		ctx.SetState("http_in.body", body)
		ctx.SetState("http_in.path", path)
		ctx.SetState("http_in.timestamp", ts)
		ctx.SetState("http_in.port", h.port)
		// 解析数据
		parseIncomingData(ctx, "http_in", body, h.parseMode)
	}
	return nil
}

func (h *HTTPInInput) start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Key 鉴权
		if h.authKey != "" {
			if r.Header.Get("X-API-Key") != h.authKey && r.URL.Query().Get("key") != h.authKey {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"unauthorized"}`))
				return
			}
		}
		body, _ := io.ReadAll(r.Body)
		h.mu.Lock()
		h.lastBody = string(body)
		h.lastPath = r.URL.Path
		h.lastTime = time.Now().Unix()
		h.mu.Unlock()
		w.Write([]byte(`{"status":"ok"}`))
	})

	h.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", h.port),
		Handler: mux,
	}
	go func() {
		h.server.ListenAndServe()
	}()
}

func (h *HTTPInInput) stop() {
	if h.server != nil {
		h.server.Close()
		h.server = nil
	}
	h.started = false
}
