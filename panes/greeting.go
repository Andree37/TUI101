package panes

import (
	"tui101/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type GreetingPane struct {
	BasePaneModel
	st *styles.Styles
}

func NewStashPane() *GreetingPane {
	base := NewBasePaneModel("Greeting", StashPaneType, "greeting")

	pane := &GreetingPane{
		BasePaneModel: base,
		st:            styles.NewStyles(),
	}

	pane.loadGreeting()
	return pane
}

func (g *GreetingPane) Init() tea.Cmd {
	return nil
}

func (g *GreetingPane) Update(msg tea.Msg) (Pane, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		if !g.IsActive() {
			return g, nil
		}
	}

	return g, nil
}

func (g *GreetingPane) View() string {
	if g.IsLoading() {
		return g.st.LoadingText.Render("Loading...")
	}

	if len(g.items) == 0 {
		return g.st.GreetingText.Render("Hi")
	}

	// Display the greeting with styling
	greeting := g.items[0].Display

	// Create a nice centered greeting
	styledGreeting := lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		g.st.GreetingText.Render(greeting),
		"",
		g.st.Dimmed.Render("Press ? for help"),
	)

	return styledGreeting
}

func (g *GreetingPane) Refresh() tea.Cmd {
	return nil
}

func (g *GreetingPane) HandleAction(action string) tea.Cmd {
	return nil
}

func (g *GreetingPane) GetAvailableActions() []string {
	return []string{}
}

func (g *GreetingPane) loadGreeting() {
	g.Clear()

	g.AddItem(PaneItem{
		Display: "Hi",
		Value:   "hi",
		Type:    "greeting",
	})
}
