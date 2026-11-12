// Package lsat implements the solver interface declared in ABC's lsat/solver.h.
// Its execution engine is the pure-Go Satoko CDCL solver.
package lsat

/****************************************************************************************[solver.h]
Copyright (c) 2008, Niklas Sorensson
              2008, Koen Claessen

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/

import (
	"fmt"
	"github.com/glycerine/abcgo/sat/satoko"
)

// Literal uses variable*2+sign; sign one denotes negation.
type Literal = int
type Variable = int
type Truth int

const (
	True Truth = iota
	False
	Undefined
)

type Solver struct {
	engine           *satoko.Solver
	clauseBuffer     []int
	assumptionBuffer []int
	verbosity        int
}

func New() *Solver                                  { return &Solver{engine: satoko.New()} }
func (s *Solver) NewVariable() Variable             { return s.engine.AddVariable(false) }
func (s *Solver) NewLiteral() Literal               { return MakeLiteral(s.NewVariable(), false) }
func MakeLiteral(v Variable, negative bool) Literal { return satoko.Literal(v, negative) }
func Negate(p Literal) Literal                      { return p ^ 1 }
func LiteralVariable(p Literal) Variable            { return p >> 1 }
func LiteralSign(p Literal) bool                    { return p&1 != 0 }
func (s *Solver) AddClause(lits []Literal) bool     { return s.engine.AddClause(lits) }
func (s *Solver) BeginClause()                      { s.clauseBuffer = s.clauseBuffer[:0] }
func (s *Solver) AddClauseLiteral(lit Literal)      { s.clauseBuffer = append(s.clauseBuffer, lit) }
func (s *Solver) CommitClause() bool                { return s.AddClause(s.clauseBuffer) }
func (s *Solver) BeginSolve()                       { s.assumptionBuffer = s.assumptionBuffer[:0] }
func (s *Solver) AddAssumption(lit Literal)         { s.assumptionBuffer = append(s.assumptionBuffer, lit) }
func (s *Solver) CommitSolve() bool                 { return s.Solve(s.assumptionBuffer) }
func (s *Solver) Solve(assumptions []Literal) bool {
	result := s.engine.SolveAssumptions(assumptions)
	if s.verbosity > 0 {
		fmt.Printf("lsat: result %d, decisions %d, conflicts %d\n", result, s.engine.Stats().Decisions, s.engine.Stats().Conflicts)
	}
	return result == satoko.SATOKO_SAT
}
func (s *Solver) Simplify() bool                        { return s.engine.Simplify() }
func (s *Solver) Okay() bool                            { return s.engine.Okay() }
func (s *Solver) SetPolarity(v Variable, negative bool) { s.engine.SetPolarity(v, negative) }
func (s *Solver) SetDecisionVariable(v Variable, enabled bool) {
	s.engine.SetDecisionVariable(v, enabled)
}
func truth(value int) Truth {
	if value < 0 {
		return Undefined
	}
	if value == 0 {
		return False
	}
	return True
}
func (s *Solver) Value(v Variable) Truth      { return truth(s.engine.CurrentValue(v)) }
func (s *Solver) ModelValue(v Variable) Truth { return truth(s.engine.ModelValue(v)) }
func negateValue(value Truth) Truth {
	if value == Undefined {
		return value
	}
	if value == True {
		return False
	}
	return True
}
func (s *Solver) LiteralValue(lit Literal) Truth {
	value := s.Value(lit >> 1)
	if lit&1 != 0 {
		return negateValue(value)
	}
	return value
}
func (s *Solver) ModelLiteralValue(lit Literal) Truth {
	value := s.ModelValue(lit >> 1)
	if lit&1 != 0 {
		return negateValue(value)
	}
	return value
}
func (s *Solver) AssignmentCount() int       { return s.engine.AssignedCount() }
func (s *Solver) ClauseCount() int           { return s.engine.ClauseCount() }
func (s *Solver) LearntCount() int           { return s.engine.LearntCount() }
func (s *Solver) VariableCount() int         { return s.engine.VariableCount() }
func (s *Solver) FreeVariableCount() int     { return s.engine.FreeVariableCount() }
func (s *Solver) Conflict() []Literal        { return s.engine.FinalConflict() }
func (s *Solver) ConflictCount() int64       { return s.engine.Stats().AllConflicts }
func (s *Solver) SetVerbosity(verbosity int) { s.verbosity = verbosity }
