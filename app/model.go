package app

import (
	"fmt"
	"tui101/panes"
	"tui101/styles"

	tea "github.com/charmbracelet/bubbletea"
)

type Focus int

const (
	FocusLeftPanes Focus = iota
	FocusDetails
)

type DetailsPane struct {
	selectedLine int
	scrollPos    int
	lines        []string
}

func (d *DetailsPane) Reset() {
	d.selectedLine = 0
	d.scrollPos = 0
}

func (d *DetailsPane) MoveDown() {
	if len(d.lines) == 0 {
		return
	}
	if d.selectedLine < len(d.lines)-1 {
		d.selectedLine++
	}
}

func (d *DetailsPane) MoveUp() {
	if d.selectedLine > 0 {
		d.selectedLine--
	}
}

func (d *DetailsPane) MoveToTop() {
	d.selectedLine = 0
	d.scrollPos = 0
}

func (d *DetailsPane) MoveToBottom() {
	if len(d.lines) > 0 {
		d.selectedLine = len(d.lines) - 1
	}
}

func (d *DetailsPane) AdjustScroll(maxLines int) {
	if maxLines < 1 {
		maxLines = 1
	}

	if d.selectedLine >= d.scrollPos+maxLines {
		d.scrollPos = d.selectedLine - maxLines + 1
	}

	if d.selectedLine < d.scrollPos {
		d.scrollPos = d.selectedLine
	}
}

func (d *DetailsPane) ScrollDown(maxLines int) {
	if len(d.lines) > 0 {
		maxScroll := len(d.lines) - maxLines
		if maxScroll < 0 {
			maxScroll = 0
		}
		if d.scrollPos < maxScroll {
			d.scrollPos++
		}
	}
}

func (d *DetailsPane) ScrollUp() {
	if d.scrollPos > 0 {
		d.scrollPos--
	}
}

type Model struct {
	panes      []panes.Pane
	activePane int
	width      int
	height     int
	styles     *styles.Styles
	quitting   bool
	filterMode bool
	filterText string
	details    DetailsPane
	focus      Focus
}

