package input

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"rulecraft/plugin"
)

// NetworkDetectInput 网络检测插件（Ping / TCP 端口 / HTTP 状态码）。
type NetworkDetectInput struct {
	target  string // IP 地址、域名 或 URL
	port    int    // TCP 端口（仅 tcp 模式）
	timeout int    // 超时秒数
	mode    string // "ping", "tcp", "http"
}

func NewNetworkDetectInput() *NetworkDetectInput {
	return &NetworkDetectInput{timeout: 5, mode: "ping"}
}

func (n *NetworkDetectInput) ID() string        { return "network_detect" }
func (n *NetworkDetectInput) Name() string      { return "网络检测传感器" }
func (n *NetworkDetectInput) IsAvailable() bool { return true }

// Configure 设置检测目标。
// mode: "ping"=ICMP探测, "tcp"=TCP端口检测, "http"=HTTP状态码检测
func (n *NetworkDetectInput) Configure(target string, port, timeout int, mode string) {
	n.target = target
	if port > 0 {
		n.port = port
	}
	if timeout > 0 {
		n.timeout = timeout
	}
	if mode != "" {
		n.mode = mode
	}
}

// Host 返回纯主机地址（移除 URL 前缀）。
func (n *NetworkDetectInput) host() string {
	if len(n.target) > 7 && n.target[:7] == "http://" {
		return n.target[7:]
	}
	if len(n.target) > 8 && n.target[:8] == "https://" {
		return n.target[8:]
	}
	return n.target
}

func (n *NetworkDetectInput) Collect(ctx *plugin.SystemContext) error {
	if n.target == "" {
		return nil
	}

	timeout := time.Duration(n.timeout) * time.Second
	start := time.Now()

	ctx.SetState("network_detect.target", n.target)
	ctx.SetState("network_detect.mode", n.mode)
	ctx.SetState("network_detect.port", n.port)
	ctx.SetState("network_detect.timeout", n.timeout)

	switch n.mode {
	case "http":
		n.checkHTTP(ctx, timeout)
	case "tcp":
		n.checkTCP(ctx, timeout)
	default: // "ping"
		n.checkPing(ctx, timeout)
	}

	elapsed := time.Since(start).Seconds()
	ctx.SetState("network_detect.latency_ms", elapsed*1000)
	ctx.SetState("network_detect.timestamp", time.Now().Unix())
	return nil
}

func (n *NetworkDetectInput) checkPing(ctx *plugin.SystemContext, timeout time.Duration) {
	host := n.host()
	reachable := pingHost(host, timeout)
	ctx.SetState("network_detect.reachable", reachable)
	if !reachable {
		ctx.SetState("network_detect.error", "ping timeout or unreachable")
	}
}

func (n *NetworkDetectInput) checkTCP(ctx *plugin.SystemContext, timeout time.Duration) {
	host := n.host()
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", n.port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		ctx.SetState("network_detect.reachable", false)
		ctx.SetState("network_detect.error", err.Error())
	} else {
		conn.Close()
		ctx.SetState("network_detect.reachable", true)
	}
}

func (n *NetworkDetectInput) checkHTTP(ctx *plugin.SystemContext, timeout time.Duration) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(n.target)
	if err != nil {
		ctx.SetState("network_detect.reachable", false)
		ctx.SetState("network_detect.http_status", 0)
		ctx.SetState("network_detect.error", err.Error())
		return
	}
	defer resp.Body.Close()
	// 读取少量 body 用于解析
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	ctx.SetState("network_detect.reachable", true)
	ctx.SetState("network_detect.http_status", resp.StatusCode)
	// 按状态码分组判定
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		ctx.SetState("network_detect.http_ok", true)
	} else {
		ctx.SetState("network_detect.http_ok", false)
	}
	// 解析响应体
	parseIncomingData(ctx, "network_detect", bodyStr, "auto")
}

func pingHost(host string, timeout time.Duration) bool {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("ping", "-n", "1", "-w", fmt.Sprintf("%d", timeout.Milliseconds()), host)
		return cmd.Run() == nil
	}
	cmd := exec.Command("ping", "-c", "1", "-W", fmt.Sprintf("%d", int(timeout.Seconds())), host)
	return cmd.Run() == nil
}
