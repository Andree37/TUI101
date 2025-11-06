package panes

import (
	"tui101/git"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusPane represents the status pane showing current branch and repo info
type StatusPane struct {
	BasePaneModel
	gitRepo *git.Repository
}

// NewStatusPane creates a new status pane
func NewStatusPane() *StatusPane {
	base := NewBasePaneModel("Status", StatusPaneType, "status")

	pane := &StatusPane{
		BasePaneModel: base,
		gitRepo:       git.NewRepository("."),
	}

	pane.loadStatusInfo()
	return pane
}

// Init initializes the status pane
func (s *StatusPane) Init() tea.Cmd {
	return s.Refresh()
}

// Update handles updates for the status pane
func (s *StatusPane) Update(msg tea.Msg) (Pane, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !s.IsActive() {
			return s, nil
		}

		switch msg.String() {
		case "r":
			return s, s.Refresh()
		case "enter":
			// Handle status pane actions (e.g., checkout branch)
			return s, s.HandleAction("toggle")
		}

	case git.StatusUpdateMsg:
		s.updateFromGitStatus(msg)
		return s, nil
	}

	return s, nil
}

// View renders the status pane
func (s *StatusPane) View() string {
	if s.IsLoading() {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#74B9FF")).
			Render("Loading status...")
	}

	if len(s.items) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#74B9FF")).
			Render("No repository information")
	}

	var lines []string
	for _, item := range s.items {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#DDD6FE"))

		switch item.Type {
		case "branch":
			style = style.Foreground(lipgloss.Color("#04B575")).Bold(true)
		case "status":
			style = style.Foreground(lipgloss.Color("#FFEAA7"))
		case "upstream":
			style = style.Foreground(lipgloss.Color("#74B9FF"))
		}

		lines = append(lines, style.Render(item.Display))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// Refresh refreshes the status pane data
func (s *StatusPane) Refresh() tea.Cmd {
	s.SetLoading(true)
	return func() tea.Msg {
		status := s.gitRepo.GetStatus()
		return git.StatusUpdateMsg{Status: status}
	}
}

// HandleAction handles pane-specific actions
func (s *StatusPane) HandleAction(action string) tea.Cmd {
	switch action {
	case "toggle":
		// Toggle between showing detailed status or simple status
		return nil
	case "fetch":
		return func() tea.Msg {
			s.gitRepo.Fetch()
			return git.StatusUpdateMsg{Status: s.gitRepo.GetStatus()}
		}
	}
	return nil
}

// GetAvailableActions returns available actions for this pane
func (s *StatusPane) GetAvailableActions() []string {
	return []string{"refresh", "fetch", "toggle"}
}

// loadStatusInfo loads initial status information
func (s *StatusPane) loadStatusInfo() {
	s.Clear()

	// Add current branch info
	branch := s.gitRepo.GetCurrentBranch()
	if branch != "" {
		s.AddItem(PaneItem{
			Display: "TUI101 → " + branch,
			Value:   branch,
			Type:    "branch",
			Icon:    "→",
		})
	} else {
		s.AddItem(PaneItem{
			Display: "TUI101 (detached HEAD)",
			Value:   "HEAD",
			Type:    "branch",
			Icon:    "⚠",
		})
	}

	// Add repository status
	status := s.gitRepo.GetStatus()
	if status.HasChanges() {
		s.AddItem(PaneItem{
			Display: "Working tree has changes",
			Value:   "dirty",
			Type:    "status",
			Icon:    "●",
		})
	} else {
		s.AddItem(PaneItem{
			Display: "Working tree clean",
			Value:   "clean",
			Type:    "status",
			Icon:    "○",
		})
	}

	// Add upstream info
	upstream := s.gitRepo.GetUpstreamInfo()
	if upstream != "" {
		s.AddItem(PaneItem{
			Display: upstream,
			Value:   upstream,
			Type:    "upstream",
			Icon:    "↑",
		})
	}
}

// updateFromGitStatus updates the pane content from git status message
func (s *StatusPane) updateFromGitStatus(msg git.StatusUpdateMsg) {
	s.SetLoading(false)
	s.Clear()

	// Update branch info
	if msg.Status.Branch != "" {
		s.AddItem(PaneItem{
			Display: "TUI101 → " + msg.Status.Branch,
			Value:   msg.Status.Branch,
			Type:    "branch",
			Icon:    "→",
		})
	}

	// Update status info
	if msg.Status.HasChanges() {
		changesText := "Working tree has changes"
		if msg.Status.ModifiedFiles > 0 {
			changesText += " (" + string(rune(msg.Status.ModifiedFiles)) + " modified)"
		}
		s.AddItem(PaneItem{
			Display: changesText,
			Value:   "dirty",
			Type:    "status",
			Icon:    "●",
		})
	} else {
		s.AddItem(PaneItem{
			Display: "Working tree clean",
			Value:   "clean",
			Type:    "status",
			Icon:    "○",
		})
	}

	// Update upstream info
	if msg.Status.Upstream != "" {
		upstreamText := msg.Status.Upstream
		if msg.Status.AheadBy > 0 {
			upstreamText += " (ahead " + string(rune(msg.Status.AheadBy)) + ")"
		}
		if msg.Status.BehindBy > 0 {
			upstreamText += " (behind " + string(rune(msg.Status.BehindBy)) + ")"
		}

		s.AddItem(PaneItem{
			Display: upstreamText,
			Value:   msg.Status.Upstream,
			Type:    "upstream",
			Icon:    "↑",
		})
	}
}
