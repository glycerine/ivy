package live

import "testing"

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
