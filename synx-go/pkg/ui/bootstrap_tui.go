package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BootstrapResult holds the wizard output
type BootstrapResult struct {
	AurHelper       string
	Packages        string // space/newline separated
	Repos           string // one per line: url | dest | cmd
	DMName          string
	DMTheme         string
	DMThemeSource   string
	DotfilesRestore bool
	Commands        string // one per line
}

// wizard steps
const (
	bsStepAUR = iota
	bsStepPackages
	bsStepRepos
	bsStepDM
	bsStepDMTheme
	bsStepRestore
	bsStepCommands
	bsStepReview
	bsStepDone
)

var aurHelpers = []string{"paru", "yay", "pikaur", "skip"}
var dmOptions = []string{"sddm", "gdm", "ly", "greetd", "lightdm", "skip"}

type bootstrapModel struct {
	step     int
	cursor   int // for select lists
	input    textinput.Model
	viewport viewport.Model

	aurHelper       string
	packages        string
	repos           string
	dmName          string
	dmTheme         string
	dmThemeSource   string
	dotfilesRestore bool
	commands        string

	canceled bool
	ready    bool
	width    int
	height   int
}

func newBootstrapModel() bootstrapModel {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 60
	ti.PromptStyle = StyleCyan
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4"))
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(Cyan)

	return bootstrapModel{
		step:   bsStepAUR,
		cursor: 0,
		input:  ti,
	}
}

func (m bootstrapModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m bootstrapModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.canceled = true
			return m, tea.Quit

		case "esc":
			if m.step > bsStepAUR {
				m.step--
				m.cursor = 0
				m.input.SetValue("")
				m.input.Focus()
			}
			return m, nil

		case "up", "k":
			if m.isSelectStep() && m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.isSelectStep() {
				max := len(m.selectOptions()) - 1
				if m.cursor < max {
					m.cursor++
				}
			}

		case "enter":
			return m.handleEnter()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.viewport = viewport.New(msg.Width-8, msg.Height-8)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 8
			m.viewport.Height = msg.Height - 8
		}
	}

	if m.isTextStep() {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m bootstrapModel) isSelectStep() bool {
	return m.step == bsStepAUR || m.step == bsStepDM || m.step == bsStepRestore
}

func (m bootstrapModel) isTextStep() bool {
	return m.step == bsStepPackages || m.step == bsStepRepos ||
		m.step == bsStepDMTheme || m.step == bsStepCommands
}

func (m bootstrapModel) selectOptions() []string {
	switch m.step {
	case bsStepAUR:
		return aurHelpers
	case bsStepDM:
		return dmOptions
	case bsStepRestore:
		return []string{"yes", "no"}
	}
	return nil
}

func (m bootstrapModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case bsStepAUR:
		choice := aurHelpers[m.cursor]
		if choice == "skip" {
			m.aurHelper = ""
		} else {
			m.aurHelper = choice
		}
		m.step = bsStepPackages
		m.cursor = 0
		m.input.SetValue("")
		m.input.Placeholder = "e.g. firefox neovim kitty foot waybar"
		m.input.Focus()

	case bsStepPackages:
		m.packages = m.input.Value()
		m.step = bsStepRepos
		m.input.SetValue("")
		m.input.Placeholder = "url | ~/dest | ./install.sh"
		m.input.Focus()

	case bsStepRepos:
		m.repos = m.input.Value()
		m.step = bsStepDM
		m.cursor = 0

	case bsStepDM:
		choice := dmOptions[m.cursor]
		if choice == "skip" {
			m.dmName = ""
			m.step = bsStepRestore
		} else {
			m.dmName = choice
			m.step = bsStepDMTheme
			m.input.SetValue("")
			m.input.Placeholder = "e.g. sddm-sugar-dark (leave empty to skip)"
			m.input.Focus()
		}
		m.cursor = 0

	case bsStepDMTheme:
		m.dmTheme = m.input.Value()
		m.step = bsStepRestore
		m.cursor = 0

	case bsStepRestore:
		m.dotfilesRestore = m.cursor == 0 // "yes" is index 0
		m.step = bsStepCommands
		m.input.SetValue("")
		m.input.Placeholder = "e.g. chsh -s /usr/bin/fish"
		m.input.Focus()

	case bsStepCommands:
		m.commands = m.input.Value()
		m.step = bsStepReview

	case bsStepReview:
		m.step = bsStepDone
		return m, tea.Quit
	}

	return m, nil
}

