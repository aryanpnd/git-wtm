package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aryan/worktree-manager/internal/git"
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
		m.applyFilter()
		return m, nil

	case branchListMsg:
		m.branches = msg
		m.applyBranchFilter()
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case statusMsg:
		m.statusMsg = string(msg)
		m.err = nil
		return m, fetchWorktrees

	case loadingMsg:
		m.loading = bool(msg)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.currentView {
		case viewList:
			return m.updateList(msg)
		case viewAdd:
			return m.updateAdd(msg)
		case viewRemoveConfirm:
			return m.updateRemoveConfirm(msg)
		case viewBranchSelect:
			return m.updateBranchSelect(msg)
		case viewDetail:
			return m.updateDetail(msg)
		}
	}

	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.searching {
		switch msg.String() {
		case "esc":
			m.searching = false
			m.searchInput.Blur()
			m.searchInput.Reset()
			m.applyFilter()
			return m, nil
		case "enter":
			m.searching = false
			m.searchInput.Blur()
			return m, nil
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			m.applyFilter()
			return m, cmd
		}
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}
	case "/":
		m.searching = true
		m.searchInput.Focus()
		return m, textinput.Blink
	case "a":
		m.currentView = viewAdd
		m.addStep = 0
		m.createNew = false
		m.branchInput.Reset()
		m.pathInput.Reset()
		m.branchInput.Focus()
		m.statusMsg = ""
		m.err = nil
		return m, textinput.Blink
	case "n":
		m.currentView = viewAdd
		m.addStep = 0
		m.createNew = true
		m.branchInput.Reset()
		m.pathInput.Reset()
		m.branchInput.Focus()
		m.statusMsg = ""
		m.err = nil
		return m, textinput.Blink
	case "b":
		m.currentView = viewBranchSelect
		m.branchCursor = 0
		m.branchSearch.Reset()
		m.branchSearch.Focus()
		m.applyBranchFilter()
		m.statusMsg = ""
		m.err = nil
		return m, tea.Batch(fetchBranches, textinput.Blink)
	case "d", "x":
		wt := m.selectedWorktree()
		if wt != nil && !wt.IsCurrent {
			m.currentView = viewRemoveConfirm
			m.confirmForce = false
			m.statusMsg = ""
			m.err = nil
		}
	case "enter":
		if m.selectedWorktree() != nil {
			m.currentView = viewDetail
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

func (m Model) updateAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.currentView = viewList
		return m, nil
	case "enter":
		if m.addStep == 0 {
			if m.branchInput.Value() == "" {
				return m, nil
			}
			m.addStep = 1
			m.branchInput.Blur()
			m.pathInput.Focus()
			return m, textinput.Blink
		}
		if m.addStep == 1 {
			branch := m.branchInput.Value()
			path := m.pathInput.Value()
			createNew := m.createNew
			m.currentView = viewList
			m.loading = true
			return m, func() tea.Msg {
				if err := git.AddWorktree(path, branch, createNew); err != nil {
					return errMsg{err}
				}
				displayPath := path
				if displayPath == "" {
					displayPath = git.DefaultWorktreePath(branch)
				}
				return statusMsg("Created worktree: " + displayPath)
			}
		}
	case "tab", "shift+tab":
		if m.addStep == 0 {
			m.addStep = 1
			m.branchInput.Blur()
			m.pathInput.Focus()
			return m, textinput.Blink
		}
		m.addStep = 0
		m.pathInput.Blur()
		m.branchInput.Focus()
		return m, textinput.Blink
	}

	var cmd tea.Cmd
	if m.addStep == 0 {
		m.branchInput, cmd = m.branchInput.Update(msg)
	} else {
		m.pathInput, cmd = m.pathInput.Update(msg)
	}
	return m, cmd
}

func (m Model) updateRemoveConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		m.currentView = viewList
	case "y":
		wt := m.selectedWorktree()
		if wt == nil {
			m.currentView = viewList
			return m, nil
		}
		path := wt.Path
		force := m.confirmForce
		m.currentView = viewList
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

func (m Model) updateBranchSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.currentView = viewList
		m.branchSearch.Blur()
		return m, nil
	case "up", "ctrl+p":
		if m.branchCursor > 0 {
			m.branchCursor--
		}
	case "down", "ctrl+n":
		if m.branchCursor < len(m.filteredBranch)-1 {
			m.branchCursor++
		}
	case "enter":
		if len(m.filteredBranch) > 0 && m.branchCursor < len(m.filteredBranch) {
			branch := m.branches[m.filteredBranch[m.branchCursor]]
			m.currentView = viewAdd
			m.addStep = 1
			m.createNew = false
			m.branchInput.SetValue(branch)
			m.pathInput.Reset()
			m.branchInput.Blur()
			m.branchSearch.Blur()
			m.pathInput.Focus()
			return m, textinput.Blink
		}
	default:
		var cmd tea.Cmd
		m.branchSearch, cmd = m.branchSearch.Update(msg)
		m.applyBranchFilter()
		return m, cmd
	}
	return m, nil
}

func (m Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.currentView = viewList
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
			m.currentView = viewRemoveConfirm
			m.confirmForce = false
		}
	}
	return m, nil
}
