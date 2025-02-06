package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gobwas/glob"
)

// Commit is a commit with a hash, title (first line of the message), and body
// (rest of the message, not including the title).
type Commit struct {
	SHA   string
	Title string
	Body  string
}

func (c Commit) String() string {
	return c.SHA + ": " + c.Title + "\n" + c.Body
}

const (
	TagModeAll     = "all"
	TagModeCurrent = "current"
)

// copied from goreleaser

// IsRepo returns true if current folder is a git repository
func IsRepo() bool {
	out, err := run("rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

func Root() string {
	out, _ := run("rev-parse", "--show-toplevel")
	return strings.TrimSpace(out)
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
	if tagMode == TagModeCurrent {
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

func Changelog(tag string, dirs []string) ([]Commit, error) {
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

func gitLog(dirs []string, refs ...string) ([]Commit, error) {
	args := []string{"log", "--no-decorate", "--no-color", `--format=%H:%B<svu-commit-end>`}
	args = append(args, refs...)
	if len(dirs) > 0 {
		args = append(args, "--")
		args = append(args, dirs...)
	}
	s, err := run(args...)
	if err != nil {
		return nil, err
	}
	var result []Commit
	for _, commit := range strings.Split(s, "<svu-commit-end>") {
		commit = strings.TrimSpace(commit)
		if commit == "" { // accounts for the last split, which will be an empty line
			continue
		}

		hashEndIdx := strings.Index(commit, ":")
		titleEndIdx := strings.Index(commit, "\n")
		if titleEndIdx < 0 {
			titleEndIdx = len(commit)
		}

		result = append(result, Commit{
			commit[:hashEndIdx],
			commit[hashEndIdx+1 : titleEndIdx],
			commit[min(titleEndIdx+1, len(commit)):],
		})
	}
	return result, nil
}
