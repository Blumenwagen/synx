# SYNX - Dotfile Backup System

A powerful Fish shell tool for managing dotfiles with git-based version control, machine-specific exclusions, and full system bootstrapping.

> [!WARNING]
> Synx is specifically designed for **Arch Linux** and Arch-based distributions. It relies on `pacman` and AUR helpers (like `paru` or `yay`) for package management and bootstrapping. Usage on other distributions is not supported and may not work as intended.

## Features

✅ Commit & push to GitHub    
✅ Symlink preservation  
✅ Dynamic target management  
✅ Git history & time-travel rollback  
✅ Exclude patterns for machine-specific configs  
✅ Auto-reloads Hyprland  
✅ System bootstrap — provision a new machine in one command  

## Quick Start

### Installation

```bash
git clone https://github.com/Blumenwagen/synx.git
cd synx
./install.sh
```

The installer will:
- Install `synx` command to `~/.local/bin/`
- Create config files in `~/.config/synx/`
- Optionally initialize your dotfiles repo

### Basic Usage

```fish
synx                       # Copy and Sync to GitHub
synx --restore             # Restore from GitHub
synx --list                # List tracked dotfiles
synx --history             # View sync history
```

## Commands

| Command | Description |
|---------|-------------|
| `synx` | Sync dotfiles to GitHub |
| `synx --restore` | Restore dotfiles from GitHub |
| `synx --add <name>` | Track a new dotfile directory |
| `synx --remove <name>` | Stop tracking a dotfile directory |
| `synx --exclude <path>` | Exclude machine-specific file |
| `synx --list` | List tracked/available dotfiles |
| `synx --history` | Show commit history |
| `synx --rollback <n>` | Rollback n commits |
| `synx --bootstrap-setup` | Create bootstrap config interactively |
| `synx --bootstrap` | Run bootstrap from local config |
| `synx --bootstrap <url>` | Clone repo & bootstrap from its config |
| `synx --help` | Show help |

## Bootstrap — Full System Setup

Set up a new machine in one command using a bootstrap config that syncs with your dotfiles.

### Create a Bootstrap Config

```fish
synx --bootstrap-setup
```

The wizard walks you through:
1. **AUR Helper** — detect/choose paru, yay
2. **Packages** — choose packages to install on bootstrap
3. **Git Repos** — repos to clone with optional install scripts
4. **Dotfile Restore** — auto-restore dotfiles after setup
5. **Custom Commands** — post-install commands (chsh, systemctl, etc.)

### Bootstrap a New Machine

```fish
synx --bootstrap https://github.com/user/dotfiles.git
```

This will:
1. Clone your dotfiles repo
2. Find the bootstrap config inside it
3. **Show it to you for review/editing** before running anything
4. Execute each step with per-step confirmations

### Bootstrap Config Example

```ini
[aur]
helper = paru

[packages]
list = firefox hyprland waybar kitty rofi-wayland

[repos]
repo = https://github.com/user/project.git | ~/project | ./install.sh

[dotfiles]
restore = true

[commands]
run = chsh -s /usr/bin/fish
```

## Multi-Machine Setup

Use the same dotfiles repo on multiple machines while keeping machine-specific configs local:

```fish
# Exclude machine-specific files
synx --exclude hypr/monitors.conf
synx --exclude hypr/workspaces.conf
```

## Configuration

**Tracked dotfiles:** `~/.config/synx/synx.conf`  
**Exclude patterns:** `~/.config/synx/exclude.conf`  
**Bootstrap config:** `~/.config/synx/bootstrap.conf`  
**Dotfiles repo:** `~/dotfiles/`

## Requirements

- Fish shell (can be installed by the synx install-script)
- Git

