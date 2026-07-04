package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func testModel(w, h int) model {
	r := lipgloss.NewRenderer(io.Discard)
	r.SetColorProfile(termenv.TrueColor)
	return newModel(newStyles(r), w, h)
}

func press(t *testing.T, m model, keys ...string) model {
	t.Helper()
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "up":
			msg = tea.KeyMsg{Type: tea.KeyUp}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		nm, _ := m.Update(msg)
		m = nm.(model)
	}
	return m
}

// TestRenderSnapshots writes plain-text dumps of every view for eyeballing.
func TestRenderSnapshots(t *testing.T) {
	var b strings.Builder
	dump := func(name string, m model) {
		b.WriteString("===== " + name + " =====\n")
		b.WriteString(ansi.Strip(m.View()))
		b.WriteString("\n\n")
	}

	wm := testModel(120, 35)
	dump("menu wide 120x35", wm)
	dump("about wide", press(t, wm, "enter"))
	dump("projects wide", press(t, wm, "j", "enter"))
	dump("project detail wide", press(t, wm, "j", "enter", "j", "j", "enter"))
	dump("projects filtered wide", press(t, wm, "j", "enter", "/", "r", "u"))
	dump("now wide", press(t, wm, "j", "j", "enter"))
	dump("contact wide", press(t, wm, "j", "j", "j", "enter"))
	dump("contact url note", press(t, wm, "j", "j", "j", "enter", "enter"))
	dump("coffee tried wide", press(t, wm, "j", "j", "j", "j", "enter", "enter"))

	nm := testModel(80, 24)
	dump("menu narrow 80x24", nm)
	dump("projects narrow", press(t, nm, "j", "enter"))
	dump("project detail narrow", press(t, nm, "j", "enter", "enter"))
	dump("about narrow", press(t, nm, "enter"))

	out := os.Getenv("SNAPSHOT_OUT")
	if out == "" {
		out = filepath.Join(t.TempDir(), "snapshots.txt")
	}
	if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("snapshots written to %s", out)
}

// TestInteraction covers the core key flows.
func TestInteraction(t *testing.T) {
	m := testModel(120, 35)

	// navigate to projects, filter to "rust", expect 2 results
	m2 := press(t, m, "j", "enter", "/", "r", "u", "s", "t")
	if got := len(m2.filteredProjects()); got != 2 {
		t.Fatalf("rust filter: want 2 projects, got %d", got)
	}
	// esc clears filter
	m3 := press(t, m2, "esc")
	if m3.filterOpen || m3.filterText != "" {
		t.Fatal("esc should close and clear the filter")
	}
	// esc again goes back to menu
	m4 := press(t, m3, "esc")
	if m4.view != viewMenu {
		t.Fatalf("esc from projects should return to menu, got view %v", m4.view)
	}
	// enter on a project opens detail
	m5 := press(t, m, "j", "enter", "j", "j", "enter")
	if m5.view != viewProject || m5.projName != "usher" {
		t.Fatalf("expected usher detail, got view=%v proj=%q", m5.view, m5.projName)
	}
	// coffee gag toggles
	m6 := press(t, m, "j", "j", "j", "j", "enter", "enter")
	if !m6.coffeeTried {
		t.Fatal("coffee order should set coffeeTried")
	}
	if v := press(t, m6, "esc"); v.coffeeTried {
		t.Fatal("leaving coffee should reset coffeeTried")
	}
	// q quits
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("q should quit")
	}
}
