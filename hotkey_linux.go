//go:build linux

package main

import (
	"fmt"
	"sync"

	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/ahmetb/RectangleWin/w32"
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

// 注册全局热键，Key为X11 keycode
func RegisterHotKey(h HotKey) bool {
	hotkeyMu.Lock()
	defer hotkeyMu.Unlock()
	if _, ok := hotkeyRegistrations[h.Id]; ok {
		panic("hotkey id already registered")
	}
	xu := w32.GetXConn()
	if xu == nil {
		return false
	}
	// 组合键字符串，如 "Control-Alt-T"
	// keyStr := keyString(h.Mod, h.Key)
	// err := keybind.Register(xu, xu.RootWin(), keyStr, func() {
	// 	if cb := h.Handler; cb != nil {
	// 		cb()
	// 	}
	// })
	// if err == nil {
	// 	hotkeyRegistrations[h.Id] = &h
	// 	return true
	// }
	// fmt.Printf("failed to register hotkey: %s (%v)\n", keyStr, err)
	return false
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
	fmt.Println("hotkey event loop started")
	xevent.Main(xu)
	fmt.Println("hotkey event loop finished")
	return nil
}
