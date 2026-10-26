package ui

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"jalin/internal/netstat"
)

// navItem is one destination in the workstation faceplate bar.
var navItems = []struct{ id, label, shortLabel string }{
	{"test", "Speed Benchmark", "Benchmark"},
	{"monitor", "Live Monitor", "Monitor"},
	{"history", "History Logs", "History"},
	{"proxy", "Proxy Router", "Proxy"},
	{"settings", "Studio Settings", "Settings"},
}

// MainView draws the complete studio workstation shell: faceplate ribbon,
// active dock view with responsive layout, and bottom status console.
func (a *App) MainView(c *ui.Context) {
	c.SetTheme(theme())

	ui.Column(c).Fill().Background(ink).Children(func() {
		a.faceplate(c)
		hairline(c, false)

		ui.Scroll(c).Grow(1).FillWidth().Children(func() {
			switch a.SelectedNav {
			case "monitor":
				a.ViewMonitor(c)
			case "history":
				a.ViewHistory(c)
			case "proxy":
				a.ViewProxy(c)
			case "settings":
				a.ViewSettings(c)
			default:
				a.ViewSpeedTest(c)
			}
		})

		hairline(c, false)
		a.statusStrip(c)
	})
}

// faceplate draws the studio header bar with brand mark, dock tab selector,
// and live telemetry chips with responsive sizing.
func (a *App) faceplate(c *ui.Context) {
	stats := a.LiveStats
	brandGap := float32(12)
	telemetryGap := float32(10)
	if compact(c) {
		brandGap = 8
		telemetryGap = 6
	}

	ui.Row(c).Height(navHeight).FillWidth().Padding(0, pad(c)).Background(panelHeader).AlignItems(ui.Center).Justify(ui.SpaceBetween).Gap(8).Children(func() {
		// 1. Brand Identity: JALIN (Makna Jaringan)
		ui.Row(c).AlignItems(ui.Center).Gap(brandGap).Children(func() {
			ui.Text(c, "JALIN").Font(fontDisplay).FontSize(15).FontWeight(750).LetterSpacing(1.6).TextColor(text)
			if !compact(c) {
				badgePill(c, "JARINGAN", accent, ui.RGBA(79, 70, 229, 0.08))
			}
		})

		// 2. Workstation Navigation Tabs (Responsive label length & padding)
		useShort := winW(c) < 1020
		tabPad := float32(11)
		if compact(c) {
			tabPad = 7
		}
		ui.Row(c).Height(32).Padding(2, 3).Background(panelHi).Radius(6).Border(1, rule).AlignItems(ui.Center).Gap(2).Children(func() {
			for _, item := range navItems {
				it := item
				tabText := it.label
				if useShort {
					tabText = it.shortLabel
				}
				isActive := a.SelectedNav == it.id
				a.navTab(c, tabText, isActive, tabPad, func() {
					a.SelectedNav = it.id
					if it.id == "history" || it.id == "proxy" || it.id == "monitor" {
						a.ReloadData()
					}
				})
			}
		})

		// 3. Quick Telemetry, Internet Speed Meter, & Route Chips
		ui.Row(c).AlignItems(ui.Center).Gap(telemetryGap).Children(func() {
			if wide(c) {
				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.Row(c).Width(86).Gap(4).AlignItems(ui.Center).Children(func() {
						ui.Text(c, "↓").FontSize(10.5).FontWeight(700).TextColor(signal)
						ui.Text(c, netstat.FormatSpeed(stats.RXBytesPerSec)).
							FontSize(10.5).Font(fontMono).FontFeatures("tnum").TextColor(textMute)
					})
					ui.Row(c).Width(86).Gap(4).AlignItems(ui.Center).Children(func() {
						ui.Text(c, "↑").FontSize(10.5).FontWeight(700).TextColor(data)
						ui.Text(c, netstat.FormatSpeed(stats.TXBytesPerSec)).
							FontSize(10.5).Font(fontMono).FontFeatures("tnum").TextColor(textMute)
					})
				})
			}

			// Today's Data Usage Chip (Internet Speed Meter)
			todayTotal := a.TodayUsage.TotalBytes
			todayStr := "TODAY " + netstat.FormatSpeed(float64(todayTotal))
			if compact(c) {
				todayStr = netstat.FormatSpeed(float64(todayTotal))
			}
			badgePill(c, todayStr, signalAmber, ui.RGBA(217, 119, 6, 0.10))

			// Route pill (hidden in compact mode to ensure comfortable space)
			if !compact(c) {
				routeText := "DIRECT"
				routeCol := good
				routeBg := ui.RGBA(5, 150, 105, 0.10)
				if a.ActiveProxy != nil && a.ActiveProxy.Active {
					routeText = "PROXY: " + trimHost(a.ActiveProxy.Name, 8)
					routeCol = accent
					routeBg = ui.RGBA(79, 70, 229, 0.12)
				}
				badgePill(c, routeText, routeCol, routeBg)
			}

			// Background / Mini Bar Toggle button
			trayBtn := ui.Row(c).Height(24).Padding(0, 8).Background(panelHi).Radius(4).Border(1, rule).AlignItems(ui.Center).Gap(4)
			if trayBtn.Hovered() {
				trayBtn.Background(ui.Hex("#E2E8F0"))
			}
			if trayBtn.Clicked() {
				a.HideToTray()
			}
			trayBtn.Children(func() {
				ui.Text(c, "▾").FontSize(9).TextColor(textFaint)
				trayLabel := "BACKGROUND"
				if compact(c) {
					trayLabel = "MINI"
				}
				ui.Text(c, trayLabel).FontSize(9.5).Font(fontMono).FontWeight(700).TextColor(textMute)
			})
		})
	})
}

