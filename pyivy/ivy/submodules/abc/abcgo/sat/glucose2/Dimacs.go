package glucose2

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

import (
	"fmt"
	"io"
)

type ClauseReceiver interface {
	NVars() int
	NewVar(bool, bool) Var
	AddClause([]Lit) bool
}
type DimacsInfo struct {
	Variables, Clauses, ReadClauses int
	Warnings                        []string
}

func ReadClause(in *StreamBuffer, solver ClauseReceiver) ([]Lit, error) {
	var ps []Lit
	for {
		n, err := ParseInt(in)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return ps, nil
		}
		negative := n < 0
		if negative {
			n = -n
		}
		if n > 1<<30 {
			return nil, fmt.Errorf("DIMACS variable exceeds literal encoding")
		}
		v := Var(n - 1)
		for solver.NVars() <= int(v) {
			solver.NewVar(true, true)
		}
		ps = append(ps, MkLit(v, negative))
	}
}

// ParseDIMACS accepts plain and gzip readers. Header mismatches remain warnings,
// while malformed data returns an error instead of terminating the process.
func ParseDIMACS(input io.Reader, solver ClauseReceiver) (DimacsInfo, error) {
	in, err := NewStreamBuffer(input)
	if err != nil {
		return DimacsInfo{}, err
	}
	defer in.Close()
	var info DimacsInfo
	for {
		SkipWhitespace(in)
		switch in.Peek() {
		case -1:
			if err := in.Err(); err != nil {
				return info, err
			}
			if info.Variables != solver.NVars() {
				info.Warnings = append(info.Warnings, "DIMACS header mismatch: wrong number of variables")
			}
			if info.Clauses != info.ReadClauses {
				info.Warnings = append(info.Warnings, "DIMACS header mismatch: wrong number of clauses")
			}
			return info, nil
		case 'c':
			SkipLine(in)
		case 'p':
			if !EagerMatch(in, "p cnf") {
				return info, fmt.Errorf("line %d: expected p cnf header", in.line)
			}
			info.Variables, err = ParseInt(in)
			if err != nil {
				return info, err
			}
			info.Clauses, err = ParseInt(in)
			if err != nil {
				return info, err
			}
			if info.Variables < 0 || info.Clauses < 0 {
				return info, fmt.Errorf("negative DIMACS header count")
			}
		default:
			ps, err := ReadClause(in, solver)
			if err != nil {
				return info, err
			}
			info.ReadClauses++
			solver.AddClause(ps)
		}
	}
}
