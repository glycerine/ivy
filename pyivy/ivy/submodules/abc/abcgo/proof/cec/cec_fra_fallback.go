package cec

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/fracomb"
	"github.com/glycerine/abcgo/proof/framgr"
)

// cecVerifyFra runs the original Fra_FraigCec fallback on a dual-output
// miter and independently replays any returned model on the original graph.
func cecVerifyFra(graph *Gia_Man_t, verbose int32) (int32, int32, error) {
	if graph == nil || graph.RegisterCount() != 0 || gia.Gia_ManPoNum(graph)%2 != 0 {
		return -1, -1, fmt.Errorf("cec: FRA fallback requires a combinational dual-output miter")
	}
	graph.ClearCombinationalCounterexample()
	transformed := gia.Gia_ManTransformMiter(graph)
	defer gia.Gia_ManStop(transformed)
	miter := gia.Gia_ManToAig(transformed, 0)
	status := fracomb.Fra_FraigCec(&miter, 10000000, verbose, framgr.SweepForCEC)
	if status != 0 {
		return status, -1, nil
	}
	if miter == nil {
		return -1, -1, fmt.Errorf("cec: FRA fallback returned no miter")
	}
	model, ok := miter.Data.([]int32)
	if !ok || len(model) != len(graph.InputLiterals()) {
		return -1, -1, fmt.Errorf("cec: FRA fallback returned no valid counterexample")
	}
	inputs := make([]bool, len(model))
	for i, value := range model {
		inputs[i] = value != 0
	}
	outputs, err := graph.Evaluate(inputs)
	if err != nil {
		return -1, -1, fmt.Errorf("cec: FRA counterexample replay: %w", err)
	}
	for i := 0; i < len(outputs); i += 2 {
		if outputs[i] != outputs[i+1] {
			if err := graph.SetCombinationalCounterexample(inputs, i/2); err != nil {
				return -1, -1, err
			}
			return 0, int32(i / 2), nil
		}
	}
	return -1, -1, fmt.Errorf("cec: FRA counterexample does not disprove an output pair")
}

var _ fracomb.Sweep = framgr.SweepForCEC
