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

func Changelog(tag string, dirs ...string) (string, error) {
	if tag == "" {
		return gitLog(dirs, "HEAD")
	} else {
		return gitLog(dirs, fmt.Sprintf("tags/%s..HEAD", tag))
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

func gitLog(dirs []string, refs ...string) (string, error) {
	args := []string{"log", "--no-decorate", "--no-color"}
	args = append(args, refs...)
	// dirs was changed from string to []string, so we handle the case where a single empty string is passed for
	// backwards compatibility
	if len(dirs) > 1 || (len(dirs) == 1 && dirs[0] != "") {
		args = append(args, "--")
		args = append(args, dirs...)
	}
	return run(args...)
}
