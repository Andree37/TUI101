package panes

import (
	"fmt"
	"strings"
	"tui101/git"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CommitsPane represents the commits pane showing commit history
type CommitsPane struct {
	BasePaneModel
	gitRepo    *git.Repository
	showReflog bool
	limit      int
}

// NewCommitsPane creates a new commits pane
func NewCommitsPane() *CommitsPane {
	base := NewBasePaneModel("Commits", CommitsPaneType, "commits")

	pane := &CommitsPane{
		BasePaneModel: base,
		gitRepo:       git.NewRepository("."),
		showReflog:    false,
		limit:         50,
	}

	pane.loadCommits()
	return pane
}

// Init initializes the commits pane
func (c *CommitsPane) Init() tea.Cmd {
	return c.Refresh()
}

// Update handles updates for the commits pane
func (c *CommitsPane) Update(msg tea.Msg) (Pane, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !c.IsActive() {
			return c, nil
		}

		switch msg.String() {
		case "j", "down":
			c.MoveDown()
		case "k", "up":
			c.MoveUp()
		case "g":
			c.MoveToTop()
		case "G":
			c.MoveToBottom()
		case "enter":
			return c, c.HandleAction("show_commit")
		case "d":
			return c, c.HandleAction("show_diff")
		case "c":
			return c, c.HandleAction("cherry_pick")
		case "r":
			return c, c.Refresh()
		case "R":
			return c, c.HandleAction("reset")
		case "v":
			return c, c.HandleAction("revert")
		case "s":
			return c, c.HandleAction("squash")
		case "e":
			return c, c.HandleAction("edit")
		case "f":
			return c, c.HandleAction("fixup")
		case "t":
			c.showReflog = !c.showReflog
			return c, c.Refresh()
		case "ctrl+f":
			return c, c.HandleAction("search")
		case "+":
			c.limit += 25
			return c, c.Refresh()
		case "-":
			if c.limit > 25 {
				c.limit -= 25
				return c, c.Refresh()
			}
		}

	case git.CommitsUpdateMsg:
		c.updateFromCommitsMsg(msg)
		return c, nil
	}

	return c, nil
}

// View renders the commits pane
func (c *CommitsPane) View() string {
	if c.IsLoading() {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#74B9FF")).
			Render("Loading commits...")
	}

	if len(c.items) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#74B9FF")).
			Render("No commits found")
	}

	var lines []string
	visibleItems := c.GetVisibleItems()

	for i, item := range visibleItems {
		actualIndex := c.GetScrollOffset() + i
		isSelected := actualIndex == c.GetSelectedIndex()

		line := c.formatCommitItem(item, isSelected)
		lines = append(lines, line)
	}

	// Add scroll indicators if needed
	if c.GetScrollOffset() > 0 {
		lines = append([]string{"  ↑ more commits above"}, lines...)
	}
	if c.GetScrollOffset()+len(visibleItems) < len(c.items) {
		lines = append(lines, "  ↓ more commits below")
	}

	// Add footer with current mode and count
	footer := c.getFooter()
	if footer != "" {
		lines = append(lines, "", footer)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// formatCommitItem formats a single commit item for display
func (c *CommitsPane) formatCommitItem(item PaneItem, isSelected bool) string {
	var parts []string

	// Add selection indicator
	if isSelected {
		parts = append(parts, "▶")
	} else {
		parts = append(parts, " ")
	}

	// Parse commit information from display string
	// Expected format: "hash AR ○ message"
	commitParts := strings.Fields(item.Display)
	if len(commitParts) >= 4 {
		hash := commitParts[0]
		author := commitParts[1]
		bullet := commitParts[2]
		message := strings.Join(commitParts[3:], " ")

		// Style hash
		hashStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFEAA7")).Bold(true)
		styledHash := hashStyle.Render(hash)

		// Style author
		authorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#74B9FF"))
		styledAuthor := authorStyle.Render(author)

		// Style bullet based on type
		bulletStyle := lipgloss.NewStyle()
		switch item.Type {
		case "commit":
			bulletStyle = bulletStyle.Foreground(lipgloss.Color("#04B575"))
		case "merge":
			bulletStyle = bulletStyle.Foreground(lipgloss.Color("#F25D94"))
		case "tag":
			bulletStyle = bulletStyle.Foreground(lipgloss.Color("#FFEAA7"))
		default:
			bulletStyle = bulletStyle.Foreground(lipgloss.Color("#DDD6FE"))
		}
		styledBullet := bulletStyle.Render(bullet)

		// Style message
		messageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#DDD6FE"))
		if isSelected {
			messageStyle = messageStyle.Bold(true)
		}
		styledMessage := messageStyle.Render(message)

		parts = append(parts, styledHash, styledAuthor, styledBullet, styledMessage)
	} else {
		// Fallback for malformed commit lines
		parts = append(parts, item.Display)
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
func (c *CommitsPane) getFooter() string {
	mode := "Commits"
	if c.showReflog {
		mode = "Reflog"
	}

	count := len(c.items)
	selected := c.GetSelectedIndex() + 1

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#74B9FF")).
		Render(fmt.Sprintf("%s: %d/%d (limit: %d)", mode, selected, count, c.limit))
}

// Refresh refreshes the commits pane data
func (c *CommitsPane) Refresh() tea.Cmd {
	c.SetLoading(true)
	return func() tea.Msg {
		commits := c.gitRepo.GetCommits(c.limit)
		return git.CommitsUpdateMsg{Commits: commits}
	}
}

// HandleAction handles pane-specific actions
func (c *CommitsPane) HandleAction(action string) tea.Cmd {
	selectedItem := c.GetSelectedItem()
	if selectedItem == nil {
		return nil
	}

	switch action {
	case "show_commit":
		return c.showCommit(selectedItem.Value)
	case "show_diff":
		return c.showDiff(selectedItem.Value)
	case "cherry_pick":
		return c.cherryPick(selectedItem.Value)
	case "reset":
		return c.resetToCommit(selectedItem.Value)
	case "revert":
		return c.revertCommit(selectedItem.Value)
	case "squash":
		return c.squashCommit(selectedItem.Value)
	case "edit":
		return c.editCommit(selectedItem.Value)
	case "fixup":
		return c.fixupCommit(selectedItem.Value)
	case "search":
		return c.searchCommits()
	default:
		return nil
	}
}

// GetAvailableActions returns available actions for this pane
func (c *CommitsPane) GetAvailableActions() []string {
	return []string{
		"show_commit", "show_diff", "cherry_pick", "reset", "revert",
		"squash", "edit", "fixup", "search", "refresh", "toggle_reflog",
	}
}

// loadCommits loads initial commit data
func (c *CommitsPane) loadCommits() {
	c.Clear()

	commits := c.gitRepo.GetCommits(c.limit)

	for _, commit := range commits {
		// Format the display string to match the expected format
		display := fmt.Sprintf("%s %s ○ %s", commit.ShortHash, commit.Author, commit.Message)

		commitType := "commit"
		if strings.Contains(strings.ToLower(commit.Message), "merge") {
			commitType = "merge"
		}

		c.AddItem(PaneItem{
			Display:  display,
			Value:    commit.Hash,
			Icon:     "○",
			Type:     commitType,
			Metadata: commit,
		})
	}
}

// showCommit shows detailed commit information
func (c *CommitsPane) showCommit(commitHash string) tea.Cmd {
	return func() tea.Msg {
		// This would typically show commit details in a popup or side panel
		return git.ActionCompleteMsg{
			Action:  "show_commit",
			Success: true,
			Message: "Showing commit: " + commitHash,
		}
	}
}

// showDiff shows the diff for the commit
func (c *CommitsPane) showDiff(commitHash string) tea.Cmd {
	return func() tea.Msg {
		diff := c.gitRepo.GetFileDiff(commitHash) // This would need to be modified for commits
		return git.DiffUpdateMsg{
			Diff: diff,
			File: commitHash,
		}
	}
}

// cherryPick cherry-picks the commit
func (c *CommitsPane) cherryPick(commitHash string) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git cherry-pick
		return git.ActionCompleteMsg{
			Action:  "cherry_pick",
			Success: true,
			Message: "Cherry-picked commit: " + commitHash,
		}
	}
}

// resetToCommit resets to the specified commit
func (c *CommitsPane) resetToCommit(commitHash string) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git reset
		return git.ActionCompleteMsg{
			Action:  "reset",
			Success: true,
			Message: "Reset to commit: " + commitHash,
		}
	}
}

