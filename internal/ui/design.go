package ui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/egoist/mygo/ui"
)

// Design tokens. Soft clean white studio aesthetic: soothing off-white canvas,
// crisp pure white panels, and slate contrast without harsh glare.
var (
	ink         = ui.Hex("#F8FAFC") // Soothing neutral off-white canvas
	panel       = ui.Hex("#FFFFFF") // Crisp pure white dock panel surface
	panelHeader = ui.Hex("#FFFFFF") // Crisp pure white header ribbon
	panelHi     = ui.Hex("#F1F5F9") // Raised row, hover, clean contrast
	rule        = ui.Hex("#E2E8F0") // Modern subtle border divider
	ruleHi      = ui.Hex("#CBD5E1") // Slate focus divider
	text        = ui.Hex("#0F172A") // Deep crisp slate text
	textMute    = ui.Hex("#475569") // Muted secondary slate
	textFaint   = ui.Hex("#64748B") // Technical metadata
	accent      = ui.Hex("#4F46E5") // Vibrant indigo accent
	accentHover = ui.Hex("#4338CA")
	signal      = ui.Hex("#059669") // High-contrast emerald for download
	signalDim   = ui.Hex("#047857")
	data        = ui.Hex("#0284C7") // Cyber ocean blue for upload
	signalAmber = ui.Hex("#D97706") // Deep amber for latency & telemetry
	good        = ui.Hex("#059669")
	bad         = ui.Hex("#DC2626")
)

// Font roles: display numbers, UI labels, machine data.
const (
	fontDisplay = "Bahnschrift"
	fontUI      = "Segoe UI"
	fontMono    = "Cascadia Mono"
)

// theme is the clean soft-light workstation theme every stock widget draws with.
var theme = sync.OnceValue(func() *ui.Theme {
	t := ui.LightTheme()
	t.Background = ink
	t.Surface = panel
	t.SurfaceHover = panelHi
	t.SurfacePressed = ui.Hex("#E2E8F0")
	t.Border = rule
	t.Text = text
	t.TextMuted = textMute
	t.Accent = accent
	t.AccentHover = accentHover
	t.AccentPressed = ui.Hex("#3730A3")
	t.AccentText = ui.Hex("#FFFFFF")
	t.Danger = bad
	t.Warning = signalAmber
	t.Success = good
	t.Selection = ui.RGBA(79, 70, 229, 0.18)
	t.Focus = ui.RGBA(79, 70, 229, 0.40)
	t.Scrollbar = ui.RGBA(15, 23, 42, 0.15)
	t.Inverse = text
	t.InverseText = ink
	t.Radius = 6
	t.Spacing = 4
	t.FontSize = 13
	t.Font = fontUI
	return t
})

// Window dimension & responsive breakpoint helpers.
func winW(c *ui.Context) float32 {
	w, _ := c.Size()
	return w
}

func wide(c *ui.Context) bool {
	return winW(c) >= 1060
}

func medium(c *ui.Context) bool {
	return winW(c) >= 880
}

func compact(c *ui.Context) bool {
	return winW(c) < 880
}

// pad is the workspace gutter: 20 when roomy, 16 when medium, 12 when compact.
func pad(c *ui.Context) float32 {
	if wide(c) {
		return 20
	}
	if medium(c) {
		return 16
	}
	return 12
}

func h1Size(c *ui.Context) float32 {
	if compact(c) {
		return 15
	}
	return 17
}

// navHeight is the faceplate bar height.
const navHeight float32 = 46

// pageHeight is the viewport the shell leaves a view.
func pageHeight(c *ui.Context) float32 {
	_, h := c.Size()
	return h - navHeight - 34
}

