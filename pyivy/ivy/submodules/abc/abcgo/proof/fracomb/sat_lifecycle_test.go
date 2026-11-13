package fracomb

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/sat/bsat"
)

// ABC's Cnf_DataWriteIntoSolverInt simplifies the loaded circuit CNF before
// Fra_FraigSat adds its output clause and performs a second simplification.
type lifecycleSolver struct {
	events            []string
	addOK, simplifyOK bool
}

func (s *lifecycleSolver) AddClause(...bsat.Literal) bool {
	s.events = append(s.events, "add")
	return s.addOK
}
func (s *lifecycleSolver) NewVariable() int32 { return 0 }
func (s *lifecycleSolver) NumVars() int32     { return 0 }
func (s *lifecycleSolver) Simplify() bool {
	s.events = append(s.events, "simplify")
	return s.simplifyOK
}
func (s *lifecycleSolver) SolveWithSourceLimits([]bsat.Literal, int64, int64, int64, int64) bsat.Status {
	return bsat.Unknown
}
func (s *lifecycleSolver) Value(int32) (bool, bool) { return false, false }

func TestSATCircuitCNFSimplifiedBeforeOutputConstraints(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		addOK, simplifyOK, want bool
		events                  []string
	}{
		{"consistent", true, true, true, []string{"add", "add", "simplify"}},
		{"conflicting clause", false, true, false, []string{"add"}},
		{"root conflict", true, false, false, []string{"add", "add", "simplify"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &lifecycleSolver{addOK: tc.addOK, simplifyOK: tc.simplifyOK}
			got := fraLoadCNF(s, [][]bsat.Literal{{0}, {1, 2}}, 0)
			if got != tc.want || !reflect.DeepEqual(s.events, tc.events) {
				t.Fatalf("result=%t events=%v", got, s.events)
			}
		})
	}
}
