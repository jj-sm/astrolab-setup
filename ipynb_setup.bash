#!/usr/bin/env bash
set -e

echo "=== 1. Installing uv (Astral's fast Python/tool manager) ==="
curl -LsSf https://astral.sh/uv/install.sh | sh

# Ensure uv is available in current PATH
export PATH="$HOME/.cargo/bin:$PATH"

echo "=== 2. Installing Python 3.11 and Python 3.13 via uv (No root required!) ==="
uv python install 3.11
uv python install 3.13

echo "=== 3. Verifying Python & uv Installations ==="
uv python find 3.11
uv python find 3.13
uv --version

echo "=== 4. ipynb virtual environment setup (using Python 3.13) ==="
# Always in the home folder, where astroenvs looks for it, wherever this script is run from
IPYNB_VENV="$HOME/.ipynb"
if [ -x "$IPYNB_VENV/bin/python" ]; then
    # Re-running the setup keeps the existing environment and the packages installed in it
    echo "Reusing the existing virtual environment at $IPYNB_VENV ($("$IPYNB_VENV/bin/python" --version))"
elif [ -e "$IPYNB_VENV" ] && [ ! -f "$IPYNB_VENV/pyvenv.cfg" ]; then
    # Not a virtual environment: never delete it, it may hold someone's files
    echo "✖ $IPYNB_VENV exists but isn't a virtual environment."
    echo "  Move or delete it, then run the setup again."
    exit 1
else
    # --clear replaces a broken environment instead of stopping at uv's "already exists" prompt
    uv venv --python 3.13 --clear "$IPYNB_VENV"
fi
source "$IPYNB_VENV/bin/activate"

echo "=== 5. Installing package and ipykernel into virtual environment ==="
uv pip install ipykernel
uv pip install "git+https://github.com/jj-sm/ipynb-jjsm-tools.git[full]"

# Register virtual environment as a Jupyter kernel
python -m ipykernel install --user --name=ipynb --display-name "Python (ipynb 3.13)"

echo "=== 6. Packages installed in ipynb virtual environment ==="
uv pip list

echo "=== 7. Installing TinyTeX (Local LaTeX distribution - No sudo required!) ==="
if [ ! -f "$HOME/.TinyTeX/bin/x86_64-linux/tlmgr" ] && [ ! -f "$HOME/.TinyTeX/bin/linux/tlmgr" ]; then
    echo "Downloading and installing TinyTeX..."
    set -o pipefail
    if ! wget -qO- "https://yihui.org/tinytex/install-unix.sh" | sh; then
        echo "  TinyTeX download/install failed."
        echo "  This usually means the cluster's network blocks yihui.org."
        echo "  Ask course staff whether outbound access to that domain is allowed,"
        echo "  or whether a pre-downloaded TinyTeX bundle is provided on the cluster."
    fi
    set +o pipefail
else
    echo "TinyTeX is already installed and configured."
fi

if [ -n "$TINYTEXT_BIN" ] && [ -f "$TINYTEXT_BIN/tlmgr" ]; then
    export PATH="$TINYTEXT_BIN:$PATH"
    echo "=== 8. Installing common LaTeX packages for Matplotlib usetex & PDF generation ==="
    tlmgr option repository ctan
    tlmgr update --self
    tlmgr install latex-bin fancyhdr geometry graphics amsmath cm-super dvipng
else
    echo "=== Warning: tlmgr binary not found automatically. You can install packages later via: ==="
    echo "tlmgr install latex-bin fancyhdr geometry graphics amsmath cm-super dvipng"
fi

echo "=== 9. Installing the astroenvs manager and setting up aliases ==="
# The manager is a Go binary published as a GitHub Release (see .github/workflows/build.yml)
ASTROENVS_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$(uname -m)" in
    x86_64|amd64) ASTROENVS_ARCH=amd64 ;;
    aarch64|arm64) ASTROENVS_ARCH=arm64 ;;
    *) ASTROENVS_ARCH=$(uname -m) ;;
