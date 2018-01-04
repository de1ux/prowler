package clientv1

type State int

const (
	passing State = iota
	pending
	failed
)

type Client struct {
	// TODO - also need to be common properties across SCM providers (GitHub, Bamboo)
}

type PullRequest struct {
	URL       string // URL to the PR
	Conflicts *bool  // whether the PR has conflicts; nullable to support *special* Bitbucket
}

type ServiceState struct {
	Name  string // Bamboo, Jenkins, Smithy etc
	URL   string // URL to the webhook page
	State State  // The Service specific state coralled into a State enum by the implementer
}

func (c *Client) getPullRequestsByRepo()

func (c *Client) getServiceStatesByPullRequest(pr *PullRequest) {}
