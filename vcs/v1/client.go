package v1

type PullRequest struct {
	URL       string // URL to the PR
	Conflicts *bool  // whether the PR has conflicts; nullable to support *special* Bitbucket
}

type Client interface {
	GetPullRequestsByRepo(repo string) ([]*PullRequest, error)
}
