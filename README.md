# astrolab-setup

Sets up your account on the lab server for astronomy work, and installs
`astroenvs`, a terminal manager for Jupyter Lab, CARTA, DS9 and CASA.
Everything goes into your home folder, without `sudo`.

**Documentation:** [docs.jjsm.science/astrophysics/astrolab-setup](https://docs.jjsm.science/astrophysics/astrolab-setup)
· [en español](https://docs.jjsm.science/es/astrophysics/astrolab-setup)

## What you get

- Python 3.11 and 3.13 via [uv](https://docs.astral.sh/uv/)
- `~/.ipynb`, a Python 3.13 environment registered as the Jupyter kernel
  *Python (ipynb 3.13)*, with [ipynb-jjsm-tools](https://github.com/jj-sm/ipynb-jjsm-tools)
- TinyTeX, for Matplotlib's `usetex` and PDF export
- `astroenvs` (alias `ipynb`): start, list and stop Jupyter and CARTA servers,
  open DS9, run CASA, read logs, all from arrow-key menus
- `tfold`: creates `out/ src/ data/ notebooks/` for a new project

## Quick start

On the server, run:

```bash
curl -fsSL https://raw.githubusercontent.com/jj-sm/astrolab-setup/main/install.sh | bash
```

It clones this repository into `~/astrolab-setup`, asks you to choose a token
(Enter picks a random one) and runs the setup. When it finishes, close and
reopen the terminal and run `astroenvs`.

Choose **Jupyter · Start server**, press Enter twice and open the URL it prints
(with the university VPN on).

Update the manager later with:

```bash
curl -fsSL https://github.com/jj-sm/astrolab-setup/releases/latest/download/astroenvs-linux-amd64 \
  -o ~/.local/bin/astroenvs && chmod +x ~/.local/bin/astroenvs
```

## Configuration

Settings live in `~/.astroenvs.env` as `KEY=value` lines: `TOKEN`,
`JUPYTER_ROOT`, `DATA_DIRS`, `CARTA_BIN`, and more. See
[Configuration](https://docs.jjsm.science/astrophysics/astrolab-setup/configuration)
for the full list.

DS9 and CASA's windows need X11 forwarding: install
[XQuartz](https://www.xquartz.org) on a Mac and connect with `ssh -Y <user>@10.254.0.10`.

## Repository

| Path | Contents |
|---|---|
| `cmd/astroenvs/` | the manager, in Go |
| `install.sh` | the one-command installer: clones the repo, asks for a token, runs the setup |
| `ipynb_setup.bash` | the setup script |
| `folder.bash` | project folders (`tfold`) |
| `docs/en/`, `docs/es/` | the documentation, published to docs.jjsm.science |

Build with `go build -o astroenvs ./cmd/astroenvs` (Go 1.24+). Pushing a `v*`
tag publishes release binaries; pushing changes to `docs/` publishes the docs.
See [Development](https://docs.jjsm.science/astrophysics/astrolab-setup/development).

## License

MIT
