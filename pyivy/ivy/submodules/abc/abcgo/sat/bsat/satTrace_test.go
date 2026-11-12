package bsat

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestTraceExplicitWriteRoundTrip(t *testing.T) {
	s := NewSolver()
	s.NewVariable()
	s.NewVariable()
	name := filepath.Join(t.TempDir(), "trace.cnf")
	if err := Sat_SolverTraceStart(s, name); err != nil {
		t.Fatal(err)
	}
	if err := Sat_SolverTraceStart(s, name); err == nil {
		t.Fatal("accepted overlapping trace")
	}
	for _, root := range []int32{1, 1, 0} {
		if err := Sat_SolverTraceWrite(s, []lit{0, 3}, root); err != nil {
			t.Fatal(err)
		}
	}
	if err := Sat_SolverTraceStop(s); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var tag string
	var variables, clauses, roots int32
	if _, err := fmt.Fscan(file, &tag, &variables, &clauses, &roots); err != nil || tag != "p" || variables != 2 || clauses != 3 || roots != 2 {
		t.Fatal(tag, variables, clauses, roots, err)
	}
	for range 3 {
		var positive, negative, end int
		if _, err := fmt.Fscan(file, &positive, &negative, &end); err != nil || positive != 1 || negative != -2 || end != 0 {
			t.Fatal(positive, negative, end, err)
		}
	}

	if err := Sat_SolverTraceStop(s); err != nil {
		t.Fatal(err)
	}
}
