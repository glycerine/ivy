package int

import (
	"errors"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/cnf"
	"reflect"
	"testing"
)

type fixtureM114p struct {
	roots  [][]int32
	chains []M114pChain
	added  [][]int32
	nvars  int32
}

func (p *fixtureM114p) SetVarNum(n int32) { p.nvars = n }
func (p *fixtureM114p) AddClause(c []int32) bool {
	p.added = append(p.added, append([]int32(nil), c...))
	return true
}
func (*fixtureM114p) Solve() int32           { return -1 }
func (*fixtureM114p) ConflictNum() int32     { return 0 }
func (p *fixtureM114p) Roots() [][]int32     { return p.roots }
func (p *fixtureM114p) Chains() []M114pChain { return p.chains }
func (*fixtureM114p) Close()                 {}

func TestMiniSatProofInterpolation(t *testing.T) {
	proof := &fixtureM114p{roots: [][]int32{{0}, {1}}, chains: []M114pChain{{Clauses: []int32{0, 1}, Pivots: []int32{0}}}}
	for _, pudlak := range []bool{false, true} {
		graph, err := InterpolateM114p(proof, []int32{0, 1}, []int32{0}, pudlak)
		if err != nil {
			t.Fatal(err)
		}
		for _, x := range []bool{false, true} {
			got, err := graph.Evaluate([]bool{x})
			if err != nil || len(got) != 1 || got[0] != x {
				t.Fatalf("pudlak=%v x=%v got %v err %v", pudlak, x, got, err)
			}
		}
	}
}

func TestMiniSatBackendUnavailableAndInvalidResolution(t *testing.T) {
	if _, err := InterpolateM114p(nil, nil, nil, false); !errors.Is(err, ErrMiniSat114pUnavailable) {
		t.Fatal(err)
	}
	if _, err := Inter_ManResolveM114p([]int32{0}, []int32{0}, 0); err == nil {
		t.Fatal("same polarity resolution accepted")
	}
}

func TestMiniSatPartitionAssemblyRestoresCNF(t *testing.T) {
	inter := aig.NewGraph()
	x := inter.NewInput()
	inter.AddOutput(x.Not())
	trans := aig.NewGraph()
	state := trans.NewInput()
	trans.AddOutput(state)
	aig.Aig_ManSetRegNum(trans, 1)
	frames := aig.NewGraph()
	next := frames.NewInput()
	frames.AddOutput(next)
	ci, ct, cf := cnf.Cnf_Derive(inter, 0), cnf.Cnf_Derive(trans, 1), cnf.Cnf_Derive(frames, 0)
	beforeI, beforeT := append([]int32(nil), ci.VarNums...), append([]int32(nil), ct.VarNums...)
	s := &fixtureM114p{}
	roots, vars, err := Inter_ManDeriveSatSolverM114p(s, inter, ci, trans, ct, frames, cf)
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != len(s.added) || len(vars) != int(s.nvars) || s.nvars != int32(ci.Variables+ct.Variables+cf.Variables) {
		t.Fatalf("partition dimensions roots=%v vars=%v solver=%d", roots, vars, s.nvars)
	}
	if vars[cf.VarNums[next.NodeID()]] != 0 {
		t.Fatal("shared variable mapping missing")
	}
	if !reflect.DeepEqual(beforeI, ci.VarNums) || !reflect.DeepEqual(beforeT, ct.VarNums) {
		t.Fatal("lifted CNF not restored")
	}
	sawB := false
	for _, part := range roots {
		if part == 1 {
			sawB = true
		}
		if part == 0 && sawB {
			t.Fatal("A clause appears after B")
		}
	}
	if !sawB {
		t.Fatal("missing B clauses")
	}
}

func TestMiniSatMcMillanCreatesOnlyReferencedInputs(t *testing.T) {
	// The shared variable is irrelevant to these partition-local clauses.
	proof := &fixtureM114p{roots: [][]int32{{0}, {1}}, chains: []M114pChain{{Clauses: []int32{0, 1}, Pivots: []int32{0}}}}
	graph, err := InterpolateM114p(proof, []int32{0, 1}, []int32{-1, 0}, false)
	if err != nil {
		t.Fatal(err)
	}
	if graph.InputCount() != 0 {
		t.Fatalf("unused shared variable created %d inputs", graph.InputCount())
	}
}
