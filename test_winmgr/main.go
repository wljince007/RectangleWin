package main

import (
	"log"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xwindow"
)

func main() {
	// Connect to the X server
	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	// 获取屏幕尺寸
	geom, err := xwindow.New(X, X.RootWin()).Geometry()
	if err != nil {
		log.Fatal("获取屏幕尺寸失败:", err)
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

	keystr := "Mod4-G" //Super+g

	keystr = "Control-mod1-1"
	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, e xevent.KeyPressEvent) {
			// 获取活动窗口
			win, err := ewmh.ActiveWindowGet(X)
			if err != nil {
				log.Fatal("获取活动窗口失败:", err)
			}
			// 计算右半区尺寸并调整窗口
			width := geom.Width() / 2
			height := geom.Height() / 2
			x := geom.Width() / 2
			y := 0

			err = ewmh.MoveresizeWindow(X, win, x, y, width, height)
			if err != nil {
				log.Fatal("调整窗口失败:", err)
			}
		},
	).Connect(X, X.RootWin(), keystr, true) // true 表示自动抓取按键

	keystr = "Control-mod1-2"
	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, e xevent.KeyPressEvent) {
			// 获取活动窗口
			win, err := ewmh.ActiveWindowGet(X)
			if err != nil {
				log.Fatal("获取活动窗口失败:", err)
			}
			// 计算右半区尺寸并调整窗口
			width := geom.Width() / 2
			height := geom.Height() / 2
			x := 0
			y := 0

			err = ewmh.MoveresizeWindow(X, win, x, y, width, height)
			if err != nil {
				log.Fatal("调整窗口失败:", err)
			}
		},
	).Connect(X, X.RootWin(), keystr, true) // true 表示自动抓取按键

	// Finally, start the main event loop. This will route any appropriate
	// KeyPressEvents to your callback function.
	log.Println("Program initialized. Start pressing keys!")
	xevent.Main(X)
}
