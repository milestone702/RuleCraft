// RuleCraft — Windows 轻量级自动化规则引擎
//
// 使用方式：
//   rulecraft                    # 启动（默认端口 19530）
//   rulecraft -port 8080         # 指定端口
//   rulecraft -config config.json # 指定配置文件
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"rulecraft/api"
	"rulecraft/config"
	"rulecraft/engine"
	"rulecraft/gui"
	"rulecraft/input"
	"rulecraft/output"
	"rulecraft/platform"
	"rulecraft/plugin"
)

// ============================================================================
// 命令行参数
// ============================================================================

var (
	flagPort        = flag.Int("port", 19530, "Web 服务端口")
	flagConfig      = flag.String("config", "config.json", "配置文件路径")
	flagPoll        = flag.Int("interval", 5, "轮询间隔（秒）")
	flagShowVersion = flag.Bool("version", false, "显示版本信息")
)

// 构建时注入的版本号
var version = "1.0.0-dev"

// ============================================================================
// 程序入口
// ============================================================================

func main() {
	flag.Parse()

	if *flagShowVersion {
		fmt.Printf("RuleCraft v%s\n", version)
		os.Exit(0)
	}

	// 0a. 单实例保护 — 防止重复运行
	if !trySingleInstance() {
		showMessageBox("RuleCraft 已在运行",
			"RuleCraft 已经启动，无法重复运行。\n请检查系统托盘区域。")
		os.Exit(0)
	}

	// 0b. 检查端口是否可用
	addr := fmt.Sprintf("127.0.0.1:%d", *flagPort)
	if err := checkPortAvailable(addr); err != nil {
		log.Printf("[main] port %d is not available: %v", *flagPort, err)
		showMessageBox("端口被占用",
			fmt.Sprintf("端口 %d 已被其他程序占用。\n请关闭占用程序后重试，或使用 -port 参数指定其他端口。", *flagPort))
		os.Exit(1)
	}

	// 1. 加载配置
	appCfg := loadConfig(*flagConfig)

	// 1b. 确保数据目录存在（全新环境自动创建）
	ensureDir("Input_Plugins")
	ensureDir("Output_Plugins")
	ensureDir("tasks")

	// 2. 初始化日志系统
	logMgr := engine.NewLogManager(*appCfg.Logging)
	if err := logMgr.InitSystem("logs/app.log"); err != nil {
		log.Fatalf("init log system failed: %v", err)
	}
	defer logMgr.CloseAll()
	log.Println("[main] RuleCraft starting...")

	// 3. 初始化平台层
	plat := platform.NewPlatform()
	log.Printf("[main] platform: %T initialized", plat)

	// 4. 初始化插件注册表
	registry := plugin.NewRegistry()

	// 注册内置输入插件
	if err := input.RegisterBuiltinInputs(registry, plat); err != nil {
		log.Fatalf("register input plugins failed: %v", err)
	}
	log.Printf("[main] registered %d input plugins", len(registry.ListInputs()))

	// 注册内置输出插件
	if err := output.RegisterBuiltinOutputs(registry, plat); err != nil {
		log.Fatalf("register output plugins failed: %v", err)
	}
	log.Printf("[main] registered %d output plugins", len(registry.ListOutputs()))

	// 5. 初始化通知管理器
	notifier := engine.NewNotifier(*appCfg.Notifications)

	// 6. 初始化引擎
	runner := engine.NewRunner(appCfg, registry)

	// 6b. 初始化文件加载器 + 任务调度器
	loader := engine.NewPluginManager("Input_Plugins", "Output_Plugins", "tasks")
	scheduler := engine.NewTaskScheduler(runner, loader)
	runner.SetScheduler(scheduler)

	// 7. 启动轮询引擎
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := runner.Start(ctx); err != nil {
		log.Fatalf("start runner failed: %v", err)
	}

	// 8. 设置 GUI 版本号
	gui.SetVersion(version)

	// 9. 启动 Web 服务 + API
	handler := api.NewHandler(runner, registry, logMgr, notifier, appCfg, "Input_Plugins", "Output_Plugins", "tasks")
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Web 服务启停控制（供 OnToggleWeb 回调使用）
	var (
		webSrvRunning atomic.Bool
		webSrvMu      sync.Mutex
		webServer     *http.Server
		webGen        atomic.Int64
	)

	startWebServer := func() {
		webSrvMu.Lock()
		if webSrvRunning.Load() {
			webSrvMu.Unlock()
			return
		}

		// 检查端口是否可用
		addr := fmt.Sprintf("127.0.0.1:%d", *flagPort)
		if err := checkPortAvailable(addr); err != nil {
			webSrvMu.Unlock()
			log.Printf("[main] port %d is not available, cannot start web server: %v", *flagPort, err)
			showMessageBox("端口被占用",
				fmt.Sprintf("端口 %d 已被占用，无法启动 Web 服务。\n请关闭占用程序后重试。", *flagPort))
			return
		}

		gen := webGen.Add(1)
		srv := &http.Server{
			Addr:    addr,
			Handler: mux,
		}
		webServer = srv
		webSrvRunning.Store(true)
		webSrvMu.Unlock()

		go func() {
			log.Printf("[main] Web UI: http://127.0.0.1:%d", *flagPort)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("[main] HTTP server error: %v", err)
			}
			// 只有当前 goroutine 对应的是最新一次 start 时才清除状态
			if gen == webGen.Load() {
				webSrvRunning.Store(false)
			}
		}()
	}

	stopWebServer := func() {
		webSrvMu.Lock()
		defer webSrvMu.Unlock()
		if !webSrvRunning.Load() || webServer == nil {
			return
		}
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		webServer.Shutdown(shutdownCtx)
		webSrvRunning.Store(false)
		log.Println("[main] Web server stopped")
	}

	// 初始启动 Web 服务
	startWebServer()

	// 10. 启动 GUI（阻塞，直到用户退出）
	gui.Run(gui.Config{
		Port:           *flagPort,
		AutoStart:      isAutoStartEnabled(),
		WebEnabled:     true,
		OnOpenWeb: func() {
			openBrowser(fmt.Sprintf("http://127.0.0.1:%d", *flagPort))
		},
		OnToggleWeb: func() bool {
			if webSrvRunning.Load() {
				stopWebServer()
				return false
			} else {
				startWebServer()
				return true
			}
		},
		OnExit: func() {
			log.Println("[main] user requested exit via GUI")
		},
		OnAutoStart: func(enabled bool) {
			setAutoStart(enabled)
		},
		OnPortChange: func(port int) {
			log.Printf("[main] port changed to %d (restart required)", port)
		},
	})

	// 11. 退出清理
	log.Println("[main] shutting down...")
	cancel()
	runner.Stop()
	stopWebServer()
	logMgr.CloseAll()
	log.Println("[main] RuleCraft shutdown complete")
}

