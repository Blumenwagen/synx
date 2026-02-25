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
	// Attach to stdout/err so interactive auth prompts work natively
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (g *GitManager) Pull() error {
	cmd := exec.Command("git", "pull", "--rebase")
	cmd.Dir = g.RepoDir
	// Attach to stdout/err so interactive auth prompts work natively
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
		// No upstream or other error, assume potentially not up to date
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
	// Setup target commit info
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
