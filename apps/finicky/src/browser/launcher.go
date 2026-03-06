package browser

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"al.essio.dev/pkg/shellescape"
	"finicky/util"
)

//go:embed browsers.json
var browsersJsonData []byte

type BrowserResult struct {
	Browser BrowserConfig `json:"browser"`
	Error   string        `json:"error"`
}

type BrowserConfig struct {
	Name             string   `json:"name"`
	AppType          string   `json:"appType"`
	OpenInBackground *bool    `json:"openInBackground"`
	Profile          string   `json:"profile"`
	Args             []string `json:"args"`
	URL              string   `json:"url"`
	NewWindow        *bool    `json:"newWindow"`
	Incognito        *bool    `json:"incognito"`
	NewTab           *bool    `json:"newTab"`
}

type browserInfo struct {
	ConfigDirRelative string `json:"config_dir_relative"`
	ID                string `json:"id"`
	AppName           string `json:"app_name"`
	Type              string `json:"type"`
}

func LaunchBrowser(config BrowserConfig, dryRun bool, openInBackgroundByDefault bool) error {
	if config.AppType == "none" {
		slog.Info("AppType is 'none', not launching any browser")
		return nil
	}

	slog.Info("Starting browser", "name", config.Name, "url", config.URL)

	// Handle profile, window flags, and custom args
	profileArgument, hasProfile := resolveBrowserProfileArgument(config.Name, config.Profile)
	windowFlags, hasWindowFlags := resolveChromiumWindowFlags(config.Name, config)
	hasCustomArgs := len(config.Args) > 0
	isChromium := isChromiumBrowser(config.Name)

	var cmd *exec.Cmd

	// For Chromium browsers with flags, execute the browser binary directly
	if isChromium && (hasProfile || hasWindowFlags || hasCustomArgs) {
		executablePath, err := getBrowserExecutablePath(config.Name, config.AppType)
		if err != nil {
			return fmt.Errorf("failed to get browser executable path: %v", err)
		}

		var chromiumArgs []string

		// Add profile argument first if present
		if hasProfile {
			chromiumArgs = append(chromiumArgs, profileArgument)
		}

		// Add window flags second (before custom args)
		if hasWindowFlags {
			chromiumArgs = append(chromiumArgs, windowFlags...)
		}

		// Add custom args or URL
		if hasCustomArgs {
			chromiumArgs = append(chromiumArgs, config.Args...)
		} else {
			chromiumArgs = append(chromiumArgs, config.URL)
		}

		cmd = exec.Command(executablePath, chromiumArgs...)
	} else {
		// Use the traditional `open` command for non-Chromium browsers or simple launches
		var openArgs []string

		if config.AppType == "bundleId" {
			openArgs = []string{"-b", config.Name}
		} else {
			openArgs = []string{"-a", config.Name}
		}

		var openInBackground bool = openInBackgroundByDefault

		if config.OpenInBackground != nil {
			openInBackground = *config.OpenInBackground
		}

		if openInBackground {
			openArgs = append(openArgs, "-g")
		}

		// Add -n flag if profile is used OR if newWindow is set
		if hasProfile || (hasWindowFlags && config.NewWindow != nil && *config.NewWindow) {
			openArgs = append(openArgs, "-n")
		}

		// Add --args if we have profile args, window flags, or custom args
		if hasProfile || hasWindowFlags || hasCustomArgs {
			if !slices.Contains(config.Args, "--args") {
				openArgs = append(openArgs, "--args")
			}

			// Add profile argument first if present
			if hasProfile {
				openArgs = append(openArgs, profileArgument)
			}

			// Add window flags second (before custom args)
			if hasWindowFlags {
				openArgs = append(openArgs, windowFlags...)
			}

			// Add custom args or URL
			if hasCustomArgs {
				openArgs = append(openArgs, config.Args...)
			} else {
				openArgs = append(openArgs, config.URL)
			}
		} else {
			// No special args, just add the URL
			openArgs = append(openArgs, config.URL)
		}

		cmd = exec.Command("open", openArgs...)
	}

	// Pretty print the command with proper escaping
	prettyCmd := formatCommand(cmd.Path, cmd.Args)

	if dryRun {
		slog.Debug("Would run command (dry run)", "command", prettyCmd)
		return nil
	} else {
		slog.Debug("Run command", "command", prettyCmd)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	stderrBytes, err := io.ReadAll(stderr)
	if err != nil {
		return fmt.Errorf("error reading stderr: %v", err)
	}

	stdoutBytes, err := io.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("error reading stdout: %v", err)
	}

	cmdErr := cmd.Wait()

	if len(stderrBytes) > 0 {
		slog.Error("Command returned error", "error", string(stderrBytes))
	}
	if len(stdoutBytes) > 0 {
		slog.Debug("Command returned output", "output", string(stdoutBytes))
	}

	if cmdErr != nil {
		return fmt.Errorf("command failed: %v", cmdErr)
	}

	return nil
}

