package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

const (
	AllBranchesTagMode   = "all-branches"
	CurrentBranchTagMode = "current-branch"
)

// copied from goreleaser

type Repository struct {
	GitWorkTree  string
	GitDirectory string
}

// NewRepository creates a Repository. gitworktree and gitdirectory will default to 'pwd' and 'pwd/.git'
func NewRepository(gitworktree, gitdirectory string) (*Repository, error) {
	var err error
	if gitworktree == "" {
		gitworktree, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	gitworktree = filepath.Clean(gitworktree)
	if gitdirectory == "" {
		gitdirectory, err = os.Getwd()
		if err != nil {
			return nil, err
		}
		gitdirectory = filepath.Join(gitdirectory, ".git")
	}
	gitdirectory = filepath.Clean(gitdirectory)

	r := &Repository{
		GitWorkTree:  gitworktree,
		GitDirectory: gitdirectory,
	}
	return r, nil
}

// IsRepo returns true if the current work-tree and git-dir are considered a git repository by the git binary.
// unless specified by --work-tree and --git-dir, the current working directory will be considered
func (r *Repository) IsRepo() bool {
	_, err := r.run("status")
	return err == nil
}

// Root returns the root of the git Repository
func (r *Repository) Root() (string, error) {
	out, err := r.run("rev-parse", "--show-toplevel")
	return strings.TrimSpace(out), err
}

func (r *Repository) getAllTags(args ...string) ([]string, error) {
	tags, err := r.run(append([]string{"-c", "versionsort.suffix=-", "tag", "--sort=-version:refname"}, args...)...)
	if err != nil {
		return nil, err
	}
	return strings.Split(tags, "\n"), nil
}

func (r *Repository) DescribeTag(tagMode string, pattern string) (string, error) {
	args := []string{}
	if tagMode == CurrentBranchTagMode {
		args = []string{"--merged"}
	}
	tags, err := r.getAllTags(args...)
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", nil
	}
	if pattern == "" {
		return tags[0], nil
	}

	g, err := glob.Compile(pattern)
	if err != nil {
		return "", err
	}
	for _, tag := range tags {
		if g.Match(tag) {
			return tag, nil
		}
	}
	return "", fmt.Errorf("no tags match '%s'", pattern)
}

func (r *Repository) Changelog(tag string, dir string) (string, error) {
	if tag == "" {
		return r.gitLog(dir, "HEAD")
	} else {
		return r.gitLog(dir, fmt.Sprintf("tags/%s..HEAD", tag))
	}
}

func (r *Repository) run(args ...string) (string, error) {
	extraArgs := []string{"-c", "log.showSignature=false"}
	if r.GitWorkTree != "" {
		extraArgs = append(extraArgs, fmt.Sprintf("--work-tree=%s", r.GitWorkTree))
	}
	if r.GitDirectory != "" {
		extraArgs = append(extraArgs, fmt.Sprintf("--git-dir=%s", r.GitDirectory))
	}
	args = append(extraArgs, args...)
	/* #nosec */
	cmd := exec.Command("git", args...)
	bts, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(string(bts))
	}
	return string(bts), nil
}

func (r *Repository) gitLog(dir string, refs ...string) (string, error) {
	args := []string{"log", "--no-decorate", "--no-color"}
	args = append(args, refs...)
	if dir != "" {
		args = append(args, "--", dir)
	}
	return r.run(args...)
}
