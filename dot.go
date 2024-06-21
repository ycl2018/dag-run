package dagRun

import "text/template"

var templateStr = `
digraph G {
"start"[shape=box,color="green"]
"end"[shape=box,color="red"]
{{- range $curNode, $nextNodes := .Edges}}
"{{printName $curNode }}" -> { {{- range $i, $node := $nextNodes }}{{- if $i}},{{end -}} " {{- printName $node -}}"{{end}}}
{{- end }}
"start" -> { {{- range $i, $node := .ToStart }} {{- if $i}},{{end -}} " {{- printName $node -}} " {{- end -}} }
{ {{- range $i, $node := .ToEnd }} {{- if $i}},{{end -}} " {{- printName $node -}} " {{- end -}} }  -> "end"
}
`

type dotContext struct {
	Edges   map[Node][]Node
	ToStart []Node
	ToEnd   []Node
}

var dotTemplate = template.Must(template.New("dag").Funcs(template.FuncMap{
	"printName": func(node Node) string {
		return node.Name()
	},
}).Parse(templateStr))
