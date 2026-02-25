package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ConfigManager struct {
	ConfigDir            string
	SynxConfig           string
	ExcludeCfg           string
	SynxConfigMachine    string
	ExcludeCfgMachine    string
	Hostname             string
	UsingMachineTargets  bool
	UsingMachineExcludes bool
	Targets              []string
	Excludes             []string
}

func NewConfigManager() (*ConfigManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	cfgDir := filepath.Join(home, ".config", "synx")

	// Create config dir if not exists
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config dir: %w", err)
	}

	hostname, _ := os.Hostname()

	return &ConfigManager{
		ConfigDir:         cfgDir,
		SynxConfig:        filepath.Join(cfgDir, "synx.conf"),
		ExcludeCfg:        filepath.Join(cfgDir, "exclude.conf"),
		SynxConfigMachine: filepath.Join(cfgDir, "synx.conf."+hostname),
		ExcludeCfgMachine: filepath.Join(cfgDir, "exclude.conf."+hostname),
		Hostname:          hostname,
		Targets:           []string{},
		Excludes:          []string{},
	}, nil
}

func (c *ConfigManager) Load() error {
	// Ensure base targets
	if _, err := os.Stat(c.SynxConfig); os.IsNotExist(err) {
		defaultTargets := []string{"hypr", "foot", "kitty", "fastfetch", "alacritty"}
		err = c.SaveTargets(defaultTargets)
		if err != nil {
			return err
		}
	}

	// Ensure base excludes
	if _, err := os.Stat(c.ExcludeCfg); os.IsNotExist(err) {
		err = c.SaveExcludes([]string{})
		if err != nil {
			return err
		}
	}

	// Load targets: machine-specific replaces base if it exists
	if _, err := os.Stat(c.SynxConfigMachine); err == nil {
		targets, err := readFileLines(c.SynxConfigMachine)
		if err != nil {
			return err
		}
		c.Targets = targets
		c.UsingMachineTargets = true
	} else {
		targets, err := readFileLines(c.SynxConfig)
		if err != nil {
			return err
		}
		c.Targets = targets
	}

	// Load excludes: base + machine-specific (appended)
	excludes, err := readFileLines(c.ExcludeCfg)
	if err != nil {
		return err
	}
	c.Excludes = excludes

	if _, err := os.Stat(c.ExcludeCfgMachine); err == nil {
		machineExcludes, err := readFileLines(c.ExcludeCfgMachine)
		if err != nil {
			return err
		}
		c.Excludes = append(c.Excludes, machineExcludes...)
		c.UsingMachineExcludes = true
	}

	return nil
}

func (c *ConfigManager) SaveTargets(targets []string) error {
	return writeLines(c.SynxConfig, targets, "Synx tracked dotfiles")
}

func (c *ConfigManager) SaveExcludes(excludes []string) error {
	return writeLines(c.ExcludeCfg, excludes, "Exclude patterns for machine-specific files\n# One pattern per line")
}

func (c *ConfigManager) SaveTargetsMachine(targets []string) error {
	return writeLines(c.SynxConfigMachine, targets, "Synx tracked dotfiles for "+c.Hostname)
}

func (c *ConfigManager) SaveExcludesMachine(excludes []string) error {
	return writeLines(c.ExcludeCfgMachine, excludes, "Exclude patterns for "+c.Hostname)
}

// MachineExcludes returns only the machine-specific excludes (not including base).
func (c *ConfigManager) MachineExcludes() ([]string, error) {
	if _, err := os.Stat(c.ExcludeCfgMachine); os.IsNotExist(err) {
		return []string{}, nil
	}
	return readFileLines(c.ExcludeCfgMachine)
}

// MachineTargets returns only the machine-specific targets.
func (c *ConfigManager) MachineTargets() ([]string, error) {
	if _, err := os.Stat(c.SynxConfigMachine); os.IsNotExist(err) {
		return []string{}, nil
	}
	return readFileLines(c.SynxConfigMachine)
}

func (c *ConfigManager) IsExcluded(path string) bool {
	if len(c.Excludes) == 0 {
		return false
	}

	// Basic matching logic analogous to fish string match -q
	for _, pattern := range c.Excludes {
		if path == pattern {
			return true
		}
		if match, _ := filepath.Match(pattern, path); match {
			return true
		}
		// Prefix matching (pattern/*)
		if strings.HasPrefix(path, pattern+"/") {
			return true
		}
		// Suffix matching (*/pattern) if pattern has no slashes
		if !strings.Contains(pattern, "/") && strings.HasSuffix(path, "/"+pattern) {
			return true
		}
	}
	return false
}

// Helpers

func readFileLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines, scanner.Err()
}

func writeLines(path string, lines []string, header string) error {
	content := ""
	if header != "" {
		content += "# " + strings.ReplaceAll(header, "\n", "\n# ") + "\n\n"
	}
	content += strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0644)
}
