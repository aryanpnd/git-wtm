package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aryanpnd/git-wtm/internal/git"
	"github.com/aryanpnd/git-wtm/internal/ui"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Printf("git-wtm %s (%s)\n", version, commit)
			return
		case "--help", "-h", "help":
			fmt.Println("git-wtm — Modern TUI for managing Git worktrees and branches")
			fmt.Println()
			fmt.Println("Usage: git-wtm    (or: git wtm)")
			fmt.Println()
			fmt.Println("Flags:")
			fmt.Println("  -v, --version   Print version")
			fmt.Println("  -h, --help      Show this help")
			return
		}
	}

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
