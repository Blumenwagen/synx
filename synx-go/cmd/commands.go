package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Blumenwagen/synx/pkg/bootstrap"
	"github.com/Blumenwagen/synx/pkg/config"
	"github.com/Blumenwagen/synx/pkg/git"
	"github.com/Blumenwagen/synx/pkg/hooks"
	"github.com/Blumenwagen/synx/pkg/packages"
	"github.com/Blumenwagen/synx/pkg/profiles"
	"github.com/Blumenwagen/synx/pkg/services"
	"github.com/Blumenwagen/synx/pkg/sync"
	"github.com/Blumenwagen/synx/pkg/ui"
	"github.com/spf13/cobra"
)

func runInit(cfg *config.ConfigManager, dotfileDir string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	ui.PrintHeader("📦", "SYNX - Setup")
	fmt.Println()

	ui.Step("Setting up configuration...")
	fmt.Println()

	if _, err := os.Stat(cfg.SynxConfig); err == nil {
		ui.Success("Created " + cfg.SynxConfig)
	} else {
		ui.Warn("Could not find or create " + cfg.SynxConfig)
	}

	if _, err := os.Stat(cfg.ExcludeCfg); err == nil {
		ui.Success("Created " + cfg.ExcludeCfg)
	} else {
		ui.Warn("Could not find or create " + cfg.ExcludeCfg)
	}

	fmt.Println()
	ui.Step("Dotfiles repository setup")
	fmt.Println()

	g := git.NewGitManager(dotfileDir)
	if g.IsRepo() {
		ui.Detail("○ Repository already exists at " + dotfileDir + ", skipping")
	} else {
		fmt.Println(ui.StyleBold.Render("Choose an option:"))
		fmt.Printf("    %s Initialize a new repo in %s\n", ui.StyleCyan.Render("1)"), dotfileDir)
		fmt.Printf("    %s Clone an existing repo to %s\n", ui.StyleCyan.Render("2)"), dotfileDir)
		fmt.Printf("    %s Skip\n", ui.StyleCyan.Render("3)"))
		fmt.Println()
		fmt.Print("  Enter choice [1/2/3]: ")

		var choice string
		fmt.Scanln(&choice)
		choice = strings.TrimSpace(choice)
		fmt.Println()

		switch choice {
		case "1":
			os.MkdirAll(dotfileDir, 0755)

			// Initialize repo manually to bypass IsRepo check restrictions
			cmd := exec.Command("git", "init")
			cmd.Dir = dotfileDir
			if err := cmd.Run(); err != nil {
				ui.Error("Failed to initialize git repo in " + dotfileDir)
			} else {
				ui.Success("Initialized git repo in " + dotfileDir)
				fmt.Println()
				fmt.Print("  Add remote? [y/N] ")

				var addRemote string
				fmt.Scanln(&addRemote)
				addRemote = strings.ToLower(strings.TrimSpace(addRemote))
				fmt.Println()

				if addRemote == "y" || addRemote == "yes" {
					fmt.Print("  Remote URL: ")
					remoteUrl, _ := reader.ReadString('\n')
					remoteUrl = strings.TrimSpace(remoteUrl)

					if remoteUrl != "" {
						err := g.AddRemote("origin", remoteUrl)
						if err != nil {
							ui.Error("Failed to add remote: " + err.Error())
						} else {
							ui.Success("Added remote: " + remoteUrl)
						}
					}
				}
			}
		case "2":
			fmt.Print("  Remote URL to clone: ")
			cloneUrl, _ := reader.ReadString('\n')
			cloneUrl = strings.TrimSpace(cloneUrl)

			if cloneUrl != "" {
				err := g.Clone(cloneUrl, dotfileDir)
				if err == nil {
					ui.Success("Cloned to " + dotfileDir)
				} else {
					ui.Error("Failed to clone repository")
				}
			} else {
				ui.Warn("No URL provided, skipping")
			}
		default:
			ui.Detail("○ Skipped dotfiles repo setup")
		}
	}

	fmt.Println()
	ui.Success("Setup complete!")
	fmt.Println()
	fmt.Println(ui.StyleBold.Render("Quick Start:"))
	fmt.Printf("  %s              Show all commands\n", ui.StyleCyan.Render("synx --help"))
	fmt.Printf("  %s              List tracked dotfiles\n", ui.StyleCyan.Render("synx --list"))
	fmt.Printf("  %s                     Sync to GitHub\n", ui.StyleCyan.Render("synx"))
	fmt.Printf("  %s           Restore from GitHub\n", ui.StyleCyan.Render("synx --restore"))
	fmt.Printf("  %s   Create bootstrap config\n", ui.StyleCyan.Render("synx --bootstrap-setup"))
	fmt.Println()
	fmt.Println(ui.StyleBold.Render("Configuration:"))
	fmt.Printf("  Tracked files: %s\n", ui.StyleCyan.Render("~/.config/synx/synx.conf"))
	fmt.Printf("  Excludes:      %s\n", ui.StyleCyan.Render("~/.config/synx/exclude.conf"))
	fmt.Printf("  Bootstrap:     %s\n", ui.StyleCyan.Render("~/.config/synx/bootstrap.conf"))
	fmt.Println()
}