// navTab renders a high-density workstation tab button with steady alignment.
func (a *App) navTab(c *ui.Context, label string, active bool, hPad float32, onClick func()) {
	bg := ui.Transparent
	border := ui.Transparent
	col := textMute
	dotCol := ui.Transparent
	if active {
		bg = panel
		border = rule
		col = text
		dotCol = accent
	}

	tab := ui.Row(c).Height(26).Padding(0, hPad).Background(bg).Radius(4).Border(1, border).AlignItems(ui.Center).Gap(5)
	if !active && tab.Hovered() {
		tab.Background(panelHi)
		col = text
	}
	if tab.Clicked() {
		onClick()
	}

	fontSize := float32(11.5)
	if compact(c) {
		fontSize = 11
	}
	tab.Children(func() {
		ui.Box(c).Width(4).Height(4).Radius(2).Background(dotCol)
		ui.Text(c, label).FontSize(fontSize).FontWeight(550).TextColor(col)
	})
}

// statusStrip renders the bottom speed meter console and telemetry bar.
func (a *App) statusStrip(c *ui.Context) {
	stats := a.LiveStats
	ui.Row(c).Height(32).FillWidth().Padding(0, pad(c)).Background(panelHeader).AlignItems(ui.Center).Justify(ui.SpaceBetween).Gap(10).Children(func() {
		// Left: Operational State with Status Pulse
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			dotCol := good
			if a.IsTesting {
				dotCol = signalAmber
			}
			ui.Box(c).Width(6).Height(6).Radius(3).Background(dotCol)
			msg := a.StatusText
			if msg == "" {
				msg = "Engine Idle · Ready"
			}
			ui.Text(c, msg).FontSize(10.5).Font(fontMono).TextColor(text)
		})

		// Center: Live Internet Speed Meter Display (RX & TX Dual-Channel Rack)
		ui.Row(c).Padding(2, 8).Background(panelHi).Radius(4).Border(1, rule).Gap(8).AlignItems(ui.Center).Children(func() {
			// Download RX
			ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "↓").FontSize(11).FontWeight(700).TextColor(signal)
				ui.Text(c, "RX:").FontSize(9.5).Font(fontMono).FontWeight(700).TextColor(textFaint)
				ui.Text(c, netstat.FormatSpeed(stats.RXBytesPerSec)).
					FontSize(10.5).Font(fontMono).FontFeatures("tnum").TextColor(signal)
			})

			// Divider
			ui.Box(c).Width(1).Height(12).Background(rule)

			// Upload TX
			ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "↑").FontSize(11).FontWeight(700).TextColor(data)
				ui.Text(c, "TX:").FontSize(9.5).Font(fontMono).FontWeight(700).TextColor(textFaint)
				ui.Text(c, netstat.FormatSpeed(stats.TXBytesPerSec)).
					FontSize(10.5).Font(fontMono).FontFeatures("tnum").TextColor(data)
			})
		})

		// Right: Daily Speed Meter Total & Hide to Tray Action
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			meta(c, fmt.Sprintf("TODAY %s", netstat.FormatSpeed(float64(a.TodayUsage.TotalBytes))))
			if !compact(c) {
				meta(c, fmt.Sprintf("RX TOTAL %s", netstat.FormatSpeed(float64(stats.TotalRXBytes))))
			}

			// "Hide to Tray" Micro Button
			bgBtn := ui.Row(c).Height(22).Padding(0, 8).Background(panelHi).Radius(4).Border(1, rule).AlignItems(ui.Center).Gap(4)
			if bgBtn.Hovered() {
				bgBtn.Background(ui.Hex("#E2E8F0"))
			}
			if bgBtn.Clicked() {
				a.HideToTray()
			}
			bgBtn.Children(func() {
				ui.Text(c, "▾").FontSize(9).TextColor(textFaint)
				ui.Text(c, "BACKGROUND").FontSize(9.5).Font(fontMono).FontWeight(700).TextColor(textMute)
			})
		})
	})
}
