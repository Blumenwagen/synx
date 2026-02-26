package cmd

import (
	"fmt"
	"os"

	"github.com/Blumenwagen/synx/pkg/config"
	"github.com/Blumenwagen/synx/pkg/services"
	"github.com/Blumenwagen/synx/pkg/ui"
	"github.com/spf13/cobra"
)

var svcCmd = &cobra.Command{
	Use:   "svc",
	Short: "Manage tracked systemd services",
	Long:  "Snapshot, diff, and restore your enabled systemd services (system + user).",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if !services.HasSystemctl() {
			ui.Error("systemctl not found — 'synx svc' requires a systemd-based system")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var svcSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Snapshot currently enabled services",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := mustLoadConfig()
		runSvcSync(cfg)
	},
}

var svcStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show service changes since last sync",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := mustLoadConfig()
		runSvcStatus(cfg)
	},
}

var svcRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Enable missing services from saved list",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := mustLoadConfig()
		runSvcRestore(cfg)
	},
}

var svcListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show saved service list",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := mustLoadConfig()
		runSvcList(cfg)
	},
}

func init() {
	rootCmd.AddCommand(svcCmd)
	svcCmd.AddCommand(svcSyncCmd)
	svcCmd.AddCommand(svcStatusCmd)
	svcCmd.AddCommand(svcRestoreCmd)
	svcCmd.AddCommand(svcListCmd)
}