func init() {
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Println("synx version " + Version)
			return
		}

		cfg, err := config.NewConfigManager()
		if err != nil {
			ui.Error(fmt.Sprintf("Init error: %v", err))
			os.Exit(1)
		}
		cfg.Load()

		eng, err := sync.NewEngine(cfg)
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}

		g := git.NewGitManager(eng.DotfileDir)

		if initFlag || !g.IsRepo() {
			runInit(cfg, eng.DotfileDir)
			if initFlag {
				return
			}
		}

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

		if rollbackFlag >= 0 {
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

		if cleanFlag {
			runClean(cfg)
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

		if statusFlag {
			runStatus(cfg)
			return
		}

		if doctorFlag {
			runDoctor(cfg)
			return
		}

		if profileFlag != "" {
			runProfileApply(cfg, profileFlag)
			return
		}

		if profileListFlag {
			runProfileList(cfg)
			return
		}

		if profileCreateFlag != "" {
			runProfileCreate(cfg, profileCreateFlag)
			return
		}

		if profileDeleteFlag != "" {
			runProfileDelete(cfg, profileDeleteFlag)
			return
		}

		if remoteDiffFlag {
			runRemoteDiff(cfg)
			return
		}

		if updateFlag {
			runUpdate()
			return
		}

		runSync(cfg)
	}
}

func runSync(cfg *config.ConfigManager) {
	title := "Dotfile Sync"
	if dryRunFlag {
		title = "Dotfile Sync (DRY RUN)"
	}
	if cfg.Hostname != "" {
		title += " (" + cfg.Hostname + ")"
	}
	ui.PrintHeader("🚀", title)
	if cfg.ActiveProfile != "" {
		ui.Detail("Using profile: " + cfg.ActiveProfile)
	}
	if dryRunFlag {
		ui.Info("Dry-run mode — no files will be modified")
		fmt.Println()
	}

	hooksDir := filepath.Join(cfg.ConfigDir, "hooks")
	if err := hooks.RunHook(hooksDir, hooks.PreSync); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Step("Syncing dotfiles...")

	eng, err := sync.NewEngine(cfg)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
	eng.DryRun = dryRunFlag

	res, err := eng.SyncToRepo()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	if dryRunFlag {
		fmt.Printf("%s would sync  %s  %s would skip\n", ui.StyleGreen.Render(fmt.Sprintf("%d", res.Synced)), ui.StyleDim.Render("│"), ui.StyleYellow.Render(fmt.Sprintf("%d", res.Skipped)))
	} else {
		fmt.Printf("%s synced  %s  %s skipped\n", ui.StyleGreen.Render(fmt.Sprintf("%d", res.Synced)), ui.StyleDim.Render("│"), ui.StyleYellow.Render(fmt.Sprintf("%d", res.Skipped)))
	}
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Println()

	if dryRunFlag {
		ui.Info("Dry-run complete — no changes were made")
		return
	}

	g := git.NewGitManager(eng.DotfileDir)
	if !g.IsRepo() {
		ui.Error("Not a git repository.")
		ui.Detail("Run: cd " + eng.DotfileDir + " && git init && git remote add origin <url>")
		os.Exit(1)
	}

	if !g.HasChanges() {
		ui.Info("No changes to commit")
		return
	}

	ui.Step("Committing changes...")
	cfiles, _ := g.ChangedFilesCount()
	ui.Detail(fmt.Sprintf("Modified: %d file(s)", cfiles))

	commitMsg := fmt.Sprintf("Update rice via synx - %s", time.Now().Format("2006-01-02 15:04:05"))
	err = g.Commit(commitMsg)
	if err != nil {
		ui.Error("Commit failed")
		os.Exit(1)
	}
	ui.Success("Committed changes")

	fmt.Println()
	ui.Step("Pushing to remote...")
	branch, _ := g.CurrentBranch()

	hasRemote, err := g.RemoteURL()
	if err != nil || hasRemote == "" {
		ui.Warn("No remote configured. Skipping push.")
		ui.Detail("Run `git remote add origin <url>` in your dotfiles repo to enable remote sync.")
	} else if err := g.Push(branch, false); err != nil {
		ui.Error("Push failed")
		ui.Detail("Check your git remote configuration and credentials")
		os.Exit(1)
	}

	hooks.RunHook(hooksDir, hooks.PostSync)

	fmt.Println()
	fmt.Printf("%s  %s\n", ui.StyleGreen.Render("☁"), ui.StyleBold.Render("Sync complete!"))
}

