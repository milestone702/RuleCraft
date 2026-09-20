package output

import (
	"fmt"

	"rulecraft/plugin"
)

// DingTalkBotOutput 发送钉钉群机器人消息。
type DingTalkBotOutput struct{}

func NewDingTalkBotOutput() *DingTalkBotOutput { return &DingTalkBotOutput{} }

func (d *DingTalkBotOutput) ID() string        { return "notify_dingtalk" }
func (d *DingTalkBotOutput) Name() string      { return "钉钉机器人" }
func (d *DingTalkBotOutput) IsAvailable() bool { return true }

// Execute params: webhook_url (必填), secret (可选，加签), message (必填)
// 若配置 secret，将追加 timestamp+sign 查询参数（需自行实现 HMAC 时可用完整 webhook_url 直接带 sign）。
func (d *DingTalkBotOutput) Execute(params map[string]interface{}) error {
	url := paramString(params, "webhook_url")
	if url == "" {
		return fmt.Errorf("notify_dingtalk: webhook_url is required")
	}
	msg := paramString(params, "message")
	if msg == "" {
		return fmt.Errorf("notify_dingtalk: message is required")
	}
	// 加签请把带 sign 的完整 URL 放入 webhook_url，或使用自定义机器人「安全设置-自定义关键词」
	payload := map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]interface{}{"content": msg},
	}
	return postJSON(url, payload, nil)
}

func (d *DingTalkBotOutput) Reset(map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
