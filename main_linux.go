//go:build linux

package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"runtime"

	"github.com/ahmetb/RectangleWin/w32"
	"github.com/getlantern/systray"

	"github.com/ahmetb/RectangleWin/w32ex"
)

var lastResized w32.HWND

func main() {
	runtime.LockOSThread() // since we bind hotkeys etc that need to dispatch their message here
	if !w32ex.SetProcessDPIAware() {
		panic("failed to set DPI aware")
	}

	autorun, err := AutoRunEnabled()
	if err != nil {
		panic(err)
	}
	fmt.Printf("autorun enabled=%v\n", autorun)
	printMonitors()

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
			panic("foreground window is NULL")
		}
		if lastResized != hwnd {
			*turns = make([]int, len(edgeFuncs)) // reset
		}
		if _, err := resize(hwnd, funcs[i][(*turns)[i]%len(funcs[i])]); err != nil {
			fmt.Printf("warn: resize: %v\n", err)
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

	// Linux 下的热键定义，Key 字符串格式如 "Control-Alt-Left"
	// 辅助函数：将字符串热键描述解析为 (mod, keycode)
	parseHotkeyString := func(s string) (mod int, key int) {
		// 简化实现：仅支持部分常用组合和方向键/数字/F键
		// 实际项目可根据 keymap.go 或 X11 键码表完善
		mod = 0
		key = 0
		if s == "" {
			return
		}
		if contains(s, "Control") {
			mod |= MOD_CONTROL
		}
		if contains(s, "Alt") {
			mod |= MOD_ALT
		}
		if contains(s, "Shift") {
			mod |= MOD_SHIFT
		}
		if contains(s, "Win") {
			mod |= MOD_WIN
		}
		switch {
		case contains(s, "Left"):
			key = 0x25 // 左方向键
		case contains(s, "Right"):
			key = 0x27 // 右方向键
		case contains(s, "Up"):
			key = 0x26 // 上方向键
		case contains(s, "Down"):
			key = 0x28 // 下方向键
		case contains(s, "1"):
			key = 0x31
		case contains(s, "2"):
			key = 0x32
		case contains(s, "3"):
			key = 0x33
		case contains(s, "4"):
			key = 0x34
		case contains(s, "F"):
			key = 0x46 // F
		case contains(s, "C"):
			key = 0x43 // C
		case contains(s, "A"):
			key = 0x41 // A
		default:
			key = 0
		}
		return
	}

	// 统一用 int 类型的 Key 字段
	hks := []HotKey{
		func() HotKey {
			m, k := parseHotkeyString("Control-Alt-Left")
			return HotKey{Id: 1, Mod: m, Key: k, Handler: func() { cycleEdgeFuncs(0) }}
		}(),
		func() HotKey {
			m, k := parseHotkeyString("Control-Alt-Right")
			return HotKey{Id: 2, Mod: m, Key: k, Handler: func() { cycleEdgeFuncs(1) }}
		}(),
		func() HotKey {
			m, k := parseHotkeyString("Control-Alt-Up")
			return HotKey{Id: 3, Mod: m, Key: k, Handler: func() { cycleEdgeFuncs(2) }}
		}(),
		func() HotKey {
			m, k := parseHotkeyString("Control-Alt-Down")
			return HotKey{Id: 4, Mod: m, Key: k, Handler: func() { cycleEdgeFuncs(3) }}
		}(),
		func() HotKey {
			m, k := parseHotkeyString("Control-Alt-2")
			return HotKey{Id: 5, Mod: m, Key: k, Handler: func() { cycleCornerFuncs(0) }}
		}(),
		func() HotKey {
			m, k := parseHotkeyString("Control-Alt-1")
			return HotKey{Id: 6, Mod: m, Key: k, Handler: func() { cycleCornerFuncs(1) }}
		}(),
		func() HotKey {
			m, k := parseHotkeyString("Control-Alt-3")
			return HotKey{Id: 7, Mod: m, Key: k, Handler: func() { cycleCornerFuncs(2) }}
		}(),
		func() HotKey {
			m, k := parseHotkeyString("Control-Alt-4")
			return HotKey{Id: 8, Mod: m, Key: k, Handler: func() { cycleCornerFuncs(3) }}
		}(),
		func() HotKey {
			m, k := parseHotkeyString("Control-Alt-Shift-F")
			return HotKey{Id: 50, Mod: m, Key: k, Handler: func() {
				lastResized = 0
				if err := maximize(); err != nil {
					fmt.Printf("warn: maximize: %v\n", err)
					return
				}
			}}
		}(),
		func() HotKey {
			m, k := parseHotkeyString("Control-Alt-Shift-C")
			return HotKey{Id: 60, Mod: m, Key: k, Handler: func() {
				lastResized = 0
				if _, err := resize(w32.GetForegroundWindow(), center); err != nil {
					fmt.Printf("warn: resize: %v\n", err)
					return
				}
			}}
		}(),
		func() HotKey {
			m, k := parseHotkeyString("Control-Alt-Shift-A")
			return HotKey{Id: 70, Mod: m, Key: k, Handler: func() {
				hwnd := w32.GetForegroundWindow()
				if err := toggleAlwaysOnTop(hwnd); err != nil {
					fmt.Printf("warn: toggleAlwaysOnTop: %v\n", err)
					return
				}
				fmt.Printf("> toggled always on top: %v\n", hwnd)
			}}
		}(),
	}

	// Linux 下批量注册热键
	var failedHotKeys []HotKey
	for i, hk := range hks {
		ok := RegisterHotKey(hk)
		if !ok {
			failedHotKeys = append(failedHotKeys, hk)
		}
	}
	if len(failedHotKeys) > 0 {
		fmt.Println("以下热键注册失败（可能已被其他进程占用）：")
		for _, hk := range failedHotKeys {
			fmt.Printf("  - %s\n", hk.Key)
		}
	}

	exitCh := make(chan os.Signal)
	signal.Notify(exitCh, os.Interrupt)
	go func() {
		<-exitCh
		fmt.Println("exit signal received")
		systray.Quit() // causes WM_CLOSE, WM_QUIT, not sure if a side-effect
	}()

	// TODO systray/systray.go already locks the OS thread in init()
	// however it's not clear if GetMessage(0,0) will continue to work
	// as we run "go initTray()" and not pin the thread that initializes the
	// tray.
	initTray()
	if err := msgLoop(); err != nil {
		panic(err)
	}
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
		fmt.Printf("warn: non-zonable window: %s\n", w32.GetWindowText(hwnd))
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

	fmt.Printf("> window: 0x%x %#v (w:%d,h:%d) mon=0x%X(@ display DPI:%d)\n", hwnd, rect, rect.Width(), rect.Height(), mon, displayDPI)
	fmt.Printf("> DWM frame:        %#v (W:%d,H:%d) @ window DPI=%v\n", frame, frame.Width(), frame.Height(), windowDPI)
	fmt.Printf("> DPI-less frame:   %#v (W:%d,H:%d)\n", resizedFrame, resizedFrame.Width(), resizedFrame.Height())

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
		fmt.Println("no resize")
		return false, nil
	}

	fmt.Printf("> resizing to: %#v (W:%d,H:%d)\n", newPos, newPos.Width(), newPos.Height())
	if !w32.ShowWindow(hwnd, w32.SW_SHOWNORMAL) { // normalize window first if it's set to SW_SHOWMAXIMIZE (and therefore stays maximized)
		return false, fmt.Errorf("failed to normalize window ShowWindow:%d", w32.GetLastError())
	}
	if !w32.SetWindowPos(hwnd, 0, int(newPos.Left), int(newPos.Top), int(newPos.Width()), int(newPos.Height()), w32.SWP_NOZORDER|w32.SWP_NOACTIVATE) {
		return false, fmt.Errorf("failed to SetWindowPos:%d", w32.GetLastError())
	}
	rect = w32.GetWindowRect(hwnd)
	fmt.Printf("> post-resize: %#v(W:%d,H:%d)\n", rect, rect.Width(), rect.Height())
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
