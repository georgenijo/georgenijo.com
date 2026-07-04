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
	{"nadirclaw", "Python", "LLM router & cost optimizer",
		"Routes every prompt to the cheapest model that can actually handle it. Scores request complexity, tracks per-provider spend in real time, and fails over cleanly when a provider rate-limits."},
	{"ghosthands", "Python", "Local computer-use for macOS",
		"Drives the Mac desktop with local models: screenshots in, accessibility-tree actions out. Nothing between your screen and the agent leaves the machine."},
	{"usher", "Go", "MCP broker",
		"One MCP endpoint that fans out to many servers. Handles auth, health checks, and tool-name collisions so agents see a single clean catalog."},
	{"hangar", "Rust", "Agent control plane",
		"Where agents live between jobs. Scheduling, budgets, kill switches, and an audit log for every tool call a fleet makes."},
	{"omen", "Rust", "Code intel for AI agents",
		"Feeds an agent the ten lines that matter instead of the whole repo. Symbol graphs, blast-radius queries, and context packing tuned for LLM windows."},
	{"hermes-agent", "Python", "Errand-running personal agent",
		"A daemon that reads the inbox, watches the calendar, and drafts the boring replies. Sandboxed tools; a human signs off on anything that sends."},
	{"fleetmap", "TypeScript", "Live map of running agents",
		"One screen showing every agent, what it is doing, and what it is spending. The ops view an agent fleet deserves."},
	{"gauge", "Go", "Eval harness for agent stacks",
		"Replayable scenario suites for routers, brokers, and agents. Catch the regression before it pages you."},
	{"aperture", "Swift", "iOS film camera",
		"A film camera in your pocket: fixed stocks, 36 exposures, no preview until the roll is developed. Patience as a feature."},
	{"star-maker", "Python", "Constellations from your data",
		"Turns any dataset into a navigable night sky. Mostly an excuse to draw stars with math."},
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
	{"projects", "projects", "ten repos, four languages"},
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
