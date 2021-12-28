package git

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	"os/exec"
	"strings"
)

// type tagMode string

const CurrentBranch = "current-branch"
const AllBranches = "all-branches"

const DefaultSort = "-version:refname"

// copied from goreleaser

// IsRepo returns true if current folder is a git repository
func IsRepo() bool {
	out, err := run([]string{"rev-parse", "--is-inside-work-tree"})
	return err == nil && strings.TrimSpace(out) == "true"
}

func getTags(tagMode, sort, pattern string) ([]string, error) {

	a := []string{"tag"}

	if pattern != "" {
		a = append(a, pattern)
	}

	if sort != "" {
		a = append(a, fmt.Sprintf("--sort=%s", sort))
	}

	if tagMode == CurrentBranch {
		a = append(a, "--merged")
	}

	tags, err := run(a)
	if err != nil {
		return nil, err
	}
	return strings.Split(tags, "\n"), nil

}

func GetTag(tagMode, pattern string) (string, error) {
	tags, err := getTags(tagMode, DefaultSort, pattern)
	if err != nil {
		return "", err
	}

	return tags[0], nil
}

func Changelog(tag string) (string, error) {
	if tag == "" {
		return gitLog("HEAD")
	} else {
		return gitLog(fmt.Sprintf("tags/%s..HEAD", tag))
	}
}

func run(args []string) (string, error) {
	args = append([]string{"-c", "log.showSignature=false"}, args...)
	//log.Println("git", strings.Join(args, " "))
	/* #nosec */
	cmd := exec.Command("git", args...)
	bts, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(string(bts))
	}
	return string(bts), nil
}

func gitLog(refs ...string) (string, error) {
	args := []string{"log", "--no-decorate", "--no-color"}
	args = append(args, refs...)
	return run(args)
}

func GetSemVer(tag, prefix string) (current *semver.Version, err error) {
	if tag == "" {
		tag = "0.0.0"
	}

	if prefix != "" {
		tag = strings.TrimPrefix(tag, prefix)
	}

	tag = "v" + tag

	return semver.NewVersion(tag)
}
