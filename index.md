# George Nijo — the factory that builds the software

Software engineer in Boston. Snapshot 2026-10-29 (UTC).

I turned agent-orchestrated delivery into a repeatable system: explicit model routing, fanned-out subagents, one independent cross-model review, and verification against the running build. The output is measurable — 7,691 GitHub contributions and 1,237 pull requests across 19 repositories in 2026 so far, against 129 contributions in 2025.

The [animated homepage](https://georgenijo.com/) presents this snapshot. The [paper dossier](https://georgenijo.com/dossier.html) preserves the September 28, 2026 edition. The [terminal version](https://georgenijo.com/terminal.html) is also available.

## By the numbers

| Metric | Value | Scope |
|---|---:|---|
| Monthly contribution rate | 79× | 2025 average 10.75/month vs. Feb–Oct 2026 average 850.8/month; October is partial. |
| GitHub contributions in 2026 | 7,691 | Commits, PRs, reviews, and issues; 13,012 automated Lab snapshot commits excluded. |
| Pull requests opened in 2026 | 1,237 | Opened, not merged; 3 in 2025. |
| Repositories receiving commits | 19 | 2 in 2025; 85 owned repositories on GitHub. |
| Lines changed, trailing 12 months | ~2.97M | ~2.45M added / ~526k deleted across 24 local owned repositories; author-matched non-merge commits on available default branches. |
| Tokens in retained usage logs | 28.5B | Four Fleet machines; log cleanup can reduce the observed floor. |
| murmur-app GitHub releases | 122 | Latest v0.52.2, published 2026-10-23. |
| Personal skills | 75 | Collecting workstation; 5 hook events. |

## Clean contributions per month

GitHub contribution-calendar activity, including private activity, minus automated Lab commits grouped by author date. October 2026 is partial, through October 29 (UTC).

| Month | Contributions |
|---|---:|
| 2025-01 | 0 |
| 2025-02 | 0 |
| 2025-03 | 27 |
| 2025-04 | 0 |
| 2025-05 | 0 |
| 2025-06 | 0 |
| 2025-07 | 46 |
| 2025-08 | 1 |
| 2025-09 | 31 |
| 2025-10 | 6 |
| 2025-11 | 10 |
| 2025-12 | 8 |
| 2026-01 | 34 |
| 2026-02 | 230 |
| 2026-03 | 578 |
| 2026-04 | 460 |
| 2026-05 | 633 |
| 2026-06 | 1,177 |
| 2026-07 | 1,338 |
| 2026-08 | 1,143 |
| 2026-09 | 1,027 |
| 2026-10 | 1,071 (partial) |

## Tokens by model

Retained-log snapshot 2026-10-29; top five models.

| Model | Tokens |
|---|---:|
| GPT-5.6 Sol | 8,998,670,238 |
| Opus 5.5 | 5,702,650,463 |
| GPT-6 Astra | 3,584,293,131 |
| Opus 4.8 | 2,375,699,238 |
| GPT-6.1 Sol | 1,961,920,727 |

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
- **Burn log** — Tracks what the agents cost against what they shipped. A collector walks usage data and git history across every repo and machine, aggregates through Fleet, and publishes a daily record correlating tokens spent to commits landed. Python pipeline + static page, 28.5B tokens in retained logs. Public — [georgenijo.com/burn](https://georgenijo.com/burn.html).
- **ssh-tui** — `ssh ssh.georgenijo.com` opens an interactive terminal. A Wish + Bubbletea server on the gateway mirrors the terminal site and embeds the same burn snapshot. Go. Public endpoint.
- **Claude Code skills and hooks** — The agent loop itself, customized until it fits the doctrine. 75 personal skills covering repo cleanup, TDD, blast-radius analysis, decision logging, artifact publishing, and video generation, plus 5 hook events. Markdown, Python, shell. Private.
- **agentos** — Collapses eleven scattered repos into one self-hosted agent OS. A single backend, one secrets ledger, and a `just`-driven dev loop unifying Claude Code, the Cursor bridge, Discord bots, Google and GitHub glue, the mesh nodes, and Home Assistant. Everything external is mocked in tests. Python, ~96k lines changed in 12 months, 196 commits. Private.
- **agent-mesh** — A shared nervous system for agents that would otherwise collide. A local-first coordination fabric letting Claude Code, Codex CLI, Cursor CLI, and Aider discover each other, announce what they are touching, and read a shared blackboard of decisions. Go stdlib only. ~58k lines changed in 12 months, 209 commits. Public — [github.com/georgenijo/agent-mesh](https://github.com/georgenijo/agent-mesh).
- **usher** — The MCP broker, one front desk every agent talks to. Routes, trims, arbitrates, gates, and audits MCP traffic so agents face a single well-behaved surface. Context trimming is the load-bearing feature. Go. Public — [github.com/georgenijo/usher](https://github.com/georgenijo/usher).
- **hangar** — Multi-session control plane for Claude Code, Codex, and friends. Runs and supervises many agent sessions at once, so parallel fan-out is an operation rather than a pile of terminal tabs. Rust. Public — [github.com/georgenijo/hangar](https://github.com/georgenijo/hangar).
- **ghosthands** — Free, local computer-use for macOS, no cursor stolen. A small MLX model drives native macOS apps through the Accessibility tree, then replays the recorded flow with no model in the loop. Python, ~32k lines changed in 12 months. Public — [github.com/georgenijo/ghosthands](https://github.com/georgenijo/ghosthands).

## Products and selected projects

- **murmur-app** — Local voice-to-text dictation for macOS, nothing leaves the machine. Tauri 2 with a Rust core (whisper-rs on the Metal GPU, cpal audio capture, global hotkey, clipboard paste) and a React/TypeScript front end. Rust + React/TypeScript, ~222k LOC, 1,358 commits, 122 GitHub releases; latest v0.52.2. Public — [github.com/georgenijo/murmur-app](https://github.com/georgenijo/murmur-app).
- **aperture** — iOS film camera with a real vintage emulation pipeline. An AVFoundation capture engine feeding a Core Image and Metal chain that emulates film stocks rather than applying a filter. Swift. Public — [github.com/georgenijo/aperture](https://github.com/georgenijo/aperture).
- **whoop-dashboard** — Personal health analytics with an AI coach on top. Pulls Whoop data into a dashboard that interprets rather than just plots, with a companion Oura version. TypeScript, ~160k lines changed in 12 months, 282 commits. Public — [github.com/georgenijo/whoop-dashboard](https://github.com/georgenijo/whoop-dashboard).
- **St. Basil's Boston** — Church website rebuild, run as an agent ticket workflow. Next.js 14 App Router with Sanity CMS, Supabase, and Resend. Work enters as tickets and is executed by agents against acceptance checks. The oldest thread here: first commit March 2025, before any of this. TypeScript, ~86k lines changed in 12 months, 279 commits. Private repo, live site.
- **fleetmap** — Live map of what your Mac is actually running. A relationship-aware process and connection monitor: processes sized by RAM, colored by CPU, wired together by their live sockets. Swift + Go. Public — [github.com/georgenijo/fleetmap](https://github.com/georgenijo/fleetmap).
- **Gauge** — Claude usage in the menu bar, and nothing else. A deliberately minimal macOS menu bar app showing consumption at a glance. Swift. Public — [github.com/georgenijo/Gauge](https://github.com/georgenijo/Gauge).

## How to read these numbers

These are dated observations. GitHub contributions include commits, pull requests, reviews, and issues, with private activity included. Automated Lab snapshot commits are excluded using site main history. Lines changed covers locally available default branches, excludes duplicate worktrees, merges, and automated Lab snapshots, and matches George's author identity. It is a floor. Token counts cover logs currently retained on four machines, and can fall when older transcripts are removed; they are not a lifetime counter.

Exact source timestamps and methodology: [site-metrics.json](/data/site-metrics.json). Usage data: [burn.json](/burn.json). Project metadata: [lab-projects.json](/data/lab-projects.json). Fleet health: [fleet-health.json](/data/fleet-health.json).

## Where to look next

- [Homepage](https://georgenijo.com/)
- [Paper dossier — September 28 edition](https://georgenijo.com/dossier.html)
- [Burn log](https://georgenijo.com/burn.html)
- [Contact](https://georgenijo.com/contact.md)
