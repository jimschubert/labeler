package labeler

import (
	"bytes"
	"context"
	"errors"
	"github.com/jimschubert/labeler/model"
	"io"
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type mockConfig struct {
	mock.Mock
}

func (m *mockConfig) LabelsFor(text ...string) map[string]model.Label {
	arguments := make([]interface{}, 0)
	for _, v := range text {
		arguments = append(arguments, v)
	}
	args := m.Called(arguments...)
	return args.Get(0).(map[string]model.Label)
}

func (m *mockConfig) FromBytes(b []byte) error {
	args := m.Called(b)
	return args.Error(0)
}

type mockRichClient struct {
	mock.Mock
}

func (m *mockRichClient) DownloadContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (io.ReadCloser, *github.Response, error) {
	args := m.Called(ctx, owner, repo, path, opts)
	return args.Get(0).(io.ReadCloser), nil, args.Error(2)
}

func (m *mockRichClient) AddLabelsToIssue(ctx context.Context, owner, repo string, number int, labels []string) ([]*github.Label, *github.Response, error) {
	args := m.Called(ctx, owner, repo, number, labels)
	return args.Get(0).([]*github.Label), nil, args.Error(2)
}

func (m *mockRichClient) CreateComment(ctx context.Context, owner, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
	args := m.Called(ctx, owner, repo, number, comment)
	return nil, nil, args.Error(2)
}

func (m *mockRichClient) GetIssue(ctx context.Context, owner, repo string, number int) (*github.Issue, *github.Response, error) {
	args := m.Called(ctx, owner, repo, number)
	return args.Get(0).(*github.Issue), nil, args.Error(2)
}

func (m *mockRichClient) GetPullRequest(ctx context.Context, owner, repo string, number int) (*github.PullRequest, *github.Response, error) {
	args := m.Called(ctx, owner, repo, number)
	return args.Get(0).(*github.PullRequest), nil, args.Error(2)
}

func TestLabeler_checkPreconditions(t *testing.T) {
	l := &Labeler{
		Owner: ptr("o"),
		Repo:  ptr("r"),
		Event: ptr("issues"),
	}
	assert.Error(t, l.checkPreconditions(), "owner too short")
	l.Owner = ptr("owner")
	assert.Error(t, l.checkPreconditions(), "repo too short")
	l.Repo = ptr("repo")
	l.Event = ptr("bad")
	assert.Error(t, l.checkPreconditions(), "bad event")
	l.Event = ptr("issues")
	assert.NoError(t, l.checkPreconditions())
}

func TestLabeler_retrieveConfig_success(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRichClient)

	mockCfg := new(mockConfig)
	l := &Labeler{
		Owner:      ptr("owner"),
		Repo:       ptr("repo"),
		Event:      ptr("issues"),
		context:    &ctx,
		client:     mockClient,
		configPath: ".github/labeler.yml",
	}

	mockClient.On("DownloadContents", mock.Anything, "owner", "repo", ".github/labeler.yml", mock.Anything).
		Return(io.NopCloser(bytes.NewReader([]byte(
			`comment: Labels
labels:
  'bug':
    - '\bbug[s]?\b'
  'duplicate':
    - '\bduplicate\b'
    - '\bdupe\b'
  'question':
    - '\bquestion\b'
`))), nil, nil)

	mockCfg.On("FromBytes", mock.Anything).Return(nil)

	_, err := l.retrieveConfig()
	assert.NoError(t, err)
}

func TestLabeler_retrieveConfig_error(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRichClient)

	l := &Labeler{
		Owner:      ptr("owner"),
		Repo:       ptr("repo"),
		Event:      ptr("issues"),
		context:    &ctx,
		client:     mockClient,
		configPath: ".github/labeler.yml",
	}
	mockClient.On("DownloadContents", mock.Anything, "owner", "repo", ".github/labeler.yml", mock.Anything).
		Return(io.NopCloser(nil), nil, errors.New("fail"))
	_, err := l.retrieveConfig()
	assert.Error(t, err)
}

func TestLabeler_applyLabels(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRichClient)

	mockCfg := new(mockConfig)
	l := &Labeler{
		Owner:   ptr("owner"),
		Repo:    ptr("repo"),
		ID:      ptr(1),
		context: &ctx,
		client:  mockClient,
		config:  mockCfg,
	}
	mockCfg.On("LabelsFor", "title", "body").Return(map[string]model.Label{
		"bug": {},
	})
	mockClient.On("AddLabelsToIssue", mock.Anything, "owner", "repo", 1, []string{"bug"}).
		Return([]*github.Label{{Name: ptr("bug")}}, nil, nil)
	ev := &testEvent{title: "title", body: "body"}
	count := l.applyLabels(ev, []*github.Label{})
	assert.Equal(t, 1, count)

	mockClient.AssertNumberOfCalls(t, "AddLabelsToIssue", 1)
	mockClient.AssertExpectations(t)
}

