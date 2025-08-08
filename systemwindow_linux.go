//go:build linux

package main

import (
	"strings"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/ahmetb/RectangleWin/w32"
)

func isZonableWindow(hwnd w32.HWND) bool {
	if hwnd == 0 {
		return false
	}
	xu := w32.GetXConn()
	if xu == nil {
		return false
	}
	// 判断窗口类型
	typ, _ := xprop.GetProperty(xu, xproto.Window(hwnd), "_NET_WM_WINDOW_TYPE")
	if typ != nil && strings.Contains(string(typ.Value), "_DOCK") {
		return false
	}
	// 判断窗口已映射且为正常窗口
	attrs, err := xproto.GetWindowAttributes(xu.Conn(), xproto.Window(hwnd)).Reply()
	if err != nil || attrs.MapState != xproto.MapStateViewable {
		return false
	}
	// 判断有无可见owner
	return isStandardWindow(hwnd) && hasNoVisibleOwner(hwnd)
}

func hasNoVisibleOwner(hwnd w32.HWND) bool {
	xu := w32.GetXConn()
	if xu == nil {
		return true
	}
	// TransientFor 相当于 owner
	prop, _ := xprop.GetProperty(xu, xproto.Window(hwnd), "WM_TRANSIENT_FOR")
	if prop == nil || len(prop.Value) < 4 {
		return true
	}
	ownerId := uint32(prop.Value[0]) | uint32(prop.Value[1])<<8 | uint32(prop.Value[2])<<16 | uint32(prop.Value[3])<<24
	attrs, err := xproto.GetWindowAttributes(xu.Conn(), xproto.Window(ownerId)).Reply()
	if err != nil || attrs.MapState != xproto.MapStateViewable {
		return true
	}
	// 判断owner窗口大小
	owner := xwindow.New(xu, xproto.Window(ownerId))
	geom, err := owner.Geometry()
	if err != nil || geom.Width() == 0 || geom.Height() == 0 {
		return true
	}
	return false
}

func isStandardWindow(hwnd w32.HWND) bool {
	if hwnd == 0 {
		return false
	}
	xu := w32.GetXConn()
	if xu == nil {
		return false
	}
	// 常规窗口类型
	typ, _ := xprop.GetProperty(xu, xproto.Window(hwnd), "_NET_WM_WINDOW_TYPE")
	if typ == nil || !strings.Contains(string(typ.Value), "NORMAL") {
		return false
	}
	attrs, err := xproto.GetWindowAttributes(xu.Conn(), xproto.Window(hwnd)).Reply()
	if err != nil || attrs.MapState != xproto.MapStateViewable {
		return false
	}
	// 简单排除桌面、托盘等特殊窗口
	wmName, _ := xprop.GetProperty(xu, xproto.Window(hwnd), "_NET_WM_NAME")
	if wmName != nil && isSystemClassName(string(wmName.Value)) {
		return false
	}
	return true
}

func isSystemClassName(className string) bool {
	// Linux下用名称/类型简单判断
	sysNames := []string{
		"Dock", "Panel", "desktop", "notification", "tray", "bar",
		"SysListView32", "WorkerW", "Shell_TrayWnd", "Shell_SecondaryTrayWnd", "Progman",
	}
	className = strings.ToLower(className)
	for _, n := range sysNames {
		if strings.Contains(className, strings.ToLower(n)) {
			return true
		}
	}
	return false
}
