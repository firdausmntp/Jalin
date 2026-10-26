package ui

import (
	"github.com/egoist/mygo/ui"
)

// ViewSettings renders the studio preferences and runtime build telemetry.
func (a *App) ViewSettings(c *ui.Context) {
	ui.Column(c).MinHeight(pageHeight(c)).AlignSelf(ui.Stretch).Padding(pad(c)).Gap(16).Children(func() {
		// 1. Header Ribbon
		ui.Row(c).FillWidth().AlignItems(ui.Center).Justify(ui.SpaceBetween).Gap(16).Children(func() {
			ui.Column(c).Gap(2).Shrink(1).Children(func() {
				ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "STUDIO PREFERENCES").Font(fontDisplay).FontSize(h1Size(c)).FontWeight(700).LetterSpacing(0.8).TextColor(text)
					badgePill(c, "CONFIG", accent, ui.RGBA(79, 70, 229, 0.12))
				})
				ui.Text(c, "Engine transfer parameters and runtime hardware environment specifications.").
					FontSize(11.5).TextColor(textMute)
			})
		})

		// 2. Responsive Workstation Layout
		if wide(c) {
			ui.Row(c).FillWidth().Gap(16).AlignItems(ui.Start).Children(func() {
				ui.Column(c).Grow(1).Children(func() { a.settingParameters(c) })
				ui.Column(c).Grow(1).Children(func() { a.settingBuild(c) })
			})
		} else {
			a.settingParameters(c)
			a.settingBuild(c)
		}
	})
}

// settingParameters renders the engine transfer settings card.
func (a *App) settingParameters(c *ui.Context) {
	panelCard(c, "Transfer Engine Parameters", "Applied Next Run", accent, func() {
		ui.Column(c).FillWidth().Gap(14).Children(func() {
			ui.Row(c).FillWidth().Gap(16).Children(func() {
				ui.Column(c).Grow(1).Children(func() {
					field(c, "Parallel Streams", "Concurrent TCP/HTTP workers.", func() {
						ui.TextInput(c, &a.StreamsInput)
					})
				})
				ui.Column(c).Grow(1).Children(func() {
					field(c, "Phase Window (Sec)", "Sampling time after warm-up.", func() {
						ui.TextInput(c, &a.DurationInput)
					})
				})
			})

			hairline(c, false)

			ui.Row(c).FillWidth().AlignItems(ui.Center).Justify(ui.SpaceBetween).Gap(16).Children(func() {
				msg, col := a.SettingsMessage, good
				if msg == "" {
					msg, col = "Settings stored in local SQLite database.", textFaint
				}
				ui.Text(c, msg).FontSize(11).Font(fontMono).TextColor(col).Shrink(1)

				if ui.PrimaryButton(c, "Save Parameters").Clicked() {
					a.SaveSettings()
				}
			})
		})
	})
}

// settingBuild renders the runtime environment specifications table.
func (a *App) settingBuild(c *ui.Context) {
	panelCard(c, "Runtime System Specifications", "Hardware Host", textFaint, func() {
		ui.Column(c).FillWidth().Dividers(1, rule).Children(func() {
			buildRow(c, "Application Version", "1.0.0 (Production Release)")
			buildRow(c, "Render Engine", "GPU Accelerated (github.com/egoist/mygo)")
			buildRow(c, "Kernel Telemetry", "Windows iphlpapi.dll (GetIfTable2)")
			buildRow(c, "Database Driver", "SQLite modernc.org/sqlite (WAL mode)")
			buildRow(c, "CGO Status", "Zero CGO · Pure Go Toolchain")
			buildRow(c, "Proxy Capabilities", "HTTP/1.1 & SOCKS5 with Auth")
		})
	})
}

// buildRow is one key/value line in the build record.
func buildRow(c *ui.Context, key, value string) {
	ui.Row(c).FillWidth().Padding(8, 0).Justify(ui.SpaceBetween).Gap(16).AlignItems(ui.Center).Children(func() {
		ui.Text(c, key).FontSize(11.5).TextColor(textMute)
		ui.Text(c, value).FontSize(11).Font(fontMono).TextColor(text)
	})
}
