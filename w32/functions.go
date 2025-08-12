//go:build linux

package w32

import (
	"log"
	"sync"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/cihub/seelog"
)

var (
	xConn     *xgbutil.XUtil
	xConnOnce sync.Once
)

func GetXConn() *xgbutil.XUtil {
	xConnOnce.Do(func() {
		xu, err := xgbutil.NewConn()
		if err == nil {
			xConn = xu
		}
	})
	return xConn
}

type HWND uint32
type HMONITOR uintptr

// GetDeviceCaps index constants
const (
	DRIVERVERSION   = 0
	TECHNOLOGY      = 2
	HORZSIZE        = 4
	VERTSIZE        = 6
	HORZRES         = 8
	VERTRES         = 10
	LOGPIXELSX      = 88
	LOGPIXELSY      = 90
	BITSPIXEL       = 12
	PLANES          = 14
	NUMBRUSHES      = 16
	NUMPENS         = 18
	NUMFONTS        = 22
	NUMCOLORS       = 24
	NUMMARKERS      = 20
	ASPECTX         = 40
	ASPECTY         = 42
	ASPECTXY        = 44
	PDEVICESIZE     = 26
	CLIPCAPS        = 36
	SIZEPALETTE     = 104
	NUMRESERVED     = 106
	COLORRES        = 108
	PHYSICALWIDTH   = 110
	PHYSICALHEIGHT  = 111
	PHYSICALOFFSETX = 112
	PHYSICALOFFSETY = 113
	SCALINGFACTORX  = 114
	SCALINGFACTORY  = 115
	VREFRESH        = 116
	DESKTOPHORZRES  = 118
	DESKTOPVERTRES  = 117
	BLTALIGNMENT    = 119
	SHADEBLENDCAPS  = 120
	COLORMGMTCAPS   = 121
	RASTERCAPS      = 38
	CURVECAPS       = 28
	LINECAPS        = 30
	POLYGONALCAPS   = 32
	TEXTCAPS        = 34
)

type RECT struct {
	Left, Top, Right, Bottom int32
}

func (r *RECT) Width() int32 {
	if r == nil {
		return 0
	}
	return r.Right - r.Left
}

func (r *RECT) Height() int32 {
	if r == nil {
		return 0
	}
	return r.Bottom - r.Top
}

func MessageBox(hwnd HWND, text, caption string, flags uint32) int {
	// Linux stub: no message box
	return 0
}
func GetActiveWindow() HWND {
	return GetForegroundWindow()
}

// X11原生获取当前活动窗口
func GetForegroundWindow() HWND {
	xu := GetXConn()
	if xu == nil {
		return 0
	}

	win, err := ewmh.ActiveWindowGet(xu)
	if err != nil {
		log.Fatal("获取活动窗口失败:", err)
	}
	return (HWND)(win)

	// prop, err := xprop.GetProperty(xu, xu.RootWin(), "_NET_ACTIVE_WINDOW")
	// if err != nil || prop == nil || len(prop.Value) < 4 {
	// 	return 0
	// }
	// // _NET_ACTIVE_WINDOW 是一个32位window id（cardinal），用小端字节序
	// winId := uint32(prop.Value[0]) | uint32(prop.Value[1])<<8 | uint32(prop.Value[2])<<16 | uint32(prop.Value[3])<<24
	// return HWND(winId)
}

func GetWindowText(hwnd HWND) string {
	xu := GetXConn()
	if xu == nil {
		return ""
	}
	// _NET_WM_NAME 是UTF-8编码的窗口标题
	prop, err := xprop.GetProperty(xu, xproto.Window(hwnd), "_NET_WM_NAME")
	if err != nil || prop == nil {
		return ""
	}
	// prop.Value 是 []byte
	return string(prop.Value)
}

// X11原生获取窗口几何信息
func GetWindowRect(hwnd HWND) *RECT {
	xu := GetXConn()
	if xu == nil {
		return &RECT{}
	}

	win := xwindow.New(xu, xproto.Window(hwnd))
	geom, err := win.DecorGeometry()
	if err != nil {
		return &RECT{}
	}
	seelog.Debugf("window: 0x%x geom:%#v", hwnd, geom)

	absX, absY := int(geom.X()), int(geom.Y())
	return &RECT{
		Left:   int32(absX),
		Top:    int32(absY),
		Right:  int32(absX) + int32(geom.Width()),
		Bottom: int32(absY) + int32(geom.Height()),
	}
}

func MonitorFromWindow(hwnd HWND, flag uint32) uintptr {
	return 0
}

func GetDC(hwnd HWND) uintptr {
	return 0
}

func GetDeviceCaps(hdc uintptr, index int) int {
	return 96 // 默认 DPI
}

