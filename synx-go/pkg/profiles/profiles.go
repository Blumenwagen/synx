package profiles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ListProfiles returns the names of all profile directories.
func ListProfiles(profilesDir string) ([]string, error) {
	entries, err := os.ReadDir(profilesDir)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// GetActiveProfile reads the active profile name from the marker file.
func GetActiveProfile(configDir string) string {
	data, err := os.ReadFile(filepath.Join(configDir, "active_profile"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// SetActiveProfile writes the active profile name to the marker file.
func SetActiveProfile(configDir, name string) error {
	return os.WriteFile(filepath.Join(configDir, "active_profile"), []byte(name+"\n"), 0644)
}

// ClearActiveProfile removes the active profile marker.
func ClearActiveProfile(configDir string) error {
	path := filepath.Join(configDir, "active_profile")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

// CreateProfile creates a new empty profile directory.
func CreateProfile(profilesDir, name string) error {
	dir := filepath.Join(profilesDir, name)
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("profile '%s' already exists", name)
	}
	return os.MkdirAll(dir, 0755)
}

// DeleteProfile removes a profile directory entirely.
func DeleteProfile(profilesDir, name string) error {
	dir := filepath.Join(profilesDir, name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", name)
	}
	return os.RemoveAll(dir)
}

// ProfileTargetCount returns the number of explicit targets tracked by the profile.
func ProfileTargetCount(profilesDir, name string) int {
	data, err := os.ReadFile(filepath.Join(profilesDir, name, "targets.conf"))
	if err != nil {
		return 0
	}
	count := 0
	lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" && !strings.HasPrefix(l, "#") {
			count++
		}
	}
	return count
}
