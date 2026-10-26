package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestCNFDiagnosticNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_print.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Readable, Empty int
		Text            string
	}
	if e = json.Unmarshal(data, &rows); e != nil || len(rows) != 4 {
		t.Fatal(len(rows), e)
	}
	for _, row := range rows {
		variables := 5
		clauses := [][]SatLit{{0, 3, 4}, {}, {7, 8, 1, 2}, {5, 6, 9}}
		if row.Empty != 0 {
			variables, clauses = 0, nil
		}
		f, e := NewCNFFormula(variables, clauses, CNFTransformOptions{})
		if e != nil {
			t.Fatal(e)
		}
		text, e := f.DiagnosticText(CNFDiagnosticOptions{Readable: row.Readable != 0})
		if e != nil || string(text) != row.Text {
			t.Fatal(row, string(text), e)
		}
	}
}

func TestCNFDiagnosticBoundsWrappersAndVariables(t *testing.T) {
	f, e := NewCNFFormula(2, [][]SatLit{{1, 2}, {}}, CNFTransformOptions{})
	if e != nil {
		t.Fatal(e)
	}
	text, e := f.DiagnosticText(CNFDiagnosticOptions{Readable: true})
	if e != nil || !strings.Contains(string(text), "-0 1 \n\n") {
		t.Fatal(string(text), e)
	}
	exact, e := f.DiagnosticText(CNFDiagnosticOptions{Readable: true, MaxBytes: len(text)})
	if e != nil || string(exact) != string(text) {
		t.Fatal(string(exact), e)
	}
	for _, o := range []CNFDiagnosticOptions{{MaxBytes: -1}, {MaxBytes: len(text) - 1}, {Limits: CNFTransformOptions{Stop: func() bool { return true }}}} {
		out, e := f.DiagnosticText(o)
		if out != nil || !(errors.Is(e, ErrDimension) || errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(out, e)
		}
	}
	if out, e := (*CNFFormula)(nil).DiagnosticText(CNFDiagnosticOptions{}); out != nil || e != ErrDimension {
		t.Fatal(out, e)
	}

	g := mappedCNFFixture(t, 0, 0)
	c, e := DeriveAigCNF4(g, 2, AigCNFDeriveOptions{})
	if e != nil {
		t.Fatal(e)
	}
	ci, e := c.CIVariables(g, CNFVariableMapOptions{})
	if e != nil {
		t.Fatal(e)
	}
	pi, e := c.PIVariables(g, CNFVariableMapOptions{})
	if e != nil || !slices.Equal(ci, pi) {
		t.Fatal(ci, pi, e)
	}
	co, e := c.COVariables(g, CNFVariableMapOptions{})
	if e != nil {
		t.Fatal(e)
	}
	objectVariables := c.ObjectVariables()
	for i, id := range g.cis {
		if ci[i] != objectVariables[id] {
			t.Fatal(i, ci, objectVariables)
		}
	}
	for i, id := range g.cos {
		if co[i] != objectVariables[id] {
			t.Fatal(i, co, objectVariables)
		}
	}
	ci[0] = -99
	co[0] = -99
	if c.ObjectVariables()[g.cis[0]] == -99 || c.ObjectVariables()[g.cos[0]] == -99 {
		t.Fatal("returned slices alias CNF")
	}
	asserted, e := DeriveAigCNF4(g, 0, AigCNFDeriveOptions{})
	if e != nil {
		t.Fatal(e)
	}
	co, e = asserted.COVariables(g, CNFVariableMapOptions{})
	if e != nil || !slices.Equal(co, []int{-1, -1}) {
		t.Fatal(co, e)
	}
	for _, o := range []CNFVariableMapOptions{{MaxObjects: -1}, {MaxObjects: 1}, {Stop: func() bool { return true }}} {
		got, e := c.CIVariables(g, o)
		if got != nil || !(errors.Is(e, ErrDimension) || errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(got, e)
		}
	}
	if got, e := c.CIVariables(nil, CNFVariableMapOptions{}); got != nil || e != ErrDimension {
		t.Fatal(got, e)
	}
	bad := c.Clone()
	bad.objectVariables[g.cis[0]] = bad.variables
	if got, e := bad.CIVariables(g, CNFVariableMapOptions{}); got != nil || e != ErrDimension {
		t.Fatal(got, e)
	}
	if mappedText, e := c.DiagnosticText(CNFDiagnosticOptions{}); e != nil || len(mappedText) == 0 {
		t.Fatal(string(mappedText), e)
	}
	simple, e := EncodeAigCNF(g)
	if e != nil {
		t.Fatal(e)
	}
	if simpleText, e := simple.DiagnosticText(CNFDiagnosticOptions{}); e != nil || len(simpleText) == 0 {
		t.Fatal(string(simpleText), e)
	}
	frames, e := c.UnrollFrames(g, 2, false, CNFFrameOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if frameText, e := frames.DiagnosticText(CNFDiagnosticOptions{}); e != nil || len(frameText) == 0 {
		t.Fatal(string(frameText), e)
	}
}

func FuzzCNFDiagnostic(f *testing.F) {
	f.Add([]byte{0, 3, 4, 7, 8, 1, 2, 5, 6, 9}, true)
	f.Fuzz(func(t *testing.T, data []byte, readable bool) {
		if len(data) > 512 {
			data = data[:512]
		}
		flat := make([]SatLit, len(data))
		for i, b := range data {
			flat[i] = SatLit(b % 32)
		}
		middle := len(flat) / 2
		clauses := [][]SatLit{flat[:middle], {}, flat[middle:]}
		formula, e := NewCNFFormula(16, clauses, CNFTransformOptions{})
		if e != nil {
			t.Fatal(e)
		}
		got, e := formula.DiagnosticText(CNFDiagnosticOptions{Readable: readable})
		if e != nil {
			t.Fatal(e)
		}
		var want strings.Builder
		want.WriteString("p cnf 16 3\n")
		for _, clause := range clauses {
			for _, l := range clause {
				if l.Negative() {
					want.WriteByte('-')
				}
				n := int(l.Var())
				if !readable {
					n++
				}
				want.WriteString(strconv.Itoa(n))
				want.WriteByte(' ')
			}
			want.WriteByte('\n')
		}
		want.WriteByte('\n')
		if string(got) != want.String() {
			t.Fatal(string(got), want.String())
		}
	})
}
