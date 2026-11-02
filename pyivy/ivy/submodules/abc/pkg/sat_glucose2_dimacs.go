package abc

import (
	"fmt"
	"os"
)

// Port of src/sat/glucose2/Dimacs.h (namespace Gluco2).

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
	lits.Clear(false)
	for {
		parsedLit := gluco2ParseInt(in)
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
			lits.PushElem(Gluco2MkLit(v, false))
		} else {
			lits.PushElem(Gluco2LitNot(Gluco2MkLit(v, false)))
		}
	}
}

// gluco2ParseDIMACSMain is Gluco2::parse_DIMACS_main (static template).
// C: template<class B, class Solver> static void parse_DIMACS_main(B& in, Solver& S)
func gluco2ParseDIMACSMain[B Gluco2CharStream, S Gluco2DimacsSolver](in B, s S) {
	var lits Gluco2Vec[Gluco2Lit]
	vars := 0
	clauses := 0
	cnt := 0
	for {
		gluco2SkipWhitespace(in)
		if in.Deref() == gluco2EOF {
			break
		} else if in.Deref() == 'p' {
			if gluco2EagerMatch(in, "p cnf") {
				vars = gluco2ParseInt(in)
				clauses = gluco2ParseInt(in)
				// SATRACE'06 hack
				// if (clauses > 4000000)
				//     S.eliminate(true);
			} else {
				fmt.Printf("PARSE ERROR! Unexpected char: %c\n", rune(in.Deref()))
				os.Exit(3)
			}
		} else if in.Deref() == 'c' || in.Deref() == 'p' {
			gluco2SkipLine(in)
		} else {
			cnt++
			gluco2ReadClause(in, s, &lits)
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

// gluco2ParseDIMACS is Gluco2::parse_DIMACS (static template).
// C: template<class Solver> static void parse_DIMACS(gzFile input_stream, Solver& S)
func gluco2ParseDIMACS[S Gluco2DimacsSolver](inputStream *GZIPReader, s S) {
	in := NewGluco2StreamBuffer(inputStream)
	gluco2ParseDIMACSMain(in, s)
}
