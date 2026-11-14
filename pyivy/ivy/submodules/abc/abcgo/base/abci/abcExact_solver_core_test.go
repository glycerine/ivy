package abci

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestExactSolverVariableIndexLayouts(t *testing.T) {
	p := &Ses_Man_t{nSpecVars: 3, nSpecFunc: 1, nGates: 2, nRows: 7, nArrTimeMax: 2}
	p.nSimOffset, p.nOutputOffset, p.nGateOffset, p.nSelectOffset, p.nDepthOffset = 50, 40, 30, 0, 100
	if got := []int32{Ses_ManSimVar(p, 1, 4), Ses_ManOutputVar(p, 0, 1), Ses_ManGateVar(p, 1, 1, 1)}; got[0] != 61 || got[1] != 41 || got[2] != 35 {
		t.Fatalf("simulation/output/gate variables = %v", got)
	}
	for pair, want := range [][2]int32{{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3}} {
		if got := Ses_ManSelectVar(p, 1, want[0], want[1]); got != int32(pair+3) {
			t.Fatalf("select variable for %v = %d, want %d", want, got, pair+3)
		}
	}
	if got := Ses_ManDepthVar(p, 1, 2); got != 105 {
		t.Fatalf("depth variable = %d, want 105", got)
	}
}

func TestExactSolverDerivesTruthFromNormalizedGateEncoding(t *testing.T) {
	p := &Ses_Man_t{nSpecVars: 2, nSpecFunc: 1, nSpecWords: 1}
	// One XOR gate: operation bits 0 and 1 select the two unequal input rows.
	solution := string([]byte{2, 1, 1, 3, 2, 0, 1, 0, 0, 1, 1})
	got := Ses_ManDeriveTruth(p, solution, 0)
	if got == nil || *got != 0x6666666666666666 {
		if got == nil {
			t.Fatal("derived XOR truth is nil")
		}
		t.Fatalf("derived XOR truth = %#016x", *got)
	}
	complemented := []byte(solution)
	complemented[7] = 1
	if inv := Ses_ManDeriveTruth(p, string(complemented), 1); inv == nil || *inv != ^uint64(0x6666666666666666) {
		if inv == nil {
			t.Fatal("complemented output truth is nil")
		}
		t.Fatalf("complemented output truth = %#016x", *inv)
	}
}

func TestExactSolverAllocationRestoresTruthAndArrivalTimes(t *testing.T) {
	truth := [4]word{1, 0x1234, 0x5678, 0x9abc}
	original := truth
	arrivals := [2]int32{3, 5}
	p := Ses_ManAlloc(&truth[0], 2, 1, 4, &arrivals[0], 0, 100, 0)
	if p.bSpecInv != 1 || truth[0]&1 != 0 || arrivals != [2]int32{0, 2} || p.nArrTimeMax != 3 {
		t.Fatalf("normalized manager state: inv=%d truth=%x arrivals=%v max=%d", p.bSpecInv, truth, arrivals, p.nArrTimeMax)
	}
	Ses_ManCleanLight(p)
	if truth != original || arrivals != [2]int32{3, 5} {
		t.Fatalf("cleanup did not restore caller data: truth=%x arrivals=%v", truth, arrivals)
	}
}

func TestExactSolverAddsRequiredSingleOutputTruthAssumption(t *testing.T) {
	truth := [4]word{0x6666666666666666}
	p := &Ses_Man_t{
		pSpec: &truth[0], nSpecVars: 2, nSpecFunc: 1, nSpecWords: 1,
		nRows: 3, nMaxDepth: -1, vAssump: Vec_IntAlloc(4), pSat: bsat.NewSolver(),
	}
	defer p.pSat.Close()
	Ses_ManCreateVars(p, 1)
	if Ses_ManCreateTruthTableClause(p, 0) != 1 {
		t.Fatal("truth clause creation failed")
	}
	want := Ses_ManSimVar(p, 0, 0) * 2
	if vec.Vec_IntSize(p.vAssump) != 1 || vec.Vec_IntEntry(p.vAssump, 0) != want {
		t.Fatalf("single-output assumption = %v, want literal %d", vec.Vec_IntArray(p.vAssump), want)
	}
}

func TestExactSolverRandomRowsMatchSrandCAFE(t *testing.T) {
	rng := newGlibcRand(0xCAFE)
	want := []int32{2, 1, 2, 4, 3, 5, 0, 2}
	for i, expected := range want {
		if got := int32(rng.next() % 7); got != expected {
			t.Fatalf("rand()%%7 draw %d = %d, want %d", i, got, expected)
		}
	}
}

func TestExactSolverFindsMinimumXORNetwork(t *testing.T) {
	truth := [4]word{0x6666666666666666}
	p := Ses_ManAlloc(&truth[0], 2, 1, -1, nil, 0, 100000, 0)
	p.nStartGates = 0
	defer Ses_ManClean(p)
	solution := Ses_ManFindMinimumSize(p)
	if solution == "" || int32(solution[ABC_EXACT_SOL_NGATES]) != 1 {
		t.Fatalf("minimum XOR solution = %v", []byte(solution))
	}
	derived := Ses_ManDeriveTruth(p, solution, 0)
	if derived == nil || *derived != truth[0] {
		if derived == nil {
			t.Fatal("minimum solution truth is nil")
		}
		t.Fatalf("minimum solution truth = %#016x, want %#016x", *derived, truth[0])
	}
}
