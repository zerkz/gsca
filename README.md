# GSCA - Global Steam Command Args

![CI](https://github.com/zerkz/gsca/workflows/CI/badge.svg)
![Release](https://github.com/zerkz/gsca/workflows/Release/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/zerkz/gsca)](https://goreportcard.com/report/github.com/zerkz/gsca)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A cross-platform CLI to manage Steam game launch options with interactive search and bulk updates.

I mainly made this for easily updating newly installed Steam games with `game-performance %command` for [CachyOS game mode](https://wiki.cachyos.org/configuration/gaming/#power-profile-switching-on-demand).

## Installation

**macOS (Homebrew):**
```bash
brew install zerkz/gsca-brew/gsca
```

**Windows (Scoop):**
```powershell
scoop bucket add gsca https://github.com/zerkz/gsca-scoop.git
scoop install gsca
```

**Arch Linux (AUR):**
```bash
yay -S gsca-bin
```

**Steam Deck / Linux (Flatpak):**
```bash
curl -LO https://github.com/zerkz/gsca/releases/download/v1.0.3/gsca-v1.0.3.flatpak
flatpak install gsca-v1.0.3.flatpak
```

**Manual (Linux/macOS):**
```bash
curl -LO https://github.com/zerkz/gsca/releases/download/v1.0.3/gsca_1.0.3_linux_amd64.tar.gz
tar -xzf gsca_1.0.3_linux_amd64.tar.gz
sudo mv gsca /usr/local/bin/
```

**From source:**
```bash
go install github.com/zerkz/gsca@latest
```

## Quick Start

```bash
# 1. Search for games and select which to update
gsca query dota

# 2. Update selected games
gsca update --args "gamemoderun %command%" --allow selected-games.txt
```

## Commands

### `gsca query [search term]`

Search for installed games and interactively select which ones to export.

```bash
gsca query baldur        # Search for "baldur"
gsca query               # Show all installed games
```

Selection syntax: `1,3,5` (specific), `1-5` (range), `*` (all)

### `gsca list [file]`

Display game details from a list file.

```bash
gsca list                # Uses selected-games.txt
gsca list my-games.txt   # Specific file
```

**Flags:**
| Flag | Description |
|------|-------------|
| `-f, --file string` | Path to game list file (default "selected-games.txt") |

### `gsca update`

Update launch options for games.

```bash
gsca update --args "gamemoderun %command%" --allow games.txt
gsca update --args "mangohud %command%" --allow games.txt --force
gsca update --args "test" --deny exclude.txt --dry-run
```

**Flags:**
| Flag | Description |
|------|-------------|
| `-a, --args string` | Launch arguments to set (required) |
| `-l, --allow string` | Path to allow list file |
| `-d, --deny string` | Path to deny list file |
| `--all` | Update all games (use with caution) |
| `-f, --force` | Automatically close Steam if running (no prompt) |
| `-o, --open` | Open the config file after updating |
| `--dry-run` | Show changes without modifying files |
| `--no-backup` | Skip creating backup file |
| `--ignore-missing` | Continue if games in list are not found |

### `gsca restore-backup`

List available config backups and interactively select one to restore.

```bash
gsca restore-backup
```

### Global Flags

| Flag | Description |
|------|-------------|
| `-s, --steam-path string` | Override Steam installation path |
| `-u, --user-id string` | Override Steam user ID |
| `--include-tools` | Include Steam tools (Proton, runtimes, etc.) |

## Steam Warning

Steam overwrites `localconfig.vdf` when it closes. The tool detects if Steam is running and will prompt you to close it (or use `--force` to auto-close).

## Common Launch Options
- `gamemoderun %command%` - Feral GameMode 
- `mangohud %command%` - MangoHud
- `PROTON_LOG=1 %command%` - Proton logging
- `game-performance %command%` - [game-performance mode (CachyOS utility)](https://wiki.cachyos.org/configuration/gaming/#power-profile-switching-on-demand)

_Note: gsca does not install utilities in the above commands for you._

## License

MIT

---

See [TECHNICAL.md](TECHNICAL.md) for build instructions, backup management, and implementation details.
