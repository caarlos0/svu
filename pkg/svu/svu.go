// Package svu provides a Go API to SVU.
package svu

import (
	"github.com/caarlos0/svu/v2/internal/git"
	"github.com/caarlos0/svu/v2/internal/svu"
)

type option func(o *svu.Options)

// Next returns the next version based on the git log.
func Next(opts ...option) (string, error) {
	return version(append(opts, cmd(svu.Next))...)
}

// Major increase the major part of the version.
func Major(opts ...option) (string, error) {
	return version(append(opts, cmd(svu.Major))...)
}

// Minor increase the minor part of the version.
func Minor(opts ...option) (string, error) {
	return version(append(opts, cmd(svu.Minor))...)
}

// Patch increase the patch part of the version.
func Patch(opts ...option) (string, error) {
	return version(append(opts, cmd(svu.Patch))...)
}

// Current returns the current version.
func Current(opts ...option) (string, error) {
	return version(append(opts, cmd(svu.Current))...)
}

// PreRelease returns the next pre-release version.
func PreRelease(opts ...option) (string, error) {
	return version(append(opts, cmd(svu.PreRelease))...)
}

// WithPattern ignores tags that do not match the given pattern.
func WithPattern(pattern string) option {
	return func(o *svu.Options) {
		o.Pattern = pattern
	}
}

// WithPrefix sets the version prefix.
func WithPrefix(prefix string) option {
	return func(o *svu.Options) {
		o.Prefix = prefix
	}
}

// WithPreRelease sets the version prerelease.
func WithPreRelease(prerelease string) option {
	return func(o *svu.Options) {
		o.PreRelease = prerelease
	}
}

// WithMetadata sets the version metadata.
func WithMetadata(metadata string) option {
	return func(o *svu.Options) {
		o.Metadata = metadata
	}
}

// WithDirectories only use commits that changed files in the given directories.
func WithDirectories(directories ...string) option {
	return func(o *svu.Options) {
		o.Directories = append(o.Directories, directories...)
	}
}

// ForCurrentBranch look for tags in the current branch only.
func ForCurrentBranch() option {
	return func(o *svu.Options) {
		o.TagMode = git.TagModeCurrent
	}
}

// ForAllBranches look for tags in all branches.
func ForAllBranches() option {
	return func(o *svu.Options) {
		o.TagMode = git.TagModeAll
	}
}

// Always if no commits would have increased the version, increase the
// patch portion anyway.
func Always() option {
	return func(o *svu.Options) {
		o.Always = true
	}
}

// KeepV0 prevents major upgrades if current version is a v0.
func KeepV0() option {
	return func(o *svu.Options) {
		o.KeepV0 = true
	}
}

func version(opts ...option) (string, error) {
	options := &svu.Options{
		Action:  svu.Next,
		Prefix:  "v",
		TagMode: git.TagModeCurrent,
	}
	for _, opt := range opts {
		opt(options)
	}
	return svu.Version(*options)
}

func cmd(cmd svu.Action) option {
	return func(o *svu.Options) {
		o.Action = cmd
	}
}
