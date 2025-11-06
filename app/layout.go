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

	for i := 0; i < 4 && i < len(m.panes); i++ {
		pane := m.panes[i]
		isActive := i == m.activePane

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
	if m.activePane == 4 && len(m.panes) > 4 {
		return m.renderStashPane(width, height)
	}

	return m.renderDiffPane(width, height)
}

func (m *Model) renderStashPane(width, height int) string {
	if len(m.panes) <= 4 {
		return m.createPaneStyle(width, height, false).Render("No stash pane available")
	}

	stashPane := m.panes[4]
	isActive := m.activePane == 4

	content := stashPane.View()
	title := m.renderPaneTitle(stashPane.GetTitle(), 5, isActive)
	fullContent := title + "\n" + content

	style := m.createPaneStyle(width, height, isActive)
	return style.Render(fullContent)
}

// renderDiffPane renders the diff pane in right column
func (m *Model) renderDiffPane(width, height int) string {
	title := m.renderPaneTitle("Diff", 0, false) // Not directly selectable

	diffContent := m.renderScrollableDiffContent(height - 4) // Reserve space for title and borders

	fullContent := title + "\n" + diffContent

	style := m.createPaneStyle(width, height, false)
	return style.Render(fullContent)
}

func (m *Model) renderPaneTitle(title string, number int, isActive bool) string {
	if number == 0 {
		if isActive {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true).Render(title)
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FFEAA7")).Render(title)
	}

	titleStyle := lipgloss.NewStyle()
	if isActive {
		titleStyle = titleStyle.Foreground(lipgloss.Color("#04B575")).Bold(true)
	} else {
		titleStyle = titleStyle.Foreground(lipgloss.Color("#FFEAA7"))
	}

	return titleStyle.Render(fmt.Sprintf("[%d] %s", number, title))
}

func (m *Model) createPaneStyle(width, height int, isActive bool) lipgloss.Style {
	baseStyle := lipgloss.NewStyle().
		Width(width-2).
		Height(height-2).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder())

	if isActive {
		return baseStyle.BorderForeground(lipgloss.Color("#04B575"))
	}

	return baseStyle.BorderForeground(lipgloss.Color("#6C5CE7"))
}

func (m *Model) renderStatusBar() string {
	currentPaneName := "Unknown"
	if m.activePane < len(m.panes) {
		currentPaneName = m.panes[m.activePane].GetTitle()
	}

	var leftStatus string
	if m.activePane != 4 {
		leftStatus = fmt.Sprintf("Active: %s | 1-5: Switch | Tab: Next | j/k: Scroll diff | q: Quit", currentPaneName)
	} else {
		leftStatus = fmt.Sprintf("Active: %s | 1-5: Switch | Tab: Next | j/k: Navigate | q: Quit", currentPaneName)
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

	return lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color("#1A202C")).
		Foreground(lipgloss.Color("#04B575")).
		Render(statusLine)
}

func (m *Model) renderScrollableDiffContent(maxLines int) string {
	diffLines := m.GetDiffLines()
	scrollPos := m.GetDiffScrollPos()

	if len(diffLines) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#74B9FF")).
			Render("Select a file to see diff")
	}

	start := scrollPos
	end := scrollPos + maxLines
	if end > len(diffLines) {
		end = len(diffLines)
	}
	if start >= len(diffLines) {
		start = len(diffLines) - 1
	}
	if start < 0 {
		start = 0
	}

	visibleLines := diffLines[start:end]

	var styledLines []string
	for _, line := range visibleLines {
		if len(line) > 120 {
			line = line[:120] + "..."
		}

		if len(line) == 0 {
			styledLines = append(styledLines, "")
			continue
		}

		switch line[0] {
		case '+':
			styledLines = append(styledLines,
				lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render(line))
		case '-':
			styledLines = append(styledLines,
				lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(line))
		case '@':
			styledLines = append(styledLines,
				lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true).Render(line))
		default:
			styledLines = append(styledLines,
				lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Render(line))
		}
	}

	result := strings.Join(styledLines, "\n")
	if scrollPos > 0 {
		result = "  ↑ more content above\n" + result
	}
	if end < len(diffLines) {
		result = result + "\n  ↓ more content below"
	}

	return result
}
