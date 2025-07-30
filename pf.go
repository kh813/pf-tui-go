package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunSudoCmd executes a command with sudo.
func RunSudoCmd(args ...string) (string, error) {
	if testMode {
		LogInfo(fmt.Sprintf("Skipping sudo command in test mode: %s", strings.Join(args, " ")))
		return "", nil
	}
	LogInfo(fmt.Sprintf("Executing sudo command: %s", strings.Join(args, " ")))
	cmd := exec.Command("sudo", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		LogError(fmt.Sprintf("Sudo command failed: %s - %v - %s", strings.Join(args, " "), err, out.String()))
	}
	return out.String(), err
}

// setupPfConf ensures that the pf.conf file is configured to load the pf-tui anchor.
func SetupPfConf() error {
	const pfConfPath = "/etc/pf.conf"
	const anchorName = "pf-tui"
	const anchorFile = "/etc/pf.anchors/pf-tui"

	// The lines we need in pf.conf
	rdrAnchorLine := fmt.Sprintf("rdr-anchor \"%s\"", anchorName)
	anchorLine := fmt.Sprintf("anchor \"%s\"", anchorName)
	loadAnchorLine := fmt.Sprintf("load anchor \"%s\" from \"%s\"", anchorName, anchorFile)

	// Read the current pf.conf
	LogInfo(fmt.Sprintf("Checking pf.conf for anchor rules at %s", pfConfPath))
	content, err := RunSudoCmd("cat", pfConfPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", pfConfPath, err)
	}

	// Check if our lines are already present
	hasRdrAnchor := strings.Contains(content, rdrAnchorLine)
	hasAnchor := strings.Contains(content, anchorLine)
	hasLoadAnchor := strings.Contains(content, loadAnchorLine)

	if hasRdrAnchor && hasAnchor && hasLoadAnchor {
		// Everything is already set up
		return nil
	}

	// If not, we need to add them.
	var toAppend strings.Builder
	toAppend.WriteString("\n# pf-tui anchor point\n")
	if !hasRdrAnchor {
		toAppend.WriteString(rdrAnchorLine + "\n")
	}
	if !hasAnchor {
		toAppend.WriteString(anchorLine + "\n")
	}
	if !hasLoadAnchor {
		toAppend.WriteString(loadAnchorLine + "\n")
	}

	// Append the new lines to pf.conf
	LogInfo(fmt.Sprintf("Updating %s with new anchor rules", pfConfPath))
	cmd := exec.Command("sudo", "tee", "-a", pfConfPath)
	cmd.Stdin = strings.NewReader(toAppend.String())
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to append to %s: %w, output: %s", pfConfPath, err, out.String())
	}

	return nil
}


// ApplyRules applies the given rules string to pf.
func ApplyRules(rules string) (string, error) {
	if testMode {
		return "", nil
	}
	// Write rules to a temporary file for inspection
	tmpfile, err := os.CreateTemp("", "pf-tui-rules-*.conf")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name()) // clean up
	defer tmpfile.Close()

	if _, err := tmpfile.WriteString(rules); err != nil {
		return "", fmt.Errorf("failed to write rules to temp file: %w", err)
	}
	LogInfo(fmt.Sprintf("Generated pf.conf content written to temporary file: %s", tmpfile.Name()))

	// Write rules to the anchor file
	anchorPath := "/etc/pf.anchors/pf-tui"
	LogInfo(fmt.Sprintf("Applying rules to %s", anchorPath))
	cmd := exec.Command("sudo", "tee", anchorPath)
	cmd.Stdin = strings.NewReader(rules)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to write to anchor file: %w, output: %s", err, out.String())
	}

	// Load the rules from the anchor
	return RunSudoCmd("pfctl", "-f", anchorPath)
}

// GetCurrentRules returns the currently loaded pf rules.
func GetCurrentRules() (string, error) {
	if testMode {
		return "pass out on lo0 all\nblock in on lo0 all", nil
	}
	out, err := RunSudoCmd("pfctl", "-s", "rules")
	if err != nil {
		return "", err
	}
	var filteredRules []string
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "ALTQ") {
			filteredRules = append(filteredRules, line)
		}
	}
	return strings.Join(filteredRules, "\n"), nil
}

