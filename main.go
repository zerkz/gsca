package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zerkz/gsca/steam"
)

// Global flags
var (
	steamPath    string
	userID       string
	includeTools bool
)

// Update command flags
var (
	launchArgs     string
	allowFile      string
	denyFile       string
	dryRun         bool
	autoCloseSteam bool
	noBackup       bool
	ignoreMissing  bool
	openConfig     bool
)

// Query command flags
var (
	queryLimit int
	queryAll   bool
)

const statusNotInstalled = " [NOT INSTALLED]"

var rootCmd = &cobra.Command{
	Use:   "gsca",
	Short: "Global Steam Command Args - Manage Steam game launch options",
	Long: `gsca is a CLI tool to manage Steam game launch options.

Commands:
  update    Update launch options for games
  query     Search for games and view their launch options`,
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update launch options for games",
	Long: `Update Steam game command arguments (launch options) for multiple games.

You can specify games using an allow list or deny list file. The tool supports both game IDs and game names.`,
	RunE: runUpdate,
}

var queryCmd = &cobra.Command{
	Use:   "query [search term]",
	Short: "Search for games interactively",
	Long: `Search for games by name and interactively select which ones to view or update.

The query command will show matching games and let you select them interactively.
Use --all without a search term to show all games in your library.`,
	RunE: runQuery,
}

var listCmd = &cobra.Command{
	Use:   "list [file]",
	Short: "Show details for games in a list file",
	Long: `Display game names and app IDs from a list file.

If a file contains app IDs, the game names will be shown (if installed).
If a file contains game names, the app IDs will be shown.`,
	RunE: runList,
}

