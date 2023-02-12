package labeler

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type args struct {
		owner string
		repo  string
		event string
		id    int
		data  *string
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		expectErrMessage string
	}{
		{
			name:             "errors on missing data",
			args:             args{},
			wantErr:          true,
			expectErrMessage: "a JSON string of event data is required",
		},
		{
			name: "constructs as expected",
			args: args{
				owner: "jimschubert",
				repo:  "example",
				event: "issue",
				id:    1,
				data:  ptr("{}"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.owner, tt.args.repo, tt.args.event, tt.args.id, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.expectErrMessage != "" {
				assert.EqualError(t, err, tt.expectErrMessage, "New() error did not match expectation")
			}

			if err == nil {
				assert.NotNil(t, got.Owner)
				assert.NotNil(t, got.Repo)
				assert.NotNil(t, got.Event)
				assert.NotNil(t, got.ID)
				assert.NotNil(t, got.context)
				assert.NotNil(t, got.client)

				assert.Equal(t, tt.args.owner, *got.Owner)
				assert.Equal(t, tt.args.repo, *got.Repo)
				assert.Equal(t, tt.args.event, *got.Event)
				assert.Equal(t, tt.args.id, *got.ID)
				assert.Equal(t, ".github/labeler.yml", got.configPath)
			}
		})
	}
}

func TestNewWithOptions(t *testing.T) {
	childContext, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		// so tooling doesn't complain about dropping a cancel func
		cancel()
	})
	type args struct {
		opts []OptFn
	}
	tests := []struct {
		name       string
		args       args
		validate   func(l *Labeler)
		errorMatch string
	}{
		// {
		// 	name:       "fails if owner and repo are both empty",
		// 	args:       args{},
		// 	errorMatch: "both a github and owner are required",
		// },

		{
			name: "fails if repo is in org/repo format",
			args: args{
				opts: []OptFn{WithRepo("jimschubert/example")},
			},
			errorMatch: "a repo must be just the repo name. Separate org/repo style into owner and repo options",
		},

		// {
		// 	name: "fails if repo defined but owner is not",
		// 	args: args{
		// 		opts: []OptFn{WithRepo("example")},
		// 	},
		// 	errorMatch: "a github owner (user or org) is required",
		// },

		{
			name: "fails if owner defined but repo is not",
			args: args{
				opts: []OptFn{WithOwner("jimschubert")},
			},
			errorMatch: "a github repo is required",
		},

		{
			name: "fails if owner,repo defined but id is not",
			args: args{
				opts: []OptFn{WithOwner("jimschubert"), WithRepo("example")},
			},
			errorMatch: "the integer id of the issue or pull request is required",
		},

		{
			name: "constructs as expected when provided all required fields",
			args: args{
				opts: []OptFn{WithOwner("jimschubert"), WithRepo("example"), WithID(1000)},
			},
			validate: func(l *Labeler) {
				assert.Equal(t, ptr("jimschubert"), l.Owner)
				assert.Equal(t, ptr("example"), l.Repo)
				assert.Equal(t, ptr(1000), l.ID)

				assert.NotNil(t, l.context, "Should have created a default context")
				assert.NotNil(t, l.client, "Should have created a default github client")
			},
		},

		{
			name: "constructs as expected when provided all fields",
			args: args{
				opts: []OptFn{
					// required fields
					WithOwner("jimschubert"), WithRepo("example"), WithID(1000),

					// optional fields
					WithContext(childContext), WithConfigPath(".github/labeler-custom.yml"), WithData("{}"), WithToken("irrelevant"),
				},
			},
			validate: func(l *Labeler) {
				assert.Equal(t, ptr("jimschubert"), l.Owner)
				assert.Equal(t, ptr("example"), l.Repo)
				assert.Equal(t, ptr("{}"), l.Data)
				assert.Equal(t, ptr(1000), l.ID)

				assert.NotNil(t, l.context, "Should have created a default context")
				assert.NotNil(t, l.client, "Should have created a default github client")

				assert.Equal(t, ".github/labeler-custom.yml", l.configPath)
				assert.Equal(t, &childContext, l.context)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWithOptions(tt.args.opts...)
			wantError := tt.errorMatch != ""
			if wantError {
				assert.EqualError(t, err, tt.errorMatch, fmt.Sprintf("NewWithOptions(%v)", tt.args.opts))
				return
			}

			if err != nil && tt.errorMatch != "" {
				t.Errorf("New() error = %v, no error expectation was defined", err)
				return
			}

			if tt.validate != nil {
				tt.validate(got)
			}
		})
	}
}
