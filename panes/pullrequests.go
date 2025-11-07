package panes

import (
	"fmt"
	"strings"
	"time"
	"tui101/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PullRequestsPane struct {
	BasePaneModel
	pullRequests []PullRequest
	st           *styles.Styles
}

type PullRequestsUpdateMsg struct {
	PullRequests []PullRequest
}

type PullRequest struct {
	ID      int
	Title   string
	Package string
	Author  string
	Status  string
	Created time.Time
}

func NewCommitsPane() *PullRequestsPane {
	base := NewBasePaneModel("Pull Requests", CommitsPaneType, "pullrequests")

	pane := &PullRequestsPane{
		BasePaneModel: base,
		pullRequests:  []PullRequest{},
		st:            styles.NewStyles(),
	}

	pane.loadPullRequests()
	return pane
}

func (p *PullRequestsPane) Init() tea.Cmd {
	return p.Refresh()
}

func (p *PullRequestsPane) Update(msg tea.Msg) (Pane, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !p.IsActive() {
			return p, nil
		}

		switch msg.String() {
		case "j", "down":
			p.MoveDown()
		case "k", "up":
			p.MoveUp()
		case "g":
			p.MoveToTop()
		case "G":
			p.MoveToBottom()
		case "r":
			return p, p.Refresh()
		}

	case PullRequestsUpdateMsg:
		p.updateFromPullRequestsMsg(msg)
		return p, nil
	}

	return p, nil
}

