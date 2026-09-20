package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// ScheduledTaskInput 查询计划任务状态。
// 通过 params.task_name 配置要监控的任务名。
type ScheduledTaskInput struct {
	plat     platform.Platform
	taskName string
}

func NewScheduledTaskInput(plat platform.Platform) *ScheduledTaskInput {
	return &ScheduledTaskInput{plat: plat}
}

func (s *ScheduledTaskInput) ID() string        { return "scheduled_task" }
func (s *ScheduledTaskInput) Name() string      { return "计划任务传感器" }
func (s *ScheduledTaskInput) IsAvailable() bool { return true }

// ConfigureFromMap params: task_name (string)
func (s *ScheduledTaskInput) ConfigureFromMap(params map[string]interface{}) error {
	if v, ok := params["task_name"].(string); ok {
		s.taskName = v
	}
	return nil
}

// Collect 写入 scheduled_task.name / status / found / is_running。
func (s *ScheduledTaskInput) Collect(ctx *plugin.SystemContext) error {
	ctx.SetState("scheduled_task.name", s.taskName)
	if s.taskName == "" {
		ctx.SetState("scheduled_task.found", false)
		ctx.SetState("scheduled_task.status", "")
		ctx.SetState("scheduled_task.is_running", false)
		return nil
	}
	status, found, err := s.plat.GetScheduledTaskStatus(s.taskName)
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("scheduled_task.found", false)
			return nil
		}
		ctx.SetState("scheduled_task.found", false)
		ctx.SetState("scheduled_task.status", "Error")
		return nil
	}
	ctx.SetState("scheduled_task.found", found)
	ctx.SetState("scheduled_task.status", status)
	ctx.SetState("scheduled_task.is_running", found && status == "Running")
	return nil
}
