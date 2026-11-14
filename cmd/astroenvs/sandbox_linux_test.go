//go:build linux

package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
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
echo "tmp=$(ls -A /tmp | grep -vx "$TMPTOP" | wc -l)"
`
	cmd := sandboxCommand(shared, []string{venv}, "/bin/sh", "-c", script)
	// The test's folders live in /tmp, so the shared folder's path shows up in the sandbox's /tmp
	tmpTop, _, _ := strings.Cut(strings.TrimPrefix(tmp, "/tmp/"), "/")
	cmd.Env = append(os.Environ(), "HOME="+homeDir, "SHARED="+shared, "VENV="+venv, "OTHER="+other, "TMPTOP="+tmpTop)
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

// terminalClient opens a Jupyter terminal over its websocket, types a command and prints what
// the terminal shows, up to the END-OF-2TEST marker (computed, so the echoed command line
// doesn't count). tornado comes with Jupyter
const terminalClient = `
import json, sys, urllib.request
from tornado.ioloop import IOLoop
from tornado.websocket import websocket_connect
base, token, cmd = sys.argv[1:4]
req = urllib.request.Request(f"http://{base}/api/terminals?token={token}", data=b"", method="POST")
name = json.load(urllib.request.urlopen(req))["name"]
async def main():
    ws = await websocket_connect(f"ws://{base}/terminals/websocket/{name}?token={token}")
    await ws.write_message(json.dumps(["stdin", cmd + "; echo END-OF-$((1+1))TEST\r"]))
    out = ""
    while "END-OF-2TEST" not in out:
        msg = await ws.read_message()
        if msg is None:
            break
        kind, *data = json.loads(msg)
        if kind == "stdout":
            out += data[0]
    print(out)
IOLoop.current().run_sync(main, timeout=60)
`

// TestSandboxJupyter starts a real sandboxed JupyterLab the way startJupyter does and tries to
// leave its folder from a Jupyter terminal. It needs a venv with JupyterLab, made with uv like
// ~/.ipynb: ASTROENVS_TEST_VENV=/path/to/venv (CI sets it)
func TestSandboxJupyter(t *testing.T) {
	venv := os.Getenv("ASTROENVS_TEST_VENV")
	if venv == "" {
		t.Skip("ASTROENVS_TEST_VENV not set")
	}
	tmp := t.TempDir()
	homeDir := filepath.Join(tmp, "home")
	shared := filepath.Join(homeDir, "shared", "T5")
	other := filepath.Join(homeDir, "other")
	for _, d := range []string{shared, other} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile(filepath.Join(shared, "notes.txt"), []byte("shared"), 0o644)
	os.WriteFile(filepath.Join(other, "secret.txt"), []byte("TOPSECRET"), 0o644)
	os.WriteFile(filepath.Join(homeDir, ".bashrc"), []byte("TOPSECRET"), 0o644)

	cfg := &Config{IpynbVenv: venv, RunDir: filepath.Join(tmp, "run")}
	os.MkdirAll(cfg.RunDir, 0o700)
	ro := sandboxReadOnly(cfg)
	t.Logf("read-only: %v", ro)

	const token = "test-token-0123456789abcdef"
	port := freePort(18888, nil)
	bin := filepath.Join(venv, "bin")
	cmd := sandboxCommand(shared, ro, filepath.Join(bin, "jupyter"), "lab", "--no-browser",
		"--ip=127.0.0.1", fmt.Sprintf("--port=%d", port), "--ServerApp.port_retries=0",
		"--ServerApp.root_dir="+shared)
	env := []string{"PATH=" + bin + ":" + os.Getenv("PATH"), "VIRTUAL_ENV=" + venv,
		"JUPYTER_TOKEN=" + token, "HOME=" + homeDir, "OTHER=" + other}
	logPath := filepath.Join(tmp, "jupyter.log")
	cmd, err := runDetached(cmd, shared, env, logPath, 2*time.Second)
	showLog := func() string { b, _ := os.ReadFile(logPath); return string(b) }
	if err != nil {
		t.Fatalf("start: %v\n%s", err, showLog())
	}
	in := Instance{Kind: "jupyter", ID: strconv.Itoa(port), PID: cmd.Process.Pid, Port: port, Dir: shared, Sandbox: true}
	saveInstance(cfg, in)
	defer stop(in)

	base := fmt.Sprintf("127.0.0.1:%d", port)
	get := func(path string) (int, string) {
		resp, err := http.Get("http://" + base + path)
		if err != nil {
			return 0, err.Error()
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(b)
	}
	deadline := time.Now().Add(60 * time.Second)
	for {
		if code, _ := get("/api/status?token=" + token); code == 200 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("Jupyter didn't answer\n%s", showLog())
		}
		time.Sleep(500 * time.Millisecond)
	}

	if code, _ := get("/api/status"); code != 403 {
		t.Errorf("no token: HTTP %d, want 403", code)
	}
	if code, body := get("/api/contents?token=" + token); code != 200 || !strings.Contains(body, "notes.txt") {
		t.Errorf("contents of the shared folder: HTTP %d %s", code, body)
	}
	if code, _ := get("/api/contents/..%2Fother?token=" + token); code == 200 {
		t.Errorf("the file browser reached a folder above its root")
	}
	if n := len(untrackedJupyter(cfg)); n != 0 {
		t.Errorf("the sandboxed Jupyter shows up as %d untracked server(s)", n)
	}

	client := exec.Command(filepath.Join(bin, "python"), "-c", terminalClient, base, token,
		`cat "$OTHER/secret.txt"; cat ~/.bashrc; cat ../../other/secret.txt; ls /proc | grep -c "^[0-9]"; touch new.txt`)
	out, err := client.CombinedOutput()
	t.Logf("Jupyter terminal:\n%s", out)
	if err != nil {
		t.Fatalf("terminal client: %v\n%s", err, showLog())
	}
	if strings.Contains(string(out), "TOPSECRET") {
		t.Errorf("the terminal read a file outside the shared folder")
	}
	if !strings.Contains(string(out), "No such file") {
		t.Errorf("expected 'No such file' for the paths outside the shared folder")
	}
	if _, err := os.Stat(filepath.Join(shared, "new.txt")); err != nil {
		t.Errorf("a file made in the terminal isn't in the shared folder: %v", err)
	}

	// Stopping signals the sandbox's PID 1, which passes it on; everything inside must go
	if err := stop(in); err != nil {
		t.Fatal(err)
	}
	for deadline := time.Now().Add(20 * time.Second); alive(in.PID); {
		if time.Now().After(deadline) {
			t.Fatalf("the sandbox didn't stop\n%s", showLog())
		}
		time.Sleep(200 * time.Millisecond)
	}
	if code, _ := get("/api/status?token=" + token); code != 0 {
		t.Errorf("Jupyter still answers after stopping")
	}
}
