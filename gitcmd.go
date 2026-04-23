package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Git struct {
	workDir string
}

func NewGit(workDir string) *Git {
	return &Git{workDir: workDir}
}

func (g *Git) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.workDir
	out, err := cmd.CombinedOutput()
	result := strings.TrimRight(string(out), "\r\n")
	if err != nil {
		if result != "" {
			return result, fmt.Errorf("%s", result)
		}
		return result, err
	}
	return result, nil
}

func (g *Git) IsRepo() bool {
	_, err := g.run("rev-parse", "--git-dir")
	return err == nil
}

func (g *Git) CurrentBranch() string {
	out, err := g.run("branch", "--show-current")
	if err != nil {
		return "HEAD"
	}
	return out
}

func (g *Git) ShortHash() string {
	out, _ := g.run("rev-parse", "--short", "HEAD")
	return out
}

type FileStatus struct {
	Path      string
	Staged    bool
	Status    string
	Modified  bool
	Deleted   bool
	Added     bool
	Renamed   bool
	Untracked bool
}

func (g *Git) Status() ([]FileStatus, []FileStatus, []FileStatus, error) {
	out, err := g.run("status", "--porcelain=v1")
	if err != nil {
		return nil, nil, nil, err
	}

	var staged, unstaged, untracked []FileStatus
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}
		index := line[0]
		worktree := line[1]
		path := strings.TrimSpace(line[3:])

		fs := FileStatus{Path: path}

		switch index {
		case 'M':
			fs.Staged = true
			fs.Modified = true
			fs.Status = "已修改"
		case 'A':
			fs.Staged = true
			fs.Added = true
			fs.Status = "已添加"
		case 'D':
			fs.Staged = true
			fs.Deleted = true
			fs.Status = "已删除"
		case 'R':
			fs.Staged = true
			fs.Renamed = true
			fs.Status = "已重命名"
		case 'C':
			fs.Staged = true
			fs.Status = "已复制"
		}

		if fs.Staged {
			staged = append(staged, fs)
		}

		ufs := FileStatus{Path: path}
		switch worktree {
		case 'M':
			ufs.Modified = true
			ufs.Status = "已修改"
			unstaged = append(unstaged, ufs)
		case 'D':
			ufs.Deleted = true
			ufs.Status = "已删除"
			unstaged = append(unstaged, ufs)
		case '?':
			ufs.Untracked = true
			ufs.Status = "未跟踪"
			untracked = append(untracked, ufs)
		case '!':
		default:
			if index == '?' {
				ufs.Untracked = true
				ufs.Status = "未跟踪"
				untracked = append(untracked, ufs)
			}
		}
	}

	if len(untracked) == 0 && len(staged) == 0 && len(unstaged) == 0 {
		for _, line := range lines {
			if len(line) >= 4 && line[0] == '?' && line[1] == '?' {
				path := strings.TrimSpace(line[3:])
				untracked = append(untracked, FileStatus{Path: path, Untracked: true, Status: "未跟踪"})
			}
		}
	}

	return staged, unstaged, untracked, nil
}

func (g *Git) Stage(path string) error {
	_, err := g.run("add", path)
	return err
}

func (g *Git) StageAll() error {
	_, err := g.run("add", "-A")
	return err
}

func (g *Git) Unstage(path string) error {
	_, err := g.run("reset", "HEAD", "--", path)
	return err
}

func (g *Git) UnstageAll() error {
	_, err := g.run("reset", "HEAD")
	return err
}

func (g *Git) Commit(message string) error {
	_, err := g.run("commit", "-m", message)
	return err
}

func (g *Git) AmendCommit(message string) error {
	if message != "" {
		_, err := g.run("commit", "--amend", "-m", message)
		return err
	}
	_, err := g.run("commit", "--amend", "--no-edit")
	return err
}

func (g *Git) HasUncommittedChanges() bool {
	out, _ := g.run("status", "--porcelain")
	return out != ""
}

