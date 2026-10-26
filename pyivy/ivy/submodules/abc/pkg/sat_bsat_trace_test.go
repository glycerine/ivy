package abc

// Stand-alone tests for the trace format from src/sat/bsat/satTrace.c.

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestSatSolverTraceExactFile(t *testing.T) {
	name := filepath.Join(t.TempDir(), "solver.trace")
	trace, err := StartSatSolverTrace(name, SatTraceLimits{})
	if err != nil {
		t.Fatal(err)
	}
	clauses := []struct {
		literals []SatLit
		root     bool
	}{{[]SatLit{0, 3, 4}, true}, {[]SatLit{7}, false}, {nil, true}}
	for _, clause := range clauses {
		if err = trace.Write(clause.literals, clause.root); err != nil {
			t.Fatal(err)
		}
	}
	if stats := trace.Stats(); stats != (SatTraceStats{3, 2, 4, true}) {
		t.Fatal(stats)
	}
	solver, _ := NewSatSolverWithLimit(10)
	if err = solver.SetVariableCount(5); err != nil {
		t.Fatal(err)
	}
	if err = trace.StopForSolver(solver); err != nil {
		t.Fatal(err)
	}
	if stats := trace.Stats(); stats != (SatTraceStats{3, 2, 4, false}) {
		t.Fatal(stats)
	}
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	header := "p 5 3 2"
	want := header + strings.Repeat(" ", satTraceHeaderWidth-len(header)) + "\n" +
		" 1 -2 3 0\n" +
		" -4 0\n" +
		" 0\n"
	if string(data) != want {
		t.Fatalf("got %q want %q", data, want)
	}
	if err = trace.Stop(5); err != nil {
		t.Fatal(err)
	}
	if err = trace.Write(nil, false); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestSatSolverTraceErrorsAndLimits(t *testing.T) {
	if trace, err := StartSatSolverTrace("", SatTraceLimits{}); trace != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(trace, err)
	}
	if trace, err := StartSatSolverTrace(filepath.Join(t.TempDir(), "x"), SatTraceLimits{MaxClauses: -1}); trace != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(trace, err)
	}
	name := filepath.Join(t.TempDir(), "limited.trace")
	trace, err := StartSatSolverTrace(name, SatTraceLimits{MaxClauses: 1, MaxLiterals: 2})
	if err != nil {
		t.Fatal(err)
	}
	if err = trace.Write([]SatLit{0, 2}, true); err != nil {
		t.Fatal(err)
	}
	if err = trace.Write(nil, false); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if err = trace.Stop(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if !trace.Stats().Open || trace.Stop(2) != nil {
		t.Fatal("invalid stop was not atomic", trace.Stats())
	}
	invalidName := filepath.Join(t.TempDir(), "invalid.trace")
	invalid, _ := StartSatSolverTrace(invalidName, SatTraceLimits{})
	if err = invalid.Write([]SatLit{SatLit(^uint32(0))}, false); !errors.Is(err, ErrInvalidLiteral) || invalid.Stats().Clauses != 0 {
		t.Fatal(err, invalid.Stats())
	}
	if err = invalid.StopForSolver(nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if !invalid.Stats().Open || invalid.Stop(0) != nil {
		t.Fatal("nil solver stop was not atomic", invalid.Stats())
	}
	wideName := filepath.Join(t.TempDir(), "wide.trace")
	wide, _ := StartSatSolverTrace(wideName, SatTraceLimits{})
	wide.clauses, wide.roots = int(^uint(0)>>1), int(^uint(0)>>1)
	if err = wide.Stop(int(^uint(0) >> 1)); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	wide.clauses, wide.roots = 0, 0
	if !wide.Stats().Open || wide.Stop(0) != nil {
		t.Fatal("oversized header stop was not atomic", wide.Stats())
	}
}

func FuzzFormatSatTraceClause(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 254, 255})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			data = data[:256]
		}
		literals := make([]SatLit, len(data))
		var want strings.Builder
		for i, value := range data {
			literals[i] = SatLit(value)
			variable := int64(value>>1) + 1
			if value&1 != 0 {
				variable = -variable
			}
			want.WriteByte(' ')
			want.WriteString(strconv.FormatInt(variable, 10))
		}
		want.WriteString(" 0\n")
		got, err := FormatSatTraceClause(literals)
		if err != nil || got != want.String() {
			t.Fatal(got, want.String(), err)
		}
	})
}
