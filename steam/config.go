package steam

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/zdware/gsca/vdf"
)

// UpdateLaunchOptions updates launch options for specified games
func UpdateLaunchOptions(localConfigPath string, appIDs []string, launchArgs string, skipBackup bool) (string, error) {
	// Read the original file
	f, err := os.Open(localConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to open localconfig.vdf: %w", err)
	}

	parser := vdf.NewParser(f)
	root, err := parser.Parse()
	_ = f.Close()

	if err != nil {
		return "", fmt.Errorf("failed to parse localconfig.vdf: %w", err)
	}

	// Update launch options for each app ID
	for _, appID := range appIDs {
		path := fmt.Sprintf("UserLocalConfigStore/Software/Valve/Steam/apps/%s/LaunchOptions", appID)
		if setErr := vdf.SetValue(root, path, launchArgs); setErr != nil {
			return "", fmt.Errorf("failed to set launch options for app %s: %w", appID, setErr)
		}
	}

	// Create backup (unless skipped)
	var backupPath string
	if !skipBackup {
		backupPath = getNextBackupPath(localConfigPath)
		if copyErr := copyFile(localConfigPath, backupPath); copyErr != nil {
			return "", fmt.Errorf("failed to create backup: %w", copyErr)
		}
	}

	// Write the updated config
	outFile, err := os.Create(localConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	writer := bufio.NewWriter(outFile)
	if err := vdf.Write(writer, root, 0); err != nil {
		return "", fmt.Errorf("failed to write VDF: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("failed to flush writer: %w", err)
	}

	return backupPath, nil
}

// LoadFilterList loads a list of game names or IDs from a file
func LoadFilterList(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open filter file: %w", err)
	}
	defer func() { _ = f.Close() }()

	var items []string
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		items = append(items, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading filter file: %w", err)
	}

	return items, nil
}

// ResolveGameIDs validates that items are numeric app IDs
// Game names are no longer supported - use query/list modes to get IDs
func ResolveGameIDs(items []string, mapping map[string]string) ([]string, []string) {
	var resolved []string
	var notFound []string

	for _, item := range items {
		// Check if it's a numeric ID
		isNumeric := true
		for _, ch := range item {
			if ch < '0' || ch > '9' {
				isNumeric = false
				break
			}
		}

		if isNumeric && len(item) > 0 {
			// It's a numeric app ID - use it directly
			resolved = append(resolved, item)
		} else {
			// Non-numeric entries are invalid
			notFound = append(notFound, item)
		}
	}

	return resolved, notFound
}

// FilterGameIDs filters game IDs based on allow/deny lists
func FilterGameIDs(allGameIDs []string, allowList, denyList []string) []string {
	if len(allowList) > 0 {
		// Only include games in the allow list
		allowSet := make(map[string]bool)
		for _, id := range allowList {
			allowSet[id] = true
		}

		var filtered []string
		for _, id := range allGameIDs {
			if allowSet[id] {
				filtered = append(filtered, id)
			}
		}
		return filtered
	}

	if len(denyList) > 0 {
		// Exclude games in the deny list
		denySet := make(map[string]bool)
		for _, id := range denyList {
			denySet[id] = true
		}

		var filtered []string
		for _, id := range allGameIDs {
			if !denySet[id] {
				filtered = append(filtered, id)
			}
		}
		return filtered
	}

	// No filtering
	return allGameIDs
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}

// getNextBackupPath finds the next available backup filename
// Returns: localconfig.vdf.backup, localconfig.vdf.backup.1, localconfig.vdf.backup.2, etc.
func getNextBackupPath(originalPath string) string {
	basePath := originalPath + ".backup"

	// Check if base backup exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return basePath
	}

	// Find next available numbered backup
	for i := 1; i < 10000; i++ {
		backupPath := fmt.Sprintf("%s.%d", basePath, i)
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			return backupPath
		}
	}

	// Fallback (should never happen unless you have 10000 backups!)
	return fmt.Sprintf("%s.%d", basePath, 10000)
}
