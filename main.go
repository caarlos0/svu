package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/alecthomas/kingpin"
	"github.com/caarlos0/svu/internal/config"
	"github.com/caarlos0/svu/internal/git"
)

var (
	svuConfig        = config.Config{}
	svuConfigPresent = false
	version          = "dev"
	app              = kingpin.New("svu", "semantic version util")
	nextCmd          = app.Command("next", "prints the next version based on the git log").Alias("n").Default()
	majorCmd         = app.Command("major", "new major version")
	minorCmd         = app.Command("minor", "new minor version").Alias("m")
	patchCmd         = app.Command("patch", "new patch version").Alias("p")
	currentCmd       = app.Command("current", "prints current version").Alias("c")
	metadata         = app.Flag("metadata", "discards pre-release and build metadata if set to false").Default("true").Bool()
	preRelease       = app.Flag("pre-release", "discards pre-release metadata if set to false").Default("true").Bool()
	build            = app.Flag("build", "discards build metadata if set to false").Default("true").Bool()
	tagMode          = app.Flag("tag-mode", "determines if latest tag of the current or all branches will be used").Default("current-branch").Enum("current-branch", "all-branches")
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

	svuConfig, err = config.Load(".svu.yml")
	if err == nil {
		svuConfigPresent = true
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

func getTypes() ([]string, []string, []string) {
	featureTypes := []string{"feat"}
	fixTypes := []string{"fix"}

	if svuConfigPresent {
		if len(svuConfig.AdditionalFeatureTypes) > 0 {
			featureTypes = append(featureTypes, svuConfig.AdditionalFeatureTypes...)
		}
		if len(svuConfig.AdditionalFixTypes) > 0 {
			fixTypes = append(fixTypes, svuConfig.AdditionalFixTypes...)
		}
	}

	allTypes := append(featureTypes, fixTypes...)

	return fixTypes, featureTypes, allTypes
}

func findNext(current *semver.Version, tag string) semver.Version {
	fixTypes, featureTypes, allTypes := getTypes()
	breaking := regexp.MustCompile("(?im).*breaking change:.*")
	breakingBang := regexp.MustCompile(fmt.Sprintf("(?im).*(%s)(\\(.*\\))?!:.*", strings.Join(allTypes, "|")))
	feature := regexp.MustCompile(fmt.Sprintf("(?im).*(%s)(\\(.*\\))?:.*", strings.Join(featureTypes, "|")))
	patch := regexp.MustCompile(fmt.Sprintf("(?im).*(%s)(\\(.*\\))?:.*", strings.Join(fixTypes, "|")))

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
	if *tagMode == "all-branches" {
		tagHash, err := git.Clean(git.Run("rev-list", "--tags", "--max-count=1"))
		if err != nil {
			return "", err
		}

		return git.Clean(git.Run("describe", "--tags", tagHash))
	}

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
