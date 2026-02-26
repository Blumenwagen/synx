package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Standard theme colors replicating the Fish script logic
var (
	Cyan    = lipgloss.Color("#74c7ec") // sapphire
	Green   = lipgloss.Color("#a6e3a1") // green
	Yellow  = lipgloss.Color("#f9e2af") // yellow
	Red     = lipgloss.Color("#f38ba8") // red
	Blue    = lipgloss.Color("#89b4fa") // blue
	Magenta = lipgloss.Color("#cba6f7") // mauve
	Dim     = lipgloss.Color("#6c7086") // overlay0

	StyleBold    = lipgloss.NewStyle().Bold(true)
	StyleDim     = lipgloss.NewStyle().Foreground(Dim)
	StyleCyan    = lipgloss.NewStyle().Foreground(Cyan)
	StyleGreen   = lipgloss.NewStyle().Foreground(Green)
	StyleYellow  = lipgloss.NewStyle().Foreground(Yellow)
	StyleRed     = lipgloss.NewStyle().Foreground(Red)
	StyleBlue    = lipgloss.NewStyle().Foreground(Blue)
	StyleMagenta = lipgloss.NewStyle().Foreground(Magenta)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Cyan).
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4")).
			Bold(true)
)

func PrintHeader(icon string, title string) {
	fmt.Println()
	content := fmt.Sprintf("%s %s - %s", icon, StyleBold.Render("SYNX"), title)
	fmt.Println(BoxStyle.Render(content))
	fmt.Println()
}

func Step(msg string) {
	fmt.Printf("%s %s\n", StyleBlue.Render("→"), StyleBold.Render(msg))
}

func SubStep(msg string) {
	fmt.Printf("  %s %s\n", StyleCyan.Render("▸"), msg)
}

func Success(msg string) {
	fmt.Printf("  %s %s\n", StyleGreen.Render("✓"), msg)
}

func Warn(msg string) {
	fmt.Printf("  %s %s\n", StyleYellow.Render("⚠"), msg)
}

func Error(msg string) {
	fmt.Printf("  %s %s\n", StyleRed.Render("✗"), msg)
}

func Info(msg string) {
	fmt.Printf("  %s %s\n", StyleBlue.Render("ℹ"), msg)
}

func Bullet(msg string) {
	fmt.Printf("    %s %s\n", StyleDim.Render("•"), msg)
}

func Detail(msg string) {
	fmt.Printf("    %s %s\n", StyleDim.Render("→"), StyleDim.Render(msg))
}

// Shared TUI Styles (for Bubbletea screens)

var (
	TUIDocStyle  = lipgloss.NewStyle().Margin(1, 2)
	TUIPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1).
			BorderForeground(Blue)
	TUIDetailPaneStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Padding(1, 2).
				BorderForeground(Blue)

	TUITitleStyle   = lipgloss.NewStyle().Foreground(Magenta).Bold(true)
	TUIHeadingStyle = lipgloss.NewStyle().Foreground(Magenta).Bold(true).Underline(true)
	TUITrackedStyle = lipgloss.NewStyle().Foreground(Green)
	TUIMutedStyle   = lipgloss.NewStyle().Foreground(Dim)
	TUIMachineStyle = lipgloss.NewStyle().Foreground(Magenta).Bold(true)
	TUIActiveStyle  = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
)

// ThemedDelegate returns a Bubbles list delegate styled with Catppuccin Mocha colors.
func ThemedDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(Cyan).
		BorderForeground(Cyan)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(Blue).
		BorderForeground(Cyan)

	d.Styles.NormalTitle = d.Styles.NormalTitle.
		Foreground(lipgloss.Color("#cdd6f4")) // Catppuccin text
	d.Styles.NormalDesc = d.Styles.NormalDesc.
		Foreground(Dim)

	return d
}
