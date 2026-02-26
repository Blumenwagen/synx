package cmd

import (
	"fmt"
	"os"

	"github.com/Blumenwagen/synx/pkg/config"
	"github.com/Blumenwagen/synx/pkg/packages"
	"github.com/Blumenwagen/synx/pkg/ui"
	"github.com/spf13/cobra"
)

var pkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: "Manage tracked packages",
	Long:  "Snapshot, diff, and restore your installed packages (pacman + AUR).",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if !packages.HasPacman() {
			ui.Error("pacman not found — 'synx pkg' requires Arch Linux or an Arch-based distro")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var pkgSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Snapshot currently installed packages",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := mustLoadConfig()
		runPkgSync(cfg)
	},
}

var pkgStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show package changes since last sync",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := mustLoadConfig()
		runPkgStatus(cfg)
	},
}

var pkgRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Install missing packages from saved list",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := mustLoadConfig()
		runPkgRestore(cfg)
	},
}

var pkgListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show saved package list",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := mustLoadConfig()
		runPkgList(cfg)
	},
}

func init() {
	rootCmd.AddCommand(pkgCmd)
	pkgCmd.AddCommand(pkgSyncCmd)
	pkgCmd.AddCommand(pkgStatusCmd)
	pkgCmd.AddCommand(pkgRestoreCmd)
	pkgCmd.AddCommand(pkgListCmd)
}

func runPkgSync(cfg *config.ConfigManager) {
	ui.PrintHeader("📦", "Package Sync")

	ui.Step("Capturing installed packages...")

	native, err := packages.CaptureNative()
	if err != nil {
		ui.Error("Failed to capture native packages: " + err.Error())
		os.Exit(1)
	}

	foreign, err := packages.CaptureForeign()
	if err != nil {
		ui.Error("Failed to capture foreign packages: " + err.Error())
		os.Exit(1)
	}

	nativePath := packages.NativeListPath(cfg.ConfigDir)
	foreignPath := packages.ForeignListPath(cfg.ConfigDir)

	if err := packages.SaveList(nativePath, "Native (repo) packages", native); err != nil {
		ui.Error("Failed to save native list: " + err.Error())
		os.Exit(1)
	}

	if err := packages.SaveList(foreignPath, "Foreign (AUR) packages", foreign); err != nil {
		ui.Error("Failed to save foreign list: " + err.Error())
		os.Exit(1)
	}

	helper := packages.DetectAURHelper()

	fmt.Println()
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Printf("%s native  %s  %s foreign (AUR)\n",
		ui.StyleGreen.Render(fmt.Sprintf("%d", len(native))),
		ui.StyleDim.Render("│"),
		ui.StyleCyan.Render(fmt.Sprintf("%d", len(foreign))),
	)
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Println()

	if helper != "" {
		ui.Detail("AUR helper: " + helper)
	}

	ui.Success("Package state saved")
}

func runPkgStatus(cfg *config.ConfigManager) {
	ui.PrintHeader("📦", "Package Status")

	nativePath := packages.NativeListPath(cfg.ConfigDir)
	foreignPath := packages.ForeignListPath(cfg.ConfigDir)

	savedNative, err := packages.LoadList(nativePath)
	if err != nil {
		ui.Error("Failed to load native list: " + err.Error())
		os.Exit(1)
	}
	if savedNative == nil {
		ui.Warn("No saved package state — run 'synx pkg sync' first")
		return
	}

	savedForeign, _ := packages.LoadList(foreignPath)

	ui.Step("Comparing saved vs current packages...")

	currentNative, err := packages.CaptureNative()
	if err != nil {
		ui.Error("Failed to capture native packages: " + err.Error())
		os.Exit(1)
	}
	currentForeign, _ := packages.CaptureForeign()

	nativeAdded, nativeRemoved := packages.Diff(savedNative, currentNative)
	foreignAdded, foreignRemoved := packages.Diff(savedForeign, currentForeign)

	fmt.Println()

	// Native changes
	if len(nativeAdded) > 0 || len(nativeRemoved) > 0 {
		fmt.Println(ui.StyleBold.Render("NATIVE PACKAGES:"))
		printDiffList(nativeAdded, nativeRemoved)
	}

	// Foreign changes
	if len(foreignAdded) > 0 || len(foreignRemoved) > 0 {
		fmt.Println(ui.StyleBold.Render("FOREIGN (AUR) PACKAGES:"))
		printDiffList(foreignAdded, foreignRemoved)
	}

	totalAdded := len(nativeAdded) + len(foreignAdded)
	totalRemoved := len(nativeRemoved) + len(foreignRemoved)

	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Printf("%s added  %s  %s removed\n",
		ui.StyleGreen.Render(fmt.Sprintf("%d", totalAdded)),
		ui.StyleDim.Render("│"),
		ui.StyleRed.Render(fmt.Sprintf("%d", totalRemoved)),
	)
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Println()

	if totalAdded == 0 && totalRemoved == 0 {
		ui.Success("Package state matches saved snapshot")
	}
}

