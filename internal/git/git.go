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
	out, err := run("rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

func DescribeTag(tagMode string, pattern string) (string, error) {
	gitDescribe := []string{"describe", "--tags", "--abbrev=0"}
	if pattern != "" {
		gitDescribe = append(gitDescribe, "--match", pattern)
	}

	if tagMode == "all-branches" {
		tagsArg := "--tags"
		if pattern != "" {
			tagsArg = tagsArg + "=" + pattern
		}

		tagHash, err := clean(run("rev-list", tagsArg, "--max-count=1"))
		if err != nil {
			return "", err
		}

		gitDescribe = append(gitDescribe, tagHash)
		return clean(run(gitDescribe...))
	}

	return clean(run(gitDescribe...))
}

func Changelog(tag string) (string, error) {
	return gitLog(fmt.Sprintf("tags/%s..HEAD", tag))
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

func clean(output string, err error) (string, error) {
	output = strings.Replace(strings.Split(output, "\n")[0], "'", "", -1)
	if err != nil {
		err = errors.New(strings.TrimSuffix(err.Error(), "\n"))
	}
	return output, err
}

func gitLog(refs ...string) (string, error) {
	args := []string{"log", "--no-decorate", "--no-color"}
	args = append(args, refs...)
	return run(args...)
}
