package cmd

import (
	"fmt"
	"os"

	"github.com/Blumenwagen/synx/pkg/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "synx",
	Short: "Dotfile backup system",
	Long:  "A powerful CLI tool for managing dotfiles with git-based version control, machine-specific exclusions, and full system bootstrapping.",
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Println("synx version " + Version)
			return
		}

		// By default run sync
		ui.PrintHeader("🚀", "Dotfile Sync")
		fmt.Println()
		ui.Step("Syncing dotfiles...")
		// TODO: Call sync engine
	},
}

var (
	Version = "dev"

	restoreFlag       bool
	listFlag          bool
	historyFlag       bool
	rollbackFlag      int
	addFlag           string
	removeFlag        string
	excludeFlag       string
	bsSetupFlag       bool
	bootstrapFlag     string
	yesFlag           bool
	machineFlag       bool
	statusFlag        bool
	dryRunFlag        bool
	doctorFlag        bool
	profileFlag       string
	profileListFlag   bool
	profileCreateFlag string
	profileDeleteFlag string
	remoteDiffFlag    bool
	updateFlag        bool
	cleanFlag         bool
	versionFlag       bool
)

func init() {

	rootCmd.Flags().BoolVarP(&restoreFlag, "restore", "r", false, "Restore dotfiles from remote repository")
	rootCmd.Flags().BoolVar(&listFlag, "list", false, "List tracked and available dotfiles")
	rootCmd.Flags().BoolVar(&historyFlag, "history", false, "Show sync history")
	rootCmd.Flags().IntVar(&rollbackFlag, "rollback", -1, "Rollback to n commits ago (omit n for interactive UI)")
	rootCmd.Flag("rollback").NoOptDefVal = "0"
	rootCmd.Flags().StringVar(&addFlag, "add", "", "Add a dotfile to track")
	rootCmd.Flags().StringVar(&removeFlag, "remove", "", "Remove a dotfile from tracking")
	rootCmd.Flags().BoolVar(&cleanFlag, "clean", false, "Clean untracked/orphaned dotfiles from the repository")
	rootCmd.Flags().StringVar(&excludeFlag, "exclude", "", "Add exclude pattern and remove from repo")
	rootCmd.Flags().BoolVar(&bsSetupFlag, "bootstrap-setup", false, "Create bootstrap config interactively")
	rootCmd.Flags().StringVar(&bootstrapFlag, "bootstrap", "", "Run bootstrap from local config or clone provided repo URL")
	rootCmd.Flags().BoolVar(&yesFlag, "yes", false, "Skip per-step confirmations on bootstrap")
	rootCmd.Flags().BoolVarP(&machineFlag, "machine", "m", false, "Target machine-specific config instead of shared base")
	rootCmd.Flags().BoolVarP(&statusFlag, "status", "s", false, "Show what changed since last sync")
	rootCmd.Flags().BoolVarP(&dryRunFlag, "dry-run", "n", false, "Preview sync/restore without making changes")
	rootCmd.Flags().BoolVar(&doctorFlag, "doctor", false, "Run health checks on synx setup")
	rootCmd.Flags().StringVar(&profileFlag, "profile", "", "Apply a named profile")
	rootCmd.Flags().BoolVar(&profileListFlag, "profile-list", false, "List available profiles")
	rootCmd.Flags().StringVar(&profileCreateFlag, "profile-create", "", "Create a new profile")
	rootCmd.Flags().StringVar(&profileDeleteFlag, "profile-delete", "", "Delete a profile")
	rootCmd.Flags().BoolVar(&remoteDiffFlag, "remote-diff", false, "Show diff between local and remote dotfiles")
	rootCmd.Flags().BoolVar(&updateFlag, "update", false, "Update synx to the latest version")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print the synx version")

	// Override the default help
	rootCmd.SetHelpFunc(customHelp)
}

