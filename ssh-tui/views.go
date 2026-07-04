package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var spinFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// slipWidth is the total width of a slip box excluding its borders
// (1-col padding each side leaves 52 text columns, like the mock).
const slipWidth = 54

// hyperlink wraps text in an OSC 8 escape so supporting terminals make it
// clickable; others render just the text.
func hyperlink(url, text string) string {
	return "\x1b]8;;" + url + "\x07" + text + "\x1b]8;;\x07"
}

// wide is consulted by model.go for spinner cadence; the paper layout is
// single-column at every width.
func (m model) wide() bool { return m.width >= 100 }

// ── layout helpers ───────────────────────────────────────────────

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(" ", n)
}

func padTo(s string, w int) string {
	return s + spaces(w-lipgloss.Width(s))
}

// fitHeight pads or truncates a block to exactly h lines.
func fitHeight(s string, h int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > h {
		lines = lines[:h]
	}
	for len(lines) < h {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

// ── top-level view ───────────────────────────────────────────────

func (m model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	if m.booting {
		return m.renderBoot()
	}
	head := m.header(false)
	status := m.statusBar()
	bodyH := m.height - lipgloss.Height(head) - lipgloss.Height(status)
	if bodyH < 1 {
		bodyH = 1
	}
	body := m.body(bodyH)
	if lipgloss.Height(body) > bodyH {
		// Tight terminal: trade the masthead's air for body lines so the
		// selectable rows under a slip stay on screen.
		head = m.header(true)
		bodyH = max(m.height-lipgloss.Height(head)-lipgloss.Height(status), 1)
		body = m.body(bodyH)
	}
	return head + "\n" + fitHeight(body, bodyH) + "\n" + status
}

// ── boot sequence ────────────────────────────────────────────────

// renderBoot draws the intro animation: the fake ssh command typing
// itself, the connection lines, then the block GEORGE wordmark
// revealed row by row. One frame ≈ 30ms; any key skips.
func (m model) renderBoot() string {
	st := m.st
	var b strings.Builder
	b.WriteString("\n")

	// typed command with a trailing block cursor while typing
	n := min(max(m.bootFrame-3, 0), len(bootCmd))
	line := "  " + st.accentBold.Render("➜ ~ $") + " " + st.fg.Render(bootCmd[:n])
	if n < len(bootCmd) {
		line += st.dim.Render("█")
	}
	b.WriteString(line + "\n")

	// staged boot lines
	after := m.bootFrame - 3 - len(bootCmd)
	if after >= 8 {
		b.WriteString("  " + st.dim.Render(bootConnLine) + "\n")
	}
	if after >= 14 {
		b.WriteString("  " + st.dim.Render("Host key: ") + st.fg.Render(bootFingerprint) +
			" " + st.accentBold.Render("✓ known") + "\n")
	}
	if after >= 20 {
		b.WriteString("  " + st.dim.Render(bootAuthLine) + "\n")
	}

	// block wordmark, one row per two frames, then the tagline
	if after >= 28 {
		b.WriteString("\n")
		rows := min((after-28)/2+1, len(asciiGeorge))
		for i := 0; i < rows; i++ {
			b.WriteString("  " + st.title.Render(asciiGeorge[i]) + "\n")
		}
		if rows == len(asciiGeorge) && after >= 28+len(asciiGeorge)*2+4 {
			b.WriteString("\n  " + st.dim.Render(tagline) + "\n")
		}
	}

	lines := strings.Split(b.String(), "\n")
	for i, l := range lines {
		lines[i] = ansi.Truncate(l, m.width, "…")
	}
	return strings.Join(lines, "\n")
}

// ── masthead / status bar ────────────────────────────────────────

// header renders the masthead on every view, plus a breadcrumb off the menu.
// compact drops the blank separator lines for short terminals.
func (m model) header(compact bool) string {
	st := m.st
	ruleW := min(mastheadWidth, max(m.width-4, 10))
	lines := []string{
		st.title.Render("GEORGE NIJO"),
		st.rule.Render(strings.Repeat("═", ruleW)),
		st.dim.Render(tagline),
	}
	if !compact {
		lines = append(lines, "")
	}
	if c := m.crumbText(); c != "" {
		lines = append(lines, st.dim.Render(c))
		if !compact {
			lines = append(lines, "")
		}
	}
	for i, l := range lines {
		if l != "" {
			lines[i] = ansi.Truncate("  "+l, m.width, "…")
		}
	}
	return strings.Join(lines, "\n")
}

func (m model) crumbText() string {
	switch m.view {
	case viewMenu:
		return ""
	case viewProjects:
		return fmt.Sprintf("~ / projects — ledger · %d repos · five languages", len(projects))
	case viewProject:
		return "~ / projects / " + m.projName
	default:
		return "~ / " + viewNames[m.view]
	}
}

func (m model) statusBar() string {
	st := m.st
	kbd := func(k, l string) string { return st.kbd.Render(k) + " " + st.hint.Render(l) }
	var hints []string
	if m.filterTyping {
		hints = []string{kbd("type", "to filter"), kbd("enter", "apply"), kbd("esc", "cancel"), kbd("↑/↓", "navigate")}
	} else {
		switch m.view {
		case viewMenu:
			hints = []string{kbd("↑/↓", "navigate"), kbd("enter", "select"), kbd("q", "quit")}
		case viewProjects:
			hints = []string{kbd("↑/↓", "navigate"), kbd("enter", "select"), kbd("/", "filter"), kbd("esc", "back"), kbd("q", "quit")}
		default:
			hints = []string{kbd("↑/↓", "navigate"), kbd("enter", "select"), kbd("esc", "back"), kbd("q", "quit")}
		}
	}
	line := "  " + strings.Join(hints, st.hint.Render(" · "))
	line = ansi.Truncate(line, m.width, "…")
	rule := st.rule.Render(strings.Repeat("─", m.width))
	return rule + "\n" + line
}

// ── list rows ────────────────────────────────────────────────────

// displayTitle maps a row's model title to its on-screen label.
func (m model) displayTitle(r row) string {
	if r.kind == rkBack {
		return "back"
	}
	if m.view == viewProject && r.kind == rkLink {
		return "open repo ↗"
	}
	return r.title
}

// renderRows draws the current view's rows; w is the column width.
// The selected row gets an accent "› " caret; unselected rows indent 4.
// maxLines > 0 enables scrolling for long lists.
func (m model) renderRows(w, maxLines int) string {
	st := m.st
	rs := m.rows()
	var lines []string
	cursorLine := 0

	for i, r := range rs {
		sel := i == m.cursor
		prefix := "    "
		if sel {
			prefix = "  " + st.accentBold.Render("› ")
		}
		name := func(s string) string {
			if sel {
				return st.title.Render(s)
			}
			return st.fg.Render(s)
		}

		var line string
		switch {
		case m.view == viewMenu:
			line = prefix + name(padTo(fmt.Sprintf("%02d  %s", i+1, r.title), 16)) +
				st.faint.Render(r.desc)
		case r.inline: // projects ledger: NN  name  tag  lang
			tag := ansi.Truncate(r.desc, 34, "…")
			line = prefix + name(padTo(fmt.Sprintf("%02d  %s", i+1, r.title), 20)) +
				st.faint.Render(padTo(tag, 36)) + st.faint.Render(r.aux)
		default: // action rows
			t := name(m.displayTitle(r))
			if r.link != "" {
				t = hyperlink(r.link, t)
			}
			line = prefix + t
			if r.desc != "" {
				line += "  " + st.faint.Render(r.desc)
			}
		}

		if sel {
			cursorLine = len(lines)
		}
		lines = append(lines, ansi.Truncate(line, w, "…"))
	}

	if maxLines > 0 && len(lines) > maxLines {
		off := 0
		if cursorLine >= maxLines {
			off = cursorLine - maxLines + 1
		}
		end := min(off+maxLines, len(lines))
		lines = lines[off:end]
	}
	return strings.Join(lines, "\n")
}

func (m model) filterBar(w int) string {
	st := m.st
	inner := st.title.Render("/") + " " + st.fg.Render(m.filterText)
	if m.filterTyping {
		inner += st.dim.Render("█")
	} else if m.filterText == "" {
		inner += st.faint.Render("filter projects…")
	}
	bw := min(w, 48) - 4
	if bw < 8 {
		bw = 8
	}
	return st.filterBox.Width(bw).Render(ansi.Truncate(inner, bw, "…"))
}

func (m model) urlNoteBlock() string {
	if m.urlNote == "" {
		return ""
	}
	st := m.st
	return st.dim.Render("open in your browser:") + "\n  " +
		hyperlink(m.urlNote, st.urlStyle.Render(m.urlNote))
}

// ── slip content builders ────────────────────────────────────────

func kvLine(st *styles, k, v string) string {
	return st.faint.Render(padTo(k, 8)) + " " + v
}

func aboutText(st *styles) string {
	p1 := "George Nijo — software engineer in Boston. Supervisor of agents."
	p2 := "I build infrastructure for AI agents: coordination meshes, MCP brokers, and control planes — the plumbing that keeps a fleet of agents fast, cheap, and accountable. Currently building AgentOS, a personal + home agent operating system."
	p3 := "Local-first, honesty-first: build tools that solve your own problems on your own machine, then open-source the parts worth sharing. Ship the menu bar app before the platform."
	p4 := st.dim.Render("Python · Go · Rust · Swift · a dash of TypeScript")
	return p1 + "\n\n" + p2 + "\n\n" + p3 + "\n\n" + p4
}

func nowText(st *styles, spin int) string {
	p1 := "Heads-down on AgentOS — one control plane for the agents that run my code, my inbox, and (eventually) the house."
	if buildStamp != "" {
		p1 += "\n" + st.dim.Render("updated "+buildStamp)
	}
	kv := func(k, v string) string { return st.faint.Render(padTo(k, 10)) + " " + st.dim.Render(v) }
	kvs := kv("hangar", "scheduling + budgets for long-lived agents") + "\n" +
		kv("usher", "auth story for brokered MCP servers") + "\n" +
		kv("agent-mesh", "keeping parallel agents out of each other's files")
	spinner := st.dim.Render(spinFrames[spin%len(spinFrames)] + " agents working…")
	return p1 + "\n\n" + kvs + "\n\n" + spinner
}

func contactText(st *styles) string {
	return st.dim.Render("no forms, no funnels — just these three:") + "\n\n" +
		kvLine(st, "mailto", hyperlink("mailto:"+email, st.fg.Render(email))) + "\n" +
		kvLine(st, "github", hyperlink(ghBase, st.fg.Render("github.com/georgenijo"))) + "\n" +
		kvLine(st, "linkedin", hyperlink(liURL, st.fg.Render("linkedin.com/in/georgenijo")))
}

func coffeeText(st *styles, tried bool) string {
	if tried {
		return st.title.Render("order failed: no coffee endpoint on this host.") + "\n\n" +
			st.dim.Render("The café is next door — terminal.shop already has coffee-over-ssh covered. This server only pours bytes.")
	}
	return "You can, famously, buy coffee over SSH." + "\n\n" +
		st.dim.Render("Not here, though. The café is next door; this host serves a portfolio and nothing hot.")
}

func projectDetailText(st *styles, p *project) string {
	url := ghBase + "/" + p.Name
	repo := hyperlink(url, st.dim.Render("github.com/georgenijo/"+p.Name))
	return p.Desc + "\n\n" +
		kvLine(st, "repo", repo) + "\n" +
		kvLine(st, "lang", st.dim.Render(p.Lang))
}

func projectTitleLine(st *styles, p *project) string {
	return st.title.Render(p.Name) + "  " + st.badge.Render("["+p.Lang+"]")
}

// ── single-column body ───────────────────────────────────────────

func (m model) body(h int) string {
	st := m.st
	cw := m.width - 4
	if cw < 20 {
		cw = max(m.width-2, 10)
	}
	slipW := min(slipWidth, cw-2)
	slip := func(text string) string { return st.pane.Width(slipW).Render(text) }

	var b []string
	add := func(s ...string) { b = append(b, s...) }

	switch m.view {
	case viewMenu:
		add(st.faint.Render("index"), "")
		add(m.renderRows(cw, max(h-2, 1)))
	case viewProjects:
		if m.filterOpen {
			add(m.filterBar(cw), "")
		}
		if len(m.rows()) == 0 {
			add(st.faint.Render("nothing matches “" + m.filterText + "” — esc to clear"))
		} else {
			used := len(b)
			add(m.renderRows(cw, max(h-used, 3)))
		}
	case viewProject:
		p := projByName(m.projName)
		if p == nil {
			break
		}
		add(projectTitleLine(st, p), st.sub.Render(p.Tag), "")
		add(slip(projectDetailText(st, p)), "")
		add(m.renderRows(cw, 0))
		if n := m.urlNoteBlock(); n != "" {
			add("", n)
		}
	case viewAbout:
		add(st.title.Render("about"), "")
		add(slip(aboutText(st)), "")
		add(m.renderRows(cw, 0))
	case viewNow:
		add(st.title.Render("now"), "")
		add(slip(nowText(st, m.spin)), "")
		add(m.renderRows(cw, 0))
	case viewContact:
		add(st.title.Render("contact"), "")
		add(slip(contactText(st)), "")
		add(m.renderRows(cw, 0))
		if n := m.urlNoteBlock(); n != "" {
			add("", n)
		}
	case viewCoffee:
		add(st.title.Render("coffee"), "")
		add(slip(coffeeText(st, m.coffeeTried)), "")
		add(m.renderRows(cw, 0))
	}

	return st.r.NewStyle().Padding(0, 2).Render(strings.Join(b, "\n"))
}
