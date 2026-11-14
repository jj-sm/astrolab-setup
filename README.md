# AstroLab Setup

A comprehensive setup utility for creating a Python development environment with Jupyter Lab, supporting multiple Python versions and integrated tools.

## Overview

This project automates the initialization of a development workspace with:
- Directory structure creation (`out/`, `src/`, `data/`, `notebooks/`)
- Python 3.11 and Python 3.13 installation
- UV package installer setup
- Jupyter Lab environment configuration with custom kernel registration
- Interactive Jupyter Lab manager with server control utilities

## Prerequisites

- Linux/Unix-based system (uses `apt-get`, `sudo`, and bash)
- Sudo access for package installation
- Approximately 500MB+ disk space for Python versions and packages

## Files

### `folder.bash`
Creates the basic directory structure for the project:
- `out/` - Output directory
- `src/` - Source code directory
- `data/` - Data directory
- `notebooks/` - Jupyter notebooks directory

### `ipynb_setup.bash`
Main setup script that:
1. Updates system package lists
2. Installs build essentials and prerequisites
3. Adds deadsnakes PPA for multiple Python versions
4. Installs Python 3.11 and Python 3.13 with development headers
5. Installs UV (Astral's fast Python package installer)
6. Creates an `ipynb` virtual environment
7. Installs `ipykernel` and `ipynb-jjsm-tools` package
8. Registers the virtual environment as a Jupyter kernel
9. Sets up shell alias for easy access

### `astroenvs` (Go, in `cmd/astroenvs/`)
Interactive manager for Jupyter Lab, CARTA, DS9 and CASA (installed as `~/.local/bin/astroenvs`, aliases `astroenvs` and `ipynb`). Every prompt is a selector: move with ↑/↓, confirm with Enter, go back with Esc.
- **Start Jupyter / CARTA Server** - Launch in the background with the default token from `.env` or one you type (Jupyter picks the next free port from 8888, CARTA from 3002)
- **Active Instances** - List running servers with their URLs and folders
- **Logs** - Pick a Jupyter or CARTA log and show its last lines
- **Open DS9** - Open a detached DS9 window, optionally with a FITS file picked from the folder (needs X11 forwarding)
- **Start CASA** - Run CASA in the current terminal; without X11 it starts with `--nologger --nogui`
- **Stop an instance** - Pick a running Jupyter, CARTA or DS9 instance to stop
- **Stop all instances** - Stop all Jupyter, all CARTA, all DS9, or everything

Every tool first asks which folder to start in: Jupyter offers `JUPYTER_ROOT` (your home) first, then `DATA_DIRS`, or a custom path with Tab completion. For Jupyter that folder is also the top of its file browser.

`astroenvs.sh` is the previous bash version. `ipynb_setup.bash` installs it only if the binary can't be downloaded.

## Quick Start

### 1. Set up directories
```bash
chmod +x folder.bash
./folder.bash
```

### 2. Set up Python environment and Jupyter
```bash
chmod +x ipynb_setup.bash
./ipynb_setup.bash
```

### 3. Run the manager
```bash
astroenvs
```
To update it later without re-running the whole setup:
```bash
curl -fsSL https://github.com/jj-sm/astrolab-setup/releases/latest/download/astroenvs-linux-amd64 -o ~/.local/bin/astroenvs && chmod +x ~/.local/bin/astroenvs
```

### Building and releasing
`.github/workflows/build.yml` builds static binaries (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`) on every push and pull request; they are attached to the workflow run as artifacts. Pushing a version tag also publishes them as a GitHub Release, which is what the setup script and the update command download:
```bash
git tag v1.0.0 && git push origin v1.0.0
```
To build locally: `go build -o astroenvs ./cmd/astroenvs` (Go 1.24+).

## Configuration

### Token
When you start Jupyter Lab or CARTA, the manager asks for a token. Press Enter to use the default, read from `TOKEN` in a `.env` file (see `.env.example`), or type another one for that instance (letters, digits and `. _ ~ -`):
```bash
cp .env.example .env   # then set TOKEN=<something random>
```
`ipynb_setup.bash` copies `.env` to `~/.astroenvs.env` (mode 600). The manager reads its settings from the first of these that exists:
1. the file named by `$ASTROENVS_ENV`
2. `~/.astroenvs.env`
3. `.env` next to the executable

Without a default you have to type a token every time; an empty token is never accepted, since the server would then have no authentication. `.env` is git-ignored; never commit it.

### Settings
Add any of these to `~/.astroenvs.env` as `KEY=value` (or `export` them before running; the environment wins over the file, except for `TOKEN`):
- `JUPYTER_ROOT` - default folder Jupyter opens in (default `~`)
- `DATA_DIRS` - folders offered as the starting point for every tool, separated by `:` (default `/data:~/UC-AST221-Rsync`)
- `IPYNB_VENV` - venv that provides `jupyter` (default `~/.ipynb`)
- `CARTA_BIN`, `DS9_BIN`, `CASA_BIN` - executables (default: found in `PATH`; CASA falls back to `/usr/local/casa/bin/casa`)
- `CARTA_TOP` - top-level folder the CARTA file browser can reach (default `/`)
- `JUPYTER_PORT_START`, `CARTA_PORT_START` - first port to try (defaults `8888`, `3002`)
- `SERVER_IP` - address shown in the URLs (default: the first network address, like `hostname -I`)

### DS9 and CASA windows (X11)
DS9 and the CASA logger/viewers are desktop windows, so the SSH session running the manager needs X11 forwarding. On a Mac, install [XQuartz](https://www.xquartz.org) and connect with:
```bash
ssh -Y <user>@10.254.0.10
```
Without it, DS9 won't open and CASA starts in terminal-only mode.

### Virtual Environment
The setup creates a virtual environment named `ipynb/` in the project root. To activate it manually:
```bash
source ipynb/bin/activate
```

## Dependencies

- Python 3.11, Python 3.13
- UV package manager
- Jupyter Lab
- ipykernel
- ipynb-jjsm-tools (from GitHub: https://github.com/jj-sm/ipynb-jjsm-tools)

## Logs

- Logs are saved to `~/.astrolab/logs/` (`jupyter-<port>.log` and `carta-<port>.log` per server, `ds9-<pid>.log` per DS9 window)
- Running instances are tracked in `~/.astrolab/run/`
- CASA writes its own `casa-*.log` files in the folder it was started in
- Use **Jupyter · Logs** and **CARTA · Logs** in the manager to view them

## Troubleshooting

### Permission Denied
Make sure scripts are executable:
```bash
chmod +x *.bash *.sh
```

### Python version not found
Verify installations:
```bash
python3.11 --version
python3.13 --version
uv --version
```

### Jupyter fails to start
The manager shows the end of the log when a server fails; the full log is `~/.astrolab/logs/jupyter-<port>.log`.

### Virtual environment activation fails
Manually create it:
```bash
uv venv ipynb
source ipynb/bin/activate
```

## License

MIT

## Author

jj-sm
