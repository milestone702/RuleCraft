package output

import (
	"fmt"

	"rulecraft/plugin"
)

// FeishuBotOutput 发送飞书群机器人消息。
type FeishuBotOutput struct{}

func NewFeishuBotOutput() *FeishuBotOutput { return &FeishuBotOutput{} }

func (f *FeishuBotOutput) ID() string        { return "notify_feishu" }
func (f *FeishuBotOutput) Name() string      { return "飞书机器人" }
func (f *FeishuBotOutput) IsAvailable() bool { return true }

// Execute params: webhook_url (必填), message (必填), title (可选)
func (f *FeishuBotOutput) Execute(params map[string]interface{}) error {
	url := paramString(params, "webhook_url")
	if url == "" {
		return fmt.Errorf("notify_feishu: webhook_url is required")
	}
	msg := paramString(params, "message")
	if msg == "" {
		return fmt.Errorf("notify_feishu: message is required")
	}
	title := paramString(params, "title")
	if title == "" {
		title = "RuleCraft"
	}
	payload := map[string]interface{}{
		"msg_type": "text",
		"content":  map[string]interface{}{"text": msg},
	}
	_ = title // text 消息无标题；如需富文本可扩展
	return postJSON(url, payload, nil)
}

func (f *FeishuBotOutput) Reset(map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
