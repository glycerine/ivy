package cec

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestNativeSATWitnessAndComplement(t *testing.T) {
	graph := gia.Gia_ManStart(16)
	x := gia.Gia_ManAppendCi(graph)
	y := gia.Gia_ManAppendCi(graph)
	a := gia.Gia_ManAppendAnd(graph, x, y)
	b := gia.Gia_ManAppendAnd(graph, x, y^1)
	zero := gia.Gia_ManAppendAnd(graph, a, b)
	p := Cec_ManSatCreate(graph, nil)
	if p.Error != nil {
		t.Fatal(p.Error)
	}
	if got := Cec_ManSatCheckNode(p, gia.Gia_ManObj(graph, zero>>1)); got != 1 {
		t.Fatalf("contradiction status %d: %v", got, p.Error)
	}
	obj := gia.Gia_ManObj(graph, x>>1)
	if got := Cec_ManSatCheckNode(p, gia.Gia_Not(obj)); got != 0 {
		t.Fatalf("negated input status %d", got)
	}
	if Cec_ObjSatVarValue(p, obj) != 0 {
		t.Fatal("negated input witness has positive input")
	}
	if got := Cec_ManSatCheckNodeTwo(p, obj, obj); got != 0 || Cec_ObjSatVarValue(p, obj) != 1 {
		t.Fatalf("identical pair status %d", got)
	}
	if got := Cec_ManSatCheckNodeTwo(p, obj, gia.Gia_Not(obj)); got != 1 {
		t.Fatalf("complementary pair status %d", got)
	}
	// The source returns from this trivial case without touching the previous model.
	if p.values == nil || Cec_ObjSatVarValue(p, obj) != 1 {
		t.Fatal("trivial query unexpectedly changed the previous SAT witness")
	}
	p.pPars.nBTLimit = -1
	if got := Cec_ManSatCheckNode(p, obj); got != -1 || p.Error == nil {
		t.Fatal("invalid budget not reported")
	}
}

func TestNativeSATVariableValueUsesSourcePolarity(t *testing.T) {
	for _, flip := range []int32{0, 1} {
		g := gia.Gia_ManStart(8)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		n := gia.Gia_ManAppendAnd(g, a^1, b^1)
		gia.Gia_ManSetPhase(g)
		pars := new(Cec_ParSat_t)
		Cec_ManSatSetDefaultParams(pars)
		pars.fPolarFlip = flip
		p := Cec_ManSatCreate(g, pars)
		obj := gia.Gia_ManObj(g, n>>1)
		if status := Cec_ManSatCheckNode(p, obj); status != 0 {
			t.Fatalf("flip %d: status %d", flip, status)
		}
		if logical, physical := Cec_ObjLogicalValue(p, obj), Cec_ObjSatVarValue(p, obj); logical != 1 || physical != 1-flip {
			t.Fatalf("flip %d: logical %d physical %d", flip, logical, physical)
		}
	}
}

func TestNativeVerifyAndWitnessOwnership(t *testing.T) {
	left, right := gia.Gia_ManStart(8), gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(left), gia.Gia_ManAppendCi(right)
	gia.Gia_ManAppendCo(left, a)
	gia.Gia_ManAppendCo(right, b^1)
	if got := Cec_ManVerifyTwo(left, right, 0); got != 0 {
		t.Fatalf("mismatch status %d", got)
	}
	witness := left.CombinationalCounterexample()
	if witness == nil || right.CombinationalCounterexample() != nil {
		t.Fatal("witness ownership")
	}
	if got := Cec_ManVerifyTwo(left, left, 0); got != 1 {
		t.Fatalf("self comparison status %d", got)
	}
	if left.CombinationalCounterexample() != nil || witness == nil {
		t.Fatal("stale witness not cleared")
	}
}

func TestNativeDualOutputVerification(t *testing.T) {
	graph := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(graph)
	gia.Gia_ManAppendCo(graph, x)
	gia.Gia_ManAppendCo(graph, x^1)
	params := new(Cec_ParCec_t)
	Cec_ManCecSetDefaultParams(params)
	if got := Cec_ManVerify(graph, params); got != 0 || params.Error != nil || params.iOutFail != 0 {
		t.Fatalf("dual miter status %d, params %+v", got, params)
	}
}

