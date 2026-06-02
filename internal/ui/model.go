package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aryanpnd/git-wtm/internal/git"
)

type view int

const (
	viewList view = iota
	viewAdd
	viewRemoveConfirm
	viewDetail
)

type Model struct {
	worktrees    []git.Worktree
	filtered     []int
	branches     []string
	cursor       int
	currentView  view
	width        int
	height       int
	err          error
	statusMsg    string
	pathInput    textinput.Model
	branchInput  textinput.Model
	searchInput  textinput.Model
	addStep      int // 0=branch, 1=path
	addMatches   []string
	addCursor    int
	confirmForce bool
	searching    bool
	loading      bool
	showHelp     bool
}

type worktreeListMsg []git.Worktree
type branchListMsg []string
type errMsg struct{ err error }
type statusMsg string
type loadingMsg bool

func (e errMsg) Error() string { return e.err.Error() }

func NewModel() Model {
	pi := textinput.New()
	pi.Placeholder = "leave empty for default path"
	pi.CharLimit = 200

	bi := textinput.New()
	bi.Placeholder = "type branch name..."
	bi.CharLimit = 100

	si := textinput.New()
	si.Placeholder = "type to filter..."
	si.CharLimit = 100

	return Model{
		currentView: viewList,
		pathInput:   pi,
		branchInput: bi,
		searchInput: si,
		addCursor:   -1,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return loadingMsg(true) },
		fetchWorktrees,
		fetchBranches,
	)
}

func fetchWorktrees() tea.Msg {
	wts, err := git.ListWorktrees()
	if err != nil {
		return errMsg{err}
	}
	return worktreeListMsg(wts)
}

func fetchBranches() tea.Msg {
	branches, err := git.ListBranches()
	if err != nil {
		return errMsg{err}
	}
	return branchListMsg(branches)
}

func (m *Model) applyFilter() {
	query := strings.ToLower(m.searchInput.Value())
	m.filtered = nil
	for i, wt := range m.worktrees {
		if query == "" ||
			strings.Contains(strings.ToLower(wt.Branch), query) ||
			strings.Contains(strings.ToLower(wt.Path), query) {
			m.filtered = append(m.filtered, i)
		}
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m *Model) updateAddMatches() {
	query := strings.ToLower(m.branchInput.Value())
	m.addMatches = nil
	if query == "" {
		m.addCursor = -1
		return
	}
	for _, b := range m.branches {
		if strings.Contains(strings.ToLower(b), query) {
			m.addMatches = append(m.addMatches, b)
		}
	}
	if m.addCursor >= len(m.addMatches) {
		m.addCursor = len(m.addMatches) - 1
	}
	if m.addCursor < -1 {
		m.addCursor = -1
	}
}

func (m Model) selectedWorktree() *git.Worktree {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		return nil
	}
	idx := m.filtered[m.cursor]
	if idx >= len(m.worktrees) {
		return nil
	}
	return &m.worktrees[idx]
}

func (m Model) branchExists(name string) bool {
	for _, b := range m.branches {
		if b == name {
			return true
		}
	}
	return false
}
