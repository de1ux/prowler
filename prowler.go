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

type status int

const (
	failed status = iota
	pending
	success
)

func (s status) String() string {
	switch s {
	case pending:
		return "#f89406"
	case success:
		return "#00bb00"
	}
	return "#bb0000"
}

type config struct {
	Username string   `json:"username"`
	Repos    []string `json:"repos,omitempty"` // blank == all
	Token    string   `json:"token"`
	Services []string `json:"services,omitempty"` // blank == all
	Successs []string `json:"successStates,omitempty"`
	Pendings []string `json:"pendingStates,omitempty"`
	Failures []string `json:"failureStates,omitempty"` // TODO: remove because it's ignored
	Conficts bool     `json:"hideMergeConflicts"`
	All      bool     `json:"showAllPrs"`

	// Used for processing
	wg       sync.WaitGroup
	color    status
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

	cfg.color = success
	for _, repo := range cfg.metadata {
		for _, pr := range repo.prs {
			if !pr.Merge {
				cfg.color = failed
			}
			if cfg.color > pr.color {
				cfg.color = pr.color
			}
		}
	}
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
	out = append(out, "Took: "+cfg.duration.String())
	return strings.Join(out, "\n---\n")
}

type repoMeta struct {
	prs []*prMeta
	res *http.Response
	err error
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
	out := make([]string, len(m.prs))
	for i, pr := range m.prs {
		out[i] = pr.String()
	}
	return strings.Join(out, "\n")
}

func check(err error, doing string) {
	if err != nil {
		fmt.Printf("Error %s: %s\n", doing, err) // TODO: make this a pretty error
		os.Exit(1)
	}
}

type prMeta struct {
	URL   string `json:"url"`
	Link  string `json:"html_url"`
	Title string `json:"title"`
	User  struct {
		Login string `json:"login"`
	} `json:"user"` // TODO(benchmark): use lazy json.RawMessage here and check for substring existance
	Stats string `json:"statuses_url"`
	Merge bool   `json:"mergeable"`

	// processing stuff
	stats []*statsMeta
	res   *http.Response
	color status
	err   error
}

func (m *prMeta) process(ctx *config) {
	// Self mining to make sure the PR is mergable
	if !ctx.Conficts {
		m.res, m.err = ctx.get(m.URL)
		if m.err == nil {
			m.err = json.NewDecoder(m.res.Body).Decode(m)
			m.res.Body.Close()
		}
	} else {
		m.Merge = true
	}

	// Get Statuses
	if m.err == nil {
		m.res, m.err = ctx.get(m.Stats)
	}
	if m.err == nil {
		m.err = json.NewDecoder(m.res.Body).Decode(&m.stats)
		m.res.Body.Close()
	}

	// Get latest status for each target (getLatestStatus)
	if m.err == nil {
		unique := make(map[string]*statsMeta, len(m.stats))
		for _, stat := range m.stats {
			entry, ok := unique[stat.URL]
			if !ok || stat.Stamp.After(entry.Stamp) {
				entry = stat
			}
			unique[stat.URL] = entry
		}
		m.stats = make([]*statsMeta, 0, len(unique))
		for _, stat := range unique {
			for _, slug := range ctx.Services {
				if strings.Contains(stat.URL, slug) {
					m.stats = append(m.stats, stat)
					stat.process(ctx) // TODO: parallel
					break
				}
			}
		}
	}

	// Setting color based on children
	m.color = success
	for _, stat := range m.stats {
		if m.color > stat.color {
			m.color = stat.color
		}
	}
}

func (m prMeta) String() string {
	if m.err != nil {
		return "error = " + m.err.Error() // TODO: pretty error
	}
	out := make([]string, len(m.stats))
	for i, stat := range m.stats {
		out[i] = stat.String()
	}
	mergable := ""
	if !m.Merge {
		mergable = "\U0001F6AB"
	}
	return fmt.Sprintf("%s %s| href=%s color=%s\n%s", m.Title, mergable, m.Link, m.color, strings.Join(out, "\n"))
}

type statsMeta struct {
	Ctx   string    `json:"context"`
	URL   string    `json:"target_url"`
	State string    `json:"state"`
	Stamp time.Time `json:"updated_at"`
	color status
}

func (m *statsMeta) process(ctx *config) {
	m.color = failed
	for _, state := range ctx.Successs {
		if m.State == state {
			m.color = success
			return
		}
	}
	for _, state := range ctx.Pendings {
		if m.State == state {
			m.color = pending
			return
		}
	}
}

func (m statsMeta) String() string {
	return fmt.Sprintf("-- %s | href=%s color=%s", m.Ctx, m.URL, m.color)
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
	fmt.Printf("\u2766 | color=%s\n---\n%s", cfg.color, cfg.String())
}
