package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"git-shippr/internal/gh"

	list "github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	stageFetch = iota
	stagePickPR
	stageViewSummary
	stageConfirmOpen
	stagePickStrategy
	stageConfirmDelete
	stageMerging
	stageDone
)

var (
	primary   = lipgloss.Color("#C471ED")
	secondary = lipgloss.Color("#DA70D6")
	accent    = lipgloss.Color("#DDA0DD")
	dark      = lipgloss.Color("#8B008B")
	light     = lipgloss.Color("#E6E6FA")

	titleStyle = lipgloss.NewStyle().
			Foreground(primary).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(light)

	successStyle = lipgloss.NewStyle().
			Foreground(secondary).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF1493")).
			Bold(true)

	branchStyle = lipgloss.NewStyle().
			Foreground(accent)

	accentStyle = lipgloss.NewStyle().
			Foreground(accent)

	prNumberStyle = lipgloss.NewStyle().
			Foreground(primary).
			Bold(true)

	highlightStyle = lipgloss.NewStyle().
			Foreground(dark).
			Background(light).
			Bold(true).
			Padding(0, 1)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primary).
			Padding(1)
)

const (
	mergeSquash = "--squash"
	mergeRebase = "--rebase"
	mergeMerge  = "--merge"
)

type prItem struct{ gh.PR }

func (i prItem) Title() string {
	return fmt.Sprintf("%s %s",
		prNumberStyle.Render(fmt.Sprintf("#%d", i.Number)),
		i.PR.Title)
}

func (i prItem) Description() string {
	return fmt.Sprintf("Branch: %s", branchStyle.Render(i.HeadRefName))
}

func (i prItem) FilterValue() string {
	return fmt.Sprintf("%d %s %s", i.Number, i.PR.Title, i.HeadRefName)
}

type strategyItem struct{ flag, label string }

func (s strategyItem) Title() string       { return s.label }
func (s strategyItem) Description() string { return s.flag }
func (s strategyItem) FilterValue() string { return s.label }

type model struct {
	ctx       context.Context
	repo      string
	prs       []gh.PR
	list      list.Model
	spinner   spinner.Model
	stage     int
	selected  *gh.PR
	prDetails *gh.PRDetails
	strat     string
	deleteBr  bool
	status    string
	err       error
}

type fetchedMsg struct {
	prs []gh.PR
	err error
}

type mergedMsg struct{ err error }

type openInBrowserMsg struct{}

type prDetailsMsg struct {
	details *gh.PRDetails
	err     error
}

type reviewActionMsg struct{ err error }

func initialModel(ctx context.Context, repo string) model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Open Pull Requests"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(primary)

	return model{
		ctx:     ctx,
		repo:    repo,
		list:    l,
		spinner: s,
		stage:   stageFetch,
		strat:   mergeSquash,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.fetchPRs(), m.spinner.Tick, tea.EnterAltScreen)
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

func (m model) fetchPRDetails() tea.Cmd {
	return func() tea.Msg {
		if m.selected == nil {
			return prDetailsMsg{err: fmt.Errorf("no PR selected")}
		}
		ctx, cancel := context.WithTimeout(m.ctx, 15*time.Second)
		defer cancel()
		details, err := gh.GetPRDetails(ctx, m.repo, m.selected.Number)
		return prDetailsMsg{details: details, err: err}
	}
}

func (m model) approvePR() tea.Cmd {
	return func() tea.Msg {
		if m.selected == nil {
			return reviewActionMsg{err: fmt.Errorf("no PR selected")}
		}
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()
		err := gh.ApprovePR(ctx, m.repo, m.selected.Number)
		return reviewActionMsg{err: err}
	}
}

