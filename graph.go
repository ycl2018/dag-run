package dagRun

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

type Graph struct {
	Nodes []Node
	Edges map[Node][]Node
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: make([]Node, 0),
		Edges: make(map[Node][]Node),
	}
}

type Node interface {
	fmt.Stringer
}

func (g *Graph) AddNode(n Node) {
	g.Nodes = append(g.Nodes, n)
}

func (g *Graph) AddEdge(from Node, to Node) {
	g.Edges[from] = append(g.Edges[from], to)
}

func (g *Graph) String() string {
	var sb strings.Builder
	for i := 0; i < len(g.Nodes); i++ {
		sb.WriteString(fmt.Sprintf("[%v]-> [", g.Nodes[i]))
		nearNodes := g.Edges[g.Nodes[i]]
		for j := 0; j < len(nearNodes); j++ {
			sb.WriteString(fmt.Sprintf("%s,", nearNodes[j]))
		}
		sb.WriteString("]\n")
	}
	return sb.String()
}

func (g *Graph) DOT() string {
	var dc dotContext
	dc.Edges = g.Edges
	var inDegree = map[Node]int{}
	// 0 outDegrees
	for _, n := range g.Nodes {
		if to, ok := g.Edges[n]; !ok {
			dc.ToEnd = append(dc.ToEnd, n)
		} else {
			for _, toN := range to {
				inDegree[toN]++
			}
		}
	}
	//0 indegrees
	for _, n := range g.Nodes {
		if inDegree[n] == 0 {
			dc.ToStart = append(dc.ToStart, n)
		}
	}

	buf := new(bytes.Buffer)
	err := dotTemplate.Execute(buf, dc)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

type Walker func(node Node) error

func NopeWalker(_ Node) error { return nil }

func (g *Graph) DFS(walker Walker) error {
	visited := make(map[Node]int, len(g.Nodes))
	for _, node := range g.Nodes {
		if err := g.dfs(node, visited, walker); err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph) dfs(node Node, visited map[Node]int, walker Walker) error {
	if g == nil || len(g.Nodes) == 0 {
		return nil
	}
	if visited[node] != 0 {
		return nil
	}
	if err := walker(node); err != nil {
		return fmt.Errorf("walf func return err:%v", err)
	}
	visited[node] = 1
	for _, v := range g.Edges[node] {
		if visited[v] == 1 {
			return fmt.Errorf("graph has circle, cur node:%s ,next node:%s", node, v)
		} else if visited[v] == -1 {
			continue
		} else {
			if err := g.dfs(v, visited, walker); err != nil {
				return err
			}
		}
	}
	visited[node] = -1
	return nil
}

func (g *Graph) BFS(walker Walker) error {
	var visitedNodesNum int
	inDegrees := make(map[Node]int, len(g.Nodes))
	for _, n := range g.Nodes {
		inDegrees[n] = 0
	}

	for _, to := range g.Edges {
		for _, n := range to {
			inDegrees[n]++
		}
	}
	var zeroDegreeNodes []Node
	for _, v := range g.Nodes {
		if inDegrees[v] == 0 {
			zeroDegreeNodes = append(zeroDegreeNodes, v)
		}
	}

	for len(zeroDegreeNodes) > 0 {
		curNode := zeroDegreeNodes[0]
		zeroDegreeNodes = zeroDegreeNodes[1:]
		if err := walker(curNode); err != nil {
			return err
		}
		visitedNodesNum++
		for _, to := range g.Edges[curNode] {
			inDegrees[to]--
			if inDegrees[to] == 0 {
				zeroDegreeNodes = append(zeroDegreeNodes, to)
			}
		}
	}
	// check circle
	if visitedNodesNum < len(g.Nodes) {
		var circleNodes []Node
		for n, inDegree := range inDegrees {
			if inDegree != 0 {
				circleNodes = append(circleNodes, n)
			}
		}
		sort.Slice(circleNodes, func(i, j int) bool {
			return circleNodes[i].String() < circleNodes[j].String()
		})
		return fmt.Errorf("graph has circle in nodes:%v", circleNodes)
	}

	return nil
}
