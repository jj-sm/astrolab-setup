# astrolab-setup

`astrolab-setup` prepares your account on the lab server for astronomy work and
gives you `astroenvs`, a terminal manager that starts and stops Jupyter Lab,
CARTA, DS9 and CASA for you.

![astrolab-screenshot.png](/astrophysics/astrolab-setup/resources/astrolab-screenshot.png =600x){.align-center .decor-border .decor-rounded .decor-shadow}

Everything installs in your home folder, without `sudo`.

Current version: 1.0.1 (see the [changelog](changelog.md)).

| Page | Contents |
|---|---|
| [Installation](install.md) | running the setup script, the token, updating |
| [Using the manager](manager.md) | every menu option, keys, URLs, X11 apps |
| [Configuration](configuration.md) | `~/.astroenvs.env` settings, where logs live |
| [Troubleshooting](troubleshooting.md) | common errors and how to fix them |
| [Development](development.md) | code layout, building, releases, these docs |
| [Changelog](changelog.md) | changes by version |

## What the setup gives you

- **Python** 3.11 and 3.13, installed with [uv](https://docs.astral.sh/uv/)
- **`~/.ipynb`**, a Python 3.13 virtual environment registered as the Jupyter
  kernel *Python (ipynb 3.13)*, with
  [ipynb-jjsm-tools](https://github.com/jj-sm/ipynb-jjsm-tools) installed
- **TinyTeX**, a small LaTeX distribution for Matplotlib's `usetex` and PDF export
- **`astroenvs`**, the manager, with the aliases `astroenvs` and `ipynb`
- **`tfold`**, which creates the `out/`, `src/`, `data/` and `notebooks/`
  folders for a new project

## Quick start

On the server, run:

```bash
curl -fsSL https://raw.githubusercontent.com/jj-sm/astrolab-setup/main/install.sh | bash
```

Choose a token when it asks (Enter picks a random one). When it finishes, close
and reopen the terminal, then run `astroenvs`. To do it step by step instead,
see [Installation](install.md).

Pick **Jupyter · Start server**, press Enter twice (your home folder, the
default token) and open the URL it prints. You need the university VPN on to
reach it.

```text
┃ ASTROENVS MANAGER
┃ Server 10.254.0.10 · v1.0.1
┃ > Jupyter  · Start server
┃   Jupyter  · Active instances
┃   Jupyter  · Logs
┃   CARTA    · Start server
┃   CARTA    · Active instances
┃   CARTA    · Logs
┃   DS9      · Open window (X11)
┃   CASA     · Start in this terminal
┃   Manage   · Stop an instance
┃   Manage   · Stop all instances
┃   Exit
↑ up • ↓ down • / filter • enter submit
```
