package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var spinFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// hyperlink wraps text in an OSC 8 escape so supporting terminals make it
// clickable; others render just the text.
func hyperlink(url, text string) string {
	return "\x1b]8;;" + url + "\x07" + text + "\x1b]8;;\x07"
}

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
	head := m.header()
	status := m.statusBar()
	bodyH := m.height - lipgloss.Height(head) - lipgloss.Height(status)
	if bodyH < 1 {
		bodyH = 1
	}

	var banner string
	if m.view == viewMenu && m.width >= bannerWidth+4 && m.height >= 27 {
		banner = m.bannerBlock()
		bodyH -= lipgloss.Height(banner)
		if bodyH < 1 {
			banner, bodyH = "", 1
		}
	}

	var body string
	if m.wide() {
		body = m.splitBody(bodyH)
	} else {
		body = m.singleBody(bodyH)
	}
	body = fitHeight(body, bodyH)

	parts := []string{head}
	if banner != "" {
		parts = append(parts, banner)
	}
	parts = append(parts, body, status)
	return strings.Join(parts, "\n")
}

// bannerBlock is the MOTD wordmark shown at the top of the main menu.
func (m model) bannerBlock() string {
	pad := m.st.r.NewStyle().Padding(0, 2)
	return pad.Render(m.st.banner+"\n"+m.st.bannerTag) + "\n"
}

// ── header / status bar ──────────────────────────────────────────

func (m model) crumb() string {
	st := m.st
	switch m.view {
	case viewMenu:
		return st.lav.Render(hostName) + st.dim.Render(" ~")
	case viewProject:
		return st.lav.Render("~") + st.dim.Render(" / projects / "+m.projName)
	default:
		return st.lav.Render("~") + st.dim.Render(" / "+viewNames[m.view])
	}
}

func (m model) header() string {
	st := m.st
	left := " " + st.chip + "  " + m.crumb()
	conn := st.faint.Render("ssh · ed25519")
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(conn) - 1
	line := left
	if gap >= 2 {
		line += spaces(gap) + conn
	}
	line = ansi.Truncate(line, m.width, "…")
	rule := st.rule.Render(strings.Repeat("─", m.width))
	return line + "\n" + rule
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
	line := " " + strings.Join(hints, st.faint.Render("  ·  "))
	line = ansi.Truncate(line, m.width, "…")
	rule := st.rule.Render(strings.Repeat("─", m.width))
	return rule + "\n" + line
}

// ── list rows ────────────────────────────────────────────────────

