package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"slices"
	"testing"
)

func TestConstraintSourceCollectorAndWorkingGraph(t *testing.T) {
	p := aig.Aig_ManStart(20)
	a, b := aig.Aig_ObjCreateCi(p), aig.Aig_ObjCreateCi(p)
	for _, e := range []aig.Edge{a, a.Not(), b, b.Not(), b} {
		aig.Aig_ObjCreateCo(p, e)
	}
	aig.Aig_ManSetRegNum(p, 1)
	n := &Network{Graph: p, OutputNames: []string{"csLiveConst_a", "csLiveTarget_first", "csLiveConst_b", "csLiveTarget_last", "next"}}
	wantSignals := []aig.Edge{a, b, b.Not()}
	for _, collect := range []func(*Network, *aig.Graph) []aig.Edge{collectCSSignals, collectCSSignalsWithDSC} {
		if got := collect(n, p); !slices.Equal(got, wantSignals) {
			t.Fatalf("collected %v, want %v", got, wantSignals)
		}
	}
	index := int32(-1)
	result := generateWorkingAig(p, n, &index)
	if index != 4 || trueOutputs(result) != 5 {
		t.Fatalf("index %d outputs %d", index, trueOutputs(result))
	}
	for bits := 0; bits < 4; bits++ {
		a, b := bits&1 != 0, bits&2 != 0
		got, err := result.Evaluate([]bool{a, b})
		if err != nil {
			t.Fatal(err)
		}
		if got[index] != (!(a && b) || !b) {
			t.Fatalf("a=%v b=%v target=%v", a, b, got[index])
		}
	}
}