func TestLabeler_applyLabelsCustomizedFields(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRichClient)

	mockCfg := new(mockConfig)
	l := &Labeler{
		Owner:     ptr("owner"),
		Repo:      ptr("repo"),
		ID:        ptr(1),
		fieldFlag: FieldTitle,
		context:   &ctx,
		client:    mockClient,
		config:    mockCfg,
	}
	mockCfg.On("LabelsFor", "title").Return(map[string]model.Label{
		"bug": {},
	})
	mockClient.On("AddLabelsToIssue", mock.Anything, "owner", "repo", 1, []string{"bug"}).
		Return([]*github.Label{{Name: ptr("bug")}}, nil, nil)
	ev := &testEvent{title: "title", body: "body"}
	count := l.applyLabels(ev, []*github.Label{})
	assert.Equal(t, 1, count)

	mockClient.AssertNumberOfCalls(t, "AddLabelsToIssue", 1)
	mockClient.AssertExpectations(t)
}

func TestLabeler_addComment(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRichClient)
	l := &Labeler{
		Owner:   ptr("owner"),
		Repo:    ptr("repo"),
		ID:      ptr(1),
		context: &ctx,
		client:  mockClient,
	}
	comment := "hello"
	mockClient.On("CreateComment", mock.Anything, "owner", "repo", 1, mock.Anything).
		Return(nil, nil, nil)
	err := l.addComment(&comment)
	assert.NoError(t, err)
}

func TestLabeler_getIssue_withData(t *testing.T) {
	mockClient := new(mockRichClient)
	l := &Labeler{
		Data:   ptr(`{"title":"t","body":"b", "issue": {"number": 1}}`),
		client: mockClient,
	}
	issue, err := l.getIssue()
	assert.NoError(t, err)
	assert.NotNil(t, issue)
	assert.Equal(t, 1, issue.GetNumber())
}

func TestLabeler_getIssue_fromAPI(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRichClient)
	l := &Labeler{
		Owner:   ptr("owner"),
		Repo:    ptr("repo"),
		ID:      ptr(1),
		context: &ctx,
		client:  mockClient,
	}
	mockClient.On("GetIssue", mock.Anything, "owner", "repo", 1).Return(&github.Issue{Title: ptr("t"), Body: ptr("b")}, nil, nil)

	issue, err := l.getIssue()
	assert.NoError(t, err)
	assert.Equal(t, "t", *issue.Title)

	mockClient.AssertNumberOfCalls(t, "GetIssue", 1)
	mockClient.AssertExpectations(t)
}

func TestLabeler_getPullRequest_withData(t *testing.T) {
	l := &Labeler{
		Data: ptr(`{"title":"t","body":"b", "pull_request": {"number": 1}}`),
	}
	pr, err := l.getPullRequest()
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 1, pr.GetNumber(), "should have number 1 from data")
}

func TestLabeler_getPullRequest_fromAPI(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRichClient)
	l := &Labeler{
		Owner:   ptr("owner"),
		Repo:    ptr("repo"),
		ID:      ptr(1),
		context: &ctx,
		client:  mockClient,
	}
	mockClient.On("GetPullRequest", mock.Anything, "owner", "repo", 1).
		Return(&github.PullRequest{Title: ptr("t"), Body: ptr("b")}, nil, nil)
	pr, err := l.getPullRequest()
	assert.NoError(t, err)
	assert.Equal(t, "t", *pr.Title)

	mockClient.AssertNumberOfCalls(t, "GetPullRequest", 1)
	mockClient.AssertExpectations(t)
}

func TestLabeler_Execute_errors(t *testing.T) {
	l := &Labeler{
		Owner: ptr("o"),
		Repo:  ptr("r"),
		Event: ptr("bad"),
	}
	err := l.Execute()
	assert.Error(t, err)
}

func TestLabeler_Execute_success_issue(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRichClient)

	l := &Labeler{
		Owner:      ptr("owner"),
		Repo:       ptr("repo"),
		Event:      ptr("issues"),
		ID:         ptr(1),
		context:    &ctx,
		client:     mockClient,
		configPath: ".github/labeler.yml",
	}
	mockClient.On("DownloadContents", mock.Anything, "owner", "repo", ".github/labeler.yml", mock.Anything).
		Return(io.NopCloser(bytes.NewReader([]byte(
			`comment: Labels
labels:
  'bug':
    - '\bbug[s]?\b'
  'duplicate':
    - '\bduplicate\b'
    - '\bdupe\b'
  'question':
    - '\bquestion\b'
`))), nil, nil)
	mockClient.On("GetIssue", mock.Anything, "owner", "repo", 1).Return(&github.Issue{Title: ptr("title"), Body: ptr("body")}, nil, nil)
	err := l.Execute()
	assert.NoError(t, err)

	mockClient.AssertNumberOfCalls(t, "GetIssue", 1)
	mockClient.AssertExpectations(t)
}

