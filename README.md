<div align="center">
  <img src="https://raw.githubusercontent.com/Blumenwagen/synx/main/readme.gif" width="0" height="0"> <!-- preload -->
  
  <h1>🚀 SYNX</h1>
  <p><b>A blazing fast CLI tool for managing dotfiles</b></p>
  <p><i>Git-based version control • Multi-machine support • Smart profiles • Full system bootstrapping</i></p>

  <p>
    <a href="https://aur.archlinux.org/packages/synx-git"><img src="https://img.shields.io/aur/version/synx-git?color=1793d1&label=AUR&logo=arch-linux&style=for-the-badge" alt="AUR Package"></a>
    <a href="https://github.com/Blumenwagen/synx/blob/main/LICENSE"><img src="https://img.shields.io/github/license/Blumenwagen/synx?color=blue&style=for-the-badge" alt="License"></a>
    <img src="https://img.shields.io/github/go-mod/go-version/Blumenwagen/synx?filename=synx-go%2Fgo.mod&style=for-the-badge&logo=go" alt="Go Version">
  </p>
</div>

<br>

<p align="center">
  <img src="readme.gif" alt="Synx Demo execution" width="100%" onerror="this.style.display='none';"> 
</p>


## 🚀 Quick Start

### For Arch-Users an AUR build is available at synx-git:
```bash
# use your preferred AUR helper (yay, paru etc.)
yay -S synx-git

synx #first run will trigger the setup wizard
```
> [!NOTE]
> Note that the AUR version does **NOT** support "synx --update", instead you will need to update using your Package Manager of choice.

---
### Other Distributions can Install via the install script:
```bash
git clone https://github.com/Blumenwagen/synx.git
cd synx
./install.sh
```

Requires **git**, the installer will try to install **Go** on it's own but if this fails, you will need to manually install go for your system.

<br>

## 💻 Supported Platforms

| Platform | Support Level | Description |
| :--- | :---: | :--- |
| 🐧 **Arch Linux** | 🟢 **Full** | All features including pacman/AUR tracking and system bootstrapping. |
| 🐧 **Other Distros** | 🟡 **Core** | Standard dotfile syncing, Git version control, and profiles work out of the box. |
| 🍏 **macOS** | 🟡 **Core** | Core syncing capabilities are functional. System packages and services are unmanaged. |
| 🪟 **Windows** | 🟡 **WSL** | Should work seamlessly inside WSL for standard syncing. |

<br>

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

## 🧩 Features

| Feature | Description |
|---------|-------------|
| 🔄 **Sync & push** | Push dotfiles and commits automatically in one command |
| ⬇️ **Restore** | Pull and restore dotfiles from remote |
| 🔍 **Status & diff** | See what changed since last sync |
| 🧪 **Dry-run** | Preview operations without touching files |
| 🎭 **Profiles** | Named presets for quick config switching |
| 🪝 **Hooks** | Custom pre/post sync and restore scripts |
| ⏪ **History & rollback** | View commits, time-travel to previous states |
| 📦 **Package tracking** | Snapshot and restore installed packages (pacman + AUR) |
| ⚙️ **Service tracking** | Snapshot and restore enabled systemd services |
| 🚀 **Bootstrap** | Provision a new machine from a declarative config (experimental) |
| 🩺 **Doctor** | Health checks and diagnostics |

> [!WARNING]
> The bootstrap feature is designed for **Arch Linux** and Arch-based distributions.
> Core syncing should work on any Unix-like system.

## 💡 Real-world Examples

Here is how `synx` looks in practice:

**1. Track a new config file and upload it to GitHub:**
```bash
synx --add nvim
synx # Backs up ~/.config/nvim to dotfiles and pushes it
```

**2. Setup a completely new laptop:**
```bash
git clone https://github.com/my/dotfiles.git ~/dotfiles
synx -r # Restores all configs and symlinks them to ~/.config
synx pkg restore # Downloads missing pacman/AUR packages
synx svc restore # Enables necessary system services
```

**3. Switch from a "battery-saver" theme to "smooth animations":**
```bash
synx --profile smooth # Changes the active profile and reloads Hyprland automatically
```

## 🔥 Why synx over GNU Stow or a Bare Git Repo?

- **Zero mental overhead:** A bare git repo forces you to manually `git add`, `commit`, and `push` every time you edit a config file. `synx` does all of this automatically in one command.
- **Explicit tracking:** `stow` blindly symlinks entire directories, meaning you often accidentally backup unnecessary cache files. `synx` only syncs the specific directories/files you explicitly tell it to via `--add`.
- **System state capture:** `synx` doesn't just manage config files. It seamlessly captures and tracks your installed system packages (pacman/AUR) and enabled systemd services.
- **Smart Profiles:** Running different configs on your desktop and laptop is painful with standard tools. `synx` lets you create profiles (e.g. `pc`, `laptop`) that act as explicit fallback overrides over your base config.


## 📖 Documentation

**Full documentation is available on the [Wiki](https://github.com/Blumenwagen/synx/wiki):**

| Page | Description |
|------|-------------|
| [Commands](https://github.com/Blumenwagen/synx/wiki/Commands) | Full command reference |
| [Packages](https://github.com/Blumenwagen/synx/wiki/Packages) | Package state tracking |
| [Services](https://github.com/Blumenwagen/synx/wiki/Services) | Service state tracking |
| [Profiles](https://github.com/Blumenwagen/synx/wiki/Profiles) | Config presets & animation switching |
| [Hooks](https://github.com/Blumenwagen/synx/wiki/Hooks) | Custom sync/restore scripts |
| [Bootstrap](https://github.com/Blumenwagen/synx/wiki/Bootstrap) | New machine provisioning |
| [Doctor](https://github.com/Blumenwagen/synx/wiki/Doctor) | Health checks & diagnostics |


 
 