func runSvcSync(cfg *config.ConfigManager) {
	ui.PrintHeader("⚙", "Service Sync")

	ui.Step("Capturing enabled services...")

	system, err := services.CaptureSystem()
	if err != nil {
		ui.Error("Failed to capture system services: " + err.Error())
		os.Exit(1)
	}

	user, err := services.CaptureUser()
	if err != nil {
		ui.Error("Failed to capture user services: " + err.Error())
		os.Exit(1)
	}

	systemPath := services.SystemListPath(cfg.ConfigDir)
	userPath := services.UserListPath(cfg.ConfigDir)

	if err := services.SaveList(systemPath, "System services (enabled)", system); err != nil {
		ui.Error("Failed to save system list: " + err.Error())
		os.Exit(1)
	}

	if err := services.SaveList(userPath, "User services (enabled)", user); err != nil {
		ui.Error("Failed to save user list: " + err.Error())
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Printf("%s system  %s  %s user\n",
		ui.StyleGreen.Render(fmt.Sprintf("%d", len(system))),
		ui.StyleDim.Render("│"),
		ui.StyleCyan.Render(fmt.Sprintf("%d", len(user))),
	)
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Println()

	ui.Success("Service state saved")
}

func runSvcStatus(cfg *config.ConfigManager) {
	ui.PrintHeader("⚙", "Service Status")

	systemPath := services.SystemListPath(cfg.ConfigDir)
	userPath := services.UserListPath(cfg.ConfigDir)

	savedSystem, err := services.LoadList(systemPath)
	if err != nil {
		ui.Error("Failed to load system list: " + err.Error())
		os.Exit(1)
	}
	if savedSystem == nil {
		ui.Warn("No saved service state — run 'synx svc sync' first")
		return
	}

	savedUser, _ := services.LoadList(userPath)

	ui.Step("Comparing saved vs current services...")

	currentSystem, err := services.CaptureSystem()
	if err != nil {
		ui.Error("Failed to capture system services: " + err.Error())
		os.Exit(1)
	}
	currentUser, _ := services.CaptureUser()

	systemAdded, systemRemoved := services.Diff(savedSystem, currentSystem)
	userAdded, userRemoved := services.Diff(savedUser, currentUser)

	fmt.Println()

	if len(systemAdded) > 0 || len(systemRemoved) > 0 {
		fmt.Println(ui.StyleBold.Render("SYSTEM SERVICES:"))
		printDiffList(systemAdded, systemRemoved)
	}

	if len(userAdded) > 0 || len(userRemoved) > 0 {
		fmt.Println(ui.StyleBold.Render("USER SERVICES:"))
		printDiffList(userAdded, userRemoved)
	}

	totalAdded := len(systemAdded) + len(userAdded)
	totalRemoved := len(systemRemoved) + len(userRemoved)

	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Printf("%s added  %s  %s removed\n",
		ui.StyleGreen.Render(fmt.Sprintf("%d", totalAdded)),
		ui.StyleDim.Render("│"),
		ui.StyleRed.Render(fmt.Sprintf("%d", totalRemoved)),
	)
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Println()

	if totalAdded == 0 && totalRemoved == 0 {
		ui.Success("Service state matches saved snapshot")
	}
}

func runSvcRestore(cfg *config.ConfigManager) {
	ui.PrintHeader("⚙", "Service Restore")

	systemPath := services.SystemListPath(cfg.ConfigDir)
	userPath := services.UserListPath(cfg.ConfigDir)

	savedSystem, err := services.LoadList(systemPath)
	if err != nil || savedSystem == nil {
		ui.Error("No saved service state — run 'synx svc sync' first")
		os.Exit(1)
	}
	savedUser, _ := services.LoadList(userPath)

	ui.Step("Comparing saved vs current services...")

	currentSystem, err := services.CaptureSystem()
	if err != nil {
		ui.Error("Failed to capture system services: " + err.Error())
		os.Exit(1)
	}
	currentUser, _ := services.CaptureUser()

	_, systemMissing := services.Diff(currentSystem, savedSystem)
	_, userMissing := services.Diff(currentUser, savedUser)

	if len(systemMissing) == 0 && len(userMissing) == 0 {
		ui.Success("All saved services are enabled")
		return
	}

	if len(systemMissing) > 0 {
		fmt.Println()
		ui.Step(fmt.Sprintf("Enabling %d missing system service(s)...", len(systemMissing)))
		for _, s := range systemMissing {
			ui.Detail(s)
		}
		if err := services.EnableSystem(systemMissing); err != nil {
			ui.Error("Some system services failed to enable")
		} else {
			ui.Success(fmt.Sprintf("Enabled %d system service(s)", len(systemMissing)))
		}
	}

	if len(userMissing) > 0 {
		fmt.Println()
		ui.Step(fmt.Sprintf("Enabling %d missing user service(s)...", len(userMissing)))
		for _, s := range userMissing {
			ui.Detail(s)
		}
		if err := services.EnableUser(userMissing); err != nil {
			ui.Error("Some user services failed to enable")
		} else {
			ui.Success(fmt.Sprintf("Enabled %d user service(s)", len(userMissing)))
		}
	}

	fmt.Println()
	ui.Success("Service restore complete")
}

func runSvcList(cfg *config.ConfigManager) {
	ui.PrintHeader("⚙", "Saved Services")

	systemPath := services.SystemListPath(cfg.ConfigDir)
	userPath := services.UserListPath(cfg.ConfigDir)

	system, _ := services.LoadList(systemPath)
	user, _ := services.LoadList(userPath)

	if system == nil && user == nil {
		ui.Warn("No saved service state — run 'synx svc sync' first")
		return
	}

	if len(system) > 0 {
		fmt.Println(ui.StyleBold.Render("SYSTEM SERVICES:"))
		ui.Detail(fmt.Sprintf("%d service(s)", len(system)))
		max := 20
		if len(system) < max {
			max = len(system)
		}
		for i := 0; i < max; i++ {
			ui.Bullet(system[i])
		}
		if len(system) > 20 {
			ui.Detail(fmt.Sprintf("... and %d more", len(system)-20))
		}
		fmt.Println()
	}

	if len(user) > 0 {
		fmt.Println(ui.StyleBold.Render("USER SERVICES:"))
		ui.Detail(fmt.Sprintf("%d service(s)", len(user)))
		max := 20
		if len(user) < max {
			max = len(user)
		}
		for i := 0; i < max; i++ {
			ui.Bullet(user[i])
		}
		if len(user) > 20 {
			ui.Detail(fmt.Sprintf("... and %d more", len(user)-20))
		}
		fmt.Println()
	}

	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Printf("%s system  %s  %s user\n",
		ui.StyleGreen.Render(fmt.Sprintf("%d", len(system))),
		ui.StyleDim.Render("│"),
		ui.StyleCyan.Render(fmt.Sprintf("%d", len(user))),
	)
	fmt.Println(ui.StyleDim.Render("─────────────────────────────────────"))
	fmt.Println()
}
