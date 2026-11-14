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

// ABC records names as length/name pairs and reuses the original index when
// an escaped identifier is referenced again.
func TestFormulaEscapedIdentifierReuse(t *testing.T) {
	man := hop.Hop_ManStart()
	defer hop.Hop_ManStop(man)
	names := vec.Vec_PtrAlloc(0)
	funcs := vec.Vec_PtrAlloc(0)
	ops := vec.Vec_IntAlloc(0)
	errbuf := make([]byte, 2000)
	root := Ver_FormulaParser(`\x.y  ^ \x.y `, man, names, funcs, ops, errbuf)
	if root == nil {
		t.Fatalf("parse failed: %s", errbuf)
	}
	if vec.Vec_PtrSize(names) != 2 || vec.Vec_PtrEntry(names, 0) != 3 || vec.Vec_PtrEntry(names, 1) != "x.y" {
		t.Fatalf("names %v", vec.Vec_PtrArray(names))
	}
	truth := hop.Hop_ManConvertAigToTruth(man, root, 1, nil, 0)
	if truth[0] != 0 {
		t.Fatalf("x xor x truth %#x", truth[0])
	}
}
