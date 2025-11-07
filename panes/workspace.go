package panes

import (
	"fmt"
	"time"
	"tui101/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StatusPane struct {
	BasePaneModel
	st *styles.Styles
}

type WorkspaceUpdateMsg struct {
	Info WorkspaceInfo
}

type WorkspaceInfo struct {
	Name       string
	VersionSet string
	UpdatedAt  time.Time
}

func NewStatusPane() *StatusPane {
	base := NewBasePaneModel("Workspace", StatusPaneType, "workspace")

	pane := &StatusPane{
		BasePaneModel: base,
		st:            styles.NewStyles(),
	}

	pane.loadWorkspaceInfo()
	return pane
}

func (s *StatusPane) Init() tea.Cmd {
	return s.Refresh()
}

func (s *StatusPane) Update(msg tea.Msg) (Pane, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !s.IsActive() {
			return s, nil
		}

		switch msg.String() {
		case "j", "down":
			s.MoveDown()
		case "k", "up":
			s.MoveUp()
		case "r":
			return s, s.Refresh()
		}

	case WorkspaceUpdateMsg:
		s.updateFromWorkspaceInfo(msg)
		return s, nil
	}

	return s, nil
}

func (s *StatusPane) View() string {
	if s.IsLoading() {
		return s.st.LoadingText.Render("Loading workspace...")
	}

	if len(s.items) == 0 {
		return s.st.InfoText.Render("No workspace information")
	}

	var lines []string

	// Add a nice header
	lines = append(lines, s.st.Dimmed.Render("━━━━━━━━━━━━━━━━━━━━━━━━"))

	for i, item := range s.items {
		isSelected := i == s.GetSelectedIndex()

		var line string
		var style lipgloss.Style

		// Get base style based on item type
		switch item.Type {
		case "name":
			style = s.st.WorkspaceName
		case "version":
			style = s.st.WorkspaceVersion
		case "metadata":
			style = s.st.WorkspaceMetadata
		default:
			style = s.st.UnselectedItem
		}

		// Override with selection style if active and selected
		if isSelected && s.IsActive() {
			style = s.st.SelectedItem
			line = s.st.RenderCursor(true) + item.Display
		} else {
			line = s.st.RenderCursor(false) + item.Display
		}

		// Add icons based on type
		switch item.Type {
		case "name":
			if !isSelected || !s.IsActive() {
				line = "  " + item.Display
			} else {
				line = s.st.RenderCursor(true) + item.Display
			}
		case "version":
			if !isSelected || !s.IsActive() {
				line = "  " + item.Display
			} else {
				line = s.st.RenderCursor(true) + item.Display
			}
		case "metadata":
			if !isSelected || !s.IsActive() {
				line = "  " + item.Display
			} else {
				line = s.st.RenderCursor(true) + item.Display
			}
		}

		lines = append(lines, style.Render(line))
	}

	// Add a footer separator
	lines = append(lines, "")
	lines = append(lines, s.st.Dimmed.Render("━━━━━━━━━━━━━━━━━━━━━━━━"))

	// Add help text if active
	if s.IsActive() {
		lines = append(lines, "")
		lines = append(lines, s.st.Dimmed.Render("↑↓: Navigate  r: Refresh"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (s *StatusPane) Refresh() tea.Cmd {
	s.SetLoading(true)
	return func() tea.Msg {
		// Simulate some loading time
		time.Sleep(500 * time.Millisecond)
		info := s.gatherWorkspaceInfo()
		return WorkspaceUpdateMsg{Info: info}
	}
}

func (s *StatusPane) HandleAction(action string) tea.Cmd {
	switch action {
	case "refresh":
		return s.Refresh()
	}
	return nil
}

func (s *StatusPane) GetAvailableActions() []string {
	return []string{"refresh"}
}

func (s *StatusPane) loadWorkspaceInfo() {
	s.Clear()

	s.AddItem(PaneItem{
		Display: "JoeWorkspace",
		Value:   "JoeWorkspace",
		Type:    "name",
	})

	s.AddItem(PaneItem{
		Display: "Version Set: v1.0.0",
		Value:   "v1.0.0",
		Type:    "version",
	})

	s.AddItem(PaneItem{
		Display: fmt.Sprintf("Updated: %s", time.Now().Format("2006-01-02 15:04")),
		Value:   time.Now().Format(time.RFC3339),
		Type:    "metadata",
	})
}

func (s *StatusPane) gatherWorkspaceInfo() WorkspaceInfo {
	return WorkspaceInfo{
		Name:       "JoeWorkspace",
		VersionSet: "v1.0.0",
		UpdatedAt:  time.Now(),
	}
}

func (s *StatusPane) updateFromWorkspaceInfo(msg WorkspaceUpdateMsg) {
	s.SetLoading(false)
	s.Clear()

	info := msg.Info

	s.AddItem(PaneItem{
		Display: info.Name,
		Value:   info.Name,
		Type:    "name",
	})

	s.AddItem(PaneItem{
		Display: fmt.Sprintf("Version Set: %s", info.VersionSet),
		Value:   info.VersionSet,
		Type:    "version",
	})

	s.AddItem(PaneItem{
		Display: fmt.Sprintf("Updated: %s", info.UpdatedAt.Format("2006-01-02 15:04")),
		Value:   info.UpdatedAt.Format(time.RFC3339),
		Type:    "metadata",
	})
}