func runRestore(cfg *config.ConfigManager) {
	title := "Restore Mode"
	if dryRunFlag {
		title = "Restore Mode (DRY RUN)"
	}
	if cfg.Hostname != "" {
		title += " (" + cfg.Hostname + ")"
	}
	ui.PrintHeader("⬇", title)

	hooksDir := filepath.Join(cfg.ConfigDir, "hooks")
	if err := hooks.RunHook(hooksDir, hooks.PreRestore); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	eng, err := sync.NewEngine(cfg)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
	eng.DryRun = dryRunFlag

	g := git.NewGitManager(eng.DotfileDir)
	if !g.IsRepo() {
		ui.Error("Not a git repository.")
		os.Exit(1)
	}

	if !dryRunFlag {
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
	}

	fmt.Println()
	ui.Step("Restoring dotfiles to ~/.config...")

	if !dryRunFlag {
		synxBackup := eng.DotfileDir + "/.synx/synx.conf"
		if _, err := os.Stat(synxBackup); err == nil {
			os.MkdirAll(cfg.ConfigDir, 0755)
			copyFileSimple(synxBackup, cfg.SynxConfig)
			copyFileSimple(eng.DotfileDir+"/.synx/exclude.conf", cfg.ExcludeCfg)
			cfg.Load()
		}

		synxDataDir := eng.DotfileDir + "/.synx"
		for _, name := range []string{"packages.native", "packages.foreign", "services.system", "services.user"} {
			src := filepath.Join(synxDataDir, name)

			if cfg.IsProfileTarget(name) {
				profSrc := filepath.Join(eng.DotfileDir, "profiles", cfg.ActiveProfile, ".synx", name)
				if _, err := os.Stat(profSrc); err == nil {
					src = profSrc
				}
			}

			if _, err := os.Stat(src); err == nil {
				copyFileSimple(src, filepath.Join(cfg.ConfigDir, name))
			}
		}
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

	if dryRunFlag {
		ui.Info("Dry-run complete — no changes were made")
		return
	}

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

	hooks.RunHook(hooksDir, hooks.PostRestore)

	ui.Success("Restore complete!")
}

func runList(cfg *config.ConfigManager) {
	targetToAdd, err := ui.RunListTUI(cfg)
	if err != nil {
		ui.Error("List UI error: " + err.Error())
		os.Exit(1)
	}

	if targetToAdd != "" {
		fmt.Println()
		runAdd(cfg, targetToAdd)
	}
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
	eng, _ := sync.NewEngine(cfg)
	g := git.NewGitManager(eng.DotfileDir)

	if !g.IsRepo() {
		ui.Error("Not a git repository.")
		os.Exit(1)
	}

	if steps <= 0 {
		var err error
		steps, err = ui.RunRollbackTUI(g)
		if err != nil {
			ui.Error("Rollback UI error: " + err.Error())
			os.Exit(1)
		}

		if steps == 0 {
			fmt.Println("Rollback canceled.")
			return
		}
	}

	title := "Rollback"
	if dryRunFlag {
		title = "Rollback (DRY RUN)"
	}
	ui.PrintHeader("⏪", title)

	if dryRunFlag {
		ui.Info(fmt.Sprintf("Would roll back %d commit(s)", steps))
		ui.Info("Would force-push to remote and run dotfile restore")
		return
	}

	fmt.Println(ui.StyleYellow.Render("⚠ WARNING"))
	fmt.Println("  This will reset your dotfiles repo to " + strconv.Itoa(steps) + " commit(s) ago.")
	fmt.Println("  Current changes will be lost, and the remote GitHub repository")
	fmt.Println("  will be force-pushed to match this rolled-back state.")
	fmt.Println()

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

	runRestore(cfg)
}

func runAdd(cfg *config.ConfigManager, target string) {
	if cfg.ActiveProfile != "" {
		ui.Info("Profile '" + cfg.ActiveProfile + "' is active.")
		fmt.Printf("  %s Add '%s' as an explicit override for this profile? [Y/n]: ", ui.StyleCyan.Render("▸"), target)

		var choice string
		fmt.Scanln(&choice)
		choice = strings.ToLower(strings.TrimSpace(choice))

		if choice == "" || choice == "y" || choice == "yes" {
			if cfg.IsProfileTarget(target) {
				ui.Warn("'" + target + "' is already in profile overrides")
				return
			}
			cfg.ProfileTargets = append(cfg.ProfileTargets, target)
			cfg.SaveProfileTargets(cfg.ActiveProfile, cfg.ProfileTargets)
			ui.Success("Added '" + target + "' to profile '" + cfg.ActiveProfile + "'")
			return
		} else {
			ui.Detail("Skipped adding to profile. Adding to base config instead.")
		}
	}

	alreadyInBase := false
	for _, bt := range cfg.Targets {
		if bt == target {
			alreadyInBase = true
			break
		}
	}
	if !alreadyInBase {
		cfg.Targets = append(cfg.Targets, target)
		cfg.SaveTargets(cfg.Targets)
		ui.Success("Added '" + target + "' to base tracked dotfiles")
	} else {
		ui.Warn("'" + target + "' is already in base tracked dotfiles")
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
	cfg.SaveTargets(cfg.Targets)
	ui.Success("Removed '" + target + "' from tracked dotfiles")
}

func runClean(cfg *config.ConfigManager) {
	ui.PrintHeader("🧹", "Clean Repository")

	globalTargets, err := cfg.GetGlobalTargets()
	if err != nil {
		ui.Error("Failed to read global targets: " + err.Error())
		os.Exit(1)
	}

	eng, _ := sync.NewEngine(cfg)
	g := git.NewGitManager(eng.DotfileDir)

	if !g.IsRepo() {
		ui.Error("Not a git repository.")
		os.Exit(1)
	}

	safeSet := map[string]bool{
		".git":       true,
		".synx":      true,
		".github":    true,
		"README.md":  true,
		"LICENSE":    true,
		"install.sh": true,
		"profiles":   true,
	}

	for _, t := range globalTargets {
		safeSet[t] = true
	}

	entries, err := os.ReadDir(eng.DotfileDir)
	if err != nil {
		ui.Error("Failed to read dotfiles repository: " + err.Error())
		os.Exit(1)
	}

	var orphans []string
	for _, e := range entries {
		name := e.Name()
		if !safeSet[name] {
			orphans = append(orphans, name)
		}
	}

	if len(orphans) == 0 {
		ui.Success("Repository is clean. No orphaned dotfiles found.")
		return
	}

	ui.Warn(fmt.Sprintf("Found %d orphaned item(s) in repository:", len(orphans)))
	for _, o := range orphans {
		ui.Bullet(ui.StyleRed.Render(o))
	}

	fmt.Println()
	fmt.Printf("  %s Delete these from the repository? [y/N]: ", ui.StyleYellow.Render("⚠"))

	var choice string
	fmt.Scanln(&choice)
	choice = strings.ToLower(strings.TrimSpace(choice))

	if choice == "y" || choice == "yes" {
		ui.Step("Removing orphaned items...")
		for _, o := range orphans {
			err := os.RemoveAll(filepath.Join(eng.DotfileDir, o))
			if err != nil {
				ui.Error("Failed to remove " + o + ": " + err.Error())
			}
		}

		if !g.HasChanges() {
			ui.Success("Removed orphaned items (no git tracking changes needed)")
		} else {
			commitMsg := fmt.Sprintf("Cleaned orphaned dotfiles - %s", time.Now().Format("2006-01-02 15:04:05"))
			err := g.Commit(commitMsg)
			if err != nil {
				ui.Error("Failed to commit cleanup: " + err.Error())
			} else {
				ui.Success("Cleaned and committed successfully")

				fmt.Printf("  %s Push changes to remote? [Y/n]: ", ui.StyleCyan.Render("▸"))
				var pushChoice string
				fmt.Scanln(&pushChoice)
				pushChoice = strings.ToLower(strings.TrimSpace(pushChoice))

				if pushChoice == "" || pushChoice == "y" || pushChoice == "yes" {
					branch, _ := g.CurrentBranch()
					ui.Step("Pushing to remote...")
					if err := g.Push(branch, false); err != nil {
						ui.Error("Failed to push: " + err.Error())
					} else {
						ui.Success("Pushed successfully")
					}
				}
			}
		}
	} else {
		ui.Detail("Cleanup aborted.")
	}
}

func runExclude(cfg *config.ConfigManager, pattern string) {
	for _, p := range cfg.Excludes {
		if p == pattern {
			ui.Warn("'" + pattern + "' is already excluded")
			return
		}
	}

	cfg.Excludes = append(cfg.Excludes, pattern)
	cfg.SaveExcludes(cfg.Excludes)
	ui.Success("Added '" + pattern + "' to exclude patterns")

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
			commitMsg := fmt.Sprintf("Exclude %s - %s", pattern, time.Now().Format("2006-01-02 15:04:05"))
			g.Commit(commitMsg)
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

	if url != "" && url != "true" {
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

// Status

func runStatus(cfg *config.ConfigManager) {
	title := "Dotfile Status"
	if cfg.Hostname != "" {
		title += " (" + cfg.Hostname + ")"
	}
	ui.PrintHeader("🔍", title)

	eng, err := sync.NewEngine(cfg)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Step("Comparing live configs vs last sync...")
	fmt.Println()

	modified := 0
	added := 0
	deleted := 0
	unchanged := 0

	for _, target := range cfg.GetAllTargets() {
		srcPath := filepath.Join(eng.ConfigDir, target)
		destPath := filepath.Join(eng.DotfileDir, target)

		isSynxDataFile := false
		if strings.HasPrefix(target, "packages.") || strings.HasPrefix(target, "services.") {
			srcPath = filepath.Join(eng.Config.ConfigDir, target)
			destPath = filepath.Join(eng.DotfileDir, ".synx", target)
			isSynxDataFile = true
		}

		if cfg.IsProfileTarget(target) {
			if isSynxDataFile {
				destPath = filepath.Join(eng.DotfileDir, "profiles", cfg.ActiveProfile, ".synx", target)
			} else {
				destPath = filepath.Join(eng.DotfileDir, "profiles", cfg.ActiveProfile, target)
			}
		}

		if resolved, err := filepath.EvalSymlinks(srcPath); err == nil {
			srcPath = resolved
		}

		srcInfo, srcErr := os.Stat(srcPath)
		_, dstErr := os.Stat(destPath)

		if srcErr != nil && dstErr != nil {
			ui.Warn(fmt.Sprintf("%s %s", target, ui.StyleDim.Render("(not found)")))
			continue
		}

		if srcErr != nil {
			fmt.Printf("  %s %s\n", ui.StyleRed.Render("✗"), target+" "+ui.StyleDim.Render("(deleted locally)"))
			deleted++
			continue
		}

		if dstErr != nil {
			fmt.Printf("  %s %s\n", ui.StyleGreen.Render("+"), target+" "+ui.StyleDim.Render("(new, not yet synced)"))
			added++
			continue
		}

		if srcInfo.IsDir() {
			cmd := exec.Command("diff", "-rq", srcPath, destPath)
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Run()

			diffOutput := strings.TrimSpace(out.String())
			if diffOutput == "" {
				fmt.Printf("  %s %s\n", ui.StyleGreen.Render("✓"), target+" "+ui.StyleDim.Render("(no changes)"))
				unchanged++
			} else {
				lines := strings.Split(diffOutput, "\n")
				changeCount := len(lines)
				fmt.Printf("  %s %s\n", ui.StyleYellow.Render("✎"), target+" "+ui.StyleDim.Render(fmt.Sprintf("(%d file(s) changed)", changeCount)))
				max := 5
				if changeCount < max {
					max = changeCount
				}
				for i := 0; i < max; i++ {
					ui.Detail(lines[i])
				}
				if changeCount > 5 {
					ui.Detail(fmt.Sprintf("... and %d more", changeCount-5))
				}
				modified++
			}
		} else {
			cmd := exec.Command("diff", "-q", srcPath, destPath)
			if err := cmd.Run(); err != nil {
				fmt.Printf("  %s %s\n", ui.StyleYellow.Render("✎"), target+" "+ui.StyleDim.Render("(modified)"))
				modified++
			} else {
				fmt.Printf("  %s %s\n", ui.StyleGreen.Render("✓"), target+" "+ui.StyleDim.Render("(no changes)"))
				unchanged++
			}
		}
	}

	fmt.Println()
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Printf("%s modified  %s  %s new  %s  %s deleted  %s  %s unchanged\n",
		ui.StyleYellow.Render(fmt.Sprintf("%d", modified)),
		ui.StyleDim.Render("│"),
		ui.StyleGreen.Render(fmt.Sprintf("%d", added)),
		ui.StyleDim.Render("│"),
		ui.StyleRed.Render(fmt.Sprintf("%d", deleted)),
		ui.StyleDim.Render("│"),
		ui.StyleDim.Render(fmt.Sprintf("%d", unchanged)),
	)
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Println()
}

// Doctor

func runDoctor(cfg *config.ConfigManager) {
	ui.PrintHeader("🩺", "Doctor")
	ui.Step("Running health checks...")
	fmt.Println()

	passed := 0
	warnings := 0
	errors := 0

	eng, _ := sync.NewEngine(cfg)
	g := git.NewGitManager(eng.DotfileDir)

	if g.IsRepo() {
		ui.Success("Git repository OK")
		passed++
	} else {
		ui.Error("Not a git repository at " + eng.DotfileDir)
		errors++
	}

	if g.IsRepo() {
		remoteURL, err := g.RemoteURL()
		if err == nil && remoteURL != "" {
			ui.Success("Remote configured (" + remoteURL + ")")
			passed++
		} else {
			ui.Error("No remote configured")
			errors++
		}
	}

	if g.IsRepo() {
		branch, _ := g.CurrentBranch()
		if branch != "" {
			count, _ := g.UnpushedCount(branch)
			if count > 0 {
				ui.Warn(fmt.Sprintf("%d unpushed commit(s)", count))
				warnings++
			} else {
				ui.Success("No unpushed commits")
				passed++
			}
		}
	}

	allTargets := cfg.GetAllTargets()
	ui.Success(fmt.Sprintf("%d targets tracked", len(allTargets)))
	passed++

	home, _ := os.UserHomeDir()
	baseConfigDir := home + "/.config"
	for _, t := range allTargets {
		path := filepath.Join(baseConfigDir, t)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			ui.Warn(fmt.Sprintf("Target '%s' does not exist in ~/.config", t))
			warnings++
		}
	}

	brokenSymlinks := 0
	for _, t := range allTargets {
		path := filepath.Join(baseConfigDir, t)
		info, err := os.Lstat(path)
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				ui.Error(fmt.Sprintf("Broken symlink: %s", t))
				brokenSymlinks++
			}
		}
	}
	if brokenSymlinks == 0 {
		ui.Success("No broken symlinks")
		passed++
	} else {
		errors += brokenSymlinks
	}

	popularConfigs := []string{"waybar", "rofi", "nvim", "fish", "zsh", "tmux", "starship", "wezterm", "sway", "i3", "polybar", "dunst", "picom", "neofetch", "wofi"}
	targetSet := make(map[string]bool)
	for _, t := range allTargets {
		targetSet[t] = true
	}
	var untracked []string
	for _, p := range popularConfigs {
		if !targetSet[p] {
			if _, err := os.Stat(filepath.Join(baseConfigDir, p)); err == nil {
				untracked = append(untracked, p)
			}
		}
	}
	if len(untracked) > 0 {
		ui.Info(fmt.Sprintf("Found %d config(s) you might want to back up:", len(untracked)))
		for _, u := range untracked {
			ui.Detail(fmt.Sprintf("%s  →  synx --add %s", u, u))
		}
		warnings++
	}

	if g.IsRepo() {
		repoFiles, _ := g.LsFiles()
		for _, exc := range cfg.Excludes {
			matches := false
			for _, f := range repoFiles {
				if cfg.IsExcluded(f) {
					matches = true
					break
				}
			}
			if !matches {
				for _, t := range allTargets {
					if strings.HasPrefix(exc, t+"/") || exc == t {
						matches = true
						break
					}
				}
			}
			if !matches {
				ui.Warn(fmt.Sprintf("Exclude '%s' matches nothing in repo", exc))
				warnings++
			}
		}
	}

	if cfg.Hostname != "" {
		msg := "Machine: " + cfg.Hostname
		if cfg.ActiveProfile != "" {
			msg += " (using profile: " + cfg.ActiveProfile + ")"
		}
		ui.Success(msg)
		passed++
	} else {
		ui.Warn("Could not detect hostname")
		warnings++
	}

	for _, path := range []string{cfg.SynxConfig, cfg.ExcludeCfg} {
		if _, err := os.Stat(path); err == nil {
			f, err := os.OpenFile(path, os.O_RDWR, 0)
			if err != nil {
				ui.Warn(fmt.Sprintf("Config not writable: %s", filepath.Base(path)))
				warnings++
			} else {
				f.Close()
			}
		}
	}

	hooksDir := filepath.Join(cfg.ConfigDir, "hooks")
	if _, err := os.Stat(hooksDir); err == nil {
		entries, _ := os.ReadDir(hooksDir)
		ui.Success(fmt.Sprintf("Hooks directory found (%d hook(s))", len(entries)))
		passed++
	}

	profilesDir := filepath.Join(cfg.ConfigDir, "profiles")
	pNames, _ := profiles.ListProfiles(profilesDir)
	if len(pNames) > 0 {
		active := profiles.GetActiveProfile(cfg.ConfigDir)
		if active != "" {
			ui.Success(fmt.Sprintf("%d profile(s) available, active: %s", len(pNames), active))
		} else {
			ui.Success(fmt.Sprintf("%d profile(s) available", len(pNames)))
		}
		passed++
	}

	nativePath := packages.NativeListPath(cfg.ConfigDir)
	foreignPath := packages.ForeignListPath(cfg.ConfigDir)
	nativeList, _ := packages.LoadList(nativePath)
	foreignList, _ := packages.LoadList(foreignPath)
	if nativeList != nil || foreignList != nil {
		ui.Success(fmt.Sprintf("Package state: %d native, %d foreign (AUR)", len(nativeList), len(foreignList)))
		passed++
	} else {
		ui.Warn("No saved package state — run 'synx pkg sync'")
		warnings++
	}

	systemPath := services.SystemListPath(cfg.ConfigDir)
	userPath := services.UserListPath(cfg.ConfigDir)
	systemList, _ := services.LoadList(systemPath)
	userList, _ := services.LoadList(userPath)
	if systemList != nil || userList != nil {
		ui.Success(fmt.Sprintf("Service state: %d system, %d user", len(systemList), len(userList)))
		passed++
	} else {
		ui.Warn("No saved service state — run 'synx svc sync'")
		warnings++
	}

	fmt.Println()
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Printf("%s passed  %s  %s warnings  %s  %s errors\n",
		ui.StyleGreen.Render(fmt.Sprintf("%d", passed)),
		ui.StyleDim.Render("│"),
		ui.StyleYellow.Render(fmt.Sprintf("%d", warnings)),
		ui.StyleDim.Render("│"),
		ui.StyleRed.Render(fmt.Sprintf("%d", errors)),
	)
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Println()
}

// Profiles

func runProfileApply(cfg *config.ConfigManager, name string) {
	ui.PrintHeader("🎨", "Apply Profile")

	profilesDir := filepath.Join(cfg.ConfigDir, "profiles")

	profileDir := filepath.Join(profilesDir, name)
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		ui.Error(fmt.Sprintf("Profile '%s' does not exist", name))
		os.Exit(1)
	}

	hooksDir := filepath.Join(cfg.ConfigDir, "hooks")
	if err := hooks.RunHook(hooksDir, hooks.PreProfile, name); err != nil {
		ui.Error("pre-profile hook failed: " + err.Error())
		os.Exit(1)
	}

	ui.Step("Syncing current state before profile switch...")
	runSync(cfg)

	if err := profiles.SetActiveProfile(cfg.ConfigDir, name); err != nil {
		ui.Warn("Could not save active profile marker")
	}

	cfg.Load()

	fmt.Println()
	runRestore(cfg)

	for _, t := range cfg.ProfileTargets {
		if strings.HasPrefix(t, "hypr/") || t == "hypr" {
			if _, err := os.Stat("/usr/bin/hyprctl"); err == nil {
				fmt.Println()
				ui.Step("Reloading Hyprland...")
				cmd := exec.Command("hyprctl", "reload")
				if err := cmd.Run(); err == nil {
					ui.Success("Hyprland reloaded")
				}
			}
			break
		}
	}

	hooks.RunHook(hooksDir, hooks.PostProfile, name)

	fmt.Println()
	fmt.Printf("%s  %s\n", ui.StyleGreen.Render("✓"), ui.StyleBold.Render("Profile '"+name+"' active"))
}

