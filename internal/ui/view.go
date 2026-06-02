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
	s.WriteString(m.renderHeader())

	switch m.activeTab {
	case tabWorktrees:
		s.WriteString(m.viewWorktreesTab())
	case tabBranches:
		s.WriteString(m.viewBranchesTab())
	}

	return s.String()
}

func (m Model) renderHeader() string {
	loading := ""
	if m.loading {
		loading = "  " + loadingStyle.Render("● loading...")
	}

	titleLine := logoStyle.Render("  🌳 git-wtm  ") + "  " + titleStyle.Render("Worktree & Branch Manager") + loading

	wtLabel := " Worktrees "
	brLabel := " Branches "

	var left, right string
	if m.activeTab == tabWorktrees {
		left = activeTabStyle.Render(wtLabel)
		right = inactiveTabStyle.Render(brLabel)
	} else {
		left = inactiveTabStyle.Render(wtLabel)
		right = activeTabStyle.Render(brLabel)
	}

	hint := dimStyle.Render("  ← → to switch")
	tabLine := "  " + left + "  " + right + hint

	sep := separatorStyle.Render(strings.Repeat("─", min(m.width-2, 70)))

	return titleLine + "\n\n" + tabLine + "\n" + sep + "\n\n"
}

func (m Model) renderTabs() string {
	return ""
}

// ==========================================
// WORKTREES TAB
// ==========================================

func (m Model) viewWorktreesTab() string {
	switch m.wtView {
	case viewList:
		return m.viewWtList()
	case viewAdd:
		return m.viewWtAdd()
	case viewRemoveConfirm:
		return m.viewWtRemove()
	case viewDetail:
		return m.viewWtDetail()
	}
	return ""
}

func (m Model) viewWtList() string {
	var s strings.Builder

	if m.wtSearching {
		s.WriteString("  " + searchStyle.Render("🔍 ") + m.wtSearch.View() + "\n\n")
	} else if m.wtSearch.Value() != "" {
		s.WriteString("  " + dimStyle.Render(fmt.Sprintf("filter: %q (%d)", m.wtSearch.Value(), len(m.wtFiltered))) + "\n\n")
	}

	if len(m.worktrees) == 0 && !m.loading {
		s.WriteString(dimStyle.Render("  No worktrees found. Press 'a' to add one.\n"))
	}

	maxCards := (m.height - 10) / 4
	if maxCards < 1 {
		maxCards = 1
	}

	visibleStart := 0
	if m.wtCursor >= maxCards {
		visibleStart = m.wtCursor - maxCards + 1
	}
	end := visibleStart + maxCards
	if end > len(m.wtFiltered) {
		end = len(m.wtFiltered)
	}

	cardWidth := m.width - 6
	if cardWidth > 76 {
		cardWidth = 76
	}
	if cardWidth < 40 {
		cardWidth = 40
	}

	for vi := visibleStart; vi < end; vi++ {
		idx := m.wtFiltered[vi]
		wt := m.worktrees[idx]
		isSelected := vi == m.wtCursor
		s.WriteString(m.renderWtCard(wt, isSelected, cardWidth))
	}

	if len(m.wtFiltered) > maxCards {
		s.WriteString(dimStyle.Render(fmt.Sprintf("  ↕ %d of %d", maxCards, len(m.wtFiltered))) + "\n")
	}

	s.WriteString(m.renderMessages())
	s.WriteString("\n")

	if m.showHelp {
		s.WriteString(m.wtFullHelp())
	} else {
		s.WriteString(m.wtShortHelp())
	}

	return s.String()
}

