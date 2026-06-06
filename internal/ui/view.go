package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/aryanpnd/git-wtm/internal/git"
)

func (m Model) contentWidth() int {
	w := m.width - 2
	if w < 30 {
		w = 30
	}
	return w
}

func (m Model) View() string {
	if m.width == 0 {
		return "\n  Loading..."
	}

	if m.showModal {
		return m.renderModal()
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	var content string
	switch m.activeTab {
	case tabWorktrees:
		content = m.viewWorktreesTab()
	case tabBranches:
		content = m.viewBranchesTab()
	}

	// Calculate vertical spacing to push footer to bottom
	headerLines := strings.Count(header, "\n")
	contentLines := strings.Count(content, "\n")
	footerLines := strings.Count(footer, "\n") + 1
	usedLines := headerLines + contentLines + footerLines
	padding := m.height - usedLines
	if padding < 0 {
		padding = 0
	}

	return header + content + strings.Repeat("\n", padding) + footer
}

func (m Model) renderModal() string {
	titleSty := modalTitleError
	modalBox := modalErrorStyle
	icon := "✗"
	if !m.modalIsError {
		titleSty = modalTitleSuccess
		modalBox = modalSuccessStyle
		icon = "✓"
	}

	modalWidth := m.width / 2
	if modalWidth > 60 {
		modalWidth = 60
	}
	if modalWidth < 30 {
		modalWidth = 30
	}

	title := titleSty.Render(icon + "  " + m.modalTitle)
	message := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "#DDDDDD"}).
		Width(modalWidth - 8).
		Render(m.modalMessage)
	dismiss := dimStyle.Render("\nPress any key to dismiss")

	content := title + "\n\n" + message + "\n" + dismiss
	modal := modalBox.Width(modalWidth).Render(content)

	// Center both horizontally and vertically
	centered := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
	return centered
}

func (m Model) renderHeader() string {
	cw := m.contentWidth()

	// Title bar — full width
	title := " git-wtm"
	subtitle := "Worktree & Branch Manager "
	pad := cw - len(title) - len(subtitle)
	if pad < 2 {
		pad = 2
		subtitle = ""
	}
	titleBar := logoStyle.Width(cw).Render(
		title + strings.Repeat(" ", pad) + subtitle,
	)

	// Loading
	loadingLine := ""
	if m.loading {
		loadingLine = "  " + loadingStyle.Render("● loading...")
	}

	// Tabs
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
	hint := dimStyle.Render("  ← → switch")
	tabLine := "  " + left + " " + right + hint

	sep := separatorStyle.Render(strings.Repeat("─", cw))

	header := titleBar + loadingLine + "\n\n" + tabLine + "\n" + sep
	if m.updateInfo != nil {
		banner := updateStyle.Render(fmt.Sprintf("  ↑ v%s available — %s  ", m.updateInfo.NewVersion, m.updateInfo.UpgradeCmd))
		header += "\n" + banner
	}

	return header + "\n\n"
}

