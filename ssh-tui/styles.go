package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Paper palette (design-paper-v2). AdaptiveColor picks the Light or Dark hex
// from the renderer's background; truecolor hexes degrade automatically to
// the closest 256-color (or 16-color) equivalent on lesser terminals.
var (
	aFG     = lipgloss.AdaptiveColor{Light: "#1c1710", Dark: "#e8e2d4"}
	aDim    = lipgloss.AdaptiveColor{Light: "#6e6353", Dark: "#a2957f"}
	aFaint  = lipgloss.AdaptiveColor{Light: "#a2957f", Dark: "#6e6353"}
	aAccent = lipgloss.AdaptiveColor{Light: "#b5351c", Dark: "#e06a50"}
)

// mastheadWidth is how many ═ the masthead rule repeats (matches the mock).
const mastheadWidth = 54

type styles struct {
	r *lipgloss.Renderer

	fg, dim, faint     lipgloss.Style
	accent, accentBold lipgloss.Style
	title, sub         lipgloss.Style
	kbd, hint, rule    lipgloss.Style
	pane, filterBox    lipgloss.Style
	badge, urlStyle    lipgloss.Style

	// pinkBold is a legacy field name kept because main.go's MOTD uses it
	// for the narrow-terminal wordmark; it now renders fg bold, not pink.
	pinkBold lipgloss.Style

	banner    string // pre-rendered masthead (name + double rule) for the MOTD
	bannerTag string // tagline under the masthead
}

func newStyles(r *lipgloss.Renderer) *styles {
	st := &styles{r: r}

	st.fg = r.NewStyle().Foreground(aFG)
	st.dim = r.NewStyle().Foreground(aDim)
	st.faint = r.NewStyle().Foreground(aFaint)
	st.accent = r.NewStyle().Foreground(aAccent)
	st.accentBold = st.accent.Bold(true)

	// Hierarchy comes from weight + spacing; accent is reserved for the
	// selection marker (and the prompt sig, when one is on screen).
	st.title = st.fg.Bold(true)
	st.sub = st.dim
	st.kbd = st.dim
	st.hint = st.faint
	st.rule = st.faint
	st.pane = r.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(aDim).
		Padding(0, 1)
	st.filterBox = r.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(aDim).
		Padding(0, 1)
	st.badge = st.faint
	st.urlStyle = st.fg.Underline(true)

	st.pinkBold = st.title

	st.banner = st.title.Render("GEORGE NIJO") + "\n" +
		st.rule.Render(strings.Repeat("═", mastheadWidth))
	st.bannerTag = st.dim.Render(tagline)
	return st
}
