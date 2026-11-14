package aig

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestFrameExpansionFullSourceFlagMatrix(t *testing.T) {
	p := NewGraph()
	a, b, s, stateT := p.NewInput(), p.NewInput(), p.NewInput(), p.NewInput()
	x, y := p.AND(a, s), p.AND(a.Not(), stateT)
	nextS, nextT := p.OR(x, y), p.AND(b, s.Not())
	po0, po1 := p.AND(s, stateT), p.OR(a, b.Not())
	p.AND(b, stateT)
	p.AddOutput(po0)
	p.AddOutput(po1)
	p.AddOutput(nextS)
	p.AddOutput(nextT)
	Aig_ManSetRegNum(p, 2)
	p.Name, p.Spec = "frame-matrix", "source-spec"
	before := p.Nodes()
	var want strings.Builder
	for frames := int32(1); frames <= 4; frames++ {
		for flags := int32(0); flags < 16; flags++ {
			var mapping []Edge
			q := Aig_ManFrames(p, frames, flags&1, (flags>>1)&1, (flags>>2)&1, (flags>>3)&1, &mapping)
			fmt.Fprintf(&want, "case %d %d ci %d co %d regs %d true %d %d objects %d max %d name %s spec %s\n", frames, flags, q.InputCount(), q.OutputCount(), q.RegisterCount(), q.TruePiCount(), q.TruePoCount(), q.NodeCount(), len(q.nodes), q.Name, q.Spec)
			fmt.Fprint(&want, "map:")
			for _, e := range mapping {
				if e == InvalidEdge {
					fmt.Fprint(&want, " -")
					continue
				}
				status := 'l'
				if !q.valid(e) {
					status = 'd'
				}
				fmt.Fprintf(&want, " %c%d", status, e)
			}
			fmt.Fprintln(&want)
			for _, e := range q.LiveEdges() {
				n := q.Node(e)
				phase := 0
				if n.Phase {
					phase = 1
				}
				fmt.Fprintf(&want, "object %d refs %d level %d phase %d", e.NodeID(), n.Refs, n.Level, phase)
				switch n.Kind {
				case Input:
					fmt.Fprintf(&want, " ci %d", n.CioID)
				case Output:
					fmt.Fprintf(&want, " co %d %d", n.CioID, n.Fanins[0])
				case And:
					fmt.Fprintf(&want, " and %d %d", n.Fanins[0], n.Fanins[1])
				}
				fmt.Fprintln(&want)
			}
			for mask := 0; mask < 1<<q.InputCount(); mask++ {
				in := make([]bool, q.InputCount())
				for i := range in {
					in[i] = mask&(1<<i) != 0
				}
				out, err := q.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				packed := 0
				for i, v := range out {
					if v {
						packed |= 1 << i
					}
				}
				fmt.Fprintf(&want, "truth %d %d\n", mask, packed)
			}
			if err := q.Check(); err != nil {
				t.Fatalf("frames %d flags %d: %v", frames, flags, err)
			}
		}
	}
	got := runAigSourceFixture(t, "frames_compare.c", "aigMan.c", "aigObj.c", "aigOper.c", "aigTable.c", "aigMem.c", "aigFanout.c", "aigUtil.c", "aigFrames.c", "../../misc/tim/timMan.c", "../../misc/mem/mem.c")
	if got != want.String() {
		c, goRows := strings.Split(got, "\n"), strings.Split(want.String(), "\n")
		for i, row := range c {
			if i >= len(goRows) {
				t.Fatalf("missing Go row %d", i)
			}
			if row != goRows[i] {
				t.Fatalf("frame matrix differs at row %d:\nC  %s\nGo %s", i, row, goRows[i])
			}
		}
		t.Fatal("frame matrix output length differs")
	}
	if !reflect.DeepEqual(before, p.Nodes()) {
		t.Fatal("frame expansion modified source objects")
	}
}
