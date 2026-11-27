package hop

import (
	"strings"
	"testing"
)

func hopPrintVerilogString(f *Hop_Obj_t) string {
	var b strings.Builder
	Hop_ObjPrintVerilog(&b, f, nil, 0, 0)
	return b.String()
}

// TestHopObjPrintVerilogLikeABC: Hop_ObjPrintVerilog collapses multi-input
// ANDs, pushes a complement through them as an OR (De Morgan) and only
// parenthesizes below the top level.
func TestHopObjPrintVerilogLikeABC(t *testing.T) {
	p := Hop_ManStart()
	names := []string{"a", "b", "c"}
	for i := range names {
		Hop_ObjSetData(Hop_IthVar(p, int32(i)), names[i])
	}
	a, b, c := Hop_IthVar(p, 0), Hop_IthVar(p, 1), Hop_IthVar(p, 2)
	for _, tc := range []struct {
		f    *Hop_Obj_t
		want string
	}{
		{Hop_And(p, a, b), "a & b"},
		{Hop_Not(Hop_And(p, a, b)), "~a | ~b"},
		{Hop_And(p, Hop_And(p, a, b), c), "a & b & c"},
		{Hop_And(p, Hop_Not(Hop_And(p, a, b)), c), "(~a | ~b) & c"},
		{Hop_Not(p.pConst1), "1'b0"},
	} {
		got := hopPrintVerilogString(tc.f)
		// fanin order follows Hop's node ordering; accept any permutation
		// of the top-level operands but require ABC's shape.
		if got != tc.want && !samePermutation(got, tc.want) {
			t.Errorf("got %q, want %q", got, tc.want)
		}
	}
}

func samePermutation(got, want string) bool {
	for _, op := range []string{" & ", " | "} {
		if strings.Contains(want, op) {
			g, w := strings.Split(got, op), strings.Split(want, op)
			if len(g) != len(w) {
				return false
			}
			seen := map[string]int{}
			for _, x := range w {
				seen[x]++
			}
			for _, x := range g {
				seen[x]--
			}
			for _, n := range seen {
				if n != 0 {
					return false
				}
			}
			return true
		}
	}
	return false
}
