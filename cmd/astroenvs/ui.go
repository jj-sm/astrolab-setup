package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	styleTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	styleOK    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleWarn  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleErr   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleBold  = lipgloss.NewStyle().Bold(true)
	styleDim   = lipgloss.NewStyle().Faint(true)
)

func heading(s string) { fmt.Println("\n" + styleTitle.Render("=== "+s+" ===")) }
func ok(format string, a ...any) {
	fmt.Println(styleOK.Render("✔ " + fmt.Sprintf(format, a...)))
}
func warn(format string, a ...any) {
	fmt.Println(styleWarn.Render(fmt.Sprintf(format, a...)))
}
func fail(format string, a ...any) {
	fmt.Println(styleErr.Render("✖ " + fmt.Sprintf(format, a...)))
}

var stdin = bufio.NewReader(os.Stdin)

func pause() {
	fmt.Print(styleDim.Render("\nPress [Enter] to return to menu..."))
	stdin.ReadString('\n')
}

func clearScreen() { fmt.Print("\033[H\033[2J") }

// keymap: Esc cancels like Ctrl+C, and Tab completes paths in text fields
func keymap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"), key.WithHelp("esc", "cancel"))
	km.Input.AcceptSuggestion = key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "complete"))
	km.Input.Next = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "next"))
	return km
}

// ask runs one prompt; errCancel means the user backed out with Esc or Ctrl+C
var errCancel = errors.New("cancelled")

func ask(field huh.Field) error {
	err := huh.NewForm(huh.NewGroup(field)).WithKeyMap(keymap()).WithShowHelp(true).Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return errCancel
	}
	return err
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

// listHeight caps long lists so they scroll; short ones get exactly the rows they need (+ title)
func listHeight(options int) int {
	return min(options, 12) + 1
}

const customPath = "\x00custom"

// pickDir asks for a starting folder: `first` (if given) and then the configured DATA_DIRS,
// or a custom path typed with Tab completion
func pickDir(cfg *Config, first ...string) (string, error) {
	var opts []huh.Option[string]
	seen := map[string]bool{}
	for _, d := range append(first, cfg.DataDirs...) {
		if seen[d] {
			continue
		}
		seen[d] = true
		label := tilde(d)
		if !isDir(d) {
			label += styleDim.Render("  (not found)")
		}
		opts = append(opts, huh.NewOption(label, d))
	}
	opts = append(opts, huh.NewOption("Custom path…", customPath))

	var dir string
	err := ask(huh.NewSelect[string]().Title("Start in").Options(opts...).Value(&dir).
		Validate(func(d string) error {
			if d != customPath && !isDir(d) {
				return fmt.Errorf("%s is not a directory", tilde(d))
			}
			return nil
		}))
	if err != nil || dir != customPath {
		return dir, err
	}

	cwd, _ := os.Getwd()
	var typed string
	err = ask(huh.NewInput().Title("Path").Description("Tab completes folder names").
		Value(&typed).
		SuggestionsFunc(func() []string { return pathSuggestions(typed, cwd, true) }, &typed).
		Validate(func(s string) error {
			if s == "" {
				return errors.New("type a folder")
			}
			if !isDir(resolve(s, cwd)) {
				return fmt.Errorf("%s is not a directory", s)
			}
			return nil
		}))
	return resolve(typed, cwd), err
}

// resolve expands ~ and makes a path absolute against base
func resolve(p, base string) string {
	p = expandHome(p)
	if !filepath.IsAbs(p) {
		p = filepath.Join(base, p)
	}
	return filepath.Clean(p)
}

// pathSuggestions completes the last part of a typed path, keeping what was typed before it
func pathSuggestions(typed, base string, dirsOnly bool) []string {
	if typed == "" {
		return nil
	}
	if typed == "~" {
		return []string{"~/"}
	}
	cut := strings.LastIndex(typed, "/") + 1
	prefix, partial := typed[:cut], typed[cut:]
	parent := base
	if prefix != "" {
		parent = resolve(prefix, base)
	}

	entries, err := os.ReadDir(parent)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, partial) || (strings.HasPrefix(name, ".") && !strings.HasPrefix(partial, ".")) {
			continue
		}
		if isDir(filepath.Join(parent, name)) {
			out = append(out, prefix+name+"/")
		} else if !dirsOnly {
			out = append(out, prefix+name)
		}
		if len(out) >= 200 {
			break
		}
	}
	sort.Strings(out)
	return out
}

var tokenRe = regexp.MustCompile(`^[A-Za-z0-9._~-]+$`)

func validToken(t string) error {
	if t == "" {
		// The server would have no authentication
		return errors.New("a token is required")
	}
	if !tokenRe.MatchString(t) {
		// It goes into URLs (?token=...), so keep to characters that need no escaping
		return errors.New("use only letters, digits and . _ ~ -")
	}
	return nil
}

// warnWeakToken flags tokens short enough to guess: the servers listen on the whole network
func warnWeakToken(t string) {
	if len(t) < 16 {
		warn("This token is short enough to guess. Use 16+ random characters, e.g. from:")
		fmt.Println(`  python3 -c "import secrets; print(secrets.token_urlsafe(24))"`)
	}
}

// pickToken offers the default token from the env file or a custom one
func pickToken(cfg *Config) (string, error) {
	if cfg.Token != "" {
		var choice string
		err := ask(huh.NewSelect[string]().Title("Token").Options(
			huh.NewOption("Default token (from "+tilde(cfg.EnvFile)+")", "default"),
			huh.NewOption("Type a different token…", "custom"),
		).Value(&choice))
		if err != nil || choice == "default" {
			return cfg.Token, err
		}
	}

	var token string
	desc := "Letters, digits and . _ ~ -"
	if cfg.Token == "" {
		desc = "No default set (add TOKEN=... to ~/.astroenvs.env). " + desc
	}
	err := ask(huh.NewInput().Title("Token").Description(desc).Value(&token).Validate(validToken))
	return token, err
}
