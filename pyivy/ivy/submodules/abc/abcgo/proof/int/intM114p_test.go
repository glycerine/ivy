package int

import (
	"errors"
	"testing"
)

type fixtureM114p struct {
	roots  [][]int32
	chains []M114pChain
}

func (*fixtureM114p) SetVarNum(int32)        {}
func (*fixtureM114p) AddClause([]int32) bool { return true }
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
