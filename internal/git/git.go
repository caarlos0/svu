package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// copied from goreleaser

// IsRepo returns true if current folder is a git repository
func IsRepo() bool {
	out, err := Run("rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

// Run runs a git command and returns its output or errors
func Run(args ...string) (string, error) {
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

// Clean the output
func Clean(output string, err error) (string, error) {
	output = strings.Replace(strings.Split(output, "\n")[0], "'", "", -1)
	if err != nil {
		err = errors.New(strings.TrimSuffix(err.Error(), "\n"))
	}
	return output, err
}

func DescribeTag(tagMode string) (string, error) {
	if tagMode == "all-branches" {
		tagHash, err := Clean(Run("rev-list", "--tags", "--max-count=1"))
		if err != nil {
			return "", err
		}

		return Clean(Run("describe", "--tags", tagHash))
	}

	return Clean(Run("describe", "--tags", "--abbrev=0"))
}

func Changelog(tag string) (string, error) {
	return gitLog(fmt.Sprintf("tags/%s..HEAD", tag))
}

func gitLog(refs ...string) (string, error) {
	args := []string{"log", "--no-decorate", "--no-color"}
	args = append(args, refs...)
	return Run(args...)
}
