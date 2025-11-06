package panes

import (
	"fmt"
	"strings"
	"tui101/git"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StashPane represents the stash pane showing git stashes
type StashPane struct {
	BasePaneModel
	gitRepo *git.Repository
}

// NewStashPane creates a new stash pane
func NewStashPane() *StashPane {
	base := NewBasePaneModel("Stash", StashPaneType, "stash")

	pane := &StashPane{
		BasePaneModel: base,
		gitRepo:       git.NewRepository("."),
	}

	pane.loadStashes()
	return pane
}

// Init initializes the stash pane
func (s *StashPane) Init() tea.Cmd {
	return s.Refresh()
}

// Update handles updates for the stash pane
func (s *StashPane) Update(msg tea.Msg) (Pane, tea.Cmd) {
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
		case "g":
			s.MoveToTop()
		case "G":
			s.MoveToBottom()
		case "enter":
			return s, s.HandleAction("apply_stash")
		case "p":
			return s, s.HandleAction("pop_stash")
		case "d":
			return s, s.HandleAction("drop_stash")
		case "s":
			return s, s.HandleAction("create_stash")
		case "S":
			return s, s.HandleAction("create_stash_include_untracked")
		case "r":
			return s, s.Refresh()
		case "v":
			return s, s.HandleAction("show_stash")
		case "D":
			return s, s.HandleAction("clear_all_stashes")
		case "b":
			return s, s.HandleAction("create_branch_from_stash")
		}

	case git.StashUpdateMsg:
		s.updateFromStashMsg(msg)
		return s, nil
	}

	return s, nil
}

// View renders the stash pane
func (s *StashPane) View() string {
	if s.IsLoading() {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#74B9FF")).
			Render("Loading stashes...")
	}

	if len(s.items) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#74B9FF")).
			Render("No stashed changes")
	}

	var lines []string
	visibleItems := s.GetVisibleItems()

	for i, item := range visibleItems {
		actualIndex := s.GetScrollOffset() + i
		isSelected := actualIndex == s.GetSelectedIndex()

		line := s.formatStashItem(item, isSelected)
		lines = append(lines, line)
	}

	// Add scroll indicators if needed
	if s.GetScrollOffset() > 0 {
		lines = append([]string{"  ↑ more stashes above"}, lines...)
	}
	if s.GetScrollOffset()+len(visibleItems) < len(s.items) {
		lines = append(lines, "  ↓ more stashes below")
	}

	// Add footer with count
	if len(s.items) > 0 {
		footer := s.getFooter()
		lines = append(lines, "", footer)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// formatStashItem formats a single stash item for display
func (s *StashPane) formatStashItem(item PaneItem, isSelected bool) string {
	var parts []string

	// Add selection indicator
	if isSelected {
		parts = append(parts, "▶")
	} else {
		parts = append(parts, " ")
	}

	// Add stash icon
	if item.Icon != "" {
		parts = append(parts, item.Icon)
	}

	// Parse stash information
	// Expected format: "stash@{n}: On branch: message"
	stashInfo := item.Display
	if strings.Contains(stashInfo, ":") {
		stashParts := strings.SplitN(stashInfo, ":", 3)
		if len(stashParts) >= 3 {
			stashRef := stashParts[0]
			branch := stashParts[1]
			message := stashParts[2]

			// Style stash reference
			refStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFEAA7")).Bold(true)
			styledRef := refStyle.Render(stashRef)

			// Style branch
			branchStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#74B9FF"))
			styledBranch := branchStyle.Render(branch)

			// Style message
			messageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#DDD6FE"))
			if isSelected {
				messageStyle = messageStyle.Bold(true)
			}
			styledMessage := messageStyle.Render(message)

			parts = append(parts, styledRef+":", styledBranch+":", styledMessage)
		} else {
			parts = append(parts, stashInfo)
		}
	} else {
		parts = append(parts, stashInfo)
	}

	line := strings.Join(parts, " ")

	// Apply selection styling
	if isSelected {
		style := lipgloss.NewStyle().
			Background(lipgloss.Color("#2D3748")).
			Padding(0, 1)
		return style.Render(line)
	}

	return line
}

// getFooter returns footer information
func (s *StashPane) getFooter() string {
	count := len(s.items)
	selected := s.GetSelectedIndex() + 1

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#74B9FF")).
		Render(fmt.Sprintf("Stashes: %d/%d", selected, count))
}

// Refresh refreshes the stash pane data
func (s *StashPane) Refresh() tea.Cmd {
	s.SetLoading(true)
	return func() tea.Msg {
		stashes := s.gitRepo.GetStashes()
		return git.StashUpdateMsg{Stashes: stashes}
	}
}

// HandleAction handles pane-specific actions
func (s *StashPane) HandleAction(action string) tea.Cmd {
	selectedItem := s.GetSelectedItem()

	switch action {
	case "create_stash":
		return s.createStash(false)
	case "create_stash_include_untracked":
		return s.createStash(true)
	case "apply_stash":
		if selectedItem == nil {
			return nil
		}
		return s.applyStash(selectedItem.Value)
	case "pop_stash":
		if selectedItem == nil {
			return nil
		}
		return s.popStash(selectedItem.Value)
	case "drop_stash":
		if selectedItem == nil {
			return nil
		}
		return s.dropStash(selectedItem.Value)
	case "show_stash":
		if selectedItem == nil {
			return nil
		}
		return s.showStash(selectedItem.Value)
	case "clear_all_stashes":
		return s.clearAllStashes()
	case "create_branch_from_stash":
		if selectedItem == nil {
			return nil
		}
		return s.createBranchFromStash(selectedItem.Value)
	default:
		return nil
	}
}

// GetAvailableActions returns available actions for this pane
func (s *StashPane) GetAvailableActions() []string {
	return []string{
		"create_stash", "create_stash_include_untracked", "apply_stash", "pop_stash",
		"drop_stash", "show_stash", "clear_all_stashes", "create_branch_from_stash", "refresh",
	}
}

// loadStashes loads initial stash data
func (s *StashPane) loadStashes() {
	s.Clear()

	stashes := s.gitRepo.GetStashes()

	if len(stashes) == 0 {
		s.AddItem(PaneItem{
			Display: "No stashed changes",
			Value:   "",
			Icon:    "",
			Type:    "empty",
		})
		return
	}

	for i, stashLine := range stashes {
		// Parse stash line format: "stash@{0}: On branch: message"
		stashRef := fmt.Sprintf("stash@{%d}", i)
		if strings.HasPrefix(stashLine, "stash@{") {
			// Extract the stash reference from the line
			endIndex := strings.Index(stashLine, "}")
			if endIndex > 0 {
				stashRef = stashLine[:endIndex+1]
			}
		}

		s.AddItem(PaneItem{
			Display: stashLine,
			Value:   stashRef,
			Icon:    "📦",
			Type:    "stash",
		})
	}
}

// createStash creates a new stash
func (s *StashPane) createStash(includeUntracked bool) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git stash or git stash -u
		action := "create_stash"
		message := "Created new stash"
		if includeUntracked {
			action = "create_stash_include_untracked"
			message = "Created new stash (including untracked files)"
		}

		return git.ActionCompleteMsg{
			Action:  action,
			Success: true,
			Message: message,
		}
	}
}

