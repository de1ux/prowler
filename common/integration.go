package common

import (
	config "github.com/de1ux/prowler/config/v1"
	services "github.com/de1ux/prowler/services/v1"
	vcs "github.com/de1ux/prowler/vcs/v1"

	bamboo "github.com/de1ux/prowler/services/v1/bamboo"
	bitbucket "github.com/de1ux/prowler/vcs/v1/bitbucket"
)

type integration func(config *config.Config) (vcs.Client, []services.Client, error)

// integrations combines vcs/services specific configurations into versioned generic vcs/services clients
var integrations = map[string]integration{
	"bitbucket_and_bamboo": func(config *config.Config) (vcs.Client, []services.Client, error) {
		bitbucketConfig, err := bitbucket.NewConfig(config)
		if err != nil {
			return nil, nil, err
		}

		bambooConfig, err := bamboo.NewConfig(config)
		if err != nil {
			return nil, nil, err
		}

		return bitbucket.NewClient(bitbucketConfig), []services.Client{bamboo.NewClient(bambooConfig)}, nil
	},
}

func RunIntegration(config *config.Config) (*Manifest, error) {
	vcs, services, err := integrations[config.Integration](config)
	if err != nil {
		return nil, err
	}

	manifest := &Manifest{
		Entries: map[string][]*Entry{},
	}

	for _, repo := range config.Vcs.Repos {
		prs, err := vcs.GetPullRequestsByRepo(repo)
		if err != nil {
			return nil, err
		}

		for _, pr := range prs {
			entry := &Entry{Pr: pr}
			for _, service := range services {
				status, err := service.GetStatusByPullRequest(pr)
				if err != nil {
					return nil, err
				}
				if status == nil {
					continue
				}
				entry.Statuses = append(entry.Statuses, status)
			}
			manifest.Entries[repo] = append(manifest.Entries[repo], entry)
		}
	}

	return manifest, nil
}
