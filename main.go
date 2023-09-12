package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/alecthomas/kingpin"
	"github.com/caarlos0/svu/internal/git"
	"github.com/caarlos0/svu/internal/svu"
)

var (
	app           = kingpin.New("svu", "semantic version util")
	nextCmd       = app.Command("next", "prints the next version based on the git log").Alias("n").Default()
	majorCmd      = app.Command("major", "new major version")
	minorCmd      = app.Command("minor", "new minor version").Alias("m")
	patchCmd      = app.Command("patch", "new patch version").Alias("p")
	currentCmd    = app.Command("current", "prints current version").Alias("c")
	preReleaseCmd = app.Command("prerelease", "new pre release version based on the next version calculated from git log").
			Alias("pr")
	preRelease = preReleaseCmd.Flag("pre-release", "adds a pre-release suffix to the version, without the semver mandatory dash prefix").
			String()
	pattern     = app.Flag("pattern", "limits calculations to be based on tags matching the given pattern").String()
	prefix      = app.Flag("prefix", "set a custom prefix").Default("v").String()
	stripPrefix = app.Flag("strip-prefix", "strips the prefix from the tag").Default("false").Bool()
	build       = app.Flag("build", "adds a build suffix to the version, without the semver mandatory plug prefix").
			String()
	directory = app.Flag("directory", "specifies directory to filter commit messages by").Default("").String()
	tagMode   = app.Flag("tag-mode", "determines if latest tag of the current or all branches will be used").
			Default("current-branch").
			Enum("current-branch", "all-branches")
	forcePatchIncrement = nextCmd.Flag("force-patch-increment", "forces a patch version increment regardless of the commit message content").
				Default("false").
				Bool()
)

func main() {
	app.Author("Carlos Alexandro Becker <carlos@becker.software>")
	app.Version(buildVersion(version, commit, date, builtBy))
	app.VersionFlag.Short('v')
	app.HelpFlag.Short('h')
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	tag, err := git.DescribeTag(*tagMode, *pattern)
	app.FatalIfError(err, "failed to get current tag for repo")

	current, err := getCurrentVersion(tag)
	app.FatalIfError(err, "could not get current version from tag: '%s'", tag)

	result, err := nextVersion(cmd, current, tag, *preRelease, *build, *forcePatchIncrement)
	app.FatalIfError(err, "could not get next tag: '%s'", tag)

	if *stripPrefix {
		fmt.Println(result.String())
		return
	}
	fmt.Println(*prefix + result.String())
}

func nextVersion(cmd string, current *semver.Version, tag, preRelease, build string, force bool) (semver.Version, error) {
	if cmd == currentCmd.FullCommand() {
		return *current, nil
	}

	if force {
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
	switch cmd {
	case nextCmd.FullCommand():
		result = findNext(current, tag, *directory)
	case majorCmd.FullCommand():
		result = current.IncMajor()
	case minorCmd.FullCommand():
		result = current.IncMinor()
	case patchCmd.FullCommand():
		result = current.IncPatch()
	}

	var err error
	if cmd == preReleaseCmd.FullCommand() {
		next := findNext(current, tag, *directory)
		result, err = nextPreRelease(current, &next, preRelease)
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
		if _, err := strconv.Atoi(suffix); err == nil {
			return current.SetPrerelease(suffix)
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

func getCurrentVersion(tag string) (*semver.Version, error) {
	var current *semver.Version
	var err error
	if tag == "" {
		current, err = semver.NewVersion(strings.TrimPrefix("0.0.0", *prefix))
	} else {
		current, err = semver.NewVersion(strings.TrimPrefix(tag, *prefix))
	}
	return current, err
}

func findNext(current *semver.Version, tag string, directory string) semver.Version {
	log, err := git.Changelog(tag, directory)
	app.FatalIfError(err, "failed to get changelog")

	return svu.FindNext(current, *forcePatchIncrement, log)
}

// nolint: gochecknoglobals
var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

func buildVersion(version, commit, date, builtBy string) string {
	result := "svu version " + version
	if commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, commit)
	}
	if date != "" {
		result = fmt.Sprintf("%s\nbuilt at: %s", result, date)
	}
	if builtBy != "" {
		result = fmt.Sprintf("%s\nbuilt by: %s", result, builtBy)
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result = fmt.Sprintf("%s\nmodule version: %s, checksum: %s", result, info.Main.Version, info.Main.Sum)
	}
	return result
}
