package ui

import (
	"math"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
	"jalin/internal/netstat"
)

// ViewMonitor renders the live network traffic analyzer dock:
// dual channel meters, continuous spectrum chart, daily data usage (Internet Speed Meter),
// and lifetime byte volume counters.
func (a *App) ViewMonitor(c *ui.Context) {
	stats := a.LiveStats

	ui.Column(c).MinHeight(pageHeight(c)).AlignSelf(ui.Stretch).Padding(pad(c)).Gap(16).Children(func() {
		// 1. Monitor Title & Kernel Source Ribbon
		ui.Row(c).FillWidth().AlignItems(ui.Center).Justify(ui.SpaceBetween).Gap(16).Children(func() {
			ui.Column(c).Gap(2).Shrink(1).Children(func() {
				ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "NETWORK SPECTRUM ANALYZER").Font(fontDisplay).FontSize(h1Size(c)).FontWeight(700).LetterSpacing(0.8).TextColor(text)
					badgePill(c, "LIVE KERNEL", signal, ui.RGBA(5, 150, 105, 0.12))
				})
				ui.Text(c, "Continuous interface throughput sampled every 500ms via Windows GetIfTable2 kernel API.").
					FontSize(11.5).TextColor(textMute)
			})
			meta(c, "SAMPLER 2.0 Hz · DUAL STREAM")
		})

		// 2. Dual Channel Metric Rack (Downstream RX & Upstream TX)
		if wide(c) {
			ui.Row(c).FillWidth().Gap(16).Children(func() {
				ui.Column(c).Grow(1).Children(func() {
					channelCard(c, "Downstream Interface", "RX CH-01", stats.RXMbps, stats.RXBytesPerSec, stats.PeakRXMbps, stats.HistoryRX, signal)
				})
				ui.Column(c).Grow(1).Children(func() {
					channelCard(c, "Upstream Interface", "TX CH-02", stats.TXMbps, stats.TXBytesPerSec, stats.PeakTXMbps, stats.HistoryTX, data)
				})
			})
		} else {
			channelCard(c, "Downstream Interface", "RX CH-01", stats.RXMbps, stats.RXBytesPerSec, stats.PeakRXMbps, stats.HistoryRX, signal)
			channelCard(c, "Upstream Interface", "TX CH-02", stats.TXMbps, stats.TXBytesPerSec, stats.PeakTXMbps, stats.HistoryTX, data)
		}

		// 3. Central Oscilloscope Spectrum Chart
		a.monitorChart(c, stats)

		// 4. Daily Data Usage Storage (Internet Speed Meter)
		a.monitorDailyUsage(c)

		// 5. Telemetry Totals Rack
		a.monitorTotals(c, stats)
	})
}

// channelCard renders one direction of traffic in a studio rack panel.
func channelCard(c *ui.Context, label, short string, rate, bps, peak float64, hist []float64, col ui.Color) {
	ceiling := niceCeiling(peak)
	panelCard(c, label, short, col, func() {
		ui.Column(c).FillWidth().Gap(12).Children(func() {
			ui.Row(c).FillWidth().Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
				ui.Column(c).Gap(2).Children(func() {
					caption(c, "CURRENT RATE")
					valueRow(c, mbps(rate), "Mbps", 38, col)
				})
				ui.Column(c).AlignItems(ui.End).Gap(4).Children(func() {
					badgePill(c, "PEAK "+mbps(peak)+" Mbps", col, ui.RGBA(col.R, col.G, col.B, 0.15))
					meta(c, netstat.FormatSpeed(bps))
				})
			})

			// Mini waveform sparkline
			spark(c, hist, col)

			// Load capacity indicator bar
			loadBar(c, float32(rate/ceiling), col)
		})
	})
}

// spark plots a channel's recent shape inside a card.
func spark(c *ui.Context, hist []float64, col ui.Color) {
	top := 0.0001
	for _, v := range hist {
		if v > top {
			top = v
		}
	}
	ui.Box(c).FillWidth().Height(46).Background(ink).Radius(4).Border(1, rule).Draw(func(p *ui.Painter, r ui.Rect) {
		if len(hist) < 2 {
			return
		}
		pts := make([][2]float32, 0, len(hist))
		for i, v := range hist {
			x := r.X + r.W*float32(i)/float32(len(hist)-1)
			y := r.Y + r.H - float32(v/top)*(r.H-8) - 4
			pts = append(pts, [2]float32{x, y})
		}
		p.StrokePolyline(pts, 1.5, col)
	})
}

