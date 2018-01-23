package bitbucket

import (
	"fmt"
	"net/http"

	"github.com/de1ux/prowler/vcs/v1"
)

const (
	baseUrl         = "https://api.bitbucket.org"
	repoUrlTemplate = baseUrl + "/2.0/repositories/%s/pullrequests"
)

func NewClient(config *Config) v1.Client {
	return &Client{config: config}
}

type Client struct {
	config *Config
}

func (c *Client) GetPullRequestsByRepo(repo string) ([]*v1.PullRequest, error) {
	_, err := http.NewRequest("GET", fmt.Sprintf(repoUrlTemplate, repo), nil)
	return nil, err
}