func NewModel() *Model {
	m := &Model{
		styles:     styles.NewStyles(),
		activePane: 0, // Start with workspace pane active
		focus:      FocusLeftPanes,
	}

	m.panes = []panes.Pane{
		panes.NewStatusPane(),   // Workspace
		panes.NewBranchesPane(), // Packages
		panes.NewCommitsPane(),  // Pull Requests
		panes.NewStashPane(),    // Greeting
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	for _, pane := range m.panes {
		cmds = append(cmds, pane.Init())
	}

	return tea.Batch(cmds...)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle space key first before anything else
		if msg.String() == " " {
			m.toggleFocus()
			return m, nil
		}

		// Handle global keybindings
		cmd := m.handleKeyMsg(msg)
		if cmd != nil {
			return m, cmd
		}

		// Don't pass keys to panes if focus is on details
		if m.focus == FocusDetails {
			return m, nil
		}

		// Pass keys to active pane when focus is on left panes
		if m.activePane < len(m.panes) {
			updatedPane, cmd := m.panes[m.activePane].Update(msg)
			m.panes[m.activePane] = updatedPane
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	default:
		for i, pane := range m.panes {
			updatedPane, cmd := pane.Update(msg)
			m.panes[i] = updatedPane
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return tea.Quit

	case "tab":
		return m.handlePaneNavigation(m.nextPane)
	case "shift+tab":
		return m.handlePaneNavigation(m.prevPane)

	case "1":
		return m.handlePaneNavigation(func() { m.setActivePane(0) })
	case "2":
		return m.handlePaneNavigation(func() { m.setActivePane(1) })
	case "3":
		return m.handlePaneNavigation(func() { m.setActivePane(2) })
	case "4":
		return m.handlePaneNavigation(func() { m.setActivePane(3) })

	case "ctrl+r":
		return m.refreshAll()

	case "?":
		return tea.Batch()

	case "j", "down":
		return m.handleVerticalNavigation(true)
	case "k", "up":
		return m.handleVerticalNavigation(false)
	case "g":
		return m.handleJumpToTop()
	case "G":
		return m.handleJumpToBottom()
	}

	return nil
}

func (m *Model) handlePaneNavigation(navFunc func()) tea.Cmd {
	if m.focus == FocusLeftPanes {
		navFunc()
	}
	return tea.Batch()
}

func (m *Model) handleVerticalNavigation(down bool) tea.Cmd {
	if m.focus == FocusDetails {
		if down {
			m.details.MoveDown()
		} else {
			m.details.MoveUp()
		}
		m.details.AdjustScroll(m.height - 5)
		return tea.Batch()
	}

	// When focus is on left panes and not on greeting pane, allow scrolling details
	if m.activePane != 3 {
		if down {
			m.details.ScrollDown(m.height - 5)
		} else {
			m.details.ScrollUp()
		}
		return tea.Batch()
	}

	return nil
}

func (m *Model) handleJumpToTop() tea.Cmd {
	if m.focus == FocusDetails {
		m.details.MoveToTop()
		return tea.Batch()
	}
	return nil
}

func (m *Model) handleJumpToBottom() tea.Cmd {
	if m.focus == FocusDetails {
		m.details.MoveToBottom()
		m.details.AdjustScroll(m.height - 5)
		return tea.Batch()
	}
	return nil
}

func (m *Model) toggleFocus() {
	if m.focus == FocusLeftPanes {
		m.focus = FocusDetails
		m.details.Reset()
	} else {
		m.focus = FocusLeftPanes
	}
}

func (m *Model) nextPane() {
	m.setActivePane((m.activePane + 1) % len(m.panes))
}

func (m *Model) prevPane() {
	m.setActivePane((m.activePane - 1 + len(m.panes)) % len(m.panes))
}

func (m *Model) setActivePane(index int) {
	if index >= 0 && index < len(m.panes) {
		m.activePane = index
		for i, pane := range m.panes {
			pane.SetActive(i == index)
		}
	}
}

func (m *Model) refreshAll() tea.Cmd {
	var cmds []tea.Cmd
	for _, pane := range m.panes {
		if cmd := pane.Refresh(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	leftPaneWidth := (m.width * 2) / 3
	rightPaneWidth := m.width - leftPaneWidth

	availableHeight := m.height - 1
	leftPaneHeight := availableHeight / len(m.panes)

	if m.activePane >= len(m.panes) {
		m.activePane = 0
	}

	for i, pane := range m.panes {
		pane.SetActive(i == m.activePane)
	}

	m.updateDiffContent()

	return m.renderLayout(leftPaneWidth, rightPaneWidth, leftPaneHeight)
}

func (m *Model) updateDiffContent() {
	if m.activePane >= len(m.panes) {
		m.details.lines = []string{"No pane selected"}
		return
	}

	activePane := m.panes[m.activePane]
	selectedItem := activePane.GetSelectedItem()

	if selectedItem == nil {
		m.details.lines = []string{"Select an item to see details"}
		return
	}

	// Show details of the selected item
	var details []string
	details = append(details, "Selected Item Details:")
	details = append(details, "")
	details = append(details, fmt.Sprintf("Name: %s", selectedItem.Display))
	details = append(details, fmt.Sprintf("Value: %s", selectedItem.Value))
	details = append(details, fmt.Sprintf("Type: %s", selectedItem.Type))
	details = append(details, "")

	// Add pane-specific details
	paneName := activePane.GetTitle()
	details = append(details, fmt.Sprintf("From: %s pane", paneName))

	m.details.lines = details
}

func (m *Model) GetDiffLines() []string {
	return m.details.lines
}

func (m *Model) GetDiffScrollPos() int {
	return m.details.scrollPos
}

func (m *Model) GetPreviewLines() []string {
	return m.details.lines
}

func (m *Model) GetPreviewScrollPos() int {
	return m.details.scrollPos
}

func (m *Model) IsFocusOnDetails() bool {
	return m.focus == FocusDetails
}

func (m *Model) GetActivePane() panes.Pane {
	if m.activePane < len(m.panes) {
		return m.panes[m.activePane]
	}
	return nil
}

func (m *Model) GetPanes() []panes.Pane {
	return m.panes
}

func (m *Model) GetDimensions() (int, int) {
	return m.width, m.height
}
