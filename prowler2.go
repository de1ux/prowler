package main

import (
	"os/user"
	"path/filepath"

	"github.com/de1ux/prowler/common"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	path := filepath.Join(user.HomeDir, ".prowler2.conf")
	config, err := common.LoadConfig(path)
	if err != nil {
		panic(err)
	}

	_, err = common.RunIntegration(config)
	if err != nil {
		panic(err)
	}
}
