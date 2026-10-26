package ui

import (
	"github.com/egoist/mygo/ui"
	"jalin/internal/netstat"
)

// MiniMeterView renders the draggable floating taskbar speed meter widget.
// Features a drag handle to move anywhere, live dual-channel RX/TX telemetry,
// a restore window button, and an application exit button.
func (a *App) MiniMeterView(c *ui.Context) {
	c.SetTheme(theme())
	stats := a.LiveStats

	meter := ui.Row(c).Fill().Padding(0, 8).Background(panel).
		Radius(6).Border(1, accent).AlignItems(ui.Center).Justify(ui.SpaceBetween).Gap(6)

	meter.Children(func() {
		// 1. Drag Grip Handle & Brand Label (Draggable)
		ui.Row(c).AlignItems(ui.Center).Gap(4).DragWindow().Children(func() {
			ui.Text(c, "⠿").FontSize(11).TextColor(textFaint)
			ui.Text(c, "JALIN").Font(fontDisplay).FontSize(10).FontWeight(700).LetterSpacing(0.8).TextColor(text)
		})

		// 2. Dual-Channel Live Throughput Rack (RX & TX) (Also Draggable)
		ui.Row(c).AlignItems(ui.Center).Gap(5).DragWindow().Children(func() {
			// Download RX
			ui.Row(c).Gap(2).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "↓").FontSize(10).FontWeight(700).TextColor(signal)
				ui.Text(c, netstat.FormatSpeed(stats.RXBytesPerSec)).
					FontSize(9.5).Font(fontMono).FontWeight(600).FontFeatures("tnum").TextColor(signal).SingleLine()
			})

			// Divider
			ui.Box(c).Width(1).Height(11).Background(rule)

			// Upload TX
			ui.Row(c).Gap(2).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "↑").FontSize(10).FontWeight(700).TextColor(data)
				ui.Text(c, netstat.FormatSpeed(stats.TXBytesPerSec)).
					FontSize(9.5).Font(fontMono).FontWeight(600).FontFeatures("tnum").TextColor(data).SingleLine()
			})
		})

		// 3. Action Controls: Restore Window & Exit Application
		ui.Row(c).AlignItems(ui.Center).Gap(3).Children(func() {
			// Restore to Main Window button
			restoreBtn := ui.Box(c).Width(18).Height(18).Radius(3).Background(panelHi).Border(1, rule).Center()
			if restoreBtn.Hovered() {
				restoreBtn.Background(ui.RGBA(79, 70, 229, 0.15))
			}
			if restoreBtn.Clicked() {
				a.RestoreFromTray()
			}
			restoreBtn.Children(func() {
				ui.Text(c, "⤢").FontSize(10).TextColor(accent)
			})

			// Exit Application button
			exitBtn := ui.Box(c).Width(18).Height(18).Radius(3).Background(panelHi).Border(1, rule).Center()
			if exitBtn.Hovered() {
				exitBtn.Background(ui.RGBA(220, 38, 38, 0.18))
			}
			if exitBtn.Clicked() {
				a.Exit()
			}
			exitBtn.Children(func() {
				ui.Text(c, "✕").FontSize(9).FontWeight(700).TextColor(bad)
			})
		})
	})
}
