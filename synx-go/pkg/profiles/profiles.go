package profiles

import (
	"fmt"
	"io"
	"io/fs"
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

// ApplyProfile copies all overlay files from the profile directory into baseConfigDir (~/.config).
// It preserves directory structure. Returns the number of files copied.
func ApplyProfile(profilesDir, name, baseConfigDir string) (int, error) {
	profileDir := filepath.Join(profilesDir, name)
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		return 0, fmt.Errorf("profile '%s' does not exist", name)
	}

	count := 0
	err := filepath.WalkDir(profileDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Skip profile metadata files
		base := d.Name()
		if base == "targets.conf" || base == "excludes.conf" {
			return nil
		}

		// Calculate relative path within the profile
		relPath, _ := filepath.Rel(profileDir, path)
		destPath := filepath.Join(baseConfigDir, relPath)

		// Ensure destination directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		if err := copyFile(path, destPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", relPath, err)
		}
		count++
		return nil
	})

	return count, err
}

// ProfileFiles returns a list of files in a profile (relative paths).
func ProfileFiles(profilesDir, name string) ([]string, error) {
	profileDir := filepath.Join(profilesDir, name)
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("profile '%s' does not exist", name)
	}

	var files []string
	filepath.WalkDir(profileDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		relPath, _ := filepath.Rel(profileDir, path)
		files = append(files, relPath)
		return nil
	})
	return files, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
