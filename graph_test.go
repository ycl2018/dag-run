package dagRun

import (
	"fmt"
	"testing"
)

type StringNode struct {
	Data string
}

func (s *StringNode) Name() string {
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
	//
	//"start" [color="green",shape=doublecircle]
	//"end" [color="red",shape=doublecircle]
	//
	//"A" -> {"C","D"}
	//"B" -> {"A","D","E"}
	//"C" -> {"D"}
	//"D" -> {"E"}
	//"E" -> {"end"}
	//"start" -> {"B"}
	//}
}

func ExampleGraph_DFS() {
	graph := g()
	// graph.AddEdge(nodes[4], nodes[0])
	err := graph.DFS(func(node Node) error {
		fmt.Println(node.Name())
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

func TestCircleBFS(t *testing.T) {
	graph := g()
	graph.AddEdge(nodes[4], nodes[0])
	err := graph.BFS(func(node Node) error {
		fmt.Println(node.Name())
		return nil
	})
	want := "graph has circle in nodes:[A C D E]"
	if err.Error() != want {
		t.Errorf("want err:%s but get:%+v", want, err)
	}
}

func TestCircleDFS(t *testing.T) {
	graph := g()
	graph.AddEdge(nodes[4], nodes[0])
	err := graph.DFS(func(node Node) error {
		fmt.Println(node.Name())
		return nil
	})
	want := "graph has circle, cur node:E ,next node:A"
	if err.Error() != want {
		t.Errorf("want err:%s but get:%v", want, err)
	}
}

func ExampleGraph_BFS() {
	graph := g()
	_ = graph.BFS(func(node Node) error {
		fmt.Println(node.Name())
		return nil
	})
	// OUTPUT:
	//B
	//A
	//C
	//D
	//E
}