var listFile string

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&steamPath, "steam-path", "s", "", "Override Steam installation path (auto-detected if not specified)")
	rootCmd.PersistentFlags().StringVarP(&userID, "user-id", "u", "", "Override Steam user ID (auto-detected if not specified)")
	rootCmd.PersistentFlags().BoolVar(&includeTools, "include-tools", false, "Include Steam tools (Proton, runtimes, etc.)")

	// Update command flags
	updateCmd.Flags().StringVarP(&launchArgs, "args", "a", "", "Launch arguments to set for games (required)")
	updateCmd.Flags().StringVarP(&allowFile, "allow", "l", "", "Path to allow list file (one game name or ID per line)")
	updateCmd.Flags().StringVarP(&denyFile, "deny", "d", "", "Path to deny list file (one game name or ID per line)")
	updateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without actually modifying files")
	updateCmd.Flags().BoolVarP(&autoCloseSteam, "force", "f", false, "Automatically close Steam if running (no prompt)")
	updateCmd.Flags().BoolVar(&noBackup, "no-backup", false, "Skip creating backup file")
	updateCmd.Flags().BoolVar(&ignoreMissing, "ignore-missing", false, "Continue even if games in allow/deny list are not found")
	updateCmd.Flags().BoolVarP(&openConfig, "open", "o", false, "Open the config file after updating")
	_ = updateCmd.MarkFlagRequired("args")

	// Query command flags
	queryCmd.Flags().IntVar(&queryLimit, "limit", 10, "Maximum number of results to show")
	queryCmd.Flags().BoolVar(&queryAll, "all", false, "Show all matches (no limit)")

	// List command flags
	listCmd.Flags().StringVarP(&listFile, "file", "f", "selected-games.txt", "Path to game list file")

	// Add subcommands
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(listCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// Validate flags
	if allowFile != "" && denyFile != "" {
		return fmt.Errorf("cannot specify both --allow and --deny flags")
	}

	// Check if Steam is running (skip in dry-run mode)
	var shouldRestartSteam bool
	if !dryRun {
		steamRunning, err := steam.IsSteamRunning()
		if err != nil {
			fmt.Printf("Warning: Could not check if Steam is running: %v\n", err)
		} else if steamRunning {
			var shouldClose bool

			if autoCloseSteam {
				// Force mode - automatically close Steam
				fmt.Println("WARNING: Steam is running - closing automatically (--force flag)")
				shouldClose = true
			} else {
				// Interactive mode - ask user
				fmt.Println("\nWARNING: Steam is currently running!")
				fmt.Println("Steam overwrites localconfig.vdf when it closes, which will undo your changes.")
				fmt.Print("\nClose Steam and apply changes? (Y/n): ")

				var response string
				_, _ = fmt.Scanln(&response)
				response = strings.ToLower(strings.TrimSpace(response))

				if response == "" || response == "y" || response == "yes" {
					shouldClose = true
				} else {
					return fmt.Errorf("aborted - Steam must be closed to apply changes safely")
				}
			}

			if shouldClose {
				fmt.Println("Closing Steam...")
				if err := steam.CloseSteam(); err != nil {
					return fmt.Errorf("failed to close Steam: %w", err)
				}

				// Wait for Steam to fully close
				fmt.Print("Waiting for Steam to close")
				for i := 0; i < 10; i++ {
					time.Sleep(1 * time.Second)
					fmt.Print(".")
					running, _ := steam.IsSteamRunning()
					if !running {
						break
					}
				}
				fmt.Println(" done!")

				// Verify Steam is closed
				stillRunning, _ := steam.IsSteamRunning()
				if stillRunning {
					return fmt.Errorf("Steam is still running after close attempt - please close it manually")
				}

				shouldRestartSteam = true
			}

			fmt.Println()
		}
	}

	// Get Steam path
	var err error
	if steamPath == "" {
		steamPath, err = steam.GetSteamPath()
		if err != nil {
			return fmt.Errorf("failed to detect Steam path: %w", err)
		}
	}
	fmt.Printf("Steam path: %s\n", steamPath)

	// Get user ID
	if userID == "" {
		userID, err = steam.GetUserID(steamPath)
		if err != nil {
			return fmt.Errorf("failed to detect user ID: %w", err)
		}
	}
	fmt.Printf("User ID: %s\n", userID)

	// Get localconfig path
	localConfigPath := steam.GetLocalConfigPath(steamPath, userID)
	fmt.Printf("Local config: %s\n", localConfigPath)

	// Get game mapping
	fmt.Println("Loading game mapping...")
	mapping, err := steam.GetGameMapping(steamPath)
	if err != nil {
		return fmt.Errorf("failed to get game mapping: %w", err)
	}
	fmt.Printf("Found %d games\n", len(mapping)/2)

	// Get all game IDs from localconfig
	allGameIDs, err := steam.GetAllGameIDs(localConfigPath)
	if err != nil {
		return fmt.Errorf("failed to get game IDs: %w", err)
	}

	// Load and resolve allow/deny lists
	var targetGameIDs []string

	if allowFile != "" {
		resolvedIDs, loadErr := loadAndResolveFilterList(allowFile, "allow", mapping, ignoreMissing)
		if loadErr != nil {
			return loadErr
		}
		targetGameIDs = steam.FilterGameIDs(allGameIDs, resolvedIDs, nil)
	} else if denyFile != "" {
		resolvedIDs, loadErr := loadAndResolveFilterList(denyFile, "deny", mapping, ignoreMissing)
		if loadErr != nil {
			return loadErr
		}
		targetGameIDs = steam.FilterGameIDs(allGameIDs, nil, resolvedIDs)
	} else {
		// No filter - update all games
		targetGameIDs = allGameIDs
	}

	fmt.Printf("\nWill update launch options for %d games\n", len(targetGameIDs))
	fmt.Printf("Launch args: %s\n", launchArgs)

	if dryRun {
		fmt.Println("\n[DRY RUN] Would update the following app IDs:")
		for _, appID := range targetGameIDs {
			fmt.Printf("  - %s\n", appID)
		}

		// Open config file if requested (useful to see current state)
		if openConfig {
			fmt.Printf("\nOpening config file: %s\n", localConfigPath)
			if openErr := steam.OpenFile(localConfigPath); openErr != nil {
				fmt.Printf("Warning: Failed to open config file: %v\n", openErr)
				fmt.Println("You can open it manually at:", localConfigPath)
			}
		}

		return nil
	}

	// Update launch options
	fmt.Println("\nUpdating launch options...")
	backupPath, err := steam.UpdateLaunchOptions(localConfigPath, targetGameIDs, launchArgs, noBackup)
	if err != nil {
		return fmt.Errorf("failed to update launch options: %w", err)
	}

	fmt.Printf("\nSuccessfully updated %d games!\n", len(targetGameIDs))
	if backupPath != "" {
		fmt.Printf("Backup created at: %s\n", backupPath)
	}

	// Restart Steam if we closed it
	if shouldRestartSteam {
		fmt.Println("\nRestarting Steam...")
		if err := steam.StartSteam(); err != nil {
			fmt.Printf("Warning: Failed to start Steam: %v\n", err)
			fmt.Println("Please start Steam manually.")
		} else {
			fmt.Println("Steam started successfully!")
		}
	}

	// Open config file if requested
	if openConfig {
		fmt.Printf("\nOpening config file: %s\n", localConfigPath)
		if err := steam.OpenFile(localConfigPath); err != nil {
			fmt.Printf("Warning: Failed to open config file: %v\n", err)
			fmt.Println("You can open it manually at:", localConfigPath)
		}
	}

	return nil
}