func (m model) requestChanges() tea.Cmd {
	return func() tea.Msg {
		if m.selected == nil {
			return reviewActionMsg{err: fmt.Errorf("no PR selected")}
		}
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()
		err := gh.RequestChanges(ctx, m.repo, m.selected.Number, "Changes requested via shippr")
		return reviewActionMsg{err: err}
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

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case fetchedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.status = fmt.Sprintf("Failed to fetch PRs: %v", msg.err)
			m.stage = stageDone
			return m, nil
		}
		m.prs = msg.prs
		if len(m.prs) == 0 {
			m.status = fmt.Sprintf("No open pull requests found for %s", titleStyle.Render(m.repo))
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

	case prDetailsMsg:
		if msg.err != nil {
			m.err = msg.err
			m.status = fmt.Sprintf("Failed to fetch PR details: %v", msg.err)
			m.stage = stageDone
			return m, nil
		}
		m.prDetails = msg.details
		m.stage = stageViewSummary
		return m, nil

	case reviewActionMsg:
		if msg.err != nil {
			m.err = msg.err
			m.status = fmt.Sprintf("Review action failed: %v", msg.err)
			m.stage = stageDone
			return m, nil
		}
		m.status = "Review submitted successfully!"
		m.stage = stageConfirmOpen
		return m, nil

	case tea.KeyMsg:
		switch m.stage {
		case stagePickPR:
			switch msg.String() {
			case "enter":
				if it, ok := m.list.SelectedItem().(prItem); ok {
					p := it.PR
					m.selected = &p
					m.status = "Fetching PR details..."
					return m, m.fetchPRDetails()
				}
				return m, nil
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
		case stageViewSummary:
			switch msg.String() {
			case "a":
				m.status = "Approving PR..."
				return m, m.approvePR()
			case "r":
				m.status = "Requesting changes..."
				return m, m.requestChanges()
			case "m", "enter":
				m.stage = stageConfirmOpen
				return m, nil
			case "b", "esc":
				m.stage = stagePickPR
				return m, nil
			case "q", "ctrl+c":
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
			m.status = fmt.Sprintf("Failed to merge PR #%d: %v", m.selected.Number, msg.err)
		} else {
			m.status = fmt.Sprintf("Successfully merged PR #%d using %s strategy",
				m.selected.Number, strings.ToUpper(m.strat[2:]))
			if m.deleteBr {
				m.status += " and deleted branch"
			}
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

func (m model) renderPRSummary() string {
	if m.prDetails == nil {
		return m.spinner.View() + " " + infoStyle.Render("Loading PR details...")
	}

	pr := m.prDetails
	var content strings.Builder

	// Header
	content.WriteString(borderStyle.Render(
		titleStyle.Render(fmt.Sprintf("PR #%d: %s", pr.Number, pr.Title)) + "\n" +
			infoStyle.Render(fmt.Sprintf("by %s • %s → %s", pr.Author.Login, pr.HeadRefName, pr.BaseRefName)),
	))
	content.WriteString("\n\n")

	// Status and stats
	statusInfo := fmt.Sprintf("%s  %s  %s",
		highlightStyle.Render(pr.State),
		successStyle.Render(fmt.Sprintf("+%d", pr.Additions)),
		errorStyle.Render(fmt.Sprintf("-%d", pr.Deletions)),
	)
	content.WriteString(statusInfo + "\n\n")

	// Description
	if pr.Body != "" {
		desc := pr.Body
		if len(desc) > 200 {
			desc = desc[:200] + "..."
		}
		content.WriteString(titleStyle.Render("Description:") + "\n")
		content.WriteString(infoStyle.Render(desc) + "\n\n")
	}

	// File changes
	content.WriteString(titleStyle.Render(fmt.Sprintf("Files Changed (%d):", pr.ChangedFiles)) + "\n")
	for i, file := range pr.Files {
		if i >= 5 {
			content.WriteString(infoStyle.Render(fmt.Sprintf("... and %d more files", len(pr.Files)-5)) + "\n")
			break
		}
		var statusIcon string
		switch file.Status {
		case "added":
			statusIcon = successStyle.Render("A")
		case "modified":
			statusIcon = accentStyle.Render("M")
		case "removed":
			statusIcon = errorStyle.Render("D")
		default:
			statusIcon = infoStyle.Render("?")
		}
		content.WriteString(fmt.Sprintf("  %s %s %s\n",
			statusIcon,
			file.Path,
			infoStyle.Render(fmt.Sprintf("(+%d -%d)", file.Additions, file.Deletions)),
		))
	}
	content.WriteString("\n")

	// Reviews
	if len(pr.Reviews) > 0 {
		content.WriteString(titleStyle.Render("Reviews:") + "\n")
		for _, review := range pr.Reviews {
			var statusIcon string
			switch review.State {
			case "APPROVED":
				statusIcon = successStyle.Render("✓")
			case "CHANGES_REQUESTED":
				statusIcon = errorStyle.Render("✗")
			case "COMMENTED":
				statusIcon = infoStyle.Render("💬")
			default:
				statusIcon = infoStyle.Render("?")
			}
			content.WriteString(fmt.Sprintf("  %s %s\n", statusIcon, review.Author.Login))
		}
		content.WriteString("\n")
	}

	// Actions
	content.WriteString(borderStyle.Render(
		titleStyle.Render("Actions:") + "\n" +
			highlightStyle.Render("a") + " Approve  " +
			highlightStyle.Render("r") + " Request Changes  " +
			highlightStyle.Render("m") + " Merge\n" +
			highlightStyle.Render("b") + " Back  " +
			highlightStyle.Render("q") + " Quit",
	))

	return content.String()
}

func (m model) View() string {
	var content string

	switch m.stage {
	case stageFetch:
		content = fmt.Sprintf("%s\n%s %s\n",
			getLogo(),
			m.spinner.View(),
			infoStyle.Render("Fetching pull requests..."))
	case stagePickPR:
		content = m.list.View()
	case stageViewSummary:
		content = m.renderPRSummary()
	case stageConfirmOpen:
		content = fmt.Sprintf("%s\n%s\n%s",
			titleStyle.Render("PR Preview"),
			fmt.Sprintf("PR %s: %s", prNumberStyle.Render(fmt.Sprintf("#%d", m.selected.Number)), m.selected.Title),
			fmt.Sprintf("\nBranch: %s\n\n%s",
				branchStyle.Render(m.selected.HeadRefName),
				"Open PR in browser before merging? (y/N)"))
	case stagePickStrategy:
		content = m.list.View()
	case stageConfirmDelete:
		content = fmt.Sprintf("%s\n%s\n%s",
			titleStyle.Render("Merge Confirmation"),
			fmt.Sprintf("Ready to merge PR %s with %s strategy",
				prNumberStyle.Render(fmt.Sprintf("#%d", m.selected.Number)),
				infoStyle.Render(strings.ToUpper(m.strat[2:]))),
			fmt.Sprintf("\nDelete branch '%s' after merging? (y/N)\n", branchStyle.Render(m.selected.HeadRefName)))
	case stageMerging:
		content = fmt.Sprintf("%s %s\n", m.spinner.View(), infoStyle.Render(m.status))
	case stageDone:
		if m.err != nil {
			content = fmt.Sprintf("%s\n%s\n\n%s",
				errorStyle.Render("❌ Error"),
				errorStyle.Render(m.status),
				infoStyle.Render("(press q/esc/ctrl+c to quit)"))
		} else {
			content = fmt.Sprintf("%s\n%s\n\n%s",
				successStyle.Render("✅ Success"),
				successStyle.Render(m.status),
				infoStyle.Render("(press q/esc/ctrl+c to quit)"))
		}
	default:
		content = ""
	}

	return content
}

func formatLabels(labels []struct {
	Name string `json:"name"`
}) string {
	if len(labels) == 0 {
		return "-"
	}
	var labelNames []string
	for _, label := range labels {
		labelNames = append(labelNames, label.Name)
	}
	return strings.Join(labelNames, ", ")
}

func formatStatus(pr gh.PR) string {
	if pr.State != "OPEN" {
		return pr.State
	}
	if pr.Mergeable == "CONFLICTING" {
		return "CONFLICT"
	}
	return "OPEN"
}

func runList(org string) error {
	rows, err := gh.ListOpenPRsForOrg(context.Background(), org, 0)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		fmt.Printf("%s\n", infoStyle.Render(fmt.Sprintf("No open PRs for %s", titleStyle.Render(org))))
		return nil
	}

	// Determine layout with additional columns
	maxRepo := 0
	maxAuthor := 0
	maxLabels := 0
	for _, r := range rows {
		if l := len(r.Repo); l > maxRepo {
			maxRepo = l
		}
		if l := len(r.PR.Author.Login); l > maxAuthor {
			maxAuthor = l
		}
		labelsStr := formatLabels(r.PR.Labels)
		if l := len(labelsStr); l > maxLabels {
			maxLabels = l
		}
	}

	width := termWidth()
	if width <= 0 {
		width = 120
	}

	// Columns: Repo, PR, Author, Title, Branch, Labels, Status
	repoW := maxRepo
	if repoW < 16 {
		repoW = 16
	}
	numW := 6
	authorW := maxAuthor
	if authorW < 12 {
		authorW = 12
	}
	branchW := 24
	labelsW := maxLabels
	if labelsW < 15 {
		labelsW = 15
	}
	statusW := 10
	titleW := width - repoW - numW - authorW - branchW - labelsW - statusW - 8
	if titleW < 25 {
		titleW = 25
	}

	// Header
	fmt.Printf("%s  %s  %s  %s  %s  %s  %s\n",
		titleStyle.Render(padRight("REPO", repoW)),
		titleStyle.Render(padRight("PR", numW)),
		titleStyle.Render(padRight("AUTHOR", authorW)),
		titleStyle.Render(padRight("TITLE", titleW)),
		titleStyle.Render(padRight("BRANCH", branchW)),
		titleStyle.Render(padRight("LABELS", labelsW)),
		titleStyle.Render(padRight("STATUS", statusW)),
	)
	fmt.Println(infoStyle.Render(stringsRepeat("─", repoW+numW+authorW+titleW+branchW+labelsW+statusW+8)))

	for _, r := range rows {
		repo := padRight(r.Repo, repoW)
		pr := fmt.Sprintf("#%-*d", numW-1, r.PR.Number)
		author := padRight(r.PR.Author.Login, authorW)
		title := padRight(truncate(r.PR.Title, titleW), titleW)
		branch := padRight(r.PR.HeadRefName, branchW)
		labels := padRight(formatLabels(r.PR.Labels), labelsW)
		status := formatStatus(r.PR)

		// Color coding
		var statusColored string
		switch status {
		case "OPEN":
			statusColored = successStyle.Render(padRight(status, statusW))
		case "CONFLICT":
			statusColored = errorStyle.Render(padRight(status, statusW))
		default:
			statusColored = infoStyle.Render(padRight(status, statusW))
		}

		fmt.Printf("%s  %s  %s  %s  %s  %s  %s\n",
			infoStyle.Render(repo),
			prNumberStyle.Render(pr),
			branchStyle.Render(author),
			title,
			branchStyle.Render(branch),
			labels,
			statusColored,
		)
	}
	return nil
}

// termWidth returns terminal width using $COLUMNS if available.
func termWidth() int {
	if v := os.Getenv("COLUMNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 0
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	b := make([]byte, 0, w)
	b = append(b, s...)
	for len(b) < w {
		b = append(b, ' ')
	}
	return string(b)
}

func truncate(s string, w int) string {
	if len(s) <= w {
		return s
	}
	if w <= 1 {
		return s[:w]
	}
	// leave space for ellipsis
	cut := w - 1
	if cut > len(s) {
		cut = len(s)
	}
	return s[:cut] + "…"
}

func stringsRepeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	b := make([]byte, 0, len(s)*count)
	for i := 0; i < count; i++ {
		b = append(b, s...)
	}
	return string(b)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "list" {
		fs := flag.NewFlagSet("list", flag.ExitOnError)
		var org string
		fs.StringVar(&org, "org", "", "GitHub organization")
		fs.Usage = func() {
			// Show logo + usage for list subcommand
			fmt.Print(getLogo() + "\n")
			fmt.Fprintln(os.Stderr, "Usage: shippr list --org <org>")
			fmt.Fprintln(os.Stderr)
			fs.PrintDefaults()
		}
		_ = fs.Parse(os.Args[2:])
		if org == "" {
			fs.Usage()
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
	// Global usage with logo
	flag.Usage = func() {
		fmt.Print(getLogo() + "\n")
		fmt.Fprintln(os.Stderr, "Usage: shippr list --org <org> | shippr --org <org> --repo <repo> | shippr <org/repo>")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
	}
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
		flag.Usage()
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
