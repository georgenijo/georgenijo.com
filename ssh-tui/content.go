package main

// Content mirrored from georgenijo.com (concept-5-ssh.html).

const (
	ghBase   = "https://github.com/georgenijo"
	email    = "george.nijo8@gmail.com"
	hostName = "georgenijo.com"
)

type project struct {
	Name, Lang, Tag, Desc string
}

var projects = []project{
	{"agent-mesh", "Go", "Coordination layer for agents",
		"A shared nervous system for the coding agents on your machine. Claude Code, Codex, and Cursor discover each other, lock files instead of fighting over them, and trade notes on a blackboard — one mesh CLI runs it all."},
	{"usher", "Go", "MCP broker",
		"One MCP endpoint that fans out to many servers. Handles auth, health checks, and tool-name collisions so agents see a single clean catalog."},
	{"hangar", "Rust", "Agent control plane",
		"Where agents live between jobs. Scheduling, budgets, kill switches, and an audit log for every tool call a fleet makes."},
	{"ghosthands", "Python", "Local computer-use for macOS",
		"Drives the Mac desktop with local models: screenshots in, accessibility-tree actions out. Nothing between your screen and the agent leaves the machine. A native Swift sibling does the low-level driving."},
	{"murmur-app", "Rust", "Local voice-to-text for macOS",
		"Hold a key, speak, release — the words land in whatever app has focus. whisper.cpp on the GPU; no cloud, no subscription, nothing leaves the machine."},
	{"fleetmap", "Swift", "Live map of what your Mac runs",
		"Every process as a node — sized by RAM, colored by CPU, wired by its real sockets. Activity Monitor tells you what is running; this shows what is talking to what."},
	{"gauge", "Swift", "Claude usage in the menu bar",
		"A one-purpose menu-bar meter: session and weekly limits, pace until reset, and spend estimated from local logs. Know the budget before the model does."},
	{"whoop-dashboard", "TypeScript", "Health analytics with an AI coach",
		"WHOOP sleep, strain, and recovery in one dashboard, with a Claude coach that reads the trends and talks back."},
	{"aperture", "Swift", "iOS film camera",
		"A film camera in your pocket: fixed stocks, 36 exposures, no preview until the roll is developed. Patience as a feature."},
}

func projByName(name string) *project {
	for i := range projects {
		if projects[i].Name == name {
			return &projects[i]
		}
	}
	return nil
}

type menuItem struct {
	id    string
	title string
	desc  string
}

var menuItems = []menuItem{
	{"about", "about", "who's at the keyboard"},
	{"projects", "projects", "nine repos, five languages"},
	{"now", "now", "what's cooking — AgentOS"},
	{"contact", "contact", "say hello"},
	{"coffee", "coffee", "order a cup over ssh"},
}

// The gradient "GEORGE" wordmark (52 columns wide).
var asciiGeorge = []string{
	"  ██████╗ ███████╗ ██████╗ ██████╗  ██████╗ ███████╗",
	" ██╔════╝ ██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██╔════╝",
	" ██║  ███╗█████╗  ██║   ██║██████╔╝██║  ███╗█████╗  ",
	" ██║   ██║██╔══╝  ██║   ██║██╔══██╗██║   ██║██╔══╝  ",
	" ╚██████╔╝███████╗╚██████╔╝██║  ██║╚██████╔╝███████╗",
	"  ╚═════╝ ╚══════╝ ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝",
}

const bannerWidth = 52

const tagline = "software engineer · boston · agent infrastructure"
