package svu

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/caarlos0/svu/v2/internal/git"
)

const (
	NextCmd       = "next"
	MajorCmd      = "major"
	MinorCmd      = "minor"
	PatchCmd      = "patch"
	CurrentCmd    = "current"
	PreReleaseCmd = "prerelease"
)

var (
	breakingBody = regexp.MustCompile("(?m).*BREAKING[ -]CHANGE:.*")
	breaking     = regexp.MustCompile(`(?im).*(\w+)(\(.*\))?!:.*`)
	feature      = regexp.MustCompile(`(?im).*feat(\(.*\))?:.*`)
	patch        = regexp.MustCompile(`(?im).*fix(\(.*\))?:.*`)
)

type Options struct {
	Cmd                       string
	Pattern                   string
	Prefix                    string
	StripPrefix               bool
	PreRelease                string
	Build                     string
	Directory                 string
	TagMode                   string
	ForcePatchIncrement       bool
	PreventMajorIncrementOnV0 bool
}

func Version(opts Options) (string, error) {
	tag, err := git.DescribeTag(string(opts.TagMode), opts.Pattern)
	if err != nil {
		return "", fmt.Errorf("failed to get current tag for repo: %w", err)
	}

	current, err := getCurrentVersion(tag, opts.Prefix)
	if err != nil {
		return "", fmt.Errorf("could not get current version from tag: '%s': %w", tag, err)
	}

	result, err := nextVersion(
		string(opts.Cmd),
		current,
		tag,
		opts.PreRelease,
		opts.Build,
		opts.Directory,
		opts.PreventMajorIncrementOnV0,
		opts.ForcePatchIncrement,
	)
	if err != nil {
		return "", fmt.Errorf("could not get next tag: '%s': %w", tag, err)
	}

	if opts.StripPrefix {
		return result.String(), nil
	}
	return opts.Prefix + result.String(), nil
}

func nextVersion(cmd string, current *semver.Version, tag, preRelease, build, directory string, preventMajorIncrementOnV0, forcePatchIncrement bool) (semver.Version, error) {
	if cmd == CurrentCmd {
		return *current, nil
	}

	if forcePatchIncrement {
		c, err := current.SetMetadata("")
		if err != nil {
			return c, err
		}
		c, err = c.SetPrerelease("")
		if err != nil {
			return c, err
		}
		current = &c
	}

	var result semver.Version
	var err error
	switch cmd {
	case NextCmd, PreReleaseCmd:
		result, err = findNextWithGitLog(current, tag, directory, preventMajorIncrementOnV0, forcePatchIncrement)
	case MajorCmd:
		result = current.IncMajor()
	case MinorCmd:
		result = current.IncMinor()
	case PatchCmd:
		result = current.IncPatch()
	}
	if err != nil {
		return result, err
	}

	if cmd == PreReleaseCmd {
		result, err = nextPreRelease(current, &result, preRelease)
		if err != nil {
			return result, err
		}
	} else {
		result, err = result.SetPrerelease(preRelease)
		if err != nil {
			return result, err
		}
	}

	result, err = result.SetMetadata(build)
	if err != nil {
		return result, err
	}
	return result, nil
}

func nextPreRelease(current, next *semver.Version, preRelease string) (semver.Version, error) {
	suffix := ""
	if preRelease != "" {
		// Check if the suffix already contains a version number, if it does assume the user wants to explicitly set the version so use that
		splitPreRelease := strings.Split(preRelease, ".")
		if len(splitPreRelease) > 1 {
			if _, err := strconv.Atoi(splitPreRelease[len(splitPreRelease)-1]); err == nil {
				return current.SetPrerelease(preRelease)
			}
		}

		suffix = preRelease

		// Check if the prerelease suffix is the same as the current prerelease
		preSuffix := strings.Split(current.Prerelease(), ".")[0]
		if preSuffix == preRelease {
			suffix = current.Prerelease()
		}
	} else if current.Prerelease() != "" {
		suffix = current.Prerelease()
	} else {
		return *current, fmt.Errorf(
			"--pre-release suffix is required to calculate next pre-release version as suffix could not be determined from current version: %s",
			current.String(),
		)
	}

	splitSuffix := strings.Split(suffix, ".")
	preReleaseName := splitSuffix[0]
	preReleaseVersion := 0

	currentWithoutPreRelease, _ := current.SetPrerelease("")

	if !next.GreaterThan(&currentWithoutPreRelease) {
		preReleaseVersion = -1
		if len(splitSuffix) == 2 {
			preReleaseName = splitSuffix[0]
			preReleaseVersion, _ = strconv.Atoi(splitSuffix[1])
		} else if len(splitSuffix) > 2 {
			preReleaseName = splitSuffix[len(splitSuffix)-1]
		}

		preReleaseVersion++
	}

	return next.SetPrerelease(fmt.Sprintf("%s.%d", preReleaseName, preReleaseVersion))
}

func getCurrentVersion(tag, prefix string) (*semver.Version, error) {
	var current *semver.Version
	var err error
	if tag == "" {
		current, err = semver.NewVersion(strings.TrimPrefix("0.0.0", prefix))
	} else {
		current, err = semver.NewVersion(strings.TrimPrefix(tag, prefix))
	}
	return current, err
}

func findNextWithGitLog(current *semver.Version, tag string, directory string, preventMajorIncrementOnV0, forcePatchIncrement bool) (semver.Version, error) {
	log, err := git.Changelog(tag, directory)
	if err != nil {
		return semver.Version{}, fmt.Errorf("failed to get changelog: %w", err)
	}

	return findNext(current, preventMajorIncrementOnV0, forcePatchIncrement, log), nil
}

func isBreaking(commit git.Commit) bool {
	return breakingBody.MatchString(commit.Body) || breaking.MatchString(commit.Title)
}

func isFeature(commit git.Commit) bool {
	return feature.MatchString(commit.Title)
}

func isPatch(commit git.Commit) bool {
	return patch.MatchString(commit.Title)
}

func findNext(current *semver.Version, preventMajorIncrementOnV0, forcePatchIncrement bool, changes []git.Commit) semver.Version {
	var major, minor, patch *git.Commit
	for _, commit := range changes {
		if isBreaking(commit) {
			major = &commit
			break // no bigger change allowed, so we're done
		}

		if minor == nil && isFeature(commit) {
			minor = &commit
		}

		if patch == nil && isPatch(commit) {
			patch = &commit
		}
	}

	if major != nil {
		if current.Major() == 0 && preventMajorIncrementOnV0 {
			_, _ = fmt.Fprintf(os.Stderr, "found major change, but prevent major increment is set: %s %s\n", major.SHA, major.Title)
			return current.IncMinor()
		}
		_, _ = fmt.Fprintf(os.Stderr, "found major change: %s %s\n", major.SHA, major.Title)
		return current.IncMajor()
	}

	if minor != nil {
		_, _ = fmt.Fprintf(os.Stderr, "found minor change: %s %s\n", minor.SHA, minor.Title)
		return current.IncMinor()
	}

	if patch != nil {
		_, _ = fmt.Fprintf(os.Stderr, "found patch change: %s %s\n", patch.SHA, patch.Title)
		return current.IncPatch()
	}

	if forcePatchIncrement {
		_, _ = fmt.Fprintln(os.Stderr, "found no changes, but force patch increment is set")
		return current.IncPatch()
	}
	return *current
}
