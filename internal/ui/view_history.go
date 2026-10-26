package ui

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
)

// ViewHistory renders the benchmark audit log:
// aggregate performance summary, historical run table, and database management.
func (a *App) ViewHistory(c *ui.Context) {
	ui.Column(c).MinHeight(pageHeight(c)).AlignSelf(ui.Stretch).Padding(pad(c)).Gap(16).Children(func() {
		// 1. Header Toolbar
		ui.Row(c).FillWidth().AlignItems(ui.Center).Justify(ui.SpaceBetween).Gap(16).Children(func() {
			ui.Column(c).Gap(2).Shrink(1).Children(func() {
				ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "BENCHMARK HISTORY LOGS").Font(fontDisplay).FontSize(h1Size(c)).FontWeight(700).LetterSpacing(0.8).TextColor(text)
					badgePill(c, "SQLITE LOCAL", accent, ui.RGBA(79, 70, 229, 0.12))
				})
				ui.Text(c, "Persistent chronological log of completed benchmark runs stored in local SQLite.").
					FontSize(11.5).TextColor(textMute)
			})
			if len(a.History) > 0 {
				if ui.Button(c, "Clear Database").Clicked() {
					a.ClearAllHistory()
				}
			}
		})

		if len(a.History) == 0 {
			panelCard(c, "Benchmark Records", "Empty", textFaint, func() {
				emptyState(c, "No Benchmark Runs Recorded",
					"Completed speed tests are automatically persisted here. Run your first benchmark from the speed workspace.",
					"Open Benchmark Workspace", func() { a.SelectedNav = "test" })
			})
			return
		}

		// 2. Aggregate Metrics Summary
		a.historyAverages(c)

		// 3. Historical Data Table
		a.historyTable(c)
	})
}

// historyAverages renders studio summary tiles of all past runs.
func (a *App) historyAverages(c *ui.Context) {
	var down, up, maxDown float64
	for _, h := range a.History {
		down += h.DownloadMbps
		up += h.UploadMbps
		if h.DownloadMbps > maxDown {
			maxDown = h.DownloadMbps
		}
	}
	n := float64(len(a.History))

	panelCard(c, "Aggregate Analytics", "Telemetry Summary", accent, func() {
		if compact(c) {
			ui.Column(c).FillWidth().Gap(10).Children(func() {
				ui.Row(c).FillWidth().Gap(10).Children(func() {
					ui.Column(c).Grow(1).Children(func() {
						metricCard(c, "Runs Recorded", fmt.Sprintf("%d", len(a.History)), "runs", "SQLite records", text)
					})
					ui.Column(c).Grow(1).Children(func() {
						metricCard(c, "Average Downstream", mbps(down/n), "Mbps", "Mean velocity", signal)
					})
				})
				ui.Row(c).FillWidth().Gap(10).Children(func() {
					ui.Column(c).Grow(1).Children(func() {
						metricCard(c, "Average Upstream", mbps(up/n), "Mbps", "Mean upload", data)
					})
					ui.Column(c).Grow(1).Children(func() {
						metricCard(c, "All-Time Best Peak", mbps(maxDown), "Mbps", "Peak saturated", signalAmber)
					})
				})
			})
		} else {
			ui.Row(c).FillWidth().Gap(12).Children(func() {
				ui.Column(c).Grow(1).Children(func() {
					metricCard(c, "Runs Recorded", fmt.Sprintf("%d", len(a.History)), "runs", "Persistent SQLite records", text)
				})
				ui.Column(c).Grow(1).Children(func() {
					metricCard(c, "Average Downstream", mbps(down/n), "Mbps", "Mean download velocity", signal)
				})
				ui.Column(c).Grow(1).Children(func() {
					metricCard(c, "Average Upstream", mbps(up/n), "Mbps", "Mean upload velocity", data)
				})
				ui.Column(c).Grow(1).Children(func() {
					metricCard(c, "All-Time Best Peak", mbps(maxDown), "Mbps", "Highest saturated run", signalAmber)
				})
			})
		}
	})
}

// historyTable renders the studio data grid with fixed tabular widths.
func (a *App) historyTable(c *ui.Context) {
	colWhen := float32(126)
	if compact(c) {
		colWhen = 114
	}
	const (
		colDown   float32 = 88
		colUp     float32 = 88
		colLatJit float32 = 98
		colRoute  float32 = 72
		colAction float32 = 50
	)

	panelCard(c, "Audit Records", fmt.Sprintf("%d Entries", len(a.History)), good, func() {
		// Header row with explicit Gap to prevent collision
		ui.Row(c).FillWidth().Padding(6, 12).Background(panelHi).Radius(4).AlignItems(ui.Center).Gap(10).Children(func() {
			headCell(c, "TIMESTAMP", colWhen, ui.Start)
			headCell(c, "DOWNLOAD", colDown, ui.End)
			headCell(c, "UPLOAD", colUp, ui.End)
			headCell(c, "LATENCY / JITTER", colLatJit, ui.End)
			headCell(c, "ROUTING", colRoute, ui.Center)
			headCell(c, "TARGET ENDPOINT", 0, ui.Start)
			headCell(c, "", colAction, ui.Center)
		})
		hairline(c, false)

		ui.Column(c).FillWidth().Dividers(1, rule).Children(func() {
			for _, item := range a.History {
				r := item
				row := ui.Row(c).Key(r.ID).FillWidth().Padding(9, 12).AlignItems(ui.Center).Radius(4).Gap(10)
				if row.Hovered() {
					row.Background(panelHi)
				}
				row.Children(func() {
					ui.Text(c, r.CreatedAt.Format(time.DateOnly+" 15:04:05")).Width(colWhen).
						FontSize(10.5).Font(fontMono).TextColor(textMute)
					reading(c, mbps(r.DownloadMbps), 13.5, signal).Width(colDown).TextAlign(ui.End)
					reading(c, mbps(r.UploadMbps), 13.5, data).Width(colUp).TextAlign(ui.End)
					ui.Textf(c, "%.1f / %.1f", r.LatencyMs, r.JitterMs).Width(colLatJit).
						FontSize(10.5).Font(fontMono).TextColor(textMute).TextAlign(ui.End)
					ui.Row(c).Width(colRoute).Justify(ui.Center).Children(func() {
						if r.ViaProxy {
							badgePill(c, "PROXY", accent, ui.RGBA(79, 70, 229, 0.15))
						} else {
							badgePill(c, "DIRECT", good, ui.RGBA(5, 150, 105, 0.12))
						}
					})
					ui.Text(c, trimHost(r.Server, 36)).Grow(1).FontSize(11).TextColor(text).SingleLine()
					removeCell(c, colAction, func() { a.DeleteHistoryItem(r.ID) })
				})
			}
		})
	})
}

// headCell is a column title in high-density studio format.
func headCell(c *ui.Context, s string, w float32, align ui.Align) {
	t := ui.Text(c, s).FontSize(9.5).Font(fontMono).FontWeight(700).TextColor(textFaint).LetterSpacing(0.5).TextAlign(align)
	if w > 0 {
		t.Width(w)
	} else {
		t.Grow(1)
	}
}

// removeCell renders the delete action button with hover glow.
func removeCell(c *ui.Context, w float32, onClick func()) {
	col := textFaint
	box := ui.Box(c).Width(w).Padding(3, 8).Radius(3).Center()
	if box.Hovered() {
		box.Background(ui.RGBA(244, 63, 94, 0.15))
		col = bad
	}
	if box.Clicked() {
		onClick()
	}
	box.Children(func() {
		ui.Text(c, "Delete").FontSize(10.5).FontWeight(600).TextColor(col)
	})
}