func (m Model) renderFooter() string {
	cw := m.contentWidth()

	var keys []helpKey
	switch m.activeTab {
	case tabWorktrees:
		keys = m.wtHelpKeys()
	case tabBranches:
		keys = m.brHelpKeys()
	}

	if m.showHelp {
		return ""
	}

	var parts []string
	lineLen := 0
	for _, k := range keys {
		partLen := len(k.key) + len(k.desc) + 4
		if lineLen+partLen > cw && len(parts) > 0 {
			break
		}
		parts = append(parts, helpKeyStyle.Render(k.key)+" "+helpStyle.Render(k.desc))
		lineLen += partLen
	}

	sep := separatorStyle.Render(strings.Repeat("─", cw))
	bar := "  " + strings.Join(parts, dimStyle.Render(" · "))
	return sep + "\n" + bar
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
		s.WriteString("  " + searchStyle.Render("/ ") + m.wtSearch.View() + "\n\n")
	} else if m.wtSearch.Value() != "" {
		s.WriteString("  " + dimStyle.Render(fmt.Sprintf("filter: %q (%d results)", m.wtSearch.Value(), len(m.wtFiltered))) + "\n\n")
	}

	if m.loading && len(m.worktrees) == 0 {
		s.WriteString(dimStyle.Render("  Loading worktrees...") + "\n")
		return s.String()
	}

	if len(m.worktrees) == 0 {
		s.WriteString(dimStyle.Render("  No worktrees found. Press 'a' to add one.") + "\n")
		return s.String()
	}

	// Calculate how many cards fit: each card is ~6 lines
	// Reserve: header(~6) + footer(2) + scroll indicator(1) + messages(2)
	availLines := m.height - 11
	cardHeight := 6
	maxCards := availLines / cardHeight
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

	// Card fills available width
	cardWidth := m.contentWidth() - 2
	if cardWidth < 36 {
		cardWidth = 36
	}

	for vi := visibleStart; vi < end; vi++ {
		idx := m.wtFiltered[vi]
		wt := m.worktrees[idx]
		isSelected := vi == m.wtCursor
		s.WriteString(m.renderWtCard(wt, isSelected, cardWidth))
	}

	if len(m.wtFiltered) > maxCards {
		s.WriteString(dimStyle.Render(fmt.Sprintf("  ↕ %d of %d worktrees", min(maxCards, len(m.wtFiltered)), len(m.wtFiltered))) + "\n")
	}

	s.WriteString(m.renderMessages())

	if m.showHelp {
		s.WriteString("\n")
		s.WriteString(m.wtFullHelp())
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
		tags = append(tags, tagUnsaved.Render("UNSAVED"))
	} else {
		tags = append(tags, tagClean.Render("clean"))
	}

	line1 := " " + branchStyle.Render(wt.Branch) + "  " + strings.Join(tags, " ")

	// Status indicators
	var parts []string
	if wt.Status.Modified > 0 {
		parts = append(parts, warningStyle.Render(fmt.Sprintf("~%d", wt.Status.Modified)))
	}
	if wt.Status.Added > 0 {
		parts = append(parts, successStyle.Render(fmt.Sprintf("+%d", wt.Status.Added)))
	}
	if wt.Status.Deleted > 0 {
		parts = append(parts, errorStyle.Render(fmt.Sprintf("-%d", wt.Status.Deleted)))
	}
	if wt.Status.Untracked > 0 {
		parts = append(parts, dimStyle.Render(fmt.Sprintf("?%d", wt.Status.Untracked)))
	}
	if wt.Ahead > 0 {
		parts = append(parts, aheadStyle.Render(fmt.Sprintf("↑%d", wt.Ahead)))
	}
	if wt.Behind > 0 {
		parts = append(parts, behindStyle.Render(fmt.Sprintf("↓%d", wt.Behind)))
	}

	line2 := ""
	if len(parts) > 0 {
		line2 = " " + strings.Join(parts, " ")
	}

	// Commit — truncate to card width
	lastMsg := git.GetLastCommitMessage(wt.Path)
	maxMsg := width - 16
	if maxMsg < 10 {
		maxMsg = 10
	}
	if len(lastMsg) > maxMsg {
		lastMsg = lastMsg[:maxMsg-3] + "..."
	}
	line3 := " " + commitStyle.Render(wt.Head) + " " + dimStyle.Render(lastMsg)

	// Path — truncate from left
	pth := shortenPath(wt.Path)
	maxPath := width - 4
	if maxPath > 0 && len(pth) > maxPath {
		pth = "..." + pth[len(pth)-maxPath+3:]
	}
	line4 := " " + pathStyle.Render(pth)

	content := line1
	if line2 != "" {
		content += "\n" + line2
	}
	content += "\n" + line3 + "\n" + line4

	// Border style based on state
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

func (m Model) renderField(content string, focused bool) string {
	cw := m.contentWidth()
	borderColor := dimGray
	if focused {
		borderColor = purple
	}
	top := lipgloss.NewStyle().Foreground(borderColor).Render("╭" + strings.Repeat("─", cw-2) + "╮")
	bot := lipgloss.NewStyle().Foreground(borderColor).Render("╰" + strings.Repeat("─", cw-2) + "╯")
	left := lipgloss.NewStyle().Foreground(borderColor).Render("│")
	right := lipgloss.NewStyle().Foreground(borderColor).Render("│")

	// Pad content to fill width (using visible length)
	line := left + " " + content + " " + right
	return top + "\n" + line + "\n" + bot
}

