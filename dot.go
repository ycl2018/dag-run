package dagRun

import (
	"sort"
	"strings"
)

type dotContext struct {
	Nodes          []Node
	Edges          map[Node][]Node
	GraphAttr      map[string]string
	NodeCommonAttr map[string]string
	EdgeCommonAttr map[string]string
	NodeAttr       map[string]map[string]string
	EdgeAttr       map[string]map[string]string // key: from->to
}

const StartNodeName = "start"
const EndNodeName = "end"

type dummyNode string

func (d dummyNode) Name() string {
	return string(d)
}

func genDot(ctx *dotContext) string {
	var sb strings.Builder
	sb.WriteString("digraph G {\n")
	// define Graph attribute
	if len(ctx.GraphAttr) > 0 {
		sb.WriteString("\n")
		var graphAttrs []string
		for k, v := range ctx.GraphAttr {
			graphAttrs = append(graphAttrs, strings.Join([]string{k, v}, "="))
		}
		sort.Strings(graphAttrs)
		for _, k := range graphAttrs {
			sb.WriteString(k)
			sb.WriteString("\n")
		}
	}
	// define common nodes attributes
	if len(ctx.NodeCommonAttr) > 0 {
		sb.WriteString("\n")
		var nodeCommonAttrs []string
		for k, v := range ctx.NodeCommonAttr {
			nodeCommonAttrs = append(nodeCommonAttrs, strings.Join([]string{k, v}, "="))
		}
		sort.Strings(nodeCommonAttrs)
		sb.WriteString("node[")
		sb.WriteString(strings.Join(nodeCommonAttrs, ","))
		sb.WriteString("]")
		sb.WriteString("\n")
	}
	// define node Attr
	sb.WriteString("\n")
	// define start and end node
	var writeNodeAttr = func(nodeName string, delKey bool) {
		if len(ctx.NodeAttr) == 0 || len(ctx.NodeAttr[nodeName]) == 0 {
			return
		}
		showName := "\"" + nodeName + "\""
		sb.WriteString(showName + " [")
		var startAttrs []string
		for k, v := range ctx.NodeAttr[nodeName] {
			startAttrs = append(startAttrs, strings.Join([]string{k, v}, "="))
		}
		sort.Strings(startAttrs)
		sb.WriteString(strings.Join(startAttrs, ","))
		sb.WriteString("]")
		sb.WriteString("\n")
		if delKey {
			delete(ctx.NodeAttr, nodeName)
		}
	}
	writeNodeAttr(StartNodeName, true)
	writeNodeAttr(EndNodeName, true)
	if len(ctx.NodeAttr) > 0 {
		var nodeAttrs []string
		for k := range ctx.NodeAttr {
			nodeAttrs = append(nodeAttrs, k)
		}
		sort.Strings(nodeAttrs)
		for _, k := range nodeAttrs {
			writeNodeAttr(k, false)
		}
	}
	// define common edge attributes
	if len(ctx.EdgeCommonAttr) > 0 {
		sb.WriteString("\n")
		var edgeCommonAttr []string
		for k, v := range ctx.EdgeCommonAttr {
			edgeCommonAttr = append(edgeCommonAttr, strings.Join([]string{k, v}, "="))
		}
		sort.Strings(edgeCommonAttr)
		sb.WriteString("edge[")
		sb.WriteString(strings.Join(edgeCommonAttr, ","))
		sb.WriteString("]")
		sb.WriteString("\n")
	}
	// define edge attr
	if len(ctx.Edges) > 0 {
		sb.WriteString("\n")
		var edgeStarts []Node
		for k := range ctx.Edges {
			edgeStarts = append(edgeStarts, k)
		}
		sort.Slice(edgeStarts, func(i, j int) bool { return edgeStarts[i].Name() < edgeStarts[j].Name() })
		for _, edgeStart := range edgeStarts {
			startName := edgeStart.Name()
			sb.WriteString("\"" + startName + "\"")
			sb.WriteString(" -> {")
			var toNodesNames []string
			toNodes := ctx.Edges[edgeStart]
			for _, to := range toNodes {
				toNodesNames = append(toNodesNames, "\""+to.Name()+"\"")
			}
			sort.Strings(toNodesNames)
			sb.WriteString(strings.Join(toNodesNames, ","))
			sb.WriteString("}")

			if ctx.EdgeAttr == nil || len(ctx.EdgeAttr[startName]) == 0 {
				sb.WriteString("\n")
				continue
			}
			sb.WriteString(" [")
			var startAttrs []string
			for k, v := range ctx.EdgeAttr[startName] {
				startAttrs = append(startAttrs, strings.Join([]string{k, v}, "="))
			}
			sort.Strings(startAttrs)
			sb.WriteString(strings.Join(startAttrs, ","))
			sb.WriteString("]")
			sb.WriteString("\n")
		}

	}
	sb.WriteString("}")
	return sb.String()
}