func runProfileList(cfg *config.ConfigManager) {
	ui.PrintHeader("🎨", "Profiles")

	profilesDir := filepath.Join(cfg.ConfigDir, "profiles")
	names, err := profiles.ListProfiles(profilesDir)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	active := profiles.GetActiveProfile(cfg.ConfigDir)

	if len(names) == 0 {
		ui.Detail("No profiles found")
		ui.Detail("Create one with: synx --profile-create <name>")
		fmt.Println()
		return
	}

	for _, name := range names {
		targetCount := profiles.ProfileTargetCount(profilesDir, name)

		label := name
		tags := []string{fmt.Sprintf("%d target(s)", targetCount)}
		if name == active {
			tags = append(tags, "active")
		}
		label += " " + ui.StyleDim.Render("("+strings.Join(tags, ", ")+")")

		if name == active {
			fmt.Printf("  %s %s\n", ui.StyleGreen.Render("●"), label)
		} else {
			fmt.Printf("  %s %s\n", ui.StyleDim.Render("○"), label)
		}
	}
	fmt.Println()
}

func runProfileCreate(cfg *config.ConfigManager, name string) {
	ui.PrintHeader("🎨", "Create Profile")

	profilesDir := filepath.Join(cfg.ConfigDir, "profiles")
	if err := profiles.CreateProfile(profilesDir, name); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	cfg.SaveProfileTargets(name, []string{})
	cfg.SaveProfileExcludes(name, []string{})

	ui.Success(fmt.Sprintf("Created profile '%s'", name))
	ui.Detail("Switch to it with: synx --profile " + name)
	fmt.Println()
}

