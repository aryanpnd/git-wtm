package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aryanpnd/git-wtm/internal/git"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case worktreeListMsg:
		m.worktrees = msg
		m.loading = false
		m.applyWtFilter()
		return m, nil

	case branchListMsg:
		m.branches = msg
		return m, nil

	case branchDetailMsg:
		m.branchList = msg
		m.loading = false
		m.applyBrFilter()
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case statusMsg:
		m.statusMsg = string(msg)
		m.err = nil
		return m, tea.Batch(fetchWorktrees, fetchBranches, fetchBranchDetails)

	case loadingMsg:
		m.loading = bool(msg)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Tab switching (only in list views, not in sub-views)
		if m.isInListView() && !m.isSearching() {
			switch msg.String() {
			case "left", "h":
				if m.activeTab > 0 {
					m.activeTab--
					m.statusMsg = ""
					m.err = nil
				}
				return m, nil
			case "right", "l":
				if m.activeTab < tabBranches {
					m.activeTab++
					m.statusMsg = ""
					m.err = nil
				}
				return m, nil
			}
		}

		switch m.activeTab {
		case tabWorktrees:
			return m.updateWorktrees(msg)
		case tabBranches:
			return m.updateBranches(msg)
		}
	}

	return m, nil
}

func (m Model) isInListView() bool {
	if m.activeTab == tabWorktrees {
		return m.wtView == viewList
	}
	return m.brView == viewList
}

func (m Model) isSearching() bool {
	if m.activeTab == tabWorktrees {
		return m.wtSearching
	}
	return m.brSearching
}

// --- Worktrees tab ---

func (m Model) updateWorktrees(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.wtView {
	case viewList:
		return m.updateWtList(msg)
	case viewAdd:
		return m.updateWtAdd(msg)
	case viewRemoveConfirm:
		return m.updateWtRemove(msg)
	case viewDetail:
		return m.updateWtDetail(msg)
	}
	return m, nil
}

func (m Model) updateWtList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.wtSearching {
		switch msg.String() {
		case "esc":
			m.wtSearching = false
			m.wtSearch.Blur()
			m.wtSearch.Reset()
			m.applyWtFilter()
			return m, nil
		case "enter":
			m.wtSearching = false
			m.wtSearch.Blur()
			return m, nil
		default:
			var cmd tea.Cmd
			m.wtSearch, cmd = m.wtSearch.Update(msg)
			m.applyWtFilter()
			return m, cmd
		}
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "up", "k":
		if m.wtCursor > 0 {
			m.wtCursor--
		}
	case "down", "j":
		if m.wtCursor < len(m.wtFiltered)-1 {
			m.wtCursor++
		}
	case "/":
		m.wtSearching = true
		m.wtSearch.Focus()
		return m, textinput.Blink
	case "a":
		m.wtView = viewAdd
		m.addStep = 0
		m.addInput.Reset()
		m.pathInput.Reset()
		m.addInput.Focus()
		m.addMatches = nil
		m.addCursor = -1
		m.statusMsg = ""
		m.err = nil
		return m, textinput.Blink
	case "d", "x":
		wt := m.selectedWorktree()
		if wt != nil && !wt.IsCurrent {
			m.wtView = viewRemoveConfirm
			m.confirmForce = false
			m.statusMsg = ""
			m.err = nil
		}
	case "enter":
		if m.selectedWorktree() != nil {
			m.wtView = viewDetail
			m.statusMsg = ""
			m.err = nil
		}
	case "o":
		wt := m.selectedWorktree()
		if wt != nil {
			return m, func() tea.Msg {
				if err := git.OpenShell(wt.Path); err != nil {
					return errMsg{err}
				}
				return statusMsg("Opened terminal at " + wt.Path)
			}
		}
	case "e":
		wt := m.selectedWorktree()
		if wt != nil {
			return m, func() tea.Msg {
				if err := git.OpenInEditor(wt.Path); err != nil {
					return errMsg{err}
				}
				return statusMsg("Opened editor at " + wt.Path)
			}
		}
	case "p":
		m.statusMsg = ""
		m.err = nil
		return m, func() tea.Msg {
			if err := git.PruneWorktrees(); err != nil {
				return errMsg{err}
			}
			return statusMsg("Pruned stale worktrees")
		}
	case "f":
		m.statusMsg = ""
		m.err = nil
		m.loading = true
		return m, func() tea.Msg {
			if err := git.FetchRemote(); err != nil {
				return errMsg{err}
			}
			return statusMsg("Fetched all remotes")
		}
	case "r":
		m.statusMsg = ""
		m.err = nil
		m.loading = true
		return m, fetchWorktrees
	case "?":
		m.showHelp = !m.showHelp
	}
	return m, nil
}

