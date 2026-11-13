package acec

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// Compiled ABC's acec_phase.c reports one tree and these input/output phases.
func TestSourceAdderTreePhaseAnalysis(t *testing.T) {
	p := reviewAdderChain(-1)
	adds := Ree_ManComputeCuts(p, nil, 0)
	var ins, outs *Vec_Int_t
	order := Gia_PolynCoreOrder(p, adds, nil, &ins, &outs)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old; r.Close(); w.Close() }()
	trees := Pas_ManComputeCuts(p, adds, order, ins, outs)
	w.Close()
	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if trees != 1 {
		t.Fatalf("tree count %d, C reports 1", trees)
	}
	for _, want := range []string{"Outputs: 11(0) 18(0) 19(1)", "Inputs: 1(0) 2(0) 3(0) 4(0) 5(0)", "Compl = 0"} {
		if !strings.Contains(string(output), want) {
			t.Fatalf("missing %q in %q", want, output)
		}
	}
	if strings.Contains(string(output), "is wrong") {
		t.Fatalf("invalid full-adder phases: %s", output)
	}
}

func TestSourceAdderPhaseCountsDisconnectedTrees(t *testing.T) {
	p := gia.Gia_ManStart(128)
	in := make([]int32, 6)
	for i := range in {
		in[i] = gia.Gia_ManAppendCi(p)
	}
	var a, b [2]int32
	Acec_InsertFadd(p, in[:3], a[:])
	Acec_InsertFadd(p, in[3:], b[:])
	for _, lit := range []int32{a[0], a[1], b[0], b[1]} {
		gia.Gia_ManAppendCo(p, lit)
	}
	adds := Ree_ManComputeCuts(p, nil, 0)
	var ins, outs *Vec_Int_t
	order := Gia_PolynCoreOrder(p, adds, nil, &ins, &outs)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old; r.Close(); w.Close() }()
	trees := Pas_ManComputeCuts(p, adds, order, ins, outs)
	w.Close()
	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if trees != 2 || !strings.Contains(string(output), "Outputs: 12(0) 13(1) 19(0) 20(1)") || strings.Contains(string(output), "is wrong") {
		t.Fatalf("C reports two trees and phases 12(0) 13(1) 19(0) 20(1); got trees=%d output=%q", trees, output)
	}
}
