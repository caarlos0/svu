package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/alecthomas/kingpin"
	"github.com/caarlos0/svu/internal/git"
	"github.com/caarlos0/svu/internal/svu"
)

var (
	app                 = kingpin.New("svu", "semantic version util")
	nextCmd             = app.Command("next", "prints the next version based on the git log").Alias("n").Default()
	majorCmd            = app.Command("major", "new major version")
	minorCmd            = app.Command("minor", "new minor version").Alias("m")
	patchCmd            = app.Command("patch", "new patch version").Alias("p")
	currentCmd          = app.Command("current", "prints current version").Alias("c")
	pattern             = app.Flag("pattern", "limits calculations to be based on tags matching the given pattern").String()
	prefix              = app.Flag("prefix", "set a custom prefix").Default("v").String()
	stripPrefix         = app.Flag("strip-prefix", "strips the prefix from the tag").Default("false").Bool()
	preRelease          = app.Flag("pre-release", "adds a pre-release suffix to the version, without the semver mandatory dash prefix").String()
	build               = app.Flag("build", "adds a build suffix to the version, without the semver mandatory plug prefix").String()
	tagMode             = app.Flag("tag-mode", "determines if latest tag of the current or all branches will be used").Default("current-branch").Enum("current-branch", "all-branches")
	forcePatchIncrement = nextCmd.Flag("force-patch-increment", "forces a patch version increment regardless of the commit message content").Default("false").Bool()
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
		result = findNext(current, tag)
	case majorCmd.FullCommand():
		result = current.IncMajor()
	case minorCmd.FullCommand():
		result = current.IncMinor()
	case patchCmd.FullCommand():
		result = current.IncPatch()
	}

	var err error
	result, err = result.SetPrerelease(preRelease)
	if err != nil {
		return result, err
	}
	result, err = result.SetMetadata(build)
	if err != nil {
		return result, err
	}
	return result, nil
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

func findNext(current *semver.Version, tag string) semver.Version {
	log, err := git.Changelog(tag)
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
