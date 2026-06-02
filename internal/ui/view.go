package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/aryan/worktree-manager/internal/git"
)

func (m Model) View() string {
	if m.width == 0 {
		return "\n  Loading..."
	}

	var s strings.Builder

	switch m.currentView {
	case viewList:
		s.WriteString(m.viewList())
	case viewAdd:
		s.WriteString(m.viewAdd())
	case viewRemoveConfirm:
		s.WriteString(m.viewRemoveConfirm())
	case viewBranchSelect:
		s.WriteString(m.viewBranchSelect())
	case viewDetail:
		s.WriteString(m.viewDetail())
	}

	return s.String()
}

func (m Model) viewList() string {
	var s strings.Builder

	header := titleStyle.Render(" 🌳 Worktree Manager ")
	if m.loading {
		header += dimStyle.Render("  ⟳ loading...")
	}
	s.WriteString(header + "\n")

	if m.searching {
		s.WriteString("  " + searchStyle.Render("/") + " " + m.searchInput.View() + "\n\n")
	} else if m.searchInput.Value() != "" {
		s.WriteString("  " + dimStyle.Render(fmt.Sprintf("filter: %q (%d results)", m.searchInput.Value(), len(m.filtered))) + "\n\n")
	} else {
		s.WriteString("\n")
	}

	if len(m.worktrees) == 0 && !m.loading {
		s.WriteString(dimStyle.Render("  No worktrees found. Press 'a' to add one.\n"))
	}

	maxVisible := m.height - 12
	if maxVisible < 3 {
		maxVisible = 3
	}

	visibleStart := 0
	if m.cursor >= maxVisible {
		visibleStart = m.cursor - maxVisible + 1
	}

	end := visibleStart + maxVisible
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for vi := visibleStart; vi < end; vi++ {
		idx := m.filtered[vi]
		wt := m.worktrees[idx]

		isSelected := vi == m.cursor
		s.WriteString(m.renderWorktreeItem(wt, isSelected))
	}

	if len(m.filtered) > maxVisible {
		s.WriteString(dimStyle.Render(fmt.Sprintf("  ↕ %d/%d shown\n", maxVisible, len(m.filtered))))
	}

	if m.statusMsg != "" {
		s.WriteString("\n " + successStyle.Render("✓ "+m.statusMsg) + "\n")
	}
	if m.err != nil {
		s.WriteString("\n " + errorStyle.Render("✗ "+m.err.Error()) + "\n")
	}

	s.WriteString("\n")
	if m.showHelp {
		s.WriteString(m.fullHelp())
	} else {
		s.WriteString(m.shortHelp())
	}

	return s.String()
}

func (m Model) renderWorktreeItem(wt git.Worktree, selected bool) string {
	var s strings.Builder

	cursor := "  "
	nameStyle := itemStyle
	if selected {
		cursor = "▸ "
		nameStyle = selectedItemStyle
	}

	name := filepath.Base(wt.Path)
	branch := branchStyle.Render(wt.Branch)
	if wt.IsCurrent {
		branch = currentBadge.Render(wt.Branch + " ●")
	}

	statusIndicator := statusClean.Render("✓")
	if wt.Status.IsDirty {
		parts := []string{}
		if wt.Status.Modified > 0 {
			parts = append(parts, fmt.Sprintf("~%d", wt.Status.Modified))
		}
		if wt.Status.Added > 0 {
			parts = append(parts, fmt.Sprintf("+%d", wt.Status.Added))
		}
		if wt.Status.Deleted > 0 {
			parts = append(parts, fmt.Sprintf("-%d", wt.Status.Deleted))
		}
		if wt.Status.Untracked > 0 {
			parts = append(parts, fmt.Sprintf("?%d", wt.Status.Untracked))
		}
		statusIndicator = statusDirty.Render(strings.Join(parts, " "))
	}

	syncInfo := ""
	if wt.Ahead > 0 {
		syncInfo += aheadStyle.Render(fmt.Sprintf("↑%d", wt.Ahead))
	}
	if wt.Behind > 0 {
		if syncInfo != "" {
			syncInfo += " "
		}
		syncInfo += behindStyle.Render(fmt.Sprintf("↓%d", wt.Behind))
	}

	line1 := fmt.Sprintf("%s%s  %s  %s  %s",
		cursor,
		nameStyle.Render(name),
		branch,
		statusIndicator,
		syncInfo,
	)

	commit := commitStyle.Render(wt.Head)
	lastMsg := git.GetLastCommitMessage(wt.Path)
	if len(lastMsg) > 50 {
		lastMsg = lastMsg[:47] + "..."
	}
	line2 := fmt.Sprintf("     %s %s  %s",
		commit,
		dimStyle.Render(lastMsg),
		pathStyle.Render(shortenPath(wt.Path)),
	)

	s.WriteString(line1 + "\n")
	s.WriteString(line2 + "\n\n")

	return s.String()
}