func (m Model) renderWtCard(wt git.Worktree, selected bool, width int) string {
	// Tags
	var tags []string
	if wt.IsPrimary {
		tags = append(tags, tagPrimary.Render("PRIMARY"))
	}
	if wt.IsCurrent {
		tags = append(tags, tagActive.Render("ACTIVE"))
	}
	if wt.IsDetached {
		tags = append(tags, tagDetached.Render("DETACHED"))
	}
	if wt.Status.IsDirty {
		tags = append(tags, tagUnsaved.Render("UNSAVED CHANGES"))
	} else {
		tags = append(tags, tagClean.Render("✓ clean"))
	}

	line1 := " " + branchStyle.Render(wt.Branch) + "  " + strings.Join(tags, " ")

	// Status line
	var parts []string
	if wt.Status.Modified > 0 {
		parts = append(parts, warningStyle.Render(fmt.Sprintf("%d modified", wt.Status.Modified)))
	}
	if wt.Status.Added > 0 {
		parts = append(parts, successStyle.Render(fmt.Sprintf("%d added", wt.Status.Added)))
	}
	if wt.Status.Deleted > 0 {
		parts = append(parts, errorStyle.Render(fmt.Sprintf("%d deleted", wt.Status.Deleted)))
	}
	if wt.Status.Untracked > 0 {
		parts = append(parts, dimStyle.Render(fmt.Sprintf("%d untracked", wt.Status.Untracked)))
	}
	if wt.Ahead > 0 {
		parts = append(parts, aheadStyle.Render(fmt.Sprintf("↑ %d ahead", wt.Ahead)))
	}
	if wt.Behind > 0 {
		parts = append(parts, behindStyle.Render(fmt.Sprintf("↓ %d behind", wt.Behind)))
	}

	line2 := ""
	if len(parts) > 0 {
		line2 = " " + strings.Join(parts, dimStyle.Render(" · "))
	}

	// Commit
	lastMsg := git.GetLastCommitMessage(wt.Path)
	if len(lastMsg) > 42 {
		lastMsg = lastMsg[:39] + "..."
	}
	line3 := " " + dimStyle.Render("commit ") + commitStyle.Render(wt.Head) + " " + dimStyle.Render(lastMsg)

	// Path
	line4 := " " + dimStyle.Render("path   ") + pathStyle.Render(shortenPath(wt.Path))

	content := line1
	if line2 != "" {
		content += "\n" + line2
	}
	content += "\n" + line3 + "\n" + line4

	// Border
	border := cardBorder.Width(width)
	if selected && wt.IsCurrent {
		border = cardBorderActive.Width(width)
	} else if selected {
		border = cardBorderSelected.Width(width)
	} else if wt.IsCurrent {
		border = cardBorderActive.Width(width)
	}

	card := border.Render(content)

	cursor := "  "
	if selected {
		cursor = "▸ "
	}

	lines := strings.Split(card, "\n")
	if len(lines) > 0 {
		lines[0] = cursor + lines[0]
		for i := 1; i < len(lines); i++ {
			lines[i] = "  " + lines[i]
		}
	}

	return strings.Join(lines, "\n") + "\n"
}

func (m Model) viewWtAdd() string {
	var s strings.Builder

	s.WriteString("  " + titleStyle.Render("Add Worktree") + "\n\n")

	branchLabel := activeInputStyle.Render("  Branch ")
	pathLabel := inactiveInputStyle.Render("  Path   ")
	if m.addStep == 1 {
		branchLabel = inactiveInputStyle.Render("  Branch ")
		pathLabel = activeInputStyle.Render("  Path   ")
	}

	s.WriteString(fmt.Sprintf("%s %s\n", branchLabel, m.addInput.View()))

	if m.addStep == 0 && len(m.addMatches) > 0 {
		maxShow := 5
		if maxShow > len(m.addMatches) {
			maxShow = len(m.addMatches)
		}
		var lines []string
		for i := 0; i < maxShow; i++ {
			prefix := "  "
			style := dimStyle
			if i == m.addCursor {
				prefix = "▸ "
				style = selectedItemStyle
			}
			lines = append(lines, prefix+style.Render(m.addMatches[i]))
		}
		if len(m.addMatches) > maxShow {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("  +%d more", len(m.addMatches)-maxShow)))
		}

		dropdown := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimGray).
			Padding(0, 1).
			MarginLeft(11).
			Render(strings.Join(lines, "\n"))
		s.WriteString(dropdown + "\n")
	}

	if m.addStep == 0 && m.addInput.Value() != "" && len(m.addMatches) == 0 {
		s.WriteString(lipgloss.NewStyle().MarginLeft(11).Foreground(green).Bold(true).Render("+ new branch will be created") + "\n")
	}

	s.WriteString("\n")
	s.WriteString(fmt.Sprintf("%s %s\n", pathLabel, m.pathInput.View()))

	branchVal := m.addInput.Value()
	if m.addCursor >= 0 && m.addCursor < len(m.addMatches) {
		branchVal = m.addMatches[m.addCursor]
	}
	if branchVal != "" {
		s.WriteString(dimStyle.Render(fmt.Sprintf("           → %s", shortenPath(git.DefaultWorktreePath(branchVal)))) + "\n")
	}

	s.WriteString("\n")
	keys := []helpKey{{"enter", "create"}, {"↑/↓", "pick"}, {"tab", "edit path"}, {"esc", "cancel"}}
	if m.addStep == 1 {
		keys = []helpKey{{"enter", "create"}, {"tab", "back"}, {"esc", "back"}}
	}
	s.WriteString(m.renderHelpLine(keys))

	return s.String()
}

