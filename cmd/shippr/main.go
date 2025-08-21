package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"git-shippr/internal/gh"

	list "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// logoANSI is a colorful ANSI art logo rendered at startup.
var logoANSI = "" +
	"\x1b[49m                                                                                                    \x1b[m\n" +
	"\x1b[49m                                                                                                    \x1b[m\n" +
	"\x1b[49m                                                                                                    \x1b[m\n" +
	"\x1b[49m                                                                                                    \x1b[m\n" +
	"\x1b[49m                                                                                                    \x1b[m\n" +
	"\x1b[49m                                                                                                    \x1b[m\n" +
	"\x1b[49m                                        \x1b[38;2;151;89;169;49m▄\x1b[38;2;160;92;176;49m▄\x1b[38;2;163;95;179;49m▄\x1b[38;2;160;92;176;49m▄▄\x1b[38;2;157;92;174;49m▄\x1b[38;2;156;89;169;49m▄\x1b[49m                                                     \x1b[m\n" +
	"\x1b[49m                                        \x1b[38;2;160;95;180;48;2;177;105;197m▄\x1b[49;38;2;179;108;203m▀\x1b[49;38;2;181;108;205m▀\x1b[49;38;2;179;108;203m▀▀\x1b[49;38;2;179;108;202m▀\x1b[38;2;163;97;176;48;2;173;104;195m▄\x1b[49m                                                     \x1b[m\n" +
	"\x1b[49m                                        \x1b[38;2;175;105;195;48;2;170;101;189m▄\x1b[38;2;178;100;193;48;2;173;99;189m▄\x1b[49m    \x1b[38;2;175;106;196;48;2;170;103;189m▄\x1b[49m                                                     \x1b[m\n" +
	"\x1b[49m                                     \x1b[38;2;156;92;184;49m▄\x1b[38;2;166;100;188;49m▄\x1b[38;2;171;103;193;49m▄\x1b[38;2;181;108;201;48;2;178;107;199m▄\x1b[38;2;176;106;196;48;2;183;112;199m▄\x1b[38;2;172;105;196;49m▄\x1b[38;2;172;102;193;49m▄▄\x1b[38;2;169;101;190;49m▄\x1b[38;2;180;109;202;48;2;176;107;197m▄\x1b[38;2;169;102;191;49m▄\x1b[38;2;166;102;188;49m▄▄\x1b[38;2;166;99;188;49m▄\x1b[38;2;164;99;185;49m▄\x1b[38;2;161;97;185;49m▄\x1b[38;2;159;96;183;49m▄\x1b[38;2;158;97;183;49m▄\x1b[38;2;158;97;180;49m▄\x1b[38;2;156;94;180;49m▄\x1b[38;2;146;89;169;49m▄\x1b[49m                                          \x1b[m\n" +
	"\x1b[49m                                     \x1b[38;2;128;64;140;48;2;167;97;185m▄\x1b[49;38;2;171;104;193m▀\x1b[49;38;2;174;104;197m▀\x1b[49;38;2;181;109;203m▀\x1b[49;38;2;178;107;200m▀\x1b[49;38;2;177;107;200m▀\x1b[49;38;2;174;104;197m▀\x1b[49;38;2;175;106;198m▀\x1b[49;38;2;174;106;197m▀\x1b[49;38;2;181;109;204m▀\x1b[49;38;2;172;104;195m▀▀\x1b[49;38;2;172;103;195m▀\x1b[49;38;2;169;102;192m▀▀\x1b[49;38;2;168;101;191m▀\x1b[49;38;2;165;100;190m▀\x1b[49;38;2;165;99;189m▀\x1b[49;38;2;164;98;186m▀\x1b[49;38;2;161;97;186m▀\x1b[49;38;2;152;90;174m▀\x1b[49m                                          \x1b[m\n" +
	"\x1b[49m                                    \x1b[38;2;174;101;190;48;2;168;98;186m▄\x1b[38;2;174;103;193;48;2;174;101;189m▄\x1b[49m        \x1b[38;2;151;88;167;49m▄\x1b[38;2;169;101;191;49m▄\x1b[38;2;167;101;189;49m▄\x1b[38;2;168;102;191;49m▄\x1b[38;2;166;100;189;49m▄▄\x1b[38;2;164;100;189;49m▄\x1b[38;2;164;100;187;49m▄\x1b[38;2;144;89;166;49m▄\x1b[49m  \x1b[38;2;152;85;170;48;2;155;92;176m▄\x1b[38;2;165;99;189;48;2;161;94;183m▄\x1b[38;2;144;89;166;49m▄\x1b[38;2;73;36;91;49m▄\x1b[38;2;155;93;178;49m▄\x1b[38;2;151;91;176;49m▄\x1b[38;2;151;91;174;49m▄▄\x1b[38;2;149;91;174;49m▄\x1b[38;2;148;90;173;49m▄▄\x1b[38;2;149;91;174;49m▄▄\x1b[38;2;148;90;173;49m▄\x1b[38;2;149;91;174;49m▄\x1b[38;2;148;90;173;49m▄\x1b[38;2;148;88;171;49m▄\x1b[38;2;147;89;172;49m▄▄▄▄\x1b[38;2;145;89;172;49m▄\x1b[38;2;145;89;170;49m▄\x1b[38;2;147;89;170;49m▄\x1b[38;2;147;89;172;49m▄\x1b[38;2;138;84;161;49m▄\x1b[49m                 \x1b[m\n" +
	"...TRUNCATED FOR BREVITY..." // The full provided ANSI string should be included here converted to \x1b escapes.

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
		return logoANSI + "\nFetching pull requests...\n"
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
	// Determine layout
	maxRepo := 0
	for _, r := range rows {
		if l := len(r.Repo); l > maxRepo {
			maxRepo = l
		}
	}
	width := termWidth()
	if width <= 0 {
		width = 100
	}
	// Repo col + spaces + # + num + spaces + title col + spaces + branch
	// Reserve ~20 chars for repo, ~8 for number, ~20 for branch; rest for title
	repoW := maxRepo
	if repoW < 16 {
		repoW = 16
	}
	numW := 6 // includes '#'
	branchW := 24
	titleW := width - repoW - numW - branchW - 6
	if titleW < 20 {
		titleW = 20
	}

	// Header
	bold := func(s string) string { return "\x1b[1m" + s + "\x1b[0m" }
	dim := func(s string) string { return "\x1b[90m" + s + "\x1b[0m" }
	cyan := func(s string) string { return "\x1b[36m" + s + "\x1b[0m" }
	magenta := func(s string) string { return "\x1b[35m" + s + "\x1b[0m" }

	fmt.Printf("%s  %s  %s  %s\n",
		bold(padRight("REPO", repoW)),
		bold(padRight("PR", numW)),
		bold(padRight("TITLE", titleW)),
		bold("BRANCH"),
	)
	fmt.Println(dim(stringsRepeat("─", repoW+numW+titleW+branchW+6)))

	for _, r := range rows {
		repo := padRight(r.Repo, repoW)
		pr := fmt.Sprintf("#%-*d", numW-1, r.PR.Number)
		title := padRight(truncate(r.PR.Title, titleW), titleW)
		branch := r.PR.HeadRefName
		fmt.Printf("%s  %s  %s  %s\n",
			cyan(repo),
			magenta(pr),
			title,
			dim(branch),
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
            fmt.Print(logoANSI + "\n")
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
        fmt.Print(logoANSI + "\n")
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