func ReleaseDC(hwnd HWND, hdc uintptr) bool {
	return true
}
func Test() bool {
	xu := GetXConn()
	if xu == nil {
		return false
	}

	desktopNames, err := ewmh.DesktopNamesGet(xu)
	if err != nil {
		seelog.Errorf("DesktopNamesGet err:%v", err)
		return false
	}
	for i, desktopName := range desktopNames {
		seelog.Debugf("i:%v, desktopName:%v", i, desktopName)
	}
	// i:0, desktopName:Workspace 1
	// i:1, desktopName:Workspace 2
	// i:2, desktopName:Workspace 3
	// i:3, desktopName:Workspace 4

	desktopGeometry, err := ewmh.DesktopGeometryGet(xu)
	if err != nil {
		seelog.Errorf("DesktopGeometryGet err:%v", err)
		return false
	}
	seelog.Debugf("desktopGeometry:%v", desktopGeometry)
	// ans: desktopGeometry:&{3840 2160}

	desktopLayout, err := ewmh.DesktopLayoutGet(xu)
	if err != nil {
		seelog.Errorf("DesktopLayoutGet err:%v", err)
		return false
	}
	seelog.Debugf("desktopLayout:%v", desktopLayout)
	// ans: desktopLayout:&{0 0 1 0}

	desktopViewport, err := ewmh.DesktopViewportGet(xu)
	if err != nil {
		seelog.Errorf("DesktopViewportGet err:%v", err)
		return false
	}
	seelog.Debugf("desktopViewport:%v", desktopViewport)
	// ans: desktopViewport:[{0 0}]

	// 桌面数量
	numberOfDesktops, err := ewmh.NumberOfDesktopsGet(xu)
	if err != nil {
		seelog.Errorf("NumberOfDesktopsGet err:%v", err)
		return false
	}
	seelog.Debugf("numberOfDesktops:%v", numberOfDesktops)
	// ans: numberOfDesktops:4

	//调用失败
	// desktopIdxArr, err := ewmh.VisibleDesktopsGet(xu)
	// if err != nil {
	// 	seelog.Errorf("VisibleDesktopsGet err:%v", err)
	// 	return false
	// }
	// for i, desktopIdx := range desktopIdxArr {
	// 	seelog.Debugf("i:%v, desktopIdx:%v", i, desktopIdx)
	// }

	//得到活动窗口是第一个桌面
	win, err := ewmh.ActiveWindowGet(xu)
	if err != nil {
		seelog.Errorf("ActiveWindowGet err:%v", err)
		return false
	}
	desktopIdx, err := ewmh.WmDesktopGet(xu, win)
	if err != nil {
		seelog.Errorf("WmDesktopGet err:%v", err)
		return false
	}
	seelog.Debugf("desktopIdx:%v", desktopIdx)
	// ans: desktopIdx:1

	WorkareaArr, err := ewmh.WorkareaGet(xu)
	if err != nil {
		seelog.Errorf("WorkareaGet err:%v", err)
		return false
	}
	for i, workarea := range WorkareaArr {
		seelog.Debugf("i:%v, workarea:%v", i, workarea)
	}
	// 得到4个工作区 ans:
	// 2025/08/12 09:34:10 i:0, workarea:{0 27 3840 2133}
	// 2025/08/12 09:34:10 i:1, workarea:{0 27 3840 2133}
	// 2025/08/12 09:34:10 i:2, workarea:{0 27 3840 2133}
	// 2025/08/12 09:34:10 i:3, workarea:{0 27 3840 2133}
	return true
}
func GetMonitorInfo(mon uintptr, info *MONITORINFO) bool {
	if info != nil {
		*info = MONITORINFO{}
	}
	xu := GetXConn()
	if xu == nil {
		return false
	}

	globalMenuHight := 0 //全局任务栏高度

	// 得到4个工作区,工作中有任务栏高度
	WorkareaArr, err := ewmh.WorkareaGet(xu)
	if err != nil {
		seelog.Errorf("WorkareaGet err:%v", err)
		return false
	}
	if len(WorkareaArr) > 0 {
		globalMenuHight = WorkareaArr[0].Y
	}

	// 获取屏幕尺寸
	geom := xwindow.RootGeometry(xu)
	geom.YSet(globalMenuHight)
	seelog.Debugf("RootGeometry geom:%v", geom)
	info.RcWork = RECT{int32(geom.X()), int32(geom.Y()), int32(geom.Width()), int32(geom.Height())}
	return true
}

func DwmGetWindowAttributeEXTENDED_FRAME_BOUNDS(hwnd HWND) (bool, RECT) {
	return true, RECT{}
}

func ShowWindow(hwnd HWND, cmdShow int) bool {
	return true
}

func GetWindowLong(hwnd HWND, index int) int32 {
	//todo 待实现
	return 0
}

func SetWindowPos(hwnd HWND, hwndInsertAfter HWND, x, y, cx, cy int, uFlags uint32) bool {
	// // 仅处理移动和缩放
	// exec.Command("xdotool", "windowmove", strconv.FormatUint(uint64(hwnd), 10), strconv.Itoa(x), strconv.Itoa(y)).Run()
	// exec.Command("xdotool", "windowsize", strconv.FormatUint(uint64(hwnd), 10), strconv.Itoa(cx), strconv.Itoa(cy)).Run()

	xu := GetXConn()
	if xu == nil {
		return false
	}
	err := ewmh.MoveresizeWindow(xu, (xproto.Window)(hwnd), x, y, cx, cy)
	if err != nil {
		log.Fatal("调整窗口失败:", err)
		return false
	}

	return true
}

func GetLastError() int {
	return 0
}

// 常量定义
const (
	MB_ICONWARNING           = 0
	MB_OK                    = 0
	SW_SHOWNORMAL            = 0
	SW_MAXIMIZE              = 0
	MONITOR_DEFAULTTONEAREST = 0
	GWL_EXSTYLE              = 0
	WS_EX_TOPMOST            = 0
	HWND_NOTOPMOST           = 0
	HWND_TOPMOST             = 0
	SWP_NOMOVE               = 0
	SWP_NOSIZE               = 0
	SWP_NOZORDER             = 0
	SWP_NOACTIVATE           = 0
)

type MONITORINFO struct {
	CbSize    uint32
	RcMonitor RECT
	RcWork    RECT
	DwFlags   uint32
}
