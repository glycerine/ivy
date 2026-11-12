package sbd

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"slices"
	"testing"
)

func TestFixedTopologyFunctionSynthesis(t *testing.T) {
	var topology LUTTopology
	topology[0][0], topology[0][1] = 0, 1
	for function := uint64(0); function < 16; function++ {
		target := function * 0x1111111111111111
		values := make([]int32, 3)
		phase := Sbd_SolverFunc(2, 1, 2, &topology, []uint64{target}, values)
		if phase < 0 {
			t.Fatalf("two-input function%x rejected", function)
		}
		got := Sbd_SolverTruth(2, 1, 2, &topology, values)
		if phase != 0 {
			got = ^got
		}
		if got != target {
			t.Fatalf("function%x reconstructed%x", function, got)
		}
	}
	topology[1][0], topology[1][1] = 2, 3
	values := make([]int32, 6)
	and3 := uint64(0x8080808080808080)
	phase := Sbd_SolverFunc(3, 2, 2, &topology, []uint64{and3}, values)
	if phase != 0 || Sbd_SolverTruth(3, 2, 2, &topology, values) != and3 {
		t.Fatal("two-LUT AND3 synthesis")
	}
	if Sbd_SolverFunc(3, 2, 2, &topology, []uint64{0xe8e8e8e8e8e8e8e8}, values) != -1 {
		t.Fatal("two-input cascade implements majority3 unexpectedly")
	}
}
func TestMultiwordTopologySynthesis(t *testing.T) {
	var topology LUTTopology
	for i, pair := range [][2]int32{{0, 1}, {2, 3}, {4, 5}, {6, 7}, {8, 9}, {10, 11}, {12, 13}} {
		topology[i][0], topology[i][1] = pair[0], pair[1]
	}
	target := []uint64{0, 0, 0, 1 << 63}
	values := make([]int32, 21)
	phase := Sbd_SolverFunc(8, 7, 2, &topology, target, values)
	if phase < 0 {
		t.Fatal("AND8 rejected")
	}
	got := Sbd_SolverTruthWord(8, 7, 2, &topology, values, sbdTruthStorage(8, 7), phase)
	if !slices.Equal(got, target) {
		t.Fatal("multiword reconstruction")
	}
}
func TestTopologyConnectivityAndDelayBounds(t *testing.T) {
	for _, test := range []struct {
		m, n, req int32
		delays    []int32
		sat       bool
	}{{2, 1, 1, []int32{0, 0}, true}, {2, 1, 0, []int32{0, 0}, false}, {3, 1, 2, []int32{0, 0, 0}, false}, {3, 2, 2, []int32{0, 0, 0}, true}, {2, 1, 1, []int32{4, 4}, false}, {2, 1, 5, []int32{4, 4}, true}} {
		var variables TopologyVariables
		var delays DelayVariables
		solver, _ := Sbd_SolverTopo(test.m, test.n, 2, &variables, &delays, test.delays, test.req)
		if (solver.Solve() == bsat.Satisfiable) != test.sat {
			t.Fatalf("topology %+v", test)
		}
	}
}
func TestRandomTopologyUsesEveryNode(t *testing.T) {
	for m := int32(2); m <= 8; m++ {
		for n := m - 1; n <= 30; n++ {
			var topology LUTTopology
			Sbd_SolverSynth(m, n, 2, &topology)
			used := make([]bool, m+n)
			for node := int32(0); node < n; node++ {
				a, b := topology[node][0], topology[node][1]
				if a == b || a < 0 || b < 0 || a >= m+node || b >= m+node {
					t.Fatal("invalid generated topology")
				}
				used[a], used[b] = true, true
			}
			for _, value := range used[:len(used)-1] {
				if !value {
					t.Fatalf("disconnected m%d n%d", m, n)
				}
			}
		}
	}
}
func TestParameterizedSetupHasDistinctOutputAndLoads(t *testing.T) {
	var problem Sbd_Pro_t
	flat := Sbd_ProblemSetup(&problem, 1, 1, 1)
	root := problem.pVars[0][1]
	if root < problem.nPars || root >= problem.nPars+problem.nVars {
		t.Fatal("output overlaps parameters")
	}
	clauses := sbdFlatClauses(flat)
	for function := int32(0); function < 4; function++ {
		for input := int32(0); input < 2; input++ {
			for output := int32(0); output < 2; output++ {
				solver := bsat.NewSolver()
				for _, clause := range clauses {
					solver.AddClause(clause...)
				}
				assumptions := []bsat.Literal{bsat.Literal(2*problem.pPars1[0][0] + 1 - (function & 1)), bsat.Literal(2*problem.pPars1[0][1] + 1 - (function >> 1 & 1)), bsat.Literal(2*problem.pDivs[0] + 1 - input), bsat.Literal(2*root + 1 - output)}
				want := output == function>>input&1
				if (solver.Solve(assumptions...) == bsat.Satisfiable) != want {
					t.Fatal("parameterized formal LUT")
				}
			}
		}
	}
	// Loading onto an external output must use the requested output variable,
	// not the source formal output's old index.
	solver := bsat.NewSolver()
	Sbd_ProblemLoad1(&problem, flat, 3, []int32{0}, 1, solver)
	assumptions := []bsat.Literal{bsat.Negative(3 + problem.pPars1[0][0]), bsat.Positive(3 + problem.pPars1[0][1]), bsat.Positive(0), bsat.Negative(1)}
	if solver.Solve(assumptions...) != bsat.Unsatisfiable {
		t.Fatal("formal output was not remapped")
	}
	levels := vec.Vec_WecAlloc(len(clauses))
	for _, clause := range clauses {
		level := vec.Vec_WecPushLevel(levels)
		for _, lit := range clause {
			vec.Vec_IntPush(level, lit)
		}
	}
	solver = bsat.NewSolver()
	Sbd_ProblemLoad2(&problem, levels, 20, []int32{1}, 1, solver)
	if solver.Solve(bsat.Negative(problem.pPars1[0][0]), bsat.Positive(problem.pPars1[0][1])) != bsat.Satisfiable {
		t.Fatal("constant substitution")
	}
}
