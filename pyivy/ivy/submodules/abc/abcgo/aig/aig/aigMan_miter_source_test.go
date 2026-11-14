package aig

import (
	"fmt"
	"strings"
	"testing"
)

func TestExtractMiterUsesSourceZeroInputPhase(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	Aig_ObjSetLevel(g, a, 3)
	Aig_ObjSetLevel(g, b, 7)
	x, y := g.AND(a.Not(), b.Not()), g.AND(a, b)
	q := Aig_ManExtractMiter(g, x, y)
	for mask := 0; mask < 4; mask++ {
		values, err := q.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
		// C normalizes the XOR to zero under the all-zero input assignment.
		want := mask == 1 || mask == 2
		if err != nil || values[0] != want {
			t.Errorf("assignment %d: %v (%v), want %t", mask, values, err, want)
		}
	}
	for _, input := range q.inputs {
		if q.node(input).Level != 0 {
			t.Errorf("miter CI level %d, want fresh source level zero", q.node(input).Level)
		}
	}
}

func TestExtractMiterRetainsSourceCopyBindings(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	x := g.AND(a.Not(), b.Not())
	q := Aig_ManExtractMiter(g, x, x)
	copy := Aig_ObjCopy(g, x)
	if !q.valid(copy) {
		t.Error("miter cleanup invalidated the source copy binding")
	}
	if count := Aig_ManNodeNum(q); count != 1 {
		t.Errorf("copied gates %d, want source's retained single gate", count)
	}
}

func TestExtractMiterAllSmallRootPhases(t *testing.T) {
	// Truth nibbles use assignment a + 2*b. The last root is a rewriting
	// buffer of !a & !b, which the source duplicator resolves transparently.
	tables := []int{15, 10, 12, 8, 4, 2, 1, 1}
	for first := range tables {
		for second := range tables {
			g := NewGraph()
			a, b := g.NewInput(), g.NewInput()
			roots := []Edge{g.Const1(), a, b, g.AND(a, b), g.AND(a.Not(), b), g.AND(a, b.Not()), g.AND(a.Not(), b.Not())}
			roots = append(roots, Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{roots[6], InvalidEdge}}))
			q := Aig_ManExtractMiter(g, roots[first], roots[second])
			want := tables[first] ^ tables[second]
			if want&1 != 0 {
				want ^= 15
			}
			for mask := 0; mask < 4; mask++ {
				out, err := q.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
				if err != nil || out[0] != (want&(1<<mask) != 0) {
					t.Fatalf("roots %d/%d assignment %d: %v (%v), want truth %x", first, second, mask, out, err, want)
				}
			}
			for _, root := range []Edge{roots[first], roots[second]} {
				if copy := Aig_ObjCopy(g, root); !q.valid(copy) {
					t.Fatalf("roots %d/%d: dangling copy %d", first, second, copy)
				}
				if err := q.Check(); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestExtractMiterAgainstCompiledSource(t *testing.T) {
	var want strings.Builder
	for mode := 0; mode < 2; mode++ {
		g := NewGraph()
		a, b := g.NewInput(), g.NewInput()
		Aig_ObjSetLevel(g, a, 3)
		Aig_ObjSetLevel(g, b, 7)
		x := g.AND(a.Not(), b.Not())
		y := x
		if mode == 0 {
			y = g.AND(a, b)
		}
		q := Aig_ManExtractMiter(g, x, y)
		truth := 0
		for mask := 0; mask < 4; mask++ {
			out, err := q.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
			if err != nil {
				t.Fatal(err)
			}
			if out[0] {
				truth |= 1 << mask
			}
		}
		live := 0
		if q.valid(Aig_ObjCopy(g, x)) {
			live = 1
		}
		fmt.Fprintf(&want, "mode %d truth %d nodes %d ci-levels %d/%d copy-live %d\n", mode, truth, Aig_ManNodeNum(q), q.node(q.inputs[0]).Level, q.node(q.inputs[1]).Level, live)
	}
	got := runAigSourceFixture(t, "miter_compare.c", "aigMan.c", "aigObj.c", "aigOper.c", "aigTable.c", "aigMem.c", "aigFanout.c", "aigUtil.c", "aigCheck.c", "../../misc/tim/timMan.c", "../../misc/mem/mem.c")
	if got != want.String() {
		t.Fatalf("source miters:\n%sGo miters:\n%s", got, want.String())
	}
}
