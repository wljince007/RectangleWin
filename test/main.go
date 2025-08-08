package main

import (
	"fmt"
	"log"
	"os/exec"
	"syscall"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
)

func main() {
	// 初始化 X11 连接
	x, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal("Failed to connect to X server:", err)
	}
	defer x.Conn().Close()

	// 获取屏幕数量
	screenCount := x.NumScreens()

	// 绑定全局快捷键 Ctrl+Shift+Z
	for i := 0; i < screenCount; i++ {
		screen := x.Screen(i)
		rootWindow := screen.Root

		// 获取按键代码
		keysym := xproto.KeysymZ
		keycode, err := xproto.KeysymToKeycode(x.Conn(), keysym)
		if err != nil {
			log.Fatal("Failed to get keycode for keysym:", err)
		}

		// 设置修饰符
		modifiers := xproto.ModifierControl | xproto.ModifierShift

		// 绑定快捷键
		xproto.GrabKey(x.Conn(), false, rootWindow, modifiers, keycode, xproto.GrabModeAsync, xproto.GrabModeAsync)
	}

	// 事件循环
	for {
		ev, err := x.Conn().NextEvent()
		if err != nil {
			log.Fatal("Failed to get event:", err)
		}

		switch ev := ev.(type) {
		case xproto.KeyPressEvent:
			fmt.Println("Key pressed:", ev.Detail)
			// 执行通知命令
			cmd := exec.Command("notify-send", "Hello")
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			if err := cmd.Start(); err != nil {
				log.Fatal("Failed to start command:", err)
			}
		}
	}
}
