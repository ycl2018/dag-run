package dagRun

import "text/template"

var templateStr = `
digraph G {
"start"[shape=box,color="green"]
"end"[shape=box,color="red"]
{{- range $curNode, $nextNodes := .Edges}}
"{{print $curNode }}" -> { {{- range $i, $node := $nextNodes }}{{- if $i}},{{end -}} " {{- print $node -}}"{{end}}}
{{- end }}
"start" -> { {{- range $i, $node := .ToStart }} {{- if $i}},{{end -}} " {{- print $node -}} " {{- end -}} }
{ {{- range $i, $node := .ToEnd }} {{- if $i}},{{end -}} " {{- print $node -}} " {{- end -}} }  -> "end"
}
`

type dotContext struct {
 Edges   map[Node][]Node
 ToStart []Node
 ToEnd   []Node
}

var dotTemplate = template.Must(template.New("dag").Parse(templateStr))
