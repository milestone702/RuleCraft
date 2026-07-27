package input

import (
	"encoding/json"
	"net/url"
	"strings"

	"rulecraft/plugin"
)

// parseIncomingData 解析入站数据，支持 json/kv/auto 模式。
// 解析结果直接写入 prefix.key，prefix._raw 存储原始数据。
func parseIncomingData(ctx *plugin.SystemContext, prefix, data, mode string) {
	if data == "" {
		return
	}
	// 始终存原始数据
	ctx.SetState(prefix+"._raw", data)

	switch mode {
	case "json":
		parseJSON(ctx, prefix, data)
	case "kv":
		parseKeyValue(ctx, prefix, data)
	case "auto":
		if !parseJSON(ctx, prefix, data) {
			parseKeyValue(ctx, prefix, data)
		}
	}
}

// parseJSON 解析 JSON 对象，字段直接写入 prefix.<key>。
func parseJSON(ctx *plugin.SystemContext, prefix, data string) bool {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(data), &parsed); err != nil {
		return false
	}
	for k, v := range parsed {
		if k == "_raw" {
			continue
		}
		ctx.SetState(prefix+"."+k, v)
	}
	ctx.SetState(prefix+"._parsed", "json")
	return true
}

// parseKeyValue 解析 key=value&key2=value2 格式，
// 字段直接写入 prefix.<key>。
func parseKeyValue(ctx *plugin.SystemContext, prefix, data string) {
	pairs := strings.Split(data, "&")
	decoded := false
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		eqIdx := strings.IndexByte(pair, '=')
		if eqIdx <= 0 {
			// 无等号：作为真值写入 prefix.<key>
			key := pair
			if d, err := url.QueryUnescape(key); err == nil {
				key = d
			}
			if key != "_raw" && key != "_parsed" {
				ctx.SetState(prefix+"."+key, true)
			}
			continue
		}
		key := pair[:eqIdx]
		value := pair[eqIdx+1:]
		if d, err := url.QueryUnescape(value); err == nil {
			value = d
			decoded = true
		}
		if d, err := url.QueryUnescape(key); err == nil {
			key = d
		}
		if key != "_raw" && key != "_parsed" {
			ctx.SetState(prefix+"."+key, value)
		}
	}
	ctx.SetState(prefix+"._parsed", "kv")
	_ = decoded
}
