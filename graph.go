package dagRun

import (
	"fmt"
	"sync"
)

type Graph struct {
	Nodes []Node
	Edges map[Node][]Node
	mutex sync.Mutex
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: make([]Node, 0),
		Edges: make(map[Node][]Node),
		mutex: sync.Mutex{},
	}
}

type Node interface {
	fmt.Stringer
}

func (g *Graph) AddNode(n Node) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.Nodes = append(g.Nodes, n)
}

func (g *Graph) AddEdge(from Node, to Node) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.Edges[from] = append(g.Edges[from], to)
}

func (g *Graph) String() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	for i := 0; i < len(g.Nodes); i++ {
		fmt.Printf("[%v]-> [", g.Nodes[i])
		nearNodes := g.Edges[g.Nodes[i]]
		for j := 0; j < len(nearNodes); j++ {
			fmt.Printf("%v,", nearNodes[j])
		}
		fmt.Print("]\n")
	}
}

type Walker func(node Node) error

func NopeWalker(_ Node) error { return nil }

func (g *Graph) DFS(walker Walker) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	visited := make(map[Node]int, len(g.Nodes))
	for _, node := range g.Nodes {
		if err := g.dfs2(node, visited, walker); err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph) dfs(node Node, visited map[Node]struct{}, walker Walker) error {
	if g == nil || len(g.Nodes) == 0 {
		return nil
	}
	if _, ok := visited[node]; ok {
		return nil
	}
	stack := make([]Node, 1)
	var peek = 0
	stack[0] = node
	for peek >= 0 {
		_node := stack[peek]
		peek--
		visited[_node] = struct{}{}
		nears := g.Edges[_node]
		for i := 0; i < len(nears); i++ {
			if _, ok := visited[nears[i]]; ok {
				continue
			}
			peek++
			// stack满了
			if len(stack) == peek {
				stack = append(stack, nears[i])
			} else {
				stack[peek] = nears[i]
			}
		}
		if err := walker(_node); err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph) dfs2(node Node, visited map[Node]int, walker Walker) error {
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
			if err := g.dfs2(v, visited, walker); err != nil {
				return err
			}
		}
	}
	visited[node] = -1
	return nil
}

func (g *Graph) BFS(walker Walker) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	visited := make(map[Node]struct{}, len(g.Nodes))
	for _, node := range g.Nodes {
		if err := g.bfs(node, visited, walker); err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph) bfs(node Node, visited map[Node]struct{}, walker Walker) error {
	if g == nil || len(g.Nodes) == 0 {
		return nil
	}
	if _, ok := visited[node]; ok {
		return nil
	}
	var queue []Node
	var head = 0
	queue = append(queue, node)
	visited[node] = struct{}{}
	for len(queue) > head {
		n := queue[head]
		head++
		nears := g.Edges[n]
		for i := 0; i < len(nears); i++ {
			if _, ok := visited[nears[i]]; ok {
				continue
			}
			queue = append(queue, nears[i])
			visited[nears[i]] = struct{}{}
		}
		if err := walker(n); err != nil {
			return err
		}
	}
	return nil
}
