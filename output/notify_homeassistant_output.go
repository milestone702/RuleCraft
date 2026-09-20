package output

import (
	"fmt"

	"rulecraft/plugin"
)

// HomeAssistantOutput 调用 Home Assistant REST API。
type HomeAssistantOutput struct{}

func NewHomeAssistantOutput() *HomeAssistantOutput { return &HomeAssistantOutput{} }

func (h *HomeAssistantOutput) ID() string        { return "notify_homeassistant" }
func (h *HomeAssistantOutput) Name() string      { return "Home Assistant" }
func (h *HomeAssistantOutput) IsAvailable() bool { return true }

// Execute 支持两种模式：
// 1) persistent_notification.create 通知：
//    base_url, token, message [, title]
// 2) service 调用：
//    base_url, token, domain, service [, service_data]
func (h *HomeAssistantOutput) Execute(params map[string]interface{}) error {
	base := paramString(params, "base_url")
	if base == "" {
		base = paramString(params, "url")
	}
	if base == "" {
		return fmt.Errorf("notify_homeassistant: base_url is required")
	}
	token := paramString(params, "token")
	if token == "" {
		return fmt.Errorf("notify_homeassistant: token is required")
	}

	// service 调用模式
	domain := paramString(params, "domain")
	service := paramString(params, "service")
	if domain != "" && service != "" {
		url := fmt.Sprintf("%s/api/services/%s/%s", trimSlash(base), domain, service)
		payload := map[string]interface{}{}
		if sd, ok := params["service_data"].(map[string]interface{}); ok {
			payload = sd
		} else if msg := paramString(params, "message"); msg != "" {
			payload["message"] = msg
		}
		return postJSON(url, payload, map[string]string{
			"Authorization": "Bearer " + token,
		})
	}

	// 默认：创建持久通知
	msg := paramString(params, "message")
	if msg == "" {
		return fmt.Errorf("notify_homeassistant: message is required")
	}
	title := paramString(params, "title")
	if title == "" {
		title = "RuleCraft"
	}
	url := trimSlash(base) + "/api/services/persistent_notification/create"
	return postJSON(url, map[string]interface{}{
		"title": title,
		"message": msg,
	}, map[string]string{"Authorization": "Bearer " + token})
}

func (h *HomeAssistantOutput) Reset(map[string]interface{}) error {
	return plugin.ErrResetNotSupported
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}
