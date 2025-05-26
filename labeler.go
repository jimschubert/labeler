package labeler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"regexp"
	"time"

	"github.com/google/go-github/v50/github"
	"github.com/jimschubert/labeler/model"
	log "github.com/sirupsen/logrus"
)

var (
	issue             = "issues"
	pullRequest       = "pull_request"
	pullRequestTarget = "pull_request_target"
)

type githubEvent interface {
	GetTitle() string
	GetBody() string
}

// Labeler is the container for the application entrypoint's logic
type Labeler struct {
	Owner      *string
	Repo       *string
	Event      *string
	Data       *string
	ID         *int
	context    *context.Context
	client     model.Client
	config     model.Config
	configPath string
}

// Execute performs the labeler logic
func (l *Labeler) Execute() error {
	err := l.checkPreconditions()
	if err != nil {
		return err
	}

	log.Debugf("executing with owner=%s repo=%s event=%s", *l.Owner, *l.Repo, *l.Event)

	c, err := l.retrieveConfig()
	if err != nil {
		return err
	}
	l.config = c

	switch *l.Event {
	case issue:
		err = l.processIssue()
		if err != nil {
			return err
		}
	case pullRequestTarget, pullRequest:
		err = l.processPullRequest()
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Labeler) retrieveConfig() (model.Config, error) {
	if l.configPath == "" {
		return nil, errors.New("the labeler configuration path can not be empty")
	}
	ctx, cancel := context.WithTimeout(*l.context, 10*time.Second)
	defer cancel()
	r, _, err := l.client.DownloadContents(ctx, *l.Owner, *l.Repo, l.configPath, &github.RepositoryContentGetOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = r.Close()
	}()

	bytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %w", l.configPath, err)
	}

	var c model.Config
	c = &model.FullConfig{}
	err = c.FromBytes(bytes)
	if err != nil {
		c = &model.SimpleConfig{}
		err = c.FromBytes(bytes)
		if err != nil {
			return nil, fmt.Errorf("could not parse %q", l.configPath)
		}
	}
	log.WithFields(log.Fields{l.configPath: c}).Debugf("Parsed %q", l.configPath)
	return c, nil
}

func (l *Labeler) checkPreconditions() error {
	if len(*l.Owner) <= 1 {
		return errors.New("owner is invalid")
	}
	if len(*l.Repo) <= 1 {
		return errors.New("repo is invalid")
	}
	if *l.Event != issue && *l.Event != pullRequest && *l.Event != pullRequestTarget {
		return fmt.Errorf("event must be one of [ %s , %s , %s ]", issue, pullRequest, pullRequestTarget)
	}

	return nil
}

