package pdr_test

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/proof/pdr"
	"testing"
	"time"
)

func publicGraph() *aig.Graph {
	g := aig.NewGraph()
	input := g.NewInput()
	latch := g.NewInput()
	g.AddOutput(g.Const0())
	g.AddOutput(g.Const1())
	g.AddOutput(latch)
	g.AddOutput(input)
	aig.Aig_ManSetRegNum(g, 1)
	return g
}
func TestPublicSolveAllWitnessStorage(t *testing.T) {
	for _, store := range []bool{false, true} {
		g := publicGraph()
		o := pdr.DefaultOptions()
		o.SolveAll = true
		o.StoreCounterexamples = store
		o.Silent = true
		callbacks := 0
		o.OnFailure = func(index int32, c *util.Abc_Cex_t) bool {
			callbacks++
			if (c != nil) != store {
				t.Errorf("output %d callback witness %v", index, c)
			}
			return false
		}
		result, err := pdr.Solve(g, &o)
		if err != nil {
			t.Fatal(err)
		}
		if result.Status != pdr.Failed || len(result.Outputs) != 3 || result.Outputs[0].Status != pdr.Proved || result.Outputs[1].Status != pdr.Failed || result.Outputs[2].Status != pdr.Failed || callbacks != 2 {
			t.Fatalf("unexpected result: %+v, callbacks %d", result, callbacks)
		}
		for _, out := range result.Outputs[1:] {
			if (out.Witness != nil) != store {
				t.Fatalf("store %v: witness %v", store, out.Witness)
			}
			if out.Witness != nil && saig.Saig_ManVerifyCex(g, out.Witness) != 1 {
				t.Fatal("invalid witness")
			}
		}
	}
}
func TestPublicSolveDefaultsAndCancellation(t *testing.T) {
	g := aig.NewGraph()
	input := g.NewInput()
	latch := g.NewInput()
	g.AddOutput(latch)
	g.AddOutput(input)
	aig.Aig_ManSetRegNum(g, 1)
	result, err := pdr.Solve(g, nil)
	if err != nil || result.Status != pdr.Failed || result.Outputs[0].Witness == nil {
		t.Fatalf("default result %+v: %v", result, err)
	}
	o := pdr.DefaultOptions()
	o.Silent = true
	o.RunID = 19
	calls := 0
	o.Stop = func(id int32) bool {
		if id != 19 {
			t.Errorf("run ID %d", id)
		}
		calls++
		return true
	}
	result, err = pdr.Solve(g, &o)
	if err != nil || result.Status != pdr.Unknown || result.Outputs[0].Status != pdr.Unknown || calls == 0 {
		t.Fatalf("cancel result %+v, calls %d: %v", result, calls, err)
	}
}
func TestPublicSolveRejectsInvalidOptions(t *testing.T) {
	if _, err := pdr.Solve(nil, nil); err == nil {
		t.Fatal("accepted nil graph")
	}
	o := pdr.DefaultOptions()
	o.Timeout = -time.Second
	if _, err := pdr.Solve(publicGraph(), &o); err == nil {
		t.Fatal("accepted negative timeout")
	}
	o = pdr.DefaultOptions()
	o.PerOutputTimeout = time.Second
	if _, err := pdr.Solve(publicGraph(), &o); err == nil {
		t.Fatal("accepted per-output timeout without SolveAll")
	}
}
func TestPublicSolveAllCancellation(t *testing.T) {
	g := publicGraph()
	o := pdr.DefaultOptions()
	o.SolveAll = true
	o.Silent = true
	o.OnFailure = func(int32, *util.Abc_Cex_t) bool { return true }
	result, err := pdr.Solve(g, &o)
	if err != nil || result.Status != pdr.Unknown {
		t.Fatalf("result %+v: %v", result, err)
	}
	if result.Outputs[0].Status != pdr.Proved || result.Outputs[1].Status != pdr.Failed || result.Outputs[2].Status != pdr.Unknown {
		t.Fatalf("outcomes %+v", result.Outputs)
	}
}
func TestPublicSolveNoOutputs(t *testing.T) {
	g := aig.NewGraph()
	o := pdr.DefaultOptions()
	o.Silent = true
	result, err := pdr.Solve(g, &o)
	if err != nil || result.Status != pdr.Proved || len(result.Outputs) != 0 {
		t.Fatalf("result %+v: %v", result, err)
	}
}

func TestPublicSolveAllStopCallback(t *testing.T) {
	g := aig.NewGraph()
	in := g.NewInput()
	latch := g.NewInput()
	g.AddOutput(g.Const0())
	g.AddOutput(latch)
	g.AddOutput(in)
	aig.Aig_ManSetRegNum(g, 1)
	o := pdr.DefaultOptions()
	o.SolveAll = true
	o.Silent = true
	o.Stop = func(int32) bool { return true }
	result, err := pdr.Solve(g, &o)
	if err != nil || result.Status != pdr.Unknown || result.Outputs[0].Status != pdr.Proved || result.Outputs[1].Status != pdr.Unknown {
		t.Fatalf("result %+v: %v", result, err)
	}
}
func TestPublicSolveAllNoOutputs(t *testing.T) {
	g := aig.NewGraph()
	o := pdr.DefaultOptions()
	o.SolveAll = true
	o.Silent = true
	result, err := pdr.Solve(g, &o)
	if err != nil || result.Status != pdr.Proved || len(result.Outputs) != 0 {
		t.Fatalf("result %+v: %v", result, err)
	}
}
