package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

// cecSolve.c's pattern visitor advances the graph traversal ID, marks the cone,
// counts each visited CI, and clears and reuses the same vector each call.
func TestSavePatternPreservesSourceTraversalAndVector(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	y := gia.Gia_ManAppendAnd(g, x, c)
	p := Cec_ManSatCreate(g, nil)
	defer Cec_ManSatStop(p)
	if status := Cec_ManSatCheckNode(p, gia.Gia_ManObj(g, y>>1)); status != 0 {
		t.Fatalf("status %d: %v", status, p.Error)
	}
	saved := Cec_ManSatReadCex(p)
	Cec_ManSavePattern(p, gia.Gia_ManObj(g, x>>1), gia.Gia_ManObj(g, y>>1))
	if Cec_ManSatReadCex(p) != saved || p.nCexLits != 3 {
		t.Fatalf("vector replaced or count %d", p.nCexLits)
	}
	for _, lit := range []int32{a, b, c, x, y} {
		if gia.Gia_ObjIsTravIdCurrent(g, gia.Gia_ManObj(g, lit>>1)) == 0 {
			t.Fatalf("object %d not visited", lit>>1)
		}
	}
	if got := vec.Vec_IntArray(saved); !reflect.DeepEqual(got, []int32{0, 2, 4}) {
		t.Fatalf("pattern %v", got)
	}
	Cec_ManSavePattern(p, gia.Gia_ManObj(g, x>>1), nil)
	if Cec_ManSatReadCex(p) != saved || p.nCexLits != 5 || !reflect.DeepEqual(vec.Vec_IntArray(saved), []int32{0, 2}) {
		t.Fatalf("second pattern %v count %d", vec.Vec_IntArray(saved), p.nCexLits)
	}
	if gia.Gia_ObjIsTravIdCurrent(g, gia.Gia_ManObj(g, c>>1)) != 0 {
		t.Fatal("old traversal leaked into new cone")
	}
}

func TestSatMiterInitializesSourcePhaseMetadata(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	node := gia.Gia_ManAppendAnd(g, a^1, b^1)
	gia.Gia_ManAppendCo(g, node)
	gia.Gia_ObjSetPhaseValue(gia.Gia_ManObj(g, node>>1), 0)
	var status *Vec_Str_t
	Cec_ManSatSolveMiter(g, nil, &status)
	if gia.Gia_ObjPhase(gia.Gia_ManObj(g, node>>1)) != 1 || gia.Gia_ObjPhase(gia.Gia_ManCo(g, 0)) != 1 {
		t.Fatal("source miter phase initialization missing")
	}
}

func TestSatConstantZeroPreservesSourceModelAndPattern(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	p := Cec_ManSatCreate(g, nil)
	defer Cec_ManSatStop(p)
	obj := gia.Gia_ManObj(g, a>>1)
	if Cec_ManSatCheckNode(p, obj) != 0 {
		t.Fatal("expected SAT")
	}
	Cec_ManSavePattern(p, obj, nil)
	vector := p.vCex
	before := append([]int32(nil), vec.Vec_IntArray(vector)...)
	source := p.source
	calls := source.calls
	values := append([]bool(nil), p.values...)
	if got := Cec_ManSatCheckNode(p, gia.Gia_ManConst0(g)); got != 1 {
		t.Fatalf("status %d", got)
	}
	if p.vCex != vector || !reflect.DeepEqual(vec.Vec_IntArray(vector), before) || !reflect.DeepEqual(p.values, values) || source.calls != calls {
		t.Fatal("constant query altered SAT state")
	}
}
