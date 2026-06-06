package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aryanpnd/git-wtm/internal/git"
)

type tab int

const (
	tabWorktrees tab = iota
	tabBranches
)

type view int

const (
	viewList view = iota
	viewAdd
	viewRemoveConfirm
	viewDetail
	viewBranchDetail
	viewBranchCreate
	viewBranchRename
	viewBranchDelete
)

type Model struct {
	// Global
	activeTab  tab
	width      int
	height     int
	err        error
	statusMsg  string
	loading    bool
	showHelp   bool
	updateInfo *git.UpdateInfo

	// Modal popup
	showModal    bool
	modalTitle   string
	modalMessage string
	modalIsError bool

	// Worktrees tab
	worktrees   []git.Worktree
	wtFiltered  []int
	wtCursor    int
	wtView      view
	wtSearch    textinput.Model
	wtSearching bool

	// Add worktree
	branches      []string
	addInput      textinput.Model
	pathInput     textinput.Model
	addMatches    []string
	addCursor     int
	addStep       int
	wtAddBaseIdx  int
	wtAddBases    []string

	// Remove worktree
	confirmForce bool

	// Branches tab
	branchList      []git.Branch
	brFiltered      []int
	brCursor        int
	brView          view
	brSearch        textinput.Model
	brSearching     bool
	brCreateInput   textinput.Model
	brRenameInput   textinput.Model
	brDeleteForce   bool
	brCreateBaseIdx int    // index into brCreateBases
	brCreateBases   []string
}

type worktreeListMsg []git.Worktree
type branchListMsg []string
type branchDetailMsg []git.Branch
type errMsg struct{ err error }
type statusMsg string
type loadingMsg bool
type folderPickedMsg string
type updateCheckMsg *git.UpdateInfo
type modalMsg struct {
	title   string
	message string
	isError bool
}

func (e errMsg) Error() string { return e.err.Error() }

var appVersion string

func NewModel(version string) Model {
	appVersion = version
	wtSearch := textinput.New()
	wtSearch.Placeholder = "filter worktrees..."
	wtSearch.CharLimit = 100

	addInput := textinput.New()
	addInput.Placeholder = "type branch name..."
	addInput.CharLimit = 100

	pathInput := textinput.New()
	pathInput.Placeholder = "leave empty for default path"
	pathInput.CharLimit = 200

	brSearch := textinput.New()
	brSearch.Placeholder = "filter branches..."
	brSearch.CharLimit = 100

	brCreate := textinput.New()
	brCreate.Placeholder = "new-branch-name"
	brCreate.CharLimit = 100

	brRename := textinput.New()
	brRename.Placeholder = "new name..."
	brRename.CharLimit = 100

	return Model{
		activeTab:     tabWorktrees,
		wtView:        viewList,
		brView:        viewList,
		wtSearch:      wtSearch,
		addInput:      addInput,
		pathInput:     pathInput,
		brSearch:      brSearch,
		brCreateInput: brCreate,
		brRenameInput: brRename,
		addCursor:     -1,
		loading:       true,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchWorktrees,
		fetchBranches,
		fetchBranchDetails,
		checkForUpdate,
	)
}

func checkForUpdate() tea.Msg {
	info, _ := git.CheckForUpdate(appVersion)
	return updateCheckMsg(info)
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

func fetchBranchDetails() tea.Msg {
	branches, err := git.ListBranchesDetailed()
	if err != nil {
		return errMsg{err}
	}
	return branchDetailMsg(branches)
}

func (m *Model) applyWtFilter() {
	query := strings.ToLower(m.wtSearch.Value())
	m.wtFiltered = nil
	for i, wt := range m.worktrees {
		if query == "" ||
			strings.Contains(strings.ToLower(wt.Branch), query) ||
			strings.Contains(strings.ToLower(wt.Path), query) {
			m.wtFiltered = append(m.wtFiltered, i)
		}
	}
	if m.wtCursor >= len(m.wtFiltered) {
		m.wtCursor = max(0, len(m.wtFiltered)-1)
	}
}

func (m *Model) applyBrFilter() {
	query := strings.ToLower(m.brSearch.Value())
	m.brFiltered = nil
	for i, b := range m.branchList {
		if query == "" || strings.Contains(strings.ToLower(b.Name), query) {
			m.brFiltered = append(m.brFiltered, i)
		}
	}
	if m.brCursor >= len(m.brFiltered) {
		m.brCursor = max(0, len(m.brFiltered)-1)
	}
}

func (m *Model) updateAddMatches() {
	query := strings.ToLower(m.addInput.Value())
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
}

func (m Model) selectedWorktree() *git.Worktree {
	if len(m.wtFiltered) == 0 || m.wtCursor >= len(m.wtFiltered) {
		return nil
	}
	idx := m.wtFiltered[m.wtCursor]
	if idx >= len(m.worktrees) {
		return nil
	}
	return &m.worktrees[idx]
}

func (m Model) selectedBranch() *git.Branch {
	if len(m.brFiltered) == 0 || m.brCursor >= len(m.brFiltered) {
		return nil
	}
	idx := m.brFiltered[m.brCursor]
	if idx >= len(m.branchList) {
		return nil
	}
	return &m.branchList[idx]
}

func buildCreateBases(branches []git.Branch) []string {
	for _, b := range branches {
		if b.Name == "main" || b.Name == "master" {
			return []string{b.Name + " (latest)", b.Name, "current HEAD"}
		}
	}
	return []string{"current HEAD"}
}

func (m Model) branchExists(name string) bool {
	for _, b := range m.branches {
		if b == name {
			return true
		}
	}
	return false
}

func (m *Model) updateFieldWidths() {
	// inner width = terminal - 2(margin) - 2(border) - 2(padding)
	w := m.width - 6
	if w < 20 {
		w = 20
	}
	m.addInput.Width = w
	m.pathInput.Width = w
	m.brCreateInput.Width = w
	m.brRenameInput.Width = w
	m.wtSearch.Width = w
	m.brSearch.Width = w
}
