package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/egoist/mygo/ui"
	"jalin/internal/speedtest"
)

// ViewSpeedTest renders the primary workstation screen:
// a central oscilloscope viewport stage paired with a telemetry inspector dock.
func (a *App) ViewSpeedTest(c *ui.Context) {
	ui.Column(c).MinHeight(pageHeight(c)).AlignSelf(ui.Stretch).Padding(pad(c)).Gap(16).Children(func() {
		a.speedHeader(c)

		if wide(c) {
			ui.Row(c).FillWidth().Gap(16).AlignItems(ui.Start).Children(func() {
				ui.Column(c).Grow(3).Children(func() {
					a.speedViewport(c)
				})
				ui.Column(c).Width(320).Shrink(0).Children(func() {
					a.speedInspector(c)
				})
			})
		} else {
			a.speedViewport(c)
			a.speedInspector(c)
		}
	})
}

// speedHeader renders the top control ribbon with server target selection.
func (a *App) speedHeader(c *ui.Context) {
	names, ids := speedTargets()
	chosen := speedtest.GetServerOrDefault(a.SelectedServerID).Name

	ui.Row(c).FillWidth().AlignItems(ui.Center).Justify(ui.SpaceBetween).Gap(16).Children(func() {
		ui.Column(c).Gap(2).Shrink(1).Children(func() {
			ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "BENCHMARK WORKSPACE").Font(fontDisplay).FontSize(h1Size(c)).FontWeight(700).LetterSpacing(0.8).TextColor(text)
				badgePill(c, "MULTI-STREAM", accent, ui.RGBA(79, 70, 229, 0.12))
			})
			ui.Text(c, "Concurrent multi-threaded TCP/HTTP bandwidth telemetry & latency analysis.").
				FontSize(11.5).TextColor(textMute)
		})

		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			caption(c, "TARGET")
			ui.Box(c).Width(210).Children(func() {
				if ui.Select(c, &chosen, names).Changed() {
					a.SelectedServerID = ids[chosen]
				}
			})
		})
	})
}

// speedViewport renders the central oscilloscope monitor stage.
func (a *App) speedViewport(c *ui.Context) {
	value := a.DownloadMbps
	phase, col := "IDLE", textFaint
	if a.IsTesting {
		value = a.CurrentMbps
		switch a.CurrentPhase {
		case speedtest.PhasePing:
			phase, col = "PING TEST", signalAmber
		case speedtest.PhaseDownload:
			phase, col = "DOWNSTREAM", signal
		case speedtest.PhaseUpload:
			phase, col = "UPSTREAM", data
		}
	} else if a.DownloadMbps > 0 || a.UploadMbps > 0 {
		phase, col = "COMPLETE", good
	}

	ceiling := niceCeiling(value)
	frac := float32(value / ceiling)
	peak := math.Max(a.DownloadMbps, a.UploadMbps)

	panelCard(c, "Viewport Monitor", phase, col, func() {
		// 1. Digital Telemetry HUD & Live Rate
		ui.Row(c).FillWidth().Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
			ui.Column(c).Gap(2).Children(func() {
				caption(c, "INSTANTANEOUS THROUGHPUT")
				valueRow(c, mbps(value), "Mbps", 58, col)
			})
			ui.Column(c).AlignItems(ui.End).Gap(4).Children(func() {
				meta(c, fmt.Sprintf("%s STREAMS · %s SEC / PHASE", valueOr(a.StreamsInput, "4"), valueOr(a.DurationInput, "8")))
				if a.LastResult != nil {
					badgePill(c, "PREV PEAK "+mbps(peak)+" Mbps", textFaint, ui.RGBA(96, 105, 120, 0.15))
				}
			})
		})

		// 2. High-Tech Oscilloscope Canvas
		a.oscilloscopeCanvas(c, ceiling, col)

		// 3. Calibrated Velocity Ruler
		ui.Column(c).FillWidth().Gap(2).Padding(4, 0, 0, 0).Children(func() {
			scale(c, frac, float32(peak/ceiling))
			scaleLabels(c, []string{
				tickText(0),
				tickText(ceiling / 4),
				tickText(ceiling / 2),
				tickText(ceiling * 3 / 4),
				tickText(ceiling),
			})
		})

		// 4. Action Bar & Execution Progress
		ui.Column(c).FillWidth().Gap(8).Padding(8, 0, 0, 0).Children(func() {
			if a.IsTesting {
				// Sleek Progress Bar
				ui.Box(c).Height(4).FillWidth().Draw(func(p *ui.Painter, r ui.Rect) {
					p.Fill(r, rule, 2)
					w := r.W * float32(a.Progress)
					if w > 0 {
						p.Fill(ui.Rect{X: r.X, Y: r.Y, W: w, H: r.H}, col, 2)
					}
				})
			}

			ui.Row(c).FillWidth().AlignItems(ui.Center).Justify(ui.SpaceBetween).Gap(16).Children(func() {
				ui.Row(c).Gap(8).AlignItems(ui.Center).Shrink(1).Children(func() {
					dotCol := good
					if a.IsTesting {
						dotCol = col
					}
					ui.Box(c).Width(6).Height(6).Radius(3).Background(dotCol)
					ui.Text(c, valueOr(a.StatusText, "Ready to benchmark")).FontSize(11.5).Font(fontMono).TextColor(textMute)
				})

				if a.IsTesting {
					if ui.Button(c, "Cancel Run").Clicked() {
						a.CancelSpeedTest()
					}
				} else if ui.PrimaryButton(c, "Start Benchmark").Clicked() {
					a.StartSpeedTest()
				}
			})
		})
	})
}

