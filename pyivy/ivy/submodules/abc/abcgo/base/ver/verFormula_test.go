package ver

import (
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestFormulaParserTruthAndNames(t *testing.T) {
	for _, tc := range []struct {
		formula string
		want    func(bool, bool, bool) bool
	}{
		{"a | b & c", func(a, b, c bool) bool { return a || b && c }},
		{"a ? b : c", func(a, b, c bool) bool {
			if a {
				return b
			}
			return c
		}},
		{"~a ^ (b | c)", func(a, b, c bool) bool { return !a != (b || c) }},
	} {
		man := hop.Hop_ManStart()
		names := vec.Vec_PtrAlloc(0)
		functions := vec.Vec_PtrAlloc(0)
		operators := vec.Vec_IntAlloc(0)
		msg := make([]byte, 2000)
		root := Ver_FormulaParser(tc.formula, man, names, functions, operators, msg)
		if root == nil {
			t.Fatalf("%q: %s", tc.formula, msg)
		}
		truth := hop.Hop_ManConvertAigToTruth(man, root, 3, nil, 0)
		for bits := 0; bits < 8; bits++ {
			got := truth[0]&(1<<bits) != 0
			want := tc.want(bits&1 != 0, bits&2 != 0, bits&4 != 0)
			if got != want {
				t.Fatalf("%q input=%d got=%v want=%v", tc.formula, bits, got, want)
			}
		}
		hop.Hop_ManStop(man)
	}
}
func TestFormulaReductionTruth(t *testing.T) {
	man := hop.Hop_ManStart()
	defer hop.Hop_ManStop(man)
	names := vec.Vec_PtrAlloc(0)
	msg := make([]byte, 2000)
	root := Ver_FormulaReduction("~&{a,b,c}", man, names, msg)
	if root == nil {
		t.Fatalf("reduction: %s", msg)
	}
	truth := hop.Hop_ManConvertAigToTruth(man, root, 3, nil, 0)
	for bits := 0; bits < 8; bits++ {
		if got := truth[0]&(1<<bits) != 0; got != (bits != 7) {
			t.Fatalf("input=%d got=%v", bits, got)
		}
	}
}
