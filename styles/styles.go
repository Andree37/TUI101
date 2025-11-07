package styles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Color constants
const (
	Green       = "#04B575"
	Yellow      = "#FFEAA7"
	Blue        = "#74B9FF"
	Purple      = "#6C5CE7"
	LightGray   = "#DDD6FE"
	DarkGray    = "#2D3748"
	Background  = "#1A202C"
	Red         = "#E53E3E"
	LightPurple = "#A78BFA"
	Orange      = "#FFA07A"
	Cyan        = "#48D1CC"
	Pink        = "#FF69B4"
	White       = "#FFFFFF"
	DimGray     = "#696969"
)

type Styles struct {
	// Border styles
	ActiveBorder   lipgloss.Style
	InactiveBorder lipgloss.Style

	// Title styles
	ActiveTitle   lipgloss.Style
	InactiveTitle lipgloss.Style

	// Item styles
	SelectedItem   lipgloss.Style
	UnselectedItem lipgloss.Style

	// Status bar style
	StatusBar lipgloss.Style

	// Loading/Info styles
	InfoText    lipgloss.Style
	LoadingText lipgloss.Style
	ErrorText   lipgloss.Style
	SuccessText lipgloss.Style
	WarningText lipgloss.Style

	// Cursor style
	Cursor lipgloss.Style

	// Package-specific styles
	PackageActive   lipgloss.Style
	PackageInactive lipgloss.Style

	// PR status styles
	PROpen   lipgloss.Style
	PRClosed lipgloss.Style
	PRMerged lipgloss.Style

	// Workspace info styles
	WorkspaceName     lipgloss.Style
	WorkspaceVersion  lipgloss.Style
	WorkspaceMetadata lipgloss.Style

	// Greeting styles
	GreetingText lipgloss.Style

	// Scrollbar indicators
	ScrollIndicator lipgloss.Style

	// Footer styles
	Footer lipgloss.Style

	// Highlighted text
	Highlight lipgloss.Style

	// Dimmed text
	Dimmed lipgloss.Style
}

func NewStyles() *Styles {
	return &Styles{
		// Border styles
		ActiveBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(Green)).
			Padding(0, 1),

		InactiveBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(Purple)).
			Padding(0, 1),

		// Title styles
		ActiveTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Green)).
			Bold(true).
			Padding(0, 1),

		InactiveTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Yellow)).
			Padding(0, 1),

		// Item styles
		SelectedItem: lipgloss.NewStyle().
			Background(lipgloss.Color(DarkGray)).
			Foreground(lipgloss.Color(Green)).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1),

		UnselectedItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color(LightGray)).
			PaddingLeft(1).
			PaddingRight(1),

		// Status bar style
		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color(DarkGray)).
			Foreground(lipgloss.Color(Green)).
			Padding(0, 1).
			Bold(true),

		// Info styles
		InfoText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Blue)).
			Italic(true),

		LoadingText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Cyan)).
			Italic(true),

		ErrorText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Red)).
			Bold(true),

		SuccessText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Green)).
			Bold(true),

		WarningText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Yellow)).
			Bold(true),

		// Cursor
		Cursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Green)).
			Bold(true),

		// Package styles
		PackageActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Green)).
			Bold(true),

		PackageInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color(DimGray)),

		// PR status styles
		PROpen: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Green)),

		PRClosed: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Red)),

		PRMerged: lipgloss.NewStyle().
			Foreground(lipgloss.Color(LightPurple)).
			Bold(true),

		// Workspace info styles
		WorkspaceName: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Green)).
			Bold(true).
			Underline(true),

		WorkspaceVersion: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Yellow)),

		WorkspaceMetadata: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Blue)).
			Italic(true),

		// Greeting styles
		GreetingText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(LightPurple)).
			Bold(true).
			Align(lipgloss.Center),

		// Scrollbar indicators
		ScrollIndicator: lipgloss.NewStyle().
			Foreground(lipgloss.Color(DimGray)).
			Italic(true),

		// Footer styles
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Blue)).
			Italic(true).
			PaddingTop(1),

		// Highlighted text
		Highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Cyan)).
			Bold(true),

		// Dimmed text
		Dimmed: lipgloss.NewStyle().
			Foreground(lipgloss.Color(DimGray)),
	}
}

// Pane creates a bordered pane style
func (s *Styles) Pane(width, height int, isActive bool) lipgloss.Style {
	if isActive {
		return s.ActiveBorder.Copy().
			Width(width - 4).
			Height(height - 4)
	}
	return s.InactiveBorder.Copy().
		Width(width - 4).
		Height(height - 4)
}

// Title creates a pane title style
func (s *Styles) Title(isActive bool) lipgloss.Style {
	if isActive {
		return s.ActiveTitle
	}
	return s.InactiveTitle
}

// Item creates an item style based on selection state
func (s *Styles) Item(isSelected bool) lipgloss.Style {
	if isSelected {
		return s.SelectedItem
	}
	return s.UnselectedItem
}

// PaneTitle renders a styled pane title with optional icons
func (s *Styles) PaneTitle(title string, isActive bool, icon string) string {
	style := s.Title(isActive)
	if icon != "" {
		return style.Render(icon + " " + title)
	}
	return style.Render(title)
}

// PRStatus returns the appropriate style for a PR status
func (s *Styles) PRStatus(status string) lipgloss.Style {
	switch status {
	case "open":
		return s.PROpen
	case "closed":
		return s.PRClosed
	case "merged":
		return s.PRMerged
	default:
		return s.UnselectedItem
	}
}

// WorkspaceItemStyle returns the appropriate style for workspace items
func (s *Styles) WorkspaceItemStyle(itemType string) lipgloss.Style {
	switch itemType {
	case "name":
		return s.WorkspaceName
	case "version":
		return s.WorkspaceVersion
	case "metadata":
		return s.WorkspaceMetadata
	default:
		return s.UnselectedItem
	}
}

// PackageItemStyle returns the appropriate style for package items
func (s *Styles) PackageItemStyle(status string) lipgloss.Style {
	switch status {
	case "active":
		return s.PackageActive
	case "inactive":
		return s.PackageInactive
	default:
		return s.UnselectedItem
	}
}

// RenderCursor renders a cursor with the appropriate style
func (s *Styles) RenderCursor(isActive bool) string {
	if isActive {
		return s.Cursor.Render("❯ ")
	}
	return "  "
}

// RenderScrollIndicator renders scroll indicators
func (s *Styles) RenderScrollIndicator(direction string) string {
	if direction == "up" {
		return s.ScrollIndicator.Render("  ↑ more items above")
	}
	return s.ScrollIndicator.Render("  ↓ more items below")
}

// RenderFooter renders a footer with count information
func (s *Styles) RenderFooter(label string, current, total int) string {
	return s.Footer.Render(lipgloss.JoinHorizontal(
		lipgloss.Left,
		label+": ",
		s.Highlight.Render(lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Render(fmt.Sprintf("%d", current)),
			"/",
			lipgloss.NewStyle().Render(fmt.Sprintf("%d", total)),
		)),
	))
}
