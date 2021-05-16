package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/alecthomas/kingpin"
	"github.com/caarlos0/svu/internal/git"
	"github.com/caarlos0/svu/internal/svu"
)

var (
	version             = "dev"
	app                 = kingpin.New("svu", "semantic version util")
	nextCmd             = app.Command("next", "prints the next version based on the git log").Alias("n").Default()
	majorCmd            = app.Command("major", "new major version")
	minorCmd            = app.Command("minor", "new minor version").Alias("m")
	patchCmd            = app.Command("patch", "new patch version").Alias("p")
	currentCmd          = app.Command("current", "prints current version").Alias("c")
	metadata            = app.Flag("metadata", "discards pre-release and build metadata if set to false").Default("true").Bool()
	preRelease          = app.Flag("pre-release", "discards pre-release metadata if set to false").Default("true").Bool()
	build               = app.Flag("build", "discards build metadata if set to false").Default("true").Bool()
	tagMode             = app.Flag("tag-mode", "determines if latest tag of the current or all branches will be used").Default("current-branch").Enum("current-branch", "all-branches")
	forcePatchIncrement = nextCmd.Flag("force-patch-increment", "forces a patch version increment regardless of the commit message content").Default("false").Bool()
)

func main() {
	app.Author("Carlos Alexandro Becker <carlos@becker.software>")
	app.Version("svu version " + version)
	app.VersionFlag.Short('v')
	app.HelpFlag.Short('h')
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	tag, err := git.DescribeTag(*tagMode)
	app.FatalIfError(err, "failed to get current tag for repo")

	current, err := semver.NewVersion(tag)
	app.FatalIfError(err, "version %s is not semantic", tag)

	if !*metadata {
		current = unsetMetadata(current)
	}

	if !*preRelease {
		current = unsetPreRelease(current)
	}

	if !*build {
		current = unsetBuild(current)
	}

	var prefix string
	if strings.HasPrefix(tag, "v") {
		prefix = "v"
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
	fmt.Printf("%s%s\n", prefix, result.String())
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
