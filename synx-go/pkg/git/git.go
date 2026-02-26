package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type GitManager struct {
	RepoDir string
}

func NewGitManager(repoDir string) *GitManager {
	return &GitManager{RepoDir: repoDir}
}

func (g *GitManager) IsRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = g.RepoDir
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func (g *GitManager) Clone(url string, dest string) error {
	cmd := exec.Command("git", "clone", url, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (g *GitManager) CurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = g.RepoDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (g *GitManager) HasChanges() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = g.RepoDir
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

func (g *GitManager) ChangedFilesCount() (int, error) {
	cmd := exec.Command("git", "status", "--short")
	cmd.Dir = g.RepoDir
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0, nil
	}
	return len(lines), nil
}

func (g *GitManager) Commit(message string) error {
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = g.RepoDir
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	commitCmd := exec.Command("git", "commit", "-m", message, "--quiet")
	commitCmd.Dir = g.RepoDir
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	return nil
}

func (g *GitManager) Push(branch string, force bool) error {
	args := []string{"push", "-u", "origin", branch}
	if force {
		args = append(args, "--force")
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = g.RepoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (g *GitManager) Pull() error {
	cmd := exec.Command("git", "pull", "--rebase")
	cmd.Dir = g.RepoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (g *GitManager) Fetch() error {
	cmd := exec.Command("git", "fetch", "origin")
	cmd.Dir = g.RepoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (g *GitManager) IsUpToDate(branch string) (bool, error) {
	localCmd := exec.Command("git", "rev-parse", "@")
	localCmd.Dir = g.RepoDir
	localHash, err := localCmd.Output()
	if err != nil {
		return false, err
	}

	remoteCmd := exec.Command("git", "rev-parse", "@{u}")
	remoteCmd.Dir = g.RepoDir
	remoteHash, err := remoteCmd.Output()
	if err != nil {
		return false, nil
	}

	return strings.TrimSpace(string(localHash)) == strings.TrimSpace(string(remoteHash)), nil
}

func (g *GitManager) Stash() (bool, error) {
	if !g.HasChanges() {
		return false, nil
	}
	cmd := exec.Command("git", "stash", "push", "-m", "synx_tmp_stash", "--quiet")
	cmd.Dir = g.RepoDir
	err := cmd.Run()
	return err == nil, err
}

func (g *GitManager) StashPop() error {
	cmd := exec.Command("git", "stash", "pop", "--quiet")
	cmd.Dir = g.RepoDir
	return cmd.Run()
}

func (g *GitManager) Log(count int) ([]string, error) {
	arg := fmt.Sprintf("-%d", count)
	cmd := exec.Command("git", "log", "--oneline", arg, "--pretty=format:%h|%ar|%s")
	cmd.Dir = g.RepoDir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}, nil
	}
	return lines, nil
}

func (g *GitManager) ResetHard(commits int) (string, error) {
	logArg := fmt.Sprintf("-%d", commits+1)
	cmdLog := exec.Command("git", "log", "--oneline", logArg)
	cmdLog.Dir = g.RepoDir
	out, err := cmdLog.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	targetLine := lines[len(lines)-1]

	resetArg := fmt.Sprintf("HEAD~%d", commits)
	cmdReset := exec.Command("git", "reset", "--hard", resetArg)
	cmdReset.Dir = g.RepoDir
	err = cmdReset.Run()
	if err != nil {
		return "", err
	}

	return targetLine, nil
}

func (g *GitManager) RmCached(file string) error {
	cmd := exec.Command("git", "rm", "--cached", file)
	cmd.Dir = g.RepoDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v: %s", err, stderr.String())
	}
	return nil
}

func (g *GitManager) LsFiles() ([]string, error) {
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = g.RepoDir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var result []string
	for _, l := range lines {
		if l != "" {
			result = append(result, l)
		}
	}
	return result, nil
}

// DiffNameStatus returns file-level status between working tree and HEAD.
// Each entry is like "M\tpath/to/file" or "A\tpath/to/file".
func (g *GitManager) DiffNameStatus() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-status", "HEAD")
	cmd.Dir = g.RepoDir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return []string{}, nil
	}
	return strings.Split(raw, "\n"), nil
}

// DiffAgainstRemote returns the full diff between local and remote branch.
func (g *GitManager) DiffAgainstRemote(branch string) (string, error) {
	cmd := exec.Command("git", "diff", "origin/"+branch, "--", ".")
	cmd.Dir = g.RepoDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// DiffStatRemote returns a --stat summary of differences with remote.
func (g *GitManager) DiffStatRemote(branch string) (string, error) {
	cmd := exec.Command("git", "diff", "--stat", "origin/"+branch, "--", ".")
	cmd.Dir = g.RepoDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// RemoteURL returns the remote origin URL.
func (g *GitManager) RemoteURL() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = g.RepoDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// UnpushedCount returns the number of commits ahead of the remote.
func (g *GitManager) UnpushedCount(branch string) (int, error) {
	cmd := exec.Command("git", "rev-list", "--count", "origin/"+branch+"..HEAD")
	cmd.Dir = g.RepoDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return 0, nil // If no upstream, return 0
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return 0, nil
	}
	count := 0
	fmt.Sscanf(raw, "%d", &count)
	return count, nil
}

// Commit structured data for Rollback TUI
type Commit struct {
	Hash    string
	Message string
	Date    string
	Author  string
}

// LogStructured returns structured commit history.
func (g *GitManager) LogStructured(limit int) ([]Commit, error) {
	cmd := exec.Command("git", "log", fmt.Sprintf("-n%d", limit), "--format=%h|%s|%ar|%an")
	cmd.Dir = g.RepoDir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}

	var commits []Commit
	for _, line := range strings.Split(raw, "\n") {
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			commits = append(commits, Commit{
				Hash:    parts[0],
				Message: parts[1],
				Date:    parts[2],
				Author:  parts[3],
			})
		}
	}
	return commits, nil
}

// ShowStat returns the stat diff for a specific commit hash.
func (g *GitManager) ShowStat(hash string) (string, error) {
	cmd := exec.Command("git", "show", "--stat", "--oneline", hash)
	cmd.Dir = g.RepoDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
