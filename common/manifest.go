package common

import (
	services "github.com/de1ux/prowler/services/v1"
	vcs "github.com/de1ux/prowler/vcs/v1"
)

type Manifest struct {
	Entries []*Entry
}

type Entry struct {
	Pr       *vcs.PullRequest
	Statuses []*services.Status
}
