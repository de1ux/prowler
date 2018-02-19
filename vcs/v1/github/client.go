package github

import (
	"fmt"
	"net/http"

	"github.com/de1ux/prowler/vcs/v1"
)

const (
	baseUrl         = "https://api.github.com"
	repoUrlTemplate = baseUrl + "/repos/%s/pulls"
)

func NewClient(config *Config) v1.Client {
	return &Client{config: config}
}

type Client struct {
	config *Config
}

func (c *Client) GetPullRequestsByRepo(repo string) ([]*v1.PullRequest, error) {
	_, err := http.Get(fmt.Sprintf("%s?oauth_token=", fmt.Sprintf(repoUrlTemplate, repo), c.config.token))
	if err != nil {
		return nil, err
	}
	return nil, nil
}
