package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"sync"
)

type config struct {
	Username string   `json:"username"`
	Repos    []string `json:"repos,omitempty"` // blank == all
	Token    string   `json:"token"`
	Services []string `json:"services,omitempty"` // blank == all
	Successs []string `json:"successStates,omitempty"`
	Pendings []string `json:"pendingStates,omitempty"`
	Failures []string `json:"failureStates,omitempty"`
	Conficts bool     `json:"hideMergeConflicts"`
	All      bool     `json:"showAllPrs"`

	metadata []meta // assigned in process
}

func (cfg *config) process(i int, wg *sync.WaitGroup) {
	// fmt.Println("processing", cfg.Repos[i])
	wg.Done()
}

func (cfg config) String() string {
	out := make([]string, 0, len(cfg.Repos))
	for i, data := range cfg.metadata {
		if o := data.String(); o != "" {
			out = append(out, cfg.Repos[i]+" | size=20\n"+o)
		}
	}
	return strings.Join(out, "\n---\n")
}

type meta struct{}

func (m meta) String() string {
	return "meta!!!"
}

func check(err error, doing string) {
	if err != nil {
		fmt.Printf("Error %s: %s\n", doing, err)
		os.Exit(1)
	}
}

func main() {
	// Parse Configuration
	usr, err := user.Current()
	check(err, "identifying user")
	data, err := ioutil.ReadFile(usr.HomeDir + "/.prowler.conf")
	check(err, "reading ~/.prowler.conf")
	var cfg config
	check(json.Unmarshal(data, &cfg), "unmarshaling json")

	// Process Repositories
	var wg sync.WaitGroup
	cfg.metadata = make([]meta, len(cfg.Repos))
	wg.Add(len(cfg.Repos))
	for i := range cfg.Repos {
		go cfg.process(i, &wg)
	}
	wg.Wait()

	// Print Results
	fmt.Println("\u2766\n---\n" + cfg.String())
}
