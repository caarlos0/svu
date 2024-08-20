package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gobwas/glob"
)

const (
	AllBranchesTagMode   = "all-branches"
	CurrentBranchTagMode = "current-branch"
)

// copied from goreleaser

// IsRepo returns true if current folder is a git repository
func IsRepo() bool {
	out, err := run("rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

// Root returns the current working directory if it's inside a git repository's working tree
func Root() (string, error) {
	// Check if the current directory is inside a Git working tree
	isRepo := IsRepo()
	if !isRepo {
		return "", errors.New("not inside a Git working tree")
	}

	// Get the current working directory using `pwd`
	cwd, err := getCurrentWorkingDirectory()
	if err != nil {
		return "", err
	}

	return cwd, nil
}

// getCurrentWorkingDirectory gets the current working directory using the `pwd` command
func getCurrentWorkingDirectory() (string, error) {
	cmd := exec.Command("pwd")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func getAllTags(args ...string) ([]string, error) {
	tags, err := run(append([]string{"-c", "versionsort.suffix=-", "tag", "--sort=-version:refname"}, args...)...)
	if err != nil {
		return nil, err
	}
	return strings.Split(tags, "\n"), nil
}

func DescribeTag(tagMode string, pattern string) (string, error) {
	args := []string{}
	if tagMode == CurrentBranchTagMode {
		args = []string{"--merged"}
	}
	tags, err := getAllTags(args...)
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

func Changelog(tag string, dir string) (string, error) {
	if tag == "" {
		return gitLog(dir, "HEAD")
	} else {
		return gitLog(dir, fmt.Sprintf("tags/%s..HEAD", tag))
	}
}

func run(args ...string) (string, error) {
	extraArgs := []string{
		"-c", "log.showSignature=false",
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

func gitLog(dir string, refs ...string) (string, error) {
	args := []string{"log", "--no-decorate", "--no-color"}
	args = append(args, refs...)
	if dir != "" {
		args = append(args, "--", dir)
	}
	return run(args...)
}