func (m Model) updateWtAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.addStep == 0 {
		switch msg.String() {
		case "esc":
			m.wtView = viewList
			return m, nil
		case "tab":
			m.addStep = 1
			m.addInput.Blur()
			m.pathInput.Focus()
			return m, textinput.Blink
		case "up", "ctrl+p":
			if m.addCursor > 0 {
				m.addCursor--
			} else if m.addCursor == 0 {
				m.addCursor = -1
			}
			return m, nil
		case "down", "ctrl+n":
			if m.addCursor < len(m.addMatches)-1 {
				m.addCursor++
			}
			return m, nil
		case "enter":
			branch := m.addInput.Value()
			if m.addCursor >= 0 && m.addCursor < len(m.addMatches) {
				branch = m.addMatches[m.addCursor]
			}
			if branch == "" {
				return m, nil
			}
			path := m.pathInput.Value()
			createNew := !m.branchExists(branch)
			m.wtView = viewList
			m.loading = true
			return m, func() tea.Msg {
				if err := git.AddWorktree(path, branch, createNew); err != nil {
					return errMsg{err}
				}
				displayPath := path
				if displayPath == "" {
					displayPath = git.DefaultWorktreePath(branch)
				}
				action := "Checked out"
				if createNew {
					action = "Created"
				}
				return statusMsg(action + " → " + displayPath)
			}
		default:
			var cmd tea.Cmd
			m.addInput, cmd = m.addInput.Update(msg)
			m.addCursor = -1
			m.updateAddMatches()
			return m, cmd
		}
	}

	if m.addStep == 1 {
		switch msg.String() {
		case "esc", "shift+tab":
			m.addStep = 0
			m.pathInput.Blur()
			m.addInput.Focus()
			return m, textinput.Blink
		case "tab":
			m.addStep = 0
			m.pathInput.Blur()
			m.addInput.Focus()
			return m, textinput.Blink
		case "enter":
			branch := m.addInput.Value()
			if branch == "" {
				return m, nil
			}
			path := m.pathInput.Value()
			createNew := !m.branchExists(branch)
			m.wtView = viewList
			m.loading = true
			return m, func() tea.Msg {
				if err := git.AddWorktree(path, branch, createNew); err != nil {
					return errMsg{err}
				}
				displayPath := path
				if displayPath == "" {
					displayPath = git.DefaultWorktreePath(branch)
				}
				action := "Checked out"
				if createNew {
					action = "Created"
				}
				return statusMsg(action + " → " + displayPath)
			}
		default:
			var cmd tea.Cmd
			m.pathInput, cmd = m.pathInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) updateWtRemove(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		m.wtView = viewList
	case "y":
		wt := m.selectedWorktree()
		if wt == nil {
			m.wtView = viewList
			return m, nil
		}
		path := wt.Path
		force := m.confirmForce
		m.wtView = viewList
		m.loading = true
		return m, func() tea.Msg {
			if err := git.RemoveWorktree(path, force); err != nil {
				return errMsg{err}
			}
			return statusMsg("Removed: " + path)
		}
	case "f":
		m.confirmForce = !m.confirmForce
	}
	return m, nil
}

func (m Model) updateWtDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.wtView = viewList
	case "o":
		wt := m.selectedWorktree()
		if wt != nil {
			return m, func() tea.Msg {
				if err := git.OpenShell(wt.Path); err != nil {
					return errMsg{err}
				}
				return statusMsg("Opened terminal")
			}
		}
	case "e":
		wt := m.selectedWorktree()
		if wt != nil {
			return m, func() tea.Msg {
				if err := git.OpenInEditor(wt.Path); err != nil {
					return errMsg{err}
				}
				return statusMsg("Opened editor")
			}
		}
	case "d":
		wt := m.selectedWorktree()
		if wt != nil && !wt.IsCurrent {
			m.wtView = viewRemoveConfirm
			m.confirmForce = false
		}
	}
	return m, nil
}

// --- Branches tab ---

func (m Model) updateBranches(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.brView {
	case viewList:
		return m.updateBrList(msg)
	case viewBranchDetail:
		return m.updateBrDetail(msg)
	case viewBranchCreate:
		return m.updateBrCreate(msg)
	case viewBranchRename:
		return m.updateBrRename(msg)
	case viewBranchDelete:
		return m.updateBrDelete(msg)
	}
	return m, nil
}