func runPkgRestore(cfg *config.ConfigManager) {
	ui.PrintHeader("📦", "Package Restore")

	nativePath := packages.NativeListPath(cfg.ConfigDir)
	foreignPath := packages.ForeignListPath(cfg.ConfigDir)

	savedNative, err := packages.LoadList(nativePath)
	if err != nil || savedNative == nil {
		ui.Error("No saved package state — run 'synx pkg sync' first")
		os.Exit(1)
	}
	savedForeign, _ := packages.LoadList(foreignPath)

	ui.Step("Comparing saved vs current packages...")

	currentNative, err := packages.CaptureNative()
	if err != nil {
		ui.Error("Failed to capture native packages: " + err.Error())
		os.Exit(1)
	}
	currentForeign, _ := packages.CaptureForeign()

	_, nativeMissing := packages.Diff(currentNative, savedNative)
	_, foreignMissing := packages.Diff(currentForeign, savedForeign)

	if len(nativeMissing) == 0 && len(foreignMissing) == 0 {
		ui.Success("All saved packages are installed")
		return
	}

	// Install native
	if len(nativeMissing) > 0 {
		fmt.Println()
		ui.Step(fmt.Sprintf("Installing %d missing native package(s)...", len(nativeMissing)))
		for _, p := range nativeMissing {
			ui.Detail(p)
		}
		if err := packages.InstallNative(nativeMissing); err != nil {
			ui.Error("Some native packages failed to install")
		} else {
			ui.Success(fmt.Sprintf("Installed %d native package(s)", len(nativeMissing)))
		}
	}

	// Install foreign
	if len(foreignMissing) > 0 {
		helper := packages.DetectAURHelper()
		if helper == "" {
			ui.Warn(fmt.Sprintf("No AUR helper found — %d foreign package(s) skipped", len(foreignMissing)))
			for _, p := range foreignMissing {
				ui.Detail(p)
			}
		} else {
			fmt.Println()
			ui.Step(fmt.Sprintf("Installing %d missing AUR package(s) via %s...", len(foreignMissing), helper))
			for _, p := range foreignMissing {
				ui.Detail(p)
			}
			if err := packages.InstallForeign(helper, foreignMissing); err != nil {
				ui.Error("Some AUR packages failed to install")
			} else {
				ui.Success(fmt.Sprintf("Installed %d AUR package(s)", len(foreignMissing)))
			}
		}
	}

	fmt.Println()
	ui.Success("Package restore complete")
}

func runPkgList(cfg *config.ConfigManager) {
	ui.PrintHeader("📦", "Saved Packages")

	nativePath := packages.NativeListPath(cfg.ConfigDir)
	foreignPath := packages.ForeignListPath(cfg.ConfigDir)

	native, _ := packages.LoadList(nativePath)
	foreign, _ := packages.LoadList(foreignPath)

	if native == nil && foreign == nil {
		ui.Warn("No saved package state — run 'synx pkg sync' first")
		return
	}

	if len(native) > 0 {
		fmt.Println(ui.StyleBold.Render("NATIVE PACKAGES:"))
		ui.Detail(fmt.Sprintf("%d package(s)", len(native)))
		max := 20
		if len(native) < max {
			max = len(native)
		}
		for i := 0; i < max; i++ {
			ui.Bullet(native[i])
		}
		if len(native) > 20 {
			ui.Detail(fmt.Sprintf("... and %d more", len(native)-20))
		}
		fmt.Println()
	}

	if len(foreign) > 0 {
		fmt.Println(ui.StyleBold.Render("FOREIGN (AUR) PACKAGES:"))
		ui.Detail(fmt.Sprintf("%d package(s)", len(foreign)))
		max := 20
		if len(foreign) < max {
			max = len(foreign)
		}
		for i := 0; i < max; i++ {
			ui.Bullet(foreign[i])
		}
		if len(foreign) > 20 {
			ui.Detail(fmt.Sprintf("... and %d more", len(foreign)-20))
		}
		fmt.Println()
	}

	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Printf("%s native  %s  %s foreign\n",
		ui.StyleGreen.Render(fmt.Sprintf("%d", len(native))),
		ui.StyleDim.Render("│"),
		ui.StyleCyan.Render(fmt.Sprintf("%d", len(foreign))),
	)
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Println()
}

// printDiffList shows added/removed items.
func printDiffList(added, removed []string) {
	max := 10
	shown := 0

	for _, p := range added {
		if shown >= max {
			break
		}
		fmt.Printf("  %s %s\n", ui.StyleGreen.Render("+"), p)
		shown++
	}
	for _, p := range removed {
		if shown >= max {
			break
		}
		fmt.Printf("  %s %s\n", ui.StyleRed.Render("-"), p)
		shown++
	}

	total := len(added) + len(removed)
	if total > max {
		ui.Detail(fmt.Sprintf("... and %d more", total-max))
	}
	fmt.Println()
}

// mustLoadConfig is a helper to load config or exit.
func mustLoadConfig() *config.ConfigManager {
	cfg, err := config.NewConfigManager()
	if err != nil {
		ui.Error(fmt.Sprintf("Init error: %v", err))
		os.Exit(1)
	}
	cfg.Load()
	return cfg
}
