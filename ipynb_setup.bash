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

echo "=== 4. ipynb virtual environment setup (using Python 3.11) ==="
uv venv --python 3.11 ipynb
source ipynb/bin/activate

echo "=== 5. Installing package and ipykernel into virtual environment ==="
uv pip install ipykernel
uv pip install "git+https://github.com/jj-sm/ipynb-jjsm-tools.git[full]"

# Register virtual environment as a Jupyter kernel
python -m ipykernel install --user --name=ipynb --display-name "Python (ipynb venv)"

echo "=== 6. Packages installed in ipynb virtual environment ==="
uv pip list

echo "=== 7. Installing TinyTeX (Local LaTeX distribution - No sudo required!) ==="
if [ ! -f "$HOME/.TinyTeX/bin/x86_64-linux/tlmgr" ] && [ ! -f "$HOME/.TinyTeX/bin/linux/tlmgr" ]; then
    echo "Downloading and installing TinyTeX..."
    set -o pipefail
    if ! wget -qO- "https://yihui.org/tinytex/install-unix.sh" | sh; then
        echo "⚠ TinyTeX download/install failed."
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

echo "=== 9. Copying manager scripts and setting up aliases ==="
# Copy manager scripts from repo folder to home root if they exist
if [ -f "ipynb.sh" ]; then
    cp ipynb.sh ~/.ipynb.sh
    chmod +x ~/.ipynb.sh
elif [ -f "astrolab-setup/ipynb.sh" ]; then
    cp astrolab-setup/ipynb.sh ~/.ipynb.sh
    chmod +x ~/.ipynb.sh
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
    
    # Add Aliases if missing
    if ! grep -q "alias ipynb=" "$PROFILE"; then
        echo "" >> "$PROFILE"
        echo "# IPYNB & Folder Manager Aliases" >> "$PROFILE"
        echo 'alias ipynb="bash ~/.ipynb.sh"' >> "$PROFILE"
        echo 'alias tfold="bash ~/.folder.sh"' >> "$PROFILE"
        echo "✔ Added aliases to $PROFILE"
    else
        echo "✔ Aliases already present in $PROFILE"
    fi
done

echo "=== Done! Run 'source ~/.bashrc' or restart your terminal to start using 'ipynb' and 'tfold'. ==="