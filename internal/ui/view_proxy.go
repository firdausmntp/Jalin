package ui

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
)

// ViewProxy renders the proxy routing & tunneling studio dock:
// visual pipeline inspector, profile directory, and modal configuration.
func (a *App) ViewProxy(c *ui.Context) {
	ui.Column(c).MinHeight(pageHeight(c)).AlignSelf(ui.Stretch).Padding(pad(c)).Gap(16).Children(func() {
		// 1. Header Ribbon
		ui.Row(c).FillWidth().AlignItems(ui.Center).Justify(ui.SpaceBetween).Gap(16).Children(func() {
			ui.Column(c).Gap(2).Shrink(1).Children(func() {
				ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "PROXY ROUTING & TUNNELING").Font(fontDisplay).FontSize(h1Size(c)).FontWeight(700).LetterSpacing(0.8).TextColor(text)
					badgePill(c, "APP-LEVEL TUNNEL", accent, ui.RGBA(79, 70, 229, 0.12))
				})
				ui.Text(c, "Configure upstream HTTP and SOCKS5 proxy gateways. System network and other apps remain untouched.").
					FontSize(11.5).TextColor(textMute)
			})
			if ui.PrimaryButton(c, "+ New Profile").Clicked() {
				a.ShowAddModal = true
			}
		})

		// 2. Active Route Pipeline Inspector
		a.proxyRoute(c)

		// 3. Saved Proxy Profiles Directory
		a.proxyProfiles(c)

		// 4. Studio Modal Dialog (Inspired by Cavalry Dialog Card)
		a.proxyForm(c)
	})
}

// proxyRoute renders the visual routing pipeline diagram and last reachability probe.
func (a *App) proxyRoute(c *ui.Context) {
	active := a.ActiveProxy != nil && a.ActiveProxy.Active

	statusBadge := "DIRECT NIC"
	statusCol := good
	if active {
		statusBadge = "ROUTED: " + strings.ToUpper(a.ActiveProxy.Type)
		statusCol = accent
	}

	panelCard(c, "Routing Pipeline", statusBadge, statusCol, func() {
		ui.Column(c).FillWidth().Gap(14).Children(func() {
			// Visual Pipeline Nodes
			ui.Row(c).FillWidth().Padding(12, 16).Background(panelHi).Radius(6).Border(1, rule).AlignItems(ui.Center).Justify(ui.SpaceBetween).Children(func() {
				// Node 1: Host Application
				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.Box(c).Width(32).Height(24).Radius(4).Background(ink).Border(1, ruleHi).Center().Children(func() {
						ui.Text(c, "SRC").FontSize(9.5).Font(fontMono).FontWeight(700).TextColor(textFaint)
					})
					ui.Column(c).Gap(1).Children(func() {
						ui.Text(c, "Jalin Client").FontSize(11.5).FontWeight(600).TextColor(text)
						meta(c, "Local Process")
					})
				})

				// Connector 1
				ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "──────►").FontSize(10).Font(fontMono).TextColor(textFaint)
				})

				// Node 2: Proxy Gateway / Direct
				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					gatewayTag := "DIR"
					gatewayTitle := "Direct Connection"
					gatewaySub := "Standard Gateway"
					iconBg := ui.RGBA(16, 185, 129, 0.15)
					iconBorder := good
					tagCol := good

					if active {
						gatewayTag = "TUN"
						gatewayTitle = a.ActiveProxy.Name
						gatewaySub = fmt.Sprintf("%s://%s:%d", a.ActiveProxy.Type, a.ActiveProxy.Host, a.ActiveProxy.Port)
						iconBg = ui.RGBA(99, 102, 241, 0.2)
						iconBorder = accent
						tagCol = accent
					}

					ui.Box(c).Width(32).Height(24).Radius(4).Background(iconBg).Border(1, iconBorder).Center().Children(func() {
						ui.Text(c, gatewayTag).FontSize(9.5).Font(fontMono).FontWeight(700).TextColor(tagCol)
					})
					ui.Column(c).Gap(1).Children(func() {
						ui.Text(c, gatewayTitle).FontSize(11.5).FontWeight(600).TextColor(text)
						meta(c, gatewaySub)
					})
				})

				// Connector 2
				ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "──────►").FontSize(10).Font(fontMono).TextColor(textFaint)
				})

				// Node 3: Internet Edge Target
				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.Box(c).Width(32).Height(24).Radius(4).Background(ink).Border(1, ruleHi).Center().Children(func() {
						ui.Text(c, "DST").FontSize(9.5).Font(fontMono).FontWeight(700).TextColor(textFaint)
					})
					ui.Column(c).Gap(1).Children(func() {
						ui.Text(c, "Anycast CDN").FontSize(11.5).FontWeight(600).TextColor(text)
						meta(c, "Target Server")
					})
				})
			})

			// Route details and probe trigger
			ui.Row(c).FillWidth().Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
				ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
					caption(c, "PIPELINE STATUS")
					if active {
						ui.Text(c, routeDetail(a, active)).FontSize(11.5).Font(fontMono).TextColor(accent)
					} else {
						ui.Text(c, "All traffic exits directly through the primary default route").
							FontSize(11.5).Font(fontMono).TextColor(textMute)
					}
				})

				ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
					probeStatus(c, a)
					if active {
						if ui.Button(c, "Disconnect Gateway").Clicked() {
							a.ToggleProxy(a.ActiveProxy)
						}
					}
				})
			})
		})
	})
}

