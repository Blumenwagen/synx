package services

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// HasSystemctl checks if systemctl is available on the system.
func HasSystemctl() bool {
	_, err := exec.LookPath("systemctl")
	return err == nil
}

// CaptureSystem returns enabled system-level service units.
func CaptureSystem() ([]string, error) {
	return captureEnabled(false)
}

// CaptureUser returns enabled user-level service units.
func CaptureUser() ([]string, error) {
	return captureEnabled(true)
}

// Diff compares two sorted service lists and returns added/removed.
func Diff(saved, current []string) (added, removed []string) {
	savedSet := make(map[string]bool, len(saved))
	currentSet := make(map[string]bool, len(current))

	for _, s := range saved {
		savedSet[s] = true
	}
	for _, s := range current {
		currentSet[s] = true
	}

	for _, s := range current {
		if !savedSet[s] {
			added = append(added, s)
		}
	}
	for _, s := range saved {
		if !currentSet[s] {
			removed = append(removed, s)
		}
	}
	return
}

// LoadList reads a service list file.
func LoadList(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
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

// SaveList writes a service list to a file with a header comment.
func SaveList(path, header string, svcs []string) error {
	sort.Strings(svcs)

	content := ""
	if header != "" {
		content += "# " + header + "\n\n"
	}
	content += strings.Join(svcs, "\n")
	if len(svcs) > 0 {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// SystemListPath returns the path to the system service list.
func SystemListPath(configDir string) string {
	return configDir + "/services.system"
}

// UserListPath returns the path to the user service list.
func UserListPath(configDir string) string {
	return configDir + "/services.user"
}

// EnableSystem enables system-level services.
func EnableSystem(svcs []string) error {
	if len(svcs) == 0 {
		return nil
	}
	args := append([]string{"systemctl", "enable"}, svcs...)
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// EnableUser enables user-level services.
func EnableUser(svcs []string) error {
	if len(svcs) == 0 {
		return nil
	}
	args := append([]string{"--user", "enable"}, svcs...)
	cmd := exec.Command("systemctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func captureEnabled(user bool) ([]string, error) {
	args := []string{"list-unit-files", "--state=enabled", "--no-legend", "--no-pager"}
	if user {
		args = append([]string{"--user"}, args...)
	}

	cmd := exec.Command("systemctl", args...)
	out, err := cmd.Output()
	if err != nil {
		kind := "system"
		if user {
			kind = "user"
		}
		return nil, fmt.Errorf("systemctl (%s) failed: %w", kind, err)
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return []string{}, nil
	}

	var units []string
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 1 {
			units = append(units, fields[0])
		}
	}
	return units, nil
}
