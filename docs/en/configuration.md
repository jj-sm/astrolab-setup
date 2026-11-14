# Configuration

## The settings file

The manager reads its settings from the first of these files that exists:

1. the file named by the `ASTROENVS_ENV` environment variable
2. `~/.astroenvs.env` (created by the setup from your `.env`)
3. `.env` next to the `astroenvs` executable

Each setting is a `KEY=value` line. Blank lines and lines starting with `#` are
ignored, and so are quotes around the value and a leading `export`:

```bash
TOKEN=k3Jq9-wX2v_LpA
JUPYTER_ROOT=~/notebooks
DATA_DIRS=/data:~/UC-AST221-Rsync:~/Tareas
CARTA_BIN=/opt/carta/bin/carta
```

An environment variable with the same name wins over the file, so you can try a
value for one run without editing anything:

```bash
DATA_DIRS=/data:/scratch/$USER astroenvs
```

`TOKEN` is the exception: it's only read from the file, because a `TOKEN`
variable in your shell is probably meant for something else.

The manager reads the file when it starts. Restart it after editing.

## Settings

| Setting | Default | What it does |
|---|---|---|
| `TOKEN` | *(none)* | default token for Jupyter and CARTA; without it you type one each time |
| `JUPYTER_ROOT` | `~` | folder offered first (and by default) when starting Jupyter |
| `DATA_DIRS` | `/data:~/UC-AST221-Rsync` | folders offered as the starting point for every tool, separated by `:` |
| `IPYNB_VENV` | `~/.ipynb` | virtual environment that provides `jupyter` |
| `CARTA_BIN` | `carta` from `PATH` | CARTA executable |
| `DS9_BIN` | `ds9` from `PATH` | DS9 executable |
| `CASA_BIN` | `casa` from `PATH`, else `/usr/local/casa/bin/casa` | CASA executable |
| `CARTA_TOP` | `/` | CARTA's file browser can't go above this folder |
| `JUPYTER_PORT_START` | `8888` | first port tried for Jupyter |
| `CARTA_PORT_START` | `3002` | first port tried for CARTA |
| `SERVER_IP` | first network address | address shown in the URLs |

Paths can start with `~`. Set `SERVER_IP` if the URLs show the wrong address,
which can happen on machines with several network interfaces.

## Tokens

A token has to be non-empty and use only letters, digits and `. _ ~ -`. The
manager won't start a server without one, because the server would then be open
to anyone who can reach it.

The default token is the one in the settings file. When starting a server you
can type a different one; it's used for that server only and shown in its URL.

## Files the manager writes

Everything lives in `~/.astrolab/`, so the manager works the same from any
folder:

| Path | Contents |
|---|---|
| `~/.astrolab/logs/jupyter-<port>.log` | output of each Jupyter server |
| `~/.astrolab/logs/carta-<port>.log` | output of each CARTA server |
| `~/.astrolab/logs/ds9-<pid>.log` | output of each DS9 window |
| `~/.astrolab/run/<kind>-<id>.json` | one file per running instance: PID, port, token, folder, log |

The files in `run/` are removed when their instance stops, and stale ones are
cleaned up automatically. Logs are kept; delete old ones whenever you like.
CASA writes its own `casa-*.log` files in the folder you started it in.