// loadBar renders a precision progress load meter.
func loadBar(c *ui.Context, frac float32, col ui.Color) {
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	ui.Box(c).Height(3).FillWidth().Draw(func(p *ui.Painter, r ui.Rect) {
		p.Fill(r, rule, 1.5)
		if frac > 0 {
			p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W * frac, H: r.H}, col, 1.5)
		}
	})
}

// monitorChart plots both channels over the sampler's history window.
func (a *App) monitorChart(c *ui.Context, stats netstat.TrafficStats) {
	maxVal := 10.0
	for _, series := range [][]float64{stats.HistoryRX, stats.HistoryTX} {
		for _, v := range series {
			if v > maxVal {
				maxVal = v
			}
		}
	}
	y := chart.NewLinear(0, maxVal*1.2).Nice(4)
	span := math.Max(1, float64(len(stats.HistoryRX)-1))
	x := chart.NewLinear(0, span)

	panelCard(c, "Realtime Dual-Channel Waveform", "Oscilloscope", accent, func() {
		ui.Column(c).FillWidth().Gap(10).Children(func() {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Justify(ui.SpaceBetween).Children(func() {
				caption(c, "TIME HISTORY (LAST 30 INTERVALS)")
				ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
					badgePill(c, "■ RX DOWNSTREAM", signal, ui.RGBA(16, 185, 129, 0.15))
					badgePill(c, "■ TX UPSTREAM", data, ui.RGBA(56, 189, 248, 0.15))
				})
			})

			plot := func(p *ui.Painter, r ui.Rect, hist []float64, col ui.Color) {
				pts := make([][2]float32, 0, len(hist))
				for i, v := range hist {
					pts = append(pts, [2]float32{chart.X(r, x.At(float64(i))), chart.Y(r, y.At(v))})
				}
				if len(pts) > 1 {
					p.StrokePolyline(pts, 2, col)
				}
			}

			ui.Box(c).FillWidth().Height(160).Background(panel).Radius(6).Border(1, rule).Padding(10).Children(func() {
				chart.Plot(c, chart.Axes{Left: y.Ticks(4)}, func(area ui.Element) {
					area.Grow(1).Fill().Draw(func(p *ui.Painter, r ui.Rect) {
						for _, tick := range y.Ticks(4) {
							py := chart.Y(r, tick.At)
							p.Line(r.X, py, r.X+r.W, py, 1, rule)
						}
						plot(p, r, stats.HistoryRX, signal)
						plot(p, r, stats.HistoryTX, data)
					})
				}).Grow(1)
			})
		})
	})
}

