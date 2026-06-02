package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aryanpnd/git-wtm/internal/git"
	"github.com/aryanpnd/git-wtm/internal/ui"
)

func main() {
	if _, err := git.GetRepoRoot(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: must be run inside a git repository")
		os.Exit(1)
	}

	p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
