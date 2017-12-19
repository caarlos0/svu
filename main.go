package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/masterminds/semver"
)

var (
	version = "dev"
	app     = kingpin.New("svu", "semantic version util")
	repo    = app.Flag("path", "git repository path").
		Default(".").
		Short('p').
		ExistingDir()
	major   = app.Command("major", "new major version")
	minor   = app.Command("minor", "new minor version").Alias("m")
	patch   = app.Command("patch", "new patch version").Alias("p")
	nothing = app.Command("current", "prints current version").Alias("c").Default()
)

func main() {
	app.Author("Carlos Alexandro Becker <caarlos0@gmail.com>")
	app.Version("svu version " + version)
	app.VersionFlag.Short('v')
	app.HelpFlag.Short('h')
	var cmd = kingpin.MustParse(app.Parse(os.Args[1:]))

	tag, err := tagFromRepo(*repo)
	app.FatalIfError(err, "failed to get current tag for repo %s", *repo)
	current, err := semver.NewVersion(tag)
	app.FatalIfError(err, "version %s is not semantic", tag)
	var prefix string
	if strings.HasPrefix(tag, "v") {
		prefix = "v"
	}

	var next semver.Version
	switch cmd {
	case major.FullCommand():
		next = current.IncMajor()
	case minor.FullCommand():
		next = current.IncMinor()
	case patch.FullCommand():
		next = current.IncPatch()
	case nothing.FullCommand():
		next = *current
	}
	fmt.Printf("%s%s\n", prefix, next.String())
}

func tagFromRepo(path string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	cmd.Dir = path
	bts, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.Split(string(bts), "\n")[0], nil
}
