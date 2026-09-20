package input

import (
	"io"
	"net/http"
	"strings"
	"time"

	"rulecraft/plugin"
)

// PublicIPInput 通过公网 HTTP 服务获取本机出口 IP。
type PublicIPInput struct {
	url     string
	timeout int
}

func NewPublicIPInput() *PublicIPInput {
	return &PublicIPInput{
		url:     "https://api.ipify.org",
		timeout: 5,
	}
}

func (p *PublicIPInput) ID() string        { return "public_ip" }
func (p *PublicIPInput) Name() string      { return "公网 IP 传感器" }
func (p *PublicIPInput) IsAvailable() bool { return true }

// ConfigureFromMap 可覆盖查询地址。
func (p *PublicIPInput) ConfigureFromMap(params map[string]interface{}) error {
	if v, ok := params["url"].(string); ok && v != "" {
		p.url = v
	}
	if v, ok := params["timeout"].(float64); ok && v > 0 {
		p.timeout = int(v)
	}
	return nil
}

// Collect 写入 public_ip.ip / error。
func (p *PublicIPInput) Collect(ctx *plugin.SystemContext) error {
	client := &http.Client{Timeout: time.Duration(p.timeout) * time.Second}
	resp, err := client.Get(p.url)
	if err != nil {
		ctx.SetState("public_ip.ip", "")
		ctx.SetState("public_ip.error", err.Error())
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		ctx.SetState("public_ip.ip", "")
		ctx.SetState("public_ip.error", err.Error())
		return nil
	}
	ip := strings.TrimSpace(string(body))
	ctx.SetState("public_ip.ip", ip)
	ctx.SetState("public_ip.error", "")
	return nil
}
