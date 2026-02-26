# SYNX — Dotfile Backup System

A fast CLI tool for managing dotfiles with git-based version control, multi-machine support, profiles, and full system bootstrapping.

## Quick Start

```bash
git clone https://github.com/Blumenwagen/synx.git
cd synx
./install.sh
```

Requires **Go** and **git**. The installer builds the binary and symlinks it to `~/.local/bin/synx`.

## Basic Usage

```bash
synx                       # Sync dotfiles to remote Git Repo
synx -r                    # Restore from remote Git Repo
synx -s                    # Show what changed
synx -n                    # Dry-run (preview only)
synx --update              # Update synx to the latest available version
synx --doctor              # Health checks
synx --profile smooth      # Switch to a profile
synx --help                # Full command list
```

## Features

- **Sync & push** — copy dotfiles to a git repo and push in one command
- **Restore** — pull and restore dotfiles from remote
- **Status & diff** — see what changed since last sync
- **Dry-run** — preview operations without touching files
- **Multi-machine** — per-hostname targets and excludes
- **Profiles** — named presets for quick config switching
- **Hooks** — custom pre/post sync and restore scripts
- **History & rollback** — view commits, time-travel to previous states
- **Package tracking** — snapshot and restore installed packages (pacman + AUR)
- **Service tracking** — snapshot and restore enabled systemd services
- **Bootstrap** — provision a new machine from a declarative config
- **Doctor** — health checks and diagnostics

> [!WARNING]
> The bootstrap feature is designed for **Arch Linux** and Arch-based distributions.
> Core syncing should work on any Unix-like system.


## 📖 Documentation

**Full documentation is available on the [Wiki](https://github.com/Blumenwagen/synx/wiki):**

| Page | Description |
|------|-------------|
| [Commands](https://github.com/Blumenwagen/synx/wiki/Commands) | Full command reference |
| [Packages](https://github.com/Blumenwagen/synx/wiki/Packages) | Package state tracking |
| [Services](https://github.com/Blumenwagen/synx/wiki/Services) | Service state tracking |
| [Profiles](https://github.com/Blumenwagen/synx/wiki/Profiles) | Config presets & animation switching |
| [Hooks](https://github.com/Blumenwagen/synx/wiki/Hooks) | Custom sync/restore scripts |
| [Multi-Machine](https://github.com/Blumenwagen/synx/wiki/Multi-Machine) | Per-hostname overrides |
| [Bootstrap](https://github.com/Blumenwagen/synx/wiki/Bootstrap) | New machine provisioning |
| [Doctor](https://github.com/Blumenwagen/synx/wiki/Doctor) | Health checks & diagnostics |


