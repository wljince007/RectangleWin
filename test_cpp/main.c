/********************************************************
 * https://www.x.org/releases/current/doc/libX11/libX11/libX11.html
 *******************************************************/
#include <stdlib.h>
#include <stdio.h>
#include <time.h>

#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <X11/Xatom.h>

// 日志记录
#define LOG(format, ...)                                \
{                                                       \
    printf(format, ##__VA_ARGS__);                      \
    printf("\n");                                       \
    char buffer[26];                                    \
    time_t timeval = time(NULL);                        \
    struct tm* tm_info = localtime(&timeval);           \
    strftime(buffer, 26, "%Y-%m-%d %H:%M:%S", tm_info); \
    FILE *log = fopen("log.txt","a");                   \
    fprintf(log,"%s : "format"\n",buffer,##__VA_ARGS__);\
    fclose(log);                                        \
}

// 按键处理
void handleShortKey(XKeyEvent xkey)
{
    LOG("%s","Key has been press and release");
    system("notify-send Hello");
}

int main()
{
    // 初始化
    Display * display = XOpenDisplay(NULL);
    if(display == NULL)
    {
        LOG("Unable to open X display\n");
    }

    // 绑定按键
    XSync(display, 0);
    for (int screen = 0; screen < ScreenCount (display); screen++)
    {
        Window grab_window = RootWindow (display, screen);
        LOG("Current screen %d, current window %ld\n", screen, grab_window);

        KeyCode customKeyCode = XKeysymToKeycode(display, XK_Z);
        uint modifiers = ControlMask | ShiftMask;

        XGrabKey(display, customKeyCode, modifiers,
                grab_window, False, GrabModeAsync, GrabModeAsync);
    }
    XSync(display, 0);

	// 事件循环
    while (True)
    {
        XEvent event;
        Bool matched = False;
        while (XPending(display) || (!matched)) 
        {
            XNextEvent(display, &event);
            switch (event.type) 
            {
                case KeyPress:
                    LOG("KeyPress");
                    break;
                case KeyRelease:
                    LOG("KeyRelease");
                    handleShortKey(event.xkey);
                    matched = True;
                    break;
                default:
                    break;
            }
        }
    }
    XCloseDisplay(display);
    return 0;
}