func (m Model) viewAdd() string {
	var s strings.Builder

	mode := "existing branch"
	if m.createNew {
		mode = "new branch"
	}
	s.WriteString(titleStyle.Render(" Add Worktree ") + "  " + dimStyle.Render("("+mode+")") + "\n\n")

	branchLabel := inactiveInputStyle.Render("  Branch ")
	pathLabel := inactiveInputStyle.Render("  Path   ")
	if m.addStep == 0 {
		branchLabel = activeInputStyle.Render("▸ Branch ")
	} else {
		pathLabel = activeInputStyle.Render("▸ Path   ")
	}

	s.WriteString(fmt.Sprintf("%s %s\n\n", branchLabel, m.branchInput.View()))
	s.WriteString(fmt.Sprintf("%s %s\n", pathLabel, m.pathInput.View()))

	defaultPath := ""
	if m.branchInput.Value() != "" {
		defaultPath = git.DefaultWorktreePath(m.branchInput.Value())
	}
	if defaultPath != "" && m.pathInput.Value() == "" {
		s.WriteString(dimStyle.Render(fmt.Sprintf("         default: %s", shortenPath(defaultPath))) + "\n")
	}

	s.WriteString("\n")
	s.WriteString(m.renderHelpLine([]helpKey{
		{"enter", "confirm"},
		{"tab", "switch field"},
		{"esc", "cancel"},
	}))

	return s.String()
}

func (m Model) viewRemoveConfirm() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(" Remove Worktree ") + "\n\n")

	wt := m.selectedWorktree()
	if wt == nil {
		s.WriteString(dimStyle.Render("  No worktree selected.\n"))
		return s.String()
	}

	s.WriteString(fmt.Sprintf("  %s %s\n\n", detailLabelStyle.Render("Path:"), errorStyle.Render(wt.Path)))
	s.WriteString(fmt.Sprintf("  %s %s\n\n", detailLabelStyle.Render("Branch:"), branchStyle.Render(wt.Branch)))

	if wt.Status.IsDirty {
		s.WriteString(warningStyle.Render("  ⚠ This worktree has uncommitted changes!") + "\n\n")
	}

	forceLabel := dimStyle.Render("off")
	if m.confirmForce {
		forceLabel = errorStyle.Render("ON")
	}
	s.WriteString(fmt.Sprintf("  %s %s\n\n", detailLabelStyle.Render("Force:"), forceLabel))

	s.WriteString(m.renderHelpLine([]helpKey{
		{"y", "confirm"},
		{"n/esc", "cancel"},
		{"f", "toggle force"},
	}))

	return s.String()
}

func (m Model) viewBranchSelect() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(" Select Branch ") + "\n\n")
	s.WriteString("  " + m.branchSearch.View() + "\n\n")

	if len(m.filteredBranch) == 0 {
		s.WriteString(dimStyle.Render("  No matching branches.\n"))
	}

	maxVisible := m.height - 10
	if maxVisible < 5 {
		maxVisible = 5
	}

	visibleStart := 0
	if m.branchCursor >= visibleStart+maxVisible {
		visibleStart = m.branchCursor - maxVisible + 1
	}

	end := visibleStart + maxVisible
	if end > len(m.filteredBranch) {
		end = len(m.filteredBranch)
	}

	for vi := visibleStart; vi < end; vi++ {
		idx := m.filteredBranch[vi]
		branch := m.branches[idx]

		cursor := "  "
		style := itemStyle
		if vi == m.branchCursor {
			cursor = "▸ "
			style = selectedItemStyle
		}
		s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(branch)))
	}

	if len(m.filteredBranch) > maxVisible {
		s.WriteString(dimStyle.Render(fmt.Sprintf("\n  %d/%d branches", maxVisible, len(m.filteredBranch))))
	}

	s.WriteString("\n\n")
	s.WriteString(m.renderHelpLine([]helpKey{
		{"enter", "select"},
		{"↑/↓", "navigate"},
		{"esc", "cancel"},
	}))

	return s.String()
}

