package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"io"
	"os"
	"strings"
	"testing"
)

func TestSplitVisitsEveryOutputAfterFirstSAT(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, a)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	status := Cec_GiaSplitTest(g, 1, 0, 0, 1, 1, 0, 1)
	w.Close()
	os.Stdout = old
	out, err := io.ReadAll(r)
	r.Close()
	if err != nil {
		t.Fatal(err)
	}
	if status != 0 || !strings.Contains(string(out), "Solving output 0:") || !strings.Contains(string(out), "Solving output 1:") {
		t.Fatalf("status %d output %s", status, out)
	}
	if c := g.CombinationalCounterexample(); c == nil || c.IPo != 0 {
		t.Fatalf("first output witness %+v", c)
	}
}

func TestSplitDirectSerialKeepsWitnessOnUnsat(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, 0)
	if err := g.SetCombinationalCounterexample([]bool{true}, 0); err != nil {
		t.Fatal(err)
	}
	before := g.CombinationalCounterexample()
	_ = a
	if status := Cec_GiaSplitTest2(g, 1, 0, 0, 1, 0, 0, 1); status != 1 || g.CombinationalCounterexample() != before {
		t.Fatalf("status %d witness changed", status)
	}
}

func TestSplitRejectsZeroProcessorCount(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	defer func() {
		if got := recover(); got != "cec: invalid split processor count" {
			t.Fatalf("processor panic = %v", got)
		}
	}()
	Cec_GiaSplitTestInt(g, 0, 0, 0, 1, 0, 0, 1)
}

func TestSplitChoiceZeroLookaheadPanics(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	defer func() {
		if got := recover(); got != "cec: cofactor selection requires a positive lookahead" {
			t.Fatalf("lookahead panic = %v", got)
		}
	}()
	splitChoose(splitPart{graph: g}, 0)
}
