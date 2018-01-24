package bamboo

import (
	"fmt"

	"github.com/de1ux/prowler/config/v1"
)

type Config struct {
	host string
}

func NewConfig(config *v1.Config) (c *Config, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Failed to create bamboo config: %s", r)
		}
	}()
	c = &Config{}

	if len(config.Services) != 1 {
		// TODO - this is a hack
		panic(fmt.Sprintf("Bamboo is only configured to work by itself"))
	}

	m, ok := config.Services[0].(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("Failed to coerce to map of interfaces: %v", ok))
	}

	c.host, ok = m["host"].(string)
	if !ok {
		panic("Failed to parse host")
	}

	return
}
