package dagRun

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type StringNode struct {
	Data string
}

func (s *StringNode) String() string {
	return s.Data
}

var nodes = []*StringNode{
	{"A"},
	{"B"},
	{"C"},
	{"D"},
	{"E"},
}

var edges = [...][2]int{
	{1, 0},
	{1, 4},
	{1, 3},
	{0, 2},
	{0, 3},
	{2, 3},
	{3, 4},
}
var g = func() *Graph {
	graph := NewGraph()
	for _, node := range nodes {
		graph.AddNode(node)
	}
	for _, edge := range edges {
		graph.AddEdge(nodes[edge[0]], nodes[edge[1]])
	}
	return graph
}

func ExampleNewGraph() {
	fmt.Println(g().String())
	// OUTPUT:
	//[A]-> [C,D,]
	//[B]-> [A,E,D,]
	//[C]-> [D,]
	//[D]-> [E,]
	//[E]-> []
}

func ExampleGraph_DOT() {
	fmt.Println(g().DOT())
	// OUTPUT:
	//digraph G {
	//"start"[shape=box,color="green"]
	//"end"[shape=box,color="red"]
	//"A" -> {"C","D"}
	//"B" -> {"A","E","D"}
	//"C" -> {"D"}
	//"D" -> {"E"}
	//"start" -> {"B"}
	//{"E"}  -> "end"
	//}
	//
}

func ExampleGraph_DFS() {
	graph := g()
	// graph.AddEdge(nodes[4], nodes[0])
	err := graph.DFS(func(node Node) error {
		fmt.Println(node.String())
		return nil
	})
	if err != nil {
		fmt.Println("err ", err)
	}
	// OUTPUT:
	//A
	//C
	//D
	//E
	//B
}

func TestCircle(t *testing.T) {
	graph := g()
	graph.AddEdge(nodes[4], nodes[0])
	err := graph.DFS(func(node Node) error {
		fmt.Println(node.String())
		return nil
	})
	assert.EqualError(t, err, "graph has circle, cur node:E ,next node:A")
}

func ExampleGraph_BFS() {
	graph := g()
	_ = graph.BFS(func(node Node) error {
		fmt.Println(node.String())
		return nil
	})
	// OUTPUT:
	//A
	//C
	//D
	//E
	//B
}
