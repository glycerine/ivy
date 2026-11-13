package fracomb

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	dar "github.com/glycerine/abcgo/opt/dar/core"
)

// fraCec.c balances XOR-heavy rewritten miters before invoking the first
// FRAIG sweep. This miter is UNSAT but exceeds the initial 600-conflict limit.
func TestCecSourceXorHeavyFallbackBalancesBeforeSweep(t *testing.T) {
	g := cecPigeonholeMiter()
	out := g.Outputs()[0]
	base := g.Node(out).Fanins[0]
	inputs := g.Inputs()
	parity := inputs[0]
	for i := 1; i < 50; i++ {
		parity = g.XOR(parity, inputs[i])
	}
	aig.Aig_ObjPatchFanin0(g, out, g.AND(base, parity))
	if status := Fra_FraigSat(g, 600, 0, 0, 0, 0, 1, 0, 0, 0); status != -1 {
		t.Fatalf("initial SAT status %d, want undecided", status)
	}
	rewritten, err := dar.Dar_ManRwsat(g, true, false)
	if err != nil || rewritten == nil {
		t.Fatalf("rewrite=%v err=%v", rewritten, err)
	}
	if Aig_ManCountXors(rewritten)*30 <= aig.Aig_ManNodeNum(rewritten)+300 {
		t.Fatal("fixture does not reach C XOR-balance branch")
	}
	expected := dar.Dar_ManBalanceXor(rewritten, true, false, false)
	calls := 0
	status := Fra_FraigCec(&g, 100000, 0, func(graph *aig.Graph, nodeLimit, miterLimit int32) *aig.Graph {
		calls++
		if nodeLimit != 2 || miterLimit != 300 {
			t.Fatalf("first sweep limits=%d,%d", nodeLimit, miterLimit)
		}
		if !reflect.DeepEqual(graph.Nodes(), expected.Nodes()) || !reflect.DeepEqual(graph.Outputs(), expected.Outputs()) {
			t.Fatal("sweep did not receive source-balanced graph")
		}
		result := aig.Aig_ManDupDfs(graph)
		result.Data = make([]int32, result.InputCount())
		return result
	})
	if status != 0 || calls != 1 {
		t.Fatalf("status=%d sweeps=%d", status, calls)
	}
}
