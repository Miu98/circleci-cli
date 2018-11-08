package cmd

import (
	"fmt"
	"os/exec"
	"regexp"

	"github.com/CircleCI-Public/circleci-cli/logger"
	"github.com/CircleCI-Public/circleci-cli/settings"
	"github.com/CircleCI-Public/circleci-cli/update"
	"github.com/CircleCI-Public/circleci-cli/version"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

type updateCommandOptions struct {
	cfg    *settings.Config
	log    *logger.Logger
	dryRun bool
	args   []string
}

func newUpdateCommand(config *settings.Config) *cobra.Command {
	opts := updateCommandOptions{
		cfg:    config,
		dryRun: false,
	}

	update := &cobra.Command{
		Use:   "update",
		Short: "Update the tool to the latest version",
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			opts.cfg.SkipUpdateCheck = true
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.args = args
			opts.log = logger.NewLogger(config.Debug)
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return updateCLI(opts)
		},
	}

	update.AddCommand(&cobra.Command{
		Use:    "check",
		Hidden: true,
		Short:  "Check if there are any updates available",
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			opts.cfg.SkipUpdateCheck = true
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.args = args
			opts.dryRun = true
			opts.log = logger.NewLogger(config.Debug)
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return updateCLI(opts)
		},
	})

	update.AddCommand(&cobra.Command{
		Use:    "install",
		Hidden: true,
		Short:  "Update the tool to the latest version",
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			opts.cfg.SkipUpdateCheck = true
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.args = args
			opts.log = logger.NewLogger(config.Debug)
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return updateCLI(opts)
		},
	})

	update.AddCommand(&cobra.Command{
		Use:    "build-agent",
		Hidden: true,
		Short:  "Update the build agent to the latest version",
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			opts.cfg.SkipUpdateCheck = true
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.args = args
			opts.log = logger.NewLogger(config.Debug)
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return updateBuildAgent(opts)
		},
	})

	update.PersistentFlags().BoolVar(&opts.dryRun, "check", false, "Check if there are any updates available without installing")

	update.Flags().BoolVar(&testing, "testing", false, "Enable test mode to bypass interactive UI.")
	if err := update.Flags().MarkHidden("testing"); err != nil {
		panic(err)
	}

	return update
}

var picardRepo = "circleci/picard"

func updateBuildAgent(opts updateCommandOptions) error {
	latestSha256, err := findLatestPicardSha()

	if err != nil {
		return err
	}

	opts.log.Infof("Latest build agent is version %s", latestSha256)

	return nil
}

// Still depends on a function in cmd/build.go
func findLatestPicardSha() (string, error) {

	if err := ensureDockerIsAvailable(); err != nil {
		return "", err
	}

	outputBytes, err := exec.Command("docker", "pull", picardRepo).CombinedOutput() // #nosec

	if err != nil {
		return "", errors.Wrap(err, "failed to pull latest docker image")
	}

	output := string(outputBytes)
	sha256 := regexp.MustCompile("(?m)sha256:[0-9a-f]+")
	latest := sha256.FindString(output)

	if latest == "" {
		return "", fmt.Errorf("failed to parse sha256 from docker pull output")
	}

	// This function still lives in cmd/build.go
	err = storeBuildAgentSha(latest)

	if err != nil {
		return "", err
	}

	return latest, nil
}

func updateCLI(opts updateCommandOptions) error {
	slug := "CircleCI-Public/circleci-cli"

	check, err := update.CheckForUpdates(opts.cfg.GitHubAPI, slug, version.Version, PackageManager)
	if err != nil {
		return err
	}

	if update.IsLatestVersion(check) {
		return errors.New("already up-to-date")
	}

	opts.log.Debug(update.DebugVersion(check))
	opts.log.Info(update.ReportVersion(check))

	if opts.dryRun {
		opts.log.Info("You can update with `circleci update install`")
		return nil
	}

	message, err := update.InstallLatest(check)
	if err != nil {
		return err
	}

	opts.log.Infof(message)

	return nil
}