func resolveBrowserProfileArgument(identifier string, profile string) (string, bool) {
	var browsersJson []browserInfo
	if err := json.Unmarshal(browsersJsonData, &browsersJson); err != nil {
		slog.Info("Error parsing browsers.json", "error", err)
		return "", false
	}

	// Try to find matching browser by bundle ID
	var matchedBrowser *browserInfo
	for _, browser := range browsersJson {
		if browser.ID == identifier || browser.AppName == identifier {
			matchedBrowser = &browser
			break
		}
	}

	if matchedBrowser == nil {
		return "", false
	}

	slog.Debug("Browser found in browsers.json", "identifier", identifier, "type", matchedBrowser.Type)

	if profile != "" {
		switch matchedBrowser.Type {
		case "Chromium":
			homeDir, err := util.UserHomeDir()
			if err != nil {
				slog.Info("Error getting home directory", "error", err)
				return "", false
			}

			localStatePath := filepath.Join(homeDir, "Library/Application Support", matchedBrowser.ConfigDirRelative, "Local State")
			profilePath, ok := parseProfiles(localStatePath, profile)
			if ok {
				return "--profile-directory=" + profilePath, true
			}
		default:
			slog.Info("Browser is not a Chromium browser, skipping profile detection", "identifier", identifier)
		}
	}

	return "", false
}

func parseProfiles(localStatePath string, profile string) (string, bool) {
	data, err := os.ReadFile(localStatePath)
	if err != nil {
		slog.Info("Error reading Local State file", "path", localStatePath, "error", err)
		return "", false
	}

	var localState map[string]interface{}
	if err := json.Unmarshal(data, &localState); err != nil {
		slog.Info("Error parsing Local State JSON", "error", err)
		return "", false
	}

	profiles, ok := localState["profile"].(map[string]interface{})
	if !ok {
		slog.Info("Could not find profile section in Local State")
		return "", false
	}

	infoCache, ok := profiles["info_cache"].(map[string]interface{})
	if !ok {
		slog.Info("Could not find info_cache in profile section")
		return "", false
	}

	// Look for the specified profile
	for profilePath, info := range infoCache {
		profileInfo, ok := info.(map[string]interface{})
		if !ok {
			continue
		}

		name, ok := profileInfo["name"].(string)
		if !ok {
			continue
		}

		if name == profile {
			slog.Info("Found profile by name", "name", name, "path", profilePath)
			return profilePath, true
		}
	}

	// If we didn't find the profile, try to find it by profile folder name
	slog.Debug("Could not find profile in browser profiles, trying to find by profile path", "profile", profile)
	for profilePath, info := range infoCache {
		if profilePath == profile {
			// Try to get the profile name of the profile we want the user to use instead
			if profileInfo, ok := info.(map[string]interface{}); ok {
				if name, ok := profileInfo["name"].(string); ok {
					slog.Warn("Found profile using profile path", "path", profilePath, "name", name, "suggestion", "Please use the profile name instead")
				}
			}
			return profilePath, true
		}
	}

	var profileNames []string
	for _, info := range infoCache {
		profileInfo, ok := info.(map[string]interface{})
		if !ok {
			continue
		}

		name, ok := profileInfo["name"].(string)
		if !ok {
			continue
		}

		profileNames = append(profileNames, name)
	}
	slog.Warn("Could not find profile in browser profiles.", "Expected profile", profile, "Available profiles", strings.Join(profileNames, ", "))

	return "", false
}

