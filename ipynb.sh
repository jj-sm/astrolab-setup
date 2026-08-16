#!/usr/bin/env bash

# Absolute path to the ipynb venv (adjust if you move it)
IPYNB_VENV="$HOME/.ipynb"

# Colors
C_RESET='\033[0m'
C_BOLD='\033[1m'
C_CYAN='\033[36m'
C_GREEN='\033[32m'
C_YELLOW='\033[33m'
C_RED='\033[31m'
C_PURPLE='\033[35m'
C_BLUE='\033[34m'

show_menu() {
    clear
    echo -e "${C_BOLD}${C_PURPLE}============================================${C_RESET}"
    echo -e "${C_BOLD}${C_CYAN}           IPYNB JUPYTER MANAGER           ${C_RESET}"
    echo -e "${C_BOLD}${C_PURPLE}============================================${C_RESET}"
    echo -e "${C_GREEN} 1)${C_RESET} Start Jupyter Server (Detached & Custom Token)"
    echo -e "${C_GREEN} 2)${C_RESET} View Active Instances (URLs & Ports)"
    echo -e "${C_YELLOW} 3)${C_RESET} View Jupyter Logs"
    echo -e "${C_RED} 4)${C_RESET} Stop a Specific Jupyter Instance"
    echo -e "${C_RED} 5)${C_RESET} Stop & Clear ALL Jupyter Instances"
    echo -e "${C_BLUE} 6)${C_RESET} Exit"
    echo -e "${C_BOLD}${C_PURPLE}============================================${C_RESET}"
    echo -n -e "${C_BOLD}Select an option [1-6]: ${C_RESET}"
}

start_jupyter() {
    echo -e "\n${C_CYAN}=== Starting Jupyter Server (Token: jsanchezm2) ===${C_RESET}"

    if [ ! -f "$IPYNB_VENV/bin/activate" ]; then
        echo -e "${C_RED}✖ Could not find venv at $IPYNB_VENV/bin/activate${C_RESET}"
        echo -e "  Edit IPYNB_VENV at the top of this script if your venv lives elsewhere.\n"
        read -p "Press [Enter] to return to menu..."
        return
    fi

    # shellcheck disable=SC1091
    source "$IPYNB_VENV/bin/activate"

    if ! command -v jupyter &> /dev/null; then
        echo -e "${C_RED}✖ 'jupyter' not found even after activating $IPYNB_VENV${C_RESET}"
        echo -e "  Try: source $IPYNB_VENV/bin/activate && uv pip install ipykernel\n"
        read -p "Press [Enter] to return to menu..."
        return
    fi

    nohup jupyter lab --no-browser --ServerApp.token='jsanchezm2' > jupyter.log 2>&1 &
    JPID=$!
    sleep 2

    if kill -0 "$JPID" 2>/dev/null; then
        echo -e "${C_GREEN}✔ Jupyter started in the background (PID $JPID)!${C_RESET}"
        echo -e "Check option 2 for URLs and ports.\n"
    else
        echo -e "${C_RED}✖ Jupyter failed to start. Last lines of jupyter.log:${C_RESET}"
        tail -n 15 jupyter.log
        echo ""
    fi
    read -p "Press [Enter] to return to menu..."
}

view_instances() {
    echo -e "\n${C_CYAN}=== Active Jupyter Instances ===${C_RESET}"
    # Find running jupyter lab processes owned by current user
    PIDS=$(pgrep -u "$USER" -f "jupyter-lab")

    if [ -z "$PIDS" ]; then
        echo -e "${C_YELLOW}No active Jupyter instances found.${C_RESET}\n"
    else
        # Extract server details via jupyter server list command if available
        if command -v jupyter &> /dev/null; then
            jupyter server list
        else
            echo "Running PIDs: $PIDS"
        fi
        echo -e "\n${C_GREEN}Token for access:${C_RESET} jsanchezm2"
    fi
    echo ""
    read -p "Press [Enter] to return to menu..."
}

view_logs() {
    echo -e "\n${C_CYAN}=== Recent Jupyter Logs (Press Ctrl+C to exit logs) ===${C_RESET}"
    if [ -f "jupyter.log" ]; then
        tail -n 30 jupyter.log
    else
        echo -e "${C_YELLOW}No jupyter.log file found in current directory.${C_RESET}"
    fi
    echo ""
    read -p "Press [Enter] to return to menu..."
}

stop_instance() {
    echo -e "\n${C_RED}=== Active Jupyter PIDs ===${C_RESET}"
    pgrep -u "$USER" -fl "jupyter-lab"
    echo -n -e "${C_BOLD}Enter the PID to stop: ${C_RESET}"
    read target_pid
    if [ -n "$target_pid" ]; then
        kill "$target_pid" 2>/dev/null && echo -e "${C_GREEN}✔ Stopped process $target_pid${C_RESET}" || echo -e "${C_RED}✖ Failed or invalid PID.${C_RESET}"
    fi
    read -p "Press [Enter] to return to menu..."
}

clear_all() {
    echo -e "\n${C_RED}=== Stopping ALL Jupyter Instances ===${C_RESET}"
    pkill -u "$USER" -f "jupyter-lab" || true
    echo -e "${C_GREEN}✔ All Jupyter server instances cleared.${C_RESET}\n"
    read -p "Press [Enter] to return to menu..."
}

while true; do
    show_menu
    read choice
    case $choice in
        1) start_jupyter ;;
        2) view_instances ;;
        3) view_logs ;;
        4) stop_instance ;;
        5) clear_all ;;
        6) echo -e "\n${C_CYAN}Goodbye!${C_RESET}\n"; exit 0 ;;
        *) echo -e "${C_RED}Invalid option. Try again.${C_RESET}"; sleep 1 ;;
    esac
done