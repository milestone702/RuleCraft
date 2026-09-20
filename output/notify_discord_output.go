package output

import (
	"fmt"

	"rulecraft/plugin"
)

// DiscordWebhookOutput 发送 Discord Webhook 消息。
type DiscordWebhookOutput struct{}

func NewDiscordWebhookOutput() *DiscordWebhookOutput { return &DiscordWebhookOutput{} }

func (d *DiscordWebhookOutput) ID() string        { return "notify_discord" }
func (d *DiscordWebhookOutput) Name() string      { return "Discord Webhook" }
func (d *DiscordWebhookOutput) IsAvailable() bool { return true }

// Execute params: webhook_url (必填), message (必填), username (可选)
func (d *DiscordWebhookOutput) Execute(params map[string]interface{}) error {
	url := paramString(params, "webhook_url")
	if url == "" {
		url = paramString(params, "url")
	}
	if url == "" {
		return fmt.Errorf("notify_discord: webhook_url is required")
	}
	msg := paramString(params, "message")
	if msg == "" {
		return fmt.Errorf("notify_discord: message is required")
	}
	payload := map[string]interface{}{"content": msg}
	if u := paramString(params, "username"); u != "" {
		payload["username"] = u
	}
	return postJSON(url, payload, nil)
}

func (d *DiscordWebhookOutput) Reset(map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
