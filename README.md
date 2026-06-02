# git-wtm

A modern, interactive TUI for managing Git worktrees. Browse, create, and remove worktrees with ease — no more memorizing `git worktree` commands.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)
![Platform](https://img.shields.io/badge/Platform-macOS%20|%20Linux%20|%20Windows-lightgrey)
![License](https://img.shields.io/badge/License-MIT-blue)

## Features

- **Visual worktree list** — see all worktrees with branch, status, commit info, and sync state at a glance
- **Smart branch picker** — type to search existing branches or create new ones inline
- **Status indicators** — colored tags show which worktree is active, primary, has unsaved changes, or is clean
- **One-key actions** — add, remove, open terminal/editor, fetch, prune — all from the keyboard
- **Auto-path** — creates worktrees at sensible default paths (skip the path prompt entirely)
- **Detail view** — drill into any worktree to see full commit, ahead/behind, and file change breakdown

## Install

### Quick install (any OS with Go)

```bash
go install github.com/aryanpnd/git-wtm@latest
```

Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your PATH:

```bash
# Add to ~/.zshrc, ~/.bashrc, or ~/.profile
export PATH="$PATH:$HOME/go/bin"
```

### From source

```bash
git clone https://github.com/aryanpnd/git-wtm.git
cd git-wtm
go build -o git-wtm .
sudo mv git-wtm /usr/local/bin/
```

### Homebrew (macOS/Linux)

```bash
# Coming soon
brew install aryanpnd/tap/git-wtm
```

## Usage

Run from any git repository:

```bash
git-wtm
```

Since the binary is prefixed with `git-`, Git also recognizes it as a subcommand:

```bash
git wtm
```

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | View worktree details |
| `/` | Search / filter worktrees |

### Actions

| Key | Action |
|-----|--------|
| `a` | Add worktree (search or create branch) |
| `d` | Remove worktree |
| `p` | Prune stale worktrees |

### Tools

| Key | Action |
|-----|--------|
| `o` | Open terminal in selected worktree |
| `e` | Open editor (`$EDITOR` or VS Code) |
| `f` | Fetch all remotes |
| `r` | Refresh worktree list |

### General

| Key | Action |
|-----|--------|
| `?` | Toggle full help |
| `q` | Quit |
| `Ctrl+C` | Force quit |

## Adding a worktree

Press `a` to open the add view:

1. **Type a branch name** — matching local branches appear as suggestions
2. **Pick a suggestion** with `↑`/`↓` and `Enter`, or just hit `Enter` to create a new branch
3. **Path is optional** — press `Tab` to customize, or let it auto-generate

The default path is `../<repo-name>-<branch-name>` (sibling to your repo directory).

## Requirements

- Git 2.15+ (worktree support)
- Go 1.21+ (for building from source)
- A terminal with 256-color support

## License

MIT
