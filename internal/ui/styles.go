package ui

import "github.com/charmbracelet/lipgloss"

var (
	purple  = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	green   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	red     = lipgloss.AdaptiveColor{Light: "#FF6347", Dark: "#FF6347"}
	yellow  = lipgloss.AdaptiveColor{Light: "#FFBF00", Dark: "#FFD700"}
	blue    = lipgloss.AdaptiveColor{Light: "#5B9BD5", Dark: "#89CFF0"}
	cyan    = lipgloss.AdaptiveColor{Light: "#00ACC1", Dark: "#4DD0E1"}
	dimGray = lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#626262"}

	// Header
	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#2D9C6F"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#2D9C6F", Dark: "#73F59F"})

	loadingStyle = lipgloss.NewStyle().
			Foreground(yellow).
			Bold(true)

	// Tabs
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(purple).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(dimGray).
				Padding(0, 2)

	// List items
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(purple).
				Bold(true)

	// Branch display
	branchStyle = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true)

	commitStyle = lipgloss.NewStyle().
			Foreground(yellow)

	pathStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	dimStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	// Messages
	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(yellow)

	// Help
	helpStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true)

	// Status indicators
	statusDirty = lipgloss.NewStyle().
			Foreground(red)

	statusClean = lipgloss.NewStyle().
			Foreground(green)

	aheadStyle = lipgloss.NewStyle().
			Foreground(blue)

	behindStyle = lipgloss.NewStyle().
			Foreground(red)

	// Tags
	tagPrimary = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#6C5CE7")).
			Padding(0, 1)

	tagActive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(green).
			Padding(0, 1)

	tagDetached = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(red).
			Padding(0, 1)

	tagUnsaved = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1a1a1a")).
			Background(yellow).
			Padding(0, 1)

	tagClean = lipgloss.NewStyle().
			Foreground(green)

	tagTracked = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(cyan).
			Padding(0, 1)

	tagLocal = lipgloss.NewStyle().
			Foreground(dimGray).
			Italic(true)

	// Cards
	cardBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimGray).
			Padding(0, 1)

	cardBorderSelected = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(purple).
				Padding(0, 1)

	cardBorderActive = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(green).
				Padding(0, 1)

	// Inputs
	activeInputStyle = lipgloss.NewStyle().
				Foreground(purple).
				Bold(true)

	inactiveInputStyle = lipgloss.NewStyle().
				Foreground(dimGray)

	focusedFieldStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(purple).
				Padding(0, 1)

	unfocusedFieldStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(dimGray).
				Padding(0, 1)

	searchStyle = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true)

	// Detail view
	detailLabelStyle = lipgloss.NewStyle().
				Foreground(cyan).
				Bold(true).
				Width(12)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "#DDDDDD"})

	separatorStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purple).
			Padding(1, 2)

	subtitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(purple)

	updateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1a1a1a")).
			Background(yellow).
			Bold(true)

	// Modal
	modalErrorStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(red).
			Padding(1, 3)

	modalSuccessStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(green).
				Padding(1, 3)

	modalTitleError = lipgloss.NewStyle().
			Bold(true).
			Foreground(red)

	modalTitleSuccess = lipgloss.NewStyle().
				Bold(true).
				Foreground(green)
)