func (m Model) viewWtRemove() string {
	var s strings.Builder

	s.WriteString("  " + titleStyle.Render("Remove Worktree") + "\n\n")

	wt := m.selectedWorktree()
	if wt == nil {
		return dimStyle.Render("  No worktree selected.\n")
	}

	content := fmt.Sprintf("%s %s\n\n%s %s",
		detailLabelStyle.Render("Branch:"), branchStyle.Render(wt.Branch),
		detailLabelStyle.Render("Path:"), errorStyle.Render(shortenPath(wt.Path)))

	if wt.Status.IsDirty {
		content += "\n\n" + warningStyle.Render("⚠ Has uncommitted changes!")
	}

	forceLabel := dimStyle.Render("off")
	if m.confirmForce {
		forceLabel = errorStyle.Render("ON")
	}
	content += fmt.Sprintf("\n\n%s %s", detailLabelStyle.Render("Force:"), forceLabel)

	s.WriteString(boxStyle.Render(content) + "\n\n")
	s.WriteString(m.renderHelpLine([]helpKey{{"y", "confirm"}, {"n/esc", "cancel"}, {"f", "toggle force"}}))

	return s.String()
}

func (m Model) viewWtDetail() string {
	var s strings.Builder

	wt := m.selectedWorktree()
	if wt == nil {
		return dimStyle.Render("  No worktree selected.\n")
	}

	s.WriteString("  " + titleStyle.Render("Worktree Details") + "\n\n")

	var tags []string
	if wt.IsPrimary {
		tags = append(tags, tagPrimary.Render("PRIMARY"))
	}
	if wt.IsCurrent {
		tags = append(tags, tagActive.Render("ACTIVE"))
	}
	if len(tags) > 0 {
		s.WriteString("  " + strings.Join(tags, " ") + "\n\n")
	}

	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Branch:"), branchStyle.Render(wt.Branch)))
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Path:"), detailValueStyle.Render(wt.Path)))
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("HEAD:"), commitStyle.Render(wt.Head)))
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Commit:"), detailValueStyle.Render(git.GetLastCommitMessage(wt.Path))))
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Time:"), dimStyle.Render(git.GetLastCommitTime(wt.Path))))

	s.WriteString("\n  " + separatorStyle.Render(strings.Repeat("─", 40)) + "\n\n")

	if wt.Status.IsDirty {
		s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Status:"), statusDirty.Render("Unsaved changes")))
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
	} else {
		s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Status:"), statusClean.Render("Clean ✓")))
	}

	s.WriteString("\n")
	keys := []helpKey{{"o", "terminal"}, {"e", "editor"}}
	if !wt.IsCurrent {
		keys = append(keys, helpKey{"d", "remove"})
	}
	keys = append(keys, helpKey{"esc", "back"})
	s.WriteString(m.renderHelpLine(keys))

	return s.String()
}

// ==========================================
// BRANCHES TAB
// ==========================================

func (m Model) viewBranchesTab() string {
	switch m.brView {
	case viewList:
		return m.viewBrList()
	case viewBranchDetail:
		return m.viewBrDetail()
	case viewBranchCreate:
		return m.viewBrCreate()
	case viewBranchRename:
		return m.viewBrRename()
	case viewBranchDelete:
		return m.viewBrDelete()
	}
	return ""
}

