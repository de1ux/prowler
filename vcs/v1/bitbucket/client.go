package bitbucket

import (
	"github.com/de1ux/prowler/vcs/v1"
)

func NewClient(config *Config) v1.Client {
	return &Client{config: config}
}

type Client struct {
	config *Config
}

func (c *Client) GetPullRequestsByRepo(repo string) ([]*v1.PullRequest, error) {
	return nil, nil
}