func TestLabeler_Execute_allows_config_override_fields_title(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRichClient)

	l := &Labeler{
		Owner:      ptr("owner"),
		Repo:       ptr("repo"),
		Event:      ptr("issues"),
		ID:         ptr(1),
		context:    &ctx,
		client:     mockClient,
		configPath: ".github/labeler.yml",
		fieldFlag:  AllFieldFlags,
	}
	mockClient.On("DownloadContents", mock.Anything, "owner", "repo", ".github/labeler.yml", mock.Anything).
		Return(io.NopCloser(bytes.NewReader([]byte(
			`%YAML 1.1
---
enable:
  issues: true
  prs: false

fields:
  - title

labels:
  'bug':
    include:
      - '\btitle[s]?\b'
    exclude: []
  'help wanted':
    include:
      - '\bbody( me)?\b'
`))), nil, nil)
	mockClient.On("GetIssue", mock.Anything, "owner", "repo", 1).Return(&github.Issue{Title: ptr("title"), Body: ptr("body")}, nil, nil)
	mockClient.On("AddLabelsToIssue", mock.Anything, "owner", "repo", 1, []string{"bug"}).Return([]*github.Label{{Name: ptr("bug")}}, nil, nil)
	err := l.Execute()
	assert.NoError(t, err)

	mockClient.AssertNumberOfCalls(t, "GetIssue", 1)
	mockClient.AssertCalled(t, "AddLabelsToIssue", mock.Anything, "owner", "repo", 1, []string{"bug"})
	mockClient.AssertNumberOfCalls(t, "AddLabelsToIssue", 1)
	mockClient.AssertExpectations(t)
}

func TestLabeler_Execute_allows_config_override_fields_body(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRichClient)

	mockCfg := new(mockConfig)
	l := &Labeler{
		Owner:      ptr("owner"),
		Repo:       ptr("repo"),
		Event:      ptr("issues"),
		ID:         ptr(1),
		context:    &ctx,
		client:     mockClient,
		configPath: ".github/labeler.yml",
		fieldFlag:  AllFieldFlags,
	}
	mockClient.On("DownloadContents", mock.Anything, "owner", "repo", ".github/labeler.yml", mock.Anything).
		Return(io.NopCloser(bytes.NewReader([]byte(
			`%YAML 1.1
---
enable:
  issues: true
  prs: false

fields:
  - body

labels:
  'bug':
    include:
      - '\btitle[s]?\b'
    exclude: []
  'help wanted':
    include:
      - '\bbody\b'

`))), nil, nil)
	mockCfg.On("FromBytes", mock.Anything).Return(nil)
	mockCfg.On("LabelsFor", "body").Return(map[string]model.Label{})
	mockClient.On("GetIssue", mock.Anything, "owner", "repo", 1).Return(&github.Issue{Title: ptr("title"), Body: ptr("body")}, nil, nil)
	mockClient.On("AddLabelsToIssue", mock.Anything, "owner", "repo", 1, []string{"help wanted"}).Return([]*github.Label{{Name: ptr("help wanted")}}, nil, nil)
	err := l.Execute()
	assert.NoError(t, err)

	mockClient.AssertNumberOfCalls(t, "GetIssue", 1)
	mockClient.AssertCalled(t, "AddLabelsToIssue", mock.Anything, "owner", "repo", 1, []string{"help wanted"})
	mockClient.AssertNumberOfCalls(t, "AddLabelsToIssue", 1)
	mockClient.AssertExpectations(t)
}

