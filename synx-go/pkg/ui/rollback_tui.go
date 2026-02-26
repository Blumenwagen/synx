package ui

import (
	"fmt"

	"github.com/Blumenwagen/synx/pkg/git"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle      = lipgloss.NewStyle().Margin(1, 2)
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	diffPaneStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).BorderForeground(lipgloss.Color("63"))
	listPaneStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).BorderForeground(lipgloss.Color("63"))
)

type commitItem struct {
	commit git.Commit
	index  int // steps back (1-indexed from HEAD)
}

func (i commitItem) Title() string { return fmt.Sprintf("%s - %s", i.commit.Hash, i.commit.Message) }
func (i commitItem) Description() string {
	return fmt.Sprintf("%s (%s)", i.commit.Author, i.commit.Date)
}
func (i commitItem) FilterValue() string { return i.commit.Message }

type rollbackModel struct {
	list        list.Model
	viewport    viewport.Model
	gitMgr      *git.GitManager
	commits     []git.Commit
	selectedIdx int // Will be set on confirm
	canceled    bool
	ready       bool
	detailCache map[string]string // Cache diff stats to avoid repeated git calls
}

func (m rollbackModel) Init() tea.Cmd {
	return nil
}

func (m rollbackModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q", "esc":
			m.canceled = true
			return m, tea.Quit

		case "enter":
			if i, ok := m.list.SelectedItem().(commitItem); ok {
				m.selectedIdx = i.index
				return m, tea.Quit
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

	// Update list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	// Update viewport content based on current list selection
	if i, ok := m.list.SelectedItem().(commitItem); ok {
		hash := i.commit.Hash
		stat, cached := m.detailCache[hash]
		if !cached {
			s, err := m.gitMgr.ShowStat(hash)
			if err != nil {
				stat = "Error fetching diff: " + err.Error()
			} else {
				stat = s
			}
			m.detailCache[hash] = stat
		}

		content := titleStyle.Render(fmt.Sprintf("Files changed in %s", hash)) + "\n\n" + stat
		m.viewport.SetContent(content)
	}

	return m, tea.Batch(cmds...)
}

func (m rollbackModel) View() string {
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

// RunRollbackTUI launches the interactive commit selector.
// Returns the number of commits to rollback (steps), or 0 if canceled.
func RunRollbackTUI(gitMgr *git.GitManager) (int, error) {
	// Fetch last 50 commits
	commits, err := gitMgr.LogStructured(50)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch commits: %w", err)
	}

	if len(commits) == 0 {
		return 0, fmt.Errorf("no commits found in dotfiles repository")
	}

	items := make([]list.Item, len(commits))
	for i, c := range commits {
		items[i] = commitItem{commit: c, index: i + 1}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Commit to Rollback To"

	m := rollbackModel{
		list:        l,
		gitMgr:      gitMgr,
		commits:     commits,
		detailCache: make(map[string]string),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return 0, err
	}

	fm, ok := finalModel.(rollbackModel)
	if !ok || fm.canceled {
		return 0, nil
	}

	return fm.selectedIdx, nil
}
