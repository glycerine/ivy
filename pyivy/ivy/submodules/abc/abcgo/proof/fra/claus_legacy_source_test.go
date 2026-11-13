package fra

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dar"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Fra_ClausProcessClausesCut3 and both record routines in fraClaus.c use
// leaf order as the minterm bit order and append cumulative clause ends.
func TestClausAigCutAndRecordWrappers(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	g.AddOutput(root)
	aig.Aig_ManSetRegNum(g, 0)
	cuts := aig.Aig_ComputeCuts(g, 8, 4, 0, 0)
	defer aig.Aig_ManCutStop(cuts)
	var cut *aig.Aig_Cut_t
	for _, c := range cuts.Cuts(root) {
		if len(c.Leaves()) == 2 {
			cut = c
			break
		}
	}
	if cut == nil {
		t.Fatal("missing two-leaf AIG cut")
	}
	ids := cut.Leaves()
	leaves := []aig.Edge{aig.Edge(ids[0] << 1), aig.Edge(ids[1] << 1)}
	s := Fra_SmlStart(g, 0, 2, 8)
	for i, e := range leaves {
		for w := 0; w < 8; w++ {
			s.NodeWords(e.NodeID())[w] = ^uint32(0)
		}
		for w := 8; w < 16; w++ {
			s.NodeWords(e.NodeID())[w] = []uint32{0xaaaaaaaa, 0xcccccccc}[i]
		}
	}
	p := &Clu_Man_t{nSimWordsPref: 8, pCnf: &cnf.Cnf_Dat_t{VarNums: []int32{0, 2, 3, 4}}, vLits: vec.Vec_IntAlloc(4), vClauses: vec.Vec_IntAlloc(2), vCosts: vec.Vec_IntAlloc(2)}
	scores := []int32{-1, -1, -1, -1, 77}
	Fra_ClausProcessClausesCut3(p, s, cut, scores)
	if !reflect.DeepEqual(scores, []int32{64, 64, 64, 64, 77}) {
		t.Fatalf("scores %v", scores)
	}
	Fra_ClausRecordClause(p, &dar.Cut{Leaves: leaves}, 1, 17)
	Fra_ClausRecordClause2(p, cut, 2, 23)
	want := []int32{2*p.pCnf.VarNums[ids[0]] + 1, 2 * p.pCnf.VarNums[ids[1]], 2 * p.pCnf.VarNums[ids[0]], 2*p.pCnf.VarNums[ids[1]] + 1}
	if got := vec.Vec_IntArray(p.vLits); !reflect.DeepEqual(got, want) {
		t.Fatalf("literals %v want %v", got, want)
	}
	if got := vec.Vec_IntArray(p.vClauses); !reflect.DeepEqual(got, []int32{2, 4}) {
		t.Fatalf("ends %v", got)
	}
	if got := vec.Vec_IntArray(p.vCosts); !reflect.DeepEqual(got, []int32{17, 23}) {
		t.Fatalf("costs %v", got)
	}
}

// The original DAR-cut path collects patterns present in combinational but
// absent in sequential simulation. With a zero self-loop latch, every emitted
// clause must hold for both PI values in the only reachable state.
func TestClausLegacyCutsExcludeUnreachableLatchStates(t *testing.T) {
	g := aig.NewGraph()
	input, state := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(input, state))
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	p := Fra_ClausAlloc(g, 1, 0, 100, 4, 10, 10, 1, 0, 0, 0, 0)
	defer Fra_ClausFree(p)
	p.nSimWords, p.nSimFrames = 16, 2
	p.nCuts, p.fFiltering = 91, 1
	p.pCnf = cnf.Cnf_DeriveSimple(g, int32(g.OutputCount()))
	if got := Fra_ClausProcessClauses(p, 0); got != 1 || p.nClauses == 0 {
		t.Fatalf("status %d cuts %d clauses %d", got, p.nCuts, p.nClauses)
	}
	if p.nCuts != 91 || p.fFiltering != 1 {
		t.Fatalf("source preserves counters: cuts=%d filtering=%d", p.nCuts, p.fFiltering)
	}
	start := 0
	for _, end := range vec.Vec_IntArray(p.vClauses) {
		for x := int32(0); x < 2; x++ {
			holds := false
			for _, lit := range vec.Vec_IntArray(p.vLits)[start:int(end)] {
				v := lit >> 1
				var value int32
				if v == p.pCnf.VarNums[input.NodeID()] {
					value = x
				} else if v != p.pCnf.VarNums[state.NodeID()] {
					t.Fatalf("unexpected literal %d", lit)
				}
				if value != lit&1 {
					holds = true
				}
			}
			if !holds {
				t.Fatalf("clause %v rejects reachable x=%d", vec.Vec_IntArray(p.vLits)[start:int(end)], x)
			}
		}
		start = int(end)
	}
}
