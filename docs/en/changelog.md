# Changelog

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
