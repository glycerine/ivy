package fracomb

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

type sequenceSolver struct {
	events    []string
	variables int32
}

func (s *sequenceSolver) AddClause(lits ...bsat.Literal) bool {
	s.events = append(s.events, fmt.Sprintf("add:%v", lits))
	return true
}
func (s *sequenceSolver) NewVariable() int32 { n := s.variables; s.variables++; return n }
func (s *sequenceSolver) NumVars() int32     { return s.variables }
func (s *sequenceSolver) Simplify() bool     { s.events = append(s.events, "simplify"); return true }
func (s *sequenceSolver) SetLearntReduction(a, b, c int32) {
	s.events = append(s.events, fmt.Sprintf("learned:%d,%d,%d", a, b, c))
}
func (s *sequenceSolver) SolveWithSourceLimits(_ []bsat.Literal, a, b, c, d int64) bsat.Status {
	s.events = append(s.events, fmt.Sprintf("solve:%d,%d,%d,%d", a, b, c, d))
	return bsat.Unknown
}
func (s *sequenceSolver) Value(int32) (bool, bool) { return false, false }

func TestSATCompleteSolverLifecycle(t *testing.T) {
	for _, mode := range []int32{0, 1} {
		t.Run(fmt.Sprint(mode), func(t *testing.T) {
			g := aig.NewGraph()
			x := g.NewInput()
			g.AddOutput(x)
			g.AddOutput(x.Not())
			s := &sequenceSolver{}
			status := fraFraigSat(g, 3, 5, 7, 11, 13, 0, mode, 0, 0, func(int32) fraCecSolver { return s })
			if status != -1 {
				t.Fatalf("status=%d", status)
			}
			expectedPrefix := []string{"add:[8]", "add:[2 7]", "add:[3 6]", "add:[4 6]", "add:[5 7]", "simplify", "learned:7,11,13"}
			expected := append([]string(nil), expectedPrefix...)
			if mode == 0 {
				expected = append(expected, "add:[2 4]")
			} else {
				expected = append(expected, "add:[2]", "add:[4]")
			}
			expected = append(expected, "simplify", "solve:3,5,0,0")
			if !reflect.DeepEqual(s.events, expected) {
				t.Fatalf("events=%v, want %v", s.events, expected)
			}
		})
	}
}
