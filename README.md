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

### `ipynb.sh`
Interactive menu-driven Jupyter Lab manager with options:
- **Start Jupyter Server** - Launch Jupyter Lab with custom token in detached mode
- **View Active Instances** - List running Jupyter servers with URLs and ports
- **View Jupyter Logs** - Display Jupyter server logs
- **Stop a Specific Instance** - Terminate a selected Jupyter server
- **Stop & Clear ALL Instances** - Stop all running Jupyter processes
- **Exit** - Close the manager

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

### 3. Run Jupyter manager
```bash
chmod +x ipynb.sh
./ipynb.sh
```

## Configuration

### Jupyter Token
The Jupyter Lab server uses a fixed token for authentication:
- **Token**: `jsanchezm2`
- Set in the `start_jupyter()` function in `ipynb.sh`

To customize the token, edit the `ipynb.sh` file and modify:
```bash
nohup jupyter lab --no-browser --IdentityProvider.token='' --ServerApp.token='jsanchezm2' > jupyter.log 2>&1 &
```

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

- Jupyter server logs are saved to `jupyter.log` in the project directory
- Use option 3 in the Jupyter manager menu to view logs in real-time

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
Check the `jupyter.log` file for error details. Ensure port 8888 (default) is not in use.

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