// panelCard wraps content in a professional studio dock panel with a distinct header bar.
func panelCard(c *ui.Context, title string, badge string, badgeColor ui.Color, children func()) ui.Element {
	return ui.Column(c).FillWidth().Background(panel).Radius(8).Border(1, rule).Clip().Shrink(0).Children(func() {
		ui.Row(c).Height(34).FillWidth().Padding(0, 14).Background(panelHi).AlignItems(ui.Center).Justify(ui.SpaceBetween).Children(func() {
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Width(3).Height(12).Radius(1.5).Background(badgeColor)
				ui.Text(c, strings.ToUpper(title)).FontSize(10.5).FontWeight(700).LetterSpacing(0.8).TextColor(text)
			})
			if badge != "" {
				badgePill(c, badge, badgeColor, ui.RGBA(badgeColor.R, badgeColor.G, badgeColor.B, 0.12))
			}
		})
		hairline(c, false)
		ui.Column(c).FillWidth().Padding(14).Children(children)
	})
}

// badgePill renders a sleek micro badge with colored border & translucent background.
func badgePill(c *ui.Context, s string, col, bg ui.Color) ui.Element {
	return ui.Box(c).Padding(3, 8).Background(bg).Radius(4).Border(1, col).Children(func() {
		ui.Text(c, s).FontSize(10).Font(fontMono).FontWeight(600).TextColor(col).SingleLine()
	})
}

// caption is a field name: small, uppercase, technical rhythm.
func caption(c *ui.Context, s string) ui.Element {
	fontSize := float32(10)
	if compact(c) {
		fontSize = 9.5
	}
	return ui.Text(c, strings.ToUpper(s)).FontSize(fontSize).FontWeight(650).TextColor(textFaint).LetterSpacing(0.6)
}

// meta is machine data — timestamps, hosts, units of record.
func meta(c *ui.Context, s string) ui.Element {
	fontSize := float32(11)
	if compact(c) {
		fontSize = 10
	}
	return ui.Text(c, s).FontSize(fontSize).Font(fontMono).TextColor(textFaint)
}

// reading prints a measured number in the display face, tabular so digits
// never jitter while the value changes.
func reading(c *ui.Context, s string, size float32, col ui.Color) ui.Element {
	if compact(c) && size > 16 {
		size = size * 0.88
	}
	return ui.Text(c, s).Font(fontDisplay).FontFeatures("tnum").FontSize(size).FontWeight(600).TextColor(col)
}

// valueRow pairs a reading with its unit, sitting on the baseline.
func valueRow(c *ui.Context, v, unit string, size float32, col ui.Color) ui.Element {
	unitSize := size * 0.45
	if unitSize < 11 {
		unitSize = 11
	}
	return ui.Row(c).AlignItems(ui.Baseline).Gap(6).Children(func() {
		reading(c, v, size, col)
		if unit != "" {
			ui.Text(c, unit).FontSize(unitSize).FontWeight(600).TextColor(textMute)
		}
	})
}

// stat is one labelled measurement in a strip.
func stat(c *ui.Context, label, v, unit string, col ui.Color) ui.Element {
	return ui.Column(c).Gap(4).Children(func() {
		caption(c, label)
		valueRow(c, v, unit, 24, col)
	})
}

// metricCard renders a studio inspector attribute tile.
func metricCard(c *ui.Context, label, v, unit, subtitle string, col ui.Color) ui.Element {
	vSize := float32(22)
	if compact(c) {
		vSize = 19
	}
	return ui.Column(c).FillWidth().Padding(10, 12).Background(panelHi).Radius(6).Border(1, rule).Gap(4).Children(func() {
		ui.Row(c).FillWidth().Justify(ui.SpaceBetween).AlignItems(ui.Center).Children(func() {
			caption(c, label)
			if subtitle != "" {
				meta(c, subtitle)
			}
		})
		valueRow(c, v, unit, vSize, col)
	})
}

