# ssh-tui

The real thing behind `ssh georgenijo.com` — a Go SSH server ([Wish](https://github.com/charmbracelet/wish)) that drops visitors into a [Bubbletea](https://github.com/charmbracelet/bubbletea) TUI. The website (`../index.html`) is a simulation of this; this is the actual server.

## Build

```sh
CGO_ENABLED=0 go build -ldflags="-s -w" -o ssh-tui .
```

## Run

```sh
./ssh-tui -addr :23231 -hostkey /path/to/ssh_host_ed25519
```

The host key is generated on first run if missing. Never commit it.

## Deployment (Oracle box)

- Binary + host key live in `~/ssh-tui/`, run as the `ubuntu` user via a systemd user unit (`ssh-tui.service`, MemoryMax=256M, lingering enabled).
- Public port 22 is redirected to 23231 by an iptables PREROUTING rule scoped to `ens3`, so tailnet admin SSH is unaffected.
- Host key fingerprint is published in the boot sequence on https://georgenijo.com.

## Tests

```sh
go test ./...
```

`render_test.go` checks view rendering; `e2e_test.go` drives a real SSH session against the server.
