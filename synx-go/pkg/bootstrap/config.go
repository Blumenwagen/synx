package bootstrap

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	AurHelper       string
	Packages        []string
	Repos           []RepoInfo
	DMName          string
	DMTheme         string
	DMThemeSource   string
	DotfilesRestore bool
	Commands        []string
}

type RepoInfo struct {
	URL     string
	Dest    string
	Command string
}

func ParseConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &Config{}
	var currentSection string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.TrimSuffix(strings.TrimPrefix(line, "["), "]")
			continue
		}

		switch currentSection {
		case "aur":
			if strings.HasPrefix(line, "helper = ") {
				cfg.AurHelper = strings.TrimPrefix(line, "helper = ")
			}
		case "packages":
			if strings.HasPrefix(line, "list = ") {
				pkgs := strings.Split(strings.TrimPrefix(line, "list = "), " ")
				for _, p := range pkgs {
					if p := strings.TrimSpace(p); p != "" {
						cfg.Packages = append(cfg.Packages, p)
					}
				}
			} else {
				pkgs := strings.Split(line, " ")
				for _, p := range pkgs {
					if p := strings.TrimSpace(p); p != "" {
						cfg.Packages = append(cfg.Packages, p)
					}
				}
			}
		case "repos":
			if strings.HasPrefix(line, "repo = ") {
				val := strings.TrimPrefix(line, "repo = ")
				parts := strings.Split(val, "|")

				repo := RepoInfo{}
				if len(parts) > 0 {
					repo.URL = strings.TrimSpace(parts[0])
				}
				if len(parts) > 1 {
					repo.Dest = strings.TrimSpace(parts[1])
				}
				if len(parts) > 2 {
					repo.Command = strings.TrimSpace(parts[2])
				}
				cfg.Repos = append(cfg.Repos, repo)
			}
		case "dm":
			if strings.HasPrefix(line, "name = ") {
				cfg.DMName = strings.TrimPrefix(line, "name = ")
			}
			if strings.HasPrefix(line, "theme = ") {
				cfg.DMTheme = strings.TrimPrefix(line, "theme = ")
			}
			if strings.HasPrefix(line, "theme_source = ") {
				cfg.DMThemeSource = strings.TrimPrefix(line, "theme_source = ")
			}
		case "dotfiles":
			if line == "restore = true" {
				cfg.DotfilesRestore = true
			}
		case "commands":
			if strings.HasPrefix(line, "run = ") {
				cfg.Commands = append(cfg.Commands, strings.TrimPrefix(line, "run = "))
			}
		}
	}

	return cfg, scanner.Err()
}

func WriteConfig(path string, cfg *Config) error {
	var lines []string

	lines = append(lines, "# ╔═══════════════════════════════════════════╗")
	lines = append(lines, "# ║  SYNX Bootstrap Configuration             ║")
	lines = append(lines, "# ╚═══════════════════════════════════════════╝")
	lines = append(lines, "")

	if cfg.AurHelper != "" {
		lines = append(lines, "# AUR Helper")
		lines = append(lines, "[aur]")
		lines = append(lines, "helper = "+cfg.AurHelper)
		lines = append(lines, "")
	}

	if len(cfg.Packages) > 0 {
		lines = append(lines, "# Packages")
		lines = append(lines, "[packages]")
		lines = append(lines, "list = "+strings.Join(cfg.Packages, " "))
		lines = append(lines, "")
	}

	if len(cfg.Repos) > 0 {
		lines = append(lines, "# Git Repositories")
		lines = append(lines, "[repos]")
		for _, r := range cfg.Repos {
			cmdStr := ""
			if r.Command != "" {
				cmdStr = " | " + r.Command
			}
			destStr := ""
			if r.Dest != "" {
				destStr = " | " + r.Dest
			}
			lines = append(lines, "repo = "+r.URL+destStr+cmdStr)
		}
		lines = append(lines, "")
	}

	if cfg.DMName != "" {
		lines = append(lines, "# Display Manager")
		lines = append(lines, "[dm]")
		lines = append(lines, "name = "+cfg.DMName)
		if cfg.DMTheme != "" {
			lines = append(lines, "theme = "+cfg.DMTheme)
		}
		if cfg.DMThemeSource != "" {
			lines = append(lines, "theme_source = "+cfg.DMThemeSource)
		}
		lines = append(lines, "")
	}

	if cfg.DotfilesRestore {
		lines = append(lines, "# Dotfiles")
		lines = append(lines, "[dotfiles]")
		lines = append(lines, "restore = true")
		lines = append(lines, "")
	}

	if len(cfg.Commands) > 0 {
		lines = append(lines, "# Custom Commands")
		lines = append(lines, "[commands]")
		for _, cmd := range cfg.Commands {
			lines = append(lines, "run = "+cmd)
		}
		lines = append(lines, "")
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}
