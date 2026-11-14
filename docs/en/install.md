# Installation

## Requirements

- An account on a Linux server with internet access (the lab server, or any
  other machine you can SSH into)
- `git`, `curl` and `wget`, which the server already has
- About 1 GB of free space in your home folder for Python, the virtual
  environment and TinyTeX

You don't need `sudo`: everything goes into your home folder.

## 1. Get the repository

Clone it into your home folder. The setup script expects to be run from there,
because it creates the virtual environment in the current folder (`~/.ipynb`),
which is where the manager looks for Jupyter.

```bash
cd ~
git clone https://github.com/jj-sm/astrolab-setup.git
```

## 2. Choose a token

Jupyter and CARTA ask for a token before they let anyone in. Set a default one
in a `.env` file:

```bash
cp astrolab-setup/.env.example astrolab-setup/.env
nano astrolab-setup/.env        # set TOKEN=<something random>
```

Use letters, digits and `. _ ~ -` only, since the token goes into URLs. For a
random one: `python3 -c "import secrets; print(secrets.token_urlsafe(24))"`.

The setup copies this file to `~/.astroenvs.env` with permissions `600`, so
only you can read it. `.env` is ignored by git; never commit it.

You can skip this step: the manager will then ask you to type a token every
time you start a server.

## 3. Run the setup

```bash
bash astrolab-setup/ipynb_setup.bash
```

It runs these steps, stopping at the first error:

1. Installs `uv`
2. Installs Python 3.11 and 3.13 with `uv`
3. Prints the installed versions
4. Creates the `~/.ipynb` virtual environment with Python 3.13
5. Installs `ipykernel` and `ipynb-jjsm-tools[full]` into it and registers the
   kernel *Python (ipynb 3.13)*
6. Lists the installed packages
7. Installs TinyTeX, unless it's already there
8. Installs the LaTeX packages Matplotlib needs, if `tlmgr` is found
9. Downloads the latest `astroenvs` release to `~/.local/bin/astroenvs`, copies
   `.env` to `~/.astroenvs.env` and `folder.bash` to `~/.folder.sh`, and adds the
   aliases to `~/.bashrc`, `~/.bash_profile` and `~/.profile`

If GitHub can't be reached in step 9, it installs the older bash version of the
manager (`astroenvs.sh`) instead, so you still get a working `astroenvs`.

When it finishes, load the new aliases:

```bash
source ~/.bashrc
```

## 4. Check it works

```bash
astroenvs --version     # astroenvs v1.0.0
astroenvs               # opens the menu
```

## Updating the manager

You don't need to run the whole setup again. Download the latest release over
the old binary:

```bash
curl -fsSL https://github.com/jj-sm/astrolab-setup/releases/latest/download/astroenvs-linux-amd64 \
  -o ~/.local/bin/astroenvs && chmod +x ~/.local/bin/astroenvs
```

On an ARM server, replace `linux-amd64` with `linux-arm64`.

Running the setup script again is also safe: it reinstalls the manager and
rewrites the aliases without duplicating them.

## Project folders (`tfold`)

`tfold` creates the usual layout for a new project in the current folder:

```bash
mkdir ~/my-project && cd ~/my-project
tfold          # creates out/ src/ data/ notebooks/, each with a .gitkeep
```
