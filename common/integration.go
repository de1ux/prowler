package common

import (
	"fmt"

	config "github.com/de1ux/prowler/config/v1"

	"github.com/de1ux/prowler/services/v1/bamboo"
	"github.com/de1ux/prowler/services/v1/travis"
	"github.com/de1ux/prowler/vcs/v1/bitbucket"
	"github.com/de1ux/prowler/vcs/v1/github"
)

type integration func(config *config.Config) (*Manifest, error)

// integrations combines vcs/services specific configurations into versioned generic vcs/services clients
var integrations = map[string]integration{
	"bitbucket_and_bamboo": func(config *config.Config) (*Manifest, error) {
		bitbucketConfig, err := bitbucket.NewConfig(config)
		if err != nil {
			return nil, err
		}

		bambooConfig, err := bamboo.NewConfig(config)
		if err != nil {
			return nil, err
		}

		bitbucketClient := bitbucket.NewClient(bitbucketConfig)
		bambooClient := bamboo.NewClient(bambooConfig)

		manifest := &Manifest{
			Entries: map[string][]*Entry{},
		}

		for _, repo := range config.Vcs.Repos {
			manifest.Entries[repo] = []*Entry{}

			prs, err := bitbucketClient.GetPullRequestsByRepo(repo)
			if err != nil {
				return nil, fmt.Errorf("Failed to get Bitbucket PRs, is the API token and username correct? %s", err)
			}

			for _, pr := range prs {
				entry := &Entry{Pr: pr}
				status, err := bambooClient.GetStatusByPullRequest(pr)
				if err != nil {
					return nil, fmt.Errorf("Failed to get Bamboo builds: %s", err)
				}
				if status == nil {
					continue
				}
				entry.Statuses = append(entry.Statuses, status)
				manifest.Entries[repo] = append(manifest.Entries[repo], entry)
			}
		}
		return manifest, nil
	},
	"github_and_travis": func(config *config.Config) (*Manifest, error) {
		githubConfig, err := github.NewConfig(config)
		if err != nil {
			return nil, err
		}

		travisConfig, err := travis.NewConfig(config)
		if err != nil {
			return nil, err
		}

		github.NewClient(githubConfig)
		travis.NewClient(travisConfig)

		return nil, nil
	},
}

func RunIntegration(config *config.Config) (*Manifest, error) {
	// TODO - check exists
	return integrations[config.Integration](config)
}
