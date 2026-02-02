package labeler

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/jimschubert/labeler"
	log "github.com/sirupsen/logrus"
)

var version = "unk"
var date = "unk"
var commit = "unk"
var projectName = "labeler"

type CLI struct {
	Owner      string           `short:"o" env:"GITHUB_ACTOR" help:"GitHub Owner/Org name [GITHUB_ACTOR]"`
	Repo       string           `short:"r" env:"GITHUB_REPO" help:"GitHub Repo name [GITHUB_REPO]"`
	Type       string           `short:"t" env:"GITHUB_EVENT_NAME" help:"The target event type to label (issues or pull_request) [GITHUB_EVENT_NAME]"`
	Fields     []string         `default:"title,body" help:"Fields to evaluate for labeling (title, body)"`
	ID         int              `help:"The integer id of the issue or pull request"`
	Data       string           `help:"A JSON string of the 'event' type (issue event or pull request event)"`
	ConfigPath string           `name:"config-path" help:"A custom config path, relative to the repository root"`
	Version    kong.VersionFlag `short:"v" help:"Print version information"`
}

func (c *CLI) Run() error {
	labelOpts := make([]labeler.OptFn, 0)
	labelOpts = append(labelOpts, labeler.WithOwner(c.Owner))
	labelOpts = append(labelOpts, labeler.WithRepo(c.Repo))
	labelOpts = append(labelOpts, labeler.WithEvent(c.Type))
	if c.ID > 0 {
		labelOpts = append(labelOpts, labeler.WithID(c.ID))
	}
	if c.Data != "" {
		labelOpts = append(labelOpts, labeler.WithData(c.Data))
	}
	if c.ConfigPath != "" {
		labelOpts = append(labelOpts, labeler.WithConfigPath(c.ConfigPath))
	}
	if len(c.Fields) > 0 {
		fieldFlags := labeler.ParseFieldFlags(c.Fields)
		labelOpts = append(labelOpts, labeler.WithFields(fieldFlags))
	}

	l, err := labeler.NewWithOptions(labelOpts...)
	if err != nil {
		return fmt.Errorf("could not initialize labeler: %w", err)
	}
	if err = l.Execute(); err != nil {
		return fmt.Errorf("labeling failed: %w", err)
	}
	log.Info("run complete!")
	return nil
}

// Execute parses CLI arguments and runs the command.
// This is called by main.main(). It only needs to happen once.
func Execute() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name(projectName),
		kong.Description("A labeler for GitHub issues and pull requests."),
		kong.Vars{"version": fmt.Sprintf("%s (%s) - built: %s", version, commit, date)},
		kong.UsageOnError(),
	)
	err := ctx.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