func (p *PullRequestsPane) View() string {
	if p.IsLoading() {
		return p.st.LoadingText.Render("Loading pull requests...")
	}

	if len(p.items) == 0 {
		return p.st.InfoText.Render("No pull requests")
	}

	var lines []string
	visibleItems := p.GetVisibleItems()

	// Show scroll indicator at top if needed
	if p.GetScrollOffset() > 0 {
		lines = append(lines, p.st.RenderScrollIndicator("up"))
	}

	for i, item := range visibleItems {
		actualIndex := p.GetScrollOffset() + i
		isSelected := actualIndex == p.GetSelectedIndex()

		line := p.formatPRItem(item, isSelected)
		lines = append(lines, line)
	}

	// Show scroll indicator at bottom if needed
	if p.GetScrollOffset()+len(visibleItems) < len(p.items) {
		lines = append(lines, p.st.RenderScrollIndicator("down"))
	}

	// Add footer with PR count
	footer := p.getFooter()
	if footer != "" {
		lines = append(lines, "")
		lines = append(lines, footer)
	}

	// Add help text if active
	if p.IsActive() {
		lines = append(lines, "")
		lines = append(lines, p.st.Dimmed.Render("j/k: Navigate  g/G: Top/Bottom  r: Refresh"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (p *PullRequestsPane) formatPRItem(item PaneItem, isSelected bool) string {
	var statusBadge string
	var statusStyle lipgloss.Style

	// Get status badge
	switch item.Type {
	case "open":
		statusBadge = "OPEN"
		statusStyle = p.st.PROpen
	case "closed":
		statusBadge = "CLOSED"
		statusStyle = p.st.PRClosed
	case "merged":
		statusBadge = "MERGED"
		statusStyle = p.st.PRMerged
	default:
		statusBadge = "UNKNOWN"
		statusStyle = p.st.Dimmed
	}

	// Format the badge
	badge := statusStyle.Render(fmt.Sprintf("[%s]", statusBadge))

	// Build the line
	var line string
	if isSelected && p.IsActive() {
		line = fmt.Sprintf("%s %s %s", p.st.RenderCursor(true), badge, item.Display)
		return p.st.SelectedItem.Render(line)
	}

	line = fmt.Sprintf("  %s %s", badge, item.Display)

	// Apply status-specific styling when not selected
	if !isSelected {
		return statusStyle.Render(line)
	}

	return p.st.UnselectedItem.Render(line)
}

func (p *PullRequestsPane) getFooter() string {
	if len(p.items) == 0 {
		return ""
	}

	count := len(p.items)
	selected := p.GetSelectedIndex() + 1

	// Count by status
	openCount := 0
	closedCount := 0
	mergedCount := 0

	for _, pr := range p.pullRequests {
		switch pr.Status {
		case "open":
			openCount++
		case "closed":
			closedCount++
		case "merged":
			mergedCount++
		}
	}

	// Build footer with counts
	footerParts := []string{
		p.st.Highlight.Render(fmt.Sprintf("%d/%d", selected, count)),
	}

	if openCount > 0 {
		footerParts = append(footerParts, p.st.PROpen.Render(fmt.Sprintf("Open: %d", openCount)))
	}
	if closedCount > 0 {
		footerParts = append(footerParts, p.st.PRClosed.Render(fmt.Sprintf("Closed: %d", closedCount)))
	}
	if mergedCount > 0 {
		footerParts = append(footerParts, p.st.PRMerged.Render(fmt.Sprintf("Merged: %d", mergedCount)))
	}

	return p.st.Footer.Render(strings.Join(footerParts, " │ "))
}

func (p *PullRequestsPane) Refresh() tea.Cmd {
	p.SetLoading(true)
	return func() tea.Msg {
		// Simulate loading time
		time.Sleep(500 * time.Millisecond)
		prs := p.gatherPullRequests()
		return PullRequestsUpdateMsg{PullRequests: prs}
	}
}

func (p *PullRequestsPane) HandleAction(action string) tea.Cmd {
	switch action {
	case "refresh":
		return p.Refresh()
	}
	return nil
}

func (p *PullRequestsPane) GetAvailableActions() []string {
	return []string{"refresh", "view", "checkout"}
}

func (p *PullRequestsPane) loadPullRequests() {
	p.Clear()

	prs := []PullRequest{
		{
			ID:      1,
			Title:   "Add new feature",
			Package: "antonio",
			Author:  "john",
			Status:  "open",
			Created: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:      2,
			Title:   "Fix bug in handler",
			Package: "miguel",
			Author:  "jane",
			Status:  "open",
			Created: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:      3,
			Title:   "Update dependencies",
			Package: "rita",
			Author:  "bob",
			Status:  "merged",
			Created: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:      4,
			Title:   "Refactor authentication",
			Package: "antonio",
			Author:  "alice",
			Status:  "closed",
			Created: time.Now().Add(-48 * time.Hour),
		},
		{
			ID:      5,
			Title:   "Add tests for API",
			Package: "miguel",
			Author:  "charlie",
			Status:  "open",
			Created: time.Now().Add(-3 * time.Hour),
		},
	}

	p.pullRequests = prs

	for _, pr := range prs {
		display := fmt.Sprintf("#%d: %s [%s]", pr.ID, pr.Title, pr.Package)

		p.AddItem(PaneItem{
			Display:  display,
			Value:    fmt.Sprintf("%d", pr.ID),
			Type:     pr.Status,
			Metadata: pr,
		})
	}
}

func (p *PullRequestsPane) gatherPullRequests() []PullRequest {
	return []PullRequest{
		{
			ID:      1,
			Title:   "Add new feature",
			Package: "antonio",
			Author:  "john",
			Status:  "open",
			Created: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:      2,
			Title:   "Fix bug in handler",
			Package: "miguel",
			Author:  "jane",
			Status:  "open",
			Created: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:      3,
			Title:   "Update dependencies",
			Package: "rita",
			Author:  "bob",
			Status:  "merged",
			Created: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:      4,
			Title:   "Refactor authentication",
			Package: "antonio",
			Author:  "alice",
			Status:  "closed",
			Created: time.Now().Add(-48 * time.Hour),
		},
		{
			ID:      5,
			Title:   "Add tests for API",
			Package: "miguel",
			Author:  "charlie",
			Status:  "open",
			Created: time.Now().Add(-3 * time.Hour),
		},
	}
}

func (p *PullRequestsPane) updateFromPullRequestsMsg(msg PullRequestsUpdateMsg) {
	p.SetLoading(false)
	p.Clear()
	p.pullRequests = msg.PullRequests

	for _, pr := range msg.PullRequests {
		display := fmt.Sprintf("#%d: %s [%s]", pr.ID, pr.Title, pr.Package)

		p.AddItem(PaneItem{
			Display:  display,
			Value:    fmt.Sprintf("%d", pr.ID),
			Type:     pr.Status,
			Metadata: pr,
		})
	}
}
