package abc

// Hand-checked skeleton of src/sat/glucose/Dimacs.h (namespace Gluco).

// GlucoDimacsSolver is what the Dimacs templates require of their Solver
// argument (instantiated with Solver and SimpSolver): nVars(), newVar() and addClause_().
type GlucoDimacsSolver interface {
	NVars() int
	NewVar(polarity bool, dvar bool) GlucoVar
	AddClause_(ps *GlucoVec[GlucoLit]) bool
}

// glucoReadClause is Gluco::readClause (static template).
// C: template<class B, class Solver> static void readClause(B& in, Solver& S, vec<Lit>& lits)
func glucoReadClause[B GlucoCharStream, S GlucoDimacsSolver](in B, s S, lits *GlucoVec[GlucoLit]) {
	panic("TODO, stubbed")
}

// glucoParseDIMACSMain is Gluco::parse_DIMACS_main (static template).
// C: template<class B, class Solver> static void parse_DIMACS_main(B& in, Solver& S)
func glucoParseDIMACSMain[B GlucoCharStream, S GlucoDimacsSolver](in B, s S) {
	panic("TODO, stubbed")
}

// glucoParseDIMACS is Gluco::parse_DIMACS (static template).
// C: template<class Solver> static void parse_DIMACS(gzFile input_stream, Solver& S)
func glucoParseDIMACS[S GlucoDimacsSolver](inputStream *GzState, s S) {
	panic("TODO, stubbed")
}
