package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Repository struct {
	path string
}

type Status struct {
	Branch         string
	Upstream       string
	AheadBy        int
	BehindBy       int
	ModifiedFiles  int
	UntrackedFiles int
	StagedFiles    int
	Dirty          bool
}

type FileInfo struct {
	Name     string
	Path     string
	IsDir    bool
	Status   string
	Modified bool
}

type Commit struct {
	Hash      string
	Author    string
	Message   string
	Date      time.Time
	ShortHash string
}

type Branch struct {
	Name      string
	IsCurrent bool
	IsRemote  bool
	Upstream  string
}

func NewRepository(path string) *Repository {
	return &Repository{path: path}
}

func (r *Repository) GetCurrentBranch() string {
	cmd := exec.Command("git", "-C", r.path, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func (r *Repository) GetStatus() *Status {
	status := &Status{
		Branch: r.GetCurrentBranch(),
	}

	cmd := exec.Command("git", "-C", r.path, "status", "--porcelain=v1")
	output, err := cmd.Output()
	if err != nil {
		return status
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		indexStatus := line[0]
		workTreeStatus := line[1]

		if indexStatus != ' ' && indexStatus != '?' {
			status.StagedFiles++
		}

		if workTreeStatus != ' ' {
			if workTreeStatus == '?' {
				status.UntrackedFiles++
			} else {
				status.ModifiedFiles++
			}
		}
	}

	status.Dirty = status.ModifiedFiles > 0 || status.UntrackedFiles > 0 || status.StagedFiles > 0

	cmd = exec.Command("git", "-C", r.path, "rev-parse", "--abbrev-ref", "@{upstream}")
	output, err = cmd.Output()
	if err == nil {
		status.Upstream = strings.TrimSpace(string(output))

		cmd = exec.Command("git", "-C", r.path, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
		output, err = cmd.Output()
		if err == nil {
			parts := strings.Fields(string(output))
			if len(parts) == 2 {

				if parts[0] != "0" {
					status.AheadBy = 1
				}
				if parts[1] != "0" {
					status.BehindBy = 1
				}
			}
		}
	}

	return status
}

func (s *Status) HasChanges() bool {
	return s.Dirty
}

func (r *Repository) GetFileStatus(filepath string) string {
	cmd := exec.Command("git", "-C", r.path, "status", "--porcelain", filepath)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	line := strings.TrimSpace(string(output))
	if len(line) < 2 {
		return ""
	}

	indexStatus := line[0]
	workTreeStatus := line[1]

	switch {
	case workTreeStatus == '?':
		return "untracked"
	case workTreeStatus == 'M':
		return "modified"
	case workTreeStatus == 'D':
		return "deleted"
	case indexStatus == 'A':
		return "added"
	case indexStatus == 'M':
		return "staged"
	case indexStatus == 'D':
		return "staged_deleted"
	case indexStatus == 'R':
		return "renamed"
	case indexStatus == 'C':
		return "copied"
	default:
		return ""
	}
}

func (r *Repository) GetCommits(limit int) []Commit {
	cmd := exec.Command("git", "-C", r.path, "log", "--oneline", "-n", fmt.Sprintf("%d", limit))
	output, err := cmd.Output()
	if err != nil {
		return []Commit{}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	commits := make([]Commit, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		commit := Commit{
			ShortHash: parts[0],
			Hash:      parts[0],
			Message:   parts[1],
			Author:    "AR",
			Date:      time.Now(),
		}

		commits = append(commits, commit)
	}

	return commits
}

func (r *Repository) GetBranches() []Branch {
	var branches []Branch

	cmd := exec.Command("git", "-C", r.path, "branch")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			isCurrent := strings.HasPrefix(line, "* ")
			name := strings.TrimPrefix(line, "* ")
			name = strings.TrimSpace(name)

			branches = append(branches, Branch{
				Name:      name,
				IsCurrent: isCurrent,
				IsRemote:  false,
			})
		}
	}

	cmd = exec.Command("git", "-C", r.path, "branch", "-r")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.Contains(line, "->") {
				continue
			}

			branches = append(branches, Branch{
				Name:     line,
				IsRemote: true,
			})
		}
	}

	return branches
}

