package steam

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/zerkz/gsca/vdf"
)

const (
	appStateKey = "AppState"
	osWindows   = "windows"
	osLinux     = "linux"
	osDarwin    = "darwin"
	keyAppID    = "appid"
	keyName     = "name"
)

// GetSteamPath returns the Steam installation path for the current platform
func GetSteamPath() (string, error) {
	var steamPath string

	switch runtime.GOOS {
	case osLinux:
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		steamPath = filepath.Join(homeDir, ".local", "share", "Steam")

	case osWindows:
		steamPath = `C:\Program Files (x86)\Steam`
		// Also check for custom install location in registry if needed

	case osDarwin:
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		steamPath = filepath.Join(homeDir, "Library", "Application Support", "Steam")

	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// Verify the path exists
	if _, err := os.Stat(steamPath); os.IsNotExist(err) {
		return "", fmt.Errorf("steam installation not found at %s", steamPath)
	}

	return steamPath, nil
}

// GetUserID returns the most recently used Steam user ID
func GetUserID(steamPath string) (string, error) {
	userdataPath := filepath.Join(steamPath, "userdata")

	entries, err := os.ReadDir(userdataPath)
	if err != nil {
		return "", fmt.Errorf("failed to read userdata directory: %w", err)
	}

	// Find the most recently modified user directory
	var latestUserID string
	var latestModTime int64

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip if not a numeric ID
		if _, err := fmt.Sscanf(entry.Name(), "%d", new(int)); err != nil {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		modTime := info.ModTime().Unix()
		if modTime > latestModTime {
			latestModTime = modTime
			latestUserID = entry.Name()
		}
	}

	if latestUserID == "" {
		return "", fmt.Errorf("no valid user ID found in userdata directory")
	}

	return latestUserID, nil
}

// GetLocalConfigPath returns the path to localconfig.vdf
func GetLocalConfigPath(steamPath, userID string) string {
	return filepath.Join(steamPath, "userdata", userID, "config", "localconfig.vdf")
}

// GameInfo represents information about a Steam game
type GameInfo struct {
	AppID         string
	Name          string
	LaunchOptions string
	Installed     bool
}

// GetGameMapping returns a map of game names (lowercase) to app IDs
func GetGameMapping(steamPath string) (map[string]string, error) {
	mapping := make(map[string]string)

	// Get all library folders
	libraryFolders, err := GetLibraryFolders(steamPath)
	if err != nil {
		return nil, err
	}

	// Scan each library folder
	for _, libraryPath := range libraryFolders {
		steamappsPath := filepath.Join(libraryPath, "steamapps")

		// Read all appmanifest files in this library
		files, err := filepath.Glob(filepath.Join(steamappsPath, "appmanifest_*.acf"))
		if err != nil {
			continue // Skip this library if glob fails
		}

		for _, file := range files {
			f, err := os.Open(file)
			if err != nil {
				continue
			}

			parser := vdf.NewParser(f)
			root, err := parser.Parse()
			_ = f.Close()

			if err != nil {
				continue
			}

			// Find AppState node
			var appState *vdf.Node
			for _, child := range root.Children {
				if child.Key == appStateKey {
					appState = child
					break
				}
			}

			if appState == nil {
				continue
			}

			var appID, name string
			for _, child := range appState.Children {
				switch child.Key {
				case keyAppID:
					appID = child.Value
				case keyName:
					name = child.Value
				}
			}

			if appID != "" && name != "" {
				// Store with lowercase name for case-insensitive matching
				mapping[strings.ToLower(name)] = appID
				// Also store with the app ID as key for direct ID lookup
				mapping[appID] = appID
			}
		}
	}

	return mapping, nil
}

// GetAllGameIDs returns all app IDs from the localconfig.vdf
func GetAllGameIDs(localConfigPath string) ([]string, error) {
	f, err := os.Open(localConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open localconfig.vdf: %w", err)
	}
	defer func() { _ = f.Close() }()

	parser := vdf.NewParser(f)
	root, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse localconfig.vdf: %w", err)
	}

	// Navigate to Software/Valve/Steam/apps
	appsNode := vdf.FindNode(root, "UserLocalConfigStore/Software/Valve/Steam/apps")
	if appsNode == nil {
		return nil, fmt.Errorf("apps node not found in localconfig.vdf")
	}

	var appIDs []string
	for _, child := range appsNode.Children {
		appIDs = append(appIDs, child.Key)
	}

	return appIDs, nil
}