// renderRows draws the current view's rows; w is the column width.
// maxLines > 0 enables scrolling for long single-line lists.
func (m model) renderRows(w, maxLines int) string {
	st := m.st
	rs := m.rows()
	var lines []string
	cursorLine := 0

	for i, r := range rs {
		sel := i == m.cursor
		bar, pad := "  ", "  "
		if sel {
			bar = st.pink.Render("┃") + " "
		}
		mark := ""
		if r.act {
			if sel {
				mark = st.pink.Render("→ ")
			} else {
				mark = st.faint.Render("→ ")
			}
		}
		var title string
		if r.inline {
			name := padTo(r.title, 14)
			if sel {
				title = st.pinkBold.Render(name)
			} else {
				title = st.fg.Render(name)
			}
		} else if sel {
			title = st.pinkBold.Render(r.title)
		} else {
			title = st.fg.Render(r.title)
		}
		if r.link != "" {
			title = hyperlink(r.link, title)
		}
		line := bar + mark + title

		if r.inline && r.desc != "" {
			d := r.desc
			budget := w - lipgloss.Width(line) - lipgloss.Width(r.aux) - 4
			if budget > 4 {
				d = ansi.Truncate(d, budget, "…")
				if sel {
					line += " " + st.lav.Render(d)
				} else {
					line += " " + st.dim.Render(d)
				}
			}
		}
		if r.aux != "" {
			aux := st.faint.Render(r.aux)
			if sel {
				aux = st.lav.Render(r.aux)
			}
			gap := w - lipgloss.Width(line) - lipgloss.Width(r.aux)
			if gap < 1 {
				gap = 1
			}
			line += spaces(gap) + aux
		}
		if sel {
			cursorLine = len(lines)
		}
		lines = append(lines, ansi.Truncate(line, w, "…"))

		if !r.inline && r.desc != "" {
			d := ansi.Truncate(r.desc, w-4, "…")
			if sel {
				lines = append(lines, pad+"  "+st.lav.Render(d))
			} else {
				lines = append(lines, pad+"  "+st.faint.Render(d))
			}
		}
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
	inner := st.pinkBold.Render("/") + " " + st.fg.Render(m.filterText)
	if m.filterTyping {
		inner += st.pink.Render("█")
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
	return st.pink.Render("→ open in your browser:") + "\n  " +
		hyperlink(m.urlNote, st.urlStyle.Render(m.urlNote))
}

// ── content builders (shared by narrow column and preview pane) ──

func kvLine(st *styles, k, v string) string {
	return st.faint.Render(padTo(k, 7)) + " " + v
}

func aboutText(st *styles) string {
	p1 := st.pink.Render("George Nijo") + " — software engineer in " + st.lav.Render("Boston") + "."
	p2 := "I build infrastructure for AI agents: " + st.lav.Render("LLM routers") + ", " +
		st.lav.Render("MCP brokers") + ", and " + st.lav.Render("agent control planes") +
		" — the plumbing that keeps a fleet of agents fast, cheap, and accountable."
	p3 := "Currently building " + st.pink.Render("AgentOS") + ", a personal + home agent operating system."
	p4 := st.dim.Render("Python · Go · Rust · just enough Swift")
	return p1 + "\n\n" + p2 + "\n\n" + p3 + "\n\n" + p4
}

func nowText(st *styles, spin int) string {
	p1 := "Heads-down on " + st.pink.Render("AgentOS") + " — a personal + home agent operating system. " +
		"One control plane for the agents that run my code, my inbox, and (eventually) the house."
	kvs := kvLine(st, "hangar", st.dim.Render("scheduling + budgets for long-lived agents")) + "\n" +
		kvLine(st, "usher", st.dim.Render("auth story for brokered MCP servers")) + "\n" +
		kvLine(st, "omen", st.dim.Render("context packing that fits the window"))
	spinner := st.pink.Render(spinFrames[spin%len(spinFrames)]) + st.dim.Render(" agents working…")
	return p1 + "\n\n" + kvs + "\n\n" + spinner
}

func contactText(st *styles) string {
	return st.dim.Render("no forms, no funnels — just these two:") + "\n\n" +
		kvLine(st, "mailto", hyperlink("mailto:"+email, st.lav.Render(email))) + "\n" +
		kvLine(st, "https", hyperlink(ghBase, st.lav.Render("github.com/georgenijo")))
}

func coffeeText(st *styles, tried bool) string {
	if tried {
		return st.pink.Render("order failed: no coffee endpoint on this host.") + "\n\n" +
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

// projectIndexText is the preview shown when "projects" is highlighted in the menu.
func projectIndexText(st *styles, w int) string {
	var lines []string
	for _, p := range projects {
		name := st.lav.Render(padTo(p.Name, 13))
		lang := st.faint.Render(p.Lang)
		tagBudget := w - 13 - lipgloss.Width(p.Lang) - 3
		tag := st.dim.Render(ansi.Truncate(p.Tag, max(tagBudget, 6), "…"))
		line := name + " " + tag
		gap := w - lipgloss.Width(line) - lipgloss.Width(p.Lang)
		if gap < 1 {
			gap = 1
		}
		lines = append(lines, ansi.Truncate(line+spaces(gap)+lang, w, "…"))
	}
	lines = append(lines, "", st.faint.Render(fmt.Sprintf("%d repos · enter to browse", len(projects))))
	return strings.Join(lines, "\n")
}

// ── narrow (single-column) body ──────────────────────────────────

func (m model) singleBody(h int) string {
	st := m.st
	cw := min(m.width-4, 78)
	if cw < 20 {
		cw = max(m.width-2, 10)
	}
	paneW := min(cw-4, 60)
	pane := func(text string) string { return st.pane.Width(paneW).Render(text) }

	var b []string
	add := func(s ...string) { b = append(b, s...) }

	switch m.view {
	case viewMenu:
		add(st.title.Render("hi, i'm george."), st.sub.Render("software engineer · boston · pick a door"), "")
		add(m.renderRows(cw, 0))
	case viewAbout:
		add(st.title.Render("about"), st.sub.Render("whoami"), "")
		add(pane(aboutText(st)), "")
		add(m.renderRows(cw, 0))
	case viewProjects:
		add(st.title.Render("projects"), st.sub.Render(fmt.Sprintf("%d items · enter for details", len(projects))), "")
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
		add(pane(projectDetailText(st, p)), "")
		add(m.renderRows(cw, 0))
		if n := m.urlNoteBlock(); n != "" {
			add("", n)
		}
	case viewNow:
		add(st.title.Render("now"), st.sub.Render("updated june 2026"), "")
		add(pane(nowText(st, m.spin)), "")
		add(m.renderRows(cw, 0))
	case viewContact:
		add(st.title.Render("contact"), st.sub.Render("no forms, no funnels — just these two"), "")
		add(m.renderRows(cw, 0))
		if n := m.urlNoteBlock(); n != "" {
			add("", n)
		}
	case viewCoffee:
		add(st.title.Render("coffee"), st.sub.Render("cream & sugar over port 22"), "")
		add(pane(coffeeText(st, m.coffeeTried)), "")
		add(m.renderRows(cw, 0))
	}

	return st.r.NewStyle().Padding(0, 2).Render(strings.Join(b, "\n"))
}

// ── wide (master-detail) body ────────────────────────────────────

func (m model) splitBody(h int) string {
	leftW := m.width * 38 / 100
	leftW = min(max(leftW, 36), 52)
	left := fitHeight(m.leftColumn(leftW, h), h)
	left = m.st.r.NewStyle().Width(leftW).Render(left)
	right := m.previewBox(m.width-leftW, h)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m model) leftColumn(w, h int) string {
	st := m.st
	cw := w - 4
	var b []string
	add := func(s ...string) { b = append(b, s...) }

	switch m.view {
	case viewMenu:
		add(st.title.Render("hi, i'm george."), st.sub.Render("software engineer · boston · pick a door"), "")
		add(m.renderRows(cw, 0))
	case viewAbout:
		add(st.title.Render("about"), st.sub.Render("whoami"), "")
		add(m.renderRows(cw, 0))
	case viewProjects:
		add(st.title.Render("projects"), st.sub.Render(fmt.Sprintf("%d items · enter for details", len(projects))), "")
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
		add(m.renderRows(cw, 0))
		if n := m.urlNoteBlock(); n != "" {
			add("", n)
		}
	case viewNow:
		add(st.title.Render("now"), st.sub.Render("updated june 2026"), "")
		add(m.renderRows(cw, 0))
	case viewContact:
		add(st.title.Render("contact"), st.sub.Render("no forms, no funnels — just these two"), "")
		add(m.renderRows(cw, 0))
		if n := m.urlNoteBlock(); n != "" {
			add("", n)
		}
	case viewCoffee:
		add(st.title.Render("coffee"), st.sub.Render("cream & sugar over port 22"), "")
		add(m.renderRows(cw, 0))
	}

	return st.r.NewStyle().Padding(0, 2).Render(strings.Join(b, "\n"))
}

// previewContent returns the breadcrumb path and body for the preview pane.
func (m model) previewContent(w int) (string, string) {
	st := m.st
	switch m.view {
	case viewMenu:
		idx := min(max(m.cursor, 0), len(menuItems)-1)
		it := menuItems[idx]
		path := "~ / " + it.id
		switch it.id {
		case "about":
			return path, aboutText(st)
		case "projects":
			return path, projectIndexText(st, w)
		case "now":
			return path, nowText(st, m.spin)
		case "contact":
			return path, contactText(st)
		default:
			return path, coffeeText(st, m.coffeeTried)
		}

	case viewProjects:
		list := m.filteredProjects()
		if len(list) == 0 {
			return "~ / projects", st.faint.Render("no matching project")
		}
		p := &list[min(max(m.cursor, 0), len(list)-1)]
		return "~ / projects / " + p.Name,
			projectTitleLine(st, p) + "\n" + st.sub.Render(p.Tag) + "\n\n" +
				projectDetailText(st, p) + "\n\n" + st.faint.Render("enter → open actions")

	case viewProject:
		p := projByName(m.projName)
		if p == nil {
			return "~ / projects", ""
		}
		return "~ / projects / " + p.Name,
			projectTitleLine(st, p) + "\n" + st.sub.Render(p.Tag) + "\n\n" + projectDetailText(st, p)

	case viewAbout:
		return "~ / about", aboutText(st)
	case viewNow:
		return "~ / now", nowText(st, m.spin)
	case viewContact:
		return "~ / contact", contactText(st)
	case viewCoffee:
		return "~ / coffee", coffeeText(st, m.coffeeTried)
	}
	return "~", ""
}

func fmtUptime(d time.Duration) string {
	t := int(d.Seconds())
	return fmt.Sprintf("%02d:%02d:%02d", t/3600, (t%3600)/60, t%60)
}

// previewBox draws the bordered live-preview pane with title and footer rows.
func (m model) previewBox(w, h int) string {
	st := m.st
	bw := w - 3 // 1 col gap left, 2 cols gap right
	if bw < 20 || h < 7 {
		return ""
	}
	iw := bw - 2 // inside the border
	tw := iw - 2 // text width inside padding

	path, body := m.previewContent(tw)
	bodyLines := strings.Split(st.r.NewStyle().Width(tw).Render(body), "\n")
	bodyH := h - 6
	if len(bodyLines) > bodyH {
		bodyLines = bodyLines[:bodyH]
	}
	for len(bodyLines) < bodyH {
		bodyLines = append(bodyLines, "")
	}

	title := st.pink.Render("◉") + " " + st.dim.Render(path)

	// footer: host · pts · COLSxROWS · up HH:MM:SS (right aligned)
	sep := st.faint.Render("  ·  ")
	footL := st.faint.Render(hostName) + sep + st.faint.Render("pts/0") + sep +
		st.faint.Render(fmt.Sprintf("%d×%d", m.width, m.height))
	footR := st.faint.Render("up " + fmtUptime(time.Since(m.start)))
	gap := tw - lipgloss.Width(footL) - lipgloss.Width(footR)
	foot := footL + spaces(max(gap, 1)) + footR

	bar := strings.Repeat("─", iw)
	rowLine := func(s string) string {
		s = ansi.Truncate(s, tw, "…")
		return st.rule.Render("│") + " " + padTo(s, tw) + " " + st.rule.Render("│")
	}

	var out []string
	out = append(out, st.rule.Render("╭"+bar+"╮"))
	out = append(out, rowLine(title))
	out = append(out, st.rule.Render("├"+bar+"┤"))
	for _, l := range bodyLines {
		out = append(out, rowLine(l))
	}
	out = append(out, st.rule.Render("├"+bar+"┤"))
	out = append(out, rowLine(foot))
	out = append(out, st.rule.Render("╰"+bar+"╯"))

	pref := " "
	for i := range out {
		out[i] = pref + out[i]
	}
	return strings.Join(out, "\n")
}
