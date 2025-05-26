package labeler

import (
	"context"
	"errors"
	"github.com/jimschubert/labeler/model"
	"os"
	"strings"

	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

var skipTokenCheck = false

// Opt is a group of options for constructing a new Labeler
type Opt struct {
	token      string
	ctx        context.Context
	client     *github.Client
	owner      string
	repo       string
	event      string
	id         int
	data       string
	configPath string
}

type OptFn func(o *Opt)

// WithToken allows configuration of the github token required for accessing the GitHub API for a given org/repo.
//
//goland:noinspection GoUnusedExportedFunction
func WithToken(value string) OptFn {
	return func(o *Opt) {
		o.token = value
	}
}

// WithContext allows configuration of the context used as a parent context for all GitHub API calls
func WithContext(ctx context.Context) OptFn {
	return func(o *Opt) {
		o.ctx = ctx
	}
}

// WithClient allows for configuration of the github client
//
//goland:noinspection GoUnusedExportedFunction
func WithClient(client *github.Client) OptFn {
	return func(o *Opt) {
		o.client = client
	}
}

// WithOwner allows for configuring the user or organization owning the target repo
func WithOwner(value string) OptFn {
	return func(o *Opt) {
		o.owner = value
	}
}

// WithRepo allows for configuring the name of the repo under a given owner; use in conjunction with WithOwner
func WithRepo(value string) OptFn {
	return func(o *Opt) {
		o.repo = value
	}
}

// WithEvent allows for configuring the target event type
func WithEvent(value string) OptFn {
	return func(o *Opt) {
		o.event = value
	}
}

// WithID allows for configuring the identifier of the issue or pull request we want to label
func WithID(value int) OptFn {
	return func(o *Opt) {
		o.id = value
	}
}

// WithData allows for configuring the JSON event data to apply a label
func WithData(value string) OptFn {
	return func(o *Opt) {
		o.data = value
	}
}

// WithConfigPath allows for configuring the labeler config path relative to the repository root (usually .github/labeler.yml)
func WithConfigPath(value string) OptFn {
	return func(o *Opt) {
		o.configPath = value
	}
}

// NewWithOptions constructs a new Labeler with functional arguments of type OptFn
func NewWithOptions(opts ...OptFn) (*Labeler, error) {
	l := Labeler{}
	options := Opt{
		token: os.Getenv("GITHUB_TOKEN"),
		owner: os.Getenv("GITHUB_ACTOR"),
		repo:  os.Getenv("GITHUB_REPO"),
		event: os.Getenv("GITHUB_EVENT_NAME"),
		id:    -1,
	}

	for _, opt := range opts {
		opt(&options)
	}

	// validations
	if options.owner == "" && options.repo == "" {
		return nil, errors.New("both a github and owner are required")
	}

	if strings.Contains(options.repo, "/") {
		return nil, errors.New("a repo must be just the repo name. Separate org/repo style into owner and repo options")
	}

	if options.owner == "" {
		return nil, errors.New("a github owner (user or org) is required")
	}

	if options.repo == "" {
		return nil, errors.New("a github repo is required")
	}

	if options.id < 0 {
		return nil, errors.New("the integer id of the issue or pull request is required")
	}

	if options.ctx == nil {
		options.ctx = context.Background()
	}

	if options.client == nil {
		// only validate the token when constructing this default client. Otherwise, assume the caller has property constructed a client
		if options.token == "" && !skipTokenCheck {
			return nil, errors.New("github token (e.g. GITHUB_TOKEN environment variable) is required")
		}

		options.client = github.NewClient(oauth2.NewClient(options.ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: options.token},
		)))
	}

	if options.configPath == "" {
		options.configPath = ".github/labeler.yml"
	}

	// assignment
	l.context = &options.ctx
	l.client = &model.RichClient{Client: options.client}
	l.Owner = &options.owner
	l.Repo = &options.repo
	l.Event = &options.event
	l.ID = &options.id
	if options.data != "" {
		l.Data = &options.data
	}
	l.configPath = options.configPath

	return &l, nil
}

// New creates a new instance of a Labeler
func New(owner string, repo string, event string, id int, data *string) (*Labeler, error) {
	if data == nil {
		return nil, errors.New("a JSON string of event data is required")
	}
	return NewWithOptions(
		WithOwner(owner),
		WithRepo(repo),
		WithEvent(event),
		WithID(id),
		WithData(*data),
	)
}