func (m Model) viewWtAdd() string {
	var s strings.Builder
	cw := m.contentWidth()

	s.WriteString("  " + titleStyle.Render("Add Worktree") + "\n\n")

	// Branch field
	s.WriteString("  " + activeInputStyle.Render("Branch") + "\n")
	s.WriteString(m.renderField(m.addInput.View(), m.addStep == 0) + "\n")

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
			MarginLeft(2).
			Render(strings.Join(lines, "\n"))
		s.WriteString(dropdown + "\n")
	}

	if m.addStep == 0 && m.addInput.Value() != "" && len(m.addMatches) == 0 {
		s.WriteString("  " + lipgloss.NewStyle().Foreground(green).Bold(true).Render("+ new branch") + "\n")
	}

	// Base selector
	s.WriteString("\n  " + activeInputStyle.Render("Base") + "\n")
	var baseRaw string
	for i, base := range m.wtAddBases {
		if i == m.wtAddBaseIdx {
			baseRaw += tagActive.Render(" " + base + " ")
		} else {
			baseRaw += dimStyle.Render(" " + base + " ")
		}
		if i < len(m.wtAddBases)-1 {
			baseRaw += " "
		}
	}
	if m.addStep == 1 {
		baseRaw += "  " + dimStyle.Render("← →")
	}
	s.WriteString(m.renderField(baseRaw, m.addStep == 1) + "\n")

	// Path field
	s.WriteString("\n  " + activeInputStyle.Render("Path") + "\n")
	s.WriteString(m.renderField(m.pathInput.View(), m.addStep == 2) + "\n")

	branchVal := m.addInput.Value()
	if m.addCursor >= 0 && m.addCursor < len(m.addMatches) {
		branchVal = m.addMatches[m.addCursor]
	}
	if branchVal != "" {
		defaultPath := shortenPath(git.DefaultWorktreePath(branchVal))
		s.WriteString("  " + dimStyle.Render("→ "+defaultPath) + "\n")
	}

	s.WriteString("\n")
	sep := separatorStyle.Render(strings.Repeat("─", cw))
	var keys []helpKey
	switch m.addStep {
	case 0:
		keys = []helpKey{{"enter", "create"}, {"↑/↓", "pick"}, {"tab", "next"}, {"esc", "cancel"}}
	case 1:
		keys = []helpKey{{"enter", "create"}, {"←/→", "select"}, {"tab", "next"}, {"esc", "cancel"}}
	case 2:
		keys = []helpKey{{"enter", "create"}, {"^o", "browse"}, {"tab", "next"}, {"esc", "cancel"}}
	}
	s.WriteString(sep + "\n  " + m.renderKeys(keys))

	return s.String()
}

func (m Model) viewWtRemove() string {
	var s strings.Builder
	cw := m.contentWidth()

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

	boxWidth := cw - 4
	if boxWidth > 60 {
		boxWidth = 60
	}
	s.WriteString("  " + boxStyle.Width(boxWidth).Render(content) + "\n\n")

	sep := separatorStyle.Render(strings.Repeat("─", cw))
	s.WriteString(sep + "\n  " + m.renderKeys([]helpKey{{"y", "confirm"}, {"n/esc", "cancel"}, {"f", "force"}}))

	return s.String()
}

