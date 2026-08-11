# ssh-tui

The real thing behind `ssh ssh.georgenijo.com` — a Go SSH server ([Wish](https://github.com/charmbracelet/wish)) that drops visitors into a [Bubbletea](https://github.com/charmbracelet/bubbletea) TUI. The website (`../index.html`) is a simulation of this; this is the actual server.

## Build

```sh
CGO_ENABLED=0 go build -ldflags="-s -w" -o ssh-tui .
```

## Run

```sh
./ssh-tui -addr :23231 -hostkey /path/to/ssh_host_ed25519
```

The host key is generated on first run if missing. Never commit it.

## Deployment (opti)

The apex is a Cloudflare-proxied tunnel now, and that proxy is HTTP-only — raw SSH
can't ride it. The public entry point is `ssh.georgenijo.com`, a dns-only A record.

- Binary + host key live in `/home/george/ssh-tui/`, run as `george` via the system unit `/etc/systemd/system/ssh-tui.service` (MemoryMax=256M, Restart=always).
- `ssh.georgenijo.com` resolves to the home IP; the router forwards external 22 → opti:23231.
- Host key fingerprint is published in the boot sequence on https://georgenijo.com — it survived the move off the Oracle box, so it stays truthful.

The old Oracle box redirected public :22 → :23231 with an iptables PREROUTING rule on
`ens3` and ran as a systemd *user* unit under `ubuntu`; both are retired.

## Tests

```sh
go test ./...
```

`render_test.go` checks view rendering; `e2e_test.go` drives a real SSH session against the server.
