# Changelog

## Unreleased

- **Sandboxed Jupyter servers**: a new **Access** question when starting Jupyter.
  A sandboxed server sees only its folder (plus `~/.ipynb` and the system,
  read-only), so its URL can be shared without exposing the rest of your
  account. It uses Linux user namespaces; nothing to install.
- **Security:** the Jupyter token is no longer on the command line, where every
  user of the server could see it with `ps`. `~/.astrolab` and its logs, which
  contain the URLs with tokens, are now readable only by you. Restart Jupyter
  servers started by 1.0.1 or earlier.
- After starting Jupyter, the manager warns that anyone with the URL can run
  commands as you. It warns about tokens shorter than 16 characters for Jupyter
  and CARTA.
- One-command install: `curl -fsSL https://raw.githubusercontent.com/jj-sm/astrolab-setup/main/install.sh | bash`
  downloads the repository to a temporary folder, asks for a token, runs the
  setup and deletes the folder. It doesn't need `git`.
- Re-running the setup reuses an existing `~/.ipynb` instead of stopping at
  uv's "already exists" prompt, and the environment is always created in your
  home folder, wherever you run the setup from.
- The setup finds `.env` and `folder.bash` next to the script, so running it
  from any folder copies your token to `~/.astroenvs.env`. It warns when that
  file has no `TOKEN`.

## 1.0.1 (2026-11-14)

- The bash manager (`astroenvs.sh`) is removed. The setup installs only the Go
  binary; if the download fails, it finishes everything else and stops with the
  command to download it by hand.

## 1.0.0 (2026-11-14)

The manager is now a Go program, `astroenvs`, downloaded as a binary by the
setup script.

- Every prompt is a selector: arrow keys and Enter, Esc to go back
- Jupyter asks which folder to open in (your home by default) and uses it as the
  top of its file browser
- Each Jupyter and CARTA server can use the default token or one you type
- Custom paths are typed with Tab completion
- DS9 lists the FITS files in the chosen folder
- Jupyter servers are tracked like CARTA: their own port, token and log, with the
  URL shown directly
- Logs are picked from a list; **Stop an instance** picks from a list instead of
  asking for a PID
- Settings moved from the top of the script to `~/.astroenvs.env`
- Instances started by the bash manager are still recognised
- The setup script falls back to the bash manager if GitHub can't be reached