func (m Model) viewDetail() string {
	var s strings.Builder

	wt := m.selectedWorktree()
	if wt == nil {
		s.WriteString(dimStyle.Render("  No worktree selected.\n"))
		return s.String()
	}

	s.WriteString(titleStyle.Render(" Worktree Details ") + "\n\n")

	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Branch:"), branchStyle.Render(wt.Branch)))
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Path:"), detailValueStyle.Render(wt.Path)))
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("HEAD:"), commitStyle.Render(wt.Head)))

	lastMsg := git.GetLastCommitMessage(wt.Path)
	lastTime := git.GetLastCommitTime(wt.Path)
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Commit:"), detailValueStyle.Render(lastMsg)))
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Time:"), dimStyle.Render(lastTime)))

	s.WriteString("\n")

	statusLabel := statusClean.Render("Clean ✓")
	if wt.Status.IsDirty {
		statusLabel = statusDirty.Render("Dirty")
	}
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Status:"), statusLabel))

	if wt.Status.IsDirty {
		if wt.Status.Modified > 0 {
			s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render(""), warningStyle.Render(fmt.Sprintf("%d modified", wt.Status.Modified))))
		}
		if wt.Status.Added > 0 {
			s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render(""), successStyle.Render(fmt.Sprintf("%d added", wt.Status.Added))))
		}
		if wt.Status.Deleted > 0 {
			s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render(""), errorStyle.Render(fmt.Sprintf("%d deleted", wt.Status.Deleted))))
		}
		if wt.Status.Untracked > 0 {
			s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render(""), dimStyle.Render(fmt.Sprintf("%d untracked", wt.Status.Untracked))))
		}
	}

	if wt.Ahead > 0 || wt.Behind > 0 {
		s.WriteString("\n")
		if wt.Ahead > 0 {
			s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Ahead:"), aheadStyle.Render(fmt.Sprintf("%d commits", wt.Ahead))))
		}
		if wt.Behind > 0 {
			s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Behind:"), behindStyle.Render(fmt.Sprintf("%d commits", wt.Behind))))
		}
	}

	if wt.IsCurrent {
		s.WriteString("\n  " + currentBadge.Render("● Current worktree") + "\n")
	}

	s.WriteString("\n")
	keys := []helpKey{{"o", "open terminal"}, {"e", "open editor"}}
	if !wt.IsCurrent {
		keys = append(keys, helpKey{"d", "remove"})
	}
	keys = append(keys, helpKey{"esc", "back"})
	s.WriteString(m.renderHelpLine(keys))

	return s.String()
}

type helpKey struct {
	key  string
	desc string
}

func (m Model) renderHelpLine(keys []helpKey) string {
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = helpKeyStyle.Render(k.key) + helpStyle.Render(": "+k.desc)
	}
	return helpStyle.Render("  ") + strings.Join(parts, helpStyle.Render("  •  "))
}

func (m Model) shortHelp() string {
	return m.renderHelpLine([]helpKey{
		{"a", "add"},
		{"n", "new"},
		{"b", "branches"},
		{"/", "search"},
		{"d", "remove"},
		{"enter", "details"},
		{"?", "help"},
		{"q", "quit"},
	})
}

func (m Model) fullHelp() string {
	var s strings.Builder

	s.WriteString(subtitleStyle.Render("  Keybindings") + "\n\n")

	sections := []struct {
		title string
		keys  []helpKey
	}{
		{"Navigation", []helpKey{
			{"j/↓", "move down"},
			{"k/↑", "move up"},
			{"enter", "view details"},
			{"/", "search/filter"},
		}},
		{"Actions", []helpKey{
			{"a", "add worktree (existing branch)"},
			{"n", "add worktree (new branch)"},
			{"b", "browse & select branch"},
			{"d", "remove worktree"},
			{"p", "prune stale worktrees"},
		}},
		{"Tools", []helpKey{
			{"o", "open terminal in worktree"},
			{"e", "open editor in worktree"},
			{"f", "fetch all remotes"},
			{"r", "refresh list"},
		}},
		{"General", []helpKey{
			{"?", "toggle help"},
			{"q", "quit"},
			{"ctrl+c", "force quit"},
		}},
	}

	for _, sec := range sections {
		s.WriteString(dimStyle.Render("  "+sec.title) + "\n")
		for _, k := range sec.keys {
			s.WriteString(fmt.Sprintf("    %s  %s\n",
				helpKeyStyle.Render(lipgloss.NewStyle().Width(10).Render(k.key)),
				helpStyle.Render(k.desc),
			))
		}
		s.WriteString("\n")
	}

	s.WriteString(helpStyle.Render("  Press ? to close help"))
	return s.String()
}

func shortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}
