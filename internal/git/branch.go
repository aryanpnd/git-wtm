package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type Branch struct {
	Name      string
	IsCurrent bool
	IsRemote  bool
	LastCommit string
	CommitTime string
	Upstream   string
	Ahead      int
	Behind     int
}

func ListBranchesDetailed() ([]Branch, error) {
	out, err := exec.Command("git", "branch", "-vv", "--format=%(refname:short)\t%(HEAD)\t%(upstream:short)\t%(objectname:short)\t%(subject)\t%(creatordate:relative)").Output()
	if err != nil {
		return nil, err
	}

	var branches []Branch
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 6)
		if len(parts) < 6 {
			continue
		}

		b := Branch{
			Name:       parts[0],
			IsCurrent:  parts[1] == "*",
			Upstream:   parts[2],
			LastCommit: parts[3],
			CommitTime: parts[5],
		}

		subject := parts[4]
		if len(subject) > 50 {
			subject = subject[:47] + "..."
		}
		b.LastCommit = parts[3] + " " + subject

		if b.Upstream != "" {
			b.Ahead, b.Behind = GetAheadBehind(".", b.Name)
		}

		branches = append(branches, b)
	}
	return branches, nil
}

func CreateBranch(name, from string) error {
	args := []string{"branch", name}
	if from != "" {
		args = append(args, from)
	}
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func DeleteBranch(name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	out, err := exec.Command("git", "branch", flag, name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func RenameBranch(oldName, newName string) error {
	out, err := exec.Command("git", "branch", "-m", oldName, newName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func CheckoutBranch(name string) error {
	out, err := exec.Command("git", "checkout", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func MergeBranch(name string) error {
	out, err := exec.Command("git", "merge", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}
