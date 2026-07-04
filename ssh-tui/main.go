// ssh-tui is the real thing georgenijo.com pretends to be: George Nijo's
// portfolio served over SSH with Wish + Bubbletea + Lipgloss.
package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	bm "github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"
	gossh "golang.org/x/crypto/ssh"
)

// idleLimit is how long a session may sit with no keyboard input before it
// is disconnected (enforced in the model; a transport deadline backstops it).
const idleLimit = 20 * time.Minute

func newServer(addr, hostKey string) (*ssh.Server, error) {
	return wish.NewServer(
		wish.WithAddress(addr),
		wish.WithHostKeyPath(hostKey),
		// Transport-level backstop; input idleness is enforced in the model
		// because the ticking UI keeps the connection byte-active.
		wish.WithIdleTimeout(idleLimit+5*time.Minute),

		// Public portfolio: accept absolutely everyone, however they knock.
		wish.WithPublicKeyAuth(func(ssh.Context, ssh.PublicKey) bool { return true }),
		wish.WithPasswordAuth(func(ssh.Context, string) bool { return true }),
		wish.WithKeyboardInteractiveAuth(func(ssh.Context, gossh.KeyboardInteractiveChallenge) bool { return true }),

		// Middleware list is innermost-first: logging runs first on connect,
		// then activeterm gates non-PTY sessions, then the MOTD prints, then
		// Bubbletea takes over; the goodbye runs after the program exits.
		wish.WithMiddleware(
			goodbyeMiddleware(),
			bm.Middleware(teaHandler),
			motdMiddleware(),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
}

func main() {
	addr := flag.String("addr", ":23231", "SSH listen address")
	hostKey := flag.String("hostkey", "./ssh_host_ed25519", "path to the SSH host key (generated on first run)")
	flag.Parse()

	srv, err := newServer(*addr, *hostKey)
	if err != nil {
		log.Fatal("could not create server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info("starting SSH server", "addr", *addr, "hostkey", *hostKey)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("server error", "error", err)
			done <- syscall.SIGTERM
		}
	}()

	<-done
	log.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("shutdown error", "error", err)
	}
}

// teaHandler builds a fresh model per session — no shared mutable state.
func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()
	w, h := pty.Window.Width, pty.Window.Height
	if w <= 0 || h <= 0 {
		// Degenerate clients (e.g. piped ssh -tt) report 0x0; assume a
		// classic terminal instead of rendering nothing.
		w, h = 80, 24
	}
	m := newModel(newStyles(sessionRenderer(s)), w, h)
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

type sshEnviron []string

func (e sshEnviron) Environ() []string { return e }
func (e sshEnviron) Getenv(k string) string {
	for _, v := range e {
		if strings.HasPrefix(v, k+"=") {
			return v[len(k)+1:]
		}
	}
	return ""
}

// sessionRenderer builds a lipgloss renderer from the session's environment
// (TERM, and COLORTERM when the client forwards it) without ever querying the
// terminal. bubbletea.MakeRenderer probes the client with OSC/DA1 escape
// queries and reads the session's input stream for the answer, which both
// delays startup and can swallow early keystrokes; deriving the profile from
// the environment is deterministic: truecolor terminals that advertise it get
// 24-bit color, everything else falls back to 256 colors (or less).
func sessionRenderer(s ssh.Session) *lipgloss.Renderer {
	pty, _, ok := s.Pty()
	if !ok || pty.Term == "" || pty.Term == "dumb" {
		return lipgloss.NewRenderer(s, termenv.WithProfile(termenv.Ascii))
	}
	env := sshEnviron(append(s.Environ(), "TERM="+pty.Term))
	r := lipgloss.NewRenderer(s,
		termenv.WithEnvironment(env),
		termenv.WithUnsafe(), // trust the env; the session writer is not a local TTY
		termenv.WithColorCache(true),
	)
	r.SetHasDarkBackground(true) // charm-dark theme; never query the client
	return r
}

func crlf(s string) string {
	return strings.ReplaceAll(s, "\n", "\r\n")
}

// motdMiddleware prints the gradient wordmark before the TUI starts, so it
// also lands in the scrollback after the alt screen is torn down on exit.
func motdMiddleware() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			if pty, _, ok := s.Pty(); ok {
				st := newStyles(sessionRenderer(s))
				var b strings.Builder
				b.WriteString("\n")
				if pty.Window.Width >= bannerWidth+2 {
					b.WriteString(st.banner)
				} else {
					b.WriteString(st.pinkBold.Render("GEORGE NIJO"))
				}
				b.WriteString("\n" + st.bannerTag + "\n")
				b.WriteString(st.dim.Render("Last login: "+time.Now().UTC().Format("Mon Jan _2 15:04:05 2006")+" from the internet") + "\n")
				wish.WriteString(s, crlf(b.String()))
				time.Sleep(600 * time.Millisecond)
			}
			next(s)
		}
	}
}

// goodbyeMiddleware runs after the Bubbletea program exits.
func goodbyeMiddleware() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			wish.WriteString(s, "Connection to "+hostName+" closed.\r\n")
			next(s)
		}
	}
}
