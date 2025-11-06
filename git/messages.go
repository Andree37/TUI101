package git

import "time"

// StatusUpdateMsg is sent when git status is updated
type StatusUpdateMsg struct {
	Status *Status
}

// FilesUpdateMsg is sent when file list is updated
type FilesUpdateMsg struct {
	Files []FileInfo
	Path  string
}

// CommitsUpdateMsg is sent when commit list is updated
type CommitsUpdateMsg struct {
	Commits []Commit
}

// BranchesUpdateMsg is sent when branch list is updated
type BranchesUpdateMsg struct {
	Branches []Branch
}

// StashUpdateMsg is sent when stash list is updated
type StashUpdateMsg struct {
	Stashes []string
}

// DiffUpdateMsg is sent when diff content is updated
type DiffUpdateMsg struct {
	Diff string
	File string
}

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
	Error error
}

// RefreshMsg is sent to trigger a refresh
type RefreshMsg struct {
	Timestamp time.Time
}

// ActionCompleteMsg is sent when an action is completed
type ActionCompleteMsg struct {
	Action  string
	Success bool
	Message string
}

// LoadingMsg is sent to indicate loading state
type LoadingMsg struct {
	Pane    string
	Loading bool
}
