package gh

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

type PR struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	HeadRefName string `json:"headRefName"`
}

func Slug(org, repo string) string { return fmt.Sprintf("%s/%s", org, repo) }

func ListPRs(ctx context.Context, repo string) ([]PR, error) {
	cmd := exec.CommandContext(ctx, "gh", "pr", "list", "--repo", repo, "--json", "number,title,headRefName")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gh pr list failed: %w\n%s", err, string(out))
	}
	var prs []PR
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, fmt.Errorf("parse gh json: %w", err)
	}
	return prs, nil
}

func ViewPRWeb(ctx context.Context, repo string, number int) error {
	cmd := exec.CommandContext(ctx, "gh", "pr", "view", fmt.Sprint(number), "--repo", repo, "--web")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gh pr view failed: %w\n%s", err, string(out))
	}
	return nil
}

func MergePR(ctx context.Context, repo string, number int, strategy string, deleteBranch bool) error {
	args := []string{"pr", "merge", fmt.Sprint(number), "--repo", repo, strategy}
	if deleteBranch {
		args = append(args, "--delete-branch")
	}
	cmd := exec.CommandContext(ctx, "gh", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gh pr merge failed: %w\n%s", err, string(out))
	}
	return nil
}

func EnsureGH(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "gh", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("'gh' not found, please install GitHub CLI: %w", err)
	}
	return nil
}

func HasFZF(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "sh", "-c", "command -v fzf >/dev/null 2>&1")
	return cmd.Run() == nil
}

func HumanStrategy(flag string) string {
	switch strings.ToLower(flag) {
	case "--rebase":
		return "rebase"
	case "--merge":
		return "merge"
	default:
		return "squash"
	}
}

type repoOwner struct{
	Login string `json:"login"`
}

type Repository struct{
	Name string `json:"name"`
	Owner repoOwner `json:"owner"`
}

func ListOrgRepos(ctx context.Context, org string, limit int) ([]Repository, error) {
	args := []string{"repo", "list", org, "--json", "name,owner"}
	if limit > 0 { args = append(args, "--limit", fmt.Sprint(limit)) }
	cmd := exec.CommandContext(ctx, "gh", args...)
	out, err := cmd.CombinedOutput()
	if err != nil { return nil, fmt.Errorf("gh repo list failed: %w\n%s", err, string(out)) }
	var repos []Repository
	if err := json.Unmarshal(out, &repos); err != nil { return nil, fmt.Errorf("parse gh json: %w", err) }
	return repos, nil
}

type RepoPR struct{
	Repo string
	PR   PR
}

func ListOpenPRsForOrg(ctx context.Context, org string, limitRepos int) ([]RepoPR, error) {
	repos, err := ListOrgRepos(ctx, org, limitRepos)
	if err != nil { return nil, err }
	concurrency := runtime.NumCPU()
	if concurrency < 4 { concurrency = 4 }
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	resCh := make(chan []RepoPR, len(repos))
	for _, r := range repos {
		repoSlug := Slug(r.Owner.Login, r.Name)
		wg.Add(1)
		go func(slug string){
			defer wg.Done()
			sem <- struct{}{}
			defer func(){ <-sem }()
			prs, e := ListPRs(ctx, slug)
			if e != nil { return }
			acc := make([]RepoPR, 0, len(prs))
			for _, p := range prs { acc = append(acc, RepoPR{Repo: slug, PR: p}) }
			resCh <- acc
		}(repoSlug)
	}
	go func(){ wg.Wait(); close(resCh) }()
	var all []RepoPR
	for batch := range resCh { all = append(all, batch...) }
	return all, nil
}
