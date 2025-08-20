package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"git-shippr/internal/gh"

	list "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	stageFetch = iota
	stagePickPR
	stageConfirmOpen
	stagePickStrategy
	stageConfirmDelete
	stageMerging
	stageDone
)

const (
	mergeSquash = "--squash"
	mergeRebase = "--rebase"
	mergeMerge  = "--merge"
)

type prItem struct{ gh.PR }

func (i prItem) Title() string       { return fmt.Sprintf("#%d %s", i.Number, i.PR.Title) }
func (i prItem) Description() string { return i.HeadRefName }
func (i prItem) FilterValue() string { return fmt.Sprintf("%d %s %s", i.Number, i.PR.Title, i.HeadRefName) }

type strategyItem struct{ flag, label string }

func (s strategyItem) Title() string       { return s.label }
func (s strategyItem) Description() string { return s.flag }
func (s strategyItem) FilterValue() string { return s.label }

type model struct {
	ctx      context.Context
	repo     string
	prs      []gh.PR
	list     list.Model
	stage    int
	selected *gh.PR
	strat    string
	deleteBr bool
	status   string
	err      error
}

type fetchedMsg struct{ prs []gh.PR; err error }

type mergedMsg struct{ err error }

type openInBrowserMsg struct{}

type mergeCmdMsg struct{}

func initialModel(ctx context.Context, repo string) model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Open Pull Requests"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	return model{
		ctx:   ctx,
		repo:  repo,
		list:  l,
		stage: stageFetch,
		strat: mergeSquash,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.fetchPRs(), tea.EnterAltScreen)
}

func (m model) fetchPRs() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 15*time.Second)
		defer cancel()
		if err := gh.EnsureGH(ctx); err != nil {
			return fetchedMsg{err: err}
		}
		prs, err := gh.ListPRs(ctx, m.repo)
		return fetchedMsg{prs: prs, err: err}
	}
}

func (m model) mergeSelected() tea.Cmd {
	return func() tea.Msg {
		if m.selected == nil {
			return mergedMsg{err: fmt.Errorf("no PR selected")}
		}
		ctx, cancel := context.WithTimeout(m.ctx, 60*time.Second)
		defer cancel()
		err := gh.MergePR(ctx, m.repo, m.selected.Number, m.strat, m.deleteBr)
		return mergedMsg{err: err}
	}
}

func (m model) openSelectedInBrowser() tea.Cmd {
	return func() tea.Msg {
		if m.selected != nil {
			_ = gh.ViewPRWeb(m.ctx, m.repo, m.selected.Number)
		}
		return openInBrowserMsg{}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-2)
		return m, nil

	case fetchedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.status = fmt.Sprintf("Error: %v", msg.err)
			m.stage = stageDone
			return m, nil
		}
		m.prs = msg.prs
		if len(m.prs) == 0 {
			m.status = fmt.Sprintf("No open PRs for %s", m.repo)
			m.stage = stageDone
			return m, nil
		}
		items := make([]list.Item, 0, len(m.prs))
		for _, p := range m.prs {
			items = append(items, prItem{PR: p})
		}
		m.list.SetItems(items)
		m.stage = stagePickPR
		return m, nil

	case tea.KeyMsg:
		switch m.stage {
		case stagePickPR:
			switch msg.String() {
			case "enter":
				if it, ok := m.list.SelectedItem().(prItem); ok {
					p := it.PR
					m.selected = &p
					m.stage = stageConfirmOpen
				}
				return m, nil
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
		case stageConfirmOpen:
			switch msg.String() {
			case "y", "Y":
				m.status = "Opening PR in browser..."
				return m, m.openSelectedInBrowser()
			case "n", "N", "enter":
				m.stage = stagePickStrategy
				return m, m.showStrategies()
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
		case stagePickStrategy:
			switch msg.String() {
			case "enter":
				if it, ok := m.list.SelectedItem().(strategyItem); ok {
					m.strat = it.flag
					m.stage = stageConfirmDelete
				}
				return m, nil
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
		case stageConfirmDelete:
			switch msg.String() {
			case "y", "Y":
				m.deleteBr = true
				m.stage = stageMerging
				m.status = "Merging PR..."
				return m, m.mergeSelected()
			case "n", "N":
				m.deleteBr = false
				m.stage = stageMerging
				m.status = "Merging PR..."
				return m, m.mergeSelected()
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
		}
		return m, nil

	case openInBrowserMsg:
		m.stage = stagePickStrategy
		return m, m.showStrategies()

	case mergedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.status = fmt.Sprintf("Merge failed: %v", msg.err)
		} else {
			m.status = "Merge succeeded."
		}
		m.stage = stageDone
		return m, nil
	}

	if m.stage == stagePickPR || m.stage == stagePickStrategy {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) showStrategies() tea.Cmd {
	items := []list.Item{
		strategyItem{flag: mergeSquash, label: "Squash (default)"},
		strategyItem{flag: mergeRebase, label: "Rebase"},
		strategyItem{flag: mergeMerge, label: "Merge"},
	}
	m.list.Title = "Choose merge strategy"
	m.list.SetItems(items)
	m.list.Select(0)
	return nil
}

func (m model) View() string {
	switch m.stage {
	case stageFetch:
		return "Fetching pull requests...\n"
	case stagePickPR:
		return m.list.View()
	case stageConfirmOpen:
		return fmt.Sprintf("Open PR #%d in browser before merging? (y/N)\n", m.selected.Number)
	case stagePickStrategy:
		return m.list.View()
	case stageConfirmDelete:
		return fmt.Sprintf("Delete branch '%s' after merging? (y/N)\n", m.selected.HeadRefName)
	case stageMerging:
		return m.status + "\n"
	case stageDone:
		return m.status + "\n(press q/esc/ctrl+c to quit)\n"
	default:
		return ""
	}
}

func runList(org string) error {
	rows, err := gh.ListOpenPRsForOrg(context.Background(), org, 0)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		fmt.Printf("No open PRs for %s\n", org)
		return nil
	}
	for _, r := range rows {
		fmt.Printf("%s #%d %s (%s)\n", r.Repo, r.PR.Number, r.PR.Title, r.PR.HeadRefName)
	}
	return nil
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "list" {
		fs := flag.NewFlagSet("list", flag.ExitOnError)
		var org string
		fs.StringVar(&org, "org", "", "GitHub organization")
		_ = fs.Parse(os.Args[2:])
		if org == "" {
			fmt.Println("Usage: shippr list --org <org>")
			os.Exit(1)
		}
		if err := runList(org); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	var org, repo string
	var noAlt bool
	flag.StringVar(&org, "org", "", "GitHub organization or user")
	flag.StringVar(&repo, "repo", "", "Repository name")
	flag.BoolVar(&noAlt, "no-alt", false, "Disable alternate screen (render in normal screen)")
	flag.Parse()

	repoSlug := ""
	if org != "" && repo != "" {
		repoSlug = gh.Slug(org, repo)
	} else if flag.NArg() == 1 {
		repoSlug = flag.Arg(0)
	} else {
		fmt.Println("Usage: shippr list --org <org> | shippr --org <org> --repo <repo> | shippr <org/repo>")
		os.Exit(1)
	}

	if len(os.Getenv("DEBUG")) > 0 {
		if f, err := tea.LogToFile("debug.log", "debug"); err == nil {
			defer f.Close()
		}
	}

	if noAlt {
		p := tea.NewProgram(initialModel(context.Background(), repoSlug))
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	p := tea.NewProgram(initialModel(context.Background(), repoSlug), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