func (m bootstrapModel) View() string {
	if m.canceled || m.step == bsStepDone {
		return ""
	}

	var b strings.Builder

	header := BoxStyle.Render(fmt.Sprintf("🔧 %s - Bootstrap Setup", StyleBold.Render("SYNX")))
	b.WriteString(header + "\n\n")

	steps := []string{"AUR", "Packages", "Repos", "Display", "Theme", "Restore", "Commands", "Review"}
	var progress strings.Builder
	for i, s := range steps {
		if i == m.step {
			progress.WriteString(TUIActiveStyle.Render(fmt.Sprintf(" ● %s ", s)))
		} else if i < m.step {
			progress.WriteString(TUITrackedStyle.Render(fmt.Sprintf(" ✓ %s ", s)))
		} else {
			progress.WriteString(TUIMutedStyle.Render(fmt.Sprintf(" ○ %s ", s)))
		}
		if i < len(steps)-1 {
			progress.WriteString(TUIMutedStyle.Render("─"))
		}
	}
	b.WriteString(progress.String() + "\n\n")

	switch m.step {
	case bsStepAUR:
		b.WriteString(TUITitleStyle.Render("1. Select AUR Helper") + "\n\n")
		for i, opt := range aurHelpers {
			if i == m.cursor {
				b.WriteString(StyleCyan.Render("  ▸ "+opt) + "\n")
			} else {
				b.WriteString(TUIMutedStyle.Render("    "+opt) + "\n")
			}
		}

	case bsStepPackages:
		b.WriteString(TUITitleStyle.Render("2. Packages to Install") + "\n")
		b.WriteString(TUIMutedStyle.Render("   Space-separated list of packages") + "\n\n")
		b.WriteString("  " + m.input.View() + "\n")

	case bsStepRepos:
		b.WriteString(TUITitleStyle.Render("3. Git Repositories") + "\n")
		b.WriteString(TUIMutedStyle.Render("   Format: https://github.com/.. | ~/dest | ./install.sh") + "\n\n")
		b.WriteString("  " + m.input.View() + "\n")

	case bsStepDM:
		b.WriteString(TUITitleStyle.Render("4. Display Manager") + "\n\n")
		for i, opt := range dmOptions {
			if i == m.cursor {
				b.WriteString(StyleCyan.Render("  ▸ "+opt) + "\n")
			} else {
				b.WriteString(TUIMutedStyle.Render("    "+opt) + "\n")
			}
		}

	case bsStepDMTheme:
		b.WriteString(TUITitleStyle.Render("5. DM Theme Package") + "\n")
		b.WriteString(TUIMutedStyle.Render("   Leave empty to skip") + "\n\n")
		b.WriteString("  " + m.input.View() + "\n")

	case bsStepRestore:
		b.WriteString(TUITitleStyle.Render("6. Restore Dotfiles After Bootstrap?") + "\n\n")
		opts := []string{"yes", "no"}
		for i, opt := range opts {
			if i == m.cursor {
				b.WriteString(StyleCyan.Render("  ▸ "+opt) + "\n")
			} else {
				b.WriteString(TUIMutedStyle.Render("    "+opt) + "\n")
			}
		}

	case bsStepCommands:
		b.WriteString(TUITitleStyle.Render("7. Custom Post-Install Commands") + "\n")
		b.WriteString(TUIMutedStyle.Render("   e.g. chsh -s /usr/bin/fish") + "\n\n")
		b.WriteString("  " + m.input.View() + "\n")

	case bsStepReview:
		b.WriteString(TUITitleStyle.Render("Review Configuration") + "\n\n")
		b.WriteString(m.renderSummary())
		b.WriteString("\n" + TUIActiveStyle.Render("  Press [Enter] to save  •  [Esc] to go back") + "\n")
	}

	b.WriteString("\n" + TUIMutedStyle.Render("  [Enter] next  •  [Esc] back  •  [Ctrl+C] cancel"))

	return b.String()
}

func (m bootstrapModel) renderSummary() string {
	var b strings.Builder
	row := func(label, value string) {
		if value == "" {
			value = TUIMutedStyle.Render("(skipped)")
		}
		b.WriteString(fmt.Sprintf("  %s  %s\n", StyleCyan.Render(fmt.Sprintf("%-20s", label)), value))
	}

	row("AUR Helper:", m.aurHelper)
	if m.packages != "" {
		pkgs := strings.Fields(m.packages)
		row("Packages:", fmt.Sprintf("%d package(s)", len(pkgs)))
	} else {
		row("Packages:", "")
	}
	row("Git Repos:", m.repos)
	row("Display Manager:", m.dmName)
	row("DM Theme:", m.dmTheme)
	if m.dotfilesRestore {
		row("Restore Dotfiles:", TUITrackedStyle.Render("yes"))
	} else {
		row("Restore Dotfiles:", "no")
	}
	row("Commands:", m.commands)
	return b.String()
}

// RunBootstrapTUI launches the interactive bootstrap setup wizard.
// Returns the result or nil if canceled.
func RunBootstrapTUI() (*BootstrapResult, error) {
	m := newBootstrapModel()

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithInput(os.Stdin), tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm, ok := finalModel.(bootstrapModel)
	if !ok || fm.canceled || fm.step != bsStepDone {
		return nil, nil
	}

	return &BootstrapResult{
		AurHelper:       fm.aurHelper,
		Packages:        fm.packages,
		Repos:           fm.repos,
		DMName:          fm.dmName,
		DMTheme:         fm.dmTheme,
		DMThemeSource:   fm.dmThemeSource,
		DotfilesRestore: fm.dotfilesRestore,
		Commands:        fm.commands,
	}, nil
}
