package v1

type Config struct {
	Integration string
	Vcs         struct {
		Repos   []string
		Options interface{}
	}
	Services []interface{}
}
