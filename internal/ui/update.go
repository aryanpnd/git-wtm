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
		m.applyFilter()
		return m, nil

	case branchListMsg:
		m.branches = msg
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
	case "a", "n":
		m.currentView = viewAdd
		m.addStep = 0
		m.branchInput.Reset()
		m.pathInput.Reset()
		m.branchInput.Focus()
		m.addMatches = nil
		m.addCursor = -1
		m.statusMsg = ""
		m.err = nil
		return m, textinput.Blink
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
	if m.addStep == 0 {
		switch msg.String() {
		case "esc":
			m.currentView = viewList
			return m, nil
		case "tab":
			m.addStep = 1
			m.branchInput.Blur()
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
			branch := m.branchInput.Value()
			if m.addCursor >= 0 && m.addCursor < len(m.addMatches) {
				branch = m.addMatches[m.addCursor]
			}
			if branch == "" {
				return m, nil
			}
			path := m.pathInput.Value()
			createNew := !m.branchExists(branch)
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
				action := "Checked out"
				if createNew {
					action = "Created"
				}
				return statusMsg(action + " branch → " + displayPath)
			}
		default:
			var cmd tea.Cmd
			m.branchInput, cmd = m.branchInput.Update(msg)
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
			m.branchInput.Focus()
			return m, textinput.Blink
		case "tab":
			m.addStep = 0
			m.pathInput.Blur()
			m.branchInput.Focus()
			return m, textinput.Blink
		case "enter":
			branch := m.branchInput.Value()
			if branch == "" {
				return m, nil
			}
			path := m.pathInput.Value()
			createNew := !m.branchExists(branch)
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
				action := "Checked out"
				if createNew {
					action = "Created"
				}
				return statusMsg(action + " branch → " + displayPath)
			}
		default:
			var cmd tea.Cmd
			m.pathInput, cmd = m.pathInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
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