func TestNativeSpecialCases(t *testing.T) {
	graph := gia.Gia_ManStart(16)
	x, y := gia.Gia_ManAppendCi(graph), gia.Gia_ManAppendCi(graph)
	gia.Gia_ManAppendCo(graph, x)
	gia.Gia_ManAppendCo(graph, y)
	params := new(Cec_ParCec_t)
	Cec_ManCecSetDefaultParams(params)
	if got := Cec_ManHandleSpecialCases(graph, params); got != 0 || graph.CombinationalCounterexample() == nil {
		t.Fatalf("PI mismatch %d", got)
	}
	graph = gia.Gia_ManStart(16)
	x, y = gia.Gia_ManAppendCi(graph), gia.Gia_ManAppendCi(graph)
	both := gia.Gia_ManAppendAnd(graph, x, y)
	gia.Gia_ManAppendCo(graph, both)
	gia.Gia_ManAppendCo(graph, both)
	if got := Cec_ManHandleSpecialCases(graph, params); got != -1 {
		t.Fatalf("nontrivial proof should remain, got %d", got)
	}
}

func TestNativeMiterStatusesAndCube(t *testing.T) {
	graph := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(graph)
	gia.Gia_ManAppendCo(graph, 0)
	gia.Gia_ManAppendCo(graph, x^1)
	var status *Vec_Str_t
	store := Cec_ManSatSolveMiter(graph, nil, &status)
	if got := vec.Vec_StrArray(status); len(got) != 2 || got[0] != 1 || got[1] != 0 {
		t.Fatalf("statuses %v", got)
	}
	// Output 1 has the cube input 0=false: literal 1.
	if got := vec.Vec_IntArray(store); !reflect.DeepEqual(got, []int32{1, 1, 1}) {
		t.Fatalf("cube store %v", got)
	}
}

func TestPublicVerifyTwoAigs(t *testing.T) {
	left, right := aig.NewGraph(), aig.NewGraph()
	a, b := left.NewInput(), right.NewInput()
	left.AddOutput(a)
	right.AddOutput(b.Not())
	if got := Cec_ManVerifyTwoAigs(left, right, 0); got != 0 {
		t.Fatalf("AIG mismatch status %d", got)
	}
	model := left.CombinationalCounterexample()
	if model == nil || right.CombinationalCounterexample() != nil {
		t.Fatal("AIG mismatch model ownership")
	}
	if got := Cec_ManVerifyTwoAigs(left, left, 0); got != 1 || left.CombinationalCounterexample() != nil {
		t.Fatalf("AIG self proof/stale witness %d", got)
	}
}

func TestPublicFraigActuallyReduces(t *testing.T) {
	graph := aig.NewGraph()
	x, y, z := graph.NewInput(), graph.NewInput(), graph.NewInput()
	left, right := graph.AND(graph.AND(x, y), z), graph.AND(x, graph.AND(y, z))
	graph.AddOutput(left)
	graph.AddOutput(right)
	reduced := Cec_FraigCombinational(graph, 1000, 0)
	if reduced == nil || aig.Aig_ManAndNum(reduced) >= aig.Aig_ManAndNum(graph) {
		t.Fatal("public fraig did not reduce equivalent cones")
	}
	if got := Cec_ManVerifyTwoAigs(graph, reduced, 0); got != 1 {
		t.Fatalf("public fraig changed outputs %d", got)
	}
}

func TestSourceNodeConstantOneAssertionAndMiterShortcut(t *testing.T) {
	graph := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(graph)
	gia.Gia_ManAppendCo(graph, 0)
	gia.Gia_ManAppendCo(graph, 1)
	var status *Vec_Str_t
	store := Cec_ManSatSolveMiter(graph, nil, &status)
	if got := vec.Vec_StrArray(status); !reflect.DeepEqual(got, []byte{1, 0}) {
		t.Fatalf("statuses %v", got)
	}
	if got := vec.Vec_IntArray(store); !reflect.DeepEqual(got, []int32{1, 0}) {
		t.Fatalf("store %v", got)
	}
	p := Cec_ManSatCreate(graph, nil)
	defer Cec_ManSatStop(p)
	defer func() {
		if recover() == nil {
			t.Fatal("constant-one source node check must assert")
		}
	}()
	Cec_ManSatCheckNode(p, gia.Gia_Not(gia.Gia_ManConst0(graph)))
}

func TestSourceConjunctionTrivialCasesAndConstantAssertion(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	p := Cec_ManSatCreate(g, nil)
	defer Cec_ManSatStop(p)
	zero := gia.Gia_ManConst0(g)
	one := gia.Gia_Not(zero)
	x := gia.Gia_ManObj(g, a>>1)
	for _, pair := range [][2]*Gia_Obj_t{{zero, x}, {x, zero}, {x, gia.Gia_Not(x)}, {zero, one}} {
		if got := Cec_ManSatCheckNodeTwo(p, pair[0], pair[1]); got != 1 {
			t.Fatalf("trivial result %d", got)
		}
	}
	if got := p.source.calls; got != 0 {
		t.Fatalf("trivial queries reached solver: %d", got)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("two constant-one inputs must assert")
		}
	}()
	Cec_ManSatCheckNodeTwo(p, one, one)
}
