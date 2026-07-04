package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Palette from georgenijo.com. Styles are built through a per-session
// lipgloss renderer, so truecolor hexes degrade automatically to the
// closest 256-color (or 16-color) equivalent on terminals that do not
// advertise truecolor support.
const (
	cBG     = "#141021"
	cBG2    = "#1c1631"
	cLine   = "#2c2347"
	cFG     = "#eae4f6"
	cDim    = "#9187ad"
	cFaint  = "#5c527d"
	cPink   = "#ff6ac1"
	cPurple = "#7d56f4"
	cLav    = "#b7a4f7"
	cGreen  = "#9ef29b"
	cCyan   = "#7ee0e6"
	cChipFG = "#1a1030"
)

type styles struct {
	r *lipgloss.Renderer

	fg, dim, faint    lipgloss.Style
	pink, pinkBold    lipgloss.Style
	lav, purple, cyan lipgloss.Style
	title, sub        lipgloss.Style
	kbd, hint, rule   lipgloss.Style
	pane, filterBox   lipgloss.Style
	badge, urlStyle   lipgloss.Style

	chip      string // pre-rendered "GEORGE NIJO" gradient chip
	banner    string // pre-rendered gradient ASCII wordmark
	bannerTag string // pre-rendered tagline under the wordmark
}

func newStyles(r *lipgloss.Renderer) *styles {
	c := func(hex string) lipgloss.Color { return lipgloss.Color(hex) }
	st := &styles{r: r}

	st.fg = r.NewStyle().Foreground(c(cFG))
	st.dim = r.NewStyle().Foreground(c(cDim))
	st.faint = r.NewStyle().Foreground(c(cFaint))
	st.pink = r.NewStyle().Foreground(c(cPink))
	st.pinkBold = st.pink.Bold(true)
	st.lav = r.NewStyle().Foreground(c(cLav))
	st.purple = r.NewStyle().Foreground(c(cPurple))
	st.cyan = r.NewStyle().Foreground(c(cCyan))

	st.title = st.pinkBold
	st.sub = st.faint
	st.kbd = r.NewStyle().Foreground(c(cLav))
	st.hint = st.faint
	st.rule = r.NewStyle().Foreground(c(cLine))
	st.pane = r.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(c(cLine)).
		Padding(0, 2)
	st.filterBox = r.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(c(cPurple)).
		Padding(0, 1)
	st.badge = st.cyan
	st.urlStyle = st.lav.Underline(true)

	st.chip = renderChip(r)
	st.banner = renderBanner(r)
	st.bannerTag = st.dim.Render(tagline)
	return st
}

// ── gradient helpers ─────────────────────────────────────────────

func hexToRGB(h string) (int, int, int) {
	h = strings.TrimPrefix(h, "#")
	if len(h) != 6 {
		return 255, 255, 255
	}
	r, _ := strconv.ParseInt(h[0:2], 16, 0)
	g, _ := strconv.ParseInt(h[2:4], 16, 0)
	b, _ := strconv.ParseInt(h[4:6], 16, 0)
	return int(r), int(g), int(b)
}

func lerpHex(a, b string, t float64) string {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	ar, ag, ab := hexToRGB(a)
	br, bg, bb := hexToRGB(b)
	mix := func(x, y int) int { return x + int(float64(y-x)*t+0.5) }
	return fmt.Sprintf("#%02x%02x%02x", mix(ar, br), mix(ag, bg), mix(ab, bb))
}

// gradAt mirrors the site's 100deg gradient: pink 5% → lavender 55% → purple 95%.
func gradAt(t float64) string {
	switch {
	case t <= 0.05:
		return cPink
	case t < 0.55:
		return lerpHex(cPink, cLav, (t-0.05)/0.50)
	case t < 0.95:
		return lerpHex(cLav, cPurple, (t-0.55)/0.40)
	default:
		return cPurple
	}
}

func renderBanner(r *lipgloss.Renderer) string {
	var lines []string
	for _, line := range asciiGeorge {
		runes := []rune(line)
		var b strings.Builder
		for i, ch := range runes {
			if ch == ' ' {
				b.WriteRune(' ')
				continue
			}
			t := float64(i) / float64(max(len(runes)-1, 1))
			b.WriteString(r.NewStyle().
				Foreground(lipgloss.Color(gradAt(t))).
				Bold(true).
				Render(string(ch)))
		}
		lines = append(lines, b.String())
	}
	return strings.Join(lines, "\n")
}

func renderChip(r *lipgloss.Renderer) string {
	text := " GEORGE NIJO "
	runes := []rune(text)
	var b strings.Builder
	for i, ch := range runes {
		t := float64(i) / float64(len(runes)-1)
		b.WriteString(r.NewStyle().
			Background(lipgloss.Color(lerpHex(cPink, cPurple, t))).
			Foreground(lipgloss.Color(cChipFG)).
			Bold(true).
			Render(string(ch)))
	}
	return b.String()
}
