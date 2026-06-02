<p align="center">
  <img src="assets/logo.svg" width="120" alt="git-wtm logo" />
</p>

<h1 align="center">git-wtm</h1>

<p align="center">
  <strong>A modern, interactive TUI for managing Git worktrees and branches.</strong><br/>
  No more memorizing commands — just navigate, search, and act.
</p>

<p align="center">
  <a href="https://github.com/aryanpnd/git-wtm/releases/latest"><img src="https://img.shields.io/github/v/release/aryanpnd/git-wtm?style=flat-square&color=73daca" alt="Release"></a>
  <a href="https://github.com/aryanpnd/git-wtm/actions"><img src="https://img.shields.io/github/actions/workflow/status/aryanpnd/git-wtm/release.yml?style=flat-square" alt="Build"></a>
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Platform-macOS%20|%20Linux%20|%20Windows-lightgrey?style=flat-square" alt="Platform">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue?style=flat-square" alt="License"></a>
</p>

---

## Installation

### macOS / Linux — Homebrew

```bash
brew install aryanpnd/tap/git-wtm
```

### macOS / Linux — Shell Script

```bash
curl -fsSL https://raw.githubusercontent.com/aryanpnd/git-wtm/main/install/install.sh | sh
```

### Windows — Scoop

```powershell
scoop bucket add aryanpnd https://github.com/aryanpnd/scoop-bucket
scoop install git-wtm
```

### Windows — Chocolatey

```powershell
choco install git-wtm
```

### Windows — Winget

```powershell
winget install aryanpnd.git-wtm
```

### Any OS — Go Install

```bash
go install github.com/aryanpnd/git-wtm@latest
```

Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your PATH:

```bash
export PATH="$PATH:$HOME/go/bin"
```

### Download Binary

Pre-built binaries for all platforms are on the [Releases](https://github.com/aryanpnd/git-wtm/releases/latest) page.

### Build from Source

```bash
git clone https://github.com/aryanpnd/git-wtm.git
cd git-wtm
go build -o git-wtm .
sudo mv git-wtm /usr/local/bin/
```

---

## Quick Start

Run from any git repository:

```bash
git-wtm
```

Since the binary is prefixed with `git-`, Git also recognizes it as a subcommand:

```bash
git wtm
```

---

## Features

### Worktree Manager

- **Visual worktree list** — see all worktrees with branch, status, commit info, and sync state at a glance
- **Smart add** — type to search existing branches or create new ones inline, auto-generates paths
- **Status tags** — colored labels: `PRIMARY`, `ACTIVE`, `UNSAVED CHANGES`, `clean`
- **Quick actions** — add, remove, open terminal/editor, fetch, prune
- **Detail view** — full commit info, ahead/behind counts, file change breakdown
- **Folder picker** — browse for a path using your OS file manager (`Ctrl+O`)

### Branch Manager

- **Browse all branches** — scrollable list with tracking status and sync info
- **Checkout** — switch branches instantly
- **Create / Rename / Delete** — full branch lifecycle management
- **Merge** — merge any branch into current
- **Tags** — `ACTIVE`, `remote`, `local only` with ahead/behind indicators

### General

- **Tabbed interface** — switch between Worktrees and Branches with `←` `→`
- **Search/filter** — instant filter in both tabs with `/`
- **Loading states** — visual feedback while fetching data
- **Cross-platform** — macOS, Linux, and Windows
- **Version flag** — `git-wtm --version`

---

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `←` / `→` | Switch between Worktrees and Branches tabs |
| `/` | Search / filter |
| `?` | Toggle full help |
| `q` | Quit |
| `Ctrl+C` | Force quit |

### Worktrees Tab

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | View worktree details |
| `a` | Add worktree |
| `d` / `x` | Remove worktree |
| `o` | Open terminal in worktree |
| `e` | Open editor in worktree |
| `p` | Prune stale worktrees |
| `f` | Fetch all remotes |
| `r` | Refresh list |

### Branches Tab

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | View branch details |
| `c` | Checkout branch |
| `a` / `n` | Create new branch |
| `R` | Rename branch |
| `d` / `x` | Delete branch |
| `m` | Merge branch into current |
| `f` | Fetch all remotes |
| `r` | Refresh list |

### Add Worktree View

| Key | Action |
|-----|--------|
| Type | Search/filter matching branches |
| `↑` / `↓` | Navigate suggestions |
| `Enter` | Create worktree (new branch if no match) |
| `Tab` | Switch to path field |
| `Ctrl+O` | Open folder picker (OS native) |
| `Esc` | Cancel |

---

## How It Works

### Adding a Worktree

Press `a` to open the add view:

1. **Type a branch name** — matching local branches appear as a dropdown
2. **Pick a suggestion** with `↑`/`↓` and `Enter`, or just hit `Enter` to create a new branch
3. **Path is optional** — press `Tab` to customize, or `Ctrl+O` to browse. Leave empty for auto-generated path

The default path is `../<repo-name>-<branch-name>` (sibling directory to your repo).

### Branch Management

Switch to the Branches tab with `→`. From there you can checkout, create, rename, delete, or merge branches — all without leaving the TUI.

---

## Requirements

- Git 2.15+
- A terminal with 256-color support
- Go 1.21+ only if building from source
- Optional: `zenity` on Linux for the native folder picker

---

## Configuration

`git-wtm` respects these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `EDITOR` | `code` | Editor opened with `e` key |
| `SHELL` | `/bin/sh` | Shell opened with `o` key |

---

## Releasing

Releases are automated with [GoReleaser](https://goreleaser.com/). On every version tag, GitHub Actions builds binaries for all platforms and publishes to Homebrew, Scoop, Chocolatey, and Winget automatically.

```bash
git tag v1.0.0
git push origin v1.0.0
```

---

## Contributing

```bash
git clone https://github.com/aryanpnd/git-wtm.git
cd git-wtm
go run .
```

Pull requests are welcome.

---

## License

[MIT](LICENSE)