// routeDetail describes the address without ever printing credentials.
func routeDetail(a *App, active bool) string {
	if !active {
		return "Direct local interface"
	}
	p := a.ActiveProxy
	s := fmt.Sprintf("%s://%s:%d", strings.ToUpper(p.Type), p.Host, p.Port)
	if p.Username != "" {
		s += " · Auth: " + p.Username
	}
	return s
}

// probeStatus reports the last reachability check.
func probeStatus(c *ui.Context, a *App) {
	text_, col := "", textFaint
	switch {
	case a.IsTestingProxy:
		text_, col = "PROBING REACHABILITY…", signalAmber
	case a.ProxyTestError != "":
		text_, col = "PROBE FAILED: "+a.ProxyTestError, bad
	case a.ProxyTestLatency != "":
		text_, col = "PROBE REACHABLE: "+a.ProxyTestLatency, good
	}
	if text_ == "" {
		return
	}
	badgePill(c, text_, col, ui.RGBA(col.R, col.G, col.B, 0.15))
}

// proxyProfiles renders the saved profiles table with management buttons.
func (a *App) proxyProfiles(c *ui.Context) {
	if len(a.Profiles) == 0 {
		panelCard(c, "Saved Gateways", "0 Profiles", textFaint, func() {
			emptyState(c, "No Proxy Profiles Configured",
				"Add an HTTP or SOCKS5 proxy profile to route speed tests and reachability probes through encrypted or remote tunnels.",
				"Add New Profile", func() { a.ShowAddModal = true })
		})
		return
	}

	panelCard(c, "Saved Gateways", fmt.Sprintf("%d Profiles", len(a.Profiles)), accent, func() {
		// Table Header
		ui.Row(c).FillWidth().Padding(6, 12).Background(panelHeader).Radius(4).AlignItems(ui.Center).Children(func() {
			headCell(c, "PROTOCOL", 70, ui.Start)
			headCell(c, "PROFILE NAME & ENDPOINT", 0, ui.Start)
			headCell(c, "AUTHENTICATION", 130, ui.Start)
			headCell(c, "ACTIONS", 230, ui.End)
		})
		hairline(c, false)

		ui.Column(c).FillWidth().Dividers(1, rule).Children(func() {
			for _, item := range a.Profiles {
				p := item
				row := ui.Row(c).Key(p.ID).FillWidth().Padding(10, 12).AlignItems(ui.Center).Radius(4)
				if row.Hovered() {
					row.Background(panelHi)
				}
				row.Children(func() {
					// Protocol tag
					ui.Row(c).Width(70).Children(func() {
						col := accent
						if p.Type == "http" {
							col = signalAmber
						}
						badgePill(c, strings.ToUpper(p.Type), col, ui.RGBA(col.R, col.G, col.B, 0.15))
					})

					// Profile info
					ui.Column(c).Grow(1).Gap(2).Children(func() {
						ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
							ui.Text(c, p.Name).FontSize(12.5).FontWeight(600).TextColor(text)
							if p.Active {
								badgePill(c, "ACTIVE", good, ui.RGBA(16, 185, 129, 0.2))
							}
						})
						ui.Textf(c, "%s:%d", p.Host, p.Port).FontSize(11).Font(fontMono).TextColor(textFaint)
					})

					// Authentication Status
					ui.Row(c).Width(130).Children(func() {
						if p.Username != "" {
							meta(c, "User: "+p.Username)
						} else {
							meta(c, "None (Anonymous)")
						}
					})

					// Actions
					ui.Row(c).Width(230).Gap(8).Justify(ui.End).AlignItems(ui.Center).Children(func() {
						if ui.Button(c, "Test").Clicked() {
							a.TestProxyConnection(&p)
						}

						connectLabel := "Connect"
						if p.Active {
							connectLabel = "Disconnect"
						}
						if ui.Button(c, connectLabel).Clicked() {
							a.ToggleProxy(&p)
						}

						removeCell(c, 56, func() { a.DeleteProxy(p.ID) })
					})
				})
			}
		})
	})
}