func (m Model) viewWtDetail() string {
	var s strings.Builder
	cw := m.contentWidth()

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
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Path:"), detailValueStyle.Render(shortenPath(wt.Path))))
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("HEAD:"), commitStyle.Render(wt.Head)))
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Commit:"), detailValueStyle.Render(git.GetLastCommitMessage(wt.Path))))
	s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Time:"), dimStyle.Render(git.GetLastCommitTime(wt.Path))))

	s.WriteString("\n  " + separatorStyle.Render(strings.Repeat("─", min(cw-4, 50))) + "\n\n")

	if wt.Status.IsDirty {
		s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Status:"), statusDirty.Render("Unsaved changes")))
		if wt.Status.Modified > 0 {
			s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render(""), warningStyle.Render(fmt.Sprintf("~%d modified", wt.Status.Modified))))
		}
		if wt.Status.Added > 0 {
			s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render(""), successStyle.Render(fmt.Sprintf("+%d added", wt.Status.Added))))
		}
		if wt.Status.Deleted > 0 {
			s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render(""), errorStyle.Render(fmt.Sprintf("-%d deleted", wt.Status.Deleted))))
		}
		if wt.Status.Untracked > 0 {
			s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render(""), dimStyle.Render(fmt.Sprintf("?%d untracked", wt.Status.Untracked))))
		}
	} else {
		s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Status:"), statusClean.Render("Clean")))
	}

	s.WriteString("\n")
	sep := separatorStyle.Render(strings.Repeat("─", cw))
	keys := []helpKey{{"o", "terminal"}, {"e", "editor"}}
	if !wt.IsCurrent {
		keys = append(keys, helpKey{"d", "remove"})
	}
	keys = append(keys, helpKey{"esc", "back"})
	s.WriteString(sep + "\n  " + m.renderKeys(keys))

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
	cw := m.contentWidth()

	if m.brSearching {
		s.WriteString("  " + searchStyle.Render("/ ") + m.brSearch.View() + "\n\n")
	} else if m.brSearch.Value() != "" {
		s.WriteString("  " + dimStyle.Render(fmt.Sprintf("filter: %q (%d)", m.brSearch.Value(), len(m.brFiltered))) + "\n\n")
	}

	if m.loading && len(m.branchList) == 0 {
		s.WriteString(dimStyle.Render("  Loading branches...") + "\n")
		return s.String()
	}

	if len(m.branchList) == 0 {
		s.WriteString(dimStyle.Render("  No branches found.") + "\n")
		return s.String()
	}

	// Responsive: each branch item is 2 lines; reserve header + footer
	availLines := m.height - 12
	maxVisible := availLines / 2
	if maxVisible < 2 {
		maxVisible = 2
	}

	visibleStart := 0
	if m.brCursor >= maxVisible {
		visibleStart = m.brCursor - maxVisible + 1
	}
	end := visibleStart + maxVisible
	if end > len(m.brFiltered) {
		end = len(m.brFiltered)
	}

	// Build list content
	var listContent strings.Builder
	for vi := visibleStart; vi < end; vi++ {
		idx := m.brFiltered[vi]
		b := m.branchList[idx]
		isSelected := vi == m.brCursor
		listContent.WriteString(m.renderBranchItem(b, isSelected))
	}

	// Scroll info
	if len(m.brFiltered) > maxVisible {
		info := fmt.Sprintf(" %d of %d", min(maxVisible, len(m.brFiltered)), len(m.brFiltered))
		if visibleStart > 0 {
			info += " ↑"
		}
		if end < len(m.brFiltered) {
			info += " ↓"
		}
		listContent.WriteString(dimStyle.Render(info))
	} else if len(m.brFiltered) > 0 {
		listContent.WriteString(dimStyle.Render(fmt.Sprintf(" %d branches", len(m.brFiltered))))
	}

	// Box fills full content width
	listBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(dimGray).
		Padding(0, 1).
		Width(cw).
		Render(listContent.String())

	s.WriteString(listBox + "\n")

	s.WriteString(m.renderMessages())

	if m.showHelp {
		s.WriteString("\n")
		s.WriteString(m.brFullHelp())
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

	var tags []string
	if b.IsCurrent {
		tags = append(tags, tagActive.Render("ACTIVE"))
	}
	if b.Upstream != "" {
		tags = append(tags, tagTracked.Render("remote"))
	} else {
		tags = append(tags, tagLocal.Render("local"))
	}
	if b.Ahead > 0 {
		tags = append(tags, aheadStyle.Render(fmt.Sprintf("↑%d", b.Ahead)))
	}
	if b.Behind > 0 {
		tags = append(tags, behindStyle.Render(fmt.Sprintf("↓%d", b.Behind)))
	}

	line := cursor + name + "  " + strings.Join(tags, " ")
	detail := "    " + dimStyle.Render(b.CommitTime)

	return line + "\n" + detail + "\n"
}