func (r *Repository) GetUpstreamInfo() string {
	cmd := exec.Command("git", "-C", r.path, "rev-parse", "--abbrev-ref", "@{upstream}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func (r *Repository) StageFile(filepath string) error {
	cmd := exec.Command("git", "-C", r.path, "add", filepath)
	return cmd.Run()
}

func (r *Repository) UnstageFile(filepath string) error {
	cmd := exec.Command("git", "-C", r.path, "reset", "HEAD", filepath)
	return cmd.Run()
}

func (r *Repository) GetFileDiff(filepath string) string {

	if info, err := os.Stat(filepath); err == nil {
		const maxFileSize = 1024 * 1024 // 1MB limit
		if info.Size() > maxFileSize {
			return fmt.Sprintf("File too large (%d bytes) - diff not shown", info.Size())
		}
	}

	cmd := exec.Command("git", "-C", r.path, "diff", filepath)
	output, err := cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		return r.truncateDiffOutput(string(output))
	}

	cmd = exec.Command("git", "-C", r.path, "diff", "--cached", filepath)
	output, err = cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		return r.truncateDiffOutput(string(output))
	}

	status := r.GetFileStatus(filepath)
	if status == "untracked" || status == "" {
		return r.getFileContentAsAddition(filepath)
	}

	return ""
}

func (r *Repository) GetCommitDiff(commitHash string) string {
	cmd := exec.Command("git", "-C", r.path, "show", commitHash)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

func (r *Repository) getFileContentAsAddition(filepath string) string {

	if info, err := os.Stat(filepath); err != nil || info.IsDir() {
		return ""
	}

	const maxFileSize = 1024 * 1024
	if info, err := os.Stat(filepath); err == nil && info.Size() > maxFileSize {
		return fmt.Sprintf("File too large (%d bytes) - showing first %d bytes only\n\n", info.Size(), maxFileSize/10)
	}

	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Sprintf("Error reading file: %s", err.Error())
	}

	if len(content) > maxFileSize {
		content = content[:maxFileSize/10] // Show only 10% of max size
	}

	sanitized := r.sanitizeContent(string(content))
	lines := strings.Split(sanitized, "\n")

	const maxLines = 100
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, fmt.Sprintf("... (file truncated, showing first %d lines)", maxLines))
	}

	var diffLines []string

	diffLines = append(diffLines, fmt.Sprintf("diff --git a/%s b/%s", filepath, filepath))
	diffLines = append(diffLines, "new file mode 100644")
	diffLines = append(diffLines, "index 0000000..0000000")
	diffLines = append(diffLines, "--- /dev/null")
	diffLines = append(diffLines, fmt.Sprintf("+++ b/%s", filepath))
	diffLines = append(diffLines, fmt.Sprintf("@@ -0,0 +1,%d @@", len(lines)))

	for _, line := range lines {

		if len(line) > 200 {
			line = line[:200] + "... (line truncated)"
		}
		diffLines = append(diffLines, "+"+line)
	}

	return strings.Join(diffLines, "\n")
}

func (r *Repository) truncateDiffOutput(output string) string {
	lines := strings.Split(output, "\n")
	const maxLines = 200

	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, fmt.Sprintf("... (diff truncated, showing first %d lines)", maxLines))
	}

	for i, line := range lines {
		if len(line) > 500 {
			lines[i] = line[:500] + "... (line truncated)"
		}
	}

	return strings.Join(lines, "\n")
}

func (r *Repository) sanitizeContent(content string) string {

	if r.isBinaryContent(content) {
		return fmt.Sprintf("Binary file (%d bytes) - content not shown", len(content))
	}

	var result strings.Builder
	for _, r := range content {
		if r >= 32 && r < 127 || r == '\n' || r == '\t' {
			result.WriteRune(r)
		} else if r > 127 {
			result.WriteRune(r)
		} else {
			result.WriteString(fmt.Sprintf("\\x%02x", r))
		}
	}
	return result.String()
}

func (r *Repository) isBinaryContent(content string) bool {

	nonPrintable := 0
	total := 0

	for _, r := range content {
		total++
		if r < 32 && r != '\n' && r != '\t' && r != '\r' {
			nonPrintable++
		}

		if total > 1000 {
			break
		}
	}

	if total == 0 {
		return false
	}

	return float64(nonPrintable)/float64(total) > 0.30
}

func (r *Repository) Fetch() error {
	cmd := exec.Command("git", "-C", r.path, "fetch")
	return cmd.Run()
}

func (r *Repository) GetStashes() []string {
	cmd := exec.Command("git", "-C", r.path, "stash", "list")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}
	}

	return lines
}
