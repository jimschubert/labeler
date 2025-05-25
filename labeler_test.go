package labeler

import (
	"bytes"
	"context"
	"errors"
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

func (m *mockConfig) LabelsFor(text ...string) []string {
	texts := make([]interface{}, len(text))
	for i, v := range text {
		texts[i] = v
	}
	args := m.Called(texts...)
	return args.Get(0).([]string)
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
	mockCfg.On("LabelsFor", "title", "body").Return([]string{"bug"})
	mockClient.On("AddLabelsToIssue", mock.Anything, "owner", "repo", 1, []string{"bug"}).
		Return([]*github.Label{{Name: ptr("bug")}}, nil, nil)
	ev := &testEvent{title: "title", body: "body"}
	count := l.applyLabels(ev, []*github.Label{})
	assert.Equal(t, 1, count)
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
	mockClient.On("GetPullRequest", mock.Anything, "owner", "repo", 1).
		Return(&github.Issue{Title: ptr("t"), Body: ptr("b")}, nil)
	mockClient.On("GetIssue", mock.Anything, "owner", "repo", 1).Return(&github.Issue{Title: ptr("t"), Body: ptr("b")}, nil, nil)

	issue, err := l.getIssue()
	assert.NoError(t, err)
	assert.Equal(t, "t", *issue.Title)
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
	mockCfg.On("LabelsFor", mock.Anything, mock.Anything).Return([]string{})
	mockClient.On("GetIssue", mock.Anything, "owner", "repo", 1).Return(&github.Issue{Title: ptr("t"), Body: ptr("b")}, nil, nil)
	err := l.Execute()
	assert.NoError(t, err)
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
	mockCfg.On("LabelsFor", mock.Anything, mock.Anything).Return([]string{})
	mockClient.On("GetIssue", mock.Anything, "owner", "repo", 1).Return(&github.Issue{Title: ptr("t"), Body: ptr("b")}, nil, nil)
	err := l.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse \".github/labeler.yml\"")
}

type testEvent struct{ title, body string }

func (t *testEvent) GetTitle() string { return t.title }
func (t *testEvent) GetBody() string  { return t.body }
