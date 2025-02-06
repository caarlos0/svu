package svu

import (
	"github.com/caarlos0/svu/v2/internal/git"
	"github.com/caarlos0/svu/v2/internal/svu"
)

type TagMode string

const (
	CurrentBranch TagMode = git.TagModeCurrent
	AllBranches   TagMode = git.TagModeAll
)

type option func(o *svu.Options)

func Next(opts ...option) (string, error) {
	return version(append(opts, cmd(svu.NextCmd))...)
}

func Major(opts ...option) (string, error) {
	return version(append(opts, cmd(svu.MajorCmd))...)
}

func Minor(opts ...option) (string, error) {
	return version(append(opts, cmd(svu.MinorCmd))...)
}

func Patch(opts ...option) (string, error) {
	return version(append(opts, cmd(svu.PatchCmd))...)
}

func Current(opts ...option) (string, error) {
	return version(append(opts, cmd(svu.CurrentCmd))...)
}

func PreRelease(opts ...option) (string, error) {
	return version(append(opts, cmd(svu.PreReleaseCmd))...)
}

func WithPattern(pattern string) option {
	return func(o *svu.Options) {
		o.Pattern = pattern
	}
}

func WithPrefix(prefix string) option {
	return func(o *svu.Options) {
		o.Prefix = prefix
	}
}

func WithPreRelease(preRelease string) option {
	return func(o *svu.Options) {
		o.PreRelease = preRelease
	}
}

func WithBuild(build string) option {
	return func(o *svu.Options) {
		o.Build = build
	}
}

func WithDirectories(directories ...string) option {
	return func(o *svu.Options) {
		o.Directories = append(o.Directories, directories...)
	}
}

func WithTagMode(tagMode TagMode) option {
	return func(o *svu.Options) {
		o.TagMode = string(tagMode)
	}
}

func ForCurrentBranch() option {
	return WithTagMode(CurrentBranch)
}

func ForAllBranches() option {
	return WithTagMode(AllBranches)
}

func WithAlwaysPatch() option {
	return func(o *svu.Options) {
		o.Always = true
	}
}

func version(opts ...option) (string, error) {
	options := &svu.Options{
		Cmd:     svu.NextCmd,
		Prefix:  "v",
		TagMode: string(CurrentBranch),
	}
	for _, opt := range opts {
		opt(options)
	}
	return svu.Version(*options)
}

func cmd(cmd string) option {
	return func(o *svu.Options) {
		o.Cmd = cmd
	}
}
