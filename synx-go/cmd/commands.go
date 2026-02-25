package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Blumenwagen/synx/pkg/bootstrap"
	"github.com/Blumenwagen/synx/pkg/config"
	"github.com/Blumenwagen/synx/pkg/git"
	"github.com/Blumenwagen/synx/pkg/sync"
	"github.com/Blumenwagen/synx/pkg/ui"
	"github.com/spf13/cobra"
)

func init() {
	// Sync logic mapping (rootCmd Run)
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		cfg, err := config.NewConfigManager()
		if err != nil {
			ui.Error(fmt.Sprintf("Init error: %v", err))
			os.Exit(1)
		}
		cfg.Load()

		if restoreFlag {
			runRestore(cfg)
			return
		}

		if listFlag {
			runList(cfg)
			return
		}

		if historyFlag {
			runHistory(cfg)
			return
		}

		if rollbackFlag > 0 {
			runRollback(cfg, rollbackFlag)
			return
		}

		if addFlag != "" {
			runAdd(cfg, addFlag)
			return
		}

		if removeFlag != "" {
			runRemove(cfg, removeFlag)
			return
		}

		if excludeFlag != "" {
			runExclude(cfg, excludeFlag)
			return
		}

		if bsSetupFlag {
			runBsSetup(cfg)
			return
		}

		if bootstrapFlag != "" || cmd.Flags().Changed("bootstrap") {
			runBootstrap(cfg, bootstrapFlag, yesFlag)
			return
		}

		// Default sync action
		runSync(cfg)
	}
}

