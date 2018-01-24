package bitbucket

import (
	"encoding/base64"
	"encoding/json"
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
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf(repoUrlTemplate, repo), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", c.basicAuth()))
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var payload = struct {
		Values []struct {
			Author struct {
				Username string
			}
			Title string
			Links struct {
				Html struct {
					Href string
				}
			}
		}
	}{}

	if err = json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}

	prs := make([]*v1.PullRequest, 0)
	for _, item := range payload.Values {
		if !c.config.showAllPrs && item.Author.Username != c.config.username {
			continue
		}

		prs = append(prs, &v1.PullRequest{
			Title:     item.Title,
			URL:       item.Links.Html.Href,
			Conflicts: nil,
		})
	}

	return prs, nil
}

func (c *Client) basicAuth() string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.config.username, c.config.token)))
}