func runProfileDelete(cfg *config.ConfigManager, name string) {
	ui.PrintHeader("🎨", "Delete Profile")

	profilesDir := filepath.Join(cfg.ConfigDir, "profiles")
	if err := profiles.DeleteProfile(profilesDir, name); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	active := profiles.GetActiveProfile(cfg.ConfigDir)
	if active == name {
		profiles.ClearActiveProfile(cfg.ConfigDir)
	}

	ui.Success(fmt.Sprintf("Deleted profile '%s'", name))
	fmt.Println()
}

// Remote Diff

func runRemoteDiff(cfg *config.ConfigManager) {
	ui.PrintHeader("🌐", "Remote Diff")

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

	ui.Step("Fetching from remote...")
	g.Fetch()

	branch, _ := g.CurrentBranch()

	ui.Step("Syncing local state for comparison...")
	syncEng, _ := sync.NewEngine(cfg)
	syncEng.SyncToRepo()

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = eng.DotfileDir
	addCmd.Run()

	stat, err := g.DiffStatRemote(branch)
	if err != nil {
		ui.Error("Failed to diff against remote: " + err.Error())
		os.Exit(1)
	}

	if stat == "" {
		fmt.Println()
		ui.Success("No differences between local and remote")
		fmt.Println()
		return
	}

	fmt.Println()
	ui.Step("Differences:")
	fmt.Println()

	diffCmd := exec.Command("git", "diff", "--color=always", "origin/"+branch, "--", ".")
	diffCmd.Dir = eng.DotfileDir
	diffCmd.Stdout = os.Stdout
	diffCmd.Stderr = os.Stderr
	diffCmd.Run()

	fmt.Println()
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	lines := strings.Split(stat, "\n")
	if len(lines) > 0 {
		fmt.Printf("  %s\n", lines[len(lines)-1])
	}
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Println()

	resetCmd := exec.Command("git", "reset", "HEAD", "--quiet")
	resetCmd.Dir = eng.DotfileDir
	resetCmd.Run()
}