func TestLabeler_Execute_targeted_branch(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRichClient)

	l := &Labeler{
		Owner:      ptr("owner"),
		Repo:       ptr("repo"),
		Event:      ptr("pull_request_target"),
		ID:         ptr(1),
		context:    &ctx,
		client:     mockClient,
		configPath: ".github/labeler.yml",
	}
	mockClient.On("DownloadContents", mock.Anything, "owner", "repo", ".github/labeler.yml", mock.Anything).
		Return(io.NopCloser(bytes.NewReader([]byte(
			`enable:
  issues: false
  prs: true

comments:
  prs: |
    I applied labels to your pull request.

    Please review the labels.

labels:
  'bug':
    include:
      - '\bbug[s]?\b'
    exclude: []
    branches:
      - main
      - develop
  'deploy':
    include:
      - '\bJIRA-\d{1,}\b'
    branches:
      - production
  'help wanted':
    include:
      - '\bhelp( me)?\b'
    exclude:
      - '\b\[test(ing)?\]\b'
  'enhancement':
    include:
      - '\bfeat\b'
    exclude: []

`))), nil, nil)
	mockClient.On("GetPullRequest", mock.Anything, "owner", "repo", 1).
		Return(&github.PullRequest{Title: ptr("JIRA-1234 DO IT"), Body: ptr("b"), Base: &github.PullRequestBranch{Ref: ptr("production")}}, nil, nil)

	mockClient.On("AddLabelsToIssue", mock.Anything, "owner", "repo", 1, []string{"deploy"}).
		Return([]*github.Label{{Name: ptr("deploy")}}, nil, nil)

	mockClient.On("CreateComment", mock.Anything, "owner", "repo", 1, mock.Anything).Return(nil, nil, nil)

	err := l.Execute()
	assert.NoError(t, err)
	mockClient.AssertNumberOfCalls(t, "AddLabelsToIssue", 1)
	mockClient.AssertNumberOfCalls(t, "CreateComment", 1)
	mockClient.AssertExpectations(t)
}

func TestLabeler_Execute_targeted_branch_regex(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRichClient)

	l := &Labeler{
		Owner:      ptr("owner"),
		Repo:       ptr("repo"),
		Event:      ptr("pull_request_target"),
		ID:         ptr(1),
		context:    &ctx,
		client:     mockClient,
		configPath: ".github/labeler.yml",
	}
	mockClient.On("DownloadContents", mock.Anything, "owner", "repo", ".github/labeler.yml", mock.Anything).
		Return(io.NopCloser(bytes.NewReader([]byte(
			`enable:
  issues: false
  prs: true

comments:
  prs: |
    I applied labels to your pull request.

    Please review the labels.

labels:
  'bug':
    include:
      - '\bbug[s]?\b'
    exclude: []
    branches:
      - main
      - develop
      - feature\/yes
  'deploy':
    include:
      - '\bJIRA-\d{1,}\b'
    branches:
      - production
  'help wanted':
    include:
      - '\bhelp( me)?\b'
    exclude:
      - '\b\[test(ing)?\]\b'
  'enhancement':
    include:
      - '\bfeat\b'
    exclude: []
    branches:
      - feature\/.+

`))), nil, nil)
	mockClient.On("GetPullRequest", mock.Anything, "owner", "repo", 1).
		Return(&github.PullRequest{Title: ptr("feat: wakka wakka"), Body: ptr("some enhancement or bug"), Base: &github.PullRequestBranch{Ref: ptr("feature/yes")}}, nil, nil)

	mockClient.On("AddLabelsToIssue", mock.Anything, "owner", "repo", 1, mock.Anything).
		Return([]*github.Label{{Name: ptr("bug")}, {Name: ptr("enhancement")}}, nil, nil)

	mockClient.On("CreateComment", mock.Anything, "owner", "repo", 1, mock.Anything).Return(nil, nil, nil)

	err := l.Execute()
	assert.NoError(t, err)
	mockClient.AssertNumberOfCalls(t, "AddLabelsToIssue", 1)
	mockClient.AssertNumberOfCalls(t, "CreateComment", 1)
	mockClient.AssertExpectations(t)
}

func TestLabeler_Execute_fail_to_parse_config(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockRichClient)

	mockCfg := new(mockConfig)
	l := &Labeler{
		Owner:      ptr("owner"),
		Repo:       ptr("repo"),
		Event:      ptr("issues"),
		ID:         ptr(1),
		context:    &ctx,
		client:     mockClient,
		configPath: ".github/labeler.yml",
	}
	mockClient.On("DownloadContents", mock.Anything, "owner", "repo", ".github/labeler.yml", mock.Anything).
		Return(io.NopCloser(bytes.NewReader([]byte(`bananas`))), nil, nil)
	mockCfg.On("FromBytes", mock.Anything).Return(nil)
	mockCfg.On("LabelsFor", mock.Anything, mock.Anything).Return(map[string]model.Label{})
	err := l.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse \".github/labeler.yml\"")

	mockClient.AssertNumberOfCalls(t, "DownloadContents", 1)
	mockClient.AssertExpectations(t)
}

type testEvent struct{ title, body string }

func (t *testEvent) GetTitle() string { return t.title }
func (t *testEvent) GetBody() string  { return t.body }
