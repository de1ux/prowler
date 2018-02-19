package bamboo

import (
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

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var issuePayload = struct {
		SearchResults []struct {
			Entity struct {
				Key      string
				PlanName string
			}
		}
	}{}

	if err = json.NewDecoder(res.Body).Decode(&issuePayload); err != nil {
		return nil, err
	}

	if len(issuePayload.SearchResults) == 0 {
		//println("Failed to find bamboo statuses for " + jira)
		return nil, nil
	}

	lowestScore := -1
	var status *services.Status
	for _, result := range issuePayload.SearchResults {
		key := result.Entity.Key

		req, err = http.NewRequest("GET", fmt.Sprintf(searchPlansTemplate, c.config.host, key), nil)
		if err != nil {
			return nil, err
		}

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
			continue
		}

		state, ok := bambooStateToState[planPayload.Results.Result[0].State]
		if !ok {
			//println("Failed to get a state from " + planPayload.Results.Result[0].State)
			return nil, nil
		}

		hrefParts := strings.Split(planPayload.Results.Result[0].Link.Href, "/")

		trimmedRepo := pr.Repo
		repoSplit := strings.Split(trimmedRepo, "/")
		if len(repoSplit) > 1 {
			trimmedRepo = strings.Join(repoSplit[1:], "/")
		}

		currentScore := wagnerFischer(result.Entity.PlanName, trimmedRepo, 1, 1, 2)
		if lowestScore == -1 || currentScore < lowestScore {
			lowestScore = currentScore
			status = &services.Status{
				Name:  "Bamboo",
				URL:   fmt.Sprintf("%s/browse/%s", c.config.host, hrefParts[len(hrefParts)-1]),
				State: state,
			}
		}
	}

	if status == nil {
		//println("Failed to find bamboo plans for " + jira)
	}
	return status, nil
}

// stolen from https://github.com/xrash/smetrics
func wagnerFischer(a, b string, icost, dcost, scost int) int {
	// Allocate both rows.
	row1 := make([]int, len(b)+1)
	row2 := make([]int, len(b)+1)
	var tmp []int

	// Initialize the first row.
	for i := 1; i <= len(b); i++ {
		row1[i] = i * icost
	}

	// For each row...
	for i := 1; i <= len(a); i++ {
		row2[0] = i * dcost

		// For each column...
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				row2[j] = row1[j-1]
			} else {
				ins := row2[j-1] + icost
				del := row1[j] + dcost
				sub := row1[j-1] + scost

				if ins < del && ins < sub {
					row2[j] = ins
				} else if del < sub {
					row2[j] = del
				} else {
					row2[j] = sub
				}
			}
		}

		// Swap the rows at the end of each row.
		tmp = row1
		row1 = row2
		row2 = tmp
	}

	// Because we swapped the rows, the final result is in row1 instead of row2.
	return row1[len(row1)-1]
}
