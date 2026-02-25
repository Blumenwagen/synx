package ui

import (
	"fmt"

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

	// Box drawing
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