func runSync(cfg *config.ConfigManager) {
	title := "Dotfile Sync"
	if cfg.Hostname != "" {
		title += " (" + cfg.Hostname + ")"
	}
	ui.PrintHeader("🚀", title)
	if cfg.UsingMachineTargets {
		ui.Detail("Using machine-specific targets")
	}
	ui.Step("Syncing dotfiles...")

	eng, err := sync.NewEngine(cfg)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	res, err := eng.SyncToRepo()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Printf("%s synced  %s  %s skipped\n", ui.StyleGreen.Render(fmt.Sprintf("%d", res.Synced)), ui.StyleDim.Render("│"), ui.StyleYellow.Render(fmt.Sprintf("%d", res.Skipped)))
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Println()

	// Git
	g := git.NewGitManager(eng.DotfileDir)
	if !g.IsRepo() {
		ui.Error("Not a git repository.")
		ui.Detail("Run: cd " + eng.DotfileDir + " && git init && git remote add origin <url>")
		os.Exit(1)
	}

	if !g.HasChanges() {
		ui.Info("No changes to commit")
		os.Exit(0)
	}

	ui.Step("Committing changes...")
	cfiles, _ := g.ChangedFilesCount()
	ui.Detail(fmt.Sprintf("Modified: %d file(s)", cfiles))

	err = g.Commit("Update rice via synx")
	if err != nil {
		ui.Error("Commit failed")
		os.Exit(1)
	}
	ui.Success("Committed changes")

	fmt.Println()
	ui.Step("Pushing to remote...")
	branch, _ := g.CurrentBranch()

	if err := g.Push(branch, false); err != nil {
		ui.Error("Push failed")
		ui.Detail("Check your git remote configuration and credentials")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("%s  %s\n", ui.StyleGreen.Render("☁"), ui.StyleBold.Render("Sync complete!"))
}

func runRestore(cfg *config.ConfigManager) {
	title := "Restore Mode"
	if cfg.Hostname != "" {
		title += " (" + cfg.Hostname + ")"
	}
	ui.PrintHeader("⬇", title)

	eng, err := sync.NewEngine(cfg)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	g := git.NewGitManager(eng.DotfileDir)
	if !g.IsRepo() {
		ui.Error("Not a git repository.")
		os.Exit(1)
	}

	ui.Step("Checking for updates...")
	branch, _ := g.CurrentBranch()
	g.Fetch()

	upToDate, _ := g.IsUpToDate(branch)
	if upToDate {
		ui.Success("Already up-to-date")
	} else {
		ui.Info("Updates available")
		fmt.Println()
		ui.Step("Pulling latest from remote...")

		stashed, _ := g.Stash()
		if err := g.Pull(); err != nil {
			ui.Error("Failed to pull from remote")
			if stashed {
				g.StashPop()
			}
			os.Exit(1)
		}
		ui.Success("Pulled latest changes")
		if stashed {
			g.StashPop()
			ui.Success("Restored local uncommitted changes")
		}
	}

	fmt.Println()
	ui.Step("Restoring dotfiles to ~/.config...")

	// Restore internal config backups immediately
	synxBackup := eng.DotfileDir + "/.synx/synx.conf"
	if _, err := os.Stat(synxBackup); err == nil {
		os.MkdirAll(cfg.ConfigDir, 0755)
		copyFileSimple(synxBackup, cfg.SynxConfig)
		copyFileSimple(eng.DotfileDir+"/.synx/exclude.conf", cfg.ExcludeCfg)
		cfg.Load()
	}

	res, err := eng.RestoreFromRepo()
	if err != nil {
		ui.Error(err.Error())
	}

	fmt.Println()
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Printf("%s restored  %s  %s failed\n", ui.StyleGreen.Render(fmt.Sprintf("%d", res.Restored)), ui.StyleDim.Render("│"), ui.StyleRed.Render(fmt.Sprintf("%d", res.Failed)))
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Println()

	if res.Restored > 0 {
		for _, t := range cfg.Targets {
			if t == "hypr" {
				ui.Step("Reloading Hyprland...")
				if _, err := os.Stat("/usr/bin/hyprctl"); err == nil {
					cmd := exec.Command("hyprctl", "reload")
					if err := cmd.Run(); err == nil {
						ui.Success("Hyprland reloaded")
					}
				}
				break
			}
		}
	}

	ui.Success("Restore complete!")
}

func runList(cfg *config.ConfigManager) {
	title := "List Dotfiles"
	if cfg.Hostname != "" {
		title += " (" + cfg.Hostname + ")"
	}
	ui.PrintHeader("📋", title)
	if cfg.UsingMachineTargets {
		ui.Detail("Using machine-specific targets from synx.conf." + cfg.Hostname)
	}

	home, _ := os.UserHomeDir()
	baseConfigDir := home + "/.config"

	fmt.Println(ui.StyleBold.Render("TRACKED DOTFILES:"))
	if len(cfg.Targets) == 0 {
		ui.Detail("(none)")
	} else {
		// Build a set of machine-only targets (not in base) for annotation
		machineOnlySet := make(map[string]bool)
		if cfg.UsingMachineTargets {
			baseTargets, _ := readBaseTargets(cfg)
			baseSet := make(map[string]bool)
			for _, bt := range baseTargets {
				baseSet[bt] = true
			}
			for _, t := range cfg.Targets {
				if !baseSet[t] {
					machineOnlySet[t] = true
				}
			}
		}

		for _, t := range cfg.Targets {
			path := baseConfigDir + "/" + t
			info, err := os.Lstat(path)

			var tags []string
			if err != nil {
				tags = append(tags, "not found")
			} else if info.Mode()&os.ModeSymlink != 0 {
				tags = append(tags, "symlink")
			}
			if machineOnlySet[t] {
				tags = append(tags, cfg.Hostname+" only")
			}

			label := t
			if len(tags) > 0 {
				label += " " + ui.StyleDim.Render("("+strings.Join(tags, ", ")+")")
			}

			if err != nil {
				ui.Error(label)
			} else {
				ui.Success(label)
			}
		}
	}

	fmt.Println()
	fmt.Println(ui.StyleBold.Render("AVAILABLE DOTFILES:"))
	// Simple lookup in ~/.config
	entries, _ := os.ReadDir(baseConfigDir)
	var available []string

	targetMap := make(map[string]bool)
	for _, t := range cfg.Targets {
		targetMap[t] = true
	}

	for _, e := range entries {
		if !targetMap[e.Name()] && !strings.HasPrefix(e.Name(), ".") {
			available = append(available, e.Name())
		}
	}

	if len(available) == 0 {
		ui.Detail("(all tracked)")
	} else {
		max := 10
		if len(available) < max {
			max = len(available)
		}
		for i := 0; i < max; i++ {
			ui.Bullet(available[i])
		}
		if len(available) > 10 {
			ui.Detail(fmt.Sprintf("... and %d more", len(available)-10))
		}
	}
	fmt.Println()
}

func runHistory(cfg *config.ConfigManager) {
	ui.PrintHeader("📜", "Sync History")

	eng, _ := sync.NewEngine(cfg)
	g := git.NewGitManager(eng.DotfileDir)

	if !g.IsRepo() {
		ui.Error("Not a git repository.")
		os.Exit(1)
	}

	logs, err := g.Log(20)
	if err != nil || len(logs) == 0 {
		ui.Detail("No history yet")
		return
	}

	for i, l := range logs {
		parts := strings.SplitN(l, "|", 3)
		if len(parts) == 3 {
			fmt.Printf("%s. %s %s\n    %s\n", ui.StyleDim.Render(fmt.Sprintf("%2d", i+1)), ui.StyleCyan.Render(parts[0]), ui.StyleDim.Render(parts[1]), parts[2])
		}
	}
	fmt.Println()
}

func runRollback(cfg *config.ConfigManager, steps int) {
	ui.PrintHeader("⏪", "Rollback")

	eng, _ := sync.NewEngine(cfg)
	g := git.NewGitManager(eng.DotfileDir)

	if !g.IsRepo() {
		ui.Error("Not a git repository.")
		os.Exit(1)
	}

	fmt.Println(ui.StyleYellow.Render("⚠ WARNING"))
	fmt.Println("  This will reset your dotfiles repo to " + strconv.Itoa(steps) + " commit(s) ago.")
	fmt.Println("  Current changes will be lost, and the remote GitHub repository")
	fmt.Println("  will be force-pushed to match this rolled-back state.")
	fmt.Println()

	// Confirmation could go here (skipped for testing automation ease, or assume yes flag)

	ui.Step(fmt.Sprintf("Rolling back %d commit(s)...", steps))
	target, err := g.ResetHard(steps)
	if err != nil {
		ui.Error("Rollback failed: " + err.Error())
		os.Exit(1)
	}

	ui.Success("Rolled back to " + target)

	ui.Step("Force pushing to remote to update history...")
	branch, _ := g.CurrentBranch()
	if err := g.Push(branch, true); err != nil {
		ui.Error("Failed to force push to remote")
		ui.Detail("(You may need to push manually: git push --force origin " + branch + ")")
	} else {
		ui.Success("Remote history updated")
	}

	// Automatically run restore
	runRestore(cfg)
}

func runAdd(cfg *config.ConfigManager, target string) {
	for _, t := range cfg.Targets {
		if t == target {
			ui.Warn("'" + target + "' is already tracked")
			return
		}
	}
	cfg.Targets = append(cfg.Targets, target)
	if machineFlag {
		cfg.SaveTargetsMachine(cfg.Targets)
		ui.Success("Added '" + target + "' to tracked dotfiles (" + cfg.Hostname + ")")
	} else {
		cfg.SaveTargets(cfg.Targets)
		ui.Success("Added '" + target + "' to tracked dotfiles")
	}
}

func runRemove(cfg *config.ConfigManager, target string) {
	var newTargets []string
	found := false
	for _, t := range cfg.Targets {
		if t == target {
			found = true
		} else {
			newTargets = append(newTargets, t)
		}
	}
	if !found {
		ui.Warn("'" + target + "' is not tracked")
		return
	}
	cfg.Targets = newTargets
	if machineFlag {
		cfg.SaveTargetsMachine(cfg.Targets)
		ui.Success("Removed '" + target + "' from tracked dotfiles (" + cfg.Hostname + ")")
	} else {
		cfg.SaveTargets(cfg.Targets)
		ui.Success("Removed '" + target + "' from tracked dotfiles")
	}
}

func runExclude(cfg *config.ConfigManager, pattern string) {
	for _, p := range cfg.Excludes {
		if p == pattern {
			ui.Warn("'" + pattern + "' is already excluded")
			return
		}
	}

	if machineFlag {
		machineExcludes, _ := cfg.MachineExcludes()
		machineExcludes = append(machineExcludes, pattern)
		cfg.SaveExcludesMachine(machineExcludes)
		cfg.Excludes = append(cfg.Excludes, pattern)
		ui.Success("Added '" + pattern + "' to exclude patterns (" + cfg.Hostname + ")")
	} else {
		cfg.Excludes = append(cfg.Excludes, pattern)
		cfg.SaveExcludes(cfg.Excludes)
		ui.Success("Added '" + pattern + "' to exclude patterns")
	}

	eng, _ := sync.NewEngine(cfg)
	g := git.NewGitManager(eng.DotfileDir)
	if g.IsRepo() {
		files, _ := g.LsFiles()
		removed := 0
		for _, f := range files {
			if cfg.IsExcluded(f) {
				g.RmCached(f)
				ui.Success("Removed from repo: " + f)
				removed++
			}
		}
		if removed > 0 {
			g.Commit("Exclude: " + pattern)
			ui.Success(fmt.Sprintf("Committed exclusion (%d file(s) removed)", removed))
		}
	}
}

func runBsSetup(cfg *config.ConfigManager) {
	bsCfg, err := bootstrap.RunWizard()
	if err != nil {
		ui.Error("Wizard cancelled or failed: " + err.Error())
		return
	}

	bsPath := cfg.ConfigDir + "/bootstrap.conf"
	bootstrap.WriteConfig(bsPath, bsCfg)
	ui.Success("Saved to " + bsPath)

	eng, _ := sync.NewEngine(cfg)
	if g := git.NewGitManager(eng.DotfileDir); g.IsRepo() {
		os.MkdirAll(eng.DotfileDir+"/.synx", 0755)
		copyFileSimple(bsPath, eng.DotfileDir+"/.synx/bootstrap.conf")
		ui.Success("Copied to dotfiles repo (.synx/bootstrap.conf)")
	}
}

func runBootstrap(cfg *config.ConfigManager, url string, yes bool) {
	eng, _ := sync.NewEngine(cfg)

	if url != "" && url != "true" { // cobra flag artifact parsing
		ui.Step("Cloning dotfiles repository...")
		ui.Detail(url)

		g := git.NewGitManager(eng.DotfileDir)
		if g.IsRepo() {
			ui.Warn("Dotfiles repo already exists, pulling...")
			g.Pull()
		} else {
			if err := g.Clone(url, eng.DotfileDir); err != nil {
				ui.Error("Failed to clone")
				os.Exit(1)
			}
		}
		ui.Success("Repository ready")

		repoBsCfg := eng.DotfileDir + "/.synx/bootstrap.conf"
		if _, err := os.Stat(repoBsCfg); err == nil {
			copyFileSimple(repoBsCfg, cfg.ConfigDir+"/bootstrap.conf")
			ui.Success("Found bootstrap config in repo")
		} else {
			ui.Error("No bootstrap config found in repo")
			os.Exit(1)
		}
	}
	bsPath := cfg.ConfigDir + "/bootstrap.conf"

reviewLoop:
	for {
		bsCfg, err := bootstrap.ParseConfig(bsPath)
		if err != nil {
			ui.Error("Failed to parse bootstrap config: " + err.Error())
			os.Exit(1)
		}

		fmt.Println()
		ui.PrintHeader("📋", "Review Bootstrap")

		fmt.Println(ui.StyleBold.Render("BOOTSTRAP CONFIGURATION:"))
		fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
		fmt.Println()

		if bsCfg.AurHelper != "" {
			fmt.Printf("  %s AUR Helper:  %s\n", ui.StyleCyan.Render("▸"), ui.StyleBold.Render(bsCfg.AurHelper))
		}

		if len(bsCfg.Packages) > 0 {
			fmt.Printf("  %s Packages:    %s packages\n", ui.StyleCyan.Render("▸"), ui.StyleBold.Render(fmt.Sprintf("%d", len(bsCfg.Packages))))
		}

		if len(bsCfg.Repos) > 0 {
			fmt.Printf("  %s Repositories: %s repos\n", ui.StyleCyan.Render("▸"), ui.StyleBold.Render(fmt.Sprintf("%d", len(bsCfg.Repos))))
		}

		if bsCfg.DMName != "" {
			fmt.Printf("  %s DM:          %s\n", ui.StyleCyan.Render("▸"), ui.StyleBold.Render(bsCfg.DMName))
		}

		if bsCfg.DotfilesRestore {
			fmt.Printf("  %s Dotfiles:    %s\n", ui.StyleCyan.Render("▸"), ui.StyleGreen.Render("restore after setup"))
		}

		if len(bsCfg.Commands) > 0 {
			fmt.Printf("  %s Commands:    %s custom commands\n", ui.StyleCyan.Render("▸"), ui.StyleBold.Render(fmt.Sprintf("%d", len(bsCfg.Commands))))
		}

		fmt.Println()
		fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
		fmt.Println()

		if yes {
			break reviewLoop
		}

		fmt.Println(ui.StyleBold.Render("OPTIONS:"))
		fmt.Printf("  %s  Edit config in $EDITOR\n", ui.StyleGreen.Render("e"))
		fmt.Printf("  %s  Continue with this config\n", ui.StyleGreen.Render("c"))
		fmt.Printf("  %s  Quit\n", ui.StyleGreen.Render("q"))
		fmt.Println()

		fmt.Print("  " + ui.StyleCyan.Render("▸") + " Choose an option [e/c/q]: ")
		var choice string
		fmt.Scanln(&choice)
		choice = strings.ToLower(strings.TrimSpace(choice))

		switch choice {
		case "e":
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "nano"
			}
			cmd := exec.Command(editor, bsPath)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		case "c":
			break reviewLoop
		case "q":
			ui.Detail("Cancelled")
			os.Exit(0)
		default:
			ui.Detail("(press e, c, or q)")
		}
	}

	bsCfg, _ := bootstrap.ParseConfig(bsPath)
	runner := bootstrap.NewRunner(bsCfg, yes)
	runner.RunAll(func() error {
		runRestore(cfg)
		return nil
	})
}

// OS simple copy helper
func copyFileSimple(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// readBaseTargets reads the base synx.conf directly, bypassing machine overrides.
func readBaseTargets(cfg *config.ConfigManager) ([]string, error) {
	data, err := os.ReadFile(cfg.SynxConfig)
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	return lines, nil
}
