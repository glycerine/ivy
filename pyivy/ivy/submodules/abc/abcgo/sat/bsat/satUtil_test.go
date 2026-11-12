package bsat

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestDimacsIncludesBinaryWatchesAndExactCount(t *testing.T) {
	s := NewSolver()
	s.AddClause(0, 2)
	s.AddClause(1, 4, 6)
	var out bytes.Buffer
	if err := WriteSolverDIMACS(&out, s, []lit{3}, true); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var p, cnf string
	var variables, clauses int
	fmt.Sscan(lines[0], &p, &cnf, &variables, &clauses)
	if p != "p" || cnf != "cnf" || variables != 4 || clauses != 3 || len(lines) != clauses+1 {
		t.Fatal(out.String())
	}
	if !strings.Contains(out.String(), "1 2 0\n") || !strings.Contains(out.String(), "-2 0\n") {
		t.Fatal(out.String())
	}
}
