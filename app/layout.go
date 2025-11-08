package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderLayout renders the complete application layout
func (m *Model) renderLayout(leftPaneWidth, rightPaneWidth, paneHeight int) string {
	totalHeight := m.height
	statusBarHeight := 1
	availableHeight := totalHeight - statusBarHeight

	leftPaneHeight := availableHeight / 4
	rightPaneHeight := availableHeight

	leftPanes := m.renderLeftColumn(leftPaneWidth, leftPaneHeight)

	rightPane := m.renderRightColumn(rightPaneWidth, rightPaneHeight)

	mainView := lipgloss.JoinHorizontal(lipgloss.Top, leftPanes, rightPane)

	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, mainView, statusBar)
}

func (m *Model) renderLeftColumn(width, paneHeight int) string {
	var panes []string

	for i := 0; i < len(m.panes) && i < 4; i++ {
		pane := m.panes[i]
		// Left panes should only be active when focus is on left panes
		isActive := i == m.activePane && m.focus == FocusLeftPanes

		content := pane.View()
		title := m.renderPaneTitle(pane.GetTitle(), i+1, isActive)
		fullContent := title + "\n" + content

		style := m.createPaneStyle(width, paneHeight, isActive)
		renderedPane := style.Render(fullContent)

		panes = append(panes, renderedPane)
	}

	return lipgloss.JoinVertical(lipgloss.Left, panes...)
}

func (m *Model) renderRightColumn(width, height int) string {
	if m.activePane == 3 && len(m.panes) > 3 {
		return m.renderGreetingPane(width, height)
	}

	return m.renderPreviewPane(width, height)
}

func (m *Model) renderGreetingPane(width, height int) string {
	if len(m.panes) <= 3 {
		return m.createPaneStyle(width, height, false).Render("No greeting pane available")
	}

	greetingPane := m.panes[3]
	// Greeting pane should only be active when focus is on left panes
	isActive := m.activePane == 3 && m.focus == FocusLeftPanes

	content := greetingPane.View()
	title := m.renderPaneTitle(greetingPane.GetTitle(), 4, isActive)
	fullContent := title + "\n" + content

	style := m.createPaneStyle(width, height, isActive)
	return style.Render(fullContent)
}

// renderPreviewPane renders the preview pane in right column
func (m *Model) renderPreviewPane(width, height int) string {
	isActive := m.focus == FocusDetails
	title := m.renderPaneTitle("Details", 0, isActive)

	previewContent := m.renderScrollablePreviewContent(height - 4) // Reserve space for title and borders

	fullContent := title + "\n" + previewContent

	style := m.createPaneStyle(width, height, isActive)
	return style.Render(fullContent)
}

func (m *Model) renderPaneTitle(title string, number int, isActive bool) string {
	titleStyle := m.styles.Title(isActive)

	if number == 0 {
		return titleStyle.Render(title)
	}

	return titleStyle.Render(fmt.Sprintf("[%d] %s", number, title))
}

func (m *Model) createPaneStyle(width, height int, isActive bool) lipgloss.Style {
	return m.styles.Pane(width, height, isActive)
}

func (m *Model) renderStatusBar() string {
	currentPaneName := "Unknown"
	if m.activePane < len(m.panes) {
		currentPaneName = m.panes[m.activePane].GetTitle()
	}

	var leftStatus string
	if m.focus == FocusDetails {
		leftStatus = "Active: Details | Space: Back to panes | j/k: Scroll | q: Quit"
	} else {
		leftStatus = fmt.Sprintf("Active: %s | 1-2: Switch | Tab: Next | Space: Details | j/k: Scroll | q: Quit", currentPaneName)
	}

	rightStatus := "TUI101 v0.1.0"

	maxLeftLen := m.width - len(rightStatus) - 5
	if len(leftStatus) > maxLeftLen {
		leftStatus = leftStatus[:maxLeftLen-3] + "..."
	}

	usedSpace := len(leftStatus) + len(rightStatus)
	padding := m.width - usedSpace
	if padding < 0 {
		padding = 0
	}

	statusLine := leftStatus + strings.Repeat(" ", padding) + rightStatus

	return m.styles.StatusBar.
		Width(m.width).
		Render(statusLine)
}

func (m *Model) renderScrollablePreviewContent(maxLines int) string {
	previewLines := m.GetPreviewLines()
	scrollPos := m.GetPreviewScrollPos()

	if len(previewLines) == 0 {
		return m.styles.InfoText.Render("Select an item to see details")
	}

	start := scrollPos
	end := scrollPos + maxLines
	if end > len(previewLines) {
		end = len(previewLines)
	}
	if start >= len(previewLines) {
		start = len(previewLines) - 1
	}
	if start < 0 {
		start = 0
	}

	visibleLines := previewLines[start:end]

	var styledLines []string
	for i, line := range visibleLines {
		actualIndex := start + i
		isSelected := m.focus == FocusDetails && actualIndex == m.details.selectedLine

		if isSelected {
			prefix := m.styles.Cursor.Render("> ")
			styledLines = append(styledLines, m.styles.SelectedItem.Render(prefix+line))
		} else {
			styledLines = append(styledLines, "  "+line)
		}
	}

	result := strings.Join(styledLines, "\n")

	if scrollPos > 0 {
		result = m.styles.Dimmed.Render("  ▲ more content above") + "\n" + result
	}
	if end < len(previewLines) {
		result = result + "\n" + m.styles.Dimmed.Render("  ▼ more content below")
	}

	return result
}
