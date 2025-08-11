//go:build linux

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"runtime/debug"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/ahmetb/RectangleLinux/w32"
	"github.com/cihub/seelog"

	"github.com/ahmetb/RectangleLinux/w32ex"
)

var lastResized w32.HWND

func main() {
	logger, err := seelog.LoggerFromConfigAsFile("RectangleLinux.xml")
	if err != nil {
		panic(err)
	}
	seelog.ReplaceLogger(logger)
	defer seelog.Flush()

	defer func() {
		if errErr := recover(); errErr != nil {
			seelog.Warnf("defer err:%v, satck:%v", errErr, string(debug.Stack()))
		}
	}()

	runtime.LockOSThread() // since we bind hotkeys etc that need to dispatch their message here
	if !w32ex.SetProcessDPIAware() {
		seelog.Criticalf("failed to set DPI aware")
		panic("failed to set DPI aware")
	}

	autorun, err := AutoRunEnabled()
	if err != nil {
		seelog.Criticalf("AutoRunEnabled err:%v", err)
		panic(err)
	}
	seelog.Debugf("autorun enabled=%v", autorun)
	printMonitors()

	// Connect to the X server using the DISPLAY environment variable.
	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	// Anytime the keybind (mousebind) package is used, keybind.Initialize
	// *should* be called once. It isn't strictly necessary, but allows your
	// keybindings to persist even if the keyboard mapping is changed during
	// run-time. (Assuming you're using the xevent package's event loop.)
	keybind.Initialize(X)

	// 退出时解关联
	defer func() {
		keybind.Detach(X, X.RootWin())
	}()

	edgeFuncs := [][]resizeFunc{
		{leftTwoThirds, leftHalf, leftOneThirds},
		{rightTwoThirds, rightHalf, rightOneThirds},
		{topTwoThirds, topHalf, topOneThirds},
		{bottomTwoThirds, bottomHalf, bottomOneThirds}}
	edgeFuncTurn := make([]int, len(edgeFuncs))
	cornerFuncs := [][]resizeFunc{
		{topLeftOneThirds, topLeftHalf, topLeftTwoThirds},
		{topRightOneThirds, topRightHalf, topRightTwoThirds},
		{bottomLeftOneThirds, bottomLeftHalf, bottomLeftTwoThirds},
		{bottomRightOneThirds, bottomRightHalf, bottomRightTwoThirds}}
	cornerFuncTurn := make([]int, len(cornerFuncs))

	cycleFuncs := func(funcs [][]resizeFunc, turns *[]int, i int) {
		hwnd := w32.GetForegroundWindow()
		if hwnd == 0 {
			seelog.Criticalf("foreground window is NULL")
			panic("foreground window is NULL")
		}
		if lastResized != hwnd {
			*turns = make([]int, len(edgeFuncs)) // reset
		}
		if _, err := resize(hwnd, funcs[i][(*turns)[i]%len(funcs[i])]); err != nil {
			seelog.Errorf("warn: resize: %v", err)
			return
		}
		(*turns)[i]++
		for j := 0; j < len(*turns); j++ {
			if j != i {
				(*turns)[j] = 0
			}
		}
	}

	cycleEdgeFuncs := func(i int) { cycleFuncs(edgeFuncs, &edgeFuncTurn, i) }
	cycleCornerFuncs := func(i int) { cycleFuncs(cornerFuncs, &cornerFuncTurn, i) }

	registerFunc := func(keystr string, Handler func()) {
		err = keybind.KeyPressFun(func(X *xgbutil.XUtil, e xevent.KeyPressEvent) {
			seelog.Debugf("keystr：%v callback", keystr)
			Handler()
		}).Connect(X, X.RootWin(), keystr, true) // true 表示自动抓取按键
		if err != nil {
			seelog.Warnf("keystr：%v 注册失败", keystr)
		}
	}

	// 按键名称参考 ：sgithub.com/BurntSushi/xgbutil/keybind/keysymdef.go
	// Linux 下的热键定义，Key 字符串格式如 "Control-Mod1-Left"
	// // 绑定全局快捷键 "Ctrl+Alt+G"，这个+的写法是不行的,应该是"Control-mod1-G"
	// keystr := "Mod4-G" //Super+g
	// keystr = "control-G"
	// keystr = "shift-G"
	// keystr = "control-shift-G"
	// keystr = "mod1-G" //alt+g
	// keystr = "Control-mod1-G"
	registerFunc("Control-Mod1-Left", func() { cycleEdgeFuncs(0) })
	registerFunc("Control-Mod1-Right", func() { cycleEdgeFuncs(1) })
	registerFunc("Control-Mod1-Up", func() { cycleEdgeFuncs(2) })
	registerFunc("Control-Mod1-Down", func() { cycleEdgeFuncs(3) })
	registerFunc("Control-Mod1-2", func() { cycleCornerFuncs(0) })
	registerFunc("Control-Mod1-1", func() { cycleCornerFuncs(1) })
	registerFunc("Control-Mod1-3", func() { cycleCornerFuncs(2) })
	registerFunc("Control-Mod1-4", func() { cycleCornerFuncs(3) })
	registerFunc("Control-Mod1-Shift-F", func() {
		lastResized = 0
		if err := maximize(); err != nil {
			seelog.Errorf("warn: maximize: %v", err)
			return
		}
	})
	registerFunc("Control-Mod1-Shift-C", func() {
		lastResized = 0
		if _, err := resize(w32.GetForegroundWindow(), center); err != nil {
			seelog.Errorf("warn: resize: %v", err)
			return
		}
	})
	registerFunc("Control-Mod1-Shift-A", func() {
		hwnd := w32.GetForegroundWindow()
		if err := toggleAlwaysOnTop(hwnd); err != nil {
			seelog.Errorf("warn: toggleAlwaysOnTop: %v", err)
			return
		}
		seelog.Debugf("> toggled always on top: %v", hwnd)
	})

	exitCh := make(chan os.Signal)
	signal.Notify(exitCh, os.Interrupt)
	go func() {
		<-exitCh
		seelog.Warnf("exit signal received")
		os.Exit(0)
		// systray.Quit() // causes WM_CLOSE, WM_QUIT, not sure if a side-effect
	}()

	// TODO systray/systray.go already locks the OS thread in init()
	// however it's not clear if GetMessage(0,0) will continue to work
	// as we run "go initTray()" and not pin the thread that initializes the
	// tray.
	initTray()

	// if err := msgLoop(); err != nil {
	// 	panic(err)
	// }
	// Finally, start the main event loop. This will route any appropriate
	// KeyPressEvents to your callback function.
	log.Println("Program initialized. Start pressing keys!")
	xevent.Main(X)
}

