package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type config struct {
	GitFolder    string `json:"git_folder"`
	AccessToken  string `json:"access_token"`
	Username     string `json:"username"`
	Repositories []struct {
		Name string `json:"name"`
	} `json:"repositories"`
}

type branch struct {
	Name         string
	Status       string
	Statuses     []github.RepoStatus
	LastStatusAt *time.Time
}

type repo struct {
	Name    string
	Results []*branch
}

var statusToEmoji = map[string]string{
	"success": "\U0001F7E2",
	"pending": "\U0001F7E1",
	"failure": "\U0001F534",
	"error":   "\U0001F518",
}

func getConfig() (*config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile(home + "/.prowler")
	if err != nil {
		return nil, err
	}

	c := &config{}
	return c, json.Unmarshal(b, c)
}

func writeSkipList(l map[string]int) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	b, err := json.Marshal(&l)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(home+"/.prowler-skip-list", b, 0644)
}

func getSkipList() (map[string]int, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile(home + "/.prowler-skip-list")
	if err != nil {
		// TODO - check for err not exists specifically. this will catch any error
		return map[string]int{}, writeSkipList(map[string]int{})
	}

	s := map[string]int{}
	return s, json.Unmarshal(b, &s)
}

func getClient(accessToken string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	return github.NewClient(tc)
}

func main() {
	c, err := getConfig()
	if err != nil {
		panic(err)
	}

	var repoResults []*repo
	for _, repository := range c.Repositories {
		org, project, err := getOrgAndProjectNameFromRepoName(repository.Name)
		if err != nil {
			panic(err)
		}

		//getPullRequests(c, org, project)

		branchResults, err := getBranches(c, org, project)
		if err != nil {
			panic(err)
		}
		repoResults = append(repoResults, &repo{Name: repository.Name, Results: branchResults})
	}

	fmt.Printf("❦\n")
	fmt.Printf("---\n")
	for _, repo := range repoResults {
		fmt.Printf(repo.Name + " | size=20\n")
		for _, branch := range repo.Results {
			status := statusToEmoji[branch.Status]
			fmt.Printf(branch.Name + " " + status + " | href=https://github.com/" + repo.Name + "/compare/" + branch.Name + "?expand=1\n")
			fmt.Printf("-- Gitlab\n")
			for _, status := range branch.Statuses {
				statusCode := statusToEmoji[*status.State]
				fmt.Printf("---- " + *status.Context + " " + statusCode + "| href=" + *status.TargetURL + "\n")
			}
			fmt.Printf("-- Delete\n")
			_, projectName, err := getOrgAndProjectNameFromRepoName(repo.Name)
			if err != nil {
				panic(err)
			}

			fmt.Printf("---- Confirm | shell=cd param1=\"" + c.GitFolder + "/" + projectName + " && git push origin :" + branch.Name + "\" terminal=true \n")
		}
	}
}

func getOrgAndProjectNameFromRepoName(name string) (string, string, error) {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) < 2 {
		return "", "", errors.New("too few parts")
	}
	return parts[0], parts[1], nil
}

func getPullRequests(c *config, org, project string) error {
	client := getClient(c.AccessToken)
	pulls, _, err := client.PullRequests.List(context.Background(), org, project, &github.PullRequestListOptions{
		State:       "all",
		Base:        "",
		Sort:        "",
		Direction:   "",
		ListOptions: github.ListOptions{},
	})
	if err != nil {
		return err
	}
	fmt.Printf("%+v", pulls)
	return nil
}

func getBranches(c *config, org, project string) ([]*branch, error) {
	client := getClient(c.AccessToken)

	skipList, err := getSkipList()
	if err != nil {
		return nil, err
	}

	var branchResults []*branch
	opt := &github.ListOptions{
		Page:    0,
		PerPage: 100, // TODO - this might need to be tweaked
	}
	for {
		// TODO - goroutines
		branches, resp, err := client.Repositories.ListBranches(context.Background(), org, project, opt)
		if err != nil {
			return nil, err
		}
		for _, branchData := range branches {
			if branchData.Commit == nil {
				continue
			}
			if _, found := skipList[*branchData.Commit.SHA]; found {
				continue
			}

			commit, _, err := client.Repositories.GetCommit(context.Background(), org, project, *branchData.Commit.SHA)
			if err != nil {
				return nil, err
			}
			if commit.Author == nil || commit.Author.Login == nil {
				continue
			}
			if *commit.Author.Login != c.Username {
				skipList[*branchData.Commit.SHA] = 1
				continue
			}

			status, statuses, err := getStatuses(client, org, project, *branchData.Commit.SHA)

			var lastStatusAt *time.Time
			if len(statuses) > 0 {
				lastStatusAt = statuses[0].CreatedAt
			}
			branchResults = append(branchResults, &branch{Name: *branchData.Name, Status: status, Statuses: statuses, LastStatusAt: lastStatusAt})
		}

		// sort branches by last status
		sort.Slice(branchResults, func(i, j int) bool {
			if branchResults[i].LastStatusAt == nil || branchResults[j].LastStatusAt == nil {
				return false
			}
			return branchResults[i].LastStatusAt.After(*branchResults[j].LastStatusAt)
		})

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		if err := writeSkipList(skipList); err != nil {
			return nil, err
		}
	}
	return branchResults, nil
}

func getStatuses(client *github.Client, org string, project string, sha string) (string, []github.RepoStatus, error) {
	var statuses []github.RepoStatus
	var status *github.CombinedStatus
	var resp *github.Response
	var err error

	opt := &github.ListOptions{PerPage: 100}
	for {
		status, resp, err = client.Repositories.GetCombinedStatus(context.Background(), org, project, sha, opt)
		if err != nil {
			return "", nil, err
		}
		statuses = append(statuses, status.Statuses...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// sort alphabetically
	sort.Slice(statuses, func(i, j int) bool {
		if *statuses[i].State < *statuses[j].State {
			return true
		}
		if *statuses[i].State > *statuses[j].State {
			return false
		}
		return *statuses[i].Context < *statuses[j].Context
	})

	return *status.State, statuses, err
}
