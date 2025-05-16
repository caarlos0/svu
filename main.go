package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	goversion "github.com/caarlos0/go-version"
	"github.com/caarlos0/svu/v3/internal/git"
	"github.com/caarlos0/svu/v3/internal/svu"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//go:embed description.txt
var description []byte

//go:embed examples.sh
var examples []byte

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

	rootCmd := &cobra.Command{
		Use:          "svu",
		Short:        "Semantic Version Utility",
		Long:         string(description),
		Version:      buildVersion(version, commit, date, builtBy).String(),
		Example:      paddingLeft(string(examples)),
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

	prereleaseCmd := &cobra.Command{
		Use:     "prerelease",
		Aliases: []string{"pr"},
		Short:   "Increases the build portion of the prerelease",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Action = svu.PreRelease
			return runFunc(cmd)
		},
	}
	nextCmd := &cobra.Command{
		Use:     "next",
		Aliases: []string{"n"},
		Short:   "Next version based on git history",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Action = svu.Next
			return runFunc(cmd)
		},
	}
	majorCmd := &cobra.Command{
		Use:   "major",
		Short: "New major release",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Action = svu.Major
			return runFunc(cmd)
		},
	}
	minorCmd := &cobra.Command{
		Use:     "minor",
		Short:   "New minor release",
		Aliases: []string{"m"},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Action = svu.Minor
			return runFunc(cmd)
		},
	}
	patchCmd := &cobra.Command{
		Use:     "patch",
		Short:   "New patch release",
		Aliases: []string{"p"},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Action = svu.Patch
			return runFunc(cmd)
		},
	}
	currentCmd := &cobra.Command{
		Use:     "current",
		Short:   "Current version",
		Aliases: []string{"c"},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Action = svu.Current
			return runFunc(cmd)
		},
	}
	initCmd := &cobra.Command{
		Use:     "init",
		Short:   "Creates a svu configuration file",
		Aliases: []string{"i"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return os.WriteFile(".svu.yaml", exampleConfig, 0o644)
		},
	}

	rootCmd.SetVersionTemplate("{{.Version}}")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable logs")
	rootCmd.AddCommand(initCmd)
	nextCmd.Flags().BoolVar(&opts.Always, "always", false, "if no commits trigger a version change, increment the patch")
	nextCmd.Flags().BoolVar(&opts.KeepV0, "v0", false, "prevent major version increments if current version is still v0")

	for _, cmd := range []*cobra.Command{
		nextCmd,
		majorCmd,
		minorCmd,
		patchCmd,
		currentCmd,
		prereleaseCmd,
	} {
		// init does not share these flags.
		cmd.Flags().StringVar(&opts.Pattern, "tag.pattern", "", "ignore tags that do not match the given pattern")
		cmd.Flags().StringVar(&opts.Prefix, "tag.prefix", "v", "sets a tag custom prefix")
		cmd.Flags().StringVar(&opts.TagMode, "tag.mode", git.TagModeAll, "determine if it should look for tags in all branches, or just the current one")
		cmd.Flags().StringVar(&opts.PreRelease, "prerelease", "", "sets the version prerelease")
		cmd.Flags().StringVar(&opts.Metadata, "metadata", "", "sets the version metadata")
		rootCmd.AddCommand(cmd)
	}

	for _, cmd := range []*cobra.Command{
		nextCmd,
		prereleaseCmd,
	} {
		cmd.Flags().StringSliceVar(&opts.Directories, "log.directory", nil, "only use commits that changed files in the given directories")
	}

	home, _ := os.UserHomeDir()
	config, _ := os.UserConfigDir()
	viper.AutomaticEnv()
	viper.SetEnvPrefix("svu")
	viper.AddConfigPath(".")
	if root, err := git.New().Root(); err == nil {
		viper.AddConfigPath(root)
	} else {
		log.Printf("warning: could not determine git root: %v", err)
	}
	viper.AddConfigPath(config)
	viper.AddConfigPath(home)
	viper.SetConfigName(".svu")
	viper.SetConfigType("yaml")
	cobra.OnInitialize(func() {
		if viper.ReadInConfig() == nil {
			presetRequiredFlags(rootCmd)
		}
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func presetRequiredFlags(cmd *cobra.Command) {
	viper.BindPFlags(cmd.Flags())
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) {
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

//go:embed example.svu.yaml
var exampleConfig []byte

func buildVersion(version, commit, date, builtBy string) goversion.Info {
	return goversion.GetVersionInfo(
		goversion.WithAppDetails("svu", "Semantic Version Utility", "https://github.com/caarlos0/svu"),
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

func paddingLeft(in string) string {
	var out []string
	for _, line := range strings.Split(in, "\n") {
		out = append(out, "  "+line)
	}
	return strings.Join(out, "\n")
}