func showMessageBox(text string) {
	w32.MessageBox(w32.GetActiveWindow(), text, "RectangleWin", w32.MB_ICONWARNING|w32.MB_OK)
}

type resizeFunc func(disp, cur w32.RECT) w32.RECT

func center(disp, cur w32.RECT) w32.RECT {
	// TODO find a way to round up divisions consistently as it causes multiple runs to shift by 1px
	w := (disp.Width() - cur.Width()) / 2
	h := (disp.Height() - cur.Height()) / 2
	return w32.RECT{
		Left:   disp.Left + w,
		Right:  disp.Left + w + cur.Width(),
		Top:    disp.Top + h,
		Bottom: disp.Top + h + cur.Height()}
}

func resize(hwnd w32.HWND, f resizeFunc) (bool, error) {
	if !isZonableWindow(hwnd) {
		seelog.Errorf("warn: non-zonable window: %s", w32.GetWindowText(hwnd))
		return false, nil
	}
	rect := w32.GetWindowRect(hwnd)
	mon := w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST)
	hdc := w32.GetDC(hwnd)
	displayDPI := w32.GetDeviceCaps(hdc, w32.LOGPIXELSY)
	if !w32.ReleaseDC(hwnd, hdc) {
		return false, fmt.Errorf("failed to ReleaseDC:%d", w32.GetLastError())
	}
	var monInfo w32.MONITORINFO
	if !w32.GetMonitorInfo(mon, &monInfo) {
		return false, fmt.Errorf("failed to GetMonitorInfo:%d", w32.GetLastError())
	}

	ok, frame := w32.DwmGetWindowAttributeEXTENDED_FRAME_BOUNDS(hwnd)
	if !ok {
		return false, fmt.Errorf("failed to DwmGetWindowAttributeEXTENDED_FRAME_BOUNDS:%d", w32.GetLastError())
	}
	windowDPI := w32ex.GetDpiForWindow(hwnd)
	resizedFrame := resizeForDpi(frame, int32(windowDPI), int32(displayDPI))

	seelog.Debugf("> window: 0x%x %#v (w:%d,h:%d) mon=0x%X(@ display DPI:%d)", hwnd, rect, rect.Width(), rect.Height(), mon, displayDPI)
	seelog.Debugf("> DWM frame:        %#v (W:%d,H:%d) @ window DPI=%v", frame, frame.Width(), frame.Height(), windowDPI)
	seelog.Debugf("> DPI-less frame:   %#v (W:%d,H:%d)", resizedFrame, resizedFrame.Width(), resizedFrame.Height())

	// calculate how many extra pixels go to win10 invisible borders
	lExtra := resizedFrame.Left - rect.Left
	rExtra := -resizedFrame.Right + rect.Right
	tExtra := resizedFrame.Top - rect.Top
	bExtra := -resizedFrame.Bottom + rect.Bottom

	newPos := f(monInfo.RcWork, resizedFrame)

	// adjust offsets based on invisible borders
	newPos.Left -= lExtra
	newPos.Top -= tExtra
	newPos.Right += rExtra
	newPos.Bottom += bExtra

	lastResized = hwnd
	if sameRect(rect, &newPos) {
		seelog.Debugf("no resize")
		return false, nil
	}

	seelog.Debugf("> resizing to: %#v (W:%d,H:%d)", newPos, newPos.Width(), newPos.Height())
	if !w32.ShowWindow(hwnd, w32.SW_SHOWNORMAL) { // normalize window first if it's set to SW_SHOWMAXIMIZE (and therefore stays maximized)
		return false, fmt.Errorf("failed to normalize window ShowWindow:%d", w32.GetLastError())
	}
	if !w32.SetWindowPos(hwnd, 0, int(newPos.Left), int(newPos.Top), int(newPos.Width()), int(newPos.Height()), w32.SWP_NOZORDER|w32.SWP_NOACTIVATE) {
		return false, fmt.Errorf("failed to SetWindowPos:%d", w32.GetLastError())
	}
	rect = w32.GetWindowRect(hwnd)
	seelog.Debugf("> post-resize: %#v(W:%d,H:%d)", rect, rect.Width(), rect.Height())
	return true, nil
}

