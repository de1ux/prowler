package common

const BitbarTemplate = `{{printf "\u2766"}} | color={{colorIcon .}}
---
{{range $repo, $entries := .Entries}}
{{$repo}} | size=20
{{range $entry := $entries}}
{{$entry.Pr.Title}} | href={{$entry.Pr.URL}} color={{colorPr $entry}}
{{range $status := $entry.Statuses}}
-- {{$status.Name}} | href={{$status.URL}} color={{colorStatus $status}}
{{end}}
{{end}}
{{end}}

---
Prowler v{{.Version}}; loaded in: {{.Duration}} | alternate=true
`
