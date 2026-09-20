package output

import (
	"fmt"

	"rulecraft/plugin"
)

// TelegramBotOutput 发送 Telegram Bot 消息。
type TelegramBotOutput struct{}

func NewTelegramBotOutput() *TelegramBotOutput { return &TelegramBotOutput{} }

func (t *TelegramBotOutput) ID() string        { return "notify_telegram" }
func (t *TelegramBotOutput) Name() string      { return "Telegram Bot" }
func (t *TelegramBotOutput) IsAvailable() bool { return true }

// Execute params: bot_token (必填), chat_id (必填), message (必填)
func (t *TelegramBotOutput) Execute(params map[string]interface{}) error {
	token := paramString(params, "bot_token")
	if token == "" {
		return fmt.Errorf("notify_telegram: bot_token is required")
	}
	chatID := paramString(params, "chat_id")
	if chatID == "" {
		return fmt.Errorf("notify_telegram: chat_id is required")
	}
	msg := paramString(params, "message")
	if msg == "" {
		return fmt.Errorf("notify_telegram: message is required")
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    msg,
	}
	return postJSON(url, payload, nil)
}

func (t *TelegramBotOutput) Reset(map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}
