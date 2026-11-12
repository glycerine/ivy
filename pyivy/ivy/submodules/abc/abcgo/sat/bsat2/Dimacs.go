package bsat2

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"strconv"
)

/****************************************************************************************[Dimacs.h]
Copyright (c) 2003-2006, Niklas Een, Niklas Sorensson
Copyright (c) 2007-2010, Niklas Sorensson

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

type ClauseSink interface {
	NewVariable() Var
	NumVars() int
	AddClause(...Lit) bool
}

// ParseDIMACS accepts plain or gzip-compressed DIMACS. Clause boundaries are zero
// tokens, independent of lines; malformed/truncated clauses are errors.
func ParseDIMACS(reader io.Reader, solver ClauseSink) error {
	buffered := bufio.NewReader(reader)
	if prefix, _ := buffered.Peek(2); len(prefix) == 2 && prefix[0] == 0x1f && prefix[1] == 0x8b {
		zipped, err := gzip.NewReader(buffered)
		if err != nil {
			return err
		}
		defer zipped.Close()
		reader = zipped
	} else {
		reader = buffered
	}
	tokens := NewTokenReader(reader)
	literals := []Lit{}
	declaredVars, declaredClauses := -1, -1
	clauses := 0
	for {
		token, err := tokens.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if token == "p" {
			if declaredVars >= 0 || clauses != 0 || len(literals) != 0 {
				return fmt.Errorf("line %d: misplaced DIMACS header", tokens.Line())
			}
			kind, err := tokens.Next()
			if err != nil || kind != "cnf" {
				return fmt.Errorf("line %d: expected p cnf header", tokens.Line())
			}
			counts := [2]int{}
			for i := range counts {
				text, err := tokens.Next()
				if err != nil {
					return err
				}
				n, err := strconv.ParseInt(text, 10, 32)
				if err != nil || n < 0 {
					return fmt.Errorf("line %d: invalid DIMACS count", tokens.Line())
				}
				counts[i] = int(n)
			}
			declaredVars, declaredClauses = counts[0], counts[1]
			for solver.NumVars() < declaredVars {
				solver.NewVariable()
			}
			continue
		}
		n, err := strconv.ParseInt(token, 10, 64)
		if err != nil || n < -(1<<30) || n > 1<<30 {
			return fmt.Errorf("line %d: invalid DIMACS literal %q", tokens.Line(), token)
		}
		if n == 0 {
			solver.AddClause(literals...)
			clauses++
			literals = literals[:0]
			continue
		}
		negative := n < 0
		if negative {
			n = -n
		}
		v := Var(n - 1)
		if declaredVars >= 0 && int(v) >= declaredVars {
			return fmt.Errorf("line %d: literal exceeds declared variable count", tokens.Line())
		}
		for solver.NumVars() <= int(v) {
			solver.NewVariable()
		}
		literals = append(literals, MakeLit(v, negative))
	}
	if len(literals) != 0 {
		return fmt.Errorf("unterminated DIMACS clause")
	}
	if declaredClauses >= 0 && clauses != declaredClauses {
		return fmt.Errorf("DIMACS clause count: header %d, read %d", declaredClauses, clauses)
	}
	return nil
}
func (s *Solver) WriteDIMACS(writer io.Writer, assumptions ...Lit) error {
	if !s.ok {
		_, err := fmt.Fprint(writer, "p cnf 1 2\n1 0\n-1 0\n")
		return err
	}
	clauses := [][]Lit{}
	for _, h := range s.clauses {
		c := s.ca.At(h)
		if s.satisfied(c) {
			continue
		}
		literals := []Lit{}
		for _, l := range c.Lits {
			if s.Value(l) != False {
				literals = append(literals, l)
			}
		}
		clauses = append(clauses, literals)
	}
	for _, l := range s.trail {
		if s.vardata[l.Var()].level == 0 {
			clauses = append(clauses, []Lit{l})
		}
	}
	for _, l := range assumptions {
		clauses = append(clauses, []Lit{l})
	}
	if _, err := fmt.Fprintf(writer, "p cnf %d %d\n", s.NumVars(), len(clauses)); err != nil {
		return err
	}
	for _, c := range clauses {
		for _, l := range c {
			n := int64(l.Var()) + 1
			if l.Sign() {
				n = -n
			}
			if _, err := fmt.Fprintf(writer, "%d ", n); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(writer, "0"); err != nil {
			return err
		}
	}
	return nil
}