func (m Model) viewBrDetail() string {
	var s strings.Builder
	cw := m.contentWidth()

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
		s.WriteString("\n  " + separatorStyle.Render(strings.Repeat("─", min(cw-4, 30))) + "\n\n")
		if b.Ahead > 0 {
			s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Ahead:"), aheadStyle.Render(fmt.Sprintf("↑ %d commits", b.Ahead))))
		}
		if b.Behind > 0 {
			s.WriteString(fmt.Sprintf("  %s %s\n", detailLabelStyle.Render("Behind:"), behindStyle.Render(fmt.Sprintf("↓ %d commits", b.Behind))))
		}
	}

	s.WriteString("\n")
	sep := separatorStyle.Render(strings.Repeat("─", cw))
	keys := []helpKey{{"c", "checkout"}}
	if !b.IsCurrent {
		keys = append(keys, helpKey{"d", "delete"})
	}
	keys = append(keys, helpKey{"esc", "back"})
	s.WriteString(sep + "\n  " + m.renderKeys(keys))

	return s.String()
}

func (m Model) viewBrCreate() string {
	var s strings.Builder
	cw := m.contentWidth()

	s.WriteString("  " + titleStyle.Render("Create Branch") + "\n\n")

	// Name field
	s.WriteString("  " + activeInputStyle.Render("Name") + "\n")
	s.WriteString(m.renderField(m.brCreateInput.View(), true) + "\n")

	// Base selector
	s.WriteString("\n  " + activeInputStyle.Render("Base") + "  " + dimStyle.Render("tab to cycle") + "\n")
	var baseRaw string
	for i, base := range m.brCreateBases {
		if i == m.brCreateBaseIdx {
			baseRaw += tagActive.Render(" " + base + " ")
		} else {
			baseRaw += dimStyle.Render(" " + base + " ")
		}
		if i < len(m.brCreateBases)-1 {
			baseRaw += " "
		}
	}
	s.WriteString(m.renderField(baseRaw, false) + "\n\n")

	sep := separatorStyle.Render(strings.Repeat("─", cw))
	s.WriteString(sep + "\n  " + m.renderKeys([]helpKey{{"enter", "create"}, {"tab", "cycle base"}, {"esc", "cancel"}}))
	return s.String()
}

func (m Model) viewBrRename() string {
	var s strings.Builder
	cw := m.contentWidth()
	b := m.selectedBranch()
	s.WriteString("  " + titleStyle.Render("Rename Branch") + "\n\n")
	if b != nil {
		s.WriteString("  " + dimStyle.Render("from: "+b.Name) + "\n\n")
	}

	s.WriteString("  " + activeInputStyle.Render("New name") + "\n")
	s.WriteString(m.renderField(m.brRenameInput.View(), true) + "\n\n")

	sep := separatorStyle.Render(strings.Repeat("─", cw))
	s.WriteString(sep + "\n  " + m.renderKeys([]helpKey{{"enter", "rename"}, {"esc", "cancel"}}))
	return s.String()
}

func (m Model) viewBrDelete() string {
	var s strings.Builder
	cw := m.contentWidth()

	s.WriteString("  " + titleStyle.Render("Delete Branch") + "\n\n")

	b := m.selectedBranch()
	if b == nil {
		return dimStyle.Render("  No branch selected.\n")
	}

	content := fmt.Sprintf("%s %s", detailLabelStyle.Render("Branch:"), errorStyle.Render(b.Name))
	forceLabel := dimStyle.Render("off")
	if m.brDeleteForce {
		forceLabel = errorStyle.Render("ON — unmerged commits will be lost")
	}
	content += fmt.Sprintf("\n\n%s %s", detailLabelStyle.Render("Force:"), forceLabel)

	boxWidth := cw - 4
	if boxWidth > 60 {
		boxWidth = 60
	}
	s.WriteString("  " + boxStyle.Width(boxWidth).Render(content) + "\n\n")

	sep := separatorStyle.Render(strings.Repeat("─", cw))
	s.WriteString(sep + "\n  " + m.renderKeys([]helpKey{{"y", "confirm"}, {"n/esc", "cancel"}, {"f", "force"}}))

	return s.String()
}

