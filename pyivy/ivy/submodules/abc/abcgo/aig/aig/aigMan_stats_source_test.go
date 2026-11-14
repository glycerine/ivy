package aig

import (
	"fmt"
	"strings"
	"testing"
)

func TestManagerStatsUseDriverLevelAndSourceFields(t *testing.T) {
	g := NewGraph()
	g.Name = "stats"
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	out := g.AddOutput(root)
	g.node(out).Level = 99 // C's CO level is auxiliary, not logic depth.
	got := captureAigOutput(t, func() { Aig_ManPrintStats(g) })
	want := "stats           : pi =     2  po =     1  and =       1  lev =   1\n"
	if got != want {
		t.Fatalf("stats = %q, want source %q", got, want)
	}
}

func TestManagerStatsAgainstCompiledSource(t *testing.T) {
	g := NewGraph()
	g.Name = "stats"
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	out := g.AddOutput(root)
	g.node(out).Level = 99
	want := captureAigOutput(t, func() {
		fmt.Printf("clean %d\n", Aig_ManHasNoGaps(g))
		Aig_ManPrintStats(g)
		g.AND(a.Not(), b)
		Aig_ManCleanup(g)
		fmt.Printf("deleted %d\n", Aig_ManHasNoGaps(g))
		Aig_ManPrintStats(g)
		Aig_ManSetRegNum(g, 1)
		Aig_ManPrintStats(g)
		xor := g.NewXor(a, b)
		Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{xor, InvalidEdge}})
		alternative := g.AND(a, root)
		g.equivs = map[int]Edge{root.NodeID(): alternative}
		fmt.Printf("typed %d\n", Aig_ManHasNoGaps(g))
		Aig_ManPrintStats(g)
	})
	got := runAigSourceFixture(t, "manager_stats_compare.c", "aigMan.c", "aigUtil.c")
	if got != want {
		t.Fatalf("source stats:\n%sGo stats:\n%s", got, want)
	}
}

func TestManagerStatsIncludeTypedObjectCounts(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	xor := g.NewXor(a, b)
	buffer := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{xor, InvalidEdge}})
	root := g.AND(buffer, c)
	g.AddOutput(root)
	g.AddOutput(a)
	Aig_ManSetRegNum(g, 1)
	g.equivs = map[int]Edge{root.NodeID(): g.AND(c, root)}
	got := captureAigOutput(t, func() { Aig_ManPrintStats(g) })
	for _, field := range []string{"lat =     1", "xor =     1", "ch =     1", "buf =     1"} {
		if !strings.Contains(got, field) {
			t.Errorf("missing source field %q in %q", field, got)
		}
	}
}

func TestHasNoGapsUsesSourceLiveObjectEquation(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	g.AND(a.Not(), b)
	if removed := Aig_ManCleanup(g); removed != 1 {
		t.Fatalf("cleanup removed %d nodes, want one", removed)
	}
	if Aig_ManHasNoGaps(g) != 1 {
		t.Fatal("source predicate counts live objects, including after deletion")
	}
	Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{a, InvalidEdge}})
	if Aig_ManHasNoGaps(g) != 0 {
		t.Fatal("source predicate must exclude managers containing buffers")
	}
}
