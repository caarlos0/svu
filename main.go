package main

import (
	"fmt"
	"os"
	"path"
	"runtime/debug"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/caarlos0/svu/v2/internal/git"
	"github.com/caarlos0/svu/v2/internal/svu"
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
	preRelease = app.Flag("pre-release", "adds a pre-release suffix to the version, without the semver mandatory dash prefix").
			String()
	pattern     = app.Flag("pattern", "limits calculations to be based on tags matching the given pattern").Default(defaults("pattern")).String()
	prefix      = app.Flag("prefix", "set a custom prefix").Default(defaults("prefix")).String()
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
	preventMajorIncrementOnV0 = nextCmd.Flag("no-increment-v0", "prevent major version increments when its still v0").
					Default("false").
					Bool()
)

func defaults(flag string) string {
	var def, pat string
	switch flag {
	case "prefix":
		def, pat = "v", "v"
	case "pattern":
		def, pat = "", "*"
	default:
		return ""
	}

	cwd, wdErr := os.Getwd()
	gitRoot, grErr := git.Root()
	if wdErr == nil && grErr == nil && cwd != gitRoot {
		prefix := strings.TrimPrefix(cwd, gitRoot)
		prefix = strings.TrimPrefix(prefix, string(os.PathSeparator))
		prefix = strings.TrimSuffix(prefix, string(os.PathSeparator))
		return path.Join(prefix, pat)
	}

	return def
}

func main() {
	app.Author("Carlos Alexandro Becker <carlos@becker.software>")
	app.Version(buildVersion(version, commit, date, builtBy))
	app.VersionFlag.Short('v')
	app.HelpFlag.Short('h')
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	version, err := svu.Version(svu.Options{
		Cmd:                       cmd,
		Pattern:                   *pattern,
		Prefix:                    *prefix,
		StripPrefix:               *stripPrefix,
		PreRelease:                *preRelease,
		Build:                     *build,
		Directory:                 *directory,
		TagMode:                   *tagMode,
		ForcePatchIncrement:       *forcePatchIncrement,
		PreventMajorIncrementOnV0: *preventMajorIncrementOnV0,
	})
	app.FatalIfError(err, "")
	fmt.Println(version)
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
