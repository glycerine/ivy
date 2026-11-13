package core

import (
	"math/rand"
	"testing"
)

// oracleValue interprets the test circuit directly, independently of Model and
// its evaluator. Test circuits are constructed in topological order.
func oracleValue(c *circuit, inputs []bool, lit int32) bool {
	values := make([]bool, len(c.nodes))
	for i, input := range c.inputs {
		values[input>>1] = inputs[i]
	}
	value := func(lit int32) bool { return values[lit>>1] != (lit&1 != 0) }
	for id, n := range c.nodes {
		switch n.kind {
		case And:
			values[id] = value(n.fanins[0]) && value(n.fanins[1])
		case Xor:
			values[id] = value(n.fanins[0]) != value(n.fanins[1])
		case Mux:
			if value(n.fanins[0]) {
				values[id] = value(n.fanins[1])
			} else {
				values[id] = value(n.fanins[2])
			}
		case Buffer, Output:
			values[id] = value(n.fanins[0])
		}
	}
	return value(lit)
}

func TestCrossModelProofsAgainstIndependentTruthTables(t *testing.T) {
	rng := rand.New(rand.NewSource(9832))
	for trial := 0; trial < 80; trial++ {
		left, right := newCircuit(3), newCircuit(3)
		// Equivalent cones with different first-use orders exercise shared PI
		// bindings; unrelated random cones then exercise unequal outputs.
		la := left.add(And, left.inputs[0], left.inputs[2], 0)
		ra := right.add(And, right.inputs[2], right.inputs[0], 0)
		left.outputs, right.outputs = []int32{la}, []int32{ra}
		for _, c := range []*circuit{left, right} {
			for i := 0; i < 12; i++ {
				lit := func() int32 { return int32(rng.Intn(2 * len(c.nodes))) }
				c.add([]int{And, Xor, Mux, Buffer}[rng.Intn(4)], lit(), lit(), lit())
			}
		}
		for i := 0; i < 3; i++ {
			left.outputs = append(left.outputs, int32(rng.Intn(2*len(left.nodes))))
			right.outputs = append(right.outputs, int32(rng.Intn(2*len(right.nodes))))
		}
		firstDifferent := -1
		for i := range left.outputs {
			for mask := 0; mask < 8; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				if oracleValue(left, inputs, left.outputs[i]) != oracleValue(right, inputs, right.outputs[i]) {
					firstDifferent = i
					break
				}
			}
			if firstDifferent >= 0 {
				break
			}
		}
		for _, flip := range []bool{false, true} {
			result, err := VerifyTwo(left, right, Options{PolarFlip: flip, MaxSATVariables: 1})
			if err != nil {
				t.Fatalf("trial %d: %v", trial, err)
			}
			if firstDifferent < 0 {
				if result.Status != Equivalent {
					t.Fatalf("trial %d: expected equivalent, got %+v", trial, result)
				}
				continue
			}
			w := result.Counterexample
			if result.Status != Different || w == nil || w.Output != firstDifferent || len(w.Inputs) != 3 {
				t.Fatalf("trial %d: expected difference at %d, got %+v", trial, firstDifferent, result)
			}
			x := oracleValue(left, w.Inputs, left.outputs[w.Output])
			y := oracleValue(right, w.Inputs, right.outputs[w.Output])
			if x == y || x != w.Left || y != w.Right {
				t.Fatalf("trial %d: invalid witness %+v", trial, w)
			}
		}
	}
}
