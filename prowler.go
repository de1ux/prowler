package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"strings"
	"sync"
	"time"
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

	// Used for processing
	wg       sync.WaitGroup
	metadata []meta
	duration time.Duration
}

func (cfg *config) process() {
	start := time.Now()
	cfg.metadata = make([]meta, len(cfg.Repos))
	cfg.wg.Add(len(cfg.Repos))
	for i, repo := range cfg.Repos {
		go cfg.metadata[i].process(cfg, repo)
	}
	cfg.wg.Wait()
	cfg.duration = time.Since(start)
}

func (cfg *config) get(uri string) (*http.Response, error) {
	return http.Get("https://api.github.com" + uri + "?oauth_token=" + cfg.Token)
}

func (cfg *config) String() string {
	out := make([]string, 0, len(cfg.Repos))
	for i, data := range cfg.metadata {
		if o := data.String(); o != "" {
			out = append(out, cfg.Repos[i]+" | size=20\n"+o)
		}
	}
	return strings.Join(out, "\n---\n") + "\n---\nTook: " + cfg.duration.String()
}

type meta struct {
	output []byte
	res    *http.Response
	err    error
}

func (m *meta) process(ctx *config, repo string) {
	m.res, m.err = ctx.get("/repos/" + repo + "/pulls")
	if m.err == nil {
		m.output, m.err = ioutil.ReadAll(m.res.Body)
		m.res.Body.Close()
	}
	ctx.wg.Done()
}

func (m meta) String() string {
	if m.err != nil {
		return "error = " + m.err.Error()
	}
	return string(m.output[:20])
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

	// Process in parallel
	cfg.process()
	fmt.Println("\u2766\n---\n" + cfg.String())
}
