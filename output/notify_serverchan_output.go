package output

import (
	"fmt"
	"net/url"

	"rulecraft/plugin"
)

// ServerChanOutput 发送 Server酱 Turbo 消息。
type ServerChanOutput struct{}

func NewServerChanOutput() *ServerChanOutput { return &ServerChanOutput{} }

func (s *ServerChanOutput) ID() string        { return "notify_serverchan" }
func (s *ServerChanOutput) Name() string      { return "Server酱" }
func (s *ServerChanOutput) IsAvailable() bool { return true }

// Execute params: sendkey (必填), title (必填), message/ desp (必填)
func (s *ServerChanOutput) Execute(params map[string]interface{}) error {
	key := paramString(params, "sendkey")
	if key == "" {
		key = paramString(params, "key")
	}
	if key == "" {
		return fmt.Errorf("notify_serverchan: sendkey is required")
	}
	title := paramString(params, "title")
	if title == "" {
		title = "RuleCraft"
	}
	msg := paramString(params, "message")
	if msg == "" {
		msg = paramString(params, "desp")
	}
	if msg == "" {
		return fmt.Errorf("notify_serverchan: message is required")
	}
	u := fmt.Sprintf("https://sctapi.ftqq.com/%s.send?title=%s&desp=%s",
		url.PathEscape(key), url.QueryEscape(title), url.QueryEscape(msg))
	return getURL(u)
}

func (s *ServerChanOutput) Reset(map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
