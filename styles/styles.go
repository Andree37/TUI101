package styles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	primaryColor    = lipgloss.Color("#04B575") // Green
	secondaryColor  = lipgloss.Color("#F25D94") // Pink
	accentColor     = lipgloss.Color("#FFEAA7") // Yellow
	mutedColor      = lipgloss.Color("#74B9FF") // Blue
	textColor       = lipgloss.Color("#DDD6FE") // Light purple
	borderColor     = lipgloss.Color("#6C5CE7") // Purple
	backgroundColor = lipgloss.Color("#1A202C") // Dark blue
	selectedBg      = lipgloss.Color("#2D3748") // Slightly lighter dark
	errorColor      = lipgloss.Color("#E53E3E") // Red
	warningColor    = lipgloss.Color("#D69E2E") // Orange
)

// Styles struct to hold all UI styles
type Styles struct {
	Border        lipgloss.Style
	ActiveBorder  lipgloss.Style
	Title         lipgloss.Style
	ActiveTitle   lipgloss.Style
	SelectedItem  lipgloss.Style
	Item          lipgloss.Style
	StatusBar     lipgloss.Style
	HelpText      lipgloss.Style
	DiffContent   lipgloss.Style
	GitHash       lipgloss.Style
	GitBranch     lipgloss.Style
	FileIcon      lipgloss.Style
	CommitMessage lipgloss.Style
}

// NewStyles creates and returns a new Styles instance
func NewStyles() *Styles {
	return &Styles{
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1),

		ActiveBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1),

		Title: lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			Padding(0, 1),

		ActiveTitle: lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1),

		SelectedItem: lipgloss.NewStyle().
			Foreground(primaryColor).
			Background(selectedBg).
			Padding(0, 1).
			Bold(true),

		Item: lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 1),

		StatusBar: lipgloss.NewStyle().
			Foreground(primaryColor).
			Background(backgroundColor).
			Padding(0, 1),

		HelpText: lipgloss.NewStyle().
			Foreground(mutedColor).
			Align(lipgloss.Right),

		DiffContent: lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(1),

		GitHash: lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true),

		GitBranch: lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true),

		FileIcon: lipgloss.NewStyle().
			Foreground(secondaryColor),

		CommitMessage: lipgloss.NewStyle().
			Foreground(textColor),
	}
}

// Panel border styles with dynamic width and height
func (s *Styles) PanelStyle(width, height int, isActive bool) lipgloss.Style {
	if isActive {
		return s.ActiveBorder.Copy().Width(width - 4).Height(height)
	}
	return s.Border.Copy().Width(width - 4).Height(height)
}

// Title style with dynamic activation
func (s *Styles) PanelTitle(title string, panelNumber int, isActive bool) string {
	titleText := lipgloss.NewStyle().
		Foreground(mutedColor).
		Render(fmt.Sprintf("[%d]-", panelNumber))

	if isActive {
		titleText += s.ActiveTitle.Render(title)
	} else {
		titleText += s.Title.Render(title)
	}

	return titleText
}