// ============================================================================
// 目录创建
// ============================================================================

// ensureDir 确保目录存在，不存在则创建。
func ensureDir(dir string) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("[main] create dir %s failed: %v", dir, err)
	}
}

// ============================================================================
// 配置加载
// ============================================================================

// loadConfig 加载配置文件，如果文件不存在则使用默认配置。
func loadConfig(path string) *config.AppConfig {
	cfg := config.DefaultAppConfig()

	// 命令行参数覆盖默认值
	cfg.Port = *flagPort
	cfg.PollInterval = *flagPoll

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[main] config file not found, using defaults: %s", path)
			return &cfg
		}
		log.Fatalf("[main] read config failed: %v", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("[main] parse config failed: %v", err)
	}

	// 命令行参数优先级更高
	cfg.Port = *flagPort
	cfg.PollInterval = *flagPoll

	return &cfg
}

// ============================================================================
// 辅助函数
// ============================================================================

// openBrowser 在默认浏览器中打开 URL。
func openBrowser(url string) error {
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}

// setAutoStart 设置或取消开机自启（HKCU 注册表）。
func setAutoStart(enabled bool) {
	exePath, _ := os.Executable()
	if exePath == "" {
		return
	}

	if enabled {
		cmd := exec.Command("reg", "add",
			`HKCU\Software\Microsoft\Windows\CurrentVersion\Run`,
			"/v", "RuleCraft", "/t", "REG_SZ", "/d", exePath, "/f")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if err := cmd.Run(); err != nil {
			log.Printf("[main] enable auto-start failed: %v", err)
		} else {
			log.Printf("[main] auto-start enabled: %s", exePath)
		}
	} else {
		cmd := exec.Command("reg", "delete",
			`HKCU\Software\Microsoft\Windows\CurrentVersion\Run`,
			"/v", "RuleCraft", "/f")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if err := cmd.Run(); err != nil {
			log.Printf("[main] disable auto-start failed: %v", err)
		} else {
			log.Println("[main] auto-start disabled")
		}
	}
}

// isAutoStartEnabled 检查开机自启是否已启用。
func isAutoStartEnabled() bool {
	cmd := exec.Command("reg", "query",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Run`,
		"/v", "RuleCraft")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run() == nil
}

// saveConfig 保存配置到文件。
func saveConfig(path string, cfg *config.AppConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config failed: %w", err)
	}

	// 保留原始文件权限
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config failed: %w", err)
	}
	return nil
}

// ============================================================================
// 单实例保护 & 端口检测
// ============================================================================

// mutexName 是用于单实例保护的命名互斥体名称。
const mutexName = "Local\\RuleCraft-Singleton-Mutex"

// trySingleInstance 尝试创建命名互斥体，如果已存在则返回 false。
func trySingleInstance() bool {
	modKernel32 := syscall.NewLazyDLL("kernel32.dll")
	procCreateMutexW := modKernel32.NewProc("CreateMutexW")
	procGetLastError := modKernel32.NewProc("GetLastError")

	namePtr, _ := syscall.UTF16PtrFromString(mutexName)
	ret, _, _ := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(namePtr)))
	if ret == 0 {
		// 创建失败，允许继续（保守策略）
		log.Printf("[main] CreateMutex failed, allowing startup")
		return true
	}

	errCode, _, _ := procGetLastError.Call()
	if errCode == 183 { // ERROR_ALREADY_EXISTS
		// 互斥体已存在，说明已有实例在运行
		procCloseHandle := modKernel32.NewProc("CloseHandle")
		procCloseHandle.Call(ret)
		log.Printf("[main] another instance is already running")
		return false
	}

	return true
}

// checkPortAvailable 检查 TCP 端口是否可用，不可用返回错误。
func checkPortAvailable(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	ln.Close()
	return nil
}

// showMessageBox 显示 Windows 消息框（在 GUI 未初始化时也可用）。
func showMessageBox(title, text string) {
	modUser32 := syscall.NewLazyDLL("user32.dll")
	procMessageBoxW := modUser32.NewProc("MessageBoxW")

	titlePtr, _ := syscall.UTF16PtrFromString(title)
	textPtr, _ := syscall.UTF16PtrFromString(text)

	procMessageBoxW.Call(0,
		uintptr(unsafe.Pointer(textPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		0x00000010|0x00000000) // MB_ICONSTOP | MB_OK
}
