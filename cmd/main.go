package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"

	"github.com/jimschubert/labeler"
)

var version = ""
var date = ""
var commit = ""
var projectName = ""

var opts struct {
	Owner string `short:"o" long:"owner" description:"GitHub Owner/Org name" env:"GITHUB_ACTOR"`

	Repo string `short:"r" long:"repo" description:"GitHub Repo name" env:"GITHUB_REPO"`

	Type string `short:"t" long:"type" description:"The target event type to label (issues or pull_request)" env:"GITHUB_EVENT_NAME"`

	ID int `long:"id" description:"The integer id of the issue or pull request"`

	Data *string `long:"data" description:"A JSON string of the 'event' type (issue event or pull request event)'"`

	Version bool `short:"v" long:"version" description:"Display version information"`
}

const parseArgs = flags.HelpFlag | flags.PassDoubleDash

func main() {
	parser := flags.NewParser(&opts, parseArgs)
	_, err := parser.Parse()
	if err != nil {
		flagError := err.(*flags.Error)
		if flagError.Type == flags.ErrHelp {
			parser.WriteHelp(os.Stdout)
			return
		}

		if flagError.Type == flags.ErrUnknownFlag {
			_, _ = fmt.Fprintf(os.Stderr, "%s. Please use --help for available options.\n", strings.Replace(flagError.Message, "unknown", "Unknown", 1))
			return
		}
		_, _ = fmt.Fprintf(os.Stderr, "Error parsing command line options: %s\n", err)
		return
	}

	if opts.Version {
		fmt.Printf("%s %s (%s)\n", projectName, version, commit)
		return
	}

	if len(opts.Owner)+len(opts.Repo) < 2 {
		fmt.Print("Looks like owner and repo aren't valid. Please enter these required options.\n")
		return
	}

	initLogging()

	l, err := labeler.New(opts.Owner, opts.Repo, opts.Type, opts.ID, opts.Data)
	if err != nil {
		log.WithError(err).Errorf("could not initialize labeler!")
		return
	}
	err = l.Execute()
	if err != nil {
		log.WithError(err).Errorf("labeling failed!")
		return
	}
	log.Info("run complete!")
}

func initLogging() {
	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		logLevel = "error"
	}
	ll, err := log.ParseLevel(logLevel)
	if err != nil {
		ll = log.DebugLevel
	}
	log.SetLevel(ll)
	log.SetOutput(os.Stderr)
}
