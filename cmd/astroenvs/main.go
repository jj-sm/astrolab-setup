// astroenvs is an interactive manager for Jupyter Lab, CARTA, DS9 and CASA on the lab server
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
)

// Set at build time with -ldflags "-X main.version=..."
var version = "dev"

// sandboxArg is the hidden first argument astroenvs restarts itself with to build a sandbox
const sandboxArg = "__sandbox"

type action struct {
	label string
	run   func(*Config)
}

func menu(cfg *Config) []action {
	count := func(kind string) string {
		n := len(instances(cfg, kind))
		if kind == "jupyter" {
			n += len(untrackedJupyter(cfg))
		}
		if n == 0 {
			return ""
		}
		return styleOK.Render(fmt.Sprintf("  (%d running)", n))
	}
	return []action{
		{"Jupyter  · Start server", startJupyter},
		{"Jupyter  · Active instances" + count("jupyter"), func(c *Config) { viewServers(c, "jupyter", "Jupyter") }},
		{"Jupyter  · Logs", func(c *Config) { viewLogs(c, "jupyter", "Jupyter") }},
		{"CARTA    · Start server", startCarta},
		{"CARTA    · Active instances" + count("carta"), func(c *Config) { viewServers(c, "carta", "CARTA") }},
		{"CARTA    · Logs", func(c *Config) { viewLogs(c, "carta", "CARTA") }},
		{"DS9      · Open window (X11)" + count("ds9"), startDS9},
		{"CASA     · Start in this terminal", startCasa},
		{"Manage   · Stop an instance", stopInstance},
		{"Manage   · Stop all instances", clearAll},
		{"Exit", nil},
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == sandboxArg {
		runSandbox(os.Args[2:])
		return
	}
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: astroenvs [--version]\n\n")
		fmt.Fprintf(os.Stderr, "Interactive manager for Jupyter Lab, CARTA, DS9 and CASA.\n")
		fmt.Fprintf(os.Stderr, "Settings are read from ~/.astroenvs.env (or the file named by ASTROENVS_ENV).\n")
	}
	flag.Parse()
	if *showVersion {
		fmt.Println("astroenvs", version)
		return
	}

	cfg := loadConfig()
	choice := 0
	for {
		clearScreen()
		actions := menu(cfg)
		opts := make([]huh.Option[int], len(actions))
		for i, a := range actions {
			opts[i] = huh.NewOption(a.label, i)
		}

		desc := []string{fmt.Sprintf("Server %s · %s", cfg.ServerIP, version)}
		if cfg.Token == "" {
			desc = append(desc, styleWarn.Render("No default TOKEN found: you'll type one when starting Jupyter or CARTA."))
		}
		err := ask(huh.NewSelect[int]().Title("ASTROENVS MANAGER").
			Description(strings.Join(desc, "\n")).Options(opts...).Value(&choice))
		if errors.Is(err, errCancel) || (err == nil && actions[choice].run == nil) {
			fmt.Println(styleTitle.Render("\nGoodbye!"))
			return
		}
		if err != nil {
			fail("%v", err)
			os.Exit(1)
		}
		actions[choice].run(cfg)
	}
}
