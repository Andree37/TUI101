package panes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tui101/git"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FilesPane struct {
	BasePaneModel
	gitRepo     *git.Repository
	currentPath string
	showHidden  bool
	maxFiles    int
}

func NewFilesPane() *FilesPane {
	base := NewBasePaneModel("Files", FilesPaneType, "files")

	pane := &FilesPane{
		BasePaneModel: base,
		gitRepo:       git.NewRepository("."),
		currentPath:   ".",
		showHidden:    false,
		maxFiles:      100, // Limit files to prevent crashes
	}

	pane.loadFiles()
	return pane
}

func (f *FilesPane) Init() tea.Cmd {
	return f.Refresh()
}

func (f *FilesPane) Update(msg tea.Msg) (Pane, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !f.IsActive() {
			return f, nil
		}

		switch msg.String() {
		case "j", "down":
			f.MoveDown()
		case "k", "up":
			f.MoveUp()
		case "g":
			f.MoveToTop()
		case "G":
			f.MoveToBottom()
		case "enter":
			return f, f.HandleAction("open")
		case "h", "left":
			return f, f.HandleAction("up_directory")
		case "l", "right":
			return f, f.HandleAction("enter_directory")
		case ".":
			f.showHidden = !f.showHidden
			return f, f.Refresh()
		case "r":
			return f, f.Refresh()
		case "a":
			return f, f.HandleAction("stage")
		case "u":
			return f, f.HandleAction("unstage")
		case "d":
			return f, f.HandleAction("diff")
		}

	case git.FilesUpdateMsg:
		f.updateFromFilesMsg(msg)
		return f, nil
	}

	return f, nil
}

func (f *FilesPane) View() string {
	if f.IsLoading() {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#74B9FF")).
			Render("Loading files...")
	}

	if len(f.items) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#74B9FF")).
			Render("No files in directory")
	}

	var lines []string
	visibleItems := f.GetVisibleItems()

	for i, item := range visibleItems {
		actualIndex := f.GetScrollOffset() + i
		isSelected := actualIndex == f.GetSelectedIndex()

		line := f.formatFileItem(item, isSelected)
		lines = append(lines, line)
	}

	if f.GetScrollOffset() > 0 {
		lines = append([]string{"  ↑ more items above"}, lines...)
	}
	if f.GetScrollOffset()+len(visibleItems) < len(f.items) {
		lines = append(lines, "  ↓ more items below")
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (f *FilesPane) formatFileItem(item PaneItem, isSelected bool) string {

	gitStatus := f.getGitStatus(item.Value)
	if gitStatus == " " {
		gitStatus = ""
	}

	var line string
	if gitStatus != "" {
		line = fmt.Sprintf("%s %s", gitStatus, item.Display)
	} else {
		line = fmt.Sprintf("  %s", item.Display)
	}

	style := lipgloss.NewStyle()

	if isSelected {
		style = style.Background(lipgloss.Color("#2D3748")).
			Foreground(lipgloss.Color("#04B575"))
	} else {

		switch item.Type {
		case "directory":
			style = style.Foreground(lipgloss.Color("#74B9FF"))
		case "modified":
			style = style.Foreground(lipgloss.Color("#FFEAA7"))
		case "added":
			style = style.Foreground(lipgloss.Color("#04B575"))
		case "deleted":
			style = style.Foreground(lipgloss.Color("#E53E3E"))
		case "untracked":
			style = style.Foreground(lipgloss.Color("#F25D94"))
		default:
			style = style.Foreground(lipgloss.Color("#DDD6FE"))
		}
	}

	return style.Render(line)
}

func (f *FilesPane) getGitStatus(filename string) string {
	if f.gitRepo == nil {
		return " "
	}

	status := f.gitRepo.GetFileStatus(filename)
	switch status {
	case "modified":
		return "M"
	case "added":
		return "A"
	case "deleted":
		return "D"
	case "renamed":
		return "R"
	case "copied":
		return "C"
	case "untracked":
		return "?"
	case "ignored":
		return "!"
	default:
		return " "
	}
}

func (f *FilesPane) Refresh() tea.Cmd {
	f.SetLoading(true)
	return func() tea.Msg {
		files := f.loadDirectoryContents()
		return git.FilesUpdateMsg{Files: files, Path: f.currentPath}
	}
}

func (f *FilesPane) HandleAction(action string) tea.Cmd {
	selectedItem := f.GetSelectedItem()
	if selectedItem == nil {
		return nil
	}

	switch action {
	case "open":
		if selectedItem.Type == "directory" {
			return f.enterDirectory(selectedItem.Value)
		}
		return f.openFile(selectedItem.Value)

	case "enter_directory":
		if selectedItem.Type == "directory" {
			return f.enterDirectory(selectedItem.Value)
		}
		return nil

	case "up_directory":
		return f.upDirectory()

	case "stage":
		return f.stageFile(selectedItem.Value)

	case "unstage":
		return f.unstageFile(selectedItem.Value)

	case "diff":
		return f.showDiff(selectedItem.Value)

	default:
		return nil
	}
}

func (f *FilesPane) GetAvailableActions() []string {
	return []string{"open", "stage", "unstage", "diff", "refresh", "toggle_hidden"}
}

func (f *FilesPane) loadFiles() {
	f.Clear()

	if f.currentPath != "." && f.currentPath != "" {
		f.AddItem(PaneItem{
			Display: "../",
			Value:   "..",
			Icon:    "📁",
			Type:    "directory",
		})
	}

	entries, err := os.ReadDir(f.currentPath)
	if err != nil {
		f.AddItem(PaneItem{
			Display: fmt.Sprintf("Error reading directory: %s", err),
			Value:   "",
			Type:    "error",
		})
		return
	}

	if len(entries) > f.maxFiles {
		f.AddItem(PaneItem{
			Display: fmt.Sprintf("Directory has %d items (showing first %d)", len(entries), f.maxFiles),
			Value:   "",
			Type:    "info",
		})
		entries = entries[:f.maxFiles]
	}

	var directories, files []PaneItem

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless showHidden is true
		if !f.showHidden && strings.HasPrefix(name, ".") {
			continue
		}

		if strings.ContainsAny(name, "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x0b\x0c\x0e\x0f") {
			continue
		}

		fullPath := filepath.Join(f.currentPath, name)

		itemType := f.getFileTypeSafe(fullPath)

		if entry.IsDir() {
			directories = append(directories, PaneItem{
				Display: name + "/",
				Value:   fullPath,
				Type:    "directory",
			})
		} else {
			files = append(files, PaneItem{
				Display: name,
				Value:   fullPath,
				Type:    itemType,
			})
		}
	}

	for _, dir := range directories {
		f.AddItem(dir)
	}
	for _, file := range files {
		f.AddItem(file)
	}
}

