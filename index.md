# George Nijo

Software engineer · Boston · builds the factory that builds the software

I turned agent-orchestrated delivery into a repeatable system: explicit model routing, fanned-out subagents, one independent cross-model review, and verification against the running build. The output is measurable — 6,300 commits and 1,105 pull requests across 18 repos so far in 2026, against 129 commits in all of 2025.

- Email: [george.nijo8@gmail.com](mailto:george.nijo8@gmail.com)
- GitHub: [github.com/georgenijo](https://github.com/georgenijo)
- LinkedIn: [linkedin.com/in/georgenijo](https://www.linkedin.com/in/georgenijo)
- Terminal: `ssh georgenijo.com` — a real endpoint, a Go Wish/Bubbletea TUI mirror of this site

This is a snapshot document dated 2026-09-16. The terminal-styled version of this site lives at [/terminal.html](https://georgenijo.com/terminal.html).

## By the numbers

| Value | Measure | Qualifier |
|---|---|---|
| 73× | Monthly commit rate, after vs. before | 2025 average 10.8 commits/mo vs. Feb–Sep 2026 average 783/mo. (6,266 ÷ 8) ÷ (129 ÷ 12) ≈ 72.9. Bot commits excluded. |
| 6,300 | Commits in 2026 so far | Jan 1 – Sep 16, 2026. Excludes 7,609 automated Fleet health-snapshot commits. |
| 3 → 1,105 | Pull requests opened, 2025 → 2026 | PRs opened, not merged. 2026 figure is through Sep 16. |
| 2 → 18 | Repos with commits, 2025 → 2026 | 70 repositories exist on GitHub; these are the ones actually receiving work. |
| ~1.06M | Lines changed, trailing 12 months | Floor, not a total: ~925k added / ~134k deleted, measured across 14 locally cloned repos only. |
| 6.2B | AI tokens burned | Burn-log snapshot dated 2026-07-17, aggregated across machines via Fleet. Scope-dependent; treat as a floor. |
| 101 | GitHub releases of one app | murmur-app, v0.1.0 → v0.50.0, latest 2026-09-14. Roughly one release every 1–2 days in Sep 2026. |
| 77 | Personal agent skills in use | Installed skills in `~/.claude/skills`, plus 5 hook events wired into the agent loop. |

## Ship rate

Clean commits per month. Bot commits excluded. Sep 2026 is partial (through the 16th).

| Month | Commits | Month | Commits |
|---|---|---|---|
| 2025-01 | 0 | 2026-01 | 34 |
| 2025-02 | 0 | 2026-02 | 230 |
| 2025-03 | 27 | 2026-03 | 578 |
| 2025-04 | 0 | 2026-04 | 460 |
| 2025-05 | 0 | 2026-05 | 633 |
| 2025-06 | 0 | 2026-06 | 1,177 |
| 2025-07 | 46 | 2026-07 | 1,338 |
| 2025-08 | 1 | 2026-08 | 1,142 |
| 2025-09 | 31 | 2026-09 | 708 (partial) |
| 2025-10 | 6 | | |
| 2025-11 | 10 | | |
| 2025-12 | 8 | | |

**Feb 2026 — the method changed, not the hours.** This is where I stopped typing code as the primary act and started running agents against a written doctrine: fan work out to subagents, route each task to a model chosen on difficulty, require one independent review before anything merges. Commits went from 34 in January to 230 in February and have not returned to the old baseline since.

## Model mix

Tokens by model, burn-log snapshot 2026-07-17.

| Model | Tokens |
|---|---|
| Opus 4.8 | 2,510,416,238 |
| Fable 5 | 1,062,481,134 |
| Sonnet 5 | 911,112,836 |
| GPT-5.5 | 660,827,177 |
| GPT-5.6 Sol | 434,067,143 |

Top 5 models by tokens, from the burn log at [georgenijo.com/burn](https://georgenijo.com/burn.html). Opus 4.8 is ~40% of the total — the hard reasoning goes to the expensive model, the mechanical work does not.

## The factory

Six rules. They are written down, they are enforced by config, and they are why the chart above looks like that.

1. **The orchestrator never leaves the chat.** The model I am talking to plans, routes, and reviews. It does not disappear into a forty-minute solo implementation. Context stays where decisions get made.
2. **Research and implementation get fanned out.** Anything that means reading across many files goes to a subagent that returns a conclusion, not a file dump. Independent workstreams run in parallel with separated edit scopes.
3. **Model choice is a routing decision, not a default.** Hardest architecture, ambiguity, and reviews go to GPT-6 Astra. Lead implementation and debugging go to Sol. Ordinary parallel work goes to Terra. Narrow mechanical volume goes to Luna. Claude Opus for hard Claude-native reasoning, Sonnet for ordinary, Haiku for narrow. Labels are literal; nothing gets relabeled to look better.
4. **One independent review, different model, fresh context, high effort.** Every meaningful behavior change gets reviewed by a model that did not write it, given the intended behavior, the diff, and the verification evidence — and nothing else. Read-only. If that review is unavailable, I report the missing gate rather than quietly substituting a weaker one.
5. **Verification is risk-based and happens against the running build.** Reproduce the bug first where feasible. Exercise the affected user flow in a real browser or a real binary. Confirm the artifact actually contains the change. Tests are evidence, not the whole case.
6. **"Fix this" means finished, not started.** A repository fix request authorizes the whole chain: implement, verify, review, commit, push, PR ready to merge, CI resolved, conflicts handled — without a second prompt. The inverse rule matters as much: a question authorizes nothing. Asking whether something could work never edits a file.

## The shop floor

One laptop and three servers on a private mesh. Which machine a job runs on is decided before anything else about it.

- **laptop · control** — Where I work. It plans, routes, reviews, and hands work out, and it is the only machine that can decrypt a secret.
- **always-on worker** — Runs headless agent sessions. Work started here keeps going after the laptop lid closes.
- **gateway · production** — Serves this page and runs the monitor that watches the other machines. Also the way in when a direct route fails.
- **workloads** — Hosts artifacts, dev servers, and anything long-running that has no business on a laptop.

1. **No agent ever types an address.** One manifest records who each machine is; the mesh supplies live addressing. An agent names the machine and something else resolves it. Addresses change, names do not.
2. **Work outlives the session that started it.** Long jobs go to a server rather than the laptop. Closing the lid does not kill them, and the result can be picked up from any machine.
3. **Every machine speaks the same language.** One skill bundle is linked into both agent runtimes on all four machines, so a session started anywhere has the same vocabulary and the same working agreements.
4. **Handoffs are artifacts, not pasted text.** A finished result is published once and gets an immutable, checksummed ID and a private URL. The next agent, or the next person, consumes that ID instead of a copy.
5. **The mesh watches itself, within limits.** A monitor on the gateway checks the workers, requires two independent observations before it acts, and is authorized to take exactly one action. It cannot widen its own permissions.

This page is the proof: it is served from the gateway, and the token counts above were collected from all four machines.

## Tools I built to make that possible

- **Fleet** — Agent-native control plane for a personal Tailscale mesh. One manifest resolves who every box is; Tailscale supplies live IPs and online state, so neither I nor an agent ever hardcodes an address or guesses an SSH user. Covers discovery, remote exec, file transfer, service management, durable artifact publishing, shared skill distribution, and token-burn analytics. Python, ~9.5k LOC, 28 subcommands. Private.
- **Family Host** — A personal Vercel, invite-only, running on my own hardware. Push a GitHub repo or container image and get an isolated HTTPS app on a real subdomain, plus persistent browser-based coding workspaces and managed Postgres. Systemd guard/monitor/backup timers, deploy locks, rollback, restore verification, secret scanning, synthetic checks. Python server + Node CLI, ~37k LOC, 169 commits. Private.
- **CPA — personal AI gateway** — Every model provider behind one Anthropic-compatible API. A reverse proxy fronting Anthropic Opus/Sonnet/Fable and OpenAI GPT-6 Astra and GPT-5.6 Sol/Terra/Luna behind a single endpoint, with per-client metering, quotas, and dashboards. Python, ~19k LOC, 77 commits plus six extension modules. Private.
- **Burn log** — Tracks what the agents cost against what they shipped. A collector walks usage data and git history across every repo and machine, aggregates through Fleet, and publishes a daily record correlating tokens spent to commits landed. Python pipeline + static page, 6.2B tokens recorded. Public — [georgenijo.com/burn](https://georgenijo.com/burn.html).
- **ssh-tui** — `ssh georgenijo.com` returns an interactive terminal, not a banner. A Wish + Bubbletea server on an Oracle Cloud box serving a full TUI mirror of this site on port 22. Go, ~2,000 LOC. Public endpoint.
- **Claude Code skills and hooks** — The agent loop itself, customized until it fits the doctrine. 77 personal skills covering repo cleanup, TDD, blast-radius analysis, decision logging, artifact publishing, and video generation, plus 5 hook events. Markdown, Python, shell. Private.
- **agentos** — Collapses eleven scattered repos into one self-hosted agent OS. A single backend, one secrets ledger, and a `just`-driven dev loop unifying Claude Code, the Cursor bridge, Discord bots, Google and GitHub glue, the mesh nodes, and Home Assistant. Everything external is mocked in tests. Python, ~96k lines changed in 12 months, 196 commits. Private.
- **agent-mesh** — A shared nervous system for agents that would otherwise collide. A local-first coordination fabric letting Claude Code, Codex CLI, Cursor CLI, and Aider discover each other, announce what they are touching, and read a shared blackboard of decisions. Go stdlib only. ~58k lines changed in 12 months, 209 commits. Public — [github.com/georgenijo/agent-mesh](https://github.com/georgenijo/agent-mesh).
- **usher** — The MCP broker, one front desk every agent talks to. Routes, trims, arbitrates, gates, and audits MCP traffic so agents face a single well-behaved surface. Context trimming is the load-bearing feature. Go. Public — [github.com/georgenijo/usher](https://github.com/georgenijo/usher).
- **hangar** — Multi-session control plane for Claude Code, Codex, and friends. Runs and supervises many agent sessions at once, so parallel fan-out is an operation rather than a pile of terminal tabs. Rust. Public — [github.com/georgenijo/hangar](https://github.com/georgenijo/hangar).
- **ghosthands** — Free, local computer-use for macOS, no cursor stolen. A small MLX model drives native macOS apps through the Accessibility tree, then replays the recorded flow with no model in the loop. Python, ~32k lines changed in 12 months. Public — [github.com/georgenijo/ghosthands](https://github.com/georgenijo/ghosthands).

## Products and selected projects

- **murmur-app** — Local voice-to-text dictation for macOS, nothing leaves the machine. Tauri 2 with a Rust core (whisper-rs on the Metal GPU, cpal audio capture, global hotkey, clipboard paste) and a React/TypeScript front end. Rust + React/TypeScript, ~222k LOC, 1,358 commits, 101 GitHub releases (103 git tags). Public — [github.com/georgenijo/murmur-app](https://github.com/georgenijo/murmur-app).
- **aperture** — iOS film camera with a real vintage emulation pipeline. An AVFoundation capture engine feeding a Core Image and Metal chain that emulates film stocks rather than applying a filter. Swift. Public — [github.com/georgenijo/aperture](https://github.com/georgenijo/aperture).
- **whoop-dashboard** — Personal health analytics with an AI coach on top. Pulls Whoop data into a dashboard that interprets rather than just plots, with a companion Oura version. TypeScript, ~160k lines changed in 12 months, 282 commits. Public — [github.com/georgenijo/whoop-dashboard](https://github.com/georgenijo/whoop-dashboard).
- **St. Basil's Boston** — Church website rebuild, run as an agent ticket workflow. Next.js 14 App Router with Sanity CMS, Supabase, and Resend. Work enters as tickets and is executed by agents against acceptance checks. The oldest thread here: first commit March 2025, before any of this. TypeScript, ~86k lines changed in 12 months, 279 commits. Private repo, live site.
- **fleetmap** — Live map of what your Mac is actually running. A relationship-aware process and connection monitor: processes sized by RAM, colored by CPU, wired together by their live sockets. Swift + Go. Public — [github.com/georgenijo/fleetmap](https://github.com/georgenijo/fleetmap).
- **Gauge** — Claude usage in the menu bar, and nothing else. A deliberately minimal macOS menu bar app showing consumption at a glance. Swift. Public — [github.com/georgenijo/Gauge](https://github.com/georgenijo/Gauge).

## Timeline

| Date | Headline | Detail |
|---|---|---|
| 2025-03 | First commit on St. Basil's | 27 commits in March, then nothing for three months. The old baseline: real projects, sporadic output. |
| 2025-07 – 2025-12 | Sparse and manual | 46, 1, 31, 6, 10, 8 commits across six months. Two repos received any work at all in the whole of 2025. |
| 2026-01 | murmur-app starts | 34 commits. Still hand-written, still one project at a time. |
| 2026-02 | The inflection | 230 commits. Agent workflows adopted as a standing method: written working agreements, subagent fan-out, explicit model routing. |
| 2026-03 | Tooling for agents begins | aperture starts. 578 commits. |
| 2026-04 | Control planes | hangar and whoop-dashboard begin. 460 commits, and the PR count starts climbing toward four figures. |
| 2026-06 | Peak infrastructure month | agent-mesh, ghosthands, Gauge, usher, fleetmap, agentos, Fleet, and Family Host all start within four weeks. 1,177 commits. |
| 2026-07 | The factory measures itself | This site and the burn log go up; Fleet starts auto-publishing health snapshots to it. 1,338 commits, the highest month so far. |
| 2026-09 | Steady state | murmur-app reaches v0.50.0 on the 14th. 708 commits through the 16th, with half the month left. |

## How to read these numbers

Commit counts exclude 7,609 automated Fleet health-snapshot commits, which would otherwise inflate 2026 by more than double. Lines changed is a floor: it covers only the 14 repositories cloned locally, not all 70 on GitHub. September 2026 is partial, through the 16th. Pull request counts are PRs opened, not merged. The token total is a burn-log snapshot dated 2026-07-17 and depends on which machines' usage logs were aggregated, so treat it as a lower bound. murmur-app has 101 GitHub releases (103 git tags); the two counts differ because not every tag has a release attached.

## Where to look next

- [Terminal site](https://georgenijo.com/terminal.html)
- [Burn log](https://georgenijo.com/burn.html)
- [About](https://georgenijo.com/about.md)
- [Projects](https://georgenijo.com/projects.md)
- [Contact](https://georgenijo.com/contact.md)
