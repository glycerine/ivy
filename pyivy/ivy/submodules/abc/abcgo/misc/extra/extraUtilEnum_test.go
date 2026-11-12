package extra

import (
	"bytes"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// Counters from the original extraUtilEnum.c recursion compiled as standalone C.
// Both AND-only and AND/XOR runs exercise filter ordering and backtracking.
func TestEnumOriginalCounters(t *testing.T) {
	cases := []struct {
		v, g, x                 int32
		tries, builds, finished uint64
	}{
		{2, 1, 0, 1, 1, 1}, {2, 2, 0, 9, 1, 0}, {2, 3, 0, 24, 5, 1}, {2, 4, 0, 88, 5, 0},
		{3, 2, 0, 3, 3, 2}, {3, 3, 0, 63, 17, 10}, {3, 4, 0, 584, 126, 57},
		{4, 3, 0, 18, 18, 12}, {4, 4, 0, 626, 254, 180},
		{2, 1, 1, 2, 2, 2}, {2, 2, 1, 22, 2, 0}, {2, 3, 1, 46, 6, 0}, {2, 4, 1, 126, 6, 0},
		{3, 2, 1, 8, 8, 6}, {3, 3, 1, 220, 95, 75}, {3, 4, 1, 3295, 1086, 831},
		{4, 3, 1, 74, 74, 56}, {4, 4, 1, 3534, 2114, 1742},
	}
	for _, c := range cases {
		p := &Abc_EnuMan_t{nVars: c.v, nNodes: c.v, nTops: c.v, nNodeMax: c.v + c.g, fUseXor: c.x}
		copy(p.pTruths[:], enumElementaryTruths[:c.v])
		Abc_EnumerateFuncs_rec(p, 1, 0)
		if p.nTries != c.tries || p.nBuilds != c.builds || p.nFinished != c.finished {
			t.Fatalf("%+v: got %d/%d/%d", c, p.nTries, p.nBuilds, p.nFinished)
		}
		if p.nNodes != c.v || p.nTops != c.v {
			t.Fatal("backtracking changed initial frontier")
		}
		for _, r := range p.pRefs {
			if r != 0 {
				t.Fatal("reference leak")
			}
		}
	}
	if enumAigGraphs(4, 1, 0) != nil {
		t.Fatal("infeasible graph accepted")
	}
}

func TestEnumDatasets(t *testing.T) {
	for id, data := range enumDatasets {
		var variables, minterms, functions [100]uint32
		n, m, f := enumDecodeDataset(id, variables[:], minterms[:], functions[:])
		if n != data.variables || int(m) != len(data.minterms) || int(f) != len(data.functions) {
			t.Fatal("dataset sizes")
		}
		for i, row := range data.minterms {
			for k, ch := range row {
				want := ch == '1'
				if (variables[k]>>i&1 != 0) != want || (minterms[i]>>k&1 != 0) != want {
					t.Fatal("transpose mismatch")
				}
			}
		}
		for i, row := range data.functions {
			for k, ch := range row {
				if (functions[i]>>k&1 != 0) != (ch == '1') {
					t.Fatal("target mismatch")
				}
			}
		}
		// C accumulates bits rather than clearing caller storage.
		variables[0] = ^uint32(0)
		enumDecodeDataset(id, variables[:], minterms[:], functions[:])
		if variables[0] != ^uint32(0) {
			t.Fatal("dataset overwrote bits")
		}
	}
}

func TestEnumFormulaTruthAndCost(t *testing.T) {
	variables := []uint32{0xaa, 0xcc, 0xf0}
	result := enumFunctions(variables, nil, 4, false)
	// Independent exhaustive formula-cost relaxation, allowing repeated operands.
	costs := map[uint32]int{0xaa: 0, 0xcc: 0, 0xf0: 0}
	for cost := 1; cost <= 4; cost++ {
		additions := map[uint32]int{}
		for a, ca := range costs {
			for b, cb := range costs {
				if ca+cb+1 != cost {
					continue
				}
				for _, v := range []uint32{a & b, a | b} {
					if _, ok := costs[v]; !ok {
						additions[v] = cost
					}
				}
			}
		}
		for v, c := range additions {
			costs[v] = c
		}
	}
	if len(result.truths) != len(costs) {
		t.Fatalf("enumerated %d, want %d", len(result.truths), len(costs))
	}
	for i, value := range result.truths {
		cost := 0
		for cost+1 < len(result.gateCounts) && i >= result.gateCounts[cost+1] {
			cost++
		}
		if costs[value] != cost {
			t.Fatalf("truth %x cost %d want %d", value, cost, costs[value])
		}
		if i < len(variables) {
			continue
		}
		a, b := int(vec.Vec_IntEntry(result.gates, 2*i)), int(vec.Vec_IntEntry(result.gates, 2*i+1))
		var computed uint32
		if a < b {
			computed = result.truths[a] & result.truths[b]
		} else {
			computed = result.truths[a] | result.truths[b]
		}
		if computed != value {
			t.Fatal("formula does not implement stored truth")
		}
	}
	target := enumFunctions(variables, []uint32{0xaa, 0x88, 0xfe}, 3, false)
	for i, cost := range []int{0, 1, 2} {
		if target.found[i].index < 0 || target.found[i].gates != cost {
			t.Fatalf("target %d: %+v", i, target.found[i])
		}
	}
	var out bytes.Buffer
	enumPrintNode(&out, target.gates, int32(target.found[1].index), 3)
	if out.String() != "ab" {
		t.Fatalf("printed %q", out.String())
	}
	out.Reset()
	enumPrintNode(&out, target.gates, 0, 3)
	if out.String() != "a" {
		t.Fatal("input printer")
	}
}

func TestEnumDimensionOverflow(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("invalid gate count accepted")
		}
	}()
	enumAigGraphs(2, 1<<31-1, 0)
}
