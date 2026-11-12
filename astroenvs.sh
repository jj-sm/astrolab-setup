#!/usr/bin/env bash

# Absolute path to the ipynb venv (adjust if you move it)
IPYNB_VENV="$HOME/.ipynb"

# Executables (override with e.g.: export CARTA_BIN=/path/to/carta)
CARTA_BIN="${CARTA_BIN:-$(command -v carta)}"
DS9_BIN="${DS9_BIN:-$(command -v ds9)}"
CASA_BIN="${CASA_BIN:-$(command -v casa || echo /usr/local/casa/bin/casa)}"

# CARTA's file browser can't go above this folder
CARTA_TOP="/"
# First CARTA port to try; the next free one is used (the host is shared)
CARTA_PORT_START=3002
# Folders offered as the starting point for CARTA, DS9 and CASA
DATA_DIRS=(
    "/data"
    "$HOME/UC-AST221-Rsync"
)

# Logs and instance tracking live here, so the manager works from any directory
STATE_DIR="$HOME/.astrolab"
LOG_DIR="$STATE_DIR/logs"
RUN_DIR="$STATE_DIR/run"
JUPYTER_LOG="$LOG_DIR/jupyter.log"
mkdir -p "$LOG_DIR" "$RUN_DIR"

SERVER_IP=$(hostname -I 2>/dev/null | awk '{print $1}')
SERVER_IP=${SERVER_IP:-$(hostname)}

# Colors
C_RESET='\033[0m'
C_BOLD='\033[1m'
C_CYAN='\033[36m'
C_GREEN='\033[32m'
C_YELLOW='\033[33m'
C_RED='\033[31m'
C_PURPLE='\033[35m'
C_BLUE='\033[34m'

# TOKEN (shared by Jupyter and CARTA) comes from the first .env file found
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
for ENV_FILE in "$ASTROENVS_ENV" "$HOME/.astroenvs.env" "$SCRIPT_DIR/.env"; do
    [ -n "$ENV_FILE" ] && [ -f "$ENV_FILE" ] && break
    ENV_FILE=""
done
TOKEN=""
if [ -n "$ENV_FILE" ]; then
    TOKEN=$(grep -E '^[[:space:]]*(export[[:space:]]+)?TOKEN=' "$ENV_FILE" | tail -n 1 | cut -d= -f2- | tr -d "\"' \r\t")
fi
if [ -z "$TOKEN" ]; then
    echo -e "${C_RED}✖ No TOKEN found.${C_RESET} Add a line TOKEN=<your-token> to one of:"
    echo -e "  ~/.astroenvs.env  (created by ipynb_setup.bash from the repo's .env)"
    echo -e "  $SCRIPT_DIR/.env"
    echo -e "  or point ASTROENVS_ENV at another file."
    exit 1
fi

pause() {
    read -r -p "Press [Enter] to return to menu..."
}

port_in_use() {
    ss -tln 2>/dev/null | awk '{print $4}' | grep -qE "[:.]$1\$"
}

free_port() {
    local p=$1
    while port_in_use "$p"; do p=$((p + 1)); done
    echo "$p"
}

