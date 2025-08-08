//go:build linux

package w32ex

import (
	"github.com/ahmetb/RectangleWin/w32"
)

// 仅用于接口兼容，Linux下无实际功能
const (
	GA_PARENT    = 1
	GA_ROOT      = 2
	GA_ROOTOWNER = 3
)

// RegisterHotKey 在 Linux 下无实际作用，仅作占位
func RegisterHotKey(hwnd w32.HWND, id, mod, vk int) bool {
	return false
}

// GetDpiForWindow Linux下直接返回96（常见DPI），仅作占位
func GetDpiForWindow(hwnd w32.HWND) int32 {
	return 96
}

// GetWindowModuleFileName Linux下无实现，返回空字符串
func GetWindowModuleFileName(hwnd w32.HWND) string {
	return ""
}

// GetAncestor Linux下直接返回0
func GetAncestor(hwnd w32.HWND, gaFlags uint) w32.HWND {
	return w32.HWND(0)
}

// GetShellWindow Linux下直接返回0
func GetShellWindow() w32.HWND {
	return w32.HWND(0)
}

// SetProcessDPIAware Linux下无实际作用，仅作占位
func SetProcessDPIAware() bool {
	return false
}
