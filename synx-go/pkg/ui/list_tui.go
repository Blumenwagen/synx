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

var (
	trackedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	untrackedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	machineStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	detailTitle    = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Underline(true)
)

type dotfileItem struct {
	name          string
	isTracked     bool
	isMachineOnly bool
	isSymlink     bool
	exists        bool
}

func (i dotfileItem) Title() string {
	if i.isTracked {
		if i.isMachineOnly {
			return trackedStyle.Render("✓ ") + machineStyle.Render(i.name)
		}
		return trackedStyle.Render("✓ " + i.name)
	}
	return untrackedStyle.Render("• " + i.name)
}

func (i dotfileItem) Description() string {
	if !i.exists {
		return "Missing in ~/.config"
	}
	if i.isSymlink {
		return "Symlinked"
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
		h, v := docStyle.GetFrameSize()
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

	// Update Details pane
	if i, ok := m.list.SelectedItem().(dotfileItem); ok {
		var b strings.Builder

		b.WriteString(detailTitle.Render(i.name) + "\n\n")

		if i.isTracked {
			b.WriteString(trackedStyle.Render("Status: Tracked by Synx") + "\n")
			if i.isMachineOnly {
				b.WriteString(machineStyle.Render(fmt.Sprintf("Scope: %s overrides only", m.cfg.Hostname)) + "\n")
			} else {
				b.WriteString("Scope: Base configuration\n")
			}
		} else {
			b.WriteString(untrackedStyle.Render("Status: Untracked") + "\n")
			b.WriteString("\nPress [Enter] to add to synx.\n")
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

	left := listPaneStyle.Render(m.list.View())
	right := diffPaneStyle.Render(m.viewport.View())

	return docStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, left, right))
}

// RunListTUI launches the interactive dashboard for browsing and adding dotfiles.
// Returns the name of an untracked dotfile if the user chose to add one.
func RunListTUI(cfg *config.ConfigManager) (string, error) {
	home, _ := os.UserHomeDir()
	baseConfigDir := filepath.Join(home, ".config")

	// 1. Find all active targets (Base + Machine)
	targetMap := make(map[string]bool)
	machineOnlyMap := make(map[string]bool)

	// In ConfigManager, `Targets` is already the unified list if UsingMachineTargets is true
	for _, t := range cfg.Targets {
		targetMap[t] = true
	}

	if cfg.UsingMachineTargets {
		// Figure out which ones are ONLY machine specific (not in base config file)
		baseTargetsRaw, _ := os.ReadFile(filepath.Join(home, ".synx", "targets.conf"))
		baseLines := strings.Split(string(baseTargetsRaw), "\n")
		baseMap := make(map[string]bool)
		for _, l := range baseLines {
			l = strings.TrimSpace(l)
			if l != "" && !strings.HasPrefix(l, "#") {
				baseMap[l] = true
			}
		}

		for _, t := range cfg.Targets {
			if !baseMap[t] {
				machineOnlyMap[t] = true
			}
		}
	}

	// 2. Read ~/.config
	entries, _ := os.ReadDir(baseConfigDir)

	// Ensure we show tracked items even if they are currently missing from ~/.config
	seenDirs := make(map[string]bool)
	var items []list.Item

	// First add the tracked ones
	for _, t := range cfg.Targets {
		seenDirs[t] = true
		info, err := os.Lstat(filepath.Join(baseConfigDir, t))
		items = append(items, dotfileItem{
			name:          t,
			isTracked:     true,
			isMachineOnly: machineOnlyMap[t],
			exists:        err == nil,
			isSymlink:     err == nil && info.Mode()&os.ModeSymlink != 0,
		})
	}

	// Then add the untracked ones currently present
	for _, e := range entries {
		name := e.Name()
		if !seenDirs[name] && e.IsDir() && !strings.HasPrefix(name, ".") {
			info, _ := e.Info()
			items = append(items, dotfileItem{
				name:          name,
				isTracked:     false,
				isMachineOnly: false,
				exists:        true,
				isSymlink:     info.Mode()&os.ModeSymlink != 0,
			})
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Synx Tracked & Available Dotfiles"

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
