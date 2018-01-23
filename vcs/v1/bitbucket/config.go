package bitbucket

type Config struct {
	username           string
	token              string
	repos              []string
	hideMergeConflicts bool
	showAllPrs         bool
}
