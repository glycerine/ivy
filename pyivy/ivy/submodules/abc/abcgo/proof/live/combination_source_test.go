package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestHintCombinationTesterRequiresCandidates(t *testing.T) {
	n := liveFixture(false)
	defer func() {
		if recover() == nil {
			t.Error("source asserts when no hint combinations exist")
		}
	}()
	generateDisjunctiveTester(n, n.Graph, 0, 1)
}

func TestGeneralCombinationTesterRequiresPositiveWidth(t *testing.T) {
	n := liveFixture(false)
	defer func() {
		if recover() == nil {
			t.Error("source asserts when no combinations are generated")
		}
	}()
	generateGeneralDisjunctiveTester(n, n.Graph, 0)
}

func TestHintCombinationRejectsSourceNoncontiguousRange(t *testing.T) {
	p := aig.Aig_ManStart(10)
	a := aig.Aig_ObjCreateCi(p)
	for i := 0; i < 3; i++ {
		aig.Aig_ObjCreateCo(p, a)
	}
	n := &Network{Graph: p, OutputNames: []string{"hint_a", "unrelated", "hint_b"}}
	defer func() {
		r := recover()
		if r != "live: source hint range is not contiguous" {
			t.Fatalf("expected source index guard, got %v", r)
		}
	}()
	generateDisjunctiveTester(n, p, 0, 1)
}
