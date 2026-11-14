#!/usr/bin/env bash
# One-command install:
#   curl -fsSL https://raw.githubusercontent.com/jj-sm/astrolab-setup/main/install.sh | bash
# Downloads the repository into a temporary folder, asks for a token, runs ipynb_setup.bash
# and deletes the folder: the setup copies everything it needs into your home folder.
#
# Everything is inside main, so bash has read the whole script before running anything
# and no command can swallow the rest of it from the curl pipe.
set -e

# ASTROLAB_TARBALL can point somewhere else (another branch, or a file:// URL for testing)
TARBALL_URL="${ASTROLAB_TARBALL:-https://github.com/jj-sm/astrolab-setup/archive/refs/heads/main.tar.gz}"

has_token() {
    grep -q '^[[:space:]]*\(export[[:space:]]\+\)\?TOKEN=.' "$1" 2>/dev/null &&
        ! grep -q '^[[:space:]]*\(export[[:space:]]\+\)\?TOKEN=["'\'']\?your_token' "$1"
}

random_token() {
    LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 24
}

# Asks on the terminal (stdin is the curl pipe); with no terminal, a random token is used
ask_token() {
    local token=""
    # -r is true even with no controlling terminal, so try opening it
    if { : </dev/tty; } 2>/dev/null; then
        while true; do
            printf "Choose a token to protect your Jupyter and CARTA servers\n" >/dev/tty
            printf "(letters, digits and . _ ~ - only; Enter for a random one): " >/dev/tty
            IFS= read -r token </dev/tty || token=""
            [ -z "$token" ] && break
            if printf '%s' "$token" | grep -q '^[A-Za-z0-9._~-]\+$'; then
                break
            fi
            echo "✖ That token has other characters; try again." >/dev/tty
        done
    fi
    [ -z "$token" ] && token=$(random_token)
    printf '%s' "$token"
}

main() {
    echo "=== Downloading astrolab-setup ==="
    # mktemp -d makes it private (700), which matters since the token is written there
    WORK_DIR=$(mktemp -d "${TMPDIR:-/tmp}/astrolab-setup.XXXXXX")
    trap 'rm -rf "$WORK_DIR"' EXIT
    # Some tars accept an empty stream, so check for the setup script rather than the exit code
    curl -fsSL "$TARBALL_URL" | tar -xz -C "$WORK_DIR" --strip-components=1 || true
    if [ ! -f "$WORK_DIR/ipynb_setup.bash" ]; then
        echo "✖ Could not download $TARBALL_URL (is GitHub reachable from here?)"
        exit 1
    fi

    # Keep the settings of an earlier install (or of an old ~/astrolab-setup clone);
    # otherwise ask for a token
    local env="$WORK_DIR/.env"
    if has_token "$HOME/.astroenvs.env"; then
        cp "$HOME/.astroenvs.env" "$env"
        echo "✔ Using the token in ~/.astroenvs.env"
    elif has_token "$HOME/astrolab-setup/.env"; then
        cp "$HOME/astrolab-setup/.env" "$env"
        echo "✔ Using the token in ~/astrolab-setup/.env"
    else
        local token
        token=$(ask_token)
        echo "TOKEN=$token" > "$env"
        echo "✔ Token saved (you'll find it in ~/.astroenvs.env): $token"
    fi

    bash "$WORK_DIR/ipynb_setup.bash"

    echo ""
    echo "=== All set. Close and reopen your terminal (or run 'source ~/.bashrc'), then run: astroenvs ==="
}

main "$@"
