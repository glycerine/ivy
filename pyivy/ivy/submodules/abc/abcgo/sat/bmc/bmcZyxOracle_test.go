package bmc

import (
	"fmt"
	"testing"

	"github.com/glycerine/abcgo/sat/bsat"
)

// simulateZyxIndependent evaluates the decoded topology directly, without the
// synthesis manager's evaluator or truth-table utilities.
func simulateZyxIndependent(t *testing.T, net *zyxNetwork, nInputs int) uint64 {
	t.Helper()
	var result uint64
	for assignment := 0; assignment < 1<<nInputs; assignment++ {
		values := make([]bool, nInputs+len(net.fanins))
		for i := 0; i < nInputs; i++ {
			values[i] = assignment&(1<<i) != 0
		}
		for i, fanin := range net.fanins {
			lut := 0
			for bit, id := range fanin {
				if id < 0 || id >= nInputs+i {
					t.Fatalf("node %d has invalid fanin %d", i, id)
				}
				if values[id] {
					lut |= 1 << bit
				}
			}
			if lut >= len(net.funcs[i]) {
				t.Fatalf("node %d lookup %d outside table", i, lut)
			}
			values[nInputs+i] = net.funcs[i][lut]
		}
		out := values[len(values)-1]
		if net.complemented {
			out = !out
		}
		if out {
			result |= uint64(1) << assignment
		}
	}
	return result
}

// A single 2-LUT must use both inputs, so exactly the functions depending on
// both variables are realizable. This oracle does not use solver internals.
func TestZyxAllTwoInputFunctions(t *testing.T) {
	for truth := 0; truth < 16; truth++ {
		for order := int32(0); order <= 1; order++ {
			p, err := newZyxSynthesis(&Bmc_EsPar_t{nVars: 2, nNodes: 1, nLutSize: 2, pTtStr: fmt.Sprintf("%x", truth), fOrderNodes: order})
			if err != nil {
				t.Fatal(err)
			}
			net, status := p.solve()
			dependsA := ((truth>>0)^(truth>>1))&1 != 0 || ((truth>>2)^(truth>>3))&1 != 0
			dependsB := ((truth>>0)^(truth>>2))&1 != 0 || ((truth>>1)^(truth>>3))&1 != 0
			possible := dependsA && dependsB
			if (status == bsat.Satisfiable) != possible {
				t.Fatalf("truth=%x ordered=%d status=%v want satisfiable=%v", truth, order, status, possible)
			}
			if possible {
				if net == nil {
					t.Fatalf("truth=%x: missing model", truth)
				}
				if got := simulateZyxIndependent(t, net, 2); got != uint64(truth) {
					t.Fatalf("truth=%x ordered=%d got=%x", truth, order, got)
				}
			}
		}
	}
}

// Enumerate every legal two-gate, two-LUT network using all three inputs.
func twoLutThreeInputOracle() [256]bool {
	var possible [256]bool
	for firstA := 0; firstA < 3; firstA++ {
		for firstB := firstA + 1; firstB < 3; firstB++ {
			remaining := 3 - firstA - firstB
			for f := 0; f < 16; f++ {
				for g := 0; g < 16; g++ {
					truth := 0
					for m := 0; m < 8; m++ {
						first := f >> (((m >> firstA) & 1) | (((m >> firstB) & 1) << 1)) & 1
						value := g >> (((m >> remaining) & 1) | (first << 1)) & 1
						truth |= value << m
					}
					possible[truth] = true
				}
			}
		}
	}
	return possible
}

func TestZyxThreeInputTruthTablesAgainstEnumeration(t *testing.T) {
	possible := twoLutThreeInputOracle()
	// Essential LUT constraints exclude degenerate gates, so check oracle's
	// impossible tables and selected realizable tables requiring all inputs.
	for _, truth := range []int{0x80, 0x96, 0xe8, 0xfe, 0x7f, 0x01, 0x69} {
		for _, ordered := range []int32{0, 1} {
			p, err := newZyxSynthesis(&Bmc_EsPar_t{nVars: 3, nNodes: 2, nLutSize: 2, pTtStr: fmt.Sprintf("%02x", truth), fOrderNodes: ordered})
			if err != nil {
				t.Fatal(err)
			}
			net, status := p.solve()
			if possible[truth] && status != bsat.Satisfiable || !possible[truth] && status != bsat.Unsatisfiable {
				t.Fatalf("truth=%02x ordered=%d status=%v exhaustive possible=%v", truth, ordered, status, possible[truth])
			}
			if net != nil {
				if got := simulateZyxIndependent(t, net, 3); got != uint64(truth) {
					t.Fatalf("truth=%02x got=%02x", truth, got)
				}
			}
		}
	}
}
