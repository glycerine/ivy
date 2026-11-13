package bmc

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

// TestZyxSevenInputWideTruth independently checks every assignment on both
// sides of the 64-bit truth-table boundary. Three 3-LUTs can implement the
// OR of seven inputs: two 3-input groups feed a third gate with the last PI.
func TestZyxSevenInputWideTruth(t *testing.T) {
	for _, tc := range []struct {
		name, truth string
		invert      bool
	}{
		{"or", "fffffffffffffffffffffffffffffffe", false},
		{"nor", "00000000000000000000000000000001", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := newZyxSynthesis(&Bmc_EsPar_t{nVars: 7, nNodes: 3, nLutSize: 3, pTtStr: tc.truth})
			if err != nil {
				t.Fatal(err)
			}
			network, status := p.solve()
			if status != bsat.Satisfiable || network == nil {
				t.Fatalf("expected realizable function, got %v", status)
			}
			for assignment := 0; assignment < 128; assignment++ {
				values := make([]bool, 7+len(network.fanins))
				for i := 0; i < 7; i++ {
					values[i] = assignment&(1<<i) != 0
				}
				for i, fanins := range network.fanins {
					index := 0
					for bit, id := range fanins {
						if id < 0 || id >= 7+i {
							t.Fatalf("node %d invalid fanin %d", i, id)
						}
						if values[id] {
							index |= 1 << bit
						}
					}
					values[7+i] = network.funcs[i][index]
				}
				result := values[len(values)-1] != network.complemented
				want := (assignment != 0) != tc.invert
				if result != want {
					t.Fatalf("assignment %d: got %v want %v", assignment, result, want)
				}
			}
		})
	}
}
