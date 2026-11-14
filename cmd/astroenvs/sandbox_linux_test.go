//go:build linux

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The sandbox restarts the running binary with sandboxArg; in tests that binary is the test
// binary, so route it here the way main does
func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == sandboxArg {
		runSandbox(os.Args[2:])
		return
	}
	os.Exit(m.Run())
}

// TestSandbox runs a shell inside the sandbox that reports what it can see and do. It needs
// unprivileged user namespaces; set ASTROENVS_SANDBOX_TEST=1 to fail instead of skipping
// when they are missing (CI does)
func TestSandbox(t *testing.T) {
	if !sandboxSupported() {
		t.Skip("user namespaces disabled")
	}
	tmp := t.TempDir()
	mk := func(parts ...string) string {
		p := filepath.Join(append([]string{tmp}, parts...)...)
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
		return p
	}
	write := func(p, s string) {
		if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	homeDir := mk("home")
	shared := mk("home", "shared", "T5")
	venv := mk("home", ".venv")
	other := mk("home", "other")
	write(filepath.Join(shared, "notes.txt"), "shared")
	write(filepath.Join(venv, "lib.py"), "venv")
	write(filepath.Join(other, "secret.txt"), "secret")
	write(filepath.Join(homeDir, ".bashrc"), "secret")

	script := `
check() { if eval "$2" >/dev/null 2>&1; then echo "$1=yes"; else echo "$1=no"; fi; }
check read_shared   'cat "$SHARED/notes.txt"'
check write_shared  'touch "$SHARED/new.txt"'
check read_venv     'cat "$VENV/lib.py"'
check write_venv    'touch "$VENV/evil.py"'
check see_other     'ls "$OTHER"'
check see_bashrc    'cat "$HOME/.bashrc"'
check write_home    'touch "$HOME/scratch"'
check up_from_share 'ls "$SHARED/../../other/secret.txt"'
check read_etc      'cat /etc/passwd'
check write_etc     'touch /etc/evil'
check umount_share  'umount -l "$SHARED"'
check nested_rw     'unshare -rm sh -c "mount -o remount,bind,rw \"$VENV\" && touch \"$VENV/evil2.py\""'
check nested_reveal 'unshare -rm sh -c "umount -l \"$SHARED\"; umount -l /; ls \"$OTHER\""'
echo "capeff=$(awk '/^CapEff/ {print $2}' /proc/self/status)"
echo "pids=$(ls -d /proc/[0-9]* | wc -l)"
echo "tmp=$(ls -A /tmp | wc -l)"
`
	cmd := sandboxCommand(shared, []string{venv}, "/bin/sh", "-c", script)
	cmd.Env = append(os.Environ(), "HOME="+homeDir, "SHARED="+shared, "VENV="+venv, "OTHER="+other)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	if strings.Contains(out.String(), "sandbox:") && os.Getenv("ASTROENVS_SANDBOX_TEST") != "1" {
		t.Skipf("sandbox unavailable here: %s", strings.TrimSpace(out.String()))
	}
	if err != nil {
		t.Fatalf("sandbox failed: %v\n%s", err, out.String())
	}
	t.Logf("inside the sandbox:\n%s", out.String())

	got := map[string]string{}
	for _, line := range strings.Split(out.String(), "\n") {
		if k, v, ok := strings.Cut(strings.TrimSpace(line), "="); ok {
			got[k] = strings.TrimSpace(v)
		}
	}
	want := map[string]string{
		"read_shared": "yes", "write_shared": "yes", "read_venv": "yes", "write_venv": "no",
		"see_other": "no", "see_bashrc": "no", "write_home": "yes", "up_from_share": "no",
		"read_etc": "yes", "write_etc": "no", "umount_share": "no", "nested_rw": "no", "nested_reveal": "no",
		"capeff": "0000000000000000", "tmp": "0",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
	// The shell, the script's subshells and PID 1: nowhere near the host's process list
	if got["pids"] == "" || len(got["pids"]) > 1 {
		t.Errorf("pids = %q, want a handful", got["pids"])
	}
	if _, err := os.Stat(filepath.Join(shared, "new.txt")); err != nil {
		t.Errorf("a file written in the shared folder isn't on the host: %v", err)
	}
	for _, p := range []string{filepath.Join(venv, "evil.py"), filepath.Join(venv, "evil2.py"), filepath.Join(homeDir, "scratch")} {
		if _, err := os.Stat(p); err == nil {
			t.Errorf("%s reached the host", p)
		}
	}
}
