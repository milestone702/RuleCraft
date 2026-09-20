package output

import (
	"fmt"
	"net/url"
	"strings"

	"rulecraft/plugin"
)

// BarkOutput 发送 iOS Bark 推送。
type BarkOutput struct{}

func NewBarkOutput() *BarkOutput { return &BarkOutput{} }

func (b *BarkOutput) ID() string        { return "notify_bark" }
func (b *BarkOutput) Name() string      { return "Bark 推送 (iOS)" }
func (b *BarkOutput) IsAvailable() bool { return true }

// Execute params: server (默认 https://api.day.app), device_key (必填), title, message, level
func (b *BarkOutput) Execute(params map[string]interface{}) error {
	key := paramString(params, "device_key")
	if key == "" {
		key = paramString(params, "key")
	}
	if key == "" {
		return fmt.Errorf("notify_bark: device_key is required")
	}
	server := paramString(params, "server")
	if server == "" {
		server = "https://api.day.app"
	}
	title := paramString(params, "title")
	msg := paramString(params, "message")
	if msg == "" {
		return fmt.Errorf("notify_bark: message is required")
	}
	level := paramString(params, "level") // active/timeSensitive/passive
	if level == "" {
		level = "active"
	}

	u := fmt.Sprintf("%s/%s/%s/%s?level=%s",
		strings.TrimRight(server, "/"),
		url.PathEscape(key),
		url.PathEscape(title),
		url.PathEscape(msg),
		url.QueryEscape(level))
	return getURL(u)
}

func (b *BarkOutput) Reset(map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
