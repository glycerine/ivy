package abci

import (
	"testing"

	aiggraph "github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	baseabc "github.com/glycerine/abcgo/base/abc"
)

func bmTestNetwork(g *aiggraph.Graph) *Abc_Ntk_t {
	manager := gia.Gia_ManFromAig(g)
	defer gia.Gia_ManStop(manager)
	return baseabc.Abc_NtkCreateFromGias("bm-test", nil, manager)
}

func TestBmDependenciesUseExactFunctionalSupport(t *testing.T) {
	g := aiggraph.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	g.AddOutput(b)
	g.AddOutput(g.OR(a, c))
	n := bmTestNetwork(g)
	iDep := []*Vec_Int_t{Vec_IntAlloc(1), Vec_IntAlloc(1), Vec_IntAlloc(1)}
	oDep := []*Vec_Int_t{Vec_IntAlloc(1), Vec_IntAlloc(1), Vec_IntAlloc(1)}
	defer func() {
		for _, v := range iDep {
			Vec_IntFree(v)
		}
		for _, v := range oDep {
			Vec_IntFree(v)
		}
		baseabc.Abc_NtkDelete(n)
	}()
	getDependencies(n, iDep, oDep)
	wantOutputs := [][]int32{{0, 1}, {1}, {0, 2}}
	for i, want := range wantOutputs {
		if got := Vec_IntSize(oDep[i]); got != int32(len(want)) {
			t.Fatalf("output %d support size = %d, want %d", i, got, len(want))
		}
		for j, value := range want {
			if got := Vec_IntEntry(oDep[i], int32(j)); got != value {
				t.Fatalf("output %d support[%d] = %d, want %d", i, j, got, value)
			}
		}
	}
}

func TestBmCofactorSubstitutesFixedAndRetainsMatchedInputs(t *testing.T) {
	g := aiggraph.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	n := bmTestNetwork(g)
	defer baseabc.Abc_NtkDelete(n)
	inputs := Vec_IntAlloc(1)
	Vec_IntPush(inputs, 1)
	defer Vec_IntFree(inputs)
	cofactor := computeCofactor(n, nil, []int32{1, 0}, inputs)
	defer baseabc.Abc_NtkDelete(cofactor)
	result, err := baseabc.Abc_NtkStrashToGia(cofactor).Evaluate([]bool{false})
	if err != nil {
		t.Fatal(err)
	}
	if result[0] {
		t.Fatal("cofactor output should be false when retained input is false")
	}
	result, err = baseabc.Abc_NtkStrashToGia(cofactor).Evaluate([]bool{true})
	if err != nil {
		t.Fatal(err)
	}
	if !result[0] {
		t.Fatal("cofactor output should be true when retained input is true")
	}
}

func TestBmSatReportsCounterexampleAndCleansMiter(t *testing.T) {
	left := aiggraph.NewGraph()
	a, b := left.NewInput(), left.NewInput()
	left.AddOutput(a)
	right := aiggraph.NewGraph()
	x, y := right.NewInput(), right.NewInput()
	right.AddOutput(y)
	n1, n2 := bmTestNetwork(left), bmTestNetwork(right)
	defer baseabc.Abc_NtkDelete(n1)
	defer baseabc.Abc_NtkDelete(n2)
	for attempt := 0; attempt < 2; attempt++ {
		mismatch := Vec_IntAlloc(4)
		status := Abc_NtkBmSat(n1, n2, nil, nil, mismatch, 0)
		if status != 0 {
			Vec_IntFree(mismatch)
			t.Fatalf("attempt %d: BMC SAT status = %d, want counterexample status 0", attempt, status)
		}
		if Vec_IntSize(mismatch) == 0 {
			Vec_IntFree(mismatch)
			t.Fatalf("attempt %d: expected a counterexample support", attempt)
		}
		Vec_IntFree(mismatch)
	}
	_ = b
	_ = x
}