// ==========================================
// SHARED HELPERS
// ==========================================

func (m Model) renderMessages() string {
	if m.statusMsg != "" {
		return "\n  " + successStyle.Render("✓ "+m.statusMsg)
	}
	return ""
}

type helpKey struct {
	key  string
	desc string
}

func (m Model) renderKeys(keys []helpKey) string {
	cw := m.contentWidth()
	var parts []string
	lineLen := 0
	for _, k := range keys {
		partLen := len(k.key) + len(k.desc) + 4
		if lineLen+partLen > cw && len(parts) > 0 {
			break
		}
		parts = append(parts, helpKeyStyle.Render(k.key)+" "+helpStyle.Render(k.desc))
		lineLen += partLen
	}
	return strings.Join(parts, dimStyle.Render(" · "))
}

func (m Model) wtHelpKeys() []helpKey {
	return []helpKey{
		{"a", "add"}, {"d", "remove"}, {"o", "term"}, {"e", "edit"},
		{"/", "search"}, {"?", "help"}, {"q", "quit"},
	}
}

func (m Model) brHelpKeys() []helpKey {
	return []helpKey{
		{"c", "checkout"}, {"a", "create"}, {"d", "delete"}, {"m", "merge"},
		{"/", "search"}, {"?", "help"}, {"q", "quit"},
	}
}

func (m Model) wtFullHelp() string {
	var s strings.Builder
	s.WriteString("  " + subtitleStyle.Render("Keybindings") + "\n\n")
	sections := []struct {
		title string
		keys  []helpKey
	}{
		{"Navigate", []helpKey{{"j/↓", "down"}, {"k/↑", "up"}, {"←/→", "tab"}, {"enter", "details"}, {"/", "filter"}}},
		{"Actions", []helpKey{{"a", "add worktree"}, {"d/x", "remove"}, {"p", "prune"}}},
		{"Open", []helpKey{{"o", "terminal"}, {"e", "editor"}, {"f", "fetch"}, {"r", "refresh"}}},
		{"General", []helpKey{{"?", "close help"}, {"q", "quit"}, {"ctrl+c", "force quit"}}},
	}

	for _, sec := range sections {
		s.WriteString("  " + dimStyle.Render(sec.title) + "\n")
		for _, k := range sec.keys {
			keyCol := lipgloss.NewStyle().Width(10).Render(k.key)
			s.WriteString("    " + helpKeyStyle.Render(keyCol) + helpStyle.Render(k.desc) + "\n")
		}
		s.WriteString("\n")
	}
	return s.String()
}

func (m Model) brFullHelp() string {
	var s strings.Builder
	s.WriteString("  " + subtitleStyle.Render("Keybindings") + "\n\n")
	sections := []struct {
		title string
		keys  []helpKey
	}{
		{"Navigate", []helpKey{{"j/↓", "down"}, {"k/↑", "up"}, {"←/→", "tab"}, {"enter", "details"}, {"/", "filter"}}},
		{"Actions", []helpKey{{"c", "checkout"}, {"a/n", "create"}, {"R", "rename"}, {"d/x", "delete"}, {"m", "merge"}}},
		{"Tools", []helpKey{{"f", "fetch"}, {"r", "refresh"}}},
		{"General", []helpKey{{"?", "close help"}, {"q", "quit"}, {"ctrl+c", "force quit"}}},
	}

	for _, sec := range sections {
		s.WriteString("  " + dimStyle.Render(sec.title) + "\n")
		for _, k := range sec.keys {
			keyCol := lipgloss.NewStyle().Width(10).Render(k.key)
			s.WriteString("    " + helpKeyStyle.Render(keyCol) + helpStyle.Render(k.desc) + "\n")
		}
		s.WriteString("\n")
	}
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
