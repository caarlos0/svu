package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/caarlos0/svu/v2/internal/git"
	"github.com/caarlos0/svu/v2/internal/svu"
	"github.com/spf13/cobra"
)

func main() {
	var opts svu.Options

	runFunc := func(cmd *cobra.Command, _ []string) error {
		version, err := svu.Version(opts)
		if err != nil {
			return err
		}
		cmd.Println(version)
		return nil
	}

	root := &cobra.Command{
		Use:          "svu",
		Short:        "Semantic Version Util",
		Long:         "Semantic Version Util (svu) is a small helper for release scripts and workflows.\nIt provides utility commands to increase specific portions of the version.\nIt can also figure the next version out automatically by looking through the git history.",
		Version:      buildVersion(version, commit, date, builtBy),
		SilenceUsage: true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			switch opts.TagMode {
			case git.AllBranchesTagMode, git.CurrentBranchTagMode:
			default:
				return fmt.Errorf("invalid tag-mode: %s", opts.TagMode)
			}
			return nil
		},
	}

	prerelease := &cobra.Command{
		Use:     "prerelease",
		Aliases: []string{"pr"},
		Short:   "Increases the build portion of the prerelease",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Cmd = svu.PreReleaseCmd
			return runFunc(cmd, args)
		},
	}
	next := &cobra.Command{
		Use:     "next",
		Aliases: []string{"n"},
		Short:   "Next version based on git history",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Cmd = svu.NextCmd
			return runFunc(cmd, args)
		},
	}
	major := &cobra.Command{
		Use:   "major",
		Short: "New major release",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Cmd = svu.MajorCmd
			return runFunc(cmd, args)
		},
	}
	minor := &cobra.Command{
		Use:     "minor",
		Short:   "New minor release",
		Aliases: []string{"m"},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Cmd = svu.MinorCmd
			return runFunc(cmd, args)
		},
	}
	patch := &cobra.Command{
		Use:     "patch",
		Short:   "New patch release",
		Aliases: []string{"p"},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Cmd = svu.PatchCmd
			return runFunc(cmd, args)
		},
	}
	current := &cobra.Command{
		Use:     "current",
		Short:   "Current version",
		Aliases: []string{"c"},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Cmd = svu.CurrentCmd
			return runFunc(cmd, args)
		},
	}

	root.PersistentFlags().StringVar(&opts.Pattern, "pattern", "", "limits calculations to be based on tags matching the given pattern")

	root.PersistentFlags().StringVar(&opts.Prefix, "prefix", "v", "sets a custom prefix")
	root.PersistentFlags().BoolVar(&opts.StripPrefix, "strip-prefix", false, "strips any prefixes the tag might have")
	root.PersistentFlags().StringVar(&opts.Directory, "directory", ".", "limit git operations to a directory")
	root.PersistentFlags().StringVar(&opts.TagMode, "tag-mode", "current-branch", "determines if latest tag of the current or all branches will be used (curent-branch, all-branches)")

	next.Flags().StringVar(&opts.Build, "build", "", "adds a build suffix to the version, without the semver mandatory plug prefix")
	next.Flags().StringVar(&opts.PreRelease, "pre-release", "", "adds a pre-release suffix to the version, without the semver mandatory dash prefix")
	next.Flags().BoolVar(&opts.ForcePatchIncrement, "force-patch-increment", false, "forces a patch version increment regardless of the commit message content")
	next.Flags().BoolVar(&opts.PreventMajorIncrementOnV0, "no-increment-v0", false, "prevent major version increments when its still v0")

	root.AddCommand(
		next,
		major,
		minor,
		patch,
		current,
		prerelease,
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
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