func (m Model) updateBrList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.brSearching {
		switch msg.String() {
		case "esc":
			m.brSearching = false
			m.brSearch.Blur()
			m.brSearch.Reset()
			m.applyBrFilter()
			return m, nil
		case "enter":
			m.brSearching = false
			m.brSearch.Blur()
			return m, nil
		default:
			var cmd tea.Cmd
			m.brSearch, cmd = m.brSearch.Update(msg)
			m.applyBrFilter()
			return m, cmd
		}
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "up", "k":
		if m.brCursor > 0 {
			m.brCursor--
		}
	case "down", "j":
		if m.brCursor < len(m.brFiltered)-1 {
			m.brCursor++
		}
	case "/":
		m.brSearching = true
		m.brSearch.Focus()
		return m, textinput.Blink
	case "enter":
		if m.selectedBranch() != nil {
			m.brView = viewBranchDetail
			m.statusMsg = ""
			m.err = nil
		}
	case "c":
		b := m.selectedBranch()
		if b != nil {
			m.loading = true
			name := b.Name
			return m, func() tea.Msg {
				if err := git.CheckoutBranch(name); err != nil {
					return errMsg{err}
				}
				return statusMsg("Switched to " + name)
			}
		}
	case "a", "n":
		m.brView = viewBranchCreate
		m.brCreateInput.Reset()
		m.brCreateInput.Focus()
		m.statusMsg = ""
		m.err = nil
		return m, textinput.Blink
	case "R":
		b := m.selectedBranch()
		if b != nil && !b.IsCurrent {
			m.brView = viewBranchRename
			m.brRenameInput.Reset()
			m.brRenameInput.SetValue(b.Name)
			m.brRenameInput.Focus()
			m.statusMsg = ""
			m.err = nil
			return m, textinput.Blink
		}
	case "d", "x":
		b := m.selectedBranch()
		if b != nil && !b.IsCurrent {
			m.brView = viewBranchDelete
			m.brDeleteForce = false
			m.statusMsg = ""
			m.err = nil
		}
	case "m":
		b := m.selectedBranch()
		if b != nil && !b.IsCurrent {
			name := b.Name
			m.loading = true
			return m, func() tea.Msg {
				if err := git.MergeBranch(name); err != nil {
					return errMsg{err}
				}
				return statusMsg("Merged " + name + " into current branch")
			}
		}
	case "f":
		m.statusMsg = ""
		m.err = nil
		m.loading = true
		return m, func() tea.Msg {
			if err := git.FetchRemote(); err != nil {
				return errMsg{err}
			}
			return statusMsg("Fetched all remotes")
		}
	case "r":
		m.statusMsg = ""
		m.err = nil
		m.loading = true
		return m, fetchBranchDetails
	case "?":
		m.showHelp = !m.showHelp
	}
	return m, nil
}

func (m Model) updateBrDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.brView = viewList
	case "c":
		b := m.selectedBranch()
		if b != nil {
			name := b.Name
			m.brView = viewList
			m.loading = true
			return m, func() tea.Msg {
				if err := git.CheckoutBranch(name); err != nil {
					return errMsg{err}
				}
				return statusMsg("Switched to " + name)
			}
		}
	case "d":
		b := m.selectedBranch()
		if b != nil && !b.IsCurrent {
			m.brView = viewBranchDelete
			m.brDeleteForce = false
		}
	}
	return m, nil
}

func (m Model) updateBrCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.brView = viewList
		return m, nil
	case "enter":
		name := m.brCreateInput.Value()
		if name == "" {
			return m, nil
		}
		m.brView = viewList
		m.loading = true
		return m, func() tea.Msg {
			if err := git.CreateBranch(name, ""); err != nil {
				return errMsg{err}
			}
			return statusMsg("Created branch: " + name)
		}
	default:
		var cmd tea.Cmd
		m.brCreateInput, cmd = m.brCreateInput.Update(msg)
		return m, cmd
	}
}

func (m Model) updateBrRename(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.brView = viewList
		return m, nil
	case "enter":
		newName := m.brRenameInput.Value()
		b := m.selectedBranch()
		if newName == "" || b == nil {
			return m, nil
		}
		oldName := b.Name
		m.brView = viewList
		m.loading = true
		return m, func() tea.Msg {
			if err := git.RenameBranch(oldName, newName); err != nil {
				return errMsg{err}
			}
			return statusMsg("Renamed: " + oldName + " → " + newName)
		}
	default:
		var cmd tea.Cmd
		m.brRenameInput, cmd = m.brRenameInput.Update(msg)
		return m, cmd
	}
}

func (m Model) updateBrDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		m.brView = viewList
	case "y":
		b := m.selectedBranch()
		if b == nil {
			m.brView = viewList
			return m, nil
		}
		name := b.Name
		force := m.brDeleteForce
		m.brView = viewList
		m.loading = true
		return m, func() tea.Msg {
			if err := git.DeleteBranch(name, force); err != nil {
				return errMsg{err}
			}
			return statusMsg("Deleted branch: " + name)
		}
	case "f":
		m.brDeleteForce = !m.brDeleteForce
	}
	return m, nil
}
