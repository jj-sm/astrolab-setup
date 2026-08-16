#!/usr/bin/env bash
set -e

echo "=== 1. Updating package lists and installing prerequisites ==="
sudo apt-get update
sudo apt-get install -y software-properties-common wget curl git build-essential

echo "=== 2. Adding deadsnakes PPA for multiple Python versions ==="
sudo add-apt-repository -y ppa:deadsnakes/ppa
sudo apt-get update

echo "=== 3. Installing Python 3.11 and Python 3.13 ==="
sudo apt-get install -y \
    python3.11 python3.11-dev python3.11-venv \
    python3.13 python3.13-dev python3.13-venv

echo "=== 4. Installing uv (Astral's fast Python package installer) ==="
curl -LsSf https://astral.sh/uv/install.sh | sh

# Ensure uv is available in the current shell session's PATH
export PATH="$HOME/.cargo/bin:$PATH"

echo "=== 5. Verifying Installations ==="
echo "Python 3.11 version:"
python3.11 --version

echo "Python 3.13 version:"
python3.13 --version

echo "uv version:"
uv --version

echo "=== 6. ipynb setup ==="
uv venv ipynb
# Note: activation in bash scripts requires sourcing the bin/activate file
source ipynb/bin/activate

echo "=== 7. Installing package and ipykernel into virtual environment ==="
uv pip install ipykernel
uv pip install "git+https://github.com/jj-sm/ipynb-jjsm-tools.git[full]"

# Register the virtual environment as a Jupyter kernel so you can select it
python -m ipykernel install --user --name=ipynb --display-name "Python (ipynb venv)"

echo "=== 8. Packages installed in ipynb virtual environment ==="
uv pip list

chmod +x ~/.ipynb.sh

echo "=== 9. Automatically registering the 'ipynb' alias in shell profiles ==="
PROFILE_FILE="$HOME/.bashrc"
if [ -n "$ZSH_VERSION" ] || [ -f "$HOME/.zshrc" ]; then
    PROFILE_FILE="$HOME/.zshrc"
fi

if ! grep -q "alias ipynb=" "$PROFILE_FILE"; then
    echo "" >> "$PROFILE_FILE"
    echo "# IPYNB Jupyter Manager Alias" >> "$PROFILE_FILE"
    echo 'alias ipynb="bash ~/.ipynb.sh"' >> "$PROFILE_FILE"
    echo 'alias tfold="bash ~/.folder.sh"' >> "$PROFILE_FILE"
    echo "✔ Added alias to $PROFILE_FILE"
else
    echo "✔ Alias 'ipynb' already exists in $PROFILE_FILE"
fi

echo "=== 10. Installing TeX"
yum install texlive texlive-latex-extra texlive-fonts-recommended dvipng cm-super

echo "=== Done! ==="