package msat

import (
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
