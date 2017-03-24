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
	metadata []repoMeta
	duration time.Duration
}

func (cfg *config) process() {
	start := time.Now()
	cfg.metadata = make([]repoMeta, len(cfg.Repos))
	cfg.wg.Add(len(cfg.Repos))
	for i, repo := range cfg.Repos {
		go cfg.metadata[i].process(cfg, repo)
	}
	cfg.wg.Wait()
	cfg.duration = time.Since(start)
}

func (cfg *config) get(uri string) (*http.Response, error) {
	if !strings.HasPrefix(uri, "https://api.github.com") {
		uri = "https://api.github.com" + uri
	}
	return http.Get(uri + "?oauth_token=" + cfg.Token)
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

type repoMeta struct {
	prs    []*prMeta
	output []byte
	res    *http.Response
	err    error
}

func (m *repoMeta) process(ctx *config, repo string) {
	m.res, m.err = ctx.get("/repos/" + repo + "/pulls")
	if m.err == nil {
		m.err = json.NewDecoder(m.res.Body).Decode(&m.prs)
		m.res.Body.Close()
	}

	// Filter to only my pull requests (filterPullRequestsByUser)
	if m.err == nil && !ctx.All {
		filtered := make([]*prMeta, 0)
		for _, pr := range m.prs {
			if pr.User.Login == ctx.Username {
				filtered = append(filtered, pr)
			}
		}
		m.prs = filtered
	}

	// Fetch Statuses of the PR (fetchStatuses)
	if m.err == nil {
		for _, pr := range m.prs {
			pr.process(ctx) // TODO: parallel
		}
	}
	ctx.wg.Done()
}

func (m repoMeta) String() string {
	if m.err != nil {
		return "error = " + m.err.Error() // TODO: make this a pretty error
	}
	if len(m.prs) == 0 {
		return ""
	}
	return fmt.Sprintf("%s", m.prs)
}

func check(err error, doing string) {
	if err != nil {
		fmt.Printf("Error %s: %s\n", doing, err) // TODO: make this a pretty error
		os.Exit(1)
	}
}

type prMeta struct {
	Num   int    `json:"number"`
	Title string `json:"title"`
	User  struct {
		Login string `json:"login"`
	} `json:"user"` // TODO(benchmark): use lazy json.RawMessage here and check for substring existance
	Stats string `json:"statuses_url"`

	// processing stuff
	err error
}

func (m *prMeta) process(ctx *config) {
	// m.err = errors.New("TODO")
}

func (m prMeta) String() string {
	if m.err != nil {
		return "error = " + m.err.Error() // TODO: pretty error
	}
	return fmt.Sprintf("#: %d; Title: %s; Usr: %s; Stats: %s", m.Num, m.Title, m.User.Login, m.Stats)
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
