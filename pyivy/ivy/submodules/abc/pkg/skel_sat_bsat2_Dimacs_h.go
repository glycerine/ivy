package abc

// Hand-checked skeleton of src/sat/bsat2/Dimacs.h (namespace Minisat).

// MinisatDimacsSolver is what the Dimacs templates require of their Solver
// argument (instantiated with Solver and SimpSolver): nVars(), newVar() and addClause_().
type MinisatDimacsSolver interface {
	NVars() int
	NewVar(polarity bool, dvar bool) MinisatVar
	AddClause_(ps *MinisatVec[MinisatLit]) bool
}

// minisatReadClause is Minisat::readClause (static template).
// C: template<class B, class Solver> static void readClause(B& in, Solver& S, vec<Lit>& lits)
func minisatReadClause[B MinisatCharStream, S MinisatDimacsSolver](in B, s S, lits *MinisatVec[MinisatLit]) {
	panic("TODO, stubbed")
}

// minisatParseDIMACSMain is Minisat::parse_DIMACS_main (static template).
// C: template<class B, class Solver> static void parse_DIMACS_main(B& in, Solver& S)
func minisatParseDIMACSMain[B MinisatCharStream, S MinisatDimacsSolver](in B, s S) {
	panic("TODO, stubbed")
}

// minisatParseDIMACS is Minisat::parse_DIMACS (static template).
// C: template<class Solver> static void parse_DIMACS(gzFile input_stream, Solver& S)
func minisatParseDIMACS[S MinisatDimacsSolver](inputStream *GzState, s S) {
	panic("TODO, stubbed")
}
