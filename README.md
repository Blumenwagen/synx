# SYNX - Dotfile Backup System

A fast CLI tool for managing dotfiles with git-based version control, multi-machine support, and full system bootstrapping. Written in Go.

> [!WARNING]
> The bootstrap feature is designed for **Arch Linux** and Arch-based distributions. It relies on `pacman` and AUR helpers (like `paru` or `yay`) for package management. Core syncing should work on any Unix-like system.

## Features

- **Sync & push** — copy dotfiles to a git repo and push to GitHub in one command
- **Symlink-safe** — follows and preserves symlinked config directories
- **Multi-machine** — per-hostname targets and excludes for different setups
- **Target management** — add, remove, and list tracked dotfiles
- **Exclude patterns** — skip machine-specific files from syncing
- **History & rollback** — view commit history, time-travel to previous states
- **Restore** — pull and restore dotfiles from remote
- **Bootstrap** — provision a new machine from a declarative config

## Quick Start

### Installation

```bash
git clone https://github.com/Blumenwagen/synx.git
cd synx
./install.sh
```

Requires **Go** and **git**. The installer builds the binary and symlinks it to `/bin/synx`.

### Basic Usage

```bash
synx                       # Sync dotfiles to GitHub
synx -r                    # Restore from GitHub
synx --list                # List tracked dotfiles
synx --history             # View sync history
```

## Commands

| Command | Description |
|---------|-------------|
| `synx` | Sync dotfiles to GitHub |
| `synx -r` / `--restore` | Restore dotfiles from GitHub |
| `synx --add <name>` | Track a new dotfile directory |
| `synx --remove <name>` | Stop tracking a dotfile directory |
| `synx --exclude <path>` | Exclude a file pattern from syncing |
| `synx --list` | List tracked and available dotfiles |
| `synx --history` | Show commit history |
| `synx --rollback <n>` | Rollback n commits and force push |
| `synx --bootstrap-setup` | Create bootstrap config interactively |
| `synx --bootstrap <url>` | Clone repo & bootstrap from its config |
| `synx --help` | Show help |

### Machine-Specific Flag

Add `-m` / `--machine` to target the current machine's config instead of the shared base:

```bash
synx --add waybar -m       # Track waybar on this machine only
synx --exclude "hypr/monitors.conf" -m   # Exclude on this machine only
synx --remove kitty -m     # Remove from this machine's targets
```

## Multi-Machine Setup

Synx auto-detects your hostname and supports per-machine config overrides:

```
~/.config/synx/
├── synx.conf              # Base targets (shared across machines)
├── synx.conf.desktop      # Targets for hostname "desktop" (replaces base)
├── synx.conf.laptop       # Targets for hostname "laptop" (replaces base)
├── exclude.conf           # Base excludes (shared)
├── exclude.conf.desktop   # Extra excludes for "desktop" (appended)
└── exclude.conf.laptop    # Extra excludes for "laptop" (appended)
```

**Targets** — a machine-specific file replaces the base entirely, so each machine defines exactly what it tracks.

**Excludes** — a machine-specific file is appended to the base, so shared excludes apply everywhere.

The active hostname is shown in the UI header (e.g. `Dotfile Sync (arch)`), and `--list` annotates machine-specific targets.

## Bootstrap — Full System Setup

Set up a new machine in one command using a declarative bootstrap config.

### Create a Bootstrap Config

```bash
synx --bootstrap-setup
```

The wizard walks you through:
1. **AUR Helper** — detect or choose paru/yay
2. **Packages** — select packages to install
3. **Git Repos** — repos to clone with optional install scripts
4. **Dotfile Restore** — auto-restore dotfiles after setup
5. **Custom Commands** — post-install commands (chsh, systemctl, etc.)

### Bootstrap a New Machine

```bash
synx --bootstrap https://github.com/user/dotfiles.git
```

This will clone your dotfiles, find the bootstrap config, show it for review/editing, then execute each step with confirmations (skip with `--yes`).

### Config Example

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

## Configuration

| File | Purpose |
|------|---------|
| `~/.config/synx/synx.conf` | Tracked dotfiles (base) |
| `~/.config/synx/synx.conf.<hostname>` | Tracked dotfiles (machine override) |
| `~/.config/synx/exclude.conf` | Exclude patterns (base) |
| `~/.config/synx/exclude.conf.<hostname>` | Exclude patterns (machine-specific) |
| `~/.config/synx/bootstrap.conf` | Bootstrap config |
| `~/dotfiles/` | Dotfiles git repository |

## Requirements

- Go (for building)
- Git
