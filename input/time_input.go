package input

import (
	"fmt"
	"time"

	"rulecraft/plugin"
)

// TimeInput 采集时间/日历信息。纯 Go time 包实现，无需 platform 层。
type TimeInput struct{}

// NewTimeInput 创建时间输入插件。
func NewTimeInput() *TimeInput {
	return &TimeInput{}
}

func (t *TimeInput) ID() string        { return "time" }
func (t *TimeInput) Name() string      { return "时间/日历传感器" }
func (t *TimeInput) IsAvailable() bool { return true }

// Collect 采集时间信息并写入全局状态。
func (t *TimeInput) Collect(ctx *plugin.SystemContext) error {
	now := time.Now()

	weekday := int(now.Weekday()) // 0=Sunday, 1=Monday, ...
	weekdayName := now.Weekday().String()
	isWeekend := weekday == 0 || weekday == 6
	zone, _ := now.Zone()

	ctx.SetState("time.hour", now.Hour())
	ctx.SetState("time.minute", now.Minute())
	ctx.SetState("time.weekday", weekday)
	ctx.SetState("time.weekday_name", weekdayName)
	ctx.SetState("time.is_weekend", isWeekend)
	ctx.SetState("time.month", int(now.Month()))
	ctx.SetState("time.day", now.Day())
	ctx.SetState("time.unix_timestamp", now.Unix())
	ctx.SetState("time.iso8601", now.Format("2006-01-02T15:04:05-07:00"))
	ctx.SetState("time.timezone", zone)
	ctx.SetState("time.now", fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()))
	return nil
}