func maximize() error {
	hwnd := w32.GetForegroundWindow()
	if !isZonableWindow(hwnd) {
		return errors.New("foreground window is not zonable")
	}
	if !w32.ShowWindow(hwnd, w32.SW_MAXIMIZE) {
		return fmt.Errorf("failed to ShowWindow:%d", w32.GetLastError())
	}
	return nil
}

func toggleAlwaysOnTop(hwnd w32.HWND) error {
	if !isZonableWindow(hwnd) {
		return errors.New("foreground window is not zonable")
	}

	if w32.GetWindowLong(hwnd, w32.GWL_EXSTYLE)&w32.WS_EX_TOPMOST != 0 {
		if !w32.SetWindowPos(hwnd, w32.HWND_NOTOPMOST, 0, 0, 0, 0, w32.SWP_NOMOVE|w32.SWP_NOSIZE) {
			return fmt.Errorf("failed to SetWindowPos(HWND_NOTOPMOST): %v", w32.GetLastError())
		}
	} else {
		if !w32.SetWindowPos(hwnd, w32.HWND_TOPMOST, 0, 0, 0, 0, w32.SWP_NOMOVE|w32.SWP_NOSIZE) {
			return fmt.Errorf("failed to SetWindowPos(HWND_TOPMOST) :%v", w32.GetLastError())
		}
	}
	return nil
}

func resizeForDpi(src w32.RECT, from, to int32) w32.RECT {
	return w32.RECT{
		Left:   src.Left * to / from,
		Right:  src.Right * to / from,
		Top:    src.Top * to / from,
		Bottom: src.Bottom * to / from,
	}
}

func sameRect(a, b *w32.RECT) bool {
	return a != nil && b != nil && reflect.DeepEqual(*a, *b)
}

// 字符串包含判断（文件级作用域）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || (len(s) > len(substr)+1 && s[len(s)-len(substr)-1:] == "-"+substr))))
}
