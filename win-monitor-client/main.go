package main

import (
	"context"
	_ "embed"
	"errors"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/systray"
)

//go:embed icon.ico
var iconbs []byte

const (
	appName = "win-monitor-client"
	appId   = "win-monitor-client"
)

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	monitor := NewMonitorWindow(ctx, cancelFunc)
	monitor.msgChan = make(chan string, 10)
	fyneApp := app.NewWithID(appId)
	monitor.fyneApp = fyneApp

	go systray.Run(monitor.onReady, monitor.onExit)

	go func() {
		for {
			select {
			case msg := <-monitor.msgChan:
				log.Println(msg)
				fyneApp.SendNotification(fyne.NewNotification("通知", msg))
			case <-ctx.Done():
				return
			}
		}
	}()

	client := NewClient(ctx)
	go client.Start()

	fyneApp.Run()
}

type MonitorWindow struct {
	AutoStart bool
	fyneApp   fyne.App
	ctx       context.Context
	cancel    context.CancelFunc
	msgChan   chan string
}

func NewMonitorWindow(ctx context.Context, cancel context.CancelFunc) *MonitorWindow {
	return &MonitorWindow{ctx: ctx, cancel: cancel}
}

func (m *MonitorWindow) checkAutoStart() bool {
	current, err := user.Current()
	if err != nil {
		return false
	}
	startupPath := filepath.Join(current.HomeDir, "AppData/Roaming/Microsoft/Windows/Start Menu/Programs/Startup")
	shortcutPath := filepath.Join(startupPath, appName+".lnk")
	if _, err := os.Stat(shortcutPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		}
	}
	return true
}

func (m *MonitorWindow) setAutoStart() error {
	if m.checkAutoStart() {
		return nil
	}
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	startupPath := filepath.Join(currentUser.HomeDir, "AppData/Roaming/Microsoft/Windows/Start Menu/Programs/Startup")
	shortcutPath := filepath.Join(startupPath, appName+".lnk")
	if err := os.Symlink(executable, shortcutPath); err != nil {
		return errors.New("权限不足，请以管理员运行")
	}
	return nil
}

func (m *MonitorWindow) removeAutoStart() error {
	current, err := user.Current()
	if err != nil {
		return err
	}
	startupPath := filepath.Join(current.HomeDir, "AppData/Roaming/Microsoft/Windows/Start Menu/Programs/Startup")
	shortcutPath := filepath.Join(startupPath, appName+".lnk")
	if err := os.Remove(shortcutPath); err != nil {
		return err
	}
	return nil
}

// 当系统托盘准备好后调用
func (m *MonitorWindow) onReady() {
	// 设置托盘图标和提示文本
	systray.SetIcon(iconbs) // 替换成你的图标路径
	systray.SetTitle(appName)
	systray.SetTooltip(appName)
	// 添加菜单项
	autoStart := systray.AddMenuItemCheckbox("开机自启动", "开机自启动", m.AutoStart)
	mQuit := systray.AddMenuItem("退出", "退出")
	for {
		select {
		case <-mQuit.ClickedCh:
			systray.Quit()
		case <-autoStart.ClickedCh:
			if !autoStart.Checked() {
				err := m.setAutoStart()
				if err != nil {
					autoStart.Uncheck()
					m.msgChan <- err.Error()
				}
				autoStart.Check()
			} else {
				err := m.removeAutoStart()
				if err != nil {
					m.msgChan <- err.Error()
				}
				autoStart.Uncheck()
			}
		}
	}
}
func (m *MonitorWindow) onExit() {
	m.cancel()
	m.fyneApp.Quit()
	os.Exit(0)
}

func (m *MonitorWindow) Exit() {
	if m.cancel != nil {
		m.cancel()
	}
	m.fyneApp.Quit()
}
