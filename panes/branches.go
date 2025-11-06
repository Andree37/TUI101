package panes

import (
	"strings"
	"tui101/git"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BranchesPane represents the branches pane showing local and remote branches
type BranchesPane struct {
	BasePaneModel
	gitRepo    *git.Repository
	showRemote bool
	showTags   bool
}

// NewBranchesPane creates a new branches pane
func NewBranchesPane() *BranchesPane {
	base := NewBasePaneModel("Branches", BranchesPaneType, "branches")

	pane := &BranchesPane{
		BasePaneModel: base,
		gitRepo:       git.NewRepository("."),
		showRemote:    true,
		showTags:      true,
	}

	pane.loadBranches()
	return pane
}

// Init initializes the branches pane
func (b *BranchesPane) Init() tea.Cmd {
	return b.Refresh()
}

// Update handles updates for the branches pane
func (b *BranchesPane) Update(msg tea.Msg) (Pane, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !b.IsActive() {
			return b, nil
		}

		switch msg.String() {
		case "j", "down":
			b.MoveDown()
		case "k", "up":
			b.MoveUp()
		case "g":
			b.MoveToTop()
		case "G":
			b.MoveToBottom()
		case "enter":
			return b, b.HandleAction("checkout")
		case "c":
			return b, b.HandleAction("create_branch")
		case "d":
			return b, b.HandleAction("delete_branch")
		case "r":
			return b, b.Refresh()
		case "m":
			return b, b.HandleAction("merge")
		case "R":
			return b, b.HandleAction("rebase")
		case "p":
			return b, b.HandleAction("pull")
		case "P":
			return b, b.HandleAction("push")
		case "t":
			b.showTags = !b.showTags
			return b, b.Refresh()
		case "o":
			b.showRemote = !b.showRemote
			return b, b.Refresh()
		case "f":
			return b, b.HandleAction("fetch")
		}

	case git.BranchesUpdateMsg:
		b.updateFromBranchesMsg(msg)
		return b, nil
	}

	return b, nil
}

// View renders the branches pane
func (b *BranchesPane) View() string {
	if b.IsLoading() {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#74B9FF")).
			Render("Loading branches...")
	}

	if len(b.items) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#74B9FF")).
			Render("No branches found")
	}

	var lines []string
	visibleItems := b.GetVisibleItems()

	for i, item := range visibleItems {
		actualIndex := b.GetScrollOffset() + i
		isSelected := actualIndex == b.GetSelectedIndex()

		line := b.formatBranchItem(item, isSelected)
		lines = append(lines, line)
	}

	// Add scroll indicators if needed
	if b.GetScrollOffset() > 0 {
		lines = append([]string{"  ↑ more branches above"}, lines...)
	}
	if b.GetScrollOffset()+len(visibleItems) < len(b.items) {
		lines = append(lines, "  ↓ more branches below")
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// formatBranchItem formats a single branch item for display
func (b *BranchesPane) formatBranchItem(item PaneItem, isSelected bool) string {
	var parts []string

	// Add selection indicator
	if isSelected {
		parts = append(parts, "▶")
	} else {
		parts = append(parts, " ")
	}

	// Add branch indicator
	if item.Type == "current" {
		parts = append(parts, "*")
	} else {
		parts = append(parts, " ")
	}

	// Add branch name with icon
	if item.Icon != "" {
		parts = append(parts, item.Icon, item.Display)
	} else {
		parts = append(parts, item.Display)
	}

	line := strings.Join(parts, " ")

	// Apply styling based on selection and branch type
	style := lipgloss.NewStyle()

	if isSelected {
		style = style.Background(lipgloss.Color("#2D3748")).
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)
	} else {
		switch item.Type {
		case "current":
			style = style.Foreground(lipgloss.Color("#04B575")).Bold(true)
		case "local":
			style = style.Foreground(lipgloss.Color("#DDD6FE"))
		case "remote":
			style = style.Foreground(lipgloss.Color("#74B9FF"))
		case "tag":
			style = style.Foreground(lipgloss.Color("#FFEAA7"))
		default:
			style = style.Foreground(lipgloss.Color("#DDD6FE"))
		}
	}

	return style.Render(line)
}

// Refresh refreshes the branches pane data
func (b *BranchesPane) Refresh() tea.Cmd {
	b.SetLoading(true)
	return func() tea.Msg {
		branches := b.gitRepo.GetBranches()
		return git.BranchesUpdateMsg{Branches: branches}
	}
}

// HandleAction handles pane-specific actions
func (b *BranchesPane) HandleAction(action string) tea.Cmd {
	selectedItem := b.GetSelectedItem()
	if selectedItem == nil {
		return nil
	}

	switch action {
	case "checkout":
		return b.checkoutBranch(selectedItem.Value)
	case "create_branch":
		return b.createBranch()
	case "delete_branch":
		return b.deleteBranch(selectedItem.Value)
	case "merge":
		return b.mergeBranch(selectedItem.Value)
	case "rebase":
		return b.rebaseBranch(selectedItem.Value)
	case "pull":
		return b.pullBranch(selectedItem.Value)
	case "push":
		return b.pushBranch(selectedItem.Value)
	case "fetch":
		return b.fetchBranches()
	default:
		return nil
	}
}

