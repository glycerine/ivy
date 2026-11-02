package abc

// Hand-checked skeleton of src/sat/glucose2/Dimacs.h (namespace Gluco2).

// Gluco2DimacsSolver is what the Dimacs templates require of their Solver
// argument (instantiated with Solver and SimpSolver): nVars(), newVar() and addClause_().
type Gluco2DimacsSolver interface {
	NVars() int
	NewVar(polarity bool, dvar bool) Gluco2Var
	AddClause_(ps *Gluco2Vec[Gluco2Lit]) bool
}

// gluco2ReadClause is Gluco2::readClause (static template).
// C: template<class B, class Solver> static void readClause(B& in, Solver& S, vec<Lit>& lits)
func gluco2ReadClause[B Gluco2CharStream, S Gluco2DimacsSolver](in B, s S, lits *Gluco2Vec[Gluco2Lit]) {
	panic("TODO, stubbed")
}

// gluco2ParseDIMACSMain is Gluco2::parse_DIMACS_main (static template).
// C: template<class B, class Solver> static void parse_DIMACS_main(B& in, Solver& S)
func gluco2ParseDIMACSMain[B Gluco2CharStream, S Gluco2DimacsSolver](in B, s S) {
	panic("TODO, stubbed")
}

// gluco2ParseDIMACS is Gluco2::parse_DIMACS (static template).
// C: template<class Solver> static void parse_DIMACS(gzFile input_stream, Solver& S)
func gluco2ParseDIMACS[S Gluco2DimacsSolver](inputStream *GzState, s S) {
	panic("TODO, stubbed")
}
