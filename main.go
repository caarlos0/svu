package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"

	goversion "github.com/caarlos0/go-version"
	"github.com/caarlos0/svu/v3/internal/git"
	"github.com/caarlos0/svu/v3/internal/svu"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	var verbose bool
	var opts svu.Options

	runFunc := func(cmd *cobra.Command) error {
		version, err := svu.Version(opts)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), version)
		return err
	}

	root := &cobra.Command{
		Use:          "svu",
		Short:        "semantic version util",
		Long:         "semantic version util (svu) is a small helper for release scripts and workflows.\nIt provides utility commands to increase specific portions of the version.\nIt can also figure the next version out automatically by looking through the git history.",
		Version:      buildVersion(version, commit, date, builtBy).String(),
		SilenceUsage: true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			switch opts.TagMode {
			case git.TagModeAll, git.TagModeCurrent:
			default:
				return fmt.Errorf(
					"invalid tag-mode: %q: valid options are %q and %q",
					opts.TagMode,
					git.TagModeCurrent,
					git.TagModeAll,
				)
			}

			if verbose {
				log.SetFlags(0)
			} else {
				log.SetOutput(io.Discard)
			}
			return nil
		},
	}

	prerelease := &cobra.Command{
		Use:     "prerelease",
		Aliases: []string{"pr"},
		Short:   "Increases the build portion of the prerelease",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Action = svu.PreRelease
			return runFunc(cmd)
		},
	}
	next := &cobra.Command{
		Use:     "next",
		Aliases: []string{"n"},
		Short:   "Next version based on git history",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Action = svu.Next
			return runFunc(cmd)
		},
	}
	major := &cobra.Command{
		Use:   "major",
		Short: "New major release",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Action = svu.Major
			return runFunc(cmd)
		},
	}
	minor := &cobra.Command{
		Use:     "minor",
		Short:   "New minor release",
		Aliases: []string{"m"},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Action = svu.Minor
			return runFunc(cmd)
		},
	}
	patch := &cobra.Command{
		Use:     "patch",
		Short:   "New patch release",
		Aliases: []string{"p"},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Action = svu.Patch
			return runFunc(cmd)
		},
	}
	current := &cobra.Command{
		Use:     "current",
		Short:   "Current version",
		Aliases: []string{"c"},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Action = svu.Current
			return runFunc(cmd)
		},
	}

	root.SetVersionTemplate("{{.Version}}")

	root.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable logs")
	root.PersistentFlags().StringVar(&opts.Pattern, "tag.pattern", "", "ignore tags that do not match the given pattern")
	root.PersistentFlags().StringVar(&opts.Prefix, "tag.prefix", "v", "sets a tag custom prefix")
	root.PersistentFlags().StringVar(&opts.TagMode, "tag.mode", git.TagModeAll, "determine if it should look for tags in all branches, or just the current one")

	next.Flags().StringSliceVar(&opts.Directories, "log.directory", nil, "only use commits that changed files in the given directories")
	next.Flags().StringVar(&opts.Metadata, "metadata", "", "sets the version metadata")
	next.Flags().StringVar(&opts.PreRelease, "prerelease", "", "sets the version prerelease")
	next.Flags().BoolVar(&opts.Always, "always", false, "if no commits trigger a version change, increment the patch")
	next.Flags().BoolVar(&opts.KeepV0, "v0", false, "prevent major version increments if current version is still v0")

	root.AddCommand(
		next,
		major,
		minor,
		patch,
		current,
		prerelease,
	)

	home, _ := os.UserHomeDir()
	config, _ := os.UserConfigDir()
	viper.AutomaticEnv()
	viper.SetEnvPrefix("svu")
	viper.AddConfigPath(".")
	viper.AddConfigPath(git.Root())
	viper.AddConfigPath(config)
	viper.AddConfigPath(home)
	viper.SetConfigName(".svu")
	viper.SetConfigType("yaml")
	cobra.OnInitialize(func() {
		if viper.ReadInConfig() == nil {
			presetRequiredFlags(root)
		}
	})

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func presetRequiredFlags(cmd *cobra.Command) {
	viper.BindPFlags(cmd.Flags())
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			cmd.Flags().Set(f.Name, viper.GetString(f.Name))
		}
	})
	for _, scmd := range cmd.Commands() {
		presetRequiredFlags(scmd)
	}
}

// nolint: gochecknoglobals
var (
	version   = ""
	commit    = ""
	date      = ""
	builtBy   = ""
	treeState = ""
)

//go:embed art.txt
var asciiArt string

func buildVersion(version, commit, date, builtBy string) goversion.Info {
	return goversion.GetVersionInfo(
		goversion.WithAppDetails("svu", "Semantic Version Util", "https://github.com/caarlos0/svu"),
		goversion.WithASCIIName(asciiArt),
		func(i *goversion.Info) {
			if commit != "" {
				i.GitCommit = commit
			}
			if treeState != "" {
				i.GitTreeState = treeState
			}
			if date != "" {
				i.BuildDate = date
			}
			if version != "" {
				i.GitVersion = version
			}
			if builtBy != "" {
				i.BuiltBy = builtBy
			}
		},
	)
}
