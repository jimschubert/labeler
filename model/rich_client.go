package model

import (
	"context"
	"github.com/google/go-github/v50/github"
	"io"
)

type Client interface {
	// DownloadContents downloads the contents of a file from a repository. (implementation of github.RepositoriesService.DownloadContents)
	DownloadContents(ctx context.Context, owner, repo, filepath string, opts *github.RepositoryContentGetOptions) (io.ReadCloser, *github.Response, error)

	// CreateComment creates a comment on the specified issue. Specifying an issue number of 0 will create a comment on the repository.
	// (implementation of github.IssuesService.CreateComment)
	CreateComment(ctx context.Context, owner string, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)

	// AddLabelsToIssue adds labels to the specified issue. Specifying an issue number of 0 will add labels to the repository.
	// (implementation of github.IssuesService.AddLabelsToIssue)
	AddLabelsToIssue(ctx context.Context, owner string, repo string, number int, labels []string) ([]*github.Label, *github.Response, error)

	// GetIssue retrieves the specified issue. Specifying an issue number of 0 will return the repository's default issue.
	// (implementation of github.IssuesService.Get)
	GetIssue(ctx context.Context, owner string, repo string, number int) (*github.Issue, *github.Response, error)

	// GetPullRequest retrieves the specified pull request. Specifying a pull request number of 0 will return the repository's default pull request.
	// (implementation of github.PullRequestsService.Get)
	GetPullRequest(ctx context.Context, owner string, repo string, number int) (*github.PullRequest, *github.Response, error)
}

type RichClient struct {
	*github.Client
}

func (r *RichClient) DownloadContents(ctx context.Context, owner, repo, filepath string, opts *github.RepositoryContentGetOptions) (io.ReadCloser, *github.Response, error) {
	if r.Client.Repositories == nil {
		return nil, nil, nil
	}
	return r.Client.Repositories.DownloadContents(ctx, owner, repo, filepath, opts)
}

func (r *RichClient) CreateComment(ctx context.Context, owner string, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
	if r.Client.Issues == nil {
		return nil, nil, nil
	}
	return r.Client.Issues.CreateComment(ctx, owner, repo, number, comment)
}

func (r *RichClient) AddLabelsToIssue(ctx context.Context, owner string, repo string, number int, labels []string) ([]*github.Label, *github.Response, error) {
	if r.Client.Issues == nil {
		return nil, nil, nil
	}
	return r.Client.Issues.AddLabelsToIssue(ctx, owner, repo, number, labels)
}

func (r *RichClient) GetIssue(ctx context.Context, owner string, repo string, number int) (*github.Issue, *github.Response, error) {
	if r.Client.Issues == nil {
		return nil, nil, nil
	}
	return r.Client.Issues.Get(ctx, owner, repo, number)
}

func (r *RichClient) GetPullRequest(ctx context.Context, owner string, repo string, number int) (*github.PullRequest, *github.Response, error) {
	if r.Client.PullRequests == nil {
		return nil, nil, nil
	}
	return r.Client.PullRequests.Get(ctx, owner, repo, number)
}
