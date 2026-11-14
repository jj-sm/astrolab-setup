#!/usr/bin/env bash
# One-command install:
#   curl -fsSL https://raw.githubusercontent.com/jj-sm/astrolab-setup/main/install.sh | bash
# Clones (or updates) ~/astrolab-setup, asks for a token, then runs ipynb_setup.bash.
#
# Everything is inside main, so bash has read the whole script before running anything
# and no command can swallow the rest of it from the curl pipe.
set -e

REPO_URL="https://github.com/jj-sm/astrolab-setup.git"
REPO_DIR="$HOME/astrolab-setup"

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
    echo "=== Getting astrolab-setup into $REPO_DIR ==="
    if [ -d "$REPO_DIR/.git" ]; then
        git -C "$REPO_DIR" pull --ff-only
    elif [ -e "$REPO_DIR" ]; then
        echo "✖ $REPO_DIR exists but isn't a git clone. Move it away and run this again."
        exit 1
    else
        git clone "$REPO_URL" "$REPO_DIR"
    fi

    # Keep a token set by an earlier install; otherwise ask for one
    local env="$REPO_DIR/.env"
    if has_token "$env"; then
        echo "✔ Using the token in $env"
    elif has_token "$HOME/.astroenvs.env"; then
        cp "$HOME/.astroenvs.env" "$env"
        echo "✔ Using the token in ~/.astroenvs.env"
    else
        local token
        token=$(ask_token)
        if [ -f "$env" ]; then
            grep -v '^[[:space:]]*\(export[[:space:]]\+\)\?TOKEN=' "$env" > "$env.tmp" || true
            mv "$env.tmp" "$env"
        fi
        echo "TOKEN=$token" >> "$env"
        echo "✔ Token saved (you'll find it in ~/.astroenvs.env): $token"
    fi
    chmod 600 "$env"

    bash "$REPO_DIR/ipynb_setup.bash"

    echo ""
    echo "=== All set. Close and reopen your terminal (or run 'source ~/.bashrc'), then run: astroenvs ==="
}

main "$@"