// GetLibraryFolders returns all Steam library folder paths
func GetLibraryFolders(steamPath string) ([]string, error) {
	libraryFoldersPath := filepath.Join(steamPath, "steamapps", "libraryfolders.vdf")

	f, err := os.Open(libraryFoldersPath)
	if err != nil {
		// If libraryfolders.vdf doesn't exist, just return default path
		return []string{steamPath}, nil
	}
	defer func() { _ = f.Close() }()

	parser := vdf.NewParser(f)
	root, err := parser.Parse()
	if err != nil {
		return []string{steamPath}, nil
	}

	// Navigate to libraryfolders node
	var libraryNode *vdf.Node
	for _, child := range root.Children {
		if child.Key == "libraryfolders" {
			libraryNode = child
			break
		}
	}

	if libraryNode == nil {
		return []string{steamPath}, nil
	}

	var paths []string
	for _, child := range libraryNode.Children {
		// Each child is a library entry
		for _, field := range child.Children {
			if field.Key == "path" {
				paths = append(paths, field.Value)
				break
			}
		}
	}

	if len(paths) == 0 {
		return []string{steamPath}, nil
	}

	return paths, nil
}

// getInstalledGameNames returns a map of app IDs to game names (with original casing)
func getInstalledGameNames(steamPath string) (map[string]string, error) {
	appIDToName := make(map[string]string)

	// Get all library folders
	libraryFolders, err := GetLibraryFolders(steamPath)
	if err != nil {
		return nil, err
	}

	// Scan each library folder
	for _, libraryPath := range libraryFolders {
		steamappsPath := filepath.Join(libraryPath, "steamapps")

		// Read all appmanifest files in this library
		files, err := filepath.Glob(filepath.Join(steamappsPath, "appmanifest_*.acf"))
		if err != nil {
			continue // Skip this library if glob fails
		}

		for _, file := range files {
			f, err := os.Open(file)
			if err != nil {
				continue
			}

			parser := vdf.NewParser(f)
			root, err := parser.Parse()
			_ = f.Close()

			if err != nil {
				continue
			}

			// Find AppState node
			var appState *vdf.Node
			for _, child := range root.Children {
				if child.Key == appStateKey {
					appState = child
					break
				}
			}

			if appState == nil {
				continue
			}

			var appID, name string
			for _, child := range appState.Children {
				switch child.Key {
				case keyAppID:
					appID = child.Value
				case keyName:
					name = child.Value
				}
			}

			if appID != "" && name != "" {
				appIDToName[appID] = name
			}
		}
	}

	return appIDToName, nil
}

// GetAllGames returns all games from localconfig with their names and launch options
func GetAllGames(steamPath, localConfigPath string) ([]GameInfo, error) {
	// Get installed game names with original casing
	installedNames, err := getInstalledGameNames(steamPath)
	if err != nil {
		return nil, err
	}

	// Get all games from localconfig
	f, err := os.Open(localConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open localconfig.vdf: %w", err)
	}
	defer func() { _ = f.Close() }()

	parser := vdf.NewParser(f)
	root, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse localconfig.vdf: %w", err)
	}

	// Navigate to Software/Valve/Steam/apps
	appsNode := vdf.FindNode(root, "UserLocalConfigStore/Software/Valve/Steam/apps")
	if appsNode == nil {
		return nil, fmt.Errorf("apps node not found in localconfig.vdf")
	}

	var games []GameInfo
	for _, appNode := range appsNode.Children {
		appID := appNode.Key

		// Get launch options if they exist
		var launchOptions string
		launchNode := vdf.FindNode(appNode, "LaunchOptions")
		if launchNode != nil {
			launchOptions = launchNode.Value
		}

		// Check if game is installed and get name
		name, installed := installedNames[appID]
		if !installed {
			// Not installed, use app ID as name
			name = appID
		}

		games = append(games, GameInfo{
			AppID:         appID,
			Name:          name,
			Installed:     installed,
			LaunchOptions: launchOptions,
		})
	}

	return games, nil
}
