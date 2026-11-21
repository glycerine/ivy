package dec

import "testing"

// Four inputs, seven AND nodes. The complemented edges exercise both input phases.
func rewriteTruthGraph(rounds int) *Dec_Graph_t {
	g := Dec_GraphCreate(4)
	root := Dec_EdgeCreate(0, 0)
	for i := 0; i < rounds; i++ {
		root = Dec_GraphAddNodeAnd(g, Dec_EdgeNot(root), Dec_EdgeCreate(int32(1+i%3), int32(i%2)))
	}
	Dec_GraphSetRoot(g, Dec_EdgeNot(root))
	return g
}

func TestWordTruthPreservesPayloadsAndSelectedRoot(t *testing.T) {
	for _, size := range []int{1, 7, 60, 61, 700} {
		g := rewriteTruthGraph(size)
		for i := range int(g.size) {
			Dec_GraphSetNodeFunc(g, int32(i), i)
		}
		native := g.Native()
		for _, root := range []int{0, 4, int(g.size) - 1} {
			for phase := range 2 {
				Dec_GraphSetRoot(g, Dec_EdgeCreate(int32(root), int32(phase)))
				native.SetRoot(Edge{root, phase != 0})
				var want uint32
				for row := range 32 {
					if native.Evaluate([]bool{row&1 != 0, row&2 != 0, row&4 != 0, row&8 != 0}) {
						want |= 1 << row
					}
				}
				if got, other := Dec_GraphDeriveTruth(g), native.Truth32(); got != want || other != want {
					t.Fatalf("size %d root %d/%d: graph %08x adapter %08x, want %08x", size, root, phase, got, other, want)
				}
			}
		}
		for i := range int(g.size) {
			if v := Dec_GraphNodeFunc(g, int32(i)); v != i {
				t.Fatalf("node %d lost payload %v", i, v)
			}
		}
	}
}

func TestSmallWordTruthNeedsNoGraphCopy(t *testing.T) {
	g := rewriteTruthGraph(7)
	if a := testing.AllocsPerRun(100, func() { _ = Dec_GraphDeriveTruth(g) }); a != 0 {
		t.Fatalf("four-input decomposition truth allocated %.0f times", a)
	}
}

func BenchmarkDecompositionTruth32(b *testing.B) {
	g := rewriteTruthGraph(7)
	b.ReportAllocs()
	for b.Loop() {
		Dec_GraphDeriveTruth(g)
	}
}

func TestWordTruthRejectsMalformedTrailingNode(t *testing.T) {
	g := rewriteTruthGraph(7)
	last := Dec_GraphNodeLast(g)
	last.eEdge1 = Dec_EdgeCreate(g.size, 0)
	Dec_GraphSetRoot(g, Dec_EdgeCreate(4, 0))
	check := func(name string, evaluate func() uint32) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Errorf("%s accepted a forward edge after the selected root", name)
			}
		}()
		evaluate()
	}
	check("Go", g.Native().Truth32)
	check("C adapter", func() uint32 { return Dec_GraphDeriveTruth(g) })
}
