package abc

import (
	"fmt"
	"os"
)

// Port of src/sat/glucose/Dimacs.h (namespace Gluco).

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
	lits.Clear(false)
	for {
		parsedLit := glucoParseInt(in)
		if parsedLit == 0 {
			break
		}
		v := parsedLit
		if v < 0 {
			v = -v
		}
		v--
		for v >= s.NVars() {
			s.NewVar(true, true)
		}
		if parsedLit > 0 {
			lits.PushElem(GlucoMkLit(v, false))
		} else {
			lits.PushElem(GlucoLitNot(GlucoMkLit(v, false)))
		}
	}
}

// glucoParseDIMACSMain is Gluco::parse_DIMACS_main (static template).
// C: template<class B, class Solver> static void parse_DIMACS_main(B& in, Solver& S)
func glucoParseDIMACSMain[B GlucoCharStream, S GlucoDimacsSolver](in B, s S) {
	var lits GlucoVec[GlucoLit]
	vars := 0
	clauses := 0
	cnt := 0
	for {
		glucoSkipWhitespace(in)
		if in.Deref() == glucoEOF {
			break
		} else if in.Deref() == 'p' {
			if glucoEagerMatch(in, "p cnf") {
				vars = glucoParseInt(in)
				clauses = glucoParseInt(in)
				// SATRACE'06 hack
				// if (clauses > 4000000)
				//     S.eliminate(true);
			} else {
				fmt.Printf("PARSE ERROR! Unexpected char: %c\n", rune(in.Deref()))
				os.Exit(3)
			}
		} else if in.Deref() == 'c' || in.Deref() == 'p' {
			glucoSkipLine(in)
		} else {
			cnt++
			glucoReadClause(in, s, &lits)
			s.AddClause_(&lits)
		}
	}
	if vars != s.NVars() {
		fmt.Fprintf(os.Stderr, "WARNING! DIMACS header mismatch: wrong number of variables.\n")
	}
	if cnt != clauses {
		fmt.Fprintf(os.Stderr, "WARNING! DIMACS header mismatch: wrong number of clauses.\n")
	}
}

// glucoParseDIMACS is Gluco::parse_DIMACS (static template).
// C: template<class Solver> static void parse_DIMACS(gzFile input_stream, Solver& S)
func glucoParseDIMACS[S GlucoDimacsSolver](inputStream *GZIPReader, s S) {
	in := NewGlucoStreamBuffer(inputStream)
	glucoParseDIMACSMain(in, s)
}
