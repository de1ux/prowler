package bamboo

import (
	"github.com/de1ux/prowler/services/v1"
)

func NewClient(config *Config) v1.Client {
	return &Client{config: config}
}

type Client struct {
	config *Config
}

func (c *Client) GetStatusesByPullRequest(pr *PullRequest) ([]*v1.Status, error) {
	return nil, nil
}
