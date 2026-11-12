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

### `astroenvs.sh`
Interactive menu-driven manager for Jupyter Lab, CARTA, DS9 and CASA (installed as `~/.astroenvs.sh`, aliases `astroenvs` and `ipynb`):
- **Start Jupyter / CARTA Server** - Launch in detached mode with the token from `.env` (CARTA picks the next free port from 3002)
- **View Active Instances** - List running servers with URLs and ports
- **View Logs** - Display Jupyter or CARTA server logs
- **Open DS9** - Open a detached DS9 window, optionally with a FITS file (needs X11 forwarding)
- **Start CASA** - Run CASA in the current terminal; without X11 it starts with `--nologger --nogui`
- **Stop a Specific Instance** - Terminate a selected Jupyter, CARTA or DS9 instance
- **Stop & Clear ALL Instances** - Stop all Jupyter, all CARTA, all DS9, or everything
- **Exit** - Close the manager

CARTA, DS9 and CASA first ask which folder to start in (`/data`, `~/UC-AST221-Rsync`, or a custom path).

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
chmod +x astroenvs.sh
./astroenvs.sh
```

## Configuration

### Token
Jupyter Lab and CARTA share one token, read from `TOKEN` in a `.env` file (see `.env.example`):
```bash
cp .env.example .env   # then set TOKEN=<something random>
```
`ipynb_setup.bash` copies `.env` to `~/.astroenvs.env` (mode 600). The manager looks for the token in, in order:
1. the file named by `$ASTROENVS_ENV`
2. `~/.astroenvs.env`
3. `.env` next to the script

It refuses to start if no token is found, since Jupyter with an empty token has no authentication. `.env` is git-ignored; never commit it.

### Tools
At the top of `astroenvs.sh`:
- `CARTA_BIN`, `DS9_BIN`, `CASA_BIN` - executables (default: found in `PATH`; can also be set with `export`)
- `CARTA_TOP` - top-level folder the CARTA file browser can reach (default `/`)
- `DATA_DIRS` - folders offered as the starting point for CARTA, DS9 and CASA

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

- Logs are saved to `~/.astrolab/logs/` (`jupyter.log`, `carta-<port>.log` per CARTA instance, `ds9-<pid>.log` per DS9 window)
- CASA writes its own `casa-*.log` files in the folder it was started in
- Use options 3 (Jupyter) and 6 (CARTA) in the manager menu to view them

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
Check `~/.astrolab/logs/jupyter.log` for error details. Ensure port 8888 (default) is not in use.

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