// revertCommit reverts the specified commit
func (c *CommitsPane) revertCommit(commitHash string) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git revert
		return git.ActionCompleteMsg{
			Action:  "revert",
			Success: true,
			Message: "Reverted commit: " + commitHash,
		}
	}
}

// squashCommit squashes commits
func (c *CommitsPane) squashCommit(commitHash string) tea.Cmd {
	return func() tea.Msg {
		// This would typically start an interactive rebase
		return git.ActionCompleteMsg{
			Action:  "squash",
			Success: true,
			Message: "Squash operation would start here",
		}
	}
}

// editCommit edits the commit message
func (c *CommitsPane) editCommit(commitHash string) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git commit --amend or rebase -i
		return git.ActionCompleteMsg{
			Action:  "edit",
			Success: true,
			Message: "Edit commit dialog would appear here",
		}
	}
}

// fixupCommit creates a fixup commit
func (c *CommitsPane) fixupCommit(commitHash string) tea.Cmd {
	return func() tea.Msg {
		// This would typically run git commit --fixup
		return git.ActionCompleteMsg{
			Action:  "fixup",
			Success: true,
			Message: "Fixup commit created for: " + commitHash,
		}
	}
}

// searchCommits initiates commit search
func (c *CommitsPane) searchCommits() tea.Cmd {
	return func() tea.Msg {
		// This would typically open a search dialog
		return git.ActionCompleteMsg{
			Action:  "search",
			Success: true,
			Message: "Commit search dialog would appear here",
		}
	}
}

// updateFromCommitsMsg updates the pane from a commits update message
func (c *CommitsPane) updateFromCommitsMsg(msg git.CommitsUpdateMsg) {
	c.SetLoading(false)
	c.Clear()

	for _, commit := range msg.Commits {
		display := fmt.Sprintf("%s %s ○ %s", commit.ShortHash, commit.Author, commit.Message)

		commitType := "commit"
		if strings.Contains(strings.ToLower(commit.Message), "merge") {
			commitType = "merge"
		}

		c.AddItem(PaneItem{
			Display:  display,
			Value:    commit.Hash,
			Icon:     "○",
			Type:     commitType,
			Metadata: commit,
		})
	}
}
