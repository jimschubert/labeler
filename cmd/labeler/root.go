package labeler

import (
	"fmt"
	"os"

	"github.com/jimschubert/labeler"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var version = "unk"
var date = "unk"
var commit = "unk"
var projectName = "labeler"

// rootCmd represents the base command when called without any subcommands
var rootCmd = newRootCmd()

//goland:noinspection GoErrorStringFormat
func newRootCmd() *cobra.Command {
	type options struct {
		Owner string
		Repo  string
		Type  string
		ID    int
		Data  string
	}
	o := options{
		Owner: os.Getenv("GITHUB_ACTOR"),
		Repo:  os.Getenv("GITHUB_REPO"),
		Type:  os.Getenv("GITHUB_EVENT_NAME"),
	}
	c := cobra.Command{
		Use:     "labeler",
		Short:   "A labeler for GitHub issues and pull requests.",
		Version: fmt.Sprintf("%s (%s) - built: %s", version, commit, date),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := labeler.New(o.Owner, o.Repo, o.Type, o.ID, &o.Data)
			if err != nil {
				return fmt.Errorf("could not initialize labeler: %w", err)
			}
			if err = l.Execute(); err != nil {
				return fmt.Errorf("labeling failed: %w", err)
			}
			log.Info("run complete!")
			return nil
		},
	}

	c.Flags().StringVarP(&o.Owner, "owner", "o", o.Owner, "GitHub Owner/Org name [GITHUB_ACTOR]")
	c.Flags().StringVarP(&o.Repo, "repo", "r", o.Repo, "GitHub Repo name [GITHUB_REPO]")
	c.Flags().StringVarP(&o.Type, "type", "t", o.Type, "The target event type to label (issues or pull_request) [GITHUB_EVENT_NAME]")
	c.Flags().IntVarP(&o.ID, "id", "", o.ID, "The integer id of the issue or pull request")
	c.Flags().StringVarP(&o.Data, "data", "", o.Data, "A JSON string of the 'event' type (issue event or pull request event)")

	return &c
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
