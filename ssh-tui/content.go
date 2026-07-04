package main

// Content mirrored from georgenijo.com (concept-5-ssh.html).

const (
	ghBase   = "https://github.com/georgenijo"
	liURL    = "https://www.linkedin.com/in/georgenijo/"
	email    = "george.nijo8@gmail.com"
	hostName = "georgenijo.com"
)

type project struct {
	Name, Lang, Tag, Desc string
}

var projects = []project{
	{"agent-mesh", "Go", "Agents orchestrating themselves",
		"Hand the mesh a vague goal and the agents work out the rest: who does what, who owns which file, what's already been answered. Presence, file claims, async questions, a shared blackboard — Claude Code, Codex, and Cursor acting like one team instead of five strangers."},
	{"usher", "Go", "MCP broker",
		"One front desk every agent talks to: a single MCP endpoint that routes, trims, arbitrates, and audits tool calls across a fleet of servers. Auth, health, and name collisions are handled before the agent even notices."},
	{"hangar", "Rust", "Agent control plane",
		"Mission control for long-lived coding agents: scheduling, budgets, kill switches, and an audit log of every tool call a fleet makes. Where agents live between jobs — and answer for what they did."},
	{"ghosthands", "Python", "Local computer-use for macOS",
		"Full computer-use with zero cloud: a local model reads the screen and drives native Mac apps through the accessibility tree — and never fakes success. A Swift-native sibling does the low-level driving. Nothing leaves the machine."},
	{"murmur-app", "Rust", "Local voice-to-text for macOS",
		"Hold a key, speak, release — the words land in whatever app has focus before your hands find the keyboard. Whisper on the GPU, fully offline: no cloud, no subscription, no audio leaving the machine."},
	{"fleetmap", "Swift", "Live map of what your Mac runs",
		"Your Mac as a living map: every process a node — sized by RAM, colored by CPU, wired by its real sockets. Activity Monitor tells you what's running; fleetmap shows what's talking to what."},
	{"gauge", "Swift", "Claude usage in the menu bar",
		"A one-purpose menu-bar meter for Claude: session and weekly limits, pace until reset, and spend charted straight from local logs. Know the budget is gone before the model does."},
	{"whoop-dashboard", "TypeScript", "Health analytics with an AI coach",
		"Sleep, strain, and recovery in one dashboard — with a Claude coach that reads months of WHOOP trends and talks back. Health data that finally answers questions."},
	{"aperture", "Swift", "iOS film camera",
		"A film camera in your pocket: fixed stocks, 36 exposures, and no preview until the roll is developed. A Metal pipeline paints the light leaks, grain, and date stamps. Patience as a feature."},
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

// Boot sequence lines, mirroring the site's fake-ssh boot.
const (
	bootCmd         = "ssh georgenijo.com"
	bootConnLine    = "Connecting to georgenijo.com port 22."
	bootFingerprint = "ssh-ed25519 SHA256:lba/GXO1RlqaWyX9VvHRpDLLK8egub1dOScGWr7PHPU"
	bootAuthLine    = "Authenticated to georgenijo.com (via publickey)."
)