func runQuery(cmd *cobra.Command, args []string) error {
	var query string
	if len(args) > 0 {
		query = strings.Join(args, " ")
	}

	// Validate flags
	if queryAll && query != "" {
		return fmt.Errorf("cannot combine --all flag with a search term")
	}

	// Get Steam path
	var err error
	if steamPath == "" {
		steamPath, err = steam.GetSteamPath()
		if err != nil {
			return fmt.Errorf("failed to detect Steam path: %w", err)
		}
	}

	// Get user ID
	if userID == "" {
		userID, err = steam.GetUserID(steamPath)
		if err != nil {
			return fmt.Errorf("failed to detect user ID: %w", err)
		}
	}

	localConfigPath := steam.GetLocalConfigPath(steamPath, userID)

	// Get all games (installed and uninstalled)
	fmt.Println("Loading game library...")
	allGames, err := steam.GetAllGames(steamPath, localConfigPath)
	if err != nil {
		return fmt.Errorf("failed to get game library: %w", err)
	}

	// Get game mapping for duplicate detection
	mapping, err := steam.GetGameMapping(steamPath)
	if err != nil {
		return fmt.Errorf("failed to get game mapping: %w", err)
	}

	// Filter to only installed games and exclude Steam tools by default
	var installedGames []steam.GameInfo
	for _, game := range allGames {
		if !game.Installed {
			continue
		}

		// Skip Steam tools unless --include-tools is set
		if !includeTools && isSteamTool(game.Name) {
			continue
		}

		installedGames = append(installedGames, game)
	}

	// Search or show all games
	var matches []steam.GameInfo
	if query == "" {
		// No search term - show all installed games
		fmt.Println("\nShowing all installed games")
		matches = installedGames
	} else {
		// Search installed games
		fmt.Printf("\nSearching for: \"%s\"\n", query)
		queryLower := strings.ToLower(query)

		for _, game := range installedGames {
			// Search by name or app ID
			if strings.Contains(strings.ToLower(game.Name), queryLower) ||
				strings.Contains(game.AppID, queryLower) {
				matches = append(matches, game)
			}
		}
	}

	if len(matches) == 0 {
		fmt.Println("\nNo games found matching your query.")
		fmt.Println("\nTips:")
		fmt.Println("   - Try a shorter search term")
		fmt.Println("   - Check for typos")
		fmt.Println("   - The game may not be installed")
		return nil
	}

	// Determine how many to show
	displayLimit := queryLimit
	if queryAll {
		displayLimit = len(matches)
	} else if len(matches) < queryLimit {
		displayLimit = len(matches)
	}

	// Display results
	fmt.Printf("\nFound %d match(es)", len(matches))
	if !queryAll && len(matches) > queryLimit {
		fmt.Printf(" (showing first %d, use --all to see all)", displayLimit)
	}
	fmt.Println(":")

	for i := 0; i < displayLimit; i++ {
		game := matches[i]
		fmt.Printf("[%d] %s\n", i+1, game.Name)
		fmt.Printf("    App ID: %s\n", game.AppID)

		if game.LaunchOptions != "" {
			fmt.Printf("    Launch Options: %s\n", game.LaunchOptions)
		} else {
			fmt.Printf("    Launch Options: (none)\n")
		}
		fmt.Println()
	}

	// Interactive selection
	fmt.Println("────────────────────────────────────────")
	fmt.Println("Select games to export to file:")
	fmt.Println("  • Enter numbers (e.g., 1,3,5 or 1-3)")
	fmt.Println("  • Enter * to select all")
	fmt.Println("  • Press Enter to skip")
	fmt.Print("\nSelection: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		fmt.Println("\nNo games selected. Exiting.")
		return nil
	}

	// Parse selection
	selected := parseSelection(input, displayLimit)
	if len(selected) == 0 {
		fmt.Println("\nInvalid selection. Exiting.")
		return nil
	}

	// Show selected games
	fmt.Println("\nSelected games:")
	var selectedIDs []string
	for _, idx := range selected {
		game := matches[idx]
		fmt.Printf("  • %s (ID: %s)\n", game.Name, game.AppID)
		selectedIDs = append(selectedIDs, game.AppID)
	}

	// Ask where to save
	fmt.Print("\nSave to file (default: selected-games.txt): ")
	filename, _ := reader.ReadString('\n')
	filename = strings.TrimSpace(filename)
	if filename == "" {
		filename = "selected-games.txt"
	}

	// Load existing entries to check for duplicates
	existingAppIDs := make(map[string]bool)
	fileExists := false

	if existingEntries, err := steam.LoadFilterList(filename); err == nil {
		fileExists = true
		// Resolve existing entries to app IDs
		resolvedIDs, _ := steam.ResolveGameIDs(existingEntries, mapping)
		for _, id := range resolvedIDs {
			existingAppIDs[id] = true
		}
	}

	// Filter out duplicates
	var newIDs []string
	var skipped []string
	for _, id := range selectedIDs {
		if existingAppIDs[id] {
			// Find the game name for the skipped ID
			gameName := id
			for _, game := range matches {
				if game.AppID == id {
					gameName = game.Name
					break
				}
			}
			skipped = append(skipped, gameName)
		} else {
			newIDs = append(newIDs, id)
		}
	}

	// Show duplicates if any
	if len(skipped) > 0 {
		fmt.Println("\nWARNING:Skipped duplicates (already in file):")
		for _, name := range skipped {
			fmt.Printf("  • %s\n", name)
		}
	}

	// Only append new entries
	if len(newIDs) > 0 {
		outputFile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer func() { _ = outputFile.Close() }()

		for _, id := range newIDs {
			_, _ = fmt.Fprintf(outputFile, "%s\n", id)
		}

		if fileExists {
			fmt.Printf("\nAppended %d game ID(s) to: %s\n", len(newIDs), filename)
		} else {
			fmt.Printf("\nCreated file and saved %d game ID(s) to: %s\n", len(newIDs), filename)
		}
	} else {
		fmt.Printf("\nWARNING:No new games to add (all selections already in %s)\n", filename)
	}

	fmt.Println("\nTo update these games, run:")
	fmt.Printf("   gsca update --args \"your launch options\" --allow %s\n", filename)

	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	// Use provided file path or default
	filePath := listFile
	if len(args) > 0 {
		filePath = args[0]
	}

	// Get Steam path
	var err error
	if steamPath == "" {
		steamPath, err = steam.GetSteamPath()
		if err != nil {
			return fmt.Errorf("failed to detect Steam path: %w", err)
		}
	}

	// Get user ID
	if userID == "" {
		userID, err = steam.GetUserID(steamPath)
		if err != nil {
			return fmt.Errorf("failed to detect user ID: %w", err)
		}
	}

	localConfigPath := steam.GetLocalConfigPath(steamPath, userID)

	// Load game mapping (for name/ID resolution)
	fmt.Println("Loading game library...")
	mapping, err := steam.GetGameMapping(steamPath)
	if err != nil {
		return fmt.Errorf("failed to get game mapping: %w", err)
	}

	// Get all games for detailed info
	allGames, err := steam.GetAllGames(steamPath, localConfigPath)
	if err != nil {
		return fmt.Errorf("failed to get game library: %w", err)
	}

	// Build app ID to game info map (filter Steam tools by default)
	gameInfoMap := make(map[string]steam.GameInfo)
	for _, game := range allGames {
		// Skip Steam tools unless --include-tools is set
		if !includeTools && isSteamTool(game.Name) {
			continue
		}
		gameInfoMap[game.AppID] = game
	}

	// Load the list file
	entries, err := steam.LoadFilterList(filePath)
	if err != nil {
		return fmt.Errorf("failed to load list file: %w", err)
	}

	if len(entries) == 0 {
		fmt.Printf("\nWARNING:File is empty: %s\n", filePath)
		return nil
	}

	// Resolve entries and display
	fmt.Printf("\nGames in %s:\n\n", filePath)

	for i, entry := range entries {
		entryLower := strings.ToLower(entry)

		// First check if entry is an app ID (numeric check or exists in gameInfoMap)
		isNumeric := true
		for _, c := range entry {
			if c < '0' || c > '9' {
				isNumeric = false
				break
			}
		}

		if isNumeric {
			// Entry looks like an app ID - check if it's in our library
			if gameInfo, found := gameInfoMap[entry]; found {
				status := ""
				if !gameInfo.Installed {
					status = statusNotInstalled
				}

				if gameInfo.Name == entry {
					// No name available (uninstalled), just show ID
					fmt.Printf("[%d] App ID: %s%s\n", i+1, entry, status)
				} else {
					// Show both name and ID
					fmt.Printf("[%d] %s\n", i+1, gameInfo.Name)
					fmt.Printf("    App ID: %s%s\n", entry, status)
				}

				if gameInfo.LaunchOptions != "" {
					fmt.Printf("    Launch Options: %s\n", gameInfo.LaunchOptions)
				}
			} else {
				fmt.Printf("[%d] App ID: %s [NOT IN LIBRARY]\n", i+1, entry)
			}
		} else if appID, exists := mapping[entryLower]; exists {
			// Entry is a game name
			if gameInfo, found := gameInfoMap[appID]; found {
				status := ""
				if !gameInfo.Installed {
					status = statusNotInstalled
				}

				fmt.Printf("[%d] %s\n", i+1, entry)
				fmt.Printf("    App ID: %s%s\n", appID, status)

				if gameInfo.LaunchOptions != "" {
					fmt.Printf("    Launch Options: %s\n", gameInfo.LaunchOptions)
				}
			} else {
				fmt.Printf("[%d] %s\n", i+1, entry)
				fmt.Printf("    App ID: %s [NOT IN LIBRARY]\n", appID)
			}
		} else {
			// Entry not found
			fmt.Printf("[%d] %s [NOT FOUND]\n", i+1, entry)
		}

		fmt.Println()
	}

	fmt.Printf("Total: %d game(s)\n", len(entries))

	return nil
}

