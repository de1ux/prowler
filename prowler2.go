package main

import (
	"os"
	"os/user"
	"path/filepath"
	"text/template"
	"time"

	"github.com/de1ux/prowler/common"
)

// Version is what version of code this is
var Version string

func main() {
	start := time.Now()
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	path := filepath.Join(user.HomeDir, ".prowler.json")
	config, err := common.LoadConfig(path)
	if err != nil {
		panic(err)
	}

	manifest, err := common.RunIntegration(config)
	if err != nil {
		panic(err)
	}
	duration := time.Now().Sub(start)

	manifest.Version = Version
	manifest.Duration = duration.String()

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"colorPr":     common.ColorPr,
		"colorStatus": common.ColorStatus,
		"colorIcon":   common.ColorIcon,
	}).Parse(common.BitbarManifestTemplate)
	if err != nil {
		panic(err)
	}

	if err = tmpl.Execute(os.Stdout, manifest); err != nil {
		panic(err)
	}
}
