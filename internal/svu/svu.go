// Package svu provides semantic version utilities.
package svu

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/caarlos0/svu/v3/internal/git"
)

type Action uint

const (
	Next Action = iota
	Major
	Minor
	Patch
	Current
	PreRelease
)

var (
	breakingBody = regexp.MustCompile("(?m).*BREAKING[ -]CHANGE:.*")
	breaking     = regexp.MustCompile(`(?im).*(\w+)(\(.*\))?!:.*`)
	feature      = regexp.MustCompile(`(?im).*feat(\(.*\))?:.*`)
	patch        = regexp.MustCompile(`(?im).*fix(\(.*\))?:.*`)
)

type Options struct {
	Ctx         context.Context
	Action      Action
	Pattern     string
	Prefix      string
	PreRelease  string
	Metadata    string
	TagMode     string
	ConfigRoot  string
	Directories []string
	Always      bool
	KeepV0      bool
	JSON        bool
}

type VersionInfo struct {
	Version    string `json:"version"`
	Major      uint64 `json:"major"`
	Minor      uint64 `json:"minor"`
	Patch      uint64 `json:"patch"`
	Prefix     string `json:"prefix,omitempty"`
	Metadata   string `json:"metadata,omitempty"`
	Prerelease string `json:"prerelease,omitempty"`
	Build      string `json:"build,omitempty"`
}

func Version(opts Options) (string, error) {
	tag, err := git.DescribeTag(opts.Ctx, opts.TagMode, opts.Pattern)
	if err != nil {
		return "", fmt.Errorf("failed to get current tag for repo: %w", err)
	}

	current, err := getCurrentVersion(tag, opts.Prefix)
	if err != nil {
		return "", fmt.Errorf("could not get current version from tag: '%s': %w", tag, err)
	}

	result, err := nextVersion(current, tag, opts)
	if err != nil {
		return "", fmt.Errorf("could not get next tag: '%s': %w", tag, err)
	}

	if opts.JSON {
		return jsonOutput(result, opts)
	}

	return opts.Prefix + result.String(), nil
}

func nextVersion(
	current *semver.Version,
	tag string,
	opts Options,
) (semver.Version, error) {
	if opts.Action == Current {
		return *current, nil
	}

	if opts.Always {
		c, _ := current.SetMetadata("")
		c, _ = c.SetPrerelease("")
		current = &c
	}

	var result semver.Version
	var err error
	switch opts.Action {
	case Next, PreRelease:
		result, err = findNextWithGitLog(current, tag, opts)
	case Major:
		result = current.IncMajor()
	case Minor:
		result = current.IncMinor()
	case Patch:
		result = current.IncPatch()
	}
	if err != nil {
		return result, err
	}

	if opts.Action == PreRelease {
		result, err = nextPreRelease(current, &result, opts.PreRelease)
		if err != nil {
			return result, err
		}
	} else {
		result, err = result.SetPrerelease(opts.PreRelease)
		if err != nil {
			return result, err
		}
	}

	result, err = result.SetMetadata(opts.Metadata)
	if err != nil {
		return result, err
	}
	return result, nil
}

func nextPreRelease(current, next *semver.Version, prerelease string) (semver.Version, error) {
	var suffix string
	switch {
	case prerelease != "":
		// Check if the suffix already contains a version number, if it does assume the user wants to explicitly set the version so use that
		splitPreRelease := strings.Split(prerelease, ".")
		if len(splitPreRelease) > 1 {
			if _, err := strconv.Atoi(splitPreRelease[len(splitPreRelease)-1]); err == nil {
				return current.SetPrerelease(prerelease)
			}
		}

		suffix = prerelease

		// Check if the prerelease suffix is the same as the current prerelease
		preSuffix := strings.Split(current.Prerelease(), ".")[0]
		if preSuffix == prerelease {
			suffix = current.Prerelease()
		}
	case current.Prerelease() != "":
		suffix = current.Prerelease()
	default:
		return *current, fmt.Errorf(
			"--prerelease suffix is required to calculate next pre-release version as suffix could not be determined from current version: %s",
			current.String(),
		)
	}

	splitSuffix := strings.Split(suffix, ".")
	preReleaseName := splitSuffix[0]
	preReleaseVersion := 0

	currentWithoutPreRelease, _ := current.SetPrerelease("")

	// If current is a normal release (no prerelease) and the computed next is not greater than current,
	// bump the base version (patch) so that the prerelease targets the next normal version.
	if current.Prerelease() == "" && !next.GreaterThan(&currentWithoutPreRelease) {
		bumped := current.IncPatch()
		next = &bumped
	}

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
		current, err = semver.NewVersion("0.0.0")
	} else {
		current, err = semver.NewVersion(strings.TrimPrefix(tag, prefix))
	}
	return current, err
}

func findNextWithGitLog(
	current *semver.Version,
	tag string,
	opts Options,
) (semver.Version, error) {
	log, err := git.Changelog(opts.Ctx, tag, opts.Directories)
	if err != nil {
		return semver.Version{}, fmt.Errorf("failed to get changelog: %w", err)
	}

	return findNext(current, log, opts), nil
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

func findNext(current *semver.Version, changes []git.Commit, opts Options) semver.Version {
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
		if current.Major() == 0 && opts.KeepV0 {
			log.Printf("found major change, but 'keep v0' is set: %s %s\n", major.SHA, major.Title)
			return current.IncMinor()
		}
		log.Printf("found major change: %s %s\n", major.SHA, major.Title)
		return current.IncMajor()
	}

	if minor != nil {
		log.Printf("found minor change: %s %s\n", minor.SHA, minor.Title)
		return current.IncMinor()
	}

	if patch != nil {
		log.Printf("found patch change: %s %s\n", patch.SHA, patch.Title)
		return current.IncPatch()
	}

	if opts.Always {
		log.Printf("found no changes, but 'always' is set")
		return current.IncPatch()
	}
	return *current
}

func jsonOutput(v semver.Version, opts Options) (string, error) {
	info := VersionInfo{
		Prefix:     opts.Prefix,
		Version:    opts.Prefix + v.String(),
		Major:      v.Major(),
		Minor:      v.Minor(),
		Patch:      v.Patch(),
		Metadata:   v.Metadata(),
		Prerelease: v.Prerelease(),
	}

	// Split prerelease into prerelease + build if it has a dot
	if release := strings.SplitN(v.Prerelease(), ".", 2); len(release) == 2 {
		info.Prerelease = release[0]
		info.Build = release[1]
	}

	b, err := json.Marshal(info)
	if err != nil {
		return "", fmt.Errorf("failed to convert version to json: %w", err)
	}

	return string(b), nil
}
