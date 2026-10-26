package abc

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCNFReadNativeLines(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_read.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Input     string
		Variables int
		Clauses   [][]SatLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 10 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		got, e := ParseCNFDIMACS([]byte(row.Input), CNFDIMACSReadOptions{NativeLines: true})
		if e != nil || got.variables != row.Variables || !equalSatClauses(got.clauses, row.Clauses) {
			t.Fatal(row, got, e)
		}
	}
}
func TestCNFReadStrictAndMalformed(t *testing.T) {
	got, e := ParseCNFDIMACS([]byte(" c comment\np cnf 3 3\n1 -2\n3 0 0\n-1 0\n"), CNFDIMACSReadOptions{})
	if e != nil || !equalSatClauses(got.clauses, [][]SatLit{{0, 3, 4}, {}, {1}}) {
		t.Fatal(got, e)
	}
	empty, e := ParseCNFDIMACS([]byte("p cnf 0 0\n"), CNFDIMACSReadOptions{})
	if e != nil || empty.variables != 0 || len(empty.clauses) != 0 {
		t.Fatal(empty, e)
	}
	malformed := []string{"", "c only comment", "1 0", "p cnf", "p dnf 1 1\n0", "p cnf -1 1\n0", "p cnf 1 -1\n0", "p cnf 1 1 extra\n0", "p cnf 1 1\np cnf 1 1\n0", "p cnf 1 1\n2 0", "p cnf 1 1\n1x 0", "p cnf 1 1\n1", "p cnf 1 1\n0 0", "p cnf 1 2\n0", "p cnf 1 1\na 1 0", "p cnf 1 1\n999999999999999999999999999999 0", "p cnf 1 1\n\x00 0"}
	for _, text := range malformed {
		if f, e := ParseCNFDIMACS([]byte(text), CNFDIMACSReadOptions{}); f != nil || !errors.Is(e, ErrFormat) {
			t.Fatal(text, f, e)
		}
	}
	for _, text := range []string{"p cnf 1 1\n1", "p cnf 0 0", "p cnf 1 1\nword 0"} {
		if f, e := ParseCNFDIMACS([]byte(text), CNFDIMACSReadOptions{NativeLines: true}); f != nil || !errors.Is(e, ErrFormat) {
			t.Fatal(text, f, e)
		}
	}
}
func TestCNFReadLimitsAndFile(t *testing.T) {
	source := []byte("p cnf 2 2\n1 2 0\n-1 0\n")
	for _, o := range []CNFDIMACSReadOptions{{MaxBytes: 1}, {MaxLineBytes: 3}, {Limits: CNFTransformOptions{MaxVariables: 1}}, {Limits: CNFTransformOptions{MaxClauses: 1}}, {Limits: CNFTransformOptions{MaxLiterals: 1}}, {Limits: CNFTransformOptions{Stop: func() bool { return true }}}} {
		f, e := ParseCNFDIMACS(source, o)
		if f != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(f, e)
		}
	}
	path := filepath.Join(t.TempDir(), "formula.cnf")
	if e := os.WriteFile(path, source, 0600); e != nil {
		t.Fatal(e)
	}
	f, e := ReadCNFDIMACSFile(path, CNFDIMACSReadOptions{MaxBytes: len(source)})
	if e != nil || f.variables != 2 {
		t.Fatal(f, e)
	}
	if f, e = ReadCNFDIMACSFile(path, CNFDIMACSReadOptions{MaxBytes: len(source) - 1}); f != nil || !errors.Is(e, ErrLimit) {
		t.Fatal(f, e)
	}
	if f, e = ReadCNFDIMACSFile(path+"missing", CNFDIMACSReadOptions{}); f != nil || e == nil {
		t.Fatal(f, e)
	}
	calls := 0
	f, e = ParseCNFDIMACS(source, CNFDIMACSReadOptions{Limits: CNFTransformOptions{Stop: func() bool { calls++; return calls > 4 }}})
	if f != nil || e != ErrCancelled {
		t.Fatal(f, e)
	}
}
func FuzzCNFRead(f *testing.F) {
	for _, s := range []string{"p cnf 2 2\n1 -2 0\n0\n", "p cnf 0 0\n", "p cnf 3 7\n1 0 ignored\n", "p cnf 2 1\n1\n-2 0\n"} {
		f.Add([]byte(s), false)
		f.Add([]byte(s), true)
	}
	f.Fuzz(func(t *testing.T, data []byte, native bool) {
		out, e := ParseCNFDIMACS(data, CNFDIMACSReadOptions{NativeLines: native, MaxBytes: 4096, MaxLineBytes: 1024, Limits: CNFTransformOptions{MaxVariables: 128, MaxClauses: 128, MaxLiterals: 512}})
		if e != nil {
			if out != nil {
				t.Fatal("partial result")
			}
			return
		}
		if e = checkCNFFormula(out, CNFTransformOptions{MaxVariables: 128, MaxClauses: 128, MaxLiterals: 512}); e != nil {
			t.Fatal(e)
		}
		text, e := out.DIMACS(CNFDIMACSOptions{})
		if e != nil {
			t.Fatal(e)
		}
		round, e := ParseCNFDIMACS(text, CNFDIMACSReadOptions{})
		if e != nil || round.variables != out.variables || !equalSatClauses(round.clauses, out.clauses) {
			t.Fatal(out, round, e)
		}
	})
}
