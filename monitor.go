//go:build linux

package main

import (
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/ahmetb/RectangleLinux/w32"
	"github.com/cihub/seelog"
)

// EnumMonitors 遍历所有物理显示器，回调每个显示器的 w32.HMONITOR（用output id代替）
func EnumMonitors(f func(d w32.HMONITOR) bool) bool {
	X, err := xgb.NewConn()
	if err != nil {
		seelog.Errorf("无法连接X server:", err)
		return false
	}
	defer X.Close()

	err = randr.Init(X)
	if err != nil {
		seelog.Errorf("无法初始化RandR:", err)
		return false
	}
	root := getRootWindow(X)
	res, err := randr.GetScreenResources(X, root).Reply()
	if err != nil {
		seelog.Errorf("无法获取显示器资源:", err)
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
			seelog.Errorf("无法连接X server:", err)
			return false
		}
		defer X.Close()
		err = randr.Init(X)
		if err != nil {
			seelog.Errorf("Init: %v", err)
			return true
		}
		output := randr.Output(uintptr(d))
		info, err := randr.GetOutputInfo(X, output, 0).Reply()
		if err != nil {
			seelog.Errorf("获取输出信息失败: %v", err)
			return true
		}
		name := string(info.Name)
		crtc := info.Crtc
		crtcInfo, err := randr.GetCrtcInfo(X, crtc, 0).Reply()
		if err != nil {
			seelog.Errorf("获取CRTC信息失败: %v", err)
			return true
		}
		seelog.Debugf("> monitor: %s (id=0x%x)", name, uint32(output))
		seelog.Debugf("    pos: (%d, %d)  size: %dx%d",
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
