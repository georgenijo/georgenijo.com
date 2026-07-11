package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type viewID int

const (
	viewMenu viewID = iota
	viewAbout
	viewProjects
	viewProject
	viewNow
	viewContact
	viewCoffee
	viewBurn
)

var viewNames = map[viewID]string{
	viewMenu:     "menu",
	viewAbout:    "about",
	viewProjects: "projects",
	viewProject:  "project",
	viewNow:      "now",
	viewContact:  "contact",
	viewCoffee:   "coffee",
	viewBurn:     "burn",
}

var parentOf = map[viewID]viewID{
	viewAbout:    viewMenu,
	viewProjects: viewMenu,
	viewNow:      viewMenu,
	viewContact:  viewMenu,
	viewCoffee:   viewMenu,
	viewBurn:     viewMenu,
	viewProject:  viewProjects,
}

var viewByID = map[string]viewID{
	"about":    viewAbout,
	"projects": viewProjects,
	"now":      viewNow,
	"contact":  viewContact,
	"coffee":   viewCoffee,
	"burn":     viewBurn,
}

type rowKind int

const (
	rkGoto rowKind = iota
	rkBack
	rkProject
	rkLink
	rkCoffee
)

// row is one selectable item in the current view (a Bubbletea list entry).
type row struct {
	title  string
	desc   string
	aux    string
	act    bool   // "→ " action marker, like the site's .act rows
	inline bool   // single-line layout (projects list)
	link   string // OSC 8 hyperlink target, if any
	kind   rowKind
	arg    string
}

type tickMsg time.Time

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

type model struct {
	st            *styles
	width, height int

	view     viewID
	projName string

	cursor  int
	cursors map[string]int

	filterOpen   bool
	filterTyping bool
	filterText   string

	coffeeTried bool
	urlNote     string

	start     time.Time
	lastInput time.Time
	spin      int

	booting   bool // intro boot animation; any key skips
	bootFrame int
}

// bootTotalFrames is the full boot animation length in 30ms frames:
// lead-in, typed command, staged boot lines, banner rows, hold.
var bootTotalFrames = 3 + len(bootCmd) + 28 + len(asciiGeorge)*2 + 4 + 15

func newModel(st *styles, w, h int) model {
	now := time.Now()
	return model{
		st:        st,
		width:     w,
		height:    h,
		view:      viewMenu,
		cursors:   map[string]int{},
		start:     now,
		lastInput: now,
	}
}

func (m model) Init() tea.Cmd {
	if m.booting {
		return tick(30 * time.Millisecond)
	}
	return tick(time.Second)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tickMsg:
		if m.booting {
			m.bootFrame++
			if m.bootFrame > bootTotalFrames {
				m.booting = false
			}
			return m, tick(m.tickInterval())
		}
		m.spin++
		if time.Since(m.lastInput) > idleLimit {
			return m, tea.Quit
		}
		return m, tick(m.tickInterval())

	case tea.KeyMsg:
		m.lastInput = time.Now()
		if m.booting {
			if msg.Type == tea.KeyCtrlC {
				return m, tea.Quit
			}
			m.booting = false // any key skips the boot animation
			return m, nil
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.filterTyping {
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			m.closeFilter()
		case tea.KeyEnter:
			m.filterTyping = false
		case tea.KeyBackspace:
			if r := []rune(m.filterText); len(r) > 0 {
				m.filterText = string(r[:len(r)-1])
				m.cursor = 0
			}
		case tea.KeyUp:
			m.move(-1)
		case tea.KeyDown:
			m.move(1)
		case tea.KeySpace:
			m.filterText += " "
			m.cursor = 0
		case tea.KeyRunes:
			m.filterText += string(msg.Runes)
			m.cursor = 0
		}
		m.clampCursor()
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		m.move(-1)
	case "down", "j":
		m.move(1)
	case "home", "g":
		m.cursor = 0
	case "end", "G":
		m.cursor = max(0, len(m.rows())-1)
	case "enter":
		m.activate()
	case "esc":
		m.back()
	case "/":
		if m.view == viewProjects {
			m.filterOpen, m.filterTyping = true, true
			m.urlNote = ""
		}
	}
	m.clampCursor()
	return m, nil
}

func (m model) tickInterval() time.Duration {
	if m.booting {
		return 30 * time.Millisecond
	}
	if m.spinnerVisible() {
		return 120 * time.Millisecond
	}
	return time.Second
}

// spinnerVisible reports whether the "agents working…" spinner is on screen.
func (m model) spinnerVisible() bool {
	if m.view == viewNow {
		return true
	}
	if m.wide() && m.view == viewMenu {
		if m.cursor >= 0 && m.cursor < len(menuItems) {
			return menuItems[m.cursor].id == "now"
		}
	}
	return false
}