func (m Model) viewBrList() string {
	var s strings.Builder

	if m.brSearching {
		s.WriteString("  " + searchStyle.Render("🔍 ") + m.brSearch.View() + "\n\n")
	} else if m.brSearch.Value() != "" {
		s.WriteString("  " + dimStyle.Render(fmt.Sprintf("filter: %q (%d)", m.brSearch.Value(), len(m.brFiltered))) + "\n\n")
	}

	if len(m.branchList) == 0 && !m.loading {
		s.WriteString(dimStyle.Render("  No branches found.\n"))
	}

	// Each branch item is 2 lines (name+tags, time)
	// Reserve: header(3) + search(2) + box border(2) + scroll(1) + messages(2) + help(2) = 12
	maxVisible := (m.height - 12) / 2
	if maxVisible < 3 {
		maxVisible = 3
	}
	if maxVisible > 10 {
		maxVisible = 10
	}

	visibleStart := 0
	if m.brCursor >= maxVisible {
		visibleStart = m.brCursor - maxVisible + 1
	}
	end := visibleStart + maxVisible
	if end > len(m.brFiltered) {
		end = len(m.brFiltered)
	}

	// Build the list content
	var listContent strings.Builder
	for vi := visibleStart; vi < end; vi++ {
		idx := m.brFiltered[vi]
		b := m.branchList[idx]
		isSelected := vi == m.brCursor
		listContent.WriteString(m.renderBranchItem(b, isSelected))
	}

	// Scroll indicator
	scrollInfo := ""
	if len(m.brFiltered) > maxVisible {
		scrollInfo = dimStyle.Render(fmt.Sprintf(" ↕ %d of %d branches", maxVisible, len(m.brFiltered)))
		if visibleStart > 0 {
			scrollInfo += dimStyle.Render("  ↑ more above")
		}
		if end < len(m.brFiltered) {
			scrollInfo += dimStyle.Render("  ↓ more below")
		}
	} else {
		scrollInfo = dimStyle.Render(fmt.Sprintf(" %d branches", len(m.brFiltered)))
	}

	listContent.WriteString(scrollInfo)

	// Wrap in a bordered box
	listWidth := m.width - 6
	if listWidth > 74 {
		listWidth = 74
	}
	if listWidth < 40 {
		listWidth = 40
	}

	listBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(dimGray).
		Padding(0, 1).
		Width(listWidth).
		Render(listContent.String())

	s.WriteString("  " + listBox + "\n")

	s.WriteString(m.renderMessages())
	s.WriteString("\n")

	if m.showHelp {
		s.WriteString(m.brFullHelp())
	} else {
		s.WriteString(m.brShortHelp())
	}

	return s.String()
}

func (m Model) renderBranchItem(b git.Branch, selected bool) string {
	cursor := "  "
	if selected {
		cursor = "▸ "
	}

	name := branchStyle.Render(b.Name)
	if selected {
		name = selectedItemStyle.Render(b.Name)
	}

	// Tags — keep it simple and compact
	var tags []string
	if b.IsCurrent {
		tags = append(tags, tagActive.Render("ACTIVE"))
	}
	if b.Upstream != "" {
		tags = append(tags, tagTracked.Render("remote"))
	} else {
		tags = append(tags, tagLocal.Render("local only"))
	}

	// Sync info
	if b.Ahead > 0 {
		tags = append(tags, aheadStyle.Render(fmt.Sprintf("↑%d", b.Ahead)))
	}
	if b.Behind > 0 {
		tags = append(tags, behindStyle.Render(fmt.Sprintf("↓%d", b.Behind)))
	}

	line := cursor + name + "  " + strings.Join(tags, " ")

	// Second line: compact time + upstream
	detail := "    " + dimStyle.Render(b.CommitTime)

	return line + "\n" + detail + "\n"
}

func (m Model) viewBrDetail() string {
	var s strings.Builder

	b := m.selectedBranch()
	if b == nil {
		return dimStyle.Render("  No branch selected.\n")
	}

	s.WriteString("  " + titleStyle.Render("Branch Details") + "\n\n")

	var tags []string
	if b.IsCurrent {
		tags = append(tags, tagActive.Render("ACTIVE"))
	}
	if b.Upstream != "" {
		tags = append(tags, tagTracked.Render("TRACKED"))
	} else {
		tags = append(tags, tagLocal.Render("local only"))
	}
	s.WriteString("  " + strings.Join(tags, " ") + "\n\n")

	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Name:"), branchStyle.Render(b.Name)))
	if b.Upstream != "" {
		s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Upstream:"), detailValueStyle.Render(b.Upstream)))
	}
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Last:"), dimStyle.Render(b.CommitTime)))
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Commit:"), commitStyle.Render(b.LastCommit)))

	if b.Ahead > 0 || b.Behind > 0 {
		s.WriteString("\n")
		if b.Ahead > 0 {
			s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Ahead:"), aheadStyle.Render(fmt.Sprintf("↑ %d commits", b.Ahead))))
		}
		if b.Behind > 0 {
			s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Behind:"), behindStyle.Render(fmt.Sprintf("↓ %d commits", b.Behind))))
		}
	}

	s.WriteString("\n")
	keys := []helpKey{{"c", "checkout"}}
	if !b.IsCurrent {
		keys = append(keys, helpKey{"d", "delete"})
	}
	keys = append(keys, helpKey{"esc", "back"})
	s.WriteString(m.renderHelpLine(keys))

	return s.String()
}

func (m Model) viewBrCreate() string {
	var s strings.Builder
	s.WriteString("  " + titleStyle.Render("Create Branch") + "\n\n")
	s.WriteString(fmt.Sprintf("  %s %s\n\n", activeInputStyle.Render("Name"), m.brCreateInput.View()))
	s.WriteString(dimStyle.Render("  Branch will be created from current HEAD") + "\n\n")
	s.WriteString(m.renderHelpLine([]helpKey{{"enter", "create"}, {"esc", "cancel"}}))
	return s.String()
}

