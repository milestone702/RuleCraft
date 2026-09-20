//go:build windows

package platform

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

// GetStartupItems 收集 Run 注册表启动项 + Startup 文件夹项。
func (p *WindowsPlatform) GetStartupItems() ([]string, error) {
	var items []string

	if k, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE); err == nil {
		if names, err := k.ReadValueNames(-1); err == nil {
			items = append(items, names...)
		}
		k.Close()
	}

	if k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE); err == nil {
		if names, err := k.ReadValueNames(-1); err == nil {
			for _, n := range names {
				items = append(items, "LM:"+n)
			}
		}
		k.Close()
	}

	if appdata := os.Getenv("APPDATA"); appdata != "" {
		dir := filepath.Join(appdata, `Microsoft\Windows\Start Menu\Programs\Startup`)
		if entries, err := os.ReadDir(dir); err == nil {
			for _, e := range entries {
				items = append(items, "Startup:"+e.Name())
			}
		}
	}

	return items, nil
}

// GetScheduledTaskStatus 查询计划任务状态。
func (p *WindowsPlatform) GetScheduledTaskStatus(name string) (string, bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", false, fmt.Errorf("task name is required")
	}
	script := fmt.Sprintf(
		`$t = Get-ScheduledTask -TaskName '%s' -ErrorAction SilentlyContinue; if ($t) { "$($t.State)" } else { "" }`,
		escapePS(name))
	out, err := runPowerShell(script, 10)
	if err != nil {
		return "", false, err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "", false, nil
	}
	return out, true, nil
}

// GetHostsInfo 读取 hosts 文件修改时间与行数。
func (p *WindowsPlatform) GetHostsInfo() (int64, int, error) {
	path := filepath.Join(os.Getenv("SystemRoot"), "System32", "drivers", "etc", "hosts")
	fi, err := os.Stat(path)
	if err != nil {
		return 0, 0, err
	}
	lines := 0
	f, err := os.Open(path)
	if err != nil {
		return fi.ModTime().Unix(), 0, nil
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines++
	}
	return fi.ModTime().Unix(), lines, nil
}

// GetFirewallStatus 查询防火墙配置文件状态。
func (p *WindowsPlatform) GetFirewallStatus() (bool, bool, bool, error) {
	out, err := runPowerShell(
		`Get-NetFirewallProfile | Select-Object Name,Enabled | ConvertTo-Json -Compress`, 12)
	if err != nil {
		return false, false, false, err
	}
	return parseFirewallJSON(out)
}

func parseFirewallJSON(out string) (bool, bool, bool, error) {
	out = strings.TrimSpace(out)
	if out == "" {
		return false, false, false, fmt.Errorf("empty firewall status")
	}
	var raw interface{}
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return false, false, false, err
	}

	var domain, private, public bool
	handle := func(m map[string]interface{}) {
		name, _ := m["Name"].(string)
		enabled := false
		switch v := m["Enabled"].(type) {
		case bool:
			enabled = v
		case string:
			enabled = strings.EqualFold(v, "true")
		}
		switch strings.ToLower(name) {
		case "domain":
			domain = enabled
		case "private":
			private = enabled
		case "public":
			public = enabled
		}
	}

	switch v := raw.(type) {
	case map[string]interface{}:
		handle(v)
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				handle(m)
			}
		}
	default:
		return false, false, false, fmt.Errorf("unexpected firewall json")
	}
	return domain, private, public, nil
}

// IsBatterySaver 判断是否节电模式。
func (p *WindowsPlatform) IsBatterySaver() (bool, error) {
	out, err := runPowerShell(`
$s = powercfg /q SCHEME_CURRENT SUB_BATTERY BATTERY_SAVER 2>$null
if ($s -match '当前电源设置索引|Current Power Setting Index') {
  if ($s -match '0x00000001') { '1' } else { '0' }
} else { '0' }
`, 10)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "1", nil
}

// SetWindowTopmost 将标题包含 title 的窗口设为置顶/取消置顶。
func (p *WindowsPlatform) SetWindowTopmost(title string, topmost bool) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return fmt.Errorf("window title is required")
	}
	after := "-2" // HWND_NOTOPMOST
	if topmost {
		after = "-1" // HWND_TOPMOST
	}
	script := fmt.Sprintf(`
Add-Type @"
using System;
using System.Runtime.InteropServices;
using System.Text;
public class RuleCraftWin {
  [DllImport("user32.dll")] public static extern bool SetWindowPos(IntPtr hWnd, IntPtr hAfter, int x, int y, int cx, int cy, uint uFlags);
  [DllImport("user32.dll", CharSet=CharSet.Unicode)] public static extern int GetWindowText(IntPtr hWnd, StringBuilder s, int n);
  [DllImport("user32.dll")] public static extern bool EnumWindows(EnumWindowsProc cb, IntPtr lp);
  public delegate bool EnumWindowsProc(IntPtr h, IntPtr lp);
  public static void SetTop(string needle, int after) {
    EnumWindows((h, lp) => {
      var sb = new StringBuilder(512);
      GetWindowText(h, sb, 512);
      if (sb.ToString().IndexOf(needle, StringComparison.OrdinalIgnoreCase) >= 0) {
        SetWindowPos(h, (IntPtr)after, 0, 0, 0, 0, 0x0001|0x0002|0x0010);
      }
      return true;
    }, IntPtr.Zero);
  }
}
"@
[RuleCraftWin]::SetTop('%s', %s)
`, escapePS(title), after)
	_, err := runPowerShell(script, 12)
	return err
}

// SetProcessPriority 设置进程优先级。
func (p *WindowsPlatform) SetProcessPriority(name, priority string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("process name is required")
	}
	procName := name
	if !strings.HasSuffix(strings.ToLower(procName), ".exe") {
		procName += ".exe"
	}

	prioMap := map[string]int{
		"idle":         64,
		"below_normal": 16384,
		"normal":       32,
		"above_normal": 32768,
		"high":         128,
	}
	prio, ok := prioMap[strings.ToLower(priority)]
	if !ok {
		return fmt.Errorf("invalid priority: %s", priority)
	}

	// wmic 优先，失败则 PowerShell
	cmd := exec.Command("wmic", "process", "where",
		fmt.Sprintf("name='%s'", strings.ReplaceAll(procName, "'", "")),
		"call", "setpriority", strconv.Itoa(prio))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := cmd.CombinedOutput(); err == nil {
		return nil
	} else {
		script := fmt.Sprintf(
			`Get-Process -Name '%s' -ErrorAction SilentlyContinue | ForEach-Object { $_.PriorityClass = [Diagnostics.ProcessPriorityClass]::%s }`,
			escapePS(strings.TrimSuffix(procName, ".exe")),
			mapPriorityName(priority))
		if _, err2 := runPowerShell(script, 10); err2 != nil {
			return fmt.Errorf("set priority failed: %v / %v: %s", err, err2, string(out))
		}
	}
	return nil
}

func mapPriorityName(p string) string {
	switch strings.ToLower(p) {
	case "idle":
		return "Idle"
	case "below_normal":
		return "BelowNormal"
	case "above_normal":
		return "AboveNormal"
	case "high":
		return "High"
	default:
		return "Normal"
	}
}