type CommitInfo struct {
	Hash    string
	Short   string
	Author  string
	Date    string
	Message string
}

func (g *Git) Log(count int) ([]CommitInfo, error) {
	out, err := g.run("log", fmt.Sprintf("-%d", count), "--pretty=format:%H|%h|%an|%ar|%s")
	if err != nil {
		return nil, err
	}
	return g.parseLog(out), nil
}

func (g *Git) LogAll(count int) ([]CommitInfo, error) {
	out, err := g.run("log", "--all", fmt.Sprintf("-%d", count), "--pretty=format:%H|%h|%an|%ar|%s")
	if err != nil {
		return nil, err
	}
	return g.parseLog(out), nil
}

func (g *Git) SearchLog(query string, count int) ([]CommitInfo, error) {
	out, err := g.run("log", fmt.Sprintf("-%d", count), "--grep="+query, "--pretty=format:%H|%h|%an|%ar|%s")
	if err != nil {
		return nil, err
	}
	return g.parseLog(out), nil
}

func (g *Git) parseLog(out string) []CommitInfo {
	var commits []CommitInfo
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}
		commits = append(commits, CommitInfo{
			Hash:    parts[0],
			Short:   parts[1],
			Author:  parts[2],
			Date:    parts[3],
			Message: parts[4],
		})
	}
	return commits
}

func (g *Git) ShowCommit(hash string) (string, error) {
	out, err := g.run("show", "--stat", hash)
	if err != nil {
		return "", err
	}
	return out, nil
}

func (g *Git) CommitDiff(hash string) (string, error) {
	out, err := g.run("show", "--format=", hash)
	if err != nil {
		return "", err
	}
	return out, nil
}

type BranchInfo struct {
	Name     string
	Current  bool
	IsRemote bool
}

func (g *Git) Branches() ([]BranchInfo, error) {
	out, err := g.run("branch", "-a", "--no-color")
	if err != nil {
		return nil, err
	}

	var branches []BranchInfo
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		current := false
		if strings.HasPrefix(line, "* ") {
			current = true
			line = line[2:]
		} else if strings.HasPrefix(line, "+ ") {
			line = line[2:]
		} else {
			line = strings.TrimPrefix(line, "  ")
		}

		isRemote := strings.HasPrefix(line, "remotes/")
		branches = append(branches, BranchInfo{
			Name:     line,
			Current:  current,
			IsRemote: isRemote,
		})
	}
	return branches, nil
}

func (g *Git) CreateBranch(name string) error {
	_, err := g.run("checkout", "-b", name)
	return err
}

func (g *Git) SwitchBranch(name string) error {
	_, err := g.run("checkout", name)
	return err
}

func (g *Git) DeleteBranch(name string, force bool) error {
	if force {
		_, err := g.run("branch", "-D", name)
		return err
	}
	_, err := g.run("branch", "-d", name)
	return err
}

func (g *Git) MergeBranch(name string) error {
	_, err := g.run("merge", name)
	return err
}

func (g *Git) RebaseBranch(name string) error {
	_, err := g.run("rebase", name)
	return err
}

func (g *Git) RenameBranch(oldName, newName string) error {
	_, err := g.run("branch", "-m", oldName, newName)
	return err
}

type StashInfo struct {
	Index   int
	Message string
}

func (g *Git) StashList() ([]StashInfo, error) {
	out, err := g.run("stash", "list")
	if err != nil {
		return nil, err
	}

	var stashes []StashInfo
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ": ", 3)
		if len(parts) < 2 {
			continue
		}
		idxStr := strings.TrimPrefix(parts[0], "stash@{")
		idxStr = strings.TrimSuffix(idxStr, "}")
		idx, _ := strconv.Atoi(idxStr)
		msg := parts[1]
		if len(parts) > 2 {
			msg = parts[1] + ": " + parts[2]
		}
		stashes = append(stashes, StashInfo{Index: idx, Message: msg})
	}
	return stashes, nil
}

