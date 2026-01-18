# Technical Documentation

## How It Works

1. Checks if Steam is running and warns you to close it
2. Detects Steam installation path for your platform
3. Finds the most recently used Steam user ID
4. Parses `appmanifest_*.acf` files to map game names to app IDs
5. Parses `localconfig.vdf` to find existing game configs
6. Applies filters based on allow/deny lists
7. Updates `LaunchOptions` for selected games
8. Creates an incremental backup before saving changes

## Steam Config Locations

- **Linux**: `~/.local/share/Steam/userdata/<userid>/config/localconfig.vdf`
- **Windows**: `C:\Program Files (x86)\Steam\userdata\<userid>\config\localconfig.vdf`
- **macOS**: `~/Library/Application Support/Steam/userdata/<userid>/config/localconfig.vdf`

## Cross-Platform Steam Management

The tool uses platform-specific methods to manage Steam:

### Linux
- **Detect**: `pgrep -x steam`
- **Close**: `steam -shutdown` (graceful shutdown)
- **Start**: `steam`

### Windows
- **Detect**: `tasklist /FI "IMAGENAME eq steam.exe"`
- **Close**: `steam://exitsteam` (graceful) → `taskkill /IM steam.exe` (fallback)
- **Start**: `steam://open/main` (works with any install location)

### macOS
- **Detect**: `pgrep -x steam`
- **Close**: AppleScript `quit app "Steam"` (graceful)
- **Start**: `open -a Steam`

All methods prioritize graceful shutdown to prevent data loss.

## Backup Management

### Automatic Incremental Backups

The tool **never overwrites existing backups**. Each run creates a new backup file:

```
localconfig.vdf.backup       # First backup
localconfig.vdf.backup.1     # Second backup
localconfig.vdf.backup.2     # Third backup
```

### Restoring from Backup

Make sure Steam is closed, then copy the backup back:

```bash
# Linux example
cp ~/.local/share/Steam/userdata/<userid>/config/localconfig.vdf.backup \
   ~/.local/share/Steam/userdata/<userid>/config/localconfig.vdf
```

## Building for Different Platforms

### Linux
```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gsca-linux
```

### Windows
```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o gsca.exe
```

### macOS
```bash
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o gsca-macos
```

## Binary Size

Approximately 2-5MB with standard Go build, or 1-3MB with stripped symbols (`-ldflags="-s -w"`).

## File Format

### Allow/Deny List Format

Plain text file with one entry per line:
- Numeric Steam app IDs: `570`, `730`, `1086940`
- Comments: lines starting with `#`
- Empty lines are ignored

Example:
```
# My favorite games
570     # Dota 2
730     # Counter-Strike 2
1086940 # Baldur's Gate 3
```

Use `gsca query` to find app IDs interactively.
