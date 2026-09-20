package input

import (
	"strings"

	"rulecraft/platform"
	"rulecraft/plugin"
)

// ServiceInput 查询 Windows 服务状态。
// 通过 Configure 指定要查询的服务名列表。
type ServiceInput struct {
	plat     platform.Platform
	services []string
}

// NewServiceInput 创建服务状态输入插件。
func NewServiceInput(plat platform.Platform) *ServiceInput {
	return &ServiceInput{plat: plat}
}

func (s *ServiceInput) ID() string        { return "service" }
func (s *ServiceInput) Name() string      { return "Windows 服务传感器" }
func (s *ServiceInput) IsAvailable() bool { return true }

// Configure 设置要监控的服务名列表。
func (s *ServiceInput) Configure(services []string) {
	s.services = services
}

// ConfigureFromMap 从任务参数配置。
// params: services ([]string 或逗号分隔 string)
func (s *ServiceInput) ConfigureFromMap(params map[string]interface{}) error {
	if params == nil {
		return nil
	}
	if v, ok := params["services"]; ok {
		switch sv := v.(type) {
		case string:
			parts := strings.Split(sv, ",")
			list := make([]string, 0, len(parts))
			for _, p := range parts {
				if p = strings.TrimSpace(p); p != "" {
					list = append(list, p)
				}
			}
			s.services = list
		case []interface{}:
			list := make([]string, 0, len(sv))
			for _, item := range sv {
				if str, ok := item.(string); ok && str != "" {
					list = append(list, str)
				}
			}
			s.services = list
		}
	}
	return nil
}

// Collect 采集服务状态。
// 状态键：
//
//	service.<name>.status  — Running / Stopped / ...
//	service.<name>.start_type
//	service.<name>.is_running — bool
//	service.running_count
func (s *ServiceInput) Collect(ctx *plugin.SystemContext) error {
	if len(s.services) == 0 {
		ctx.SetState("service.configured", false)
		return nil
	}
	ctx.SetState("service.configured", true)

	running := 0
	for _, name := range s.services {
		if name == "" {
			continue
		}
		info, err := s.plat.GetServiceStatus(name)
		if err != nil {
			if platform.IsErrNotImplemented(err) {
				ctx.SetState("service."+name+".status", "Unknown")
				ctx.SetState("service."+name+".is_running", false)
				continue
			}
			ctx.SetState("service."+name+".status", "Error")
			ctx.SetState("service."+name+".is_running", false)
			continue
		}
		ctx.SetState("service."+name+".status", info.Status)
		ctx.SetState("service."+name+".start_type", info.StartType)
		isRunning := info.Status == "Running"
		ctx.SetState("service."+name+".is_running", isRunning)
		if isRunning {
			running++
		}
	}
	ctx.SetState("service.running_count", running)
	return nil
}
