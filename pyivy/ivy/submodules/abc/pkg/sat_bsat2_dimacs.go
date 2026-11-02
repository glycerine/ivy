package abc

import (
	"fmt"
	"os"
)

// Port of src/sat/bsat2/Dimacs.h (namespace Minisat).

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
	lits.Clear(false)
	for {
		parsedLit := minisatParseInt(in)
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
			lits.PushElem(MinisatMkLit(v, false))
		} else {
			lits.PushElem(MinisatLitNot(MinisatMkLit(v, false)))
		}
	}
}

// minisatParseDIMACSMain is Minisat::parse_DIMACS_main (static template).
// C: template<class B, class Solver> static void parse_DIMACS_main(B& in, Solver& S)
func minisatParseDIMACSMain[B MinisatCharStream, S MinisatDimacsSolver](in B, s S) {
	var lits MinisatVec[MinisatLit]
	vars := 0
	clauses := 0
	cnt := 0
	for {
		minisatSkipWhitespace(in)
		if in.Deref() == minisatEOF {
			break
		} else if in.Deref() == 'p' {
			if minisatEagerMatch(in, "p cnf") {
				vars = minisatParseInt(in)
				clauses = minisatParseInt(in)
				// SATRACE'06 hack
				// if (clauses > 4000000)
				//     S.eliminate(true);
			} else {
				fmt.Printf("PARSE ERROR! Unexpected char: %c\n", rune(in.Deref()))
				os.Exit(3)
			}
		} else if in.Deref() == 'c' || in.Deref() == 'p' {
			minisatSkipLine(in)
		} else {
			cnt++
			minisatReadClause(in, s, &lits)
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

// minisatParseDIMACS is Minisat::parse_DIMACS (static template).
// C: template<class Solver> static void parse_DIMACS(gzFile input_stream, Solver& S)
func minisatParseDIMACS[S MinisatDimacsSolver](inputStream *GZIPReader, s S) {
	in := NewMinisatStreamBuffer(inputStream)
	minisatParseDIMACSMain(in, s)
}