func (g *Git) StashSave(message string) error {
	if message != "" {
		_, err := g.run("stash", "push", "-m", message)
		return err
	}
	_, err := g.run("stash", "push")
	return err
}

func (g *Git) StashPop(index int) error {
	_, err := g.run("stash", "pop", fmt.Sprintf("stash@{%d}", index))
	return err
}

func (g *Git) StashApply(index int) error {
	_, err := g.run("stash", "apply", fmt.Sprintf("stash@{%d}", index))
	return err
}

func (g *Git) StashDrop(index int) error {
	_, err := g.run("stash", "drop", fmt.Sprintf("stash@{%d}", index))
	return err
}

func (g *Git) StashClear() error {
	_, err := g.run("stash", "clear")
	return err
}

func (g *Git) Diff(paths ...string) (string, error) {
	args := []string{"diff"}
	args = append(args, paths...)
	out, err := g.run(args...)
	if err != nil {
		return "", err
	}
	return out, nil
}

func (g *Git) DiffStaged(paths ...string) (string, error) {
	args := []string{"diff", "--cached"}
	args = append(args, paths...)
	out, err := g.run(args...)
	if err != nil {
		return "", err
	}
	return out, nil
}

func (g *Git) DiffFile(path string, staged bool) (string, error) {
	if staged {
		return g.DiffStaged(path)
	}
	return g.Diff(path)
}

func (g *Git) Fetch(remote string) error {
	if remote == "" {
		remote = "origin"
	}
	_, err := g.run("fetch", remote)
	return err
}

func (g *Git) Pull(remote, branch string) error {
	args := []string{"pull"}
	if remote != "" {
		args = append(args, remote)
		if branch != "" {
			args = append(args, branch)
		}
	}
	_, err := g.run(args...)
	return err
}

func (g *Git) Push(remote, branch string, force bool) error {
	args := []string{"push"}
	if force {
		args = append(args, "--force")
	}
	if remote != "" {
		args = append(args, remote)
	}
	if branch != "" {
		args = append(args, branch)
	}
	_, err := g.run(args...)
	return err
}

func (g *Git) Remotes() ([]string, error) {
	out, err := g.run("remote")
	if err != nil {
		return nil, err
	}
	var remotes []string
	for _, r := range strings.Split(out, "\n") {
		r = strings.TrimSpace(r)
		if r != "" {
			remotes = append(remotes, r)
		}
	}
	return remotes, nil
}

func (g *Git) RemoteURL(name string) string {
	out, _ := g.run("remote", "get-url", name)
	return out
}

func (g *Git) AheadBehind() (int, int) {
	out, _ := g.run("rev-list", "--left-right", "--count", "@{upstream}...HEAD")
	if out == "" {
		return 0, 0
	}
	parts := strings.Split(strings.TrimSpace(out), "\t")
	if len(parts) != 2 {
		return 0, 0
	}
	behind, _ := strconv.Atoi(parts[0])
	ahead, _ := strconv.Atoi(parts[1])
	return ahead, behind
}

func (g *Git) CheckoutCommit(hash string) error {
	_, err := g.run("checkout", hash)
	return err
}

func (g *Git) CherryPick(hash string) error {
	_, err := g.run("cherry-pick", hash)
	return err
}

func (g *Git) RevertCommit(hash string) error {
	_, err := g.run("revert", hash, "--no-edit")
	return err
}

func (g *Git) Config(key string) string {
	out, _ := g.run("config", key)
	return out
}

func (g *Git) SetConfig(key, value string) error {
	_, err := g.run("config", "--global", key, value)
	return err
}

func (g *Git) AddRemote(name, url string) error {
	_, err := g.run("remote", "add", name, url)
	return err
}

func (g *Git) RemoveRemote(name string) error {
	_, err := g.run("remote", "remove", name)
	return err
}

func (g *Git) AddToGitignore(path string) error {
	gitignorePath := g.workDir + string(os.PathSeparator) + ".gitignore"
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("\n" + path + "\n")
	return err
}
