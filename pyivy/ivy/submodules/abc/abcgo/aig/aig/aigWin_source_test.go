package aig

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
)

func TestWindowFrontiersMatchOriginalC(t *testing.T) {
	rng := rand.New(rand.NewPCG(946, 295))
	var input, want strings.Builder
	fmt.Fprintln(&input, 32)
	for trial := 0; trial < 32; trial++ {
		g := NewGraph()
		pool := []Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		pool = append(pool, g.AND(pool[0], pool[1]))
		for step := 0; step < 10; step++ {
			a, b := rng.IntN(len(pool)*2), rng.IntN(len(pool)*2)
			fmt.Fprintln(&input, a, b)
			pool = append(pool, g.AND(dupPhase(pool[a/2], a&1 != 0), dupPhase(pool[b/2], b&1 != 0)))
		}
		root := pool[4]
		for _, e := range pool {
			if g.Node(e).Kind == And {
				root = Aig_Regular(e)
			}
		}
		g.AddOutput(root)
		for i := 0; i < 4; i++ {
			a := rng.IntN(len(pool) * 2)
			fmt.Fprintln(&input, a)
			g.AddOutput(dupPhase(pool[a/2], a&1 != 0))
		}
		for _, size := range []int32{2, 3, 4, 6} {
			for _, limit := range []int32{0, 1, 3, 100} {
				var front, visited []Edge
				Aig_ManFindCut(g, root, &front, &visited, size, limit)
				fmt.Fprintf(&want, "case %d root %d size %d limit %d front:", trial, root.NodeID(), size, limit)
				for _, e := range front {
					fmt.Fprintf(&want, " %d", e.NodeID())
				}
				fmt.Fprint(&want, " visited:")
				for _, e := range visited {
					fmt.Fprintf(&want, " %d", e.NodeID())
				}
				fmt.Fprint(&want, " marks:")
				for _, e := range g.LiveEdges() {
					if g.Node(e).MarkA {
						fmt.Fprintf(&want, " %d", e.NodeID())
					}
				}
				fmt.Fprintln(&want)
				Aig_ManCheckMarkA(g)
				if len(front) > int(size) {
					t.Fatal("frontier exceeded size limit")
				}
			}
		}
	}
	got := runAigSourceFixtureInput(t, "window_compare.c", input.String(), "aigMan.c", "aigObj.c", "aigOper.c", "aigTable.c", "aigMem.c", "aigFanout.c", "aigUtil.c", "aigWin.c", "../../misc/tim/timMan.c", "../../misc/mem/mem.c")
	if got != want.String() {
		t.Fatalf("C windows:\n%sGo windows:\n%s", got, want.String())
	}
}
