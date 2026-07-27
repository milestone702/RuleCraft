package output

// LockOutput 锁定工作站。
type LockOutput struct {
	plat PlatformSubset
}

// NewLockOutput 创建锁定工作站输出插件。
func NewLockOutput(plat PlatformSubset) *LockOutput {
	return &LockOutput{plat: plat}
}

func (l *LockOutput) ID() string        { return "lock" }
func (l *LockOutput) Name() string      { return "锁定工作站" }
func (l *LockOutput) IsAvailable() bool { return true }

// Execute 锁定工作站。不需要参数。
func (l *LockOutput) Execute(params map[string]interface{}) error {
	return l.plat.LockWorkstation()
}

// Reset 不支持 Reset。
func (l *LockOutput) Reset(params map[string]interface{}) error {
	return nil // 锁定操作无法撤销
}
