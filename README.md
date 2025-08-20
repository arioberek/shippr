# git-shippr

Interactive Bubble Tea CLI to quickly find and merge GitHub PRs.

- Lists open PRs for a repo with search/filter.
- Optional: open selected PR in the browser.
- Choose merge strategy: squash (default), rebase, merge.
- Optional: delete branch after merge.
- Uses `gh` (GitHub CLI) under the hood.

## Requirements

- Go 1.25+
- GitHub CLI (`gh`) installed and authenticated: `gh auth login`

## Installation

Option A: via npm (scoped package)

```bash
npm i -g @arioberek/shippr
# or
npx @arioberek/shippr --help
```

This uses a postinstall script that builds the Go binary locally. Go must be installed and on PATH.

Option B: build from source

## Build

```bash
# after renaming folder below, or build with -o
go build -o git-shippr ./cmd/git-shippr
```
This produces a `git-shippr` binary.

## Usage

```bash
# using flags
./git-shippr --org <org> --repo <repo>

# list open PRs across an organization
./git-shippr list --org <org>

# installed via npm (global)
shippr --org <org> --repo <repo>
shippr list --org <org>

# if your terminal clears the alt screen, disable it
shippr --no-alt --org <org> --repo <repo>

# or a single slug argument
./git-shippr <org/repo>
```

Keyboard:
- Enter: select
- q / esc / Ctrl+C: quit
- Type to filter lists

## Notes

- The tool shells out to `gh` for listing, viewing, and merging PRs.
- Deleting the branch uses `--delete-branch` flag from `gh pr merge`.

## Project structure

- `cmd/git-shippr/main.go`: Bubble Tea TUI and flow control.
- `internal/gh/gh.go`: small wrapper around `gh` commands.
