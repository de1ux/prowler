package bitbucket

import (
	"fmt"

	"github.com/de1ux/prowler/config/v1"
)

type Config struct {
	username           string
	token              string
	hideMergeConflicts bool
	showAllPrs         bool
}

func NewConfig(config *v1.Config) (c *Config, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Failed to create bitbucket config: %s", r)
		}
	}()
	c = &Config{}

	m, ok := config.Vcs.Options.(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("Failed to coerce: %v", ok))
	}

	c.username, ok = m["username"].(string)
	if !ok {
		panic("Failed to parse username")
	}
	c.token, ok = m["token"].(string)
	if !ok {
		panic("Failed to parse token")
	}
	c.hideMergeConflicts, ok = m["hideMergeConflicts"].(bool)
	if !ok {
		panic("Failed to parse hideMergeConflicts")
	}
	c.showAllPrs, ok = m["showAllPrs"].(bool)
	if !ok {
		panic("Failed to parse showAllPrs")
	}

	return
}
