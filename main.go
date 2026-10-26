package main

import (
	_ "embed"
	"log"
	"runtime"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/egoist/mygo"
	mygo_ui "github.com/egoist/mygo/ui"
	"jalin/internal/store"
	jalin_ui "jalin/internal/ui"
)

//go:embed assets/icon.png
var appIcon []byte

func main() {
	// Initialize local SQLite storage
	dbPath := store.DefaultDBPath()
	db, err := store.OpenDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database at %s: %v", dbPath, err)
	}
	defer db.Close()

	// Initialize UI state coordinator
	app := jalin_ui.NewApp(db)

	mygo.App.WhenReady(func() {
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:           "Jalin v1.0.0 - Network Intelligence & Speed Benchmark",
			Width:           1040,
			Height:          740,
			MinWidth:        760,
			MinHeight:       520,
			BackgroundColor: "#F8FAFC",
			StateKey:        "jalin-main-window",
			Content:         mygo_ui.View(app.MainView),
		})

		applyLightTitleBar(win.NativeHandle())

		if len(appIcon) > 0 {
			_ = win.SetIcon(appIcon)
		}

		// Floating mini speed meter window docked at bottom-right (above Windows taskbar)
		var miniWin *mygo.Window
		disp := mygo.Screen.PrimaryDisplay()
		wa := disp.WorkArea
		if wa.Width > 0 && wa.Height > 0 {
			miniW := 218
			miniH := 32
			miniX := wa.X + wa.Width - miniW - 12
			miniY := wa.Y + wa.Height - miniH - 6

			// Load saved custom coordinates if previously moved by user
			savedX, _ := strconv.Atoi(db.GetSetting("mini_x", "0"))
			savedY, _ := strconv.Atoi(db.GetSetting("mini_y", "0"))
			if savedX > 0 && savedY > 0 {
				miniX = savedX
				miniY = savedY
			}
			app.UpdateMiniPosition(miniX, miniY)

			miniWin = mygo.NewWindow(mygo.WindowOptions{
				Title:           "Jalin Mini Meter",
				Width:           miniW,
				Height:          miniH,
				X:               miniX,
				Y:               miniY,
				Frameless:       true,
				AlwaysOnTop:     true,
				Panel:           true,
				SkipTaskbar:     true,
				DisableResize:   true,
				DisableClose:    false,
				Hidden:          true,
				BackgroundColor: "#FFFFFF",
				Content:         mygo_ui.View(app.MiniMeterView),
			})

			if len(appIcon) > 0 {
				_ = miniWin.SetIcon(appIcon)
			}

			// Keep window position persistent when dragged by user
			miniWin.OnMove(func() {
				x, y := miniWin.Position()
				if x > 0 && y > 0 {
					app.UpdateMiniPosition(x, y)
				}
			})

			// Ensure window stays permanently topmost when focus changes or moves
			miniWin.OnBlur(func() {
				ensureTopMost(miniWin.NativeHandle())
			})
			miniWin.OnShow(func() {
				ensureTopMost(miniWin.NativeHandle())
			})

			app.SetMiniWindow(miniWin)
		}

		// System Tray Notification Area Setup
		speedItem := &mygo.MenuItem{
			Label:    "Live: ↓ 0 B/s   ↑ 0 B/s",
			Disabled: true,
		}
		usageItem := &mygo.MenuItem{
			Label:    "Today: 0 B",
			Disabled: true,
		}

		trayMenu := mygo.NewMenu([]*mygo.MenuItem{
			speedItem,
			usageItem,
			mygo.Separator(),
			{
				Label: "Show Jalin",
				Click: func(item *mygo.MenuItem, w *mygo.Window) {
					app.RestoreFromTray()
				},
			},
			{
				Label: "Hide to Background",
				Click: func(item *mygo.MenuItem, w *mygo.Window) {
					app.HideToTray()
				},
			},
			mygo.Separator(),
			{
				Label: "Exit Jalin",
				Click: func(item *mygo.MenuItem, w *mygo.Window) {
					app.Exit()
				},
			},
		})

		tray, err := mygo.NewTray(mygo.TrayOptions{
			ID:      "jalin-workstation-tray",
			Icon:    appIcon,
			ToolTip: "Jalin - Internet Speed Meter",
			Menu:    trayMenu,
		})
		if err == nil {
			app.SetTray(tray, speedItem, usageItem)
			tray.OnClick(func() {
				if win.IsVisible() && !win.IsMinimized() {
					win.Focus()
				} else {
					app.RestoreFromTray()
				}
			})
		}

		// When main window is closed via [X], terminate the application cleanly
		win.OnClose(func(e *mygo.CloseEvent) {
			app.Exit()
		})

		app.SetWindow(win)
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatalf("Application runtime error: %v", err)
	}
}

// applyLightTitleBar sets Windows 10/11 DWM title bar theme to match the light UI.
func applyLightTitleBar(hwnd uintptr) {
	if runtime.GOOS != "windows" || hwnd == 0 {
		return
	}
	dwm := syscall.NewLazyDLL("dwmapi.dll")
	proc := dwm.NewProc("DwmSetWindowAttribute")
	// DWMWA_USE_IMMERSIVE_DARK_MODE = 20 (0 = Light title bar, 1 = Dark title bar)
	var isDark uint32 = 0
	_, _, _ = proc.Call(hwnd, 20, uintptr(unsafe.Pointer(&isDark)), 4)
}

// ensureTopMost forces the floating statusbar window to remain permanently in the topmost Z-order.
func ensureTopMost(hwnd uintptr) {
	if runtime.GOOS != "windows" || hwnd == 0 {
		return
	}
	user32 := syscall.NewLazyDLL("user32.dll")
	procSetWindowPos := user32.NewProc("SetWindowPos")
	// HWND_TOPMOST = -1, SWP_NOSIZE(0x0001) | SWP_NOMOVE(0x0002) | SWP_NOACTIVATE(0x0010) = 0x0013
	_, _, _ = procSetWindowPos.Call(hwnd, ^uintptr(0), 0, 0, 0, 0, 0x0013)
}
