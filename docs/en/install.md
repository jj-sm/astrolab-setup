# Installation

## Requirements

- An account on a Linux server with internet access (the lab server, or any
  other machine you can SSH into)
- `git`, `curl` and `wget`, which the server already has
- About 1 GB of free space in your home folder for Python, the virtual
  environment and TinyTeX

You don't need `sudo`: everything goes into your home folder.

## Quick install

On the server, run:

```bash
curl -fsSL https://raw.githubusercontent.com/jj-sm/astrolab-setup/main/install.sh | bash
```

This does steps 1 to 3 below for you: it downloads the repository into a
temporary folder, asks you to choose a token, runs the setup and deletes the
folder, so nothing is left in your home folder but what the setup installs. It
doesn't need `git`. Press Enter at the token prompt for a random one; it's printed and saved
in `~/.astroenvs.env`. If you already set a token in an earlier install, it
keeps that one and doesn't ask, along with your other settings in
`~/.astroenvs.env`.

When it finishes, close and reopen the terminal and go to step 4. Running the
same command again later updates everything.

If you installed by hand before, you can delete `~/astrolab-setup` afterwards:
nothing uses it once the setup has run.

The rest of this page does the same by hand.

## 1. Get the repository

Clone it into your home folder. The setup reads `.env` and `folder.bash` from
the repository folder, wherever you run it from.

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
4. Creates the `~/.ipynb` virtual environment with Python 3.13, or reuses it if
   it already exists, keeping the packages you installed in it
5. Installs `ipykernel` and `ipynb-jjsm-tools[full]` into it and registers the
   kernel *Python (ipynb 3.13)*
6. Lists the installed packages
7. Installs TinyTeX, unless it's already there
8. Installs the LaTeX packages Matplotlib needs, if `tlmgr` is found
9. Downloads the latest `astroenvs` release to `~/.local/bin/astroenvs`, copies
   `.env` to `~/.astroenvs.env` and `folder.bash` to `~/.folder.sh`, and adds the
   aliases to `~/.bashrc`, `~/.bash_profile` and `~/.profile`

If the download in step 9 fails, for example because GitHub can't be reached,
the setup still finishes the rest of the step and then stops with an error
showing the `curl` command to download the manager by hand. The aliases already
point to `~/.local/bin/astroenvs`, so the manager works as soon as you download it.

When it finishes, load the new aliases:

```bash
source ~/.bashrc
```

## 4. Check it works

```bash
astroenvs --version     # astroenvs v1.0.1
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

If you installed before version 1.0.0, running the setup again also moves the
`astroenvs` and `ipynb` aliases from the old `bash ~/.astroenvs.sh` to the new
binary. You can then delete `~/.astroenvs.sh`.

## Project folders (`tfold`)

`tfold` creates the usual layout for a new project in the current folder:

```bash
mkdir ~/my-project && cd ~/my-project
tfold          # creates out/ src/ data/ notebooks/, each with a .gitkeep
```
