package common

import (
	config "github.com/de1ux/prowler/config/v1"
	services "github.com/de1ux/prowler/services/v1"
	vcs "github.com/de1ux/prowler/vcs/v1"
)

type integration func(config *config.Config) (vcs.Client, []services.Client, error)

// integrations combines vcs/services specific configurations into versioned generic vcs/services clients
var integrations = map[string]integration{
	"bitbucket_and_bamboo": func(config *config.Config) (vcs.Client, []services.Client, error) {
		return nil, nil, nil
	},
}

func RunIntegration(config *config.Config) (*Manifest, error) {
	vcs, services, err := integrations[config.Integration](config)
	if err != nil {
		return nil, err
	}

	manifest := &Manifest{}

	for _, repo := range config.Vcs.Repos {
		prs, err := vcs.GetPullRequestsByRepo(repo)
		if err != nil {
			return nil, err
		}

		for _, pr := range prs {
			entry := &Entry{Pr: pr}
			for _, service := range services {
				statuses, err := service.GetStatusesByPullRequest(pr)
				if err != nil {
					return nil, err
				}
				entry.Statuses = append(entry.Statuses, statuses...)
			}
			manifest.Entries = append(manifest.Entries, entry)
		}
	}

	return manifest, nil
}
