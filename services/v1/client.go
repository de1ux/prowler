package v1

import (
	"github.com/de1ux/prowler/vcs/v1"
)

type State int

const (
	Passing State = iota
	Pending
	Failed
	Errored
)

type Status struct {
	Name  string // Bamboo, Jenkins, Smithy etc
	URL   string // URL to the webhook page
	State State  // The Service specific state coralled into a State enum by the implementer
}

type Client interface {
	GetStatusByPullRequest(pr *v1.PullRequest) (*Status, error)
}
