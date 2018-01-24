package bamboo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	services "github.com/de1ux/prowler/services/v1"
	vcs "github.com/de1ux/prowler/vcs/v1"
)

var jiraRegex = regexp.MustCompile("([A-z]+-[0-9]+)")
var bambooStateToState = map[string]services.State{
	"Successful": services.Passing,
	"Failed":     services.Failed,
}

const (
	baseUrl             = "%s/rest/api"
	searchIssueTemplate = baseUrl + "/latest/quicksearch.json?searchTerm=%s"
	searchPlansTemplate = baseUrl + "/latest/result/%s.json"
)

func NewClient(config *Config) services.Client {
	return &Client{config: config}
}

type Client struct {
	config *Config
}

func (c *Client) GetStatusByPullRequest(pr *vcs.PullRequest) (*services.Status, error) {
	// Try to parse a JIRA ticket out of the PR title
	matches := jiraRegex.FindAllString(pr.Title, 1)
	if matches == nil || len(matches) < 1 {
		return nil, nil
	}
	jira := matches[0]

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf(searchIssueTemplate, c.config.host, jira), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", c.basicAuth()))
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var issuePayload = struct {
		SearchResults []struct {
			Entity struct {
				Key string
			}
		}
	}{}

	if err = json.NewDecoder(res.Body).Decode(&issuePayload); err != nil {
		return nil, err
	}

	if len(issuePayload.SearchResults) == 0 {
		println("Failed to find bamboo statuses for " + jira)
		return nil, nil
	}
	key := issuePayload.SearchResults[0].Entity.Key

	req, err = http.NewRequest("GET", fmt.Sprintf(searchPlansTemplate, c.config.host, key), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", c.basicAuth()))
	res, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	var planPayload = struct {
		Results struct {
			Result []struct {
				Link struct {
					Href string
				}
				State          string
				LifeCycleState string
			}
		}
	}{}

	if err = json.NewDecoder(res.Body).Decode(&planPayload); err != nil {
		return nil, err
	}

	if len(planPayload.Results.Result) == 0 {
		println("Failed to find bamboo plans for " + jira)
		return nil, nil
	}

	state, ok := bambooStateToState[planPayload.Results.Result[0].State]
	if !ok {
		println("Failed to get a state from " + planPayload.Results.Result[0].State)
		return nil, nil
	}

	hrefParts := strings.Split(planPayload.Results.Result[0].Link.Href, "/")

	return &services.Status{
		Name:  "Bamboo",
		URL:   fmt.Sprintf("%s/browse/%s", c.config.host, hrefParts[len(hrefParts)-1]),
		State: state,
	}, nil
}

func (c *Client) basicAuth() string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.config.username, c.config.password)))
}