// GetAvailableActions returns available actions for this pane
func (b *BranchesPane) GetAvailableActions() []string {
	return []string{
		"checkout", "create_branch", "delete_branch", "merge", "rebase",
		"pull", "push", "fetch", "refresh", "toggle_remote", "toggle_tags",
	}
}

// loadBranches loads initial branch data
func (b *BranchesPane) loadBranches() {
	b.Clear()

	branches := b.gitRepo.GetBranches()

	// Add local branches first
	for _, branch := range branches {
		if !branch.IsRemote {
			itemType := "local"
			icon := ""

			if branch.IsCurrent {
				itemType = "current"
				icon = "●"
			}

			b.AddItem(PaneItem{
				Display: branch.Name,
				Value:   branch.Name,
				Icon:    icon,
				Type:    itemType,
			})
		}
	}

	// Add remote branches if enabled
	if b.showRemote {
		for _, branch := range branches {
			if branch.IsRemote {
				b.AddItem(PaneItem{
					Display: branch.Name,
					Value:   branch.Name,
					Icon:    "↑",
					Type:    "remote",
				})
			}
		}
	}

	// Add tags if enabled (placeholder - would need actual git tag implementation)
	if b.showTags {
		tags := []string{"v1.0.0", "v1.1.0", "v2.0.0-beta"}
		for _, tag := range tags {
			b.AddItem(PaneItem{
				Display: tag,
				Value:   tag,
				Icon:    "🏷️",
				Type:    "tag",
			})
		}
	}
}

// checkoutBranch checks out the specified branch
func (b *BranchesPane) checkoutBranch(branchName string) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git checkout
		// For now, just return a success message
		return git.ActionCompleteMsg{
			Action:  "checkout",
			Success: true,
			Message: "Checked out branch: " + branchName,
		}
	}
}

// createBranch creates a new branch
func (b *BranchesPane) createBranch() tea.Cmd {
	return func() tea.Msg {
		// This would typically prompt for branch name and create it
		return git.ActionCompleteMsg{
			Action:  "create_branch",
			Success: true,
			Message: "Branch creation dialog would appear here",
		}
	}
}

// deleteBranch deletes the specified branch
func (b *BranchesPane) deleteBranch(branchName string) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git branch -d
		return git.ActionCompleteMsg{
			Action:  "delete_branch",
			Success: true,
			Message: "Deleted branch: " + branchName,
		}
	}
}

// mergeBranch merges the specified branch
func (b *BranchesPane) mergeBranch(branchName string) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git merge
		return git.ActionCompleteMsg{
			Action:  "merge",
			Success: true,
			Message: "Merged branch: " + branchName,
		}
	}
}

// rebaseBranch rebases the specified branch
func (b *BranchesPane) rebaseBranch(branchName string) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git rebase
		return git.ActionCompleteMsg{
			Action:  "rebase",
			Success: true,
			Message: "Rebased branch: " + branchName,
		}
	}
}

// pullBranch pulls the specified branch
func (b *BranchesPane) pullBranch(branchName string) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git pull
		return git.ActionCompleteMsg{
			Action:  "pull",
			Success: true,
			Message: "Pulled branch: " + branchName,
		}
	}
}

// pushBranch pushes the specified branch
func (b *BranchesPane) pushBranch(branchName string) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git push
		return git.ActionCompleteMsg{
			Action:  "push",
			Success: true,
			Message: "Pushed branch: " + branchName,
		}
	}
}

// fetchBranches fetches all branches
func (b *BranchesPane) fetchBranches() tea.Cmd {
	return func() tea.Msg {
		err := b.gitRepo.Fetch()
		if err != nil {
			return git.ErrorMsg{Error: err}
		}
		branches := b.gitRepo.GetBranches()
		return git.BranchesUpdateMsg{Branches: branches}
	}
}

// updateFromBranchesMsg updates the pane from a branches update message
func (b *BranchesPane) updateFromBranchesMsg(msg git.BranchesUpdateMsg) {
	b.SetLoading(false)
	b.Clear()

	// Add local branches first
	for _, branch := range msg.Branches {
		if !branch.IsRemote {
			itemType := "local"
			icon := ""

			if branch.IsCurrent {
				itemType = "current"
				icon = "●"
			}

			b.AddItem(PaneItem{
				Display: branch.Name,
				Value:   branch.Name,
				Icon:    icon,
				Type:    itemType,
			})
		}
	}

	// Add remote branches if enabled
	if b.showRemote {
		for _, branch := range msg.Branches {
			if branch.IsRemote {
				b.AddItem(PaneItem{
					Display: branch.Name,
					Value:   branch.Name,
					Icon:    "↑",
					Type:    "remote",
				})
			}
		}
	}
}
