package aig

import (
	"reflect"
	"testing"
)

func TestDuplicationCopiesOneHotMetadata(t *testing.T) {
	graph := NewGraph()
	graph.OneHotGroups = [][]int32{{0, 1}}
	copied := Aig_ManDupOrdered(graph)
	copied.OneHotGroups[0][0] = 7
	if graph.OneHotGroups[0][0] != 0 {
		t.Fatal("duplicate shares one-hot metadata")
	}
}

func TestSequentialCleanupRemapsOneHotMetadata(t *testing.T) {
	g := NewGraph()
	pi, a, b, c := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	_ = a
	g.AddOutput(g.OR(b, c))
	g.AddOutput(pi)
	g.AddOutput(pi)
	g.AddOutput(pi)
	Aig_ManSetRegNum(g, 3)
	g.OneHotGroups = [][]int32{{0, 1, 2}}
	g.OneHotInputGroups = [][]int32{{0, 2, 3}}
	Aig_ManSeqCleanup(g)
	if !reflect.DeepEqual(g.OneHotGroups, [][]int32{{0, 1}}) {
		t.Fatalf("register groups %v", g.OneHotGroups)
	}
	if !reflect.DeepEqual(g.OneHotInputGroups, [][]int32{{0, 1, 2}}) {
		t.Fatalf("input groups %v", g.OneHotInputGroups)
	}
}
