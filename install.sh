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

# Check git
if ! command -v git &> /dev/null; then
    echo -e "  ${RED}✗${RESET} git not found"
    echo -e "  ${YELLOW}⚠${RESET}  Please install git first: ${CYAN}sudo pacman -S git${RESET}"
    exit 1
else
    echo -e "  ${GREEN}✓${RESET} git"
fi

# Check fish — offer to install if missing
if ! command -v fish &> /dev/null; then
    echo -e "  ${RED}✗${RESET} fish shell not found"
    echo
    read -p "  Install fish shell? [Y/n] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        echo -e "  ${BLUE}⠿${RESET} Installing fish..."
        if sudo pacman -S --needed --noconfirm fish; then
            echo -e "  ${GREEN}✓${RESET} fish shell installed"
        else
            echo -e "  ${RED}✗${RESET} Failed to install fish"
            exit 1
        fi
    else
        echo -e "${RED}✗${RESET} fish shell is required for synx."
        exit 1
    fi
else
    echo -e "  ${GREEN}✓${RESET} fish shell"
fi

# Offer to set fish as default shell
CURRENT_SHELL=$(basename "$SHELL")
if [ "$CURRENT_SHELL" != "fish" ]; then
    echo
    echo -e "  ${YELLOW}○${RESET} Current default shell: ${BOLD}$CURRENT_SHELL${RESET}"
    read -p "  Set fish as your default shell? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        FISH_PATH=$(which fish)
        # Ensure fish is in /etc/shells
        if ! grep -q "$FISH_PATH" /etc/shells 2>/dev/null; then
            echo -e "  ${BLUE}⠿${RESET} Adding fish to /etc/shells..."
            echo "$FISH_PATH" | sudo tee -a /etc/shells > /dev/null
        fi
        if chsh -s "$FISH_PATH"; then
            echo -e "  ${GREEN}✓${RESET} Default shell set to fish"
            echo -e "  ${YELLOW}⚠${RESET}  Log out and back in for this to take effect"
        else
            echo -e "  ${RED}✗${RESET} Failed to change shell (you can run ${CYAN}chsh -s $FISH_PATH${RESET} manually)"
        fi
    fi
fi

echo

# Install synx command
echo -e "${BLUE}→${RESET} Installing synx command..."
echo

INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"

# Use symlink so updates are instant
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [ -L "$INSTALL_DIR/synx" ] || [ -f "$INSTALL_DIR/synx" ]; then
    rm -f "$INSTALL_DIR/synx"
fi
ln -s "$SCRIPT_DIR/synx" "$INSTALL_DIR/synx"
chmod +x "$SCRIPT_DIR/synx"

echo -e "  ${GREEN}✓${RESET} Linked to $INSTALL_DIR/synx"

# Add ~/.local/bin to fish PATH
echo -e "  ${BLUE}⠿${RESET} Configuring fish PATH..."
fish -c "fish_add_path -g $INSTALL_DIR" 2>/dev/null
if [ $? -eq 0 ]; then
    echo -e "  ${GREEN}✓${RESET} Added $INSTALL_DIR to fish PATH"
else
    echo -e "  ${YELLOW}⚠${RESET}  Could not add to PATH automatically"
    echo -e "  Run this in fish: ${CYAN}fish_add_path ~/.local/bin${RESET}"
fi

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
read -p "  Initialize dotfiles repo in ~/dotfiles? [y/N] " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -d ~/dotfiles/.git ]; then
        echo -e "  ${YELLOW}○${RESET} Repository already exists, skipping"
    else
        mkdir -p ~/dotfiles
        cd ~/dotfiles
        git init
        echo -e "  ${GREEN}✓${RESET} Initialized git repo in ~/dotfiles"
        echo
        read -p "  Add GitHub remote? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            read -p "  GitHub repo URL: " REPO_URL
            if [ -n "$REPO_URL" ]; then
                git remote add origin "$REPO_URL"
                echo -e "  ${GREEN}✓${RESET} Added remote: $REPO_URL"
            fi
        fi
    fi
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