// resolveChromiumWindowFlags converts convenience fields (newWindow, incognito, newTab)
// into Chromium command-line flags for browsers identified as type="Chromium" in browsers.json
func resolveChromiumWindowFlags(identifier string, config BrowserConfig) ([]string, bool) {
	var browsersJson []browserInfo
	if err := json.Unmarshal(browsersJsonData, &browsersJson); err != nil {
		slog.Debug("Error parsing browsers.json", "error", err)
		return nil, false
	}

	// Find browser in browsers.json
	var matchedBrowser *browserInfo
	for _, browser := range browsersJson {
		if browser.ID == identifier || browser.AppName == identifier {
			matchedBrowser = &browser
			break
		}
	}

	if matchedBrowser == nil || matchedBrowser.Type != "Chromium" {
		return nil, false
	}

	slog.Debug("Resolving Chromium window flags", "identifier", identifier)

	var flags []string

	// Convert convenience fields to Chrome flags
	if config.NewWindow != nil && *config.NewWindow {
		flags = append(flags, "--new-window")
	}

	if config.Incognito != nil && *config.Incognito {
		flags = append(flags, "--incognito")
	}

	// Note: newTab is default Chrome behavior, included for completeness
	// No flag needed since Chrome opens new tabs by default

	if len(flags) > 0 {
		return flags, true
	}

	return nil, false
}

// formatCommand returns a properly shell-escaped string representation of the command
func formatCommand(path string, args []string) string {
	if len(args) == 0 {
		return shellescape.Quote(path)
	}

	quotedArgs := make([]string, len(args))
	for i, arg := range args {
		quotedArgs[i] = shellescape.Quote(arg)
	}

	return strings.Join(quotedArgs, " ")
}

// isChromiumBrowser checks if the given browser identifier is a Chromium-based browser
func isChromiumBrowser(identifier string) bool {
	var browsersJson []browserInfo
	if err := json.Unmarshal(browsersJsonData, &browsersJson); err != nil {
		slog.Debug("Error parsing browsers.json", "error", err)
		return false
	}

	for _, browser := range browsersJson {
		if (browser.ID == identifier || browser.AppName == identifier) && browser.Type == "Chromium" {
			return true
		}
	}

	return false
}

// getBrowserExecutablePath returns the path to the browser's executable binary
func getBrowserExecutablePath(identifier string, appType string) (string, error) {
	var bundleID string

	if appType == "bundleId" {
		bundleID = identifier
	} else {
		// Get bundle ID from app name
		var browsersJson []browserInfo
		if err := json.Unmarshal(browsersJsonData, &browsersJson); err != nil {
			return "", fmt.Errorf("error parsing browsers.json: %v", err)
		}

		for _, browser := range browsersJson {
			if browser.AppName == identifier {
				bundleID = browser.ID
				break
			}
		}

		if bundleID == "" {
			return "", fmt.Errorf("could not find bundle ID for app: %s", identifier)
		}
	}

	// Use mdfind to locate the app bundle
	cmd := exec.Command("mdfind", fmt.Sprintf("kMDItemCFBundleIdentifier == '%s'", bundleID))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error finding app bundle: %v", err)
	}

	appPath := strings.TrimSpace(string(output))
	if appPath == "" {
		return "", fmt.Errorf("could not find app bundle for: %s", bundleID)
	}

	// Split by newline and take the first result
	appPaths := strings.Split(appPath, "\n")
	if len(appPaths) > 0 {
		appPath = appPaths[0]
	}

	// For Chromium browsers, the executable is typically at:
	// /Applications/Google Chrome.app/Contents/MacOS/Google Chrome
	executableName := filepath.Base(strings.TrimSuffix(appPath, ".app"))
	executablePath := filepath.Join(appPath, "Contents", "MacOS", executableName)

	// Verify the executable exists
	if _, err := os.Stat(executablePath); err != nil {
		return "", fmt.Errorf("executable not found at %s: %v", executablePath, err)
	}

	return executablePath, nil
}
