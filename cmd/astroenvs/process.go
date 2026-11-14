package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var execLookPath = exec.LookPath

// Instance is a server or window started by the manager, saved as ~/.astrolab/run/<kind>-<id>.json
type Instance struct {
	Kind    string    `json:"kind"` // jupyter, carta, ds9
	ID      string    `json:"id"`   // port for servers, PID for DS9
	PID     int       `json:"pid"`
	Port    int       `json:"port,omitempty"`
	Token   string    `json:"token,omitempty"`
	Dir     string    `json:"dir"`
	File    string    `json:"file,omitempty"` // FITS file DS9 opened
	Log     string    `json:"log"`
	Sandbox bool      `json:"sandbox,omitempty"` // Jupyter that only sees Dir
	Started time.Time `json:"started"`

	runFile string
}

func (in Instance) URL(cfg *Config) string {
	path := "/"
	if in.Kind == "jupyter" {
		path = "/lab"
	}
	return fmt.Sprintf("http://%s:%d%s?token=%s", cfg.ServerIP, in.Port, path, in.Token)
}

func alive(pid int) bool {
	return pid > 0 && syscall.Kill(pid, 0) == nil
}

func saveInstance(cfg *Config, in Instance) error {
	data, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(cfg.RunDir, in.Kind+"-"+in.ID+".json"), data, 0o600)
}

// instances returns the live tracked instances of a kind, oldest first, and drops stale entries.
// Plain-text files left by the bash manager ("pid [token] dir") are still read
func instances(cfg *Config, kind string) []Instance {
	files, _ := filepath.Glob(filepath.Join(cfg.RunDir, kind+"-*"))
	var out []Instance
	for _, f := range files {
		in, err := readInstance(f, kind)
		if err != nil || !alive(in.PID) {
			os.Remove(f)
			continue
		}
		in.runFile = f
		out = append(out, in)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Started.Before(out[j].Started) })
	return out
}

func readInstance(path, kind string) (Instance, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Instance{}, err
	}
	var in Instance
	if strings.HasSuffix(path, ".json") {
		err = json.Unmarshal(data, &in)
		return in, err
	}

	// Legacy bash format
	id := strings.TrimPrefix(filepath.Base(path), kind+"-")
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return in, errors.New("empty run file")
	}
	in = Instance{Kind: kind, ID: id}
	in.PID, err = strconv.Atoi(fields[0])
	rest := fields[1:]
	if kind == "carta" {
		in.Port, _ = strconv.Atoi(id)
		if len(rest) > 0 && !strings.HasPrefix(rest[0], "/") {
			in.Token, rest = rest[0], rest[1:]
		}
	}
	in.Dir = strings.Join(rest, " ")
	return in, err
}

func removeInstance(in Instance) {
	if in.runFile != "" {
		os.Remove(in.runFile)
	}
}

// untrackedJupyter lists jupyter-lab processes of this user that the manager didn't start
func untrackedJupyter(cfg *Config) []int {
	tracked := map[int]bool{}
	for _, in := range instances(cfg, "jupyter") {
		tracked[in.PID] = true
	}
	out, _ := exec.Command("pgrep", "-u", strconv.Itoa(os.Getuid()), "-f", "jupyter-lab").Output()
	var pids []int
	for _, f := range strings.Fields(string(out)) {
		// A sandboxed Jupyter runs as a child of the tracked sandbox process
		if pid, err := strconv.Atoi(f); err == nil && !tracked[pid] && !tracked[parentPID(pid)] {
			pids = append(pids, pid)
		}
	}
	return pids
}

func parentPID(pid int) int {
	out, _ := exec.Command("ps", "-o", "ppid=", "-p", strconv.Itoa(pid)).Output()
	ppid, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	return ppid
}

// processArg returns the value of --name=value in a process's command line, if present
func processArg(pid int, name string) string {
	out, _ := exec.Command("ps", "-o", "args=", "-p", strconv.Itoa(pid)).Output()
	_, v, ok := strings.Cut(strings.TrimSpace(string(out)), name+"=")
	if !ok {
		return ""
	}
	return v
}

// stopGroup stops a detached instance; it runs in its own session, so the whole group goes
func stopGroup(pid int) error {
	if err := syscall.Kill(-pid, syscall.SIGTERM); err == nil {
		return nil
	}
	return syscall.Kill(pid, syscall.SIGTERM)
}

// stop ends an instance. Jupyter only gets SIGTERM on its main process, so it can shut its kernels down
func stop(in Instance) error {
	var err error
	if in.Kind == "jupyter" {
		err = syscall.Kill(in.PID, syscall.SIGTERM)
	} else {
		err = stopGroup(in.PID)
	}
	if err == nil {
		removeInstance(in)
	}
	return err
}

// startDetached starts a command in its own session with output going to logPath, waits
// `settle` and reports whether it is still running; it keeps running after the manager exits
func startDetached(name string, args []string, dir string, env []string, logPath string, settle time.Duration) (*exec.Cmd, error) {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return runDetached(cmd, dir, env, logPath, settle)
}

// runDetached is startDetached for a prepared command, whose SysProcAttr must start a new session
func runDetached(cmd *exec.Cmd, dir string, env []string, logPath string, settle time.Duration) (*exec.Cmd, error) {
	// Logs print the URL with the token, so only this account may read them
	log, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, err
	}
	defer log.Close()
	log.Chmod(0o600) // O_CREATE's mode doesn't apply to a log left by an earlier run

	cmd.Dir = dir
	cmd.Stdout = log
	cmd.Stderr = log
	cmd.Env = append(os.Environ(), env...)
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err == nil {
			err = errors.New("exited right after starting")
		}
		return cmd, err
	case <-time.After(settle):
		return cmd, nil
	}
}

func portInUse(port int) bool {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return true
	}
	l.Close()
	return false
}

// freePort returns the first port from start on that nobody is listening on (the host is shared)
func freePort(start int, taken map[int]bool) int {
	p := start
	for taken[p] || portInUse(p) {
		p++
	}
	return p
}

// serverIP returns the first non-loopback IPv4 address, like `hostname -I`
func serverIP() string {
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			if ipn, ok := a.(*net.IPNet); ok && ipn.IP.To4() != nil {
				return ipn.IP.String()
			}
		}
	}
	h, _ := os.Hostname()
	return h
}

// tail returns the last n lines of a file
func tail(path string, n int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	const chunk = 64 * 1024
	st, err := f.Stat()
	if err != nil {
		return "", err
	}
	if st.Size() > chunk {
		f.Seek(-chunk, io.SeekEnd)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	lines := bytes.Split(bytes.TrimRight(data, "\n"), []byte("\n"))
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return string(bytes.Join(lines, []byte("\n"))), nil
}
