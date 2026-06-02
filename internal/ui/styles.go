package ui

import "github.com/charmbracelet/lipgloss"

var (
	purple  = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	green   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	red     = lipgloss.AdaptiveColor{Light: "#FF6347", Dark: "#FF6347"}
	yellow  = lipgloss.AdaptiveColor{Light: "#FFBF00", Dark: "#FFD700"}
	blue    = lipgloss.AdaptiveColor{Light: "#5B9BD5", Dark: "#89CFF0"}
	dimGray = lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#626262"}

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(purple).
			Padding(0, 1).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(purple).
			MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(purple).
				Bold(true)

	currentBadge = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	branchStyle = lipgloss.NewStyle().
			Foreground(purple)

	commitStyle = lipgloss.NewStyle().
			Foreground(yellow)

	pathStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	dimStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(yellow)

	helpStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true)

	statusDirty = lipgloss.NewStyle().
			Foreground(red)

	statusClean = lipgloss.NewStyle().
			Foreground(green)

	aheadStyle = lipgloss.NewStyle().
			Foreground(blue)

	behindStyle = lipgloss.NewStyle().
			Foreground(red)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purple).
			Padding(1, 2)

	activeInputStyle = lipgloss.NewStyle().
				Foreground(purple).
				Bold(true)

	inactiveInputStyle = lipgloss.NewStyle().
				Foreground(dimGray)

	searchStyle = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true)

	detailLabelStyle = lipgloss.NewStyle().
				Foreground(purple).
				Bold(true).
				Width(12)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "#DDDDDD"})
)
