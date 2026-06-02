package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aryan/worktree-manager/internal/git"
)

type view int

const (
	viewList view = iota
	viewAdd
	viewRemoveConfirm
	viewBranchSelect
	viewDetail
)

type Model struct {
	worktrees      []git.Worktree
	filtered       []int
	branches       []string
	filteredBranch []int
	cursor         int
	currentView    view
	width          int
	height         int
	err            error
	statusMsg      string
	pathInput      textinput.Model
	branchInput    textinput.Model
	searchInput    textinput.Model
	branchSearch   textinput.Model
	addStep        int
	createNew      bool
	branchCursor   int
	confirmForce   bool
	searching      bool
	loading        bool
	showHelp       bool
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
	bi.Placeholder = "branch-name"
	bi.CharLimit = 100

	si := textinput.New()
	si.Placeholder = "type to filter..."
	si.CharLimit = 100

	bs := textinput.New()
	bs.Placeholder = "search branches..."
	bs.CharLimit = 100

	return Model{
		currentView:  viewList,
		pathInput:    pi,
		branchInput:  bi,
		searchInput:  si,
		branchSearch: bs,
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

func (m *Model) applyBranchFilter() {
	query := strings.ToLower(m.branchSearch.Value())
	m.filteredBranch = nil
	for i, b := range m.branches {
		if query == "" || strings.Contains(strings.ToLower(b), query) {
			m.filteredBranch = append(m.filteredBranch, i)
		}
	}
	if m.branchCursor >= len(m.filteredBranch) {
		m.branchCursor = max(0, len(m.filteredBranch)-1)
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
