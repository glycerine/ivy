package cudd

import (
	"bytes"
	"math"
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

type nestedMintermWriter struct {
	output bytes.Buffer
	before func()
}

func (w *nestedMintermWriter) Write(p []byte) (int, error) {
	if f := w.before; f != nil {
		w.before = nil
		f()
	}
	return w.output.Write(p)
}

func TestMintermWriterCanPrintAnotherManager(t *testing.T) {
	a, b := NewBDDManager(1), NewBDDManager(1)
	defer core.Cudd_Quit(a)
	defer core.Cudd_Quit(b)
	var nested bytes.Buffer
	core.Cudd_SetOutput(b, &nested, nil)
	w := nestedMintermWriter{before: func() { Cudd_PrintMinterm(b, b.One) }}
	core.Cudd_SetOutput(a, &w, nil)
	Cudd_PrintMinterm(a, Cudd_Not(core.Cudd_bddIthVar(a, 0)))
	if got := w.output.String(); got != "0  1\n" || nested.String() != "-  1\n" {
		t.Fatalf("outer and nested covers %q, %q", got, nested.String())
	}
}

func TestMintermPrintsNativeTerminalValues(t *testing.T) {
	dd := NewBDDManager(1)
	defer core.Cudd_Quit(dd)
	var out bytes.Buffer
	core.Cudd_SetOutput(dd, &out, nil)
	for _, test := range []struct {
		value float64
		line  string
	}{
		{12345.6789, "-  12345.7\n"},
		{-0.000123456789, "- -0.000123457\n"},
		{1e7, "-  1e+07\n"},
		{math.Inf(1), "-  inf\n"},
		{math.Inf(-1), "- -inf\n"},
	} {
		out.Reset()
		Cudd_PrintMinterm(dd, core.Cudd_addConst(dd, test.value))
		if got := out.String(); got != test.line {
			t.Errorf("value %g gave %q, want %q", test.value, got, test.line)
		}
	}
}

func BenchmarkPrintParityMinterms(b *testing.B) {
	dd := NewBDDManager(8)
	defer core.Cudd_Quit(dd)
	f := dd.LogicZero
	for i := range 8 {
		f = core.Cudd_bddXor(dd, f, core.Cudd_bddIthVar(dd, i))
	}
	var out bytes.Buffer
	core.Cudd_SetOutput(dd, &out, nil)
	b.ReportAllocs()
	for b.Loop() {
		out.Reset()
		if Cudd_PrintMinterm(dd, f) == 0 || out.Len() != 128*12 {
			b.Fatal("incomplete parity cover", out.Len())
		}
	}
}
