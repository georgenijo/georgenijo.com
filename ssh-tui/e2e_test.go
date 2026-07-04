package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"net"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	gossh "golang.org/x/crypto/ssh"
)

// TestEndToEnd starts the real Wish server, connects with an SSH client,
// requests a PTY, verifies the MOTD/menu bytes, presses 'q', and checks that
// the session closes with the goodbye message.
func TestEndToEnd(t *testing.T) {
	dir := t.TempDir()
	srv, err := newServer("127.0.0.1:0", filepath.Join(dir, "hostkey"))
	if err != nil {
		t.Fatal(err)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve(ln) //nolint:errcheck
	defer srv.Close()

	conf := &gossh.ClientConfig{
		User:            "guest",
		Auth:            []gossh.AuthMethod{gossh.Password("anything")},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	client, err := gossh.Dial("tcp", ln.Addr().String(), conf)
	if err != nil {
		t.Fatalf("dial (password auth should always be accepted): %v", err)
	}
	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	var mu sync.Mutex
	var out bytes.Buffer
	stdout := writerFunc(func(p []byte) (int, error) {
		mu.Lock()
		defer mu.Unlock()
		return out.Write(p)
	})
	sess.Stdout = stdout
	sess.Stderr = stdout
	stdin, err := sess.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := sess.RequestPty("xterm-256color", 35, 120, gossh.TerminalModes{}); err != nil {
		t.Fatal(err)
	}
	if err := sess.Shell(); err != nil {
		t.Fatal(err)
	}

	// snapshot returns the ANSI-stripped output seen so far; the styling
	// wraps every gradient rune in its own escape sequence, so plain-text
	// matching needs the escapes removed.
	snapshot := func() string {
		mu.Lock()
		defer mu.Unlock()
		return ansi.Strip(out.String())
	}
	waitFor := func(what string, timeout time.Duration) {
		t.Helper()
		deadline := time.Now().Add(timeout)
		for time.Now().Before(deadline) {
			if strings.Contains(snapshot(), what) {
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
		t.Fatalf("timed out waiting for %q; got:\n%s", what, snapshot())
	}

	// MOTD masthead and menu must show up.
	waitFor("GEORGE NIJO", 10*time.Second)           // masthead wordmark
	waitFor("agent infrastructure", 10*time.Second)  // masthead tagline
	waitFor("who's at the keyboard", 10*time.Second) // menu row

	// Navigate: j (down to projects), enter — expect the projects list.
	if _, err := stdin.Write([]byte("j")); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if _, err := stdin.Write([]byte("\r")); err != nil {
		t.Fatal(err)
	}
	waitFor("ledger", 5*time.Second) // projects breadcrumb
	waitFor("nadirclaw", 5*time.Second)

	// q must disconnect cleanly with the goodbye line.
	if _, err := stdin.Write([]byte("q")); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- sess.Wait() }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatalf("session did not close after q; got:\n%s", snapshot())
	}
	if !strings.Contains(snapshot(), "Connection to georgenijo.com closed.") {
		t.Fatalf("missing goodbye message; got tail:\n%s", tail(snapshot(), 600))
	}
}

// TestAnonymousAuthVariants ensures every auth flavor is accepted.
func TestAnonymousAuthVariants(t *testing.T) {
	dir := t.TempDir()
	srv, err := newServer("127.0.0.1:0", filepath.Join(dir, "hostkey"))
	if err != nil {
		t.Fatal(err)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve(ln) //nolint:errcheck
	defer srv.Close()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	key, err := gossh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string][]gossh.AuthMethod{
		"publickey": {gossh.PublicKeys(key)},
		"password":  {gossh.Password("")},
		"keyboard-interactive": {gossh.KeyboardInteractive(
			func(string, string, []string, []bool) ([]string, error) { return nil, nil },
		)},
	}
	for name, auth := range cases {
		conf := &gossh.ClientConfig{
			User:            "visitor",
			Auth:            auth,
			HostKeyCallback: gossh.InsecureIgnoreHostKey(),
			Timeout:         5 * time.Second,
		}
		c, err := gossh.Dial("tcp", ln.Addr().String(), conf)
		if err != nil {
			t.Fatalf("%s auth rejected: %v", name, err)
		}
		c.Close()
	}
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }
