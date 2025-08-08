//go:build linux

package main

import (
	"fmt"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/ahmetb/RectangleWin/w32"
)

// EnumMonitors 遍历所有物理显示器，回调每个显示器的 w32.HMONITOR（用output id代替）
func EnumMonitors(f func(d w32.HMONITOR) bool) bool {
	X, err := xgb.NewConn()
	if err != nil {
		fmt.Println("无法连接X server:", err)
		return false
	}
	defer X.Close()

	err = randr.Init(X)
	if err != nil {
		fmt.Println("无法初始化RandR:", err)
		return false
	}
	root := getRootWindow(X)
	res, err := randr.GetScreenResources(X, root).Reply()
	if err != nil {
		fmt.Println("无法获取显示器资源:", err)
		return false
	}
	for _, output := range res.Outputs {
		info, err := randr.GetOutputInfo(X, output, 0).Reply()
		if err != nil || info.Connection != randr.ConnectionConnected {
			continue
		}
		if !f(w32.HMONITOR(uintptr(output))) {
			break
		}
	}
	return true
}

// printMonitors 打印所有显示器信息
func printMonitors() {
	EnumMonitors(func(d w32.HMONITOR) bool {
		X, err := xgb.NewConn()
		if err != nil {
			fmt.Println("无法连接X server:", err)
			return false
		}
		defer X.Close()
		output := randr.Output(uintptr(d))
		info, err := randr.GetOutputInfo(X, output, 0).Reply()
		if err != nil {
			fmt.Printf("获取输出信息失败: %v\n", err)
			return true
		}
		name := string(info.Name)
		crtc := info.Crtc
		crtcInfo, err := randr.GetCrtcInfo(X, crtc, 0).Reply()
		if err != nil {
			fmt.Printf("获取CRTC信息失败: %v\n", err)
			return true
		}
		fmt.Printf("> monitor: %s (id=0x%x)\n", name, uint32(output))
		fmt.Printf("    pos: (%d, %d)  size: %dx%d\n",
			crtcInfo.X, crtcInfo.Y, crtcInfo.Width, crtcInfo.Height)
		// Linux下主显示器判断不统一，暂不区分
		// 工作区可通过 _NET_WORKAREA 获取，这里略
		return true
		return true
	})
}

func getRootWindow(X *xgb.Conn) xproto.Window {
	setup := xproto.Setup(X)
	return setup.DefaultScreen(X).Root
}