// proxyForm renders the studio dialog modal inspired by Cavalry's dialog card.
func (a *App) proxyForm(c *ui.Context) {
	if !a.ShowAddModal {
		return
	}
	ui.Modal(c, &a.ShowAddModal, func() {
		// Cavalry-Style Floating Dialog Card
		ui.Column(c).Width(480).Padding(24).Gap(18).Background(ui.Hex("#191C23")).Radius(10).
			Border(1, ui.Hex("#394252")).Children(func() {
			// Dialog Header with Step Badge (Identical rhythm to Cavalry dialog)
			ui.Row(c).FillWidth().Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
				badgePill(c, "1/1 CONFIGURATION", accent, ui.RGBA(99, 102, 241, 0.25))
				ui.Text(c, "✕").FontSize(13).TextColor(textFaint)
			})

			ui.Column(c).Gap(4).Children(func() {
				ui.Text(c, "Register Proxy Gateway").Font(fontDisplay).FontSize(18).FontWeight(700).TextColor(text)
				ui.Text(c, "Configure upstream routing. All credentials are encrypted & stored in local SQLite only.").
					FontSize(11.5).TextColor(textMute)
			})

			hairline(c, false)

			// Form Fields
			field(c, "Profile Name", "Descriptive label for this proxy tunnel.", func() {
				ui.TextInput(c, &a.NewProxyName)
			})

			ui.Row(c).FillWidth().Gap(12).Children(func() {
				ui.Column(c).Width(130).Children(func() {
					field(c, "Protocol", "", func() {
						ui.Select(c, &a.NewProxyType, []string{"http", "socks5"})
					})
				})
				ui.Column(c).Grow(1).Children(func() {
					field(c, "Host / IP Address", "", func() {
						ui.TextInput(c, &a.NewProxyHost)
					})
				})
				ui.Column(c).Width(90).Children(func() {
					field(c, "Port", "", func() {
						ui.TextInput(c, &a.NewProxyPort)
					})
				})
			})

			ui.Row(c).FillWidth().Gap(12).Children(func() {
				ui.Column(c).Grow(1).Children(func() {
					field(c, "Username", "Optional credentials", func() {
						ui.TextInput(c, &a.NewProxyUser)
					})
				})
				ui.Column(c).Grow(1).Children(func() {
					field(c, "Password", "Never exported in plain views", func() {
						ui.TextInput(c, &a.NewProxyPass)
					})
				})
			})

			hairline(c, false)

			// Dialog Action Buttons (like Cavalry's Discard & Next buttons)
			ui.Row(c).FillWidth().Justify(ui.End).Gap(10).Children(func() {
				if ui.Button(c, "Discard").Clicked() {
					a.ShowAddModal = false
				}
				if ui.PrimaryButton(c, "Save Gateway").Clicked() {
					a.SaveNewProxy()
				}
			})
		})
	})
}
