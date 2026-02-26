package bootstrap

import (
	"strings"

	"github.com/Blumenwagen/synx/pkg/ui"
)

func RunWizard() (*Config, error) {
	result, err := ui.RunBootstrapTUI()
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil // user canceled
	}

	cfg := &Config{
		AurHelper:       result.AurHelper,
		DMName:          result.DMName,
		DMTheme:         result.DMTheme,
		DMThemeSource:   result.DMThemeSource,
		DotfilesRestore: result.DotfilesRestore,
	}

	// Parse packages (space/newline separated)
	if result.Packages != "" {
		for _, p := range strings.Fields(result.Packages) {
			if p = strings.TrimSpace(p); p != "" {
				cfg.Packages = append(cfg.Packages, p)
			}
		}
	}

	// Parse repos (pipe-separated per line)
	if result.Repos != "" {
		for _, line := range strings.Split(result.Repos, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.Split(line, "|")
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
			if repo.URL != "" {
				cfg.Repos = append(cfg.Repos, repo)
			}
		}
	}

	// Parse commands (one per line)
	if result.Commands != "" {
		for _, c := range strings.Split(result.Commands, "\n") {
			if c = strings.TrimSpace(c); c != "" {
				cfg.Commands = append(cfg.Commands, c)
			}
		}
	}

	return cfg, nil
}
