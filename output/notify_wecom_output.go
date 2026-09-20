package output

import (
	"fmt"

	"rulecraft/plugin"
)

// WeComBotOutput 发送企业微信群机器人消息。
type WeComBotOutput struct{}

func NewWeComBotOutput() *WeComBotOutput { return &WeComBotOutput{} }

func (w *WeComBotOutput) ID() string        { return "notify_wecom" }
func (w *WeComBotOutput) Name() string      { return "企业微信机器人" }
func (w *WeComBotOutput) IsAvailable() bool { return true }

// Execute params: webhook_url 或 key, message (必填), mentioned_list (可选字符串逗号分隔)
func (w *WeComBotOutput) Execute(params map[string]interface{}) error {
	url := paramString(params, "webhook_url")
	if url == "" {
		key := paramString(params, "key")
		if key == "" {
			return fmt.Errorf("notify_wecom: webhook_url or key is required")
		}
		url = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=" + key
	}
	msg := paramString(params, "message")
	if msg == "" {
		return fmt.Errorf("notify_wecom: message is required")
	}
	content := map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]interface{}{"content": msg},
	}
	return postJSON(url, content, nil)
}

func (w *WeComBotOutput) Reset(map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
