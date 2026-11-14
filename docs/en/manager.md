# Using the manager

Run `astroenvs` (or `ipynb`) in a terminal on the server. Every question is a
list you move through with the keyboard:

| Key | Action |
|---|---|
| ↑ / ↓ | move |
| Enter | choose |
| Esc or Ctrl+C | go back (in the main menu: exit) |
| / | filter the list by typing |
| Tab | complete a folder or file name, in text fields |

Servers keep running after you close the manager or log out. Start the manager
again at any time to see them, read their logs or stop them.

## Starting Jupyter

**Jupyter · Start server** asks two questions:

1. **Start in**: the folder Jupyter opens in. Your home folder is first and is
   the default, followed by the folders in `DATA_DIRS`, and **Custom path…**
   for any other folder.
2. **Token**: the default one from `~/.astroenvs.env`, or **Type a different
   token…** to use another one for this server only.

Then it prints the address to open:

```text
✔ Jupyter started in the background (PID 48211, port 8888)!
Open in your browser (VPN on):
  http://10.254.0.10:8888/lab?token=...
```

The folder you choose is where Jupyter's file browser starts and the highest
folder it shows. It's not a security boundary: Jupyter's terminal and notebooks
run as your account and can reach every file you can. See
[Sharing and security](#sharing-and-security).

Each server gets the first free port from 8888 up. The server is shared, so
someone else may already be on 8888; the manager checks and moves on.

## Picking a folder

The folder list shows `(not found)` next to folders that don't exist on this
machine, and won't let you choose them. With **Custom path…** you type a path:

- `~` means your home folder; relative paths start from where you ran the manager
- **Tab** completes folder names, like in the shell
- Enter checks that the folder exists before going on

## Starting CARTA

**CARTA · Start server** asks for the folder and the token, like Jupyter, and
prints the URL. CARTA's ports start at 3002.

The folder is only where CARTA's file browser opens. It can still go up to the
folder set in `CARTA_TOP` (`/` by default); see [Configuration](configuration.md).

## Active instances

**Jupyter · Active instances** and **CARTA · Active instances** list the
running servers with their folder and full URL, token included. The main menu
also shows how many are running, e.g. `(2 running)`.

Jupyter servers that weren't started by this manager (by hand, or by the bash
manager used before 1.0.0) are listed too, with the output of `jupyter server list`.

## Logs

**Jupyter · Logs** and **CARTA · Logs** list that tool's log files, newest
first, marking the ones whose server is still running. Choose one to see its
last 30 lines. When a server fails to start, the manager shows the end of its
log right away.

## DS9 and CASA

DS9 windows and CASA's logger and plots are desktop windows, so they need X11
forwarding: the window runs on the server and is drawn on your computer.

On a Mac, install [XQuartz](https://www.xquartz.org), log out and back in, then
connect with:

```bash
ssh -Y <user>@10.254.0.10
```

On Linux, `ssh -Y` works as is. On Windows, use an X server such as VcXsrv or
MobaXterm.

**DS9 · Open window** asks for a folder and then for a FITS file: the `.fits`,
`.fit`, `.fts`, `.fits.gz` and `.fz` files in that folder are listed, plus
**No file** and **Other file…**. DS9 opens in its own window and keeps running
in the background; its File dialog starts in the folder you chose.

**CASA · Start in this terminal** runs CASA in the current terminal, in the
folder you chose (CASA writes its `casa-*.log` files there). Type `exit` to come
back to the menu. Without X11, CASA starts with `--nologger --nogui`: tasks and
tools work, only the windows are missing.

## Stopping

**Manage · Stop an instance** lists everything that's running and stops the one
you choose:

```text
┃ Instance to stop
┃ > jupyter  PID 48211   port 8888  ~/Tareas/T5/data
┃   carta    PID 48302   port 3002  /data
┃   ds9      PID 48390   ~/data  m31.fits
┃   Cancel
```

**Manage · Stop all instances** stops everything, or all the servers of one kind.

Jupyter gets a normal shutdown signal so it can close its kernels. CARTA and DS9
are stopped together with any processes they started.

## Sharing and security

A Jupyter URL with its token is as good as your password. Whoever opens it gets
a terminal and notebooks running as you: they can read, change and delete
anything in your account, whichever folder you started Jupyter in. A CARTA URL
gives access to the images CARTA can open.

- **Don't share Jupyter URLs**, not even for a folder like `~/shared/T5`.
- **Use a long random token.** The servers can be reached from the whole
  network, so a word like `Capybara` can be guessed. The manager warns about
  tokens shorter than 16 characters. To make one:
  `python3 -c "import secrets; print(secrets.token_urlsafe(24))"`
- **To work on files with classmates**, each person runs their own server with
  their own account, and the shared folder is given access through Unix groups
  or ACLs. Ask the server's administrators to set this up.

The manager keeps tokens away from other users of the server. It passes the
token to Jupyter through the environment (command lines are visible to every
user with `ps`), and keeps `~/.astrolab`, which holds the logs and URLs,
readable only by you.

## Command line

```bash
astroenvs --version   # print the version
astroenvs --help      # short usage
```
