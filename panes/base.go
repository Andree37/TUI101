package panes

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// PaneType represents different types of panes
type PaneType int

const (
	StatusPaneType PaneType = iota
	FilesPaneType
	BranchesPaneType
	CommitsPaneType
	StashPaneType
	DiffPaneType
)

// PaneItem represents an item within a pane
type PaneItem struct {
	Display  string      // What to show to the user
	Value    string      // Underlying value (file path, commit hash, etc.)
	Icon     string      // Icon to display
	Type     string      // Type of item (file, directory, commit, etc.)
	Selected bool        // Whether this item is selected
	Metadata interface{} // Additional metadata
	Color    string      // Optional color override
}

// Pane interface defines the contract for all pane types
type Pane interface {
	// Core lifecycle methods
	Init() tea.Cmd
	Update(tea.Msg) (Pane, tea.Cmd)
	View() string

	// Pane identification
	GetTitle() string
	GetType() PaneType
	GetID() string

	// Selection and navigation
	GetSelectedItem() *PaneItem
	GetItems() []PaneItem
	GetItemCount() int
	MoveUp()
	MoveDown()
	MoveToTop()
	MoveToBottom()
	SelectItem(index int)

	// State management
	IsActive() bool
	SetActive(bool)
	IsLoading() bool
	SetLoading(bool)

	// Data operations
	Refresh() tea.Cmd
	Filter(string) []PaneItem
	Clear()
	AddItem(PaneItem)
	RemoveItem(index int)

	// Actions
	HandleAction(action string) tea.Cmd
	GetAvailableActions() []string

	// Display options
	ShowLineNumbers() bool
	SetShowLineNumbers(bool)
	GetMaxDisplayItems() int
	SetMaxDisplayItems(int)
}

// BasePaneModel provides common functionality for all panes
type BasePaneModel struct {
	title           string
	paneType        PaneType
	id              string
	items           []PaneItem
	selectedIndex   int
	active          bool
	loading         bool
	showLineNumbers bool
	maxDisplayItems int
	filter          string
	scrollOffset    int
}

// NewBasePaneModel creates a new base pane model
func NewBasePaneModel(title string, paneType PaneType, id string) BasePaneModel {
	return BasePaneModel{
		title:           title,
		paneType:        paneType,
		id:              id,
		items:           []PaneItem{},
		selectedIndex:   0,
		active:          false,
		loading:         false,
		showLineNumbers: false,
		maxDisplayItems: 50,
	}
}

// GetTitle returns the pane title
func (b *BasePaneModel) GetTitle() string {
	return b.title
}

// GetType returns the pane type
func (b *BasePaneModel) GetType() PaneType {
	return b.paneType
}

// GetID returns the pane ID
func (b *BasePaneModel) GetID() string {
	return b.id
}

// GetSelectedItem returns the currently selected item
func (b *BasePaneModel) GetSelectedItem() *PaneItem {
	if len(b.items) == 0 || b.selectedIndex >= len(b.items) || b.selectedIndex < 0 {
		return nil
	}
	return &b.items[b.selectedIndex]
}

// GetItems returns all items
func (b *BasePaneModel) GetItems() []PaneItem {
	return b.items
}

// GetItemCount returns the number of items
func (b *BasePaneModel) GetItemCount() int {
	return len(b.items)
}

// MoveUp moves selection up
func (b *BasePaneModel) MoveUp() {
	if len(b.items) == 0 {
		return
	}
	if b.selectedIndex > 0 {
		b.selectedIndex--
	} else {
		b.selectedIndex = len(b.items) - 1
	}
	b.adjustScrollOffset()
}

// MoveDown moves selection down
func (b *BasePaneModel) MoveDown() {
	if len(b.items) == 0 {
		return
	}
	if b.selectedIndex < len(b.items)-1 {
		b.selectedIndex++
	} else {
		b.selectedIndex = 0
	}
	b.adjustScrollOffset()
}

// MoveToTop moves selection to the first item
func (b *BasePaneModel) MoveToTop() {
	b.selectedIndex = 0
	b.scrollOffset = 0
}