// monitorDailyUsage renders the Internet Speed Meter daily data storage panel.
func (a *App) monitorDailyUsage(c *ui.Context) {
	today := time.Now().Format("2006-01-02")
	todayWiFi := netstat.FormatSpeed(float64(a.TodayUsage.WiFiBytes))
	todayOther := netstat.FormatSpeed(float64(a.TodayUsage.OtherBytes))
	todayTotal := netstat.FormatSpeed(float64(a.TodayUsage.TotalBytes))

	panelCard(c, "Daily Data Storage · Internet Speed Meter", today, signalAmber, func() {
		ui.Column(c).FillWidth().Gap(14).Children(func() {
			// Summary Cards: WiFi, Ethernet / Other, and Total
			ui.Row(c).FillWidth().Gap(14).Children(func() {
				ui.Column(c).Grow(1).Children(func() {
					metricCard(c, "Today WiFi Usage", todayWiFi, "", "Wireless 802.11 adapter", data)
				})
				ui.Column(c).Grow(1).Children(func() {
					metricCard(c, "Today LAN / Ethernet", todayOther, "", "Physical / virtual adapters", signal)
				})
				ui.Column(c).Grow(1).Children(func() {
					metricCard(c, "Today Total Transferred", todayTotal, "", "Total usage today", signalAmber)
				})
			})

			// Daily Usage Table (Date, WiFi, LAN / Other, Total, Distribution)
			if len(a.DailyUsages) > 0 {
				colDate := float32(136)
				colWiFi := float32(110)
				colOther := float32(115)
				colTotal := float32(110)
				if compact(c) {
					colDate = 110
					colWiFi = 90
					colOther = 95
					colTotal = 90
				}

				ui.Column(c).FillWidth().Background(panelHi).Radius(6).Border(1, rule).Clip().Children(func() {
					ui.Row(c).FillWidth().Padding(6, 12).Background(panelHeader).AlignItems(ui.Center).Gap(10).Children(func() {
						headCell(c, "DATE", colDate, ui.Start)
						headCell(c, "WIFI DATA", colWiFi, ui.End)
						headCell(c, "LAN / ETHERNET", colOther, ui.End)
						headCell(c, "TOTAL USAGE", colTotal, ui.End)
						headCell(c, "DISTRIBUTION (WIFI / LAN)", 0, ui.Start)
					})
					hairline(c, false)

					ui.Column(c).FillWidth().Dividers(1, rule).Children(func() {
						for _, item := range a.DailyUsages {
							u := item
							isToday := u.Date == today
							row := ui.Row(c).Key(u.Date).FillWidth().Padding(9, 12).AlignItems(ui.Center).Gap(10)
							if row.Hovered() {
								row.Background(panel)
							}
							row.Children(func() {
								ui.Row(c).Width(colDate).Gap(6).AlignItems(ui.Center).Children(func() {
									ui.Text(c, u.Date).FontSize(10.5).Font(fontMono).TextColor(text)
									if isToday && !compact(c) {
										badgePill(c, "TODAY", signalAmber, ui.RGBA(217, 119, 6, 0.12))
									}
								})
								ui.Text(c, netstat.FormatSpeed(float64(u.WiFiBytes))).Width(colWiFi).
									FontSize(10.5).Font(fontMono).TextColor(data).TextAlign(ui.End)
								ui.Text(c, netstat.FormatSpeed(float64(u.OtherBytes))).Width(colOther).
									FontSize(10.5).Font(fontMono).TextColor(signal).TextAlign(ui.End)
								ui.Text(c, netstat.FormatSpeed(float64(u.TotalBytes))).Width(colTotal).
									FontSize(11).Font(fontMono).FontWeight(600).TextColor(text).TextAlign(ui.End)

								// Distribution bar: WiFi (cyan) vs Other/LAN (emerald)
								ui.Box(c).Grow(1).Padding(0, 10).Children(func() {
									ui.Box(c).Height(4).FillWidth().Draw(func(p *ui.Painter, r ui.Rect) {
										p.Fill(r, rule, 2)
										if u.TotalBytes > 0 {
											wifiRatio := float32(float64(u.WiFiBytes) / float64(u.TotalBytes))
											if wifiRatio > 0 {
												p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W * wifiRatio, H: r.H}, data, 2)
											}
											if 1.0-wifiRatio > 0 {
												p.Fill(ui.Rect{X: r.X + r.W*wifiRatio, Y: r.Y, W: r.W * (1.0 - wifiRatio), H: r.H}, signal, 2)
											}
										}
									})
								})
							})
						}
					})
				})
			}
		})
	})
}

// monitorTotals renders cumulative byte metrics since system boot.
func (a *App) monitorTotals(c *ui.Context, stats netstat.TrafficStats) {
	rxValue, rxUnit := splitRate(netstat.FormatSpeed(float64(stats.TotalRXBytes)))
	txValue, txUnit := splitRate(netstat.FormatSpeed(float64(stats.TotalTXBytes)))
	rxUnit = strings.TrimSuffix(rxUnit, "/s")
	txUnit = strings.TrimSuffix(txUnit, "/s")

	panelCard(c, "Hardware Interface Telemetry", "Since Boot", textFaint, func() {
		ui.Row(c).FillWidth().Gap(14).Children(func() {
			ui.Column(c).Grow(1).Children(func() {
				metricCard(c, "Total Received", rxValue, rxUnit, "InOctets kernel counter", signal)
			})
			ui.Column(c).Grow(1).Children(func() {
				metricCard(c, "Total Sent", txValue, txUnit, "OutOctets kernel counter", data)
			})
			ui.Column(c).Grow(1).Children(func() {
				metricCard(c, "Combined Peak", mbps(stats.PeakRXMbps+stats.PeakTXMbps), "Mbps", "Peak aggregate bandwidth", signalAmber)
			})
		})
	})
}

// splitRate separates "1.23 GB/s" into number and unit.
func splitRate(s string) (string, string) {
	value, unit, ok := strings.Cut(s, " ")
	if !ok {
		return s, ""
	}
	return value, unit
}