// parseSelection parses user input like "1,3,5", "1-3", or "*" into indices
func parseSelection(input string, max int) []int {
	input = strings.TrimSpace(input)

	// Check for wildcard - select all
	if input == "*" {
		indices := make([]int, max)
		for i := 0; i < max; i++ {
			indices[i] = i
		}
		return indices
	}

	var indices []int
	seen := make(map[int]bool)

	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Check for range (e.g., "1-3")
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) == 2 {
				start, err1 := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
				end, err2 := strconv.Atoi(strings.TrimSpace(rangeParts[1]))

				if err1 == nil && err2 == nil && start > 0 && end <= max && start <= end {
					for i := start; i <= end; i++ {
						if !seen[i-1] {
							indices = append(indices, i-1)
							seen[i-1] = true
						}
					}
				}
			}
		} else {
			// Single number
			num, err := strconv.Atoi(part)
			if err == nil && num > 0 && num <= max {
				if !seen[num-1] {
					indices = append(indices, num-1)
					seen[num-1] = true
				}
			}
		}
	}

	return indices
}

// isSteamTool checks if a game name is a Steam tool (Proton, Runtime, etc.)
func isSteamTool(name string) bool {
	return strings.Contains(name, "Proton") || strings.Contains(name, "Runtime")
}

// loadAndResolveFilterList loads a filter list file and resolves game IDs
func loadAndResolveFilterList(filePath, listType string, mapping map[string]string, ignoreMissing bool) ([]string, error) {
	fmt.Printf("Loading %s list from: %s\n", listType, filePath)
	items, err := steam.LoadFilterList(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s list: %w", listType, err)
	}

	resolvedIDs, notFound := steam.ResolveGameIDs(items, mapping)
	if len(notFound) > 0 {
		fmt.Printf("\nERROR: Invalid entries in %s list (%d non-numeric entries):\n", listType, len(notFound))
		for _, item := range notFound {
			fmt.Printf("  - %s\n", item)
		}

		if !ignoreMissing {
			fmt.Println("\nAllow/deny lists only support numeric Steam app IDs.")
			fmt.Println("Use 'gsca query' to search for games and get their app IDs.")
			fmt.Println("Use 'gsca list' to view app IDs from existing lists.")
			fmt.Printf("\nUse --ignore-missing to continue anyway, or fix the %s list.\n", listType)
			return nil, fmt.Errorf("refusing to continue with missing games in %s list", listType)
		}

		fmt.Println("\nWARNING: Continuing anyway due to --ignore-missing flag")
	}

	return resolvedIDs, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
