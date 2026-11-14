package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds every setting the manager uses. Each one can be set in the env file
// (KEY=value) or, except TOKEN, overridden with an environment variable of the same name.
type Config struct {
	EnvFile string // file the settings were read from ("" if none was found)

	Token       string   // default token for Jupyter and CARTA
	IpynbVenv   string   // venv that provides jupyter
	JupyterRoot string   // default folder Jupyter opens in
	DataDirs    []string // folders offered as the starting point for every tool
	CartaBin    string
	DS9Bin      string
	CasaBin     string
	CartaTop    string // CARTA's file browser can't go above this folder
	CartaPort   int    // first CARTA port to try; the next free one is used (the host is shared)
	JupyterPort int    // same, for Jupyter
	ServerIP    string // address shown in the URLs

	StateDir string
	LogDir   string
	RunDir   string
}

func home() string {
	h, _ := os.UserHomeDir()
	return h
}

func expandHome(p string) string {
	if p == "~" {
		return home()
	}
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(home(), p[2:])
	}
	return p
}

// findEnvFile returns the first settings file that exists, in the same order the bash manager used
func findEnvFile() string {
	candidates := []string{os.Getenv("ASTROENVS_ENV"), filepath.Join(home(), ".astroenvs.env")}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), ".env"))
	}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	return ""
}

// readEnvFile parses KEY=value lines; later lines win, quotes and "export " are ignored
func readEnvFile(path string) map[string]string {
	vals := map[string]string{}
	if path == "" {
		return vals
	}
	f, err := os.Open(path)
	if err != nil {
		return vals
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		vals[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"'`)
	}
	return vals
}

// tilde shortens a path under the home folder to ~/..., for display
func tilde(p string) string {
	h := home()
	if p == h {
		return "~"
	}
	if strings.HasPrefix(p, h+"/") {
		return "~" + p[len(h):]
	}
	return p
}

func lookPath(name string) string {
	if p, err := execLookPath(name); err == nil {
		return p
	}
	return ""
}

func loadConfig() *Config {
	c := &Config{EnvFile: findEnvFile()}
	file := readEnvFile(c.EnvFile)

	// Environment variable first, then the env file, then the default
	get := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		if v := file[key]; v != "" {
			return v
		}
		return def
	}
	getInt := func(key string, def int) int {
		if n, err := strconv.Atoi(get(key, "")); err == nil && n > 0 {
			return n
		}
		return def
	}

	// TOKEN only comes from the file: a generic TOKEN variable in the shell is likely something else
	c.Token = file["TOKEN"]
	c.IpynbVenv = expandHome(get("IPYNB_VENV", "~/.ipynb"))
	c.JupyterRoot = expandHome(get("JUPYTER_ROOT", "~"))
	for _, d := range strings.Split(get("DATA_DIRS", "/data:~/UC-AST221-Rsync"), ":") {
		if d = strings.TrimSpace(d); d != "" {
			c.DataDirs = append(c.DataDirs, expandHome(d))
		}
	}
	c.CartaBin = expandHome(get("CARTA_BIN", lookPath("carta")))
	c.DS9Bin = expandHome(get("DS9_BIN", lookPath("ds9")))
	casa := lookPath("casa")
	if casa == "" {
		casa = "/usr/local/casa/bin/casa"
	}
	c.CasaBin = expandHome(get("CASA_BIN", casa))
	c.CartaTop = expandHome(get("CARTA_TOP", "/"))
	c.CartaPort = getInt("CARTA_PORT_START", 3002)
	c.JupyterPort = getInt("JUPYTER_PORT_START", 8888)
	c.ServerIP = get("SERVER_IP", serverIP())

	// Logs and instance tracking live here, so the manager works from any directory
	c.StateDir = filepath.Join(home(), ".astrolab")
	c.LogDir = filepath.Join(c.StateDir, "logs")
	c.RunDir = filepath.Join(c.StateDir, "run")
	os.MkdirAll(c.LogDir, 0o755)
	os.MkdirAll(c.RunDir, 0o755)
	return c
}
