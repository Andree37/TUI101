package app

import (
	"os"
	"strings"
	"tui101/git"
	"tui101/panes"
	"tui101/styles"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	panes         []panes.Pane
	activePane    int
	width         int
	height        int
	styles        *styles.Styles
	quitting      bool
	filterMode    bool
	filterText    string
	diffScrollPos int
	diffLines     []string
}

func NewModel() *Model {
	m := &Model{
		styles:     styles.NewStyles(),
		activePane: 1, // Start with files pane active
	}

	m.panes = []panes.Pane{
		panes.NewStatusPane(),
		panes.NewFilesPane(),
		panes.NewBranchesPane(),
		panes.NewCommitsPane(),
		panes.NewStashPane(),
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	for _, pane := range m.panes {
		cmds = append(cmds, pane.Init())
	}

	return tea.Batch(cmds...)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle global keybindings first
		cmd := m.handleKeyMsg(msg)
		if cmd != nil {
			return m, cmd
		}

		if m.activePane < len(m.panes) {
			updatedPane, cmd := m.panes[m.activePane].Update(msg)
			m.panes[m.activePane] = updatedPane
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	default:
		for i, pane := range m.panes {
			updatedPane, cmd := pane.Update(msg)
			m.panes[i] = updatedPane
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return tea.Quit

	case "tab":
		m.nextPane()
		return tea.Batch()
	case "shift+tab":
		m.prevPane()
		return tea.Batch()

	case "1":
		m.setActivePane(0)
		return tea.Batch()

	case "2":
		m.setActivePane(1)
		return tea.Batch()

	case "3":
		m.setActivePane(2)
		return tea.Batch()

	case "4":
		m.setActivePane(3)
		return tea.Batch()

	case "5":
		m.setActivePane(4)
		return tea.Batch()

	case "ctrl+r":
		return m.refreshAll()

	case "?":
		return tea.Batch()

	// j/k scroll diff when not in stash pane to avoid conflict with pane navigation
	case "j", "down":
		if m.activePane != 4 {
			m.scrollDiffDown()
			return tea.Batch()
		}
	case "k", "up":
		if m.activePane != 4 {
			m.scrollDiffUp()
			return tea.Batch()
		}
	}

	return nil
}

func (m *Model) scrollDiffDown() {
	if len(m.diffLines) > 0 {
		maxScroll := len(m.diffLines) - (m.height - 5)
		if maxScroll < 0 {
			maxScroll = 0
		}
		if m.diffScrollPos < maxScroll {
			m.diffScrollPos++
		}
	}
}

func (m *Model) scrollDiffUp() {
	if m.diffScrollPos > 0 {
		m.diffScrollPos--
	}
}

func (m *Model) getRealGitDiff(filepath string) string {
	if strings.HasSuffix(filepath, "/") || filepath == ".." {
		return "Directory selected - no diff to show"
	}

	if info, err := os.Stat(filepath); err == nil && info.IsDir() {
		return "Directory selected - no diff to show"
	}

	gitRepo := git.NewRepository(".")

	if m.activePane == 1 {
		return gitRepo.GetFileDiff(filepath)
	}

	if m.activePane == 3 {
		// TODO: show commit diff instead of file diff
		return gitRepo.GetFileDiff(filepath)
	}

	return ""
}

func (m *Model) nextPane() {
	m.setActivePane((m.activePane + 1) % len(m.panes))
}

func (m *Model) prevPane() {
	m.setActivePane((m.activePane - 1 + len(m.panes)) % len(m.panes))
}

func (m *Model) setActivePane(index int) {
	if index >= 0 && index < len(m.panes) {
		m.activePane = index
		for i, pane := range m.panes {
			pane.SetActive(i == index)
		}
	}
}

func (m *Model) refreshAll() tea.Cmd {
	var cmds []tea.Cmd
	for _, pane := range m.panes {
		if cmd := pane.Refresh(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.quitting {
		return "Goodbye! 👋\n"
	}

	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	leftPaneWidth := (m.width * 2) / 3
	rightPaneWidth := m.width - leftPaneWidth

	availableHeight := m.height - 1
	paneHeight := availableHeight / 4

	if m.activePane >= len(m.panes) {
		m.activePane = 0
	}

	for i, pane := range m.panes {
		pane.SetActive(i == m.activePane)
	}

	m.updateDiffContent()

	return m.renderLayout(leftPaneWidth, rightPaneWidth, paneHeight)
}

func (m *Model) updateDiffContent() {
	selectedItem := m.getSelectedItemForDiff()
	if selectedItem != "" {
		diffContent := m.getRealGitDiff(selectedItem)
		m.diffLines = strings.Split(diffContent, "\n")
	} else {
		m.diffLines = []string{"Select a file to see diff"}
	}
}

func (m *Model) getSelectedItemForDiff() string {
	if m.activePane >= len(m.panes) {
		return ""
	}

	activePane := m.panes[m.activePane]
	selectedItem := activePane.GetSelectedItem()
	if selectedItem == nil {
		return ""
	}

	return selectedItem.Value
}

func (m *Model) GetDiffLines() []string {
	return m.diffLines
}

func (m *Model) GetDiffScrollPos() int {
	return m.diffScrollPos
}

func (m *Model) GetActivePane() panes.Pane {
	if m.activePane < len(m.panes) {
		return m.panes[m.activePane]
	}
	return nil
}

func (m *Model) GetPanes() []panes.Pane {
	return m.panes
}

func (m *Model) GetDimensions() (int, int) {
	return m.width, m.height
}
