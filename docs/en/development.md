# Development

## Repository layout

| Path | Contents |
|---|---|
| `cmd/astroenvs/` | the manager, in Go |
| `ipynb_setup.bash` | the setup script |
| `folder.bash` | creates a project's folders (installed as `tfold`) |
| `docs/en/`, `docs/es/` | these docs, in English and Spanish |
| `.github/workflows/build.yml` | builds the binaries and publishes releases |
| `.github/workflows/wikijs-docs.yml` | publishes `docs/` to docs.jjsm.science |

## The manager's code

`cmd/astroenvs/` is one Go package:

| File | Contents |
|---|---|
| `main.go` | flags and the main menu |
| `config.go` | reading `~/.astroenvs.env` and the defaults |
| `ui.go` | the prompts: folder picker, token, path completion, styles |
| `actions.go` | one function per menu option |
| `process.go` | starting detached processes, tracking instances, ports, logs |
| `sandbox_linux.go` | sandboxed servers (`sandbox_other.go`: the stub for other systems) |

The prompts use [huh](https://github.com/charmbracelet/huh). Servers start in
their own session (`setsid`), so they survive the manager and the SSH session
ending. Each instance is saved as JSON in `~/.astrolab/run/`; the plain-text
files written by the bash manager used before 1.0.0 are still read, so servers it
left running show up.

A sandboxed server works like [bubblewrap](https://github.com/containers/bubblewrap).
The manager starts itself again as `astroenvs __sandbox` in new user, mount and
PID namespaces. That process builds a root holding only the system folders and
the venv (read-only bind mounts) and the chosen folder, pivots into it, drops
every capability and starts Jupyter, staying behind as PID 1. Dropping
capabilities on every thread needs a build without cgo (`CGO_ENABLED=0`); with
cgo the sandbox refuses to start rather than run unsandboxed.

## Building

You need Go 1.24 or newer.

```bash
go build -o astroenvs ./cmd/astroenvs
./astroenvs --version              # astroenvs dev
```

For the server, cross-compile a static Linux binary:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
  -ldflags "-s -w -X main.version=v1.0.0" -o astroenvs-linux-amd64 ./cmd/astroenvs
```

Run `go vet ./...` and `gofmt -l cmd/` before committing.

## Tests

`go test ./...` tests the sandbox on Linux. `TestSandbox` runs a shell in it and
checks what it can see and change. `TestSandboxJupyter` starts a real sandboxed
JupyterLab and tries to leave the folder from a Jupyter terminal; it needs a venv
with JupyterLab in `ASTROENVS_TEST_VENV`. Both skip where user namespaces are
unavailable, unless `ASTROENVS_SANDBOX_TEST=1`. CI runs them in every build,
building the venv with uv like the setup does.

```bash
CGO_ENABLED=0 ASTROENVS_TEST_VENV=~/.ipynb go test -v ./...
```

## Releases

`build.yml` builds `linux/amd64`, `linux/arm64`, `darwin/amd64` and
`darwin/arm64` binaries on every push and pull request, and attaches them to the
workflow run. Pushing a tag that starts with `v` also publishes them, with a
`checksums.txt`, as a GitHub Release:

```bash
git tag v1.1.0
git push origin v1.1.0
```

The setup script and the update command download from
`releases/latest`, so a new release reaches everyone who updates.

## These docs

The docs are Markdown in `docs/en/` and `docs/es/`. Files with the same name are
translations of each other and are linked on the wiki, so keep both languages in
step.

- `README.md` is the overview page at `/astrophysics/astrolab-setup`
- every other file `name.md` becomes `/astrophysics/astrolab-setup/name`, and
  `/es/astrophysics/astrolab-setup/name` for Spanish
- the first `# Heading` is the page title
- link between pages with relative links (`[Configuration](configuration.md)`);
  they are rewritten to the right language
- `_sidebar.json` in each folder is that language's sidebar

`wikijs-docs.yml` publishes them on every push to `main` that changes `docs/`,
using `.github/scripts/wikijs_sync.py` and the project entry in
`.github/wikijs-projects.json`. It can also be run by hand from the Actions tab.

To preview locally without publishing:

```bash
python3 .github/scripts/wikijs_sync.py --src docs --locales en,es \
  --prefix astrophysics/astrolab-setup --dry-run
```

This writes the pages as they would be published into `wikisync-preview/`.

The workflow needs these repository secrets: `WIKI_USERNAME`, `WIKI_PASSWORD`,
`WIKI_SITE_ID`, `CF_API_TOKEN` and `CF_ZONE_ID`. The wiki user's group also needs
a page rule that lets it write pages under `astrophysics/astrolab-setup`, in both
locales.