esac
ASTROENVS_URL="https://github.com/jj-sm/astrolab-setup/releases/latest/download/astroenvs-$ASTROENVS_OS-$ASTROENVS_ARCH"
ASTROENVS_CMD="$HOME/.local/bin/astroenvs"
ASTROENVS_FAILED=""
mkdir -p ~/.local/bin
if curl -fsSL "$ASTROENVS_URL" -o ~/.local/bin/astroenvs.tmp; then
    chmod +x ~/.local/bin/astroenvs.tmp
    mv ~/.local/bin/astroenvs.tmp ~/.local/bin/astroenvs
    echo "✔ Installed $(~/.local/bin/astroenvs --version) to ~/.local/bin/astroenvs"
else
    # The rest of step 9 still runs, so downloading the binary by hand later is enough
    rm -f ~/.local/bin/astroenvs.tmp
    ASTROENVS_FAILED=1
    echo "✖ Could not download $ASTROENVS_URL"
fi

# The manager reads TOKEN from ~/.astroenvs.env (private: it grants access to your servers)
if [ -f ".env" ]; then
    cp .env ~/.astroenvs.env
    chmod 600 ~/.astroenvs.env
elif [ -f "astrolab-setup/.env" ]; then
    cp astrolab-setup/.env ~/.astroenvs.env
    chmod 600 ~/.astroenvs.env
else
    echo "⚠ No .env found: create ~/.astroenvs.env with TOKEN=<your-token> before running astroenvs"
fi

if [ -f "folder.bash" ]; then
    cp folder.bash ~/.folder.sh
    chmod +x ~/.folder.sh
elif [ -f "astrolab-setup/folder.bash" ]; then
    cp astrolab-setup/folder.bash ~/.folder.sh
    chmod +x ~/.folder.sh
fi

# Add TinyTeX PATH and Aliases to all profile files (.bashrc, .bash_profile, .profile)
for PROFILE in "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.profile"; do
    touch "$PROFILE"
    
    # Add TinyTeX PATH if missing
    if ! grep -q ".TinyTeX/bin" "$PROFILE"; then
        echo "" >> "$PROFILE"
        echo "# TinyTeX PATH" >> "$PROFILE"
        echo 'export PATH="$HOME/.TinyTeX/bin/x86_64-linux:$HOME/.TinyTeX/bin/linux:$PATH"' >> "$PROFILE"
    fi
    
    # Aliases are rewritten every run, which also repoints the ones older installs made
    # (bash ~/.astroenvs.sh, ~/.ipynb.sh)
    grep -v -E '^(# Astroenvs \(Jupyter \+ CARTA\) & Folder Manager Aliases$|alias (astroenvs|ipynb|tfold)=)' \
        "$PROFILE" > "$PROFILE.astroenvs-tmp" || true
    cat "$PROFILE.astroenvs-tmp" > "$PROFILE"
    rm -f "$PROFILE.astroenvs-tmp"
    {
        echo "# Astroenvs (Jupyter + CARTA) & Folder Manager Aliases"
        echo "alias astroenvs=\"$ASTROENVS_CMD\""
        echo "alias ipynb=\"$ASTROENVS_CMD\""
        echo 'alias tfold="bash ~/.folder.sh"'
    } >> "$PROFILE"
    echo "✔ Aliases point to $ASTROENVS_CMD in $PROFILE"
done

touch ~/.root_ident

if [ -n "$ASTROENVS_FAILED" ]; then
    echo "=== Done, except the astroenvs manager: its download failed (is GitHub reachable from here?). ==="
    echo "Everything else is installed. Download it by hand, then run 'source ~/.bashrc':"
    echo "  curl -fsSL $ASTROENVS_URL -o ~/.local/bin/astroenvs && chmod +x ~/.local/bin/astroenvs"
    exit 1
fi

echo "=== Done! Run 'source ~/.bashrc' or restart your terminal to start using 'astroenvs' (or 'ipynb') and 'tfold'. ==="