func (m Model) viewBrRename() string {
	var s strings.Builder
	b := m.selectedBranch()
	s.WriteString("  " + titleStyle.Render("Rename Branch") + "\n\n")
	if b != nil {
		s.WriteString("  " + dimStyle.Render("Renaming: "+b.Name) + "\n\n")
	}
	s.WriteString(fmt.Sprintf("  %s %s\n\n", activeInputStyle.Render("New name"), m.brRenameInput.View()))
	s.WriteString(m.renderHelpLine([]helpKey{{"enter", "rename"}, {"esc", "cancel"}}))
	return s.String()
}

func (m Model) viewBrDelete() string {
	var s strings.Builder
	s.WriteString("  " + titleStyle.Render("Delete Branch") + "\n\n")

	b := m.selectedBranch()
	if b == nil {
		return dimStyle.Render("  No branch selected.\n")
	}

	content := fmt.Sprintf("%s %s", detailLabelStyle.Render("Branch:"), errorStyle.Render(b.Name))
	forceLabel := dimStyle.Render("off")
	if m.brDeleteForce {
		forceLabel = errorStyle.Render("ON (unmerged commits will be lost)")
	}
	content += fmt.Sprintf("\n\n%s %s", detailLabelStyle.Render("Force:"), forceLabel)

	s.WriteString(boxStyle.Render(content) + "\n\n")
	s.WriteString(m.renderHelpLine([]helpKey{{"y", "confirm"}, {"n/esc", "cancel"}, {"f", "toggle force"}}))

	return s.String()
}

// ==========================================
// SHARED HELPERS
// ==========================================

func (m Model) renderMessages() string {
	var s strings.Builder
	if m.statusMsg != "" {
		s.WriteString("\n " + successStyle.Render("✓ "+m.statusMsg))
	}
	if m.err != nil {
		s.WriteString("\n " + errorStyle.Render("✗ "+m.err.Error()))
	}
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

func (m Model) wtShortHelp() string {
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

func (m Model) brShortHelp() string {
	return m.renderHelpLine([]helpKey{
		{"c", "checkout"},
		{"a", "create"},
		{"R", "rename"},
		{"d", "delete"},
		{"/", "search"},
		{"?", "help"},
		{"q", "quit"},
	})
}

func (m Model) wtFullHelp() string {
	var s strings.Builder
	s.WriteString(subtitleStyle.Render("  Worktree Keys") + "\n\n")
	sections := []struct {
		title string
		keys  []helpKey
	}{
		{"Navigate", []helpKey{{"j/↓", "down"}, {"k/↑", "up"}, {"enter", "details"}, {"/", "search"}}},
		{"Actions", []helpKey{{"a", "add worktree"}, {"d", "remove"}, {"p", "prune stale"}}},
		{"Tools", []helpKey{{"o", "open terminal"}, {"e", "open editor"}, {"f", "fetch"}, {"r", "refresh"}}},
	}
	for _, sec := range sections {
		s.WriteString(dimStyle.Render("  "+sec.title) + "\n")
		for _, k := range sec.keys {
			s.WriteString(fmt.Sprintf("    %s  %s\n",
				helpKeyStyle.Render(lipgloss.NewStyle().Width(8).Render(k.key)),
				helpStyle.Render(k.desc)))
		}
		s.WriteString("\n")
	}
	s.WriteString(helpStyle.Render("  ? to close"))
	return s.String()
}

func (m Model) brFullHelp() string {
	var s strings.Builder
	s.WriteString(subtitleStyle.Render("  Branch Keys") + "\n\n")
	sections := []struct {
		title string
		keys  []helpKey
	}{
		{"Navigate", []helpKey{{"j/↓", "down"}, {"k/↑", "up"}, {"enter", "details"}, {"/", "search"}}},
		{"Actions", []helpKey{{"c", "checkout"}, {"a/n", "create"}, {"R", "rename"}, {"d", "delete"}, {"m", "merge into current"}}},
		{"Tools", []helpKey{{"f", "fetch remotes"}, {"r", "refresh"}}},
	}
	for _, sec := range sections {
		s.WriteString(dimStyle.Render("  "+sec.title) + "\n")
		for _, k := range sec.keys {
			s.WriteString(fmt.Sprintf("    %s  %s\n",
				helpKeyStyle.Render(lipgloss.NewStyle().Width(8).Render(k.key)),
				helpStyle.Render(k.desc)))
		}
		s.WriteString("\n")
	}
	s.WriteString(helpStyle.Render("  ? to close"))
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
