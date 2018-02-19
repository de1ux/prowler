package travis

import (
	services "github.com/de1ux/prowler/services/v1"
	vcs "github.com/de1ux/prowler/vcs/v1"
)

func NewClient(config *Config) services.Client {
	return &Client{config: config}
}

type Client struct {
	config *Config
}

func (c *Client) GetStatusByPullRequest(pr *vcs.PullRequest) (*services.Status, error) {
	return nil, nil
}
