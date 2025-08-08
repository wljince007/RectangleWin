//go:build linux

package main

import (
	"fmt"
	"sync"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/ahmetb/RectangleLinux/w32"
)

var (
	hotkeyRegistrations = make(map[int]*HotKey)
	hotkeyMu            sync.Mutex
)

// HotKey结构体复用
type HotKey struct {
	Id      int
	Mod     int
	Key     int
	Handler func()
}

func (h HotKey) String() string { return fmt.Sprintf("mod=0x%x,Key=%d", h.Mod, h.Key) }
func (h HotKey) Describe() string {
	var out string
	if h.Mod&MOD_WIN == MOD_WIN {
		out += "Win + "
	}
	if h.Mod&MOD_CONTROL == MOD_CONTROL {
		out += "Ctrl + "
	}
	if h.Mod&MOD_ALT == MOD_ALT {
		out += "Alt + "
	}
	if h.Mod&MOD_SHIFT == MOD_SHIFT {
		out += "Shift + "
	}
	out += fmt.Sprintf("VK_%d", h.Key)
	return out
}

func RegisterHotKey(h HotKey) bool {
	hotkeyMu.Lock()
	defer hotkeyMu.Unlock()
	if _, ok := hotkeyRegistrations[h.Id]; ok {
		panic("hotkey id already registered")
	}
	hotkeyRegistrations[h.Id] = &h
	return true
	// 删除多余的右括号，只保留一个函数结束标志
}

// 生成 keybind 识别的组合键字符串
func keyString(mod, vk int) string {
	var mods []string
	if mod&MOD_CONTROL != 0 {
		mods = append(mods, "Control")
	}
	if mod&MOD_ALT != 0 {
		mods = append(mods, "Mod1")
	}
	if mod&MOD_SHIFT != 0 {
		mods = append(mods, "Shift")
	}
	if mod&MOD_WIN != 0 {
		mods = append(mods, "Mod4")
	}
	// vk 需映射为 X11 key符号，如 "T"、"F1"。这里只作简单数字映射。
	mods = append(mods, fmt.Sprintf("%d", vk))
	return fmt.Sprintf("%s", joinWithDash(mods))
}

func joinWithDash(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += "-" + parts[i]
	}
	return out
}

// 主循环，监听热键事件
func msgLoop() error {
	xu := w32.GetXConn()
	if xu == nil {
		return fmt.Errorf("X connection unavailable")
	}
	// 注册 KeyPress 事件监听器
	xevent.KeyPressFun(func(xu *xgbutil.XUtil, ev xevent.KeyPressEvent) {
		hotkeyMu.Lock()
		defer hotkeyMu.Unlock()
		for _, h := range hotkeyRegistrations {
			// 这里需要你根据项目具体实现补全mod和key的判断逻辑
			if h.Mod == int(ev.State) && h.Key == int(ev.Detail) {
				if cb := h.Handler; cb != nil {
					cb()
				}
			}
		}
	}).Connect(xu, xu.RootWin())
	fmt.Println("hotkey event loop started")
	xevent.Main(xu)
	fmt.Println("hotkey event loop finished")
	return nil
}
