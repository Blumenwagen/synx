package packages

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// HasPacman checks if pacman is available on the system.
func HasPacman() bool {
	_, err := exec.LookPath("pacman")
	return err == nil
}

// CaptureNative returns explicitly installed native (repo) packages.
func CaptureNative() ([]string, error) {
	return runPacman("-Qqen")
}

// CaptureForeign returns explicitly installed foreign (AUR) packages.
func CaptureForeign() ([]string, error) {
	return runPacman("-Qqem")
}

// DetectAURHelper returns the first AUR helper found on $PATH.
func DetectAURHelper() string {
	helpers := []string{"paru", "yay", "trizen", "pikaur"}
	for _, h := range helpers {
		if _, err := exec.LookPath(h); err == nil {
			return h
		}
	}
	return ""
}

// Diff compares two sorted package lists and returns added/removed.
func Diff(saved, current []string) (added, removed []string) {
	savedSet := make(map[string]bool, len(saved))
	currentSet := make(map[string]bool, len(current))

	for _, p := range saved {
		savedSet[p] = true
	}
	for _, p := range current {
		currentSet[p] = true
	}

	for _, p := range current {
		if !savedSet[p] {
			added = append(added, p)
		}
	}
	for _, p := range saved {
		if !currentSet[p] {
			removed = append(removed, p)
		}
	}
	return
}

// LoadList reads a package list file (one per line, skip comments and blanks).
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

// SaveList writes a package list to a file with a header comment.
func SaveList(path, header string, pkgs []string) error {
	sort.Strings(pkgs)

	content := ""
	if header != "" {
		content += "# " + header + "\n\n"
	}
	content += strings.Join(pkgs, "\n")
	if len(pkgs) > 0 {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// NativeListPath returns the path to the native package list.
func NativeListPath(configDir string) string {
	return configDir + "/packages.native"
}

// ForeignListPath returns the path to the foreign package list.
func ForeignListPath(configDir string) string {
	return configDir + "/packages.foreign"
}

// InstallNative installs native packages via pacman.
func InstallNative(pkgs []string) error {
	if len(pkgs) == 0 {
		return nil
	}
	args := append([]string{"-S", "--needed", "--noconfirm"}, pkgs...)
	cmd := exec.Command("sudo", append([]string{"pacman"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// InstallForeign installs foreign packages via the detected AUR helper.
func InstallForeign(helper string, pkgs []string) error {
	if len(pkgs) == 0 || helper == "" {
		return nil
	}
	args := append([]string{"-S", "--needed", "--noconfirm"}, pkgs...)
	cmd := exec.Command(helper, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func runPacman(flag string) ([]string, error) {
	cmd := exec.Command("pacman", flag)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pacman %s failed: %w", flag, err)
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return []string{}, nil
	}
	lines := strings.Split(raw, "\n")
	return lines, nil
}