// noinspection GoNilness
func (l *Labeler) processIssue() error {
	issue, err := l.getIssue()
	if err != nil {
		return err
	}

	existingLabels := issue.Labels
	count := l.applyLabels(issue, existingLabels)
	if count > 0 {
		var comment *string
		switch v := (l.config).(type) {
		case *model.FullConfig:
			if v != nil && v.Comments != nil {
				comment = v.Comments.Issues
			}
		case *model.SimpleConfig:
			if v != nil {
				comment = &v.Comment
			}
		}

		err := l.addComment(comment)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Labeler) processPullRequest() error {
	pr, err := l.getPullRequest()
	if err != nil {
		return err
	}

	existingLabels := make([]*github.Label, 0)
	existingLabels = append(existingLabels, pr.Labels...)

	count := l.applyLabels(pr, existingLabels)
	if count > 0 {
		var comment *string
		switch v := (l.config).(type) {
		case *model.FullConfig:
			if v != nil && v.Comments != nil {
				comment = v.Comments.PullRequests
			}
		case *model.SimpleConfig:
			if v != nil {
				comment = &v.Comment
			}
		}

		err := l.addComment(comment)
		if err != nil {
			return err
		}
	}

	return nil
}

func labelExists(s []*github.Label, name *string) bool {
	if name != nil {
		for _, a := range s {
			if *a.Name == *name {
				return true
			}
		}
	}
	return false
}

func (l *Labeler) addComment(comment *string) error {
	if comment != nil && len(*comment) > 0 {
		ctx, cancel := context.WithTimeout(*l.context, 10*time.Second)
		defer cancel()

		issueComment := &github.IssueComment{
			Body: newComment(*comment),
		}
		_, _, err := l.client.CreateComment(ctx, *l.Owner, *l.Repo, *l.ID, issueComment)
		return err
	}
	return nil
}

func newComment(comment string) *string {
	fullComment := fmt.Sprintf("<!-- Labeler (https://github.com/jimschubert/labeler) -->\n%s", comment)
	return &fullComment
}

func (l *Labeler) applyLabels(i githubEvent, existingLabels []*github.Label) int {
	targetBranch := ""
	if pr, ok := i.(*github.PullRequest); ok && pr != nil && pr.Base != nil && pr.Base.Ref != nil {
		targetBranch = *pr.Base.Ref
	}

	labels := l.config.LabelsFor(i.GetTitle(), i.GetBody())
	filteredLabels := make(map[string]model.Label)
	for name, label := range labels {
		if len(label.Branches) > 0 && targetBranch != "" {
			for _, branch := range label.Branches {
				re := regexp.MustCompile(branch)
				if re.Match([]byte(targetBranch)) {
					filteredLabels[name] = label
					break
				}
			}
		} else if len(label.Branches) == 0 {
			filteredLabels[name] = label
		}
	}

	hasNew := false

	for name := range filteredLabels {
		if hasNew {
			break
		}
		hasNew = !labelExists(existingLabels, &name)
	}

	if hasNew {
		ctx, cancel := context.WithTimeout(*l.context, 10*time.Second)
		defer cancel()

		newLabels := make([]string, 0, len(filteredLabels))
		for m := range maps.Keys(filteredLabels) {
			newLabels = append(newLabels, m)
		}
		added, _, err := l.client.AddLabelsToIssue(ctx, *l.Owner, *l.Repo, *l.ID, newLabels)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Debug("Unable to add labels to issue.")
			return 0
		}

		num := len(added)
		log.Debugf("Found %d new labels to apply", num)
		return num
	} else {
		log.Debug("Found 0 labels to apply")
	}

	return 0
}

func (l *Labeler) getPullRequest() (*github.PullRequest, error) {
	var pr *github.PullRequest
	if l.Data != nil {
		var pre *github.PullRequestEvent = nil
		b := []byte(*l.Data)
		err := json.Unmarshal(b, &pre)
		if err != nil {
			err = json.Unmarshal(b, &pr)
			if err != nil {
				return nil, err
			}
		} else {
			//noinspection GoNilness
			pr = pre.GetPullRequest()
		}
	} else {
		ctx, cancel := context.WithTimeout(*l.context, 10*time.Second)
		defer cancel()
		pull, _, err := l.client.GetPullRequest(ctx, *l.Owner, *l.Repo, *l.ID)
		if err != nil {
			return nil, err
		}
		pr = pull
	}
	return pr, nil
}

func (l *Labeler) getIssue() (*github.Issue, error) {
	var i *github.Issue
	if l.Data != nil {
		var iss *github.IssuesEvent = nil
		b := []byte(*l.Data)
		err := json.Unmarshal(b, &iss)
		if err != nil {
			err = json.Unmarshal(b, &i)
			if err != nil {
				return nil, err
			}
		} else {
			//noinspection GoNilness
			i = iss.GetIssue()
		}
	} else {
		ctx, cancel := context.WithTimeout(*l.context, 10*time.Second)
		defer cancel()
		issue, _, err := l.client.GetIssue(ctx, *l.Owner, *l.Repo, *l.ID)
		if err != nil {
			return nil, err
		}
		i = issue
	}
	return i, nil
}
