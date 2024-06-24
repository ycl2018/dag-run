package dagRun

import "text/template"

var templateStr = `
digraph G {
{{ if len .GraphAttr}}
{{- range $k, $v := .GraphAttr}}
{{print $v}}
{{- end -}}{{- end -}}
{{ if len .NodeCommonAttr }}

node[
{{- range $i, $v := .NodeCommonAttr}}
{{- if $i}},{{end}}
{{- print $v -}}
{{- end -}}
]
{{- end}}
"start"[shape=box,color="green"]
"end"[shape=box,color="red"]
{{ if len .NodeAttr }}
{{- range $i, $v := .NodeAttr }}
{{- print $v -}}
{{- end -}}{{- end -}}
{{ if len .EdgeCommonAttr}}
edge[
{{- range $i, $v := .EdgeCommonAttr}}
{{- if $i}},{{end}}
{{- print $v -}}
{{- end -}}
]
{{- end -}}
{{- range $curNode, $nextNodes := .Edges}}
"{{printName $curNode }}" -> {
{{- range $i, $node := $nextNodes }}{{if $i}},{{end}}"{{printName $node}}"{{end}}}
{{- end }}
"start" -> { {{- range $i, $node := .ToStart }}{{if $i}},{{end}}"{{printName $node}}"{{end}}}
{ {{- range $i, $node := .ToEnd }}{{if $i}},{{end}}"{{printName $node}}"{{end}}} -> "end"
}
`

type dotContext struct {
	Edges          map[Node][]Node
	ToStart        []Node
	ToEnd          []Node
	GraphAttr      []string
	NodeAttr       []string
	NodeCommonAttr []string
	EdgeCommonAttr []string
}

var dotTemplate = template.Must(template.New("dag").Funcs(template.FuncMap{
	"printName": func(node Node) string {
		return node.Name()
	},
}).Parse(templateStr))