func customHelp(cmd *cobra.Command, args []string) {
	h := ui.StyleBold.Render
	d := ui.StyleDim.Render
	c := ui.StyleCyan.Render
	g := ui.StyleGreen.Render

	fmt.Println()
	fmt.Println(h("SYNX") + d(" — dotfile backup system"))
	fmt.Println(d("A fast CLI tool for managing dotfiles with git-backed sync,"))
	fmt.Println(d("multi-machine support, profiles, and full system bootstrapping."))
	fmt.Println()

	// Sync & Restore
	fmt.Println(h("SYNC & RESTORE"))
	fmt.Printf("  %s                         Sync dotfiles to remote\n", g("synx"))
	fmt.Printf("  %s %s                  Restore from remote\n", g("synx"), c("-r, --restore"))
	fmt.Printf("  %s %s                  Preview without changes\n", g("synx"), c("-n, --dry-run"))
	fmt.Println()

	// Status & Diagnostics
	fmt.Println(h("STATUS & DIAGNOSTICS"))
	fmt.Printf("  %s %s                   Show changes since last sync\n", g("synx"), c("-s, --status"))
	fmt.Printf("  %s %s               Compare local vs remote\n", g("synx"), c("--remote-diff"))
	fmt.Printf("  %s %s                    Run health checks\n", g("synx"), c("--doctor"))
	fmt.Printf("  %s %s                   Show sync history\n", g("synx"), c("--history"))
	fmt.Println()

	// Target Management
	fmt.Println(h("TARGET MANAGEMENT"))
	fmt.Printf("  %s %s               Track a new dotfile directory\n", g("synx"), c("--add <name>"))
	fmt.Printf("  %s %s            Stop tracking a dotfile\n", g("synx"), c("--remove <name>"))
	fmt.Printf("  %s %s             Clean orphaned files from repo\n", g("synx"), c("--clean"))
	fmt.Printf("  %s %s          Exclude a file pattern\n", g("synx"), c("--exclude <path>"))
	fmt.Printf("  %s %s                      List tracked dotfiles\n", g("synx"), c("--list"))
	fmt.Printf("  %s %s              Rollback n commits\n", g("synx"), c("--rollback <n>"))
	fmt.Println()

	// Profiles
	fmt.Println(h("PROFILES"))
	fmt.Printf("  %s %s         Apply a named profile\n", g("synx"), c("--profile <name>"))
	fmt.Printf("  %s %s              List available profiles\n", g("synx"), c("--profile-list"))
	fmt.Printf("  %s %s  Create a new empty profile\n", g("synx"), c("--profile-create <name>"))
	fmt.Printf("  %s %s  Delete a profile\n", g("synx"), c("--profile-delete <name>"))
	fmt.Println()

	// Bootstrap
	fmt.Println(h("BOOTSTRAP"))
	fmt.Printf("  %s %s          Create bootstrap config\n", g("synx"), c("--bootstrap-setup"))
	fmt.Printf("  %s %s       Clone & bootstrap from URL\n", g("synx"), c("--bootstrap <url>"))
	fmt.Println()

	// Modifiers
	fmt.Println(h("MODIFIERS"))
	fmt.Printf("  %s              Target machine-specific config\n", c("-m, --machine"))
	fmt.Printf("  %s                   Skip confirmations (bootstrap)\n", c("--yes"))
	fmt.Printf("  %s                Update synx to the latest version\n", c("--update"))
	fmt.Printf("  %s               Print the current synx version\n", c("-v, --version"))
	fmt.Println()

	// Packages & Services
	fmt.Println(h("PACKAGES & SERVICES"))
	fmt.Printf("  %s %s              Snapshot installed packages\n", g("synx pkg"), c("sync"))
	fmt.Printf("  %s %s            Package changes since last sync\n", g("synx pkg"), c("status"))
	fmt.Printf("  %s %s           Install missing packages\n", g("synx pkg"), c("restore"))
	fmt.Printf("  %s %s              Show saved package list\n", g("synx pkg"), c("list"))
	fmt.Printf("  %s %s              Snapshot enabled services\n", g("synx svc"), c("sync"))
	fmt.Printf("  %s %s            Service changes since last sync\n", g("synx svc"), c("status"))
	fmt.Printf("  %s %s           Enable missing services\n", g("synx svc"), c("restore"))
	fmt.Printf("  %s %s              Show saved service list\n", g("synx svc"), c("list"))
	fmt.Println()
	fmt.Println()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
