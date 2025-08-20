# shippr 🚢

Interactive CLI for finding and merging GitHub PRs quickly.

<img width="1024" height="1024" alt="shippr" src="https://github.com/user-attachments/assets/88ae292a-b06f-41ce-9741-de991650ef1d" />

## Features

- Interactive TUI for listing and filtering open PRs
- View, merge, and manage PRs right from your terminal
- Merge options: squash (default), rebase, or merge commit
- Option to delete branches after merging
- Support for listing PRs across an entire organization
- Lightweight Go app that wraps the GitHub CLI

## Requirements

- Go 1.21+ (if building from source)
- GitHub CLI (`gh`) installed and authenticated:
  ```bash
  gh auth login
  ```

## Installation

### Option A: via npm (Recommended)

```bash
# Install globally
npm install -g @arioberek/shippr

# Or run directly
npx @arioberek/shippr --help
```

> **Note**: This builds the Go binary using a postinstall script. You'll need Go installed and in your PATH.

### Option B: Build from Source

```bash
git clone <repository-url>
cd shippr
go build -o shippr ./cmd/git-shippr
```

This creates a `shippr` binary in the current directory.

## Usage

### Basic Commands

```bash
# Browse PRs for a specific repo
shippr --org <org> --repo <repo>

# Shorthand slug format
shippr <org/repo>

# List open PRs across an organization
shippr list --org <org>

# Disable alt screen (if your terminal clears on exit)
shippr --no-alt --org <org> --repo <repo>
```

### Examples

```bash
# Browse PRs in microsoft/vscode
shippr microsoft/vscode

# List PRs in your org
shippr list --org mycompany

# With flags
shippr --org facebook --repo react
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Select or confirm |
| `q` / `Esc` / `Ctrl+C` | Quit |
| Typing | Filter the list |
| `↑` / `↓` | Navigate |

## How It Works

shippr wraps the GitHub CLI (`gh`) to keep things simple:

1. Lists PRs with `gh pr list`
2. Shows details via `gh pr view`
3. Merges using `gh pr merge` and your chosen method
4. Deletes branches with the `--delete-branch` flag if you want

## Project Structure

```text
shippr/
├─ cmd/
│  └─ git-shippr/
│     └─ main.go          # Main entry point with Bubble Tea TUI
├─ internal/
│  └─ gh/
│     └─ gh.go            # GitHub CLI wrappers
├─ package.json           # npm config
└─ README.md
```

## Contributing

1. Fork the repo
2. Create a feature branch (`git checkout -b feature/something-cool`)
3. Commit your changes (`git commit -m 'Add something cool'`)
4. Push it (`git push origin feature/something-cool`)
5. Open a PR

## License

MIT License—check the LICENSE file for details.


## Support

If you encounter any issues or have questions:
- Open an issue on GitHub
- Check existing issues for solutions
- Make sure `gh` is properly authenticated: `gh auth status`

---

Made with ❤️ for developers who love shipping code fast! 🚀
