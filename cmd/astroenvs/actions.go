package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/huh"
)

// stopped reports whether a prompt ended the action: a cancel goes quietly back to the menu
func stopped(err error) bool {
	if err == nil {
		return false
	}
	if !errors.Is(err, errCancel) {
		fail("%v", err)
		pause()
	}
	return true
}

func executable(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir() && st.Mode()&0o111 != 0
}

// needBin checks that a tool's executable exists; key is its setting name
func needBin(path, name, key string) bool {
	if path != "" && executable(path) {
		return true
	}
	fail("%s executable not found.", name)
	fmt.Printf("  Set %s=/path/to/%s in ~/.astroenvs.env, or export %s before running the manager.\n", key, name, key)
	return false
}

// needDisplay checks for an X display, i.e. an SSH session with X11 forwarding
func needDisplay(cfg *Config, app string) bool {
	if os.Getenv("DISPLAY") != "" {
		return true
	}
	fail("No X display ($DISPLAY is empty), so %s can't open a window.", app)
	fmt.Println("  Reconnect with X11 forwarding (needs XQuartz on your Mac) and run the manager again:")
	fmt.Printf("    ssh -Y %s@%s\n", os.Getenv("USER"), cfg.ServerIP)
	return false
}

func showStartFailure(err error, log string) {
	fail("Failed to start: %v. Last lines of %s:", err, tilde(log))
	if t, terr := tail(log, 15); terr == nil {
		fmt.Println(t)
	}
}

func takenPorts(cfg *Config, kind string) map[int]bool {
	taken := map[int]bool{}
	for _, in := range instances(cfg, kind) {
		taken[in.Port] = true
	}
	return taken
}

// ---------------------------------------------------------------- Servers (Jupyter, CARTA)

