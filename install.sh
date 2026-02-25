#!/bin/bash
# SYNX Installer
# Install the synx dotfile backup system

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

echo
echo -e "${CYAN}${BOLD}╭─────────────────────────────────────╮${RESET}"
echo -e "${CYAN}${BOLD}│${RESET}  📦 ${BOLD}SYNX${RESET} - Installer          ${CYAN}${BOLD}│${RESET}"
echo -e "${CYAN}${BOLD}╰─────────────────────────────────────╯${RESET}"
echo

# Check dependencies
echo -e "${BLUE}→${RESET} Checking dependencies..."
echo

install_pkg() {
    local pkg=$1
    if command -v pacman &> /dev/null; then
        sudo pacman -S --needed --noconfirm "$pkg"
    elif command -v apt-get &> /dev/null; then
        sudo apt-get update -y && sudo apt-get install -y "$pkg"
    elif command -v dnf &> /dev/null; then
        sudo dnf install -y "$pkg"
    elif command -v zypper &> /dev/null; then
        sudo zypper install -y "$pkg"
    elif command -v brew &> /dev/null; then
        brew install "$pkg"
    else
        return 1
    fi
}

# Check git
if ! command -v git &> /dev/null; then
    echo -e "  ${RED}✗${RESET} git not found"
    echo -e "  ${YELLOW}⚠${RESET}  Please install git using your system's package manager."
    exit 1
else
    echo -e "  ${GREEN}✓${RESET} git"
fi

# Check go
if ! command -v go &> /dev/null; then
    echo -e "  ${YELLOW}⚠${RESET}  go not found, attempting to install..."
    if install_pkg go || install_pkg golang; then
        echo -e "  ${GREEN}✓${RESET} go (installed)"
    else
        echo -e "  ${RED}✗${RESET} Failed to install go automatically"
        echo -e "  ${YELLOW}⚠${RESET}  Please install go (golang) manually."
        exit 1
    fi
else
    echo -e "  ${GREEN}✓${RESET} go"
fi


# Install synx command
echo -e "${BLUE}→${RESET} Building and Installing synx command..."
echo

# Build go binary
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR/synx-go"
echo -e "  ${BLUE}⠿${RESET} Compiling go binary..."
go build -o synx main.go
if [ $? -ne 0 ]; then
    echo -e "  ${RED}✗${RESET} Failed to build synx-go"
    exit 1
fi
echo -e "  ${GREEN}✓${RESET} Build complete"

# Symlink to /bin so synx is available system-wide
echo -e "  ${BLUE}⠿${RESET} Linking to /bin/synx..."
sudo ln -sf "$SCRIPT_DIR/synx-go/synx" /bin/synx
echo -e "  ${GREEN}✓${RESET} Linked to /bin/synx"

echo

# Create config directory
echo -e "${BLUE}→${RESET} Setting up configuration..."
echo

mkdir -p ~/.config/synx

# Create default config if it doesn't exist
if [ ! -f ~/.config/synx/synx.conf ]; then
    cat > ~/.config/synx/synx.conf << 'EOF'
# Synx tracked dotfiles
# One per line, relative to ~/.config/

hypr
foot
kitty
fastfetch
alacritty
EOF
    echo -e "  ${GREEN}✓${RESET} Created ~/.config/synx/synx.conf"
else
    echo -e "  ${YELLOW}○${RESET} Config already exists, keeping it"
fi

# Create exclude config if it doesn't exist
if [ ! -f ~/.config/synx/exclude.conf ]; then
    cat > ~/.config/synx/exclude.conf << 'EOF'
# Exclude patterns for machine-specific files
# One pattern per line
#
# Examples:
#   hypr/monitors.conf
#   hypr/workspaces.conf
#   */local.conf

EOF
    echo -e "  ${GREEN}✓${RESET} Created ~/.config/synx/exclude.conf"
else
    echo -e "  ${YELLOW}○${RESET} Exclude config already exists, keeping it"
fi

echo

# Ask about dotfiles repo setup
echo -e "${BLUE}→${RESET} Dotfiles repository setup"
echo

if [ -d ~/dotfiles/.git ]; then
    echo -e "  ${YELLOW}○${RESET} Repository already exists at ~/dotfiles, skipping"
else
    echo -e "  ${BOLD}Choose an option:${RESET}"
    echo -e "    ${CYAN}1)${RESET} Initialize a new repo in ~/dotfiles"
    echo -e "    ${CYAN}2)${RESET} Clone an existing repo to ~/dotfiles"
    echo -e "    ${CYAN}3)${RESET} Skip"
    echo
    read -p "  Enter choice [1/2/3]: " -n 1 -r REPO_CHOICE
    echo

    case "$REPO_CHOICE" in
        1)
            mkdir -p ~/dotfiles
            cd ~/dotfiles
            git init
            echo -e "  ${GREEN}✓${RESET} Initialized git repo in ~/dotfiles"
            echo
            read -p "  Add remote? [y/N] " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                read -p "  Remote URL: " REPO_URL
                if [ -n "$REPO_URL" ]; then
                    git remote add origin "$REPO_URL"
                    echo -e "  ${GREEN}✓${RESET} Added remote: $REPO_URL"
                fi
            fi
            ;;
        2)
            read -p "  Remote URL to clone: " CLONE_URL
            if [ -n "$CLONE_URL" ]; then
                git clone "$CLONE_URL" ~/dotfiles && CLONE_OK=1 || CLONE_OK=0
                if [ "$CLONE_OK" -eq 1 ]; then
                    echo -e "  ${GREEN}✓${RESET} Cloned to ~/dotfiles"
                else
                    echo -e "  ${RED}✗${RESET} Failed to clone repository"
                fi
            else
                echo -e "  ${YELLOW}⚠${RESET}  No URL provided, skipping"
            fi
            ;;
        *)
            echo -e "  ${YELLOW}○${RESET} Skipped dotfiles repo setup"
            ;;
    esac
fi

echo
echo -e "${GREEN}${BOLD}✓ Installation complete!${RESET}"
echo
echo -e "${BOLD}Quick Start:${RESET}"
echo -e "  ${CYAN}synx --help${RESET}              Show all commands"
echo -e "  ${CYAN}synx --list${RESET}              List tracked dotfiles"
echo -e "  ${CYAN}synx${RESET}                     Sync to GitHub"
echo -e "  ${CYAN}synx --restore${RESET}           Restore from GitHub"
echo -e "  ${CYAN}synx --bootstrap-setup${RESET}   Create bootstrap config"
echo
echo -e "${BOLD}Configuration:${RESET}"
echo -e "  Tracked files: ${CYAN}~/.config/synx/synx.conf${RESET}"
echo -e "  Excludes:      ${CYAN}~/.config/synx/exclude.conf${RESET}"
echo -e "  Bootstrap:     ${CYAN}~/.config/synx/bootstrap.conf${RESET}"
echo
