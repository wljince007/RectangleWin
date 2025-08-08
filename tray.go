//go:build linux

package main

import (
	_ "embed"
	"os/exec"

	"github.com/cihub/seelog"
)

// --- systray dummy stub for Linux ---
type dummyMenuItem struct {
	ClickedCh chan struct{}
}

func (d *dummyMenuItem) Checked() bool     { return false }
func (d *dummyMenuItem) Check()            {}
func (d *dummyMenuItem) Uncheck()          {}
func (d *dummyMenuItem) SetTitle(s string) {}

var systray = struct {
	Run                 func(onReady, onExit func())
	SetIcon             func([]byte)
	SetTitle            func(string)
	SetTooltip          func(string)
	AddMenuItem         func(string, string) *dummyMenuItem
	AddMenuItemCheckbox func(string, string, bool) *dummyMenuItem
	AddSeparator        func()
	Quit                func()
}{
	Run:                 func(onReady, onExit func()) { onReady(); onExit() },
	SetIcon:             func([]byte) {},
	SetTitle:            func(string) {},
	SetTooltip:          func(string) {},
	AddMenuItem:         func(string, string) *dummyMenuItem { return &dummyMenuItem{ClickedCh: make(chan struct{})} },
	AddMenuItemCheckbox: func(string, string, bool) *dummyMenuItem { return &dummyMenuItem{ClickedCh: make(chan struct{})} },
	AddSeparator:        func() {},
	Quit:                func() {},
}

// --- end systray dummy stub ---

//go:embed assets/tray_icon.png
var icon []byte

const repo = "https://github.com/ahmetb/RectangleLinux"

func initTray() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(icon)
	systray.SetTitle("RectangleWin")
	systray.SetTooltip("RectangleWin")

	autorun, err := AutoRunEnabled()
	if err != nil {
		seelog.Errorf("autorun state error: %v\n", err)
		autorun = false
	}

	mRepo := systray.AddMenuItem("Documentation", "")
	go func() {
		for range mRepo.ClickedCh {
			// Linux下用xdg-open打开浏览器
			if err := exec.Command("xdg-open", repo).Start(); err != nil {
				seelog.Errorf("failed to launch browser: %v\n", err)
			}
		}
	}()

	systray.AddSeparator()

	mAutoRun := systray.AddMenuItemCheckbox("Run on startup", "", autorun)
	go func() {
		for range mAutoRun.ClickedCh {
			if mAutoRun.Checked() {
				if err := AutoRunDisable(); err != nil {
					mAutoRun.SetTitle(err.Error())
					seelog.Errorf("warn: autorun disable: %v\n", err)
					continue
				}
				seelog.Debugf("disabled autorun")
				mAutoRun.Uncheck()
			} else {
				if err := AutoRunEnable(); err != nil {
					mAutoRun.SetTitle(err.Error())
					seelog.Errorf("warn: autorun enable: %v\n", err)
					continue
				}
				seelog.Debugf("enabled autorun")
				mAutoRun.Check()
			}
		}
	}()

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "")
	go func() {
		<-mQuit.ClickedCh
		seelog.Debugf("clicked Quit")
		systray.Quit()
	}()

	seelog.Debugf("tray ready")
}

func onExit() {
	seelog.Debugf("onExit invoked")
}