func startJupyter(cfg *Config) {
	heading("Starting Jupyter Server")
	bin := filepath.Join(cfg.IpynbVenv, "bin")
	jupyter := filepath.Join(bin, "jupyter")
	if !executable(jupyter) {
		fail("Could not find jupyter in %s", bin)
		fmt.Println("  Set IPYNB_VENV in ~/.astroenvs.env if your venv lives elsewhere, or install it with:")
		fmt.Printf("    source %s/bin/activate && uv pip install jupyterlab ipykernel\n", cfg.IpynbVenv)
		pause()
		return
	}

	dir, err := pickDir(cfg, cfg.JupyterRoot)
	if stopped(err) {
		return
	}
	sandbox, err := pickSandbox(dir)
	if stopped(err) {
		return
	}
	if sandbox && (dir == "/" || strings.HasPrefix(home()+"/", dir+"/")) {
		// The folder would hold ~/.bashrc and the like: changing them reaches outside the sandbox
		fail("A sandboxed server can't use %s, since it contains your whole home folder. Pick a subfolder.", tilde(dir))
		pause()
		return
	}
	token, err := pickToken(cfg)
	if stopped(err) {
		return
	}

	port := freePort(cfg.JupyterPort, takenPorts(cfg, "jupyter"))
	log := filepath.Join(cfg.LogDir, fmt.Sprintf("jupyter-%d.log", port))
	args := []string{"lab", "--no-browser", "--ip=0.0.0.0", "--port=" + strconv.Itoa(port),
		// Fail instead of silently moving to another port, so the URL shown is right
		"--ServerApp.port_retries=0",
		"--ServerApp.allow_remote_access=True",
		"--ServerApp.root_dir=" + dir}
	// Same as activating the venv, so kernels and extensions resolve from it. The token goes in
	// the environment: command lines are visible to every user on the host (ps), environments aren't
	env := []string{"PATH=" + bin + ":" + os.Getenv("PATH"), "VIRTUAL_ENV=" + cfg.IpynbVenv,
		"JUPYTER_TOKEN=" + token}

	var cmd *exec.Cmd
	if sandbox {
		cmd = sandboxCommand(dir, sandboxReadOnly(cfg), jupyter, args...)
	} else {
		cmd = exec.Command(jupyter, args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	}
	cmd, err = runDetached(cmd, dir, env, log, 3*time.Second)
	if err != nil {
		showStartFailure(err, log)
		pause()
		return
	}
	in := Instance{Kind: "jupyter", ID: strconv.Itoa(port), PID: cmd.Process.Pid, Port: port,
		Token: token, Dir: dir, Log: log, Started: time.Now(), Sandbox: sandbox}
	saveInstance(cfg, in)
	ok("Jupyter started in the background (PID %d, port %d)!", in.PID, port)
	fmt.Println("Open in your browser (VPN on):")
	fmt.Println("  " + styleBold.Render(in.URL(cfg)))
	if sandbox {
		warn("Anyone with this URL can read, change and delete the files in %s.", tilde(dir))
	} else {
		// The folder only limits the file browser: terminals and notebooks run as this account
		warn("Anyone with this URL can run commands as you, in any folder. Don't share it.")
	}
	warnWeakToken(token)
	pause()
}

// sandboxReadOnly lists what a sandboxed Jupyter needs besides its folder: the venv, the
// Python it was made from (uv keeps it in the home folder) and TinyTeX for Matplotlib's usetex
func sandboxReadOnly(cfg *Config) []string {
	ro := []string{cfg.IpynbVenv}
	if py := venvPythonHome(cfg.IpynbVenv); py != "" {
		ro = append(ro, py)
	}
	if tex := filepath.Join(home(), ".TinyTeX"); isDir(tex) {
		ro = append(ro, tex)
	}
	return ro
}

// venvPythonHome returns the installation a venv's python links to, from pyvenv.cfg's "home"
// (its bin folder), or "" if it can't be read
func venvPythonHome(venv string) string {
	vals := readEnvFile(filepath.Join(venv, "pyvenv.cfg"))
	if bin := vals["home"]; filepath.IsAbs(bin) {
		return filepath.Dir(bin)
	}
	return ""
}

func startCarta(cfg *Config) {
	heading("Starting CARTA Server")
	if !needBin(cfg.CartaBin, "carta", "CARTA_BIN") {
		pause()
		return
	}
	dir, err := pickDir(cfg)
	if stopped(err) {
		return
	}
	token, err := pickToken(cfg)
	if stopped(err) {
		return
	}

	port := freePort(cfg.CartaPort, takenPorts(cfg, "carta"))
	log := filepath.Join(cfg.LogDir, fmt.Sprintf("carta-%d.log", port))
	args := []string{"--no_browser", "--port", strconv.Itoa(port), "--top_level_folder", cfg.CartaTop, dir}

	cmd, err := startDetached(cfg.CartaBin, args, dir, []string{"CARTA_AUTH_TOKEN=" + token}, log, 3*time.Second)
	if err != nil {
		showStartFailure(err, log)
		pause()
		return
	}
	in := Instance{Kind: "carta", ID: strconv.Itoa(port), PID: cmd.Process.Pid, Port: port,
		Token: token, Dir: dir, Log: log, Started: time.Now()}
	saveInstance(cfg, in)
	ok("CARTA started in the background (PID %d, port %d)!", in.PID, port)
	fmt.Println("Open in your browser (VPN on):")
	fmt.Println("  " + styleBold.Render(in.URL(cfg)))
	warnWeakToken(token)
	pause()
}

func viewServers(cfg *Config, kind, name string) {
	heading("Active " + name + " Instances")
	list := instances(cfg, kind)
	for _, in := range list {
		sb := ""
		if in.Sandbox {
			sb = styleDim.Render("  (sandboxed)")
		}
		fmt.Printf("%s PID %d  port %d  folder: %s%s\n", styleOK.Render("●"), in.PID, in.Port, tilde(in.Dir), sb)
		if in.Token != "" {
			fmt.Println("  " + styleBold.Render(in.URL(cfg)))
		}
	}

	others := 0
	if kind == "jupyter" {
		others = len(untrackedJupyter(cfg))
	}
	if others > 0 {
		// Started outside the manager (or by an older version): Jupyter itself knows their URLs
		warn("\n%d Jupyter server(s) not started by this manager:", others)
		cmd := exec.Command(filepath.Join(cfg.IpynbVenv, "bin", "jupyter"), "server", "list")
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if cmd.Run() != nil {
			fmt.Println("  (couldn't run `jupyter server list`)")
		}
	}
	if len(list) == 0 && others == 0 {
		warn("No active %s instances found.", name)
	}
	pause()
}

// viewLogs lets you pick one of a kind's log files, newest first, and shows its end
func viewLogs(cfg *Config, kind, name string) {
	heading("Recent " + name + " Logs")
	files, _ := filepath.Glob(filepath.Join(cfg.LogDir, kind+"-*.log"))
	if len(files) == 0 {
		warn("No %s logs found in %s", name, cfg.LogDir)
		pause()
		return
	}
	mtime := func(f string) time.Time {
		st, err := os.Stat(f)
		if err != nil {
			return time.Time{}
		}
		return st.ModTime()
	}
	sort.Slice(files, func(i, j int) bool { return mtime(files[i]).After(mtime(files[j])) })

	running := map[string]bool{}
	for _, in := range instances(cfg, kind) {
		running[in.Log] = true
	}
	var opts []huh.Option[string]
	for _, f := range files {
		label := fmt.Sprintf("%-22s %s", filepath.Base(f), mtime(f).Format("Jan 02 15:04"))
		if running[f] {
			label += styleOK.Render("  running")
		}
		opts = append(opts, huh.NewOption(label, f))
	}

	var log string
	if stopped(ask(huh.NewSelect[string]().Title("Log").Options(opts...).Value(&log).Height(listHeight(len(opts))))) {
		return
	}
	t, err := tail(log, 30)
	if err != nil {
		fail("%v", err)
	} else {
		fmt.Println(styleDim.Render("--- " + tilde(log)))
		fmt.Println(t)
	}
	pause()
}

// ---------------------------------------------------------------- DS9 & CASA

var fitsExt = []string{".fits", ".fit", ".fts", ".fits.gz", ".fz"}

func isFITS(name string) bool {
	lower := strings.ToLower(name)
	for _, ext := range fitsExt {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// pickFITS offers the FITS files in dir, none, or another path typed with Tab completion
func pickFITS(dir string) (string, error) {
	const none, other = "\x00none", "\x00other"
	opts := []huh.Option[string]{huh.NewOption("No file (open DS9 empty)", none)}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if !e.IsDir() && isFITS(e.Name()) {
			opts = append(opts, huh.NewOption(e.Name(), e.Name()))
		}
	}
	opts = append(opts, huh.NewOption("Other file…", other))

	var file string
	err := ask(huh.NewSelect[string]().Title("FITS file to open").Options(opts...).Value(&file).Height(listHeight(len(opts))))
	switch {
	case err != nil || file == none:
		return "", err
	case file != other:
		return file, nil
	}

	var typed string
	err = ask(huh.NewInput().Title("File").Description("Relative to "+tilde(dir)+"; Tab completes").
		Value(&typed).
		SuggestionsFunc(func() []string { return pathSuggestions(typed, dir, false) }, &typed).
		Validate(func(s string) error {
			st, err := os.Stat(resolve(s, dir))
			if err != nil || st.IsDir() {
				return fmt.Errorf("%s is not a file", s)
			}
			return nil
		}))
	return resolve(typed, dir), err
}

func startDS9(cfg *Config) {
	heading("Opening DS9")
	if !needBin(cfg.DS9Bin, "ds9", "DS9_BIN") || !needDisplay(cfg, "DS9") {
		pause()
		return
	}
	dir, err := pickDir(cfg)
	if stopped(err) {
		return
	}
	file, err := pickFITS(dir)
	if stopped(err) {
		return
	}

	var args []string
	if file != "" {
		args = []string{file}
	}
	// The PID isn't known until it starts, so the log is renamed afterwards
	log := filepath.Join(cfg.LogDir, "ds9-starting.log")
	// DS9's file dialog opens in its working directory
	cmd, err := startDetached(cfg.DS9Bin, args, dir, nil, log, 2*time.Second)
	if err != nil {
		showStartFailure(err, log)
		pause()
		return
	}
	pid := cmd.Process.Pid
	final := filepath.Join(cfg.LogDir, fmt.Sprintf("ds9-%d.log", pid))
	os.Rename(log, final)
	saveInstance(cfg, Instance{Kind: "ds9", ID: strconv.Itoa(pid), PID: pid, Dir: dir, File: file,
		Log: final, Started: time.Now()})
	ok("DS9 opened (PID %d). The window should appear on your screen.", pid)
	pause()
}

func startCasa(cfg *Config) {
	heading("Starting CASA")
	if !needBin(cfg.CasaBin, "casa", "CASA_BIN") {
		pause()
		return
	}
	dir, err := pickDir(cfg)
	if stopped(err) {
		return
	}

	var args []string
	if os.Getenv("DISPLAY") == "" {
		args = []string{"--nologger", "--nogui"}
		warn("No X display: starting CASA without the logger/GUI windows (tasks and tools still work).")
	}
	fmt.Printf("CASA logs go to %s. Type %s in CASA to come back to this menu.\n\n", dir, styleBold.Render("exit"))

	cmd := exec.Command(cfg.CasaBin, args...)
	cmd.Dir = dir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	// Ctrl+C belongs to CASA while it runs. Catching (not ignoring) it here keeps the manager
	// alive, and CASA still gets the default handling since caught signals reset on exec
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGQUIT)
	if err := cmd.Run(); err != nil {
		warn("CASA exited: %v", err)
	}
	signal.Stop(sig)
	pause()
}

// ---------------------------------------------------------------- Manage

func describe(in Instance) string {
	switch in.Kind {
	case "ds9":
		s := fmt.Sprintf("ds9      PID %-7d %s", in.PID, tilde(in.Dir))
		if in.File != "" {
			s += "  " + filepath.Base(in.File)
		}
		return s
	default:
		s := fmt.Sprintf("%-8s PID %-7d port %d  %s", in.Kind, in.PID, in.Port, tilde(in.Dir))
		if in.Sandbox {
			s += styleDim.Render("  (sandboxed)")
		}
		return s
	}
}

func allInstances(cfg *Config) []Instance {
	var all []Instance
	for _, kind := range []string{"jupyter", "carta", "ds9"} {
		all = append(all, instances(cfg, kind)...)
	}
	for _, pid := range untrackedJupyter(cfg) {
		dir := processArg(pid, "--ServerApp.root_dir")
		all = append(all, Instance{Kind: "jupyter", PID: pid, Dir: dir + styleDim.Render("  (not started by this manager)")})
	}
	return all
}

func stopInstance(cfg *Config) {
	heading("Stop an instance")
	all := allInstances(cfg)
	if len(all) == 0 {
		warn("No active instances found.")
		pause()
		return
	}

	opts := make([]huh.Option[int], 0, len(all)+1)
	for i, in := range all {
		opts = append(opts, huh.NewOption(describe(in), i))
	}
	opts = append(opts, huh.NewOption("Cancel", -1))
	var pick int
	if stopped(ask(huh.NewSelect[int]().Title("Instance to stop").Options(opts...).Value(&pick))) || pick < 0 {
		return
	}

	in := all[pick]
	if err := stop(in); err != nil {
		fail("Failed to stop PID %d: %v", in.PID, err)
	} else {
		ok("Stopped %s (PID %d)", in.Kind, in.PID)
	}
	pause()
}

func clearAll(cfg *Config) {
	heading("Stop ALL instances")
	var which string
	err := ask(huh.NewSelect[string]().Title("Which?").Options(
		huh.NewOption("Everything", "all"),
		huh.NewOption("All Jupyter servers", "jupyter"),
		huh.NewOption("All CARTA servers", "carta"),
		huh.NewOption("All DS9 windows", "ds9"),
		huh.NewOption("Cancel", "cancel"),
	).Value(&which))
	if stopped(err) || which == "cancel" {
		return
	}

	for _, kind := range []string{"jupyter", "carta", "ds9"} {
		if which != "all" && which != kind {
			continue
		}
		n := 0
		for _, in := range allInstances(cfg) {
			if in.Kind == kind && stop(in) == nil {
				n++
			}
		}
		ok("Stopped %d %s instance(s).", n, kind)
	}
	pause()
}
