package engine

import (
	"testing"
	"time"

	"rulecraft/config"
)

func TestCompareBetween(t *testing.T) {
	cases := []struct {
		val     interface{}
		rng     interface{}
		want    bool
		wantErr bool
	}{
		{50, []interface{}{20.0, 80.0}, true, false},
		{10, []interface{}{20.0, 80.0}, false, false},
		{90, []interface{}{20.0, 80.0}, false, false},
		{"50", "20,80", true, false},
		{50, "bad", false, true},
		{50, []interface{}{1.0}, false, true},
	}
	for _, c := range cases {
		got, err := compareBetween(c.val, c.rng)
		if c.wantErr {
			if err == nil {
				t.Fatalf("compareBetween(%v,%v): expected error", c.val, c.rng)
			}
			continue
		}
		if err != nil {
			t.Fatalf("compareBetween(%v,%v): unexpected error: %v", c.val, c.rng, err)
		}
		if got != c.want {
			t.Fatalf("compareBetween(%v,%v)=%v want %v", c.val, c.rng, got, c.want)
		}
	}
}

func TestEvaluatorNewOperators(t *testing.T) {
	ce := NewConditionEvaluator()
	states := map[string]interface{}{
		"sysres.cpu_percent": 95.0,
		"flag":               true,
		"mode":               "Night",
		"text":               "HelloWorld",
	}

	cases := []struct {
		name string
		node config.ConditionNode
		want bool
	}{
		{"between", config.ConditionNode{Type: "state", StateKey: "sysres.cpu_percent", Operator: "between", Value: []interface{}{90.0, 100.0}}, true},
		{"not_between", config.ConditionNode{Type: "state", StateKey: "sysres.cpu_percent", Operator: "not_between", Value: []interface{}{0.0, 50.0}}, true},
		{"is_true", config.ConditionNode{Type: "state", StateKey: "flag", Operator: "is_true"}, true},
		{"is_false", config.ConditionNode{Type: "state", StateKey: "missing", Operator: "is_false"}, true},
		{"equals_ignore_case", config.ConditionNode{Type: "state", StateKey: "mode", Operator: "equals_ignore_case", Value: "night"}, true},
		{"contains_ignore_case", config.ConditionNode{Type: "state", StateKey: "text", Operator: "contains_ignore_case", Value: "world"}, true},
	}

	for _, c := range cases {
		got, err := ce.Eval(&c.node, states)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", c.name, err)
		}
		if got != c.want {
			t.Fatalf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

func TestEvaluatorChangedAndIncreased(t *testing.T) {
	ce := NewConditionEvaluator()
	node := config.ConditionNode{Type: "state", StateKey: "wifi.ssid", Operator: "changed"}

	// 首轮无历史 → false
	got, err := ce.Eval(&node, map[string]interface{}{"wifi.ssid": "Home"})
	if err != nil || got {
		t.Fatalf("first poll changed should be false, got %v err=%v", got, err)
	}

	// 第二轮值变化 → true
	got, err = ce.Eval(&node, map[string]interface{}{"wifi.ssid": "Office"})
	if err != nil || !got {
		t.Fatalf("second poll changed should be true, got %v err=%v", got, err)
	}

	// increased
	inc := config.ConditionNode{Type: "state", StateKey: "usb.count", Operator: "increased"}
	_, _ = ce.Eval(&inc, map[string]interface{}{"usb.count": 1})
	got, err = ce.Eval(&inc, map[string]interface{}{"usb.count": 3})
	if err != nil || !got {
		t.Fatalf("increased should be true, got %v err=%v", got, err)
	}
}

func TestThresholdEdgeSequence(t *testing.T) {
	// 模拟调度器阈值+边缘：连续 3 次 true 后才应触发 TriggerTrue
	ts := &TaskScheduler{
		edgeTracker:       NewEdgeTriggerTracker(),
		thresholdCounters: make(map[string]int),
		cooldowns:         make(map[string]time.Time),
	}

	key := "task/out"

	// 基线：首次评估为 false（记录但不触发）
	if d := ts.edgeTracker.Check(key, false); d != TriggerNone {
		t.Fatalf("baseline should be TriggerNone, got %v", d)
	}

	// 连续 true：阈值未满时 effective=false
	e1 := ts.applyThreshold(key, true, 3)
	d1 := ts.edgeTracker.Check(key, e1)
	if e1 || d1 != TriggerNone {
		t.Fatalf("poll1: effective=%v dir=%v, want false/None", e1, d1)
	}

	e2 := ts.applyThreshold(key, true, 3)
	d2 := ts.edgeTracker.Check(key, e2)
	if e2 || d2 != TriggerNone {
		t.Fatalf("poll2: effective=%v dir=%v, want false/None", e2, d2)
	}

	// 第三次达阈值：effective=true → 边缘 false→true
	e3 := ts.applyThreshold(key, true, 3)
	d3 := ts.edgeTracker.Check(key, e3)
	if !e3 || d3 != TriggerTrue {
		t.Fatalf("poll3: effective=%v dir=%v, want true/TriggerTrue", e3, d3)
	}

	// 条件变 false：立即清零并触发 TriggerFalse
	e4 := ts.applyThreshold(key, false, 3)
	d4 := ts.edgeTracker.Check(key, e4)
	if e4 || d4 != TriggerFalse {
		t.Fatalf("poll4: effective=%v dir=%v, want false/TriggerFalse", e4, d4)
	}
}