# Prints "id pid info" for each live tracked instance of a kind (carta, ds9) and drops stale entries
instances() {
    local f id pid info
    for f in "$RUN_DIR/$1"-*; do
        [ -e "$f" ] || continue
        id=${f##*/"$1"-}
        read -r pid info < "$f"
        if kill -0 "$pid" 2>/dev/null; then
            echo "$id $pid $info"
        else
            rm -f "$f"
        fi
    done
}

# Tracked instances run in their own process group (setsid), so stop the whole group
stop_tracked_pid() {
    kill -TERM -- "-$1" 2>/dev/null || kill -TERM "$1" 2>/dev/null
}

# Asks for a starting folder and sets PICKED_DIR; returns 1 if the choice is invalid
pick_dir() {
    local i pick
    echo "Start in:"
    for i in "${!DATA_DIRS[@]}"; do
        echo -e "${C_GREEN} $((i + 1)))${C_RESET} ${DATA_DIRS[$i]}"
    done
    echo -e "${C_GREEN} c)${C_RESET} Custom path"
    echo -n -e "${C_BOLD}Choice [1]: ${C_RESET}"
    read -r pick
    pick=${pick:-1}

    if [ "$pick" = "c" ]; then
        read -r -e -p "Path: " PICKED_DIR
        PICKED_DIR=${PICKED_DIR/#\~/$HOME}
    elif [[ "$pick" =~ ^[0-9]+$ ]] && [ "$pick" -ge 1 ] && [ "$pick" -le "${#DATA_DIRS[@]}" ]; then
        PICKED_DIR=${DATA_DIRS[$((pick - 1))]}
    else
        echo -e "${C_RED}✖ Invalid choice.${C_RESET}"
        return 1
    fi

    if [ ! -d "$PICKED_DIR" ]; then
        echo -e "${C_RED}✖ $PICKED_DIR is not a directory.${C_RESET}"
        return 1
    fi
}

# Checks that an executable exists; $1 = path, $2 = name, $3 = config variable
need_bin() {
    [ -n "$1" ] && [ -x "$1" ] && return 0
    echo -e "${C_RED}✖ $2 executable not found.${C_RESET}"
    echo -e "  Set $3 at the top of this script, or: export $3=/path/to/$2\n"
    return 1
}

# GUI apps need an X display, i.e. an SSH session with X11 forwarding
need_display() {
    [ -n "$DISPLAY" ] && return 0
    echo -e "${C_RED}✖ No X display (\$DISPLAY is empty), so $1 can't open a window.${C_RESET}"
    echo -e "  Reconnect with X11 forwarding (needs XQuartz on your Mac) and run the manager again:"
    echo -e "    ssh -Y $USER@$SERVER_IP\n"
    return 1
}

show_menu() {
    clear
    echo -e "${C_BOLD}${C_PURPLE}============================================${C_RESET}"
    echo -e "${C_BOLD}${C_CYAN}             ASTROENVS MANAGER              ${C_RESET}"
    echo -e "${C_BOLD}${C_PURPLE}============================================${C_RESET}"
    echo -e "${C_BOLD} Jupyter${C_RESET}"
    echo -e "${C_GREEN} 1)${C_RESET} Start Jupyter Server (Detached & Custom Token)"
    echo -e "${C_GREEN} 2)${C_RESET} View Active Jupyter Instances (URLs & Ports)"
    echo -e "${C_YELLOW} 3)${C_RESET} View Jupyter Logs"
    echo -e "${C_BOLD} CARTA${C_RESET}"
    echo -e "${C_GREEN} 4)${C_RESET} Start CARTA Server (Detached & Custom Token)"
    echo -e "${C_GREEN} 5)${C_RESET} View Active CARTA Instances (URLs & Ports)"
    echo -e "${C_YELLOW} 6)${C_RESET} View CARTA Logs"
    echo -e "${C_BOLD} Desktop apps (X11)${C_RESET}"
    echo -e "${C_GREEN} 7)${C_RESET} Open DS9 (Detached Window)"
    echo -e "${C_GREEN} 8)${C_RESET} Start CASA (Runs in this Terminal)"
    echo -e "${C_BOLD} Manage${C_RESET}"
    echo -e "${C_RED} 9)${C_RESET} Stop a Specific Instance"
    echo -e "${C_RED}10)${C_RESET} Stop & Clear ALL Instances"
    echo -e "${C_BLUE} 0)${C_RESET} Exit"
    echo -e "${C_BOLD}${C_PURPLE}============================================${C_RESET}"
    echo -n -e "${C_BOLD}Select an option [0-10]: ${C_RESET}"
}

# ---------------------------------------------------------------- Jupyter

start_jupyter() {
    echo -e "\n${C_CYAN}=== Starting Jupyter Server ===${C_RESET}"

    if [ ! -f "$IPYNB_VENV/bin/activate" ]; then
        echo -e "${C_RED}✖ Could not find venv at $IPYNB_VENV/bin/activate${C_RESET}"
        echo -e "  Edit IPYNB_VENV at the top of this script if your venv lives elsewhere.\n"
        pause
        return
    fi

    # shellcheck disable=SC1091
    source "$IPYNB_VENV/bin/activate"

    if ! command -v jupyter &> /dev/null; then
        echo -e "${C_RED}✖ 'jupyter' not found even after activating $IPYNB_VENV${C_RESET}"
        echo -e "  Try: source $IPYNB_VENV/bin/activate && uv pip install ipykernel\n"
        pause
        return
    fi

    nohup jupyter lab --no-browser --ip=0.0.0.0 --ServerApp.allow_remote_access=True --ServerApp.token="$TOKEN" > "$JUPYTER_LOG" 2>&1 &
    JPID=$!
    sleep 2

    if kill -0 "$JPID" 2>/dev/null; then
        echo -e "${C_GREEN}✔ Jupyter started in the background (PID $JPID)!${C_RESET}"
        echo -e "Check option 2 for URLs and ports.\n"
    else
        echo -e "${C_RED}✖ Jupyter failed to start. Last lines of $JUPYTER_LOG:${C_RESET}"
        tail -n 15 "$JUPYTER_LOG"
        echo ""
    fi
    pause
}

view_jupyter() {
    echo -e "\n${C_CYAN}=== Active Jupyter Instances ===${C_RESET}"
    # Find running jupyter lab processes owned by current user
    PIDS=$(pgrep -u "$USER" -f "jupyter-lab")

    if [ -z "$PIDS" ]; then
        echo -e "${C_YELLOW}No active Jupyter instances found.${C_RESET}\n"
    else
        if [ -x "$IPYNB_VENV/bin/jupyter" ]; then
            "$IPYNB_VENV/bin/jupyter" server list
        elif command -v jupyter &> /dev/null; then
            jupyter server list
        else
            echo "Running PIDs: $PIDS"
        fi
    fi
    echo ""
    pause
}

view_jupyter_logs() {
    echo -e "\n${C_CYAN}=== Recent Jupyter Logs ===${C_RESET}"
    if [ -f "$JUPYTER_LOG" ]; then
        tail -n 30 "$JUPYTER_LOG"
    else
        echo -e "${C_YELLOW}No log found at $JUPYTER_LOG${C_RESET}"
    fi
    echo ""
    pause
}

# ---------------------------------------------------------------- CARTA

start_carta() {
    echo -e "\n${C_CYAN}=== Starting CARTA Server ===${C_RESET}"
    need_bin "$CARTA_BIN" carta CARTA_BIN || { pause; return; }
    pick_dir || { echo ""; pause; return; }

    local port log cpid
    port=$(free_port "$CARTA_PORT_START")
    log="$LOG_DIR/carta-$port.log"

    CARTA_AUTH_TOKEN="$TOKEN" setsid nohup "$CARTA_BIN" --no_browser --port "$port" \
        --top_level_folder "$CARTA_TOP" "$PICKED_DIR" > "$log" 2>&1 < /dev/null &
    cpid=$!
    sleep 3

    if kill -0 "$cpid" 2>/dev/null; then
        echo "$cpid $PICKED_DIR" > "$RUN_DIR/carta-$port"
        echo -e "${C_GREEN}✔ CARTA started in the background (PID $cpid, port $port)!${C_RESET}"
        echo -e "Open in your browser (VPN on):"
        echo -e "  ${C_BOLD}http://$SERVER_IP:$port/?token=$TOKEN${C_RESET}\n"
    else
        echo -e "${C_RED}✖ CARTA failed to start. Last lines of $log:${C_RESET}"
        tail -n 15 "$log"
        echo ""
    fi
    pause
}

view_carta() {
    echo -e "\n${C_CYAN}=== Active CARTA Instances ===${C_RESET}"
    local found=0 port pid dir
    while read -r port pid dir; do
        found=1
        echo -e "${C_GREEN}●${C_RESET} PID $pid  port $port  start folder: $dir"
        echo -e "  ${C_BOLD}http://$SERVER_IP:$port/?token=$TOKEN${C_RESET}"
    done < <(instances carta)

    if [ "$found" -eq 0 ]; then
        echo -e "${C_YELLOW}No active CARTA instances found.${C_RESET}"
    fi
    echo ""
    pause
}

view_carta_logs() {
    echo -e "\n${C_CYAN}=== Recent CARTA Logs ===${C_RESET}"
    local latest port log
    latest=$(ls -t "$LOG_DIR"/carta-*.log 2>/dev/null | head -n 1)
    if [ -z "$latest" ]; then
        echo -e "${C_YELLOW}No CARTA logs found in $LOG_DIR${C_RESET}\n"
        pause
        return
    fi

    latest=${latest##*/carta-}
    latest=${latest%.log}
    echo -n -e "${C_BOLD}Port of the instance [$latest]: ${C_RESET}"
    read -r port
    port=${port:-$latest}
    log="$LOG_DIR/carta-$port.log"

    if [ -f "$log" ]; then
        tail -n 30 "$log"
    else
        echo -e "${C_YELLOW}No log found at $log${C_RESET}"
    fi
    echo ""
    pause
}

# ---------------------------------------------------------------- DS9 & CASA

start_ds9() {
    echo -e "\n${C_CYAN}=== Opening DS9 ===${C_RESET}"
    need_bin "$DS9_BIN" ds9 DS9_BIN || { pause; return; }
    need_display DS9 || { pause; return; }
    pick_dir || { echo ""; pause; return; }

    local fits log dpid
    read -r -e -p "FITS file to open (blank for none, relative to that folder): " fits
    fits=${fits/#\~/$HOME}
    log="$LOG_DIR/ds9-starting.log"

    # DS9's file dialog opens in its working directory
    (cd "$PICKED_DIR" && exec setsid nohup "$DS9_BIN" ${fits:+"$fits"} > "$log" 2>&1 < /dev/null) &
    dpid=$!
    sleep 2

    if kill -0 "$dpid" 2>/dev/null; then
        mv "$log" "$LOG_DIR/ds9-$dpid.log"
        echo "$dpid $PICKED_DIR" > "$RUN_DIR/ds9-$dpid"
        echo -e "${C_GREEN}✔ DS9 opened (PID $dpid). The window should appear on your screen.${C_RESET}\n"
    else
        echo -e "${C_RED}✖ DS9 failed to start. Last lines of $log:${C_RESET}"
        tail -n 15 "$log"
        echo ""
    fi
    pause
}

start_casa() {
    echo -e "\n${C_CYAN}=== Starting CASA ===${C_RESET}"
    need_bin "$CASA_BIN" casa CASA_BIN || { pause; return; }
    pick_dir || { echo ""; pause; return; }

    local args=()
    if [ -z "$DISPLAY" ]; then
        args=(--nologger --nogui)
        echo -e "${C_YELLOW}No X display: starting CASA without the logger/GUI windows (tasks and tools still work).${C_RESET}"
    fi
    echo -e "CASA logs go to $PICKED_DIR. Type ${C_BOLD}exit${C_RESET} in CASA to come back to this menu.\n"

    (cd "$PICKED_DIR" && "$CASA_BIN" "${args[@]}")
    echo ""
    pause
}

# ---------------------------------------------------------------- Manage

stop_instance() {
    echo -e "\n${C_RED}=== Active Jupyter PIDs ===${C_RESET}"
    pgrep -u "$USER" -fl "jupyter-lab"
    local id pid info
    echo -e "${C_RED}=== Active CARTA PIDs ===${C_RESET}"
    while read -r id pid info; do
        echo "$pid carta (port $id)"
    done < <(instances carta)
    echo -e "${C_RED}=== Active DS9 PIDs ===${C_RESET}"
    while read -r id pid info; do
        echo "$pid ds9 ($info)"
    done < <(instances ds9)

    echo -n -e "${C_BOLD}Enter the PID to stop: ${C_RESET}"
    read -r target_pid
    [ -z "$target_pid" ] && return

    local tracked
    tracked=$(grep -l "^$target_pid " "$RUN_DIR"/* 2>/dev/null | head -n 1)
    if [ -n "$tracked" ]; then
        stop_tracked_pid "$target_pid" && rm -f "$tracked" \
            && echo -e "${C_GREEN}✔ Stopped process $target_pid${C_RESET}" \
            || echo -e "${C_RED}✖ Failed to stop $target_pid.${C_RESET}"
    else
        kill "$target_pid" 2>/dev/null && echo -e "${C_GREEN}✔ Stopped process $target_pid${C_RESET}" || echo -e "${C_RED}✖ Failed or invalid PID.${C_RESET}"
    fi
    pause
}

clear_all() {
    echo -e "\n${C_RED}=== Stop ALL instances ===${C_RESET}"
    echo -n -e "${C_BOLD}Which? [j]upyter / [c]arta / [d]s9 / [a]ll (default a): ${C_RESET}"
    local which kind id pid info
    read -r which
    which=${which:-a}

    if [ "$which" = "j" ] || [ "$which" = "a" ]; then
        pkill -u "$USER" -f "jupyter-lab" || true
        echo -e "${C_GREEN}✔ All Jupyter server instances cleared.${C_RESET}"
    fi
    for kind in carta ds9; do
        if [ "$which" = "${kind:0:1}" ] || [ "$which" = "a" ]; then
            while read -r id pid info; do
                stop_tracked_pid "$pid"
                rm -f "$RUN_DIR/$kind-$id"
            done < <(instances "$kind")
            echo -e "${C_GREEN}✔ All $kind instances cleared.${C_RESET}"
        fi
    done
    echo ""
    pause
}

while true; do
    show_menu
    read -r choice
    case $choice in
        1) start_jupyter ;;
        2) view_jupyter ;;
        3) view_jupyter_logs ;;
        4) start_carta ;;
        5) view_carta ;;
        6) view_carta_logs ;;
        7) start_ds9 ;;
        8) start_casa ;;
        9) stop_instance ;;
        10) clear_all ;;
        0) echo -e "\n${C_CYAN}Goodbye!${C_RESET}\n"; exit 0 ;;
        *) echo -e "${C_RED}Invalid option. Try again.${C_RESET}"; sleep 1 ;;
    esac
done
