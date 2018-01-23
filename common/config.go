package common

import (
	"encoding/json"
	"io/ioutil"

	"github.com/de1ux/prowler/config/v1"
)

func LoadConfig(path string) (*v1.Config, error) {
	c := &v1.Config{}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return c, json.Unmarshal(bytes, c)
}
