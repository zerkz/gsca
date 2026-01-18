# GSCA - Global Steam Command Args

![CI](https://github.com/zerkz/gsca/workflows/CI/badge.svg)
![Release](https://github.com/zerkz/gsca/workflows/Release/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/zerkz/gsca)](https://goreportcard.com/report/github.com/zerkz/gsca)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A cross-platform CLI to manage Steam game launch options with interactive search and bulk updates.

## Installation

```bash
go build -o gsca
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

Search for installed games and export selections to a file.

```bash
gsca query baldur        # Search for "baldur"
gsca query               # Show first 10 installed games
gsca query --all         # Show all installed games
```

**Flags:**
| Flag | Description |
|------|-------------|
| `--limit int` | Maximum number of results to show (default 10) |
| `--all` | Show all matches (no limit) |

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
| `-f, --force` | Automatically close Steam if running (no prompt) |
| `-o, --open` | Open the config file after updating |
| `--dry-run` | Show changes without modifying files |
| `--no-backup` | Skip creating backup file |
| `--ignore-missing` | Continue if games in list are not found |

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
