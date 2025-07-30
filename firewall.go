package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	
)

// FirewallRule represents a single filter rule.
type FirewallRule struct {
	Action      string `json:"action"`
	Direction   string `json:"direction"`
	Quick       bool   `json:"quick"`
	Interface   string `json:"interface"`
	Protocol    string `json:"protocol"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Port        string `json:"port"`
	KeepState   bool   `json:"keep_state"`
	Description string `json:"description"`
}

// PortForwardingRule represents a single port forwarding (RDR) rule.
type PortForwardingRule struct {
	Interface    string `json:"interface"`
	Protocol     string `json:"protocol"`
	ExternalIP   string `json:"external_ip"`
	ExternalPort string `json:"external_port"`
	InternalIP   string `json:"internal_ip"`
	InternalPort string `json:"internal_port"`
	Description  string `json:"description"`
}

// Config holds all firewall and port forwarding rules.
type Config struct {
	FirewallRules      []FirewallRule       `json:"filter_rules"`
	PortForwardingRules []PortForwardingRule `json:"rdr_rules"`
}

// FirewallManager handles loading, saving, and generating firewall configurations.
type FirewallManager struct {
	Config *Config
}

// NewFirewallManager creates a new FirewallManager.
func NewFirewallManager() *FirewallManager {
	return &FirewallManager{
		Config: &Config{
			FirewallRules:      []FirewallRule{},
			PortForwardingRules: []PortForwardingRule{},
		},
	}
}

func getDefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "pf-tui", "rules.json"), nil
}

func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(home, ".config", "pf-tui")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return "", err
	}
	return configPath, nil
}

// LoadConfig loads the firewall configuration from the default JSON file.
func (fm *FirewallManager) LoadConfig() error {
	path, err := getDefaultConfigPath()
	if err != nil {
		LogInfo(fmt.Sprintf("Error getting default config path: %v", err))
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			LogWarn("Configuration file not found. A new empty configuration will be created on next save.")
				fm.Config = &Config{
				FirewallRules:      []FirewallRule{},
				PortForwardingRules: []PortForwardingRule{},
			}
			return nil
		}
		LogError(fmt.Sprintf("Failed to read configuration file %s: %v", path, err))
		return err
	}

	if err := json.Unmarshal(data, fm.Config); err != nil {
		LogError(fmt.Sprintf("Failed to parse JSON from configuration file %s: %v", path, err))
		return err
	}

	LogInfo(fmt.Sprintf("Successfully loaded configuration from %s", path))
	return nil
}

// ImportConfigFile backs up the existing config and replaces it with a new one.
func (fm *FirewallManager) ImportConfigFile(sourcePath string) error {
	defaultPath, err := getDefaultConfigPath()
	if err != nil {
		LogInfo(fmt.Sprintf("Error getting default config path: %v", err))
		return err
	}

	// Create backup if the default config file exists
	if _, err := os.Stat(defaultPath); err == nil {
		backupPath := defaultPath + ".bak"
		if err := os.Rename(defaultPath, backupPath); err != nil {
			LogError(fmt.Sprintf("Failed to create backup file %s: %v", backupPath, err))
			return fmt.Errorf("failed to create backup: %w", err)
		}
	} else if !os.IsNotExist(err) {
		LogError(fmt.Sprintf("Error checking for existing config file: %v", err))
		return err // Other error like permission denied
	}

	// Read the new config file
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		LogError(fmt.Sprintf("Failed to read import file %s: %v", sourcePath, err))
		return fmt.Errorf("failed to read import file: %w", err)
	}

	// Ensure the config directory exists
	if err := os.MkdirAll(filepath.Dir(defaultPath), 0755); err != nil {
		LogError(fmt.Sprintf("Error creating config directory: %v", err))
		return err
	}

	// Write the new config file to the default path
	LogInfo(fmt.Sprintf("Importing configuration from %s", sourcePath))
	if err := os.WriteFile(defaultPath, data, 0644); err != nil {
		LogError(fmt.Sprintf("Failed to write new config file %s: %v", defaultPath, err))
		return fmt.Errorf("failed to write new config file: %w", err)
	}

	LogInfo(fmt.Sprintf("Imported configuration from %s. Previous config backed up to %s.bak", sourcePath, defaultPath))

	// Load the new config into the manager
	return fm.LoadConfig()
}


// SaveConfig saves the firewall configuration to the default JSON file.
func (fm *FirewallManager) SaveConfig() error {
	path, err := getDefaultConfigPath()
	if err != nil {
		LogInfo(fmt.Sprintf("Error getting default config path: %v", err))
		return err
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		LogError(fmt.Sprintf("Error creating config directory: %v", err))
		return err
	}

	data, err := json.MarshalIndent(fm.Config, "", "  ")
	if err != nil {
		LogError(fmt.Sprintf("Failed to marshal config to JSON: %v", err))
		return err
	}

	LogInfo(fmt.Sprintf("Saving configuration to %s", path))
	if err := os.WriteFile(path, data, 0644); err != nil {
		LogError(fmt.Sprintf("Failed to write to configuration file %s: %v", path, err))
		return err
	}

	LogInfo(fmt.Sprintf("Saved configuration to %s", path))
	return nil
}

// SaveConfigAs saves the current configuration to a different file.
func (fm *FirewallManager) SaveConfigAs(path string) error {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		LogError(fmt.Sprintf("Error creating config directory: %v", err))
		return err
	}

	data, err := json.MarshalIndent(fm.Config, "", "  ")
	if err != nil {
		LogError(fmt.Sprintf("Failed to marshal config to JSON: %v", err))
		return err
	}

	LogInfo(fmt.Sprintf("Exporting configuration to %s", path))
	if err := os.WriteFile(path, data, 0644); err != nil {
		LogError(fmt.Sprintf("Failed to write to configuration file %s: %v", path, err))
		return err
	}

		LogInfo(fmt.Sprintf("Exported configuration to %s", path))
	return nil
}

// AddFirewallRule adds a new firewall rule to the configuration file.
func (fm *FirewallManager) AddFirewallRule(rule FirewallRule) error {
	if err := fm.LoadConfig(); err != nil {
		return err
	}
	fm.Config.FirewallRules = append(fm.Config.FirewallRules, rule)
	LogInfo(fmt.Sprintf("Added firewall rule: %+v", rule))
	return fm.SaveConfig()
}

// UpdateFirewallRule updates an existing firewall rule in the configuration file.
func (fm *FirewallManager) UpdateFirewallRule(index int, rule FirewallRule) error {
	if err := fm.LoadConfig(); err != nil {
		return err
	}
	if index < 0 || index >= len(fm.Config.FirewallRules) {
		return fmt.Errorf("invalid rule index")
	}
	fm.Config.FirewallRules[index] = rule
	LogInfo(fmt.Sprintf("Updated firewall rule at index %d: %+v", index, rule))
	return fm.SaveConfig()
}

// DeleteFirewallRule deletes a firewall rule from the configuration file.
func (fm *FirewallManager) DeleteFirewallRule(index int) error {
	if err := fm.LoadConfig(); err != nil {
		return err
	}
	if index < 0 || index >= len(fm.Config.FirewallRules) {
		return fmt.Errorf("invalid rule index")
	}
	LogInfo(fmt.Sprintf("Deleted firewall rule at index %d: %+v", index, fm.Config.FirewallRules[index]))
	fm.Config.FirewallRules = append(fm.Config.FirewallRules[:index], fm.Config.FirewallRules[index+1:]...)
	return fm.SaveConfig()
}

// MoveFirewallRule moves a firewall rule from one index to another.
func (fm *FirewallManager) MoveFirewallRule(from, to int) {
	if from < 0 || from >= len(fm.Config.FirewallRules) || to < 0 || to >= len(fm.Config.FirewallRules) {
		return
	}
	if from == to {
		return
	}

	rule := fm.Config.FirewallRules[from]

	// Remove element
	tmp := append(fm.Config.FirewallRules[:from], fm.Config.FirewallRules[from+1:]...)

	// Insert element at new position
	final := make([]FirewallRule, 0, len(fm.Config.FirewallRules))
	final = append(final, tmp[:to]...)
	final = append(final, rule)
	final = append(final, tmp[to:]...)

	fm.Config.FirewallRules = final
}

// AddPortForwardingRule adds a new port forwarding rule to the configuration file.
func (fm *FirewallManager) AddPortForwardingRule(rule PortForwardingRule) error {
	if err := fm.LoadConfig(); err != nil {
		return err
	}
	fm.Config.PortForwardingRules = append(fm.Config.PortForwardingRules, rule)
	LogInfo(fmt.Sprintf("Added port forwarding rule: %+v", rule))
	return fm.SaveConfig()
}

// UpdatePortForwardingRule updates an existing port forwarding rule in the configuration file.
func (fm *FirewallManager) UpdatePortForwardingRule(index int, rule PortForwardingRule) error {
	if err := fm.LoadConfig(); err != nil {
		return err
	}
	if index < 0 || index >= len(fm.Config.PortForwardingRules) {
		return fmt.Errorf("invalid rule index")
	}
	fm.Config.PortForwardingRules[index] = rule
	LogInfo(fmt.Sprintf("Updated port forwarding rule at index %d: %+v", index, rule))
	return fm.SaveConfig()
}

// DeletePortForwardingRule deletes a port forwarding rule from the configuration file.
func (fm *FirewallManager) DeletePortForwardingRule(index int) error {
	if err := fm.LoadConfig(); err != nil {
		return err
	}
	if index < 0 || index >= len(fm.Config.PortForwardingRules) {
		return fmt.Errorf("invalid rule index")
	}
	LogInfo(fmt.Sprintf("Deleted port forwarding rule at index %d: %+v", index, fm.Config.PortForwardingRules[index]))
	fm.Config.PortForwardingRules = append(fm.Config.PortForwardingRules[:index], fm.Config.PortForwardingRules[index+1:]...)
	return fm.SaveConfig()
}

// MovePortForwardingRule moves a port forwarding rule from one index to another.
func (fm *FirewallManager) MovePortForwardingRule(from, to int) {
	if from < 0 || from >= len(fm.Config.PortForwardingRules) || to < 0 || to >= len(fm.Config.PortForwardingRules) {
		return
	}
	if from == to {
		return
	}

	rule := fm.Config.PortForwardingRules[from]

	// Remove element
	tmp := append(fm.Config.PortForwardingRules[:from], fm.Config.PortForwardingRules[from+1:]...)

	// Insert element at new position
	final := make([]PortForwardingRule, 0, len(fm.Config.PortForwardingRules))
	final = append(final, tmp[:to]...)
	final = append(final, rule)
	final = append(final, tmp[to:]...)

	fm.Config.PortForwardingRules = final
}

// GeneratePfConf generates the content of the pf.conf file from the current rules.
func (fm *FirewallManager) GeneratePfConf() string {
	var builder strings.Builder

	// Port Forwarding Rules
	for _, rule := range fm.Config.PortForwardingRules {
		if rule.Description != "" {
			builder.WriteString(fmt.Sprintf("# %s\n", rule.Description))
		}

		var rdrStr string
		if rule.Interface == "any" {
			rdrStr = fmt.Sprintf("rdr proto %s from any to %s port %s -> %s port %s",
				rule.Protocol, rule.ExternalIP, rule.ExternalPort, rule.InternalIP, rule.InternalPort)
		} else {
			// If ExternalIP is "any", it means the rule applies to any IP on the specified interface.
			// In pf, "to (interface)" is used for this.
			toPart := rule.ExternalIP
			if toPart == "any" {
				toPart = fmt.Sprintf("(%s)", rule.Interface)
			}
			rdrStr = fmt.Sprintf("rdr on %s proto %s from any to %s port %s -> %s port %s",
				rule.Interface, rule.Protocol, toPart, rule.ExternalPort, rule.InternalIP, rule.InternalPort)
		}
		builder.WriteString(rdrStr + "\n")
	}

	// Firewall Rules
	for _, rule := range fm.Config.FirewallRules {
		if rule.Description != "" {
			builder.WriteString(fmt.Sprintf("# %s\n", rule.Description))
		}

		var protocols []string
		if rule.Protocol == "any" && rule.Port != "any" {
			protocols = []string{"tcp", "udp"}
		} else {
			protocols = strings.Split(rule.Protocol, ",")
		}

		for _, proto := range protocols {
			proto = strings.TrimSpace(proto)
			var parts []string
			parts = append(parts, rule.Action)
			parts = append(parts, rule.Direction)
			if rule.Quick {
				parts = append(parts, "quick")
			}
			if rule.Interface != "any" {
				parts = append(parts, "on", rule.Interface)
			}

			if proto == "any" && rule.Source == "any" && rule.Destination == "any" && rule.Port == "any" {
				parts = append(parts, "all")
			} else {
				if proto != "any" {
					parts = append(parts, "proto", proto)
				}

				if rule.Source != "any" || rule.Destination != "any" {
					parts = append(parts, "from", rule.Source, "to", rule.Destination)
				} else if rule.Source == "any" && rule.Destination == "any" && rule.Port != "any" {
					parts = append(parts, "from", "any", "to", "any")
				}

				if rule.Port != "any" && (proto == "tcp" || proto == "udp") {
					portStr := rule.Port
					// If the port string contains a comma, it's a list of ports, so wrap in curly braces.
					// If it contains a colon or hyphen, it's a range, so replace hyphen with colon and wrap in curly braces.
					if strings.Contains(portStr, ",") || strings.Contains(portStr, "-") || strings.Contains(portStr, ":") {
						portStr = strings.ReplaceAll(portStr, "-", ":") // Replace hyphen with colon for ranges
						portStr = fmt.Sprintf("{%s}", portStr)
					}
					parts = append(parts, "port", portStr)
				}
			}

			if rule.KeepState {
				parts = append(parts, "keep state")
			}

			builder.WriteString(strings.Join(parts, " ") + "\n")
		}
	}

	return builder.String()
}

func EnsureConfigDirExists() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}
	return os.MkdirAll(path, 0755)
}