// oscilloscopeCanvas draws a studio-grade waveform oscilloscope with grid lines.
func (a *App) oscilloscopeCanvas(c *ui.Context, ceiling float64, col ui.Color) {
	hist := a.ChartHistory

	ui.Box(c).FillWidth().Height(130).Background(panel).Radius(6).Border(1, rule).Draw(func(p *ui.Painter, r ui.Rect) {
		// Draw Horizontal Oscilloscope Gridlines
		for i := 1; i <= 3; i++ {
			gy := r.Y + r.H*float32(i)/4.0
			p.Line(r.X, gy, r.X+r.W, gy, 1, ui.RGBA(215, 223, 233, 0.8))
		}
		// Draw Vertical Time Marks
		for i := 1; i <= 5; i++ {
			gx := r.X + r.W*float32(i)/6.0
			p.Line(gx, r.Y, gx, r.Y+r.H, 1, ui.RGBA(215, 223, 233, 0.6))
		}

		if len(hist) < 2 || ceiling <= 0.001 {
			return
		}

		// Calculate waveform coordinates
		pts := make([][2]float32, 0, len(hist))
		for i, v := range hist {
			x := r.X + r.W*float32(i)/float32(len(hist)-1)
			normalized := float32(v / ceiling)
			if normalized > 1 {
				normalized = 1
			}
			y := r.Y + r.H - normalized*(r.H-16) - 8
			pts = append(pts, [2]float32{x, y})
		}

		// Render waveform polyline
		p.StrokePolyline(pts, 2, col)

		// Render glowing head cursor on the latest sample
		last := pts[len(pts)-1]
		p.Fill(ui.Rect{X: last[0] - 4, Y: last[1] - 4, W: 8, H: 8}, col, 4)
		p.Fill(ui.Rect{X: last[0] - 1.5, Y: last[1] - 1.5, W: 3, H: 3}, panel, 1.5)
	})
}

// speedInspector renders the right-hand studio inspector dock with telemetry channels.
func (a *App) speedInspector(c *ui.Context) {
	panelCard(c, "Telemetry Channels", "Inspector", textFaint, func() {
		ui.Column(c).FillWidth().Gap(10).Children(func() {
			// 1. Latency & Jitter Tile
			latSubtitle := "Excellent"
			if a.LatencyMs > 60 {
				latSubtitle = "Moderate"
			} else if a.LatencyMs > 120 {
				latSubtitle = "High"
			}
			metricCard(c, "Latency / Jitter", fmt.Sprintf("%.1f", a.LatencyMs), "ms", fmt.Sprintf("±%.1f ms · %s", a.JitterMs, latSubtitle), signalAmber)

			// 2. Download Throughput Tile
			metricCard(c, "Download Throughput", mbps(a.DownloadMbps), "Mbps", "Post-warmup saturated", signal)

			// 3. Upload Throughput Tile
			metricCard(c, "Upload Throughput", mbps(a.UploadMbps), "Mbps", "Multi-stream POST", data)

			// 4. Route & Protocol Inspector Tile
			ui.Column(c).FillWidth().Padding(12, 14).Background(panelHi).Radius(6).Border(1, rule).Gap(6).Children(func() {
				caption(c, "Active Pipeline")
				ui.Row(c).Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
					if a.ActiveProxy != nil && a.ActiveProxy.Active {
						ui.Text(c, a.ActiveProxy.Name).FontSize(12).FontWeight(600).TextColor(accent)
						badgePill(c, strings.ToUpper(a.ActiveProxy.Type), accent, ui.RGBA(99, 102, 241, 0.2))
					} else {
						ui.Text(c, "Local Direct Interface").FontSize(12).FontWeight(600).TextColor(good)
						badgePill(c, "DIRECT", good, ui.RGBA(16, 185, 129, 0.18))
					}
				})
				meta(c, "Zero interception · Pure Go stack")
			})
		})
	})
}

// speedTargets maps the select's labels back to server ids.
func speedTargets() ([]string, map[string]string) {
	ids := make(map[string]string, len(speedtest.DefaultServers))
	names := make([]string, 0, len(speedtest.DefaultServers))
	for _, s := range speedtest.DefaultServers {
		names = append(names, s.Name)
		ids[s.Name] = s.ID
	}
	return names, ids
}

// niceCeiling rounds up to the next readable scale end.
func niceCeiling(v float64) float64 {
	for _, m := range []float64{10, 25, 50, 100, 250, 500, 1000, 2500, 5000} {
		if v <= m {
			return m
		}
	}
	return 10000
}

// tickText keeps a scale label short.
func tickText(v float64) string {
	if v == math.Trunc(v) {
		return fmt.Sprintf("%.0f", v)
	}
	return fmt.Sprintf("%.1f", v)
}

// valueOr falls back when a stored setting is empty.
func valueOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
