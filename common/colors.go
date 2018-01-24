package common

import (
	services "github.com/de1ux/prowler/services/v1"
)

const (
	yellow = "#f89406"
	green  = "#00bb00"
	red    = "#bb0000"
)

func ColorPr(entry *Entry) string {
	for _, status := range entry.Statuses {
		if status.State != services.Passing {
			return red
		}
	}
	return green
}

func ColorStatus(status *services.Status) string {
	if status.State != services.Passing {
		return red
	}
	return green
}

func ColorIcon(manifest *Manifest) string {
	for _, value := range manifest.Entries {
		for _, entry := range value {
			for _, status := range entry.Statuses {
				if status.State != services.Passing {
					return red
				}
			}
		}
	}
	return green
}