func runUpdate() {
	ui.PrintHeader("⬆", "Self-Update")
	fmt.Println()

	exePath, err := os.Executable()
	if err != nil {
		ui.Error("Cannot determine binary location: " + err.Error())
		os.Exit(1)
	}

	// Check if installed via AUR/system package manager
	if strings.HasPrefix(exePath, "/usr/bin/") || strings.HasPrefix(exePath, "/usr/local/bin/") || exePath == "/bin/synx" {
		// Check if it's a symlink (our install.sh creates /bin/synx -> ...)
		isSymlink := false
		if info, err := os.Lstat(exePath); err == nil {
			isSymlink = info.Mode()&os.ModeSymlink != 0
		}

		if !isSymlink {
			ui.Info("synx was installed via a package manager.")
			ui.Detail("Please use your package manager (e.g., yay -Syu) to update.")
			return
		}
	}

	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		ui.Error("Cannot resolve binary path: " + err.Error())
		os.Exit(1)
	}

	binaryDir := filepath.Dir(realPath)
	sourceDir := binaryDir // Assume binary is in synx-go/
	goMod := filepath.Join(sourceDir, "go.mod")
	if _, err := os.Stat(goMod); err != nil {
		sourceDir = filepath.Dir(binaryDir)
		goMod = filepath.Join(sourceDir, "synx-go", "go.mod")
		if _, err := os.Stat(goMod); err == nil {
			sourceDir = filepath.Join(sourceDir, "synx-go")
		} else {
			ui.Error("Cannot find synx source directory")
			ui.Detail("Binary at: " + realPath)
			ui.Detail("Looked for go.mod in: " + binaryDir)
			os.Exit(1)
		}
	}

	repoDir := sourceDir
	for repoDir != "/" {
		if _, err := os.Stat(filepath.Join(repoDir, ".git")); err == nil {
			break
		}
		repoDir = filepath.Dir(repoDir)
	}
	if repoDir == "/" {
		ui.Error("Cannot find git repository for synx source")
		os.Exit(1)
	}

	ui.Step("Pulling latest changes...")
	pullCmd := exec.Command("git", "pull", "--rebase")
	pullCmd.Dir = repoDir
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		ui.Error("Failed to pull: " + err.Error())
		os.Exit(1)
	}
	ui.Success("Source updated")

	fmt.Println()
	ui.Step("Building new binary via build.sh...")

	buildCmd := exec.Command("./build.sh")
	buildCmd.Dir = sourceDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		ui.Error("Build script failed: " + err.Error())
		os.Exit(1)
	}

	newBin := filepath.Join(sourceDir, "synx")
	if _, err := os.Stat(newBin); err != nil {
		ui.Error("Cannot find newly compiled binary at " + newBin)
		os.Exit(1)
	}

	if newBin == realPath {
		ui.Success("Binary updated in-place at " + realPath)
	} else {
		ui.Step("Replacing binary...")
		tmpBin := realPath + ".old"
		os.Rename(realPath, tmpBin) // Move current out of the way

		if err := os.Rename(newBin, realPath); err != nil {
			os.Rename(tmpBin, realPath)
			ui.Error("Failed to replace binary: " + err.Error())
			os.Exit(1)
		}
		os.Remove(tmpBin)

		ui.Success("Binary updated at " + realPath)
	}

	fmt.Println()
	verCmd := exec.Command("git", "log", "--oneline", "-1")
	verCmd.Dir = repoDir
	out, err := verCmd.Output()
	if err == nil {
		ui.Info("Now at: " + strings.TrimSpace(string(out)))
	}

	fmt.Println()
	ui.Success("synx updated successfully! 🎉")
}
