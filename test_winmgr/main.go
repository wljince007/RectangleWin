package main

import (
	"log"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xwindow"
)

func moveToNextMonitor(X *xgbutil.XUtil) {
	// 获取所有显示器信息
	screens, err := xinerama.PhysicalHeads(X)
	if err != nil {
		log.Fatal("获取显示器失败:", err)
	}

	// 获取活动窗口
	win, err := ewmh.ActiveWindowGet(X)
	if err != nil {
		log.Fatal("获取活动窗口失败:", err)
	}

	// 获取窗口当前位置
	winGeom, err := xwindow.New(X, win).Geometry()
	if err != nil {
		log.Fatal(err)
	}

	// 查找当前显示器索引
	currentIdx := 0
	for i, screen := range screens {
		if winGeom.X() >= screen.X() && winGeom.X() < screen.X()+screen.Width() {
			currentIdx = i
			break
		}
	}

	// 计算下一个显示器索引
	nextIdx := (currentIdx + 1) % len(screens)
	nextScreen := screens[nextIdx]

	// 移动窗口到下一个显示器
	err = ewmh.MoveresizeWindow(X, win,
		nextScreen.X()+(winGeom.X()%screens[currentIdx].Width()),
		nextScreen.Y()+(winGeom.Y()%screens[currentIdx].Height()),
		winGeom.Width(),
		winGeom.Height())
	if err != nil {
		log.Fatal("移动窗口失败:", err)
	}
}

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

	// 绑定快捷键
	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, e xevent.KeyPressEvent) {
			moveToNextMonitor(X)
		},
	).Connect(X, X.RootWin(), "Control-mod1-n", true)

	// Finally, start the main event loop. This will route any appropriate
	// KeyPressEvents to your callback function.
	log.Println("Program initialized. Start pressing keys!")
	xevent.Main(X)
}
