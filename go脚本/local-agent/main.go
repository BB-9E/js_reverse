package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows/registry"
)

const (
	listenAddr = "127.0.0.1:17890"
	runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	runKeyName = "LocalGoAgent"
)

// ========================
// 程序入口
// ========================
func main() {
	// ---- 日志 ----
	f, _ := os.OpenFile("agent.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	log.SetOutput(f)

	// ---- 防重复启动 ----
	if !singleInstance() {
		log.Println("agent already running")
		return
	}

	// ---- HTTP 服务 ----
	go startHTTP()

	// ---- 托盘 ----
	systray.Run(onTrayReady, onTrayExit)
}

// ========================
// HTTP 服务
// ========================
func startHTTP() {
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/open-folder", openFolderHandler)

	mux.HandleFunc("/autostart/status", autoStartStatusHandler)
	mux.HandleFunc("/autostart/enable", autoStartEnableHandler)
	mux.HandleFunc("/autostart/disable", autoStartDisableHandler)

	mux.HandleFunc("/uninstall", uninstallHandler)

	log.Println("listening on", listenAddr)
	err := http.ListenAndServe(listenAddr, mux)
	if err != nil {
		log.Println("http error:", err)
	}
}

// ========================
// 托盘
// ========================
func onTrayReady() {
	systray.SetTooltip("Local Go Agent")

	mOpen := systray.AddMenuItem("打开 Web 界面", "")
	systray.AddSeparator()

	var mToggle *systray.MenuItem
	if isAutoStartEnabled() {
		mToggle = systray.AddMenuItem("关闭开机启动", "")
	} else {
		mToggle = systray.AddMenuItem("开启开机启动", "")
	}

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出", "")

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				exec.Command("cmd", "/c", "start", "http://127.0.0.1:17890").Start()

			case <-mToggle.ClickedCh:
				if isAutoStartEnabled() {
					_ = disableAutoStart()
					mToggle.SetTitle("开启开机启动")
				} else {
					_ = enableAutoStart()
					mToggle.SetTitle("关闭开机启动")
				}
			case <-mQuit.ClickedCh:
				_ = disableAutoStart()
				systray.Quit()
				os.Exit(0)
			}
		}
	}()
}

func onTrayExit() {}

// ========================
// 文件夹打开
// ========================
func openFolderHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path is required", 400)
		return
	}

	path = strings.ReplaceAll(path, "/", `\`)
	absPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if _, err := os.Stat(absPath); err != nil {
		http.Error(w, "path not exists", 400)
		return
	}

	exec.Command("explorer.exe", absPath).Start()
	w.Write([]byte("opened"))
}

// ========================
// HTTP - 自启动
// ========================
func autoStartStatusHandler(w http.ResponseWriter, r *http.Request) {
	if isAutoStartEnabled() {
		w.Write([]byte("enabled"))
	} else {
		w.Write([]byte("disabled"))
	}
}

func autoStartEnableHandler(w http.ResponseWriter, r *http.Request) {
	if err := enableAutoStart(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte("enabled"))
}

func autoStartDisableHandler(w http.ResponseWriter, r *http.Request) {
	if err := disableAutoStart(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte("disabled"))
}

// 卸载
func uninstallHandler(w http.ResponseWriter, r *http.Request) {
	_ = disableAutoStart()
	w.Write([]byte("uninstalled"))

	go func() {
		systray.Quit()
		os.Exit(0)
	}()
}

// 注册表
func isAutoStartEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue(runKeyName)
	return err == nil
}

func enableAutoStart() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	key, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.SetStringValue(runKeyName, exePath)
}

func disableAutoStart() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.DeleteValue(runKeyName)
}

// 防重复启动
func singleInstance() bool {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procCreateMutex := kernel32.NewProc("CreateMutexW")

	name, _ := syscall.UTF16PtrFromString("LocalGoAgentSingleton")
	handle, _, _ := procCreateMutex.Call(0, 1, uintptr(unsafe.Pointer(name)))

	const ERROR_ALREADY_EXISTS = 183
	return handle != 0 && syscall.GetLastError() != syscall.Errno(ERROR_ALREADY_EXISTS)
}
