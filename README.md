# 🚢 shippr

> Interactive CLI to quickly find and merge GitHub PRs with style

## Logo

```ansi
\x1b[36m
    ┌─────────────────────────────────┐
    │  ⚓  Welcome aboard shippr! ⚓  │
    └─────────────────────────────────┘
         🚢 ∼∼∼∼∼∼∼∼∼∼∼∼∼∼∼ 🌊
    
     _____ _     _                   
    / ____| |   (_)                  
   | (___ | |__  _ _ __  _ __  _ __   
    \___ \| '_ \| | '_ \| '_ \| '__|  
    ____) | | | | | |_) | |_) | |     
   |_____/|_| |_|_| .__/| .__/|_|     
                   | |   | |          
                   |_|   |_|          
\x1b[0m
```

## ✨ Features

- **Interactive PR Browser**: Beautiful TUI to list and filter open PRs
- **Quick Actions**: View, merge, and manage PRs without leaving your terminal  
- **Flexible Merge Options**: Choose between squash (default), rebase, or merge strategies
- **Branch Cleanup**: Optionally delete branches after merging
- **Organization Support**: List PRs across entire organizations
- **Fast & Lightweight**: Built with Go and powered by GitHub CLI

## 📋 Requirements

- **Go 1.21+** (for building from source)
- **GitHub CLI (`gh`)** installed and authenticated
  ```bash
  gh auth login
  ```

## 🚀 Installation

### Option A: via npm (Recommended)

```bash
# Install globally
npm install -g @arioberek/shippr

# Or run directly
npx @arioberek/shippr --help
```

> **Note**: This method uses a postinstall script to build the Go binary locally. Go must be installed and available in your PATH.

### Option B: Build from Source

```bash
git clone <repository-url>
cd shippr
go build -o shippr ./cmd/git-shippr
```

This produces a `shippr` binary in your current directory.

## 🎯 Usage

### Basic Commands

```bash
# Interactive PR browser for a specific repository
shippr --org <org> --repo <repo>

# Or use the shorthand slug format
shippr <org/repo>

# List all open PRs across an organization
shippr list --org <org>

# Disable alternate screen buffer (if your terminal clears on exit)
shippr --no-alt --org <org> --repo <repo>
```

### Examples

```bash
# Browse PRs for a specific repo
shippr microsoft/vscode

# List all PRs in your organization
shippr list --org mycompany

# Use with flags
shippr --org facebook --repo react
```

## ⌨️ Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Select PR / Confirm action |
| `q` / `Esc` / `Ctrl+C` | Quit application |
| `Type` | Filter/search PRs in real-time |
| `↑` / `↓` | Navigate through PR list |

## 🛠️ How It Works

shippr is a thin wrapper around the GitHub CLI (`gh`) that provides:

1. **PR Listing**: Uses `gh pr list` to fetch open PRs
2. **PR Details**: Uses `gh pr view` for detailed information
3. **Merging**: Uses `gh pr merge` with your chosen strategy
4. **Branch Cleanup**: Uses `--delete-branch` flag when requested

## 📁 Project Structure

shippr/
├── cmd/git-shippr/
│   └── main.go          # Main application entry point & Bubble Tea TUI
├── internal/gh/
│   └── gh.go           # GitHub CLI wrapper functions
├── package.json        # npm package configuration
└── README.md


## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

MIT License - see LICENSE file for details

## 🙋‍♂️ Support

If you encounter any issues or have questions:
- Open an issue on GitHub
- Check existing issues for solutions
- Make sure `gh` is properly authenticated: `gh auth status`

---

Made with ❤️ for developers who love shipping code fast! 🚀
