package msat

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceDIMACSWriterIncludesLearnedAndLevelZeroAssignments(t *testing.T) {
	p := New(3)
	if !p.AddClause([]int{0, 2}) {
		t.Fatal("clause")
	}
	p.vLearned.data = append(p.vLearned.data, &Msat_Clause_t{learned: true, lits: []int{1, 4}})
	p.pLevel[2], p.pAssigns[2] = 0, 4
	path := filepath.Join(t.TempDir(), "source.cnf")
	Msat_SolverWriteDimacs(p, path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"c Produced by Msat_SolverWriteDimacs() on ", "p cnf 3 3\n", "1 2 0\n", "-1 3 0\n", "3 0\n"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in %q", want, text)
		}
	}
}

func captureMSATOutput(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old; r.Close() }()
	fn()
	w.Close()
	os.Stdout = old
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
func TestSourceAssignmentAndClauseReports(t *testing.T) {
	p := New(3)
	p.pAssigns[0], p.pAssigns[1] = 0, 3
	got := captureMSATOutput(t, func() { Msat_SolverPrintAssignment(p) })
	if got != "Current assignments are: \n012\n10.\n" {
		t.Fatalf("assignment report %q", got)
	}
	c := &Msat_Clause_t{learned: true, activity: 1.5, lits: []int{0, 3}}
	if got = captureMSATOutput(t, func() { Msat_ClausePrint(c) }); got != "Act = 1.5000   1 -2\n" {
		t.Fatalf("learned clause report %q", got)
	}
}
