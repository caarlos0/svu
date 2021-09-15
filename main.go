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
	metadata            = app.Flag("metadata", "discards pre-release and build metadata if set to false").Default("true").Bool()
	pattern             = app.Flag("pattern", "limits calculations to be based on tags matching the given pattern").String()
	preRelease          = app.Flag("pre-release", "discards pre-release metadata if set to false").Default("true").Bool()
	build               = app.Flag("build", "discards build metadata if set to false").Default("true").Bool()
	prefix              = app.Flag("prefix", "set a custom prefix").Default("v").String()
	stripPrefix         = app.Flag("strip-prefix", "strips the prefix from the tag").Default("false").Bool()
	tagMode             = app.Flag("tag-mode", "determines if latest tag of the current or all branches will be used").Default("current-branch").Enum("current-branch", "all-branches")
	forcePatchIncrement = nextCmd.Flag("force-patch-increment", "forces a patch version increment regardless of the commit message content").Default("false").Bool()
)

func main() {
	app.Author("Carlos Alexandro Becker <carlos@becker.software>")
	app.Version(buildVersion(version, commit, date, builtBy))
	app.VersionFlag.Short('v')
	app.HelpFlag.Short('h')
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	var current *semver.Version
	tag, err := git.DescribeTag(*tagMode, *pattern)
	app.FatalIfError(err, "issues doing git.DescribeTag on %v", tag)
	if tag != "" {
		current, err = semver.NewVersion(strings.TrimPrefix(tag, *prefix))
		app.FatalIfError(err, "version %s is not semantic", tag)
	} else {
		current, err = semver.NewVersion("0.0.0")
		app.FatalIfError(err, "version %s is not semantic", current)
	}

	if !*metadata {
		current = unsetMetadata(current)
	}

	if !*preRelease {
		current = unsetPreRelease(current)
	}

	if !*build {
		current = unsetBuild(current)
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
	case currentCmd.FullCommand():
		result = *current
	}
	fmt.Println(getVersion(tag, *prefix, result.String(), *stripPrefix))
}

func getVersion(tag, prefix, result string, stripPrefix bool) string {
	if stripPrefix {
		prefix = ""
	}
	return prefix + result
}

func unsetPreRelease(current *semver.Version) *semver.Version {
	newV, _ := current.SetPrerelease("")

	return &newV
}

func unsetBuild(current *semver.Version) *semver.Version {
	newV, _ := current.SetMetadata("")

	return &newV
}

func unsetMetadata(current *semver.Version) *semver.Version {
	return unsetBuild(unsetPreRelease(current))
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
