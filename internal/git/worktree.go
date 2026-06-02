package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type WorktreeStatus struct {
	Modified  int
	Added     int
	Deleted   int
	Untracked int
	IsDirty   bool
}

type Worktree struct {
	Path       string
	Branch     string
	Head       string
	IsBare     bool
	IsDetached bool
	IsCurrent  bool
	IsPrimary  bool
	Status     WorktreeStatus
	Ahead      int
	Behind     int
}

func GetRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

func GetCurrentBranch() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func GetWorktreeStatus(path string) WorktreeStatus {
	cmd := exec.Command("git", "-C", path, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return WorktreeStatus{}
	}

	var status WorktreeStatus
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) < 2 {
			continue
		}
		switch {
		case strings.HasPrefix(line, "??"):
			status.Untracked++
		case line[0] == 'A' || line[1] == 'A':
			status.Added++
		case line[0] == 'D' || line[1] == 'D':
			status.Deleted++
		case line[0] == 'M' || line[1] == 'M':
			status.Modified++
		}
	}
	status.IsDirty = status.Modified+status.Added+status.Deleted+status.Untracked > 0
	return status
}

func GetAheadBehind(path, branch string) (int, int) {
	cmd := exec.Command("git", "-C", path, "rev-list", "--left-right", "--count", branch+"...origin/"+branch)
	out, err := cmd.Output()
	if err != nil {
		return 0, 0
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) != 2 {
		return 0, 0
	}
	var ahead, behind int
	fmt.Sscanf(parts[0], "%d", &ahead)
	fmt.Sscanf(parts[1], "%d", &behind)
	return ahead, behind
}

func ListWorktrees() ([]Worktree, error) {
	out, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	currentBranch, _ := GetCurrentBranch()
	repoRoot, _ := GetRepoRoot()

	var worktrees []Worktree
	var current Worktree

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "worktree "):
			current = Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "HEAD "):
			current.Head = strings.TrimPrefix(line, "HEAD ")[:7]
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "bare":
			current.IsBare = true
		case line == "detached":
			current.IsDetached = true
			current.Branch = "(detached)"
		case line == "":
			if current.Path != "" {
				if current.Branch == currentBranch && current.Path == repoRoot {
					current.IsCurrent = true
				}
				if len(worktrees) == 0 {
					current.IsPrimary = true
				}
				current.Status = GetWorktreeStatus(current.Path)
				if current.Branch != "" && current.Branch != "(detached)" {
					current.Ahead, current.Behind = GetAheadBehind(current.Path, current.Branch)
				}
				worktrees = append(worktrees, current)
				current = Worktree{}
			}
		}
	}

	return worktrees, nil
}

func ListBranches() ([]string, error) {
	out, err := exec.Command("git", "branch", "--format=%(refname:short)").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var branches []string
	for _, l := range lines {
		if l = strings.TrimSpace(l); l != "" {
			branches = append(branches, l)
		}
	}
	return branches, nil
}

func ListRemoteBranches() ([]string, error) {
	out, err := exec.Command("git", "branch", "-r", "--format=%(refname:short)").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var branches []string
	for _, l := range lines {
		if l = strings.TrimSpace(l); l != "" && !strings.Contains(l, "HEAD") {
			branches = append(branches, l)
		}
	}
	return branches, nil
}

func DefaultWorktreePath(branch string) string {
	root, err := GetRepoRoot()
	if err != nil {
		return branch
	}
	parent := filepath.Dir(root)
	repoName := filepath.Base(root)
	safeBranch := strings.ReplaceAll(branch, "/", "-")
	return filepath.Join(parent, repoName+"-"+safeBranch)
}

func AddWorktree(path, branch string, createBranch bool) error {
	if path == "" {
		path = DefaultWorktreePath(branch)
	} else if !filepath.IsAbs(path) {
		root, err := GetRepoRoot()
		if err != nil {
			return err
		}
		path = filepath.Join(filepath.Dir(root), path)
	}

	args := []string{"worktree", "add"}
	if createBranch {
		args = append(args, "-b", branch, path)
	} else {
		args = append(args, path, branch)
	}

	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func RemoveWorktree(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func PruneWorktrees() error {
	out, err := exec.Command("git", "worktree", "prune").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func FetchRemote() error {
	out, err := exec.Command("git", "fetch", "--all", "--prune").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func OpenShell(path string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := exec.Command("open", "-a", "Terminal", path)
	if err := cmd.Run(); err != nil {
		cmd = exec.Command(shell, "-c", fmt.Sprintf("cd %q && exec %s", path, shell))
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return nil
}

func OpenInEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "code"
	}
	cmd := exec.Command(editor, path)
	return cmd.Start()
}

func WorktreeExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func GetLastCommitMessage(path string) string {
	cmd := exec.Command("git", "-C", path, "log", "-1", "--format=%s")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func GetLastCommitTime(path string) string {
	cmd := exec.Command("git", "-C", path, "log", "-1", "--format=%cr")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func PickFolder() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("osascript", "-e",
			`set chosenFolder to POSIX path of (choose folder with prompt "Choose worktree location")`,
		).Output()
		if err != nil {
			return "", fmt.Errorf("folder picker cancelled")
		}
		return strings.TrimSpace(string(out)), nil
	case "linux":
		out, err := exec.Command("zenity", "--file-selection", "--directory", "--title=Choose worktree location").Output()
		if err != nil {
			return "", fmt.Errorf("folder picker cancelled")
		}
		return strings.TrimSpace(string(out)), nil
	case "windows":
		script := `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.FolderBrowserDialog; $f.Description = "Choose worktree location"; if ($f.ShowDialog() -eq "OK") { $f.SelectedPath } else { exit 1 }`
		out, err := exec.Command("powershell", "-Command", script).Output()
		if err != nil {
			return "", fmt.Errorf("folder picker cancelled")
		}
		return strings.TrimSpace(string(out)), nil
	default:
		return "", fmt.Errorf("folder picker not supported on this OS")
	}
}