// MoveToBottom moves selection to the last item
func (b *BasePaneModel) MoveToBottom() {
	if len(b.items) > 0 {
		b.selectedIndex = len(b.items) - 1
		b.adjustScrollOffset()
	}
}

// SelectItem selects an item by index
func (b *BasePaneModel) SelectItem(index int) {
	if index >= 0 && index < len(b.items) {
		b.selectedIndex = index
		b.adjustScrollOffset()
	}
}

// IsActive returns whether the pane is active
func (b *BasePaneModel) IsActive() bool {
	return b.active
}

// SetActive sets the active state
func (b *BasePaneModel) SetActive(active bool) {
	b.active = active
}

// IsLoading returns whether the pane is loading
func (b *BasePaneModel) IsLoading() bool {
	return b.loading
}

// SetLoading sets the loading state
func (b *BasePaneModel) SetLoading(loading bool) {
	b.loading = loading
}

// Clear clears all items
func (b *BasePaneModel) Clear() {
	b.items = []PaneItem{}
	b.selectedIndex = 0
	b.scrollOffset = 0
}

// AddItem adds an item to the pane
func (b *BasePaneModel) AddItem(item PaneItem) {
	b.items = append(b.items, item)
}

// RemoveItem removes an item by index
func (b *BasePaneModel) RemoveItem(index int) {
	if index >= 0 && index < len(b.items) {
		b.items = append(b.items[:index], b.items[index+1:]...)
		if b.selectedIndex >= len(b.items) && len(b.items) > 0 {
			b.selectedIndex = len(b.items) - 1
		}
		b.adjustScrollOffset()
	}
}

// Filter filters items based on a query string
func (b *BasePaneModel) Filter(query string) []PaneItem {
	if query == "" {
		return b.items
	}

	var filtered []PaneItem
	for _, item := range b.items {
		// Simple case-insensitive substring match
		if containsIgnoreCase(item.Display, query) || containsIgnoreCase(item.Value, query) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// ShowLineNumbers returns whether to show line numbers
func (b *BasePaneModel) ShowLineNumbers() bool {
	return b.showLineNumbers
}

// SetShowLineNumbers sets whether to show line numbers
func (b *BasePaneModel) SetShowLineNumbers(show bool) {
	b.showLineNumbers = show
}

// GetMaxDisplayItems returns the maximum number of items to display
func (b *BasePaneModel) GetMaxDisplayItems() int {
	return b.maxDisplayItems
}

// SetMaxDisplayItems sets the maximum number of items to display
func (b *BasePaneModel) SetMaxDisplayItems(max int) {
	b.maxDisplayItems = max
}

// GetVisibleItems returns the items that should be visible based on scroll offset
func (b *BasePaneModel) GetVisibleItems() []PaneItem {
	if len(b.items) == 0 {
		return []PaneItem{}
	}

	start := b.scrollOffset
	end := start + b.maxDisplayItems
	if end > len(b.items) {
		end = len(b.items)
	}
	if start > end {
		start = end
	}

	return b.items[start:end]
}

// adjustScrollOffset adjusts the scroll offset to keep selected item visible
func (b *BasePaneModel) adjustScrollOffset() {
	if len(b.items) == 0 {
		b.scrollOffset = 0
		return
	}

	// If selected item is above visible area, scroll up
	if b.selectedIndex < b.scrollOffset {
		b.scrollOffset = b.selectedIndex
	}

	// If selected item is below visible area, scroll down
	if b.selectedIndex >= b.scrollOffset+b.maxDisplayItems {
		b.scrollOffset = b.selectedIndex - b.maxDisplayItems + 1
	}

	// Ensure scroll offset is not negative
	if b.scrollOffset < 0 {
		b.scrollOffset = 0
	}
}

// GetSelectedIndex returns the currently selected index
func (b *BasePaneModel) GetSelectedIndex() int {
	return b.selectedIndex
}

// GetScrollOffset returns the current scroll offset
func (b *BasePaneModel) GetScrollOffset() int {
	return b.scrollOffset
}

// containsIgnoreCase performs case-insensitive substring matching
func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}
