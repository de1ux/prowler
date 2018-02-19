package common

const BitbarManifestTemplate = "{{printf \"\u2766\"}} | color={{colorIcon .}}\n" +
	"---\n" +
	"{{range $repo, $entries := .Entries}}" +
	"" +
	"{{$repo}} | size=20\n" +
	"" +
	"{{$length := len $entries}}" +
	"{{if eq $length 0}}" +
	"No PRs found... | alternative=true\n" +
	"{{end}}" +
	"" +
	"{{range $entry := $entries}}" +
	"{{$entry.Pr.Title}} | href={{$entry.Pr.URL}} color={{colorPr $entry}}\n" +
	"{{range $status := $entry.Statuses}}" +
	"-- {{$status.Name}} | href={{$status.URL}} color={{colorStatus $status}}\n" +
	"{{end}}" +
	"{{end}}" +
	"" +
	"{{end}}" +
	"---\n" +
	"SoFi Prowler v{{.Version}}; loaded in: {{.Duration}} | alternate=true\n" +
	"---\n"
