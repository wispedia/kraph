package kraph

import (
	"fmt"
	"testing"
)

func TestNewGraph(t *testing.T) {

	g := NewGraph()

	id := NewNID("node1")
	n := NewNode(id)

	pid := NewNID("node2")
	pn := NewNode(pid)

	pid2 := NewNID("node3")
	pn2 := NewNode(pid2)

	g.AddNode(n)
	g.AddNode(pn)
	g.AddNode(pn2)
	g.AddNode(n)

	g.AddEdge(id, pid, 12.0)
	g.AddEdge(id, pid2, 10.0)
	g.AddEdge(id, pid, 1.0)

	g.ReplaceEdge(id, pid2, 0.9)

	smap, _ := g.GetSources(id)
	tmap, _ := g.GetTargets(pid)
	num := g.GetNodeCount()
	nodes := g.GetNodes()
	wgt, _ := g.GetWeight(id, pid)

	g.DeleteEdge(id, pid2)
	g.DeleteNode(pid2)

	nd := g.GetNode(id)

	fmt.Println(nd, smap, tmap, num, nodes, wgt, g.String())
	g.Init()
}
