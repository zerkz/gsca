package steam

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// IsSteamRunning checks if Steam is currently running
func IsSteamRunning() (bool, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case osLinux:
		cmd = exec.Command("pgrep", "-x", "steam")
	case osDarwin:
		cmd = exec.Command("pgrep", "-x", "steam_osx")
	case osWindows:
		cmd = exec.Command("tasklist", "/FI", "IMAGENAME eq steam.exe", "/NH")
	default:
		return false, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		// pgrep returns exit code 1 if no process found
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, err
	}

	// Check output
	outputStr := strings.TrimSpace(string(output))

	// On Windows, tasklist returns "INFO: No tasks..." when process not found
	if runtime.GOOS == osWindows {
		return !strings.Contains(outputStr, "No tasks"), nil
	}

	return outputStr != "", nil
}

// CloseSteam attempts to gracefully close Steam
func CloseSteam() error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case osLinux:
		// Use steam's own shutdown command
		cmd = exec.Command("steam", "-shutdown")
	case osDarwin:
		// macOS: Use AppleScript to quit gracefully
		// Note: osascript may return exit code 1 even when quit succeeds,
		// so we ignore the error and let the caller poll IsSteamRunning()
		cmd := exec.Command("osascript", "-e", "quit app \"Steam\"")
		_ = cmd.Run()
		return nil
	case osWindows:
		// Windows: Force kill Steam - graceful shutdown doesn't work reliably
		cmd = exec.Command("taskkill", "/F", "/IM", "steam.exe")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Run()
}

// StartSteam attempts to start Steam
func StartSteam() error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case osLinux:
		cmd = exec.Command("steam")
	case osDarwin:
		// macOS: Use open command
		cmd = exec.Command("open", "-a", "Steam")
	case osWindows:
		// Windows: Use steam:// protocol which works regardless of install location
		// The empty string "" is needed as the window title parameter for start command
		cmd = exec.Command("cmd", "/C", "start", "", "steam://open/main")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// OpenFile opens a file with the default system application
func OpenFile(filePath string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case osLinux:
		// Linux: Use xdg-open
		cmd = exec.Command("xdg-open", filePath)
	case osDarwin:
		// macOS: Use open command
		cmd = exec.Command("open", filePath)
	case osWindows:
		// Windows: Use start command
		cmd = exec.Command("cmd", "/C", "start", "", filePath)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
