package fra

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/opt/dar"
)

// The scalar counting oracle follows the bit-by-bit loop in ABC's
// Fra_ClausProcessClausesCut2; its optimized Cut variant must match it.
func TestClausCutPatternsExcludePrefix(t *testing.T) {
	g := aig.NewGraph()
	leaves := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
	s := Fra_SmlStart(g, 0, 2, 8)
	p := &Clu_Man_t{nSimWordsPref: 8}
	for j, e := range leaves {
		words := s.NodeWords(e.NodeID())
		for i := 0; i < 8; i++ {
			words[i] = ^uint32(0)
		}
		for i := 8; i < 16; i++ {
			words[i] = []uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00}[j]
		}
	}
	for count := 2; count <= 4; count++ {
		cut := &dar.Cut{Leaves: leaves[:count]}
		var want [16]int32
		for w := 8; w < 16; w++ {
			for bit := uint(0); bit < 32; bit++ {
				pattern := 0
				for j, e := range cut.Leaves {
					pattern |= int((s.NodeWords(e.NodeID())[w]>>bit)&1) << j
				}
				want[pattern]++
			}
		}
		var mask int32
		for i, n := range want {
			if n != 0 {
				mask |= 1 << i
			}
		}
		for _, run := range []struct {
			name string
			fn   func(*Clu_Man_t, *Fra_Sml_t, *dar.Cut, []int32) int32
		}{{"transpose", Fra_ClausProcessClausesCut}, {"scalar", Fra_ClausProcessClausesCut2}} {
			scores := make([]int32, 16)
			for i := range scores {
				scores[i] = -1
			}
			if got := run.fn(p, s, cut, scores); got != mask {
				t.Fatalf("%s %d leaves mask %#x, want %#x", run.name, count, got, mask)
			}
			for i, value := range scores {
				if value != want[i] {
					t.Fatalf("%s %d leaves score %d = %d, want %d", run.name, count, i, value, want[i])
				}
			}
		}
	}
}

// fraClaus.c ignores the prefix and checks a & ^b for implication and a & b
// for implication to the complement.
func TestClausSimulatedImplicationsRespectPrefix(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	s := Fra_SmlStart(g, 1, 2, 1)
	aw, bw := s.NodeWords(a.NodeID()), s.NodeWords(b.NodeID())
	aw[0], bw[0] = 1, 0 // prefix alone violates a -> b
	aw[1], bw[1] = 3, 7
	if got := Fra_ClausSmlNodesAreImp(s, a, b); got != 1 {
		t.Fatalf("implication = %d", got)
	}
	if got := Fra_ClausSmlNodesAreImpC(s, a, b); got != 0 {
		t.Fatalf("complement implication = %d", got)
	}
	aw[1], bw[1] = 8, 7
	if got := Fra_ClausSmlNodesAreImp(s, a, b); got != 0 {
		t.Fatalf("implication = %d", got)
	}
	if got := Fra_ClausSmlNodesAreImpC(s, a, b); got != 1 {
		t.Fatalf("complement implication = %d", got)
	}
}
