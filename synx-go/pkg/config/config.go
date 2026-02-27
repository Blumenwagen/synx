package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ConfigManager struct {
	ConfigDir       string
	SynxConfig      string
	ExcludeCfg      string
	Hostname        string
	ActiveProfile   string
	Targets         []string
	Excludes        []string
	ProfileTargets  []string
	ProfileExcludes []string
}

func NewConfigManager() (*ConfigManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	cfgDir := filepath.Join(home, ".config", "synx")

	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config dir: %w", err)
	}

	hostname, _ := os.Hostname()

	return &ConfigManager{
		ConfigDir:       cfgDir,
		SynxConfig:      filepath.Join(cfgDir, "synx.conf"),
		ExcludeCfg:      filepath.Join(cfgDir, "exclude.conf"),
		Hostname:        hostname,
		Targets:         []string{},
		Excludes:        []string{},
		ProfileTargets:  []string{},
		ProfileExcludes: []string{},
	}, nil
}

func (c *ConfigManager) Load() error {
	if _, err := os.Stat(c.SynxConfig); os.IsNotExist(err) {
		defaultTargets := []string{"hypr", "foot", "kitty", "fastfetch", "alacritty"}
		err = c.SaveTargets(defaultTargets)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(c.ExcludeCfg); os.IsNotExist(err) {
		err = c.SaveExcludes([]string{})
		if err != nil {
			return err
		}
	}

	targets, err := readFileLines(c.SynxConfig)
	if err != nil {
		return err
	}
	c.Targets = targets

	excludes, err := readFileLines(c.ExcludeCfg)
	if err != nil {
		return err
	}
	c.Excludes = excludes

	activeProfilePath := filepath.Join(c.ConfigDir, "active_profile")
	if data, err := os.ReadFile(activeProfilePath); err == nil {
		c.ActiveProfile = strings.TrimSpace(string(data))
	}

	if c.ActiveProfile != "" {
		profileDir := filepath.Join(c.ConfigDir, "profiles", c.ActiveProfile)
		if profTargets, err := readFileLines(filepath.Join(profileDir, "targets.conf")); err == nil {
			c.ProfileTargets = profTargets
		}
		if profExcludes, err := readFileLines(filepath.Join(profileDir, "excludes.conf")); err == nil {
			c.ProfileExcludes = profExcludes
		}
	}

	return nil
}

func (c *ConfigManager) SaveTargets(targets []string) error {
	return writeLines(c.SynxConfig, targets, "Synx tracked dotfiles")
}

func (c *ConfigManager) SaveExcludes(excludes []string) error {
	return writeLines(c.ExcludeCfg, excludes, "Exclude patterns for machine-specific files\n# One pattern per line")
}

func (c *ConfigManager) SaveProfileTargets(profile string, targets []string) error {
	path := filepath.Join(c.ConfigDir, "profiles", profile, "targets.conf")
	os.MkdirAll(filepath.Dir(path), 0755)
	return writeLines(path, targets, "Synx tracked dotfiles for profile "+profile)
}

func (c *ConfigManager) SaveProfileExcludes(profile string, excludes []string) error {
	path := filepath.Join(c.ConfigDir, "profiles", profile, "excludes.conf")
	os.MkdirAll(filepath.Dir(path), 0755)
	return writeLines(path, excludes, "Exclude patterns for profile "+profile)
}

// IsProfileTarget checks if a target is explicitly tracked by the active profile.
func (c *ConfigManager) IsProfileTarget(target string) bool {
	if c.ActiveProfile == "" {
		return false
	}
	for _, pt := range c.ProfileTargets {
		if pt == target {
			return true
		}
	}
	return false
}

// GetAllTargets returns a deduplicated list of base targets and active profile targets.
func (c *ConfigManager) GetAllTargets() []string {
	m := make(map[string]bool)
	for _, t := range c.Targets {
		m[t] = true
	}
	if c.ActiveProfile != "" {
		for _, pt := range c.ProfileTargets {
			m[pt] = true
		}
	}
	var res []string
	for t := range m {
		res = append(res, t)
	}
	return res
}

// GetGlobalTargets scans all synx.conf* files in ~/.config/synx
// and returns a deduplicated list of all tracked dotfiles across all machines.
func (c *ConfigManager) GetGlobalTargets() ([]string, error) {
	globalSet := make(map[string]bool)

	entries, err := os.ReadDir(c.ConfigDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read config dir: %w", err)
	}

	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "synx.conf") && !e.IsDir() {
			path := filepath.Join(c.ConfigDir, name)
			lines, err := readFileLines(path)
			if err != nil {
				continue
			}
			for _, line := range lines {
				globalSet[line] = true
			}
		}
	}

	var allTargets []string
	for t := range globalSet {
		allTargets = append(allTargets, t)
	}
	return allTargets, nil
}

func (c *ConfigManager) IsExcluded(path string) bool {
	allExcludes := append([]string{}, c.Excludes...)
	if c.ActiveProfile != "" {
		allExcludes = append(allExcludes, c.ProfileExcludes...)
	}

	if len(allExcludes) == 0 {
		return false
	}

	for _, pattern := range allExcludes {
		if path == pattern {
			return true
		}
		if match, _ := filepath.Match(pattern, path); match {
			return true
		}
		if strings.HasPrefix(path, pattern+"/") {
			return true
		}
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
