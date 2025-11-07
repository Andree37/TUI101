package panes

import (
	"fmt"
	"tui101/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PackagesPane struct {
	BasePaneModel
	packages []Package
	st       *styles.Styles
}

type PackagesUpdateMsg struct {
	Packages []Package
}

type Package struct {
	Name   string
	Status string
}

func NewBranchesPane() *PackagesPane {
	base := NewBasePaneModel("Packages", BranchesPaneType, "packages")

	pane := &PackagesPane{
		BasePaneModel: base,
		packages:      []Package{},
		st:            styles.NewStyles(),
	}

	pane.loadPackages()
	return pane
}

func (p *PackagesPane) Init() tea.Cmd {
	return p.Refresh()
}

func (p *PackagesPane) Update(msg tea.Msg) (Pane, tea.Cmd) {
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

	case PackagesUpdateMsg:
		p.updateFromPackagesMsg(msg)
		return p, nil
	}

	return p, nil
}

func (p *PackagesPane) View() string {
	if p.IsLoading() {
		return p.st.LoadingText.Render("Loading packages...")
	}

	if len(p.items) == 0 {
		return p.st.InfoText.Render("No packages found")
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

		line := p.formatPackageItem(item, isSelected)
		lines = append(lines, line)
	}

	// Show scroll indicator at bottom if needed
	if p.GetScrollOffset()+len(visibleItems) < len(p.items) {
		lines = append(lines, p.st.RenderScrollIndicator("down"))
	}

	// Add footer with package count
	if len(p.items) > 0 {
		lines = append(lines, "")
		footer := p.st.RenderFooter("Packages", p.GetSelectedIndex()+1, len(p.items))
		lines = append(lines, footer)
	}

	// Add help text if active
	if p.IsActive() {
		lines = append(lines, "")
		lines = append(lines, p.st.Dimmed.Render("j/k: Navigate  g/G: Top/Bottom  r: Refresh"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (p *PackagesPane) formatPackageItem(item PaneItem, isSelected bool) string {
	var style lipgloss.Style

	// Choose style based on status
	switch item.Type {
	case "active":
		style = p.st.PackageActive
	case "inactive":
		style = p.st.PackageInactive
	default:
		style = p.st.UnselectedItem
	}

	// Override with selection style if selected and active
	if isSelected && p.IsActive() {
		style = p.st.SelectedItem
		return style.Render(fmt.Sprintf("%s %s", p.st.RenderCursor(true), item.Display))
	}

	return style.Render(fmt.Sprintf("  %s", item.Display))
}

func (p *PackagesPane) Refresh() tea.Cmd {
	p.SetLoading(true)
	return func() tea.Msg {
		packages := p.gatherPackages()
		return PackagesUpdateMsg{Packages: packages}
	}
}

func (p *PackagesPane) HandleAction(action string) tea.Cmd {
	switch action {
	case "refresh":
		return p.Refresh()
	}
	return nil
}

func (p *PackagesPane) GetAvailableActions() []string {
	return []string{"refresh"}
}

func (p *PackagesPane) loadPackages() {
	p.Clear()

	packages := []Package{
		{Name: "antonio", Status: "active"},
		{Name: "miguel", Status: "active"},
		{Name: "rita", Status: "active"},
	}

	p.packages = packages

	for _, pkg := range packages {
		p.AddItem(PaneItem{
			Display: pkg.Name,
			Value:   pkg.Name,
			Type:    pkg.Status,
		})
	}
}

func (p *PackagesPane) gatherPackages() []Package {
	return []Package{
		{Name: "antonio", Status: "active"},
		{Name: "miguel", Status: "active"},
		{Name: "rita", Status: "active"},
	}
}

func (p *PackagesPane) updateFromPackagesMsg(msg PackagesUpdateMsg) {
	p.SetLoading(false)
	p.Clear()
	p.packages = msg.Packages

	for _, pkg := range msg.Packages {
		p.AddItem(PaneItem{
			Display: pkg.Name,
			Value:   pkg.Name,
			Type:    pkg.Status,
		})
	}
}
