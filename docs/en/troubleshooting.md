# Troubleshooting

## `astroenvs: command not found`

The aliases are loaded by new shells. Run `source ~/.bashrc`, or log out and
back in. If it still fails, check the binary is there:

```bash
ls -l ~/.local/bin/astroenvs
~/.local/bin/astroenvs --version
```

If it's missing, download it again (see [Updating the manager](install.md#updating-the-manager)).

## The setup says the astroenvs download failed

The setup downloads the manager from GitHub's releases. If that fails, the rest
of the setup is still done, and the last lines show the command to download it
by hand. Check that the server can reach GitHub (`curl -I https://github.com`),
then run that command, or the one in
[Updating the manager](install.md#updating-the-manager).

## "~/.ipynb exists but isn't a virtual environment"

The setup found a `~/.ipynb` folder that isn't a Python environment, and stopped
rather than delete it. Look at what's inside, move it somewhere else (or delete
it if you don't need it), and run the setup again.

## "Could not find jupyter in ~/.ipynb/bin"

The manager runs Jupyter from the `~/.ipynb` virtual environment. Either the
environment is somewhere else, or Jupyter isn't installed in it.

- If the environment is elsewhere, point the manager at it with
  `IPYNB_VENV=/path/to/venv` in `~/.astroenvs.env`.
- If Jupyter is missing, install it:

  ```bash
  source ~/.ipynb/bin/activate
  uv pip install jupyterlab ipykernel
  ```

## "No default set" when asked for a token

The manager reads the default token only from `~/.astroenvs.env`, not from the
`.env` in the repository. The setup copies one to the other, but only when it
gets to its last step. Run the [quick install](install.md) again, which asks for
a token, or add one yourself:

```bash
echo "TOKEN=<your-token>" >> ~/.astroenvs.env && chmod 600 ~/.astroenvs.env
```

## The browser can't open the URL

- Turn on the university VPN: the server isn't reachable from outside.
- Check the server is still running in **Active instances**.
- If the address in the URL isn't the server's, set `SERVER_IP` in
  `~/.astroenvs.env` (see [Configuration](configuration.md)).

## "Failed to start" when starting a server

The manager shows the last lines of the log. The full log is in
`~/.astrolab/logs/` and also under **Logs** in the menu. Common causes:

- **Port already in use**: someone took the port between the check and the
  start. Start it again; the manager picks the next free port.
- **Missing package or extension**: install it in `~/.ipynb`, as above.

## "carta executable not found" (or ds9, casa)

The tool isn't in your `PATH`. Find where it's installed and set it in
`~/.astroenvs.env`:

```bash
CARTA_BIN=/path/to/carta
DS9_BIN=/path/to/ds9
CASA_BIN=/path/to/casa
```

## "No X display" for DS9, or CASA without windows

Your SSH session has no X11 forwarding. Reconnect with `ssh -Y`, with XQuartz
installed on a Mac (see [DS9 and CASA](manager.md#ds9-and-casa)). To check, run
`echo $DISPLAY`: it should print something like `localhost:10.0`.

## The folder I want isn't in the list

Choose **Custom path…** and type it (Tab completes). To have it listed every
time, add it to `DATA_DIRS` in `~/.astroenvs.env`.

## A server keeps running after I stopped it

Look at **Active instances** again: Jupyter can take a few seconds to close its
kernels. If a process is still there, find and stop it by hand:

```bash
pgrep -u $USER -fl 'jupyter-lab|carta|ds9'
kill <PID>
```

## Python or uv not found after the setup

`uv` installs into `~/.local/bin` (older versions used `~/.cargo/bin`). Make
sure that folder is in your `PATH` and check:

```bash
uv --version
uv python find 3.13
```