// GetPfStatus returns the status of pf ("Enabled" or "Disabled").
func GetPfStatus() (string, error) {
	if testMode {
		return "Enabled", nil
	}
	out, err := RunSudoCmd("pfctl", "-s", "info")
	if err != nil {
		// If pfctl returns an error, it might be because PF is disabled.
		// The output often contains "pf not running".
		if strings.Contains(out, "pf not running") {
			return "Disabled", nil
		}
		return "", err
	}

	// Parse the output to find the status line.
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "Status:") {
			status := strings.TrimSpace(strings.TrimPrefix(line, "Status:"))
			if strings.HasPrefix(status, "Enabled") {
				return "Enabled", nil
			}
		}
	}

	return "Disabled", nil
}

// EnablePf enables the pf firewall.
func EnablePf() (string, error) {
	if testMode {
		return "", nil
	}
	return RunSudoCmd("pfctl", "-e")
}

// DisablePf disables the pf firewall.
func DisablePf() (string, error) {
	if testMode {
		return "", nil
	}
	return RunSudoCmd("pfctl", "-d")
}

// GetPfInfo returns detailed statistics from pf.
func GetPfInfo() (string, error) {
	if testMode {
		return "State Table      Total             0", nil
	}
	return RunSudoCmd("pfctl", "-s", "info")
}

// ParseLiveRules parses the output of `pfctl -s rules` and returns a slice of FirewallRule structs.
func ParseLiveRules(output string) ([]FirewallRule, error) {
	var rules []FirewallRule
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue // Not a valid rule
		}

		rule := FirewallRule{}

		// Basic rule components
		rule.Action = parts[0]
		rule.Direction = parts[1]

		// Extract other parts of the rule
		for i := 2; i < len(parts); i++ {
			switch parts[i] {
			case "quick":
				rule.Quick = true
			case "on":
				i++
				rule.Interface = parts[i]
			case "proto":
				i++
				rule.Protocol = parts[i]
			case "from":
				i++
				rule.Source = parts[i]
			case "to":
				i++
				rule.Destination = parts[i]
			case "port":
				i++
				rule.Port = parts[i]
			case "keep":
				i++ // state
				rule.KeepState = true
			}
		}

		rules = append(rules, rule)
	}
	return rules, nil
}


const plistPath = "/Library/LaunchDaemons/com.user.pftui.plist"

// CheckPfStartupStatus checks if the launchd plist exists.
func CheckPfStartupStatus() (string, error) {
	if testMode {
		return "Enabled", nil
	}
	if _, err := os.Stat(plistPath); err == nil {
		return "Enabled", nil
	} else if os.IsNotExist(err) {
		return "Disabled", nil
	} else {
		return "Unknown", err
	}
}


// EnablePfOnStartup configures pf to start on boot.
func EnablePfOnStartup() (string, error) {
	if testMode {
		return "", nil
	}
	LogInfo(fmt.Sprintf("Enabling pf on startup by creating %s", plistPath))
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.user.pftui</string>
    <key>ProgramArguments</key>
    <array>
        <string>/sbin/pfctl</string>
        <string>-e</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>StandardErrorPath</key>
    <string>/tmp/com.user.pftui.stderr</string>
    <key>StandardOutPath</key>
    <string>/tmp/com.user.pftui.stdout</string>
</dict>
</plist>`

	// Write the plist file
	cmd := exec.Command("sudo", "tee", plistPath)
	cmd.Stdin = strings.NewReader(plistContent)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to write plist file: %w, output: %s", err, out.String())
	}

	// Load the launchd job
	return RunSudoCmd("launchctl", "load", "-w", plistPath)
}

// DisablePfOnStartup prevents pf from starting on boot.
func DisablePfOnStartup() (string, error) {
	if testMode {
		return "", nil
	}
	LogInfo(fmt.Sprintf("Disabling pf on startup by removing %s", plistPath))
	// Unload the launchd job
	_, err := RunSudoCmd("launchctl", "unload", "-w", plistPath)
	if err != nil {
		// Ignore errors if the job is not loaded
	}

	// Remove the plist file
	return RunSudoCmd("rm", plistPath)
}