// ── rows per view ────────────────────────────────────────────────

func (m model) filteredProjects() []project {
	q := strings.ToLower(strings.TrimSpace(m.filterText))
	if q == "" {
		return projects
	}
	var out []project
	for _, p := range projects {
		hay := strings.ToLower(p.Name + " " + p.Tag + " " + p.Lang)
		if strings.Contains(hay, q) {
			out = append(out, p)
		}
	}
	return out
}

func (m model) rows() []row {
	switch m.view {
	case viewMenu:
		rs := make([]row, 0, len(menuItems))
		for _, it := range menuItems {
			rs = append(rs, row{title: it.title, desc: it.desc, kind: rkGoto, arg: it.id})
		}
		return rs

	case viewAbout:
		return []row{
			{title: "see the projects", act: true, kind: rkGoto, arg: "projects"},
			{title: "← back", kind: rkBack},
		}

	case viewProjects:
		list := m.filteredProjects()
		rs := make([]row, 0, len(list))
		for _, p := range list {
			rs = append(rs, row{title: p.Name, desc: p.Tag, aux: p.Lang, inline: true, kind: rkProject, arg: p.Name})
		}
		return rs

	case viewProject:
		p := projByName(m.projName)
		if p == nil {
			return nil
		}
		url := ghBase + "/" + p.Name
		return []row{
			{title: "open github.com/georgenijo/" + p.Name + " ↗", act: true, link: url, kind: rkLink, arg: url},
			{title: "← back to projects", kind: rkBack},
		}

	case viewNow:
		return []row{
			{title: "browse the parts", act: true, kind: rkGoto, arg: "projects"},
			{title: "← back", kind: rkBack},
		}

	case viewContact:
		mailto := "mailto:" + email
		return []row{
			{title: email, desc: "email — the reliable channel", aux: "mailto", link: mailto, kind: rkLink, arg: mailto},
			{title: "github.com/georgenijo", desc: "the code lives here", aux: "https", link: ghBase, kind: rkLink, arg: ghBase},
			{title: "linkedin.com/in/georgenijo", desc: "the professional channel", aux: "https", link: liURL, kind: rkLink, arg: liURL},
			{title: "← back", kind: rkBack},
		}

	case viewCoffee:
		order := "place order"
		if m.coffeeTried {
			order = "try ordering again (it won't help)"
		}
		return []row{
			{title: order, act: true, kind: rkCoffee},
			{title: "← back", kind: rkBack},
		}

	case viewBurn:
		// No selectable rows except back — detail is in the slip
		return []row{
			{title: "← back", kind: rkBack},
		}
	}
	return nil
}

// ── navigation ───────────────────────────────────────────────────

func (m *model) move(delta int) {
	n := len(m.rows())
	if n == 0 {
		return
	}
	m.cursor = min(max(m.cursor+delta, 0), n-1)
}

func (m *model) clampCursor() {
	n := len(m.rows())
	if n == 0 {
		m.cursor = 0
		return
	}
	m.cursor = min(max(m.cursor, 0), n-1)
}

func (m *model) activate() {
	rs := m.rows()
	if m.cursor < 0 || m.cursor >= len(rs) {
		return
	}
	r := rs[m.cursor]
	switch r.kind {
	case rkGoto:
		m.goTo(viewByID[r.arg], "")
	case rkBack:
		m.back()
	case rkProject:
		m.goTo(viewProject, r.arg)
	case rkLink:
		m.urlNote = r.arg
	case rkCoffee:
		m.coffeeTried = true
	}
}

func (m model) viewKey() string {
	if m.view == viewProject {
		return "project:" + m.projName
	}
	return viewNames[m.view]
}

func (m *model) goTo(v viewID, proj string) {
	m.cursors[m.viewKey()] = m.cursor
	m.view = v
	m.projName = proj
	m.filterOpen, m.filterTyping, m.filterText = false, false, ""
	m.urlNote = ""
	if v != viewCoffee {
		m.coffeeTried = false
	}
	if c, ok := m.cursors[m.viewKey()]; ok && c >= 0 && c < len(m.rows()) {
		m.cursor = c
	} else {
		m.cursor = 0
	}
}

func (m *model) closeFilter() {
	m.filterOpen, m.filterTyping, m.filterText = false, false, ""
	if c, ok := m.cursors["projects"]; ok && c < len(m.rows()) {
		m.cursor = c
	} else {
		m.cursor = 0
	}
}

func (m *model) back() {
	if m.filterOpen {
		m.closeFilter()
		return
	}
	if m.urlNote != "" {
		m.urlNote = ""
		return
	}
	if p, ok := parentOf[m.view]; ok {
		proj := ""
		if p == viewProject {
			proj = m.projName
		}
		m.goTo(p, proj)
	}
}
