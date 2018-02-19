package github

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/de1ux/prowler/vcs/v1"
)

const (
	baseUrl         = "https://api.github.com"
	repoUrlTemplate = baseUrl + "/repos/%s/pulls"
)

type prMeta struct {
	URL   string `json:"url"`
	Link  string `json:"html_url"`
	Title string `json:"title"`
	User  struct {
		Login string `json:"login"`
	} `json:"user"` // TODO(benchmark): use lazy json.RawMessage here and check for substring existance
	Stats string `json:"statuses_url"`
	Merge bool   `json:"mergeable"`
}

func NewClient(config *Config) v1.Client {
	return &Client{config: config}
}

type Client struct {
	config *Config
}

func (c *Client) GetPullRequestsByRepo(repo string) ([]*v1.PullRequest, error) {
	res, err := http.Get(fmt.Sprintf("%s?oauth_token=%s", fmt.Sprintf(repoUrlTemplate, repo), c.config.token))
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	prs := make([]prMeta, 0)
	if err = json.NewDecoder(res.Body).Decode(&prs); err != nil {
		return nil, err
	}

	return nil, nil
}
