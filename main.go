package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/caarlos0/svu/internal/git"
	"github.com/masterminds/semver"
)

var (
	version      = "dev"
	app          = kingpin.New("svu", "semantic version util")
	nextCmd      = app.Command("next", "prints the next version based on the git log").Alias("n").Default()
	majorCmd     = app.Command("major", "new major version")
	minorCmd     = app.Command("minor", "new minor version").Alias("m")
	patchCmd     = app.Command("patch", "new patch version").Alias("p")
	currentCmd   = app.Command("current", "prints current version").Alias("c")
	noMetadata   = app.Flag("no-metadata", "discards pre-release and build metadata").Bool()
	noPreRelease = app.Flag("no-pre-release", "discards pre-release metadata").Bool()
	noBuild      = app.Flag("no-build", "discards build metadata").Bool()
)

func main() {
	app.Author("Carlos Alexandro Becker <caarlos0@gmail.com>")
	app.Version("svu version " + version)
	app.VersionFlag.Short('v')
	app.HelpFlag.Short('h')
	var cmd = kingpin.MustParse(app.Parse(os.Args[1:]))

	tag, err := getTag()
	app.FatalIfError(err, "failed to get current tag for repo")

	current, err := semver.NewVersion(tag)
	app.FatalIfError(err, "version %s is not semantic", tag)

	if *noMetadata {
		current = unsetMetadata(current)
	}

	if *noPreRelease {
		current = unsetPreRelease(current)
	}

	if *noBuild {
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

var breaking = regexp.MustCompile("(?im).*breaking change:.*")
var breakingBang = regexp.MustCompile("(?im).*(feat|fix)(\\(.*\\))?!:.*")
var feature = regexp.MustCompile("(?im).*feat(\\(.*\\))?:.*")
var patch = regexp.MustCompile("(?im).*fix(\\(.*\\))?:.*")

func unsetPreRelease(current *semver.Version) *semver.Version {
	newV, _ := current.SetPrerelease("")

	return &newV
}

func unsetBuild(current *semver.Version) *semver.Version {
	newV, _ := current.SetMetadata("")

	return &newV
}

func unsetMetadata(current *semver.Version) *semver.Version {
	newV := unsetBuild(current)
	newV = unsetPreRelease(newV)

	return newV
}

func findNext(current *semver.Version, tag string) semver.Version {
	log, err := getChangelog(tag)
	app.FatalIfError(err, "failed to get changelog")

	if breaking.MatchString(log) || breakingBang.MatchString(log) {
		return current.IncMajor()
	}

	if feature.MatchString(log) {
		return current.IncMinor()
	}

	if patch.MatchString(log) {
		return current.IncPatch()
	}

	return *current
}

func getTag() (string, error) {
	return git.Clean(git.Run("describe", "--tags", "--abbrev=0"))
}

func getChangelog(tag string) (string, error) {
	return gitLog(fmt.Sprintf("tags/%s..HEAD", tag))
}

func gitLog(refs ...string) (string, error) {
	var args = []string{"log", "--no-decorate", "--no-color"}
	args = append(args, refs...)
	return git.Run(args...)
}
