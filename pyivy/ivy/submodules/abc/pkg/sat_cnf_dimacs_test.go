package abc

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestCNFDIMACSNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_dimacs.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Readable, Prefix, Empty int
		Text                    string
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 10 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		variables := 5
		clauses := [][]SatLit{{0, 3, 4}, {}, {7, 8, 1, 2}, {5, 6, 9}}
		if row.Empty != 0 {
			variables = 0
			clauses = nil
		}
		f, e := NewCNFFormula(variables, clauses, CNFTransformOptions{})
		if e != nil {
			t.Fatal(e)
		}
		options := CNFDIMACSOptions{Readable: row.Readable != 0}
		if row.Prefix&1 != 0 {
			options.ForAll = []SatVar{0, 2, 4}
		}
		if row.Prefix&2 != 0 {
			options.Exists = []SatVar{1, 3}
		}
		got, e := f.DIMACS(options)
		if e != nil || string(got) != row.Text {
			t.Fatal(row, string(got), e)
		}
	}
}
func TestCNFDIMACSBoundsAndWrappers(t *testing.T) {
	f, e := NewCNFFormula(2, [][]SatLit{{0, 3}, {}}, CNFTransformOptions{})
	if e != nil {
		t.Fatal(e)
	}
	out, e := f.DIMACS(CNFDIMACSOptions{ForAll: []SatVar{}, Exists: []SatVar{}})
	if e != nil || !strings.Contains(string(out), "a 0\ne 0\n") {
		t.Fatal(string(out), e)
	}
	plain, e := f.DIMACS(CNFDIMACSOptions{})
	if e != nil {
		t.Fatal(e)
	}
	exact, e := f.DIMACS(CNFDIMACSOptions{MaxBytes: len(plain)})
	if e != nil || string(exact) != string(plain) {
		t.Fatal(e)
	}
	for _, o := range []CNFDIMACSOptions{{MaxBytes: len(plain) - 1}, {Limits: CNFTransformOptions{Stop: func() bool { return true }}}} {
		out, e = f.DIMACS(o)
		if out != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(out, e)
		}
	}
	if out, e = f.DIMACS(CNFDIMACSOptions{ForAll: []SatVar{2}}); out != nil || e != ErrInvalidLiteral {
		t.Fatal(out, e)
	}
	g := mappedCNFFixture(t, 0, 0)
	mapped, e := DeriveAigCNF4(g, 2, AigCNFDeriveOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = mapped.DIMACS(CNFDIMACSOptions{}); e != nil {
		t.Fatal(e)
	}
	frames, e := mapped.UnrollFrames(g, 2, false, CNFFrameOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = frames.DIMACS(CNFDIMACSOptions{}); e != nil {
		t.Fatal(e)
	}
}
func FuzzCNFDIMACS(f *testing.F) {
	f.Add([]byte{0, 3, 4, 7, 8, 1, 2, 5, 6, 9})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			data = data[:256]
		}
		flat := make([]SatLit, len(data))
		for i, b := range data {
			flat[i] = SatLit(b % 16)
		}
		middle := len(flat) / 2
		clauses := [][]SatLit{flat[:middle], {}, flat[middle:]}
		formula, e := NewCNFFormula(8, clauses, CNFTransformOptions{})
		if e != nil {
			t.Fatal(e)
		}
		out, e := formula.DIMACS(CNFDIMACSOptions{})
		if e != nil {
			t.Fatal(e)
		}
		var decoded [][]SatLit
		var current []SatLit
		header := false
		for _, line := range strings.Split(string(out), "\n") {
			fields := strings.Fields(line)
			if len(fields) == 0 || fields[0] == "c" {
				continue
			}
			if fields[0] == "p" {
				if len(fields) != 4 || fields[1] != "cnf" || fields[2] != "8" || fields[3] != "3" {
					t.Fatal(line)
				}
				header = true
				continue
			}
			for _, field := range fields {
				v, e := strconv.Atoi(field)
				if e != nil {
					t.Fatal(e)
				}
				if v == 0 {
					decoded = append(decoded, current)
					current = nil
					continue
				}
				negative := v < 0
				if negative {
					v = -v
				}
				l := SatLit(v-1) << 1
				if negative {
					l ^= 1
				}
				current = append(current, l)
			}
		}
		if !header || len(current) != 0 || !equalSatClauses(decoded, clauses) {
			t.Fatal(string(out), decoded, clauses)
		}
	})
}
