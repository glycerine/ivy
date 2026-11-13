package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSweepOptionAdaptersPreserveLogicAndCallback(t *testing.T) {
	for _, kind := range []string{"plain", "mapped", "callback"} {
		t.Run(kind, func(t *testing.T) {
			g := aig.NewGraph()
			a, b := g.NewInput(), g.NewInput()
			gate := g.AND(a, b)
			g.AddOutput(gate)
			var options CombOptions
			Fra_ParamsDefault(&options)
			options.FConeBias = 0
			var result *aig.Graph
			var mapping []aig.Edge
			called := false
			switch kind {
			case "plain":
				result = Sweep(g, options)
			case "mapped":
				result, mapping = SweepMapped(g, options)
			default:
				result, mapping = SweepMappedCallback(g, options, func(p *Manager) {
					called = true
					if p.PManAig != g || p.PManFraig == nil {
						t.Error("callback lacks live manager")
					}
				})
			}
			if result == nil {
				t.Fatal("missing result")
			}
			if kind == "callback" && !called {
				t.Error("callback not invoked")
			}
			if kind != "plain" && (len(mapping) <= gate.NodeID() || mapping[gate.NodeID()] == aig.InvalidEdge) {
				t.Fatalf("missing gate mapping: %v", mapping)
			}
			for mask := 0; mask < 4; mask++ {
				values := []bool{mask&1 != 0, mask&2 != 0}
				out, err := result.Evaluate(values)
				if err != nil || len(out) != 1 || out[0] != (values[0] && values[1]) {
					t.Fatalf("input=%v output=%v err=%v", values, out, err)
				}
			}
		})
	}
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	called := false
	var options CombOptions
	Fra_ParamsDefault(&options)
	result, _ := SweepMappedCallback(g, options, func(*Manager) { called = true })
	if result == nil || called {
		t.Fatal("source fast path should copy graph without invoking callback")
	}
}
