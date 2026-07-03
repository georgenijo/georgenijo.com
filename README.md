# georgenijo.com

Personal site of George Nijo — software engineer building infrastructure for AI agents
(LLM routers, MCP brokers, agent control planes).

**Live:** https://georgenijo.com

The site is a browser simulation of SSHing into `georgenijo.com`: a typed connection
sequence, key exchange, and an ASCII MOTD, then a [Charm](https://charm.sh)-style
Bubbletea TUI — list navigation with arrows/j-k, `/` to filter projects, `q` to
disconnect. On wide screens it runs a master-detail split with a live preview pane.
Fully static, zero dependencies, one HTML file.

## The design process

Five concepts were prototyped before picking this one, each riffing on
"a familiar interface, made real" for an agent-infrastructure engineer.
All of them live on as working demos:

| Concept | Demo |
|---|---|
| GEORGE OS v1 — desktop OS with boot sequence, draggable windows, terminal | [/concepts/george-os-v1](https://georgenijo.com/concepts/george-os-v1.html) |
| GEORGE OS v2 — the OS as an agent control plane: Mission Control, live routing feed, visible agent handoffs | [/concepts/concept-1-os-v2](https://georgenijo.com/concepts/concept-1-os-v2.html) |
| Operations dashboard — NOC/Grafana panel wall; projects as monitored services, career as an alert log | [/concepts/concept-2-noc](https://georgenijo.com/concepts/concept-2-noc.html) |
| htop for agents — TUI process monitor; projects as killable, self-healing processes | [/concepts/concept-3-htop](https://georgenijo.com/concepts/concept-3-htop.html) |
| Approach control — ATC radar scope; projects as aircraft, flight strips as content | [/concepts/concept-4-atc](https://georgenijo.com/concepts/concept-4-atc.html) |
| ssh georgenijo.com — Charm-style TUI session **(chosen — this is the site)** | [/](https://georgenijo.com) |

The concepts came out of a research pass across four portfolio genres
(desktop-OS sites, terminal/ASCII sites, 3D/WebGL worlds, and
familiar-interface recreations). The through-line that survived: the
memorable ones make the interface *real* rather than skinning static
content — which is also the plan here.

## Roadmap

- **Phase 2 — the real thing:** an actual SSH server (Go, Charm Wish/Bubbletea)
  serving this same TUI at `ssh ssh.georgenijo.com`, running on the homelab.
- Live data in the TUI (GitHub activity, router telemetry).

## Development

Everything is hand-rolled HTML/CSS/JS in `index.html`. Deploys via GitHub Pages
on push to `main`; `CNAME` pins the custom domain.

🤖 Built with [Claude Code](https://claude.com/claude-code)
