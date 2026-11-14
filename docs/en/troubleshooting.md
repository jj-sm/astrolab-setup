# Troubleshooting

## `astroenvs: command not found`

The aliases are loaded by new shells. Run `source ~/.bashrc`, or log out and
back in. If it still fails, check the binary is there:

```bash
ls -l ~/.local/bin/astroenvs
~/.local/bin/astroenvs --version
```

If it's missing, download it again (see [Updating the manager](install.md#updating-the-manager)).

## "Could not find jupyter in ~/.ipynb/bin"

The manager runs Jupyter from the `~/.ipynb` virtual environment. Either the
setup was run from another folder (it creates `.ipynb` in the current folder),
or Jupyter isn't installed in it.

- If the environment is elsewhere, point the manager at it with
  `IPYNB_VENV=/path/to/venv` in `~/.astroenvs.env`.
- If Jupyter is missing, install it:

  ```bash
  source ~/.ipynb/bin/activate
  uv pip install jupyterlab ipykernel
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