// applyStash applies the specified stash
func (s *StashPane) applyStash(stashRef string) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git stash apply
		return git.ActionCompleteMsg{
			Action:  "apply_stash",
			Success: true,
			Message: "Applied stash: " + stashRef,
		}
	}
}

// popStash pops the specified stash
func (s *StashPane) popStash(stashRef string) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git stash pop
		return git.ActionCompleteMsg{
			Action:  "pop_stash",
			Success: true,
			Message: "Popped stash: " + stashRef,
		}
	}
}

// dropStash drops the specified stash
func (s *StashPane) dropStash(stashRef string) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git stash drop
		return git.ActionCompleteMsg{
			Action:  "drop_stash",
			Success: true,
			Message: "Dropped stash: " + stashRef,
		}
	}
}

// showStash shows the contents of the specified stash
func (s *StashPane) showStash(stashRef string) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git stash show or git show
		diff := "Stash contents would be shown here"
		return git.DiffUpdateMsg{
			Diff: diff,
			File: stashRef,
		}
	}
}

// clearAllStashes clears all stashes
func (s *StashPane) clearAllStashes() tea.Cmd {
	return func() tea.Msg {
		// This would typically run git stash clear
		return git.ActionCompleteMsg{
			Action:  "clear_all_stashes",
			Success: true,
			Message: "Cleared all stashes",
		}
	}
}

// createBranchFromStash creates a new branch from the specified stash
func (s *StashPane) createBranchFromStash(stashRef string) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git stash branch
		return git.ActionCompleteMsg{
			Action:  "create_branch_from_stash",
			Success: true,
			Message: "Branch creation dialog would appear here",
		}
	}
}

// updateFromStashMsg updates the pane from a stash update message
func (s *StashPane) updateFromStashMsg(msg git.StashUpdateMsg) {
	s.SetLoading(false)
	s.Clear()

	if len(msg.Stashes) == 0 {
		s.AddItem(PaneItem{
			Display: "No stashed changes",
			Value:   "",
			Icon:    "",
			Type:    "empty",
		})
		return
	}

	for i, stashLine := range msg.Stashes {
		stashRef := fmt.Sprintf("stash@{%d}", i)
		if strings.HasPrefix(stashLine, "stash@{") {
			endIndex := strings.Index(stashLine, "}")
			if endIndex > 0 {
				stashRef = stashLine[:endIndex+1]
			}
		}

		s.AddItem(PaneItem{
			Display: stashLine,
			Value:   stashRef,
			Icon:    "📦",
			Type:    "stash",
		})
	}
}
