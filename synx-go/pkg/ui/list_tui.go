package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Blumenwagen/synx/pkg/config"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type dotfileItem struct {
	name          string
	isTracked     bool
	isProfileOnly bool
	isSymlink     bool
	exists        bool
}

func (i dotfileItem) Title() string {
	if i.isTracked {
		if i.isProfileOnly {
			return TUITrackedStyle.Render("✓ ") + TUIMachineStyle.Render(i.name)
		}
		return TUITrackedStyle.Render("✓ " + i.name)
	}
	return TUIMutedStyle.Render("• " + i.name)
}

func (i dotfileItem) Description() string {
	if !i.exists {
		return StyleYellow.Render("Missing in ~/.config")
	}
	if i.isSymlink {
		return StyleCyan.Render("Symlinked")
	}
	return "Local Directory"
}

func (i dotfileItem) FilterValue() string { return i.name }

type listModel struct {
	list        list.Model
	viewport    viewport.Model
	cfg         *config.ConfigManager
	selectedAdd string
	canceled    bool
	ready       bool
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q", "esc":
			m.canceled = true
			return m, tea.Quit

		case "enter":
			if i, ok := m.list.SelectedItem().(dotfileItem); ok {
				if !i.isTracked {
					m.selectedAdd = i.name
					return m, tea.Quit
				}
			}
		}

	case tea.WindowSizeMsg:
		h, v := TUIDocStyle.GetFrameSize()
		m.list.SetSize(msg.Width/2-h-2, msg.Height-v-2)

		if !m.ready {
			m.viewport = viewport.New(msg.Width/2-h-2, msg.Height-v-4)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width/2 - h - 2
			m.viewport.Height = msg.Height - v - 4
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	if i, ok := m.list.SelectedItem().(dotfileItem); ok {
		var b strings.Builder

		b.WriteString(TUIHeadingStyle.Render(i.name) + "\n\n")

		if i.isTracked {
			b.WriteString(TUITrackedStyle.Render("Status: Tracked by Synx") + "\n")
			if i.isProfileOnly {
				b.WriteString(TUIMachineStyle.Render(fmt.Sprintf("Scope: %s profile overrides", m.cfg.ActiveProfile)) + "\n")
			} else {
				b.WriteString("Scope: Base configuration\n")
			}
		} else {
			b.WriteString(TUIMutedStyle.Render("Status: Untracked") + "\n")
			b.WriteString("\n" + TUIActiveStyle.Render("Press [Enter] to add to synx base targets.") + "\n")
		}

		b.WriteString("\n")
		if !i.exists {
			b.WriteString(StyleYellow.Render("Location: Missing from ~/.config"))
		} else if i.isSymlink {
			b.WriteString(StyleCyan.Render("Location: Synx Symlink active"))
		} else {
			b.WriteString("Location: Standard local directory")
		}

		m.viewport.SetContent(b.String())
	}

	return m, tea.Batch(cmds...)
}

func (m listModel) View() string {
	if !m.ready {
		return "Initializing UI..."
	}
	if m.canceled {
		return ""
	}

	left := TUIPaneStyle.Render(m.list.View())
	right := TUIDetailPaneStyle.Render(m.viewport.View())

	return TUIDocStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, left, right))
}

// RunListTUI launches the interactive dashboard for browsing and adding dotfiles.
// Returns the name of an untracked dotfile if the user chose to add one.
func RunListTUI(cfg *config.ConfigManager) (string, error) {
	home, _ := os.UserHomeDir()
	baseConfigDir := filepath.Join(home, ".config")

	targetMap := make(map[string]bool)
	profileOnlyMap := make(map[string]bool)

	for _, t := range cfg.Targets {
		targetMap[t] = true
	}

	for _, pt := range cfg.ProfileTargets {
		if !targetMap[pt] {
			profileOnlyMap[pt] = true
		}
		targetMap[pt] = true // Include profile targets in the main list
	}

	entries, _ := os.ReadDir(baseConfigDir)

	seenDirs := make(map[string]bool)
	var items []list.Item

	for t := range targetMap {
		seenDirs[t] = true
		info, err := os.Lstat(filepath.Join(baseConfigDir, t))
		items = append(items, dotfileItem{
			name:          t,
			isTracked:     true,
			isProfileOnly: profileOnlyMap[t],
			exists:        err == nil,
			isSymlink:     err == nil && info.Mode()&os.ModeSymlink != 0,
		})
	}

	for _, e := range entries {
		name := e.Name()
		if !seenDirs[name] && e.IsDir() && !strings.HasPrefix(name, ".") {
			info, _ := e.Info()
			items = append(items, dotfileItem{
				name:          name,
				isTracked:     false,
				isProfileOnly: false,
				exists:        true,
				isSymlink:     info.Mode()&os.ModeSymlink != 0,
			})
		}
	}

	delegate := ThemedDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Synx Tracked & Available Dotfiles"
	l.Styles.Title = TUITitleStyle

	m := listModel{
		list: l,
		cfg:  cfg,
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithInput(os.Stdin), tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	fm, ok := finalModel.(listModel)
	if !ok || fm.canceled {
		return "", nil
	}

	return fm.selectedAdd, nil
}
