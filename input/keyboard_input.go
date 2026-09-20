package input

import (
	"rulecraft/platform"
	"rulecraft/plugin"
)

// KeyboardInput 采集键盘锁定键状态。
type KeyboardInput struct {
	plat platform.Platform
}

func NewKeyboardInput(plat platform.Platform) *KeyboardInput {
	return &KeyboardInput{plat: plat}
}

func (k *KeyboardInput) ID() string        { return "keyboard" }
func (k *KeyboardInput) Name() string      { return "键盘锁定键传感器" }
func (k *KeyboardInput) IsAvailable() bool { return true }

// Collect 写入 keyboard.caps_lock / num_lock / scroll_lock。
func (k *KeyboardInput) Collect(ctx *plugin.SystemContext) error {
	caps, num, scroll, err := k.plat.GetKeyboardLockStates()
	if err != nil {
		if platform.IsErrNotImplemented(err) {
			ctx.SetState("keyboard.caps_lock", false)
			ctx.SetState("keyboard.num_lock", false)
			ctx.SetState("keyboard.scroll_lock", false)
			return nil
		}
		return err
	}
	ctx.SetState("keyboard.caps_lock", caps)
	ctx.SetState("keyboard.num_lock", num)
	ctx.SetState("keyboard.scroll_lock", scroll)
	return nil
}
