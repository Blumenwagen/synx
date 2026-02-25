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
		// By default run sync
		ui.PrintHeader("🚀", "Dotfile Sync")
		fmt.Println()
		ui.Step("Syncing dotfiles...")
		// TODO: Call sync engine
	},
}

var (
	restoreFlag   bool
	listFlag      bool
	historyFlag   bool
	rollbackFlag  int
	addFlag       string
	removeFlag    string
	excludeFlag   string
	bsSetupFlag   bool
	bootstrapFlag string
	yesFlag       bool
	machineFlag   bool
)

func init() {

	rootCmd.Flags().BoolVarP(&restoreFlag, "restore", "r", false, "Restore dotfiles from remote repository")
	rootCmd.Flags().BoolVar(&listFlag, "list", false, "List tracked and available dotfiles")
	rootCmd.Flags().BoolVar(&historyFlag, "history", false, "Show sync history")
	rootCmd.Flags().IntVar(&rollbackFlag, "rollback", 0, "Rollback to n commits ago")
	rootCmd.Flags().StringVar(&addFlag, "add", "", "Add a dotfile to track")
	rootCmd.Flags().StringVar(&removeFlag, "remove", "", "Remove a dotfile from tracking")
	rootCmd.Flags().StringVar(&excludeFlag, "exclude", "", "Add exclude pattern and remove from repo")
	rootCmd.Flags().BoolVar(&bsSetupFlag, "bootstrap-setup", false, "Create bootstrap config interactively")
	rootCmd.Flags().StringVar(&bootstrapFlag, "bootstrap", "", "Run bootstrap from local config or clone provided repo URL")
	rootCmd.Flags().BoolVar(&yesFlag, "yes", false, "Skip per-step confirmations on bootstrap")
	rootCmd.Flags().BoolVarP(&machineFlag, "machine", "m", false, "Target machine-specific config instead of shared base")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
