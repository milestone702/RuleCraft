package output

import (
	"fmt"

	"rulecraft/plugin"
)

// SlackWebhookOutput 发送 Slack Incoming Webhook。
type SlackWebhookOutput struct{}

func NewSlackWebhookOutput() *SlackWebhookOutput { return &SlackWebhookOutput{} }

func (s *SlackWebhookOutput) ID() string        { return "notify_slack" }
func (s *SlackWebhookOutput) Name() string      { return "Slack Webhook" }
func (s *SlackWebhookOutput) IsAvailable() bool { return true }

// Execute params: webhook_url (必填), message (必填)
func (s *SlackWebhookOutput) Execute(params map[string]interface{}) error {
	url := paramString(params, "webhook_url")
	if url == "" {
		return fmt.Errorf("notify_slack: webhook_url is required")
	}
	msg := paramString(params, "message")
	if msg == "" {
		return fmt.Errorf("notify_slack: message is required")
	}
	return postJSON(url, map[string]interface{}{"text": msg}, nil)
}

func (s *SlackWebhookOutput) Reset(map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