// scale draws the signature: a precision studio ruler with the value marked on it.
func scale(c *ui.Context, frac float32, peak float32) ui.Element {
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	return ui.Box(c).Height(38).FillWidth().Draw(func(p *ui.Painter, r ui.Rect) {
		base := r.Y + r.H - 8
		// Background Track
		p.Fill(ui.Rect{X: r.X, Y: base - 4, W: r.W, H: 4}, rule, 2)
		if frac > 0 {
			p.Fill(ui.Rect{X: r.X, Y: base - 4, W: r.W * frac, H: 4}, signal, 2)
		}
		// Ticks: 24 subdivisions for high-density instrument precision
		for i := 0; i <= 24; i++ {
			x := r.X + r.W*float32(i)/24
			h := float32(4)
			col := ruleHi
			if i%6 == 0 {
				h = 8
				col = textFaint
			} else if i%3 == 0 {
				h = 6
			}
			p.Line(x, base+2, x, base+2+h, 1, col)
		}
		// Peak marker: needle where the run's best sat
		if peak > 0 && peak < 1 {
			px := r.X + r.W*peak
			p.Line(px, base-12, px, base+10, 1.5, data)
		}
		// Value marker: glowing cyber caret
		if frac > 0 {
			x := r.X + r.W*frac
			p.Line(x, base-14, x, base+2, 2.5, signal)
			p.Fill(ui.Rect{X: x - 4, Y: base - 18, W: 8, H: 5}, signal, 1.5)
		}
	})
}

// scaleLabels writes the ruler's numbers under it.
func scaleLabels(c *ui.Context, marks []string) ui.Element {
	return ui.Row(c).Justify(ui.SpaceBetween).FillWidth().Padding(4, 0, 0, 0).Children(func() {
		for _, m := range marks {
			ui.Text(c, m).FontSize(9.5).Font(fontMono).TextColor(textFaint)
		}
	})
}

// hairline is the 1px rule that separates without drawing a box.
func hairline(c *ui.Context, vertical bool) ui.Element {
	if vertical {
		return ui.Box(c).Width(1).FillHeight().Background(rule)
	}
	return ui.Box(c).Height(1).FillWidth().Background(rule)
}

// mbps is the display form of a rate: no unit, two decimals past ten.
func mbps(v float64) string {
	if v >= 100 {
		return fmt.Sprintf("%.0f", v)
	}
	if v >= 10 {
		return fmt.Sprintf("%.1f", v)
	}
	return fmt.Sprintf("%.2f", v)
}

// trimHost keeps long proxy addresses inside a row.
func trimHost(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n-1]) + "…"
}

// field is a labelled control in the form views.
func field(c *ui.Context, label, help string, control func()) {
	ui.Column(c).Gap(6).Children(func() {
		ui.Text(c, label).FontSize(12).FontWeight(600).TextColor(text)
		if help != "" {
			ui.Text(c, help).FontSize(11).TextColor(textFaint)
		}
		control()
	})
}

// emptyState fills a plate when there is nothing to show yet, and says what
// to do about it.
func emptyState(c *ui.Context, title, help, action string, onAction func()) {
	ui.Column(c).Grow(1).Center().Gap(12).Padding(40).Children(func() {
		ui.Box(c).Width(40).Height(40).Radius(20).Background(panelHi).Border(1, ruleHi).Center().Children(func() {
			ui.Text(c, "◎").FontSize(18).TextColor(textMute)
		})
		ui.Text(c, title).FontSize(15).FontWeight(600).TextColor(text)
		ui.Text(c, help).FontSize(12).TextColor(textFaint).TextAlign(ui.Center)
		if action != "" {
			if ui.PrimaryButton(c, action).Clicked() {
				onAction()
			}
		}
	})
}

// tag is a small monospace label for state, styled like studio chip.
func tag(c *ui.Context, s string, col ui.Color) ui.Element {
	return ui.Box(c).Padding(1.5, 6).Background(ui.RGBA(col.R, col.G, col.B, 0.12)).Radius(3).Border(1, ui.RGBA(col.R, col.G, col.B, 0.35)).Children(func() {
		ui.Text(c, strings.ToUpper(s)).FontSize(9.5).Font(fontMono).FontWeight(600).TextColor(col)
	})
}
