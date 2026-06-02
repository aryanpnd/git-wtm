package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/aryanpnd/git-wtm/internal/git"
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
	count := dimStyle.Render(fmt.Sprintf("  %d worktrees", len(m.worktrees)))
	s.WriteString(header + count + "\n\n")

	if m.searching {
		s.WriteString("  " + searchStyle.Render("🔍 ") + m.searchInput.View() + "\n\n")
	} else if m.searchInput.Value() != "" {
		s.WriteString("  " + dimStyle.Render(fmt.Sprintf("filter: %q (%d results)", m.searchInput.Value(), len(m.filtered))) + "\n\n")
	}

	if len(m.worktrees) == 0 && !m.loading {
		s.WriteString(dimStyle.Render("  No worktrees found. Press 'a' to add one.\n"))
	}

	maxVisible := m.height - 10
	if maxVisible < 2 {
		maxVisible = 2
	}
	// Each card takes ~4 lines, so divide available space
	maxCards := maxVisible / 4
	if maxCards < 1 {
		maxCards = 1
	}

	visibleStart := 0
	if m.cursor >= maxCards {
		visibleStart = m.cursor - maxCards + 1
	}

	end := visibleStart + maxCards
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	cardWidth := m.width - 4
	if cardWidth > 80 {
		cardWidth = 80
	}
	if cardWidth < 40 {
		cardWidth = 40
	}

	for vi := visibleStart; vi < end; vi++ {
		idx := m.filtered[vi]
		wt := m.worktrees[idx]
		isSelected := vi == m.cursor
		s.WriteString(m.renderWorktreeCard(wt, isSelected, cardWidth))
	}

	if len(m.filtered) > maxCards {
		s.WriteString(dimStyle.Render(fmt.Sprintf("  ↕ showing %d of %d", maxCards, len(m.filtered))) + "\n")
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

func (m Model) renderWorktreeCard(wt git.Worktree, selected bool, width int) string {
	// Build tags
	var tags []string
	if wt.IsPrimary {
		tags = append(tags, tagMain.Render("PRIMARY"))
	}
	if wt.IsCurrent {
		tags = append(tags, tagCurrent.Render("ACTIVE"))
	}
	if wt.IsDetached {
		tags = append(tags, tagDetached.Render("DETACHED"))
	}
	if wt.Status.IsDirty {
		tags = append(tags, tagDirty.Render("UNSAVED CHANGES"))
	} else {
		tags = append(tags, tagClean.Render("✓ clean"))
	}

	// Line 1: branch + tags
	branchDisplay := branchStyle.Render(" " + wt.Branch)
	tagLine := strings.Join(tags, " ")
	line1 := branchDisplay + "  " + tagLine

	// Line 2: status details with labels
	var statusParts []string
	if wt.Status.Modified > 0 {
		statusParts = append(statusParts, warningStyle.Render(fmt.Sprintf("%d modified", wt.Status.Modified)))
	}
	if wt.Status.Added > 0 {
		statusParts = append(statusParts, successStyle.Render(fmt.Sprintf("%d added", wt.Status.Added)))
	}
	if wt.Status.Deleted > 0 {
		statusParts = append(statusParts, errorStyle.Render(fmt.Sprintf("%d deleted", wt.Status.Deleted)))
	}
	if wt.Status.Untracked > 0 {
		statusParts = append(statusParts, dimStyle.Render(fmt.Sprintf("%d untracked", wt.Status.Untracked)))
	}

	syncParts := ""
	if wt.Ahead > 0 {
		syncParts += aheadStyle.Render(fmt.Sprintf("↑ %d ahead", wt.Ahead))
	}
	if wt.Behind > 0 {
		if syncParts != "" {
			syncParts += "  "
		}
		syncParts += behindStyle.Render(fmt.Sprintf("↓ %d behind", wt.Behind))
	}

	line2 := ""
	if len(statusParts) > 0 || syncParts != "" {
		line2 = " " + strings.Join(statusParts, " ")
		if syncParts != "" {
			if line2 != " " {
				line2 += "  "
			}
			line2 += syncParts
		}
	}

	// Line 3: commit + message
	lastMsg := git.GetLastCommitMessage(wt.Path)
	if len(lastMsg) > 45 {
		lastMsg = lastMsg[:42] + "..."
	}
	line3 := " " + dimStyle.Render("commit ") + commitStyle.Render(wt.Head) + " " + dimStyle.Render(lastMsg)

	// Line 4: path
	line4 := " " + dimStyle.Render("path   ") + pathStyle.Render(shortenPath(wt.Path))

	// Compose card content
	content := line1
	if line2 != "" && line2 != " " {
		content += "\n" + line2
	}
	content += "\n" + line3
	content += "\n" + line4

	// Choose border style
	border := cardBorder
	if selected && wt.IsCurrent {
		border = cardBorderCurrent.Width(width)
	} else if selected {
		border = cardBorderSelected.Width(width)
	} else if wt.IsCurrent {
		border = cardBorderCurrent.Width(width)
	} else {
		border = cardBorder.Width(width)
	}

	// Cursor indicator
	cursor := "  "
	if selected {
		cursor = "▸ "
	}

	card := border.Render(content)

	// Prepend cursor to first line of card
	lines := strings.Split(card, "\n")
	if len(lines) > 0 {
		lines[0] = cursor + lines[0]
		for i := 1; i < len(lines); i++ {
			lines[i] = "  " + lines[i]
		}
	}

	return strings.Join(lines, "\n") + "\n"
}

func (m Model) viewAdd() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(" Add Worktree ") + "\n\n")

	// Branch input
	branchLabel := activeInputStyle.Render("  Branch ")
	pathLabel := inactiveInputStyle.Render("  Path   ")
	if m.addStep == 1 {
		branchLabel = inactiveInputStyle.Render("  Branch ")
		pathLabel = activeInputStyle.Render("  Path   ")
	}

	s.WriteString(fmt.Sprintf("%s %s\n", branchLabel, m.branchInput.View()))

	// Matching branches dropdown
	if m.addStep == 0 && len(m.addMatches) > 0 {
		maxShow := 6
		if maxShow > len(m.addMatches) {
			maxShow = len(m.addMatches)
		}

		var dropdownLines []string
		for i := 0; i < maxShow; i++ {
			prefix := "  "
			style := dimStyle
			if i == m.addCursor {
				prefix = "▸ "
				style = selectedItemStyle
			}
			dropdownLines = append(dropdownLines, prefix+style.Render(m.addMatches[i]))
		}
		if len(m.addMatches) > maxShow {
			dropdownLines = append(dropdownLines, dimStyle.Render(fmt.Sprintf("  ... %d more", len(m.addMatches)-maxShow)))
		}

		dropdown := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimGray).
			Padding(0, 1).
			MarginLeft(11).
			Render(strings.Join(dropdownLines, "\n"))

		s.WriteString(dropdown + "\n")
	}

	// "New branch" hint
	if m.addStep == 0 && m.branchInput.Value() != "" && len(m.addMatches) == 0 {
		hint := lipgloss.NewStyle().
			MarginLeft(11).
			Foreground(green).
			Bold(true).
			Render("+ new branch will be created")
		s.WriteString("\n" + hint + "\n")
	}

	s.WriteString("\n")
	s.WriteString(fmt.Sprintf("%s %s\n", pathLabel, m.pathInput.View()))

	// Default path hint
	branchVal := m.branchInput.Value()
	if m.addCursor >= 0 && m.addCursor < len(m.addMatches) {
		branchVal = m.addMatches[m.addCursor]
	}
	if branchVal != "" {
		defaultPath := git.DefaultWorktreePath(branchVal)
		s.WriteString(dimStyle.Render(fmt.Sprintf("           → %s", shortenPath(defaultPath))) + "\n")
	}

	s.WriteString("\n")
	helpKeys := []helpKey{
		{"enter", "create"},
		{"↑/↓", "pick branch"},
		{"tab", "edit path"},
		{"esc", "cancel"},
	}
	if m.addStep == 1 {
		helpKeys = []helpKey{
			{"enter", "create"},
			{"tab", "back to branch"},
			{"esc", "back"},
		}
	}
	s.WriteString(m.renderHelpLine(helpKeys))

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

	content := fmt.Sprintf("%s %s\n\n", detailLabelStyle.Render("Path:"), errorStyle.Render(shortenPath(wt.Path)))
	content += fmt.Sprintf("%s %s\n", detailLabelStyle.Render("Branch:"), branchStyle.Render(wt.Branch))

	if wt.Status.IsDirty {
		content += "\n" + warningStyle.Render("⚠ This worktree has uncommitted changes!")
	}

	forceLabel := dimStyle.Render("off")
	if m.confirmForce {
		forceLabel = errorStyle.Render("ON")
	}
	content += fmt.Sprintf("\n\n%s %s", detailLabelStyle.Render("Force:"), forceLabel)

	s.WriteString(boxStyle.Render(content) + "\n\n")

	s.WriteString(m.renderHelpLine([]helpKey{
		{"y", "confirm"},
		{"n/esc", "cancel"},
		{"f", "toggle force"},
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

	// Tags
	var tags []string
	if wt.IsPrimary {
		tags = append(tags, tagMain.Render("PRIMARY"))
	}
	if wt.IsCurrent {
		tags = append(tags, tagCurrent.Render("ACTIVE"))
	}
	if wt.IsDetached {
		tags = append(tags, tagDetached.Render("DETACHED"))
	}
	if len(tags) > 0 {
		s.WriteString("  " + strings.Join(tags, " ") + "\n\n")
	}

	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Branch:"), branchStyle.Render(wt.Branch)))
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Path:"), detailValueStyle.Render(wt.Path)))
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("HEAD:"), commitStyle.Render(wt.Head)))

	lastMsg := git.GetLastCommitMessage(wt.Path)
	lastTime := git.GetLastCommitTime(wt.Path)
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Commit:"), detailValueStyle.Render(lastMsg)))
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Time:"), dimStyle.Render(lastTime)))

	s.WriteString("\n  " + separatorStyle.Render(strings.Repeat("─", 40)) + "\n\n")

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
			s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Ahead:"), aheadStyle.Render(fmt.Sprintf("↑ %d commits", wt.Ahead))))
		}
		if wt.Behind > 0 {
			s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Behind:"), behindStyle.Render(fmt.Sprintf("↓ %d commits", wt.Behind))))
		}
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
		{"/", "search"},
		{"d", "remove"},
		{"r", "refresh"},
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
			{"a", "add worktree (type to search or create)"},
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