func (f *FilesPane) loadDirectoryContents() []git.FileInfo {
	var files []git.FileInfo

	entries, err := os.ReadDir(f.currentPath)
	if err != nil {
		return files
	}

	if len(entries) > f.maxFiles {
		entries = entries[:f.maxFiles]
	}

	for _, entry := range entries {
		name := entry.Name()
		if !f.showHidden && strings.HasPrefix(name, ".") {
			continue
		}

		if strings.ContainsAny(name, "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x0b\x0c\x0e\x0f") {
			continue
		}

		fullPath := filepath.Join(f.currentPath, name)

		status := ""
		if f.gitRepo != nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
						status = "unknown"
					}
				}()
				status = f.gitRepo.GetFileStatus(fullPath)
			}()
		}

		fileInfo := git.FileInfo{
			Name:     name,
			Path:     fullPath,
			IsDir:    entry.IsDir(),
			Status:   status,
			Modified: false,
		}

		files = append(files, fileInfo)
	}

	return files
}

func (f *FilesPane) getFileIcon(filename string) string {
	return ""
}

func (f *FilesPane) getFileType(filepath string) string {
	status := f.gitRepo.GetFileStatus(filepath)
	if status != "" {
		return status
	}
	return "tracked"
}

func (f *FilesPane) getFileTypeSafe(filepath string) string {
	defer func() {
		if r := recover(); r != nil {

		}
	}()

	if f.gitRepo != nil {
		status := f.gitRepo.GetFileStatus(filepath)
		if status != "" {
			return status
		}
	}
	return "untracked"
}

func (f *FilesPane) enterDirectory(dirPath string) tea.Cmd {
	if dirPath == ".." {
		return f.upDirectory()
	}

	f.currentPath = dirPath
	f.selectedIndex = 0
	f.scrollOffset = 0
	return f.Refresh()
}

func (f *FilesPane) upDirectory() tea.Cmd {
	if f.currentPath == "." || f.currentPath == "" {
		return nil
	}

	f.currentPath = filepath.Dir(f.currentPath)
	if f.currentPath == "/" || f.currentPath == "\\" {
		f.currentPath = "."
	}

	f.selectedIndex = 0
	f.scrollOffset = 0
	return f.Refresh()
}

func (f *FilesPane) openFile(filepath string) tea.Cmd {

	return nil
}

func (f *FilesPane) stageFile(filepath string) tea.Cmd {
	return func() tea.Msg {
		err := f.gitRepo.StageFile(filepath)
		if err != nil {
			return git.ErrorMsg{Error: err}
		}
		return git.FilesUpdateMsg{Files: f.loadDirectoryContents(), Path: f.currentPath}
	}
}

func (f *FilesPane) unstageFile(filepath string) tea.Cmd {
	return func() tea.Msg {
		err := f.gitRepo.UnstageFile(filepath)
		if err != nil {
			return git.ErrorMsg{Error: err}
		}
		return git.FilesUpdateMsg{Files: f.loadDirectoryContents(), Path: f.currentPath}
	}
}

func (f *FilesPane) showDiff(filepath string) tea.Cmd {
	return func() tea.Msg {
		diff := f.gitRepo.GetFileDiff(filepath)
		return git.DiffUpdateMsg{Diff: diff, File: filepath}
	}
}

func (f *FilesPane) updateFromFilesMsg(msg git.FilesUpdateMsg) {
	f.SetLoading(false)
	f.Clear()

	if f.currentPath != "." && f.currentPath != "" {
		f.AddItem(PaneItem{
			Display: "../",
			Value:   "..",
			Icon:    "📁",
			Type:    "directory",
		})
	}

	for _, fileInfo := range msg.Files {
		icon := "📄"
		if fileInfo.IsDir {
			icon = "📁"
		} else {
			icon = f.getFileIcon(fileInfo.Name)
		}

		display := fileInfo.Name
		if fileInfo.IsDir {
			display += "/"
		}

		itemType := "tracked"
		if fileInfo.Status != "" {
			itemType = fileInfo.Status
		}
		if fileInfo.IsDir {
			itemType = "directory"
		}

		f.AddItem(PaneItem{
			Display: display,
			Value:   fileInfo.Path,
			Icon:    icon,
			Type:    itemType,
		})
	}
}
