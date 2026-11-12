package satcost

import (
	"github.com/glycerine/abcgo/aig/gia/mf"
	"github.com/glycerine/abcgo/sat/satoko"
	"testing"
)

func TestRunCNFCostAndEarlyContradiction(t *testing.T) {
	for _, cnf := range []*mf.CNF{
		{Variables: 1, Clauses: [][]int32{{0}, {1}}},
		{Variables: 0, Clauses: [][]int32{{}}},
	} {
		result, err := RunCNF(cnf, satoko.DefaultOptions())
		if err != nil || result.Status != satoko.SATOKO_UNSAT || result.Conflicts != 0 {
			t.Fatal("early contradiction", result, err)
		}
	}
	// Four clauses exclude every assignment to two variables. No initial units.
	cnf := &mf.CNF{Variables: 2, Clauses: [][]int32{{0, 2}, {0, 3}, {1, 2}, {1, 3}}}
	result, err := RunCNF(cnf, satoko.DefaultOptions())
	if err != nil || result.Status != satoko.SATOKO_UNSAT || result.Conflicts == 0 {
		t.Fatal("real search cost", result, err)
	}
	if _, err := RunCNF(&mf.CNF{Variables: 1, Clauses: [][]int32{{2}}}, satoko.DefaultOptions()); err == nil {
		t.Fatal("out-of-range literal accepted")
	}
}
func TestMissingBackendIsExplicit(t *testing.T) {
	if _, err := (Evaluator{}).Cost(nil, satoko.DefaultOptions()); err == nil {
		t.Fatal("missing backend substituted an encoding")
	}
}
