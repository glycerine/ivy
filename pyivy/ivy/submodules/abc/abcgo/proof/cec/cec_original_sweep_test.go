package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestOriginalSweepPreservesSourcePhaseNormalizedMiter(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, x^1)
	p := Cec_ParFra_t{nWords: 1, nRounds: 1, nItersMax: 2, nBTLimit: 100, fCheckMiter: 1, fDualOut: 1, iOutFail: -1}
	out, err := cecOriginalSweep(g, &p)
	if err != nil || out == nil || p.iOutFail != -1 || g.CombinationalCounterexample() != nil {
		t.Fatalf("out %v error %v failed output %d witness %+v", out, err, p.iOutFail, g.CombinationalCounterexample())
	}
}

func TestOriginalSweepEquivalentAlternatives(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, a, b), c)
	y := gia.Gia_ManAppendAnd(g, a, gia.Gia_ManAppendAnd(g, b, c))
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, y)
	p := Cec_ParFra_t{nWords: 1, nRounds: 1, nItersMax: 4, nBTLimit: 100, nDepthMax: 1, fCheckMiter: 1, fDualOut: 1}
	out, err := cecOriginalSweep(g, &p)
	if err != nil || out == nil {
		t.Fatalf("out %v error %v", out, err)
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, e := out.Evaluate(in)
		// Dual-output reduction may replace a proven equal pair by zero.
		if e != nil || len(got) != 2 || got[1] != got[0] {
			t.Fatalf("bits %d out %v error %v", bits, got, e)
		}
	}
}

func TestOriginalSweepPreservesNonMiterOutputs(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, a, b), c)
	y := gia.Gia_ManAppendAnd(g, a, gia.Gia_ManAppendAnd(g, b, c))
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, y^1)
	p := Cec_ParFra_t{nWords: 1, nRounds: 1, nItersMax: 4, nBTLimit: 100, nDepthMax: 1}
	out, err := cecOriginalSweep(g, &p)
	if err != nil || out == nil {
		t.Fatalf("out %v error %v", out, err)
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, e := out.Evaluate(in)
		if e != nil || len(got) != 2 || got[0] != (bits == 7) || got[1] != (bits != 7) {
			t.Fatalf("bits %d out %v error %v", bits, got, e)
		}
	}
}

func TestOriginalSweepUsesPreloadedSimulation(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, b)
	gia.Gia_ManCecSimAlloc(g, 1)
	// Pattern zero is reserved for the all-zero input; pattern one differs.
	gia.Gia_ManCecSaveCIWord(g, 2)
	gia.Gia_ManCecSaveCIWord(g, 0)
	p := Cec_ParFra_t{nWords: 1, nRounds: 1, nItersMax: 2, nBTLimit: 100, fCheckMiter: 1, fDualOut: 1}
	out, err := cecOriginalSweep(g, &p)
	if err != nil || out != nil || p.iOutFail != 0 || g.CombinationalCounterexample() == nil {
		t.Fatalf("out %v error %v failed output %d witness %+v", out, err, p.iOutFail, g.CombinationalCounterexample())
	}
	cex := g.CombinationalCounterexample()
	values := []bool{cex.PData[0]&1 != 0, cex.PData[0]&2 != 0}
	outputs, err := g.Evaluate(values)
	if err != nil || outputs[0] == outputs[1] {
		t.Fatalf("invalid witness %v outputs %v error %v", values, outputs, err)
	}
}

func TestOriginalSweepReplaysRareMiterWitness(t *testing.T) {
	g := gia.Gia_ManStart(64)
	inputs := make([]int32, 16)
	for i := range inputs {
		inputs[i] = gia.Gia_ManAppendCi(g)
	}
	x := inputs[0]
	for _, in := range inputs[1:] {
		x = gia.Gia_ManAppendAnd(g, x, in)
	}
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, x)
	p := Cec_ParFra_t{nWords: 1, nRounds: 1, nItersMax: 100, nBTLimit: 1000, nDepthMax: 1, fCheckMiter: 1, fDualOut: 1}
	out, err := cecOriginalSweep(g, &p)
	if err != nil || out != nil || p.iOutFail != 0 || g.CombinationalCounterexample() == nil {
		t.Fatalf("out %v error %v failed output %d witness %+v", out, err, p.iOutFail, g.CombinationalCounterexample())
	}
	cex := g.CombinationalCounterexample()
	values := make([]bool, len(inputs))
	for i := range values {
		values[i] = cex.PData[i>>5]&(uint32(1)<<uint(i&31)) != 0
	}
	outputs, err := g.Evaluate(values)
	if err != nil || outputs[0] == outputs[1] {
		t.Fatalf("invalid witness %v outputs %v error %v", values, outputs, err)
	}
}

// TestOriginalSweepABC compares the same source workflow against the checked-out
// ABC executable when available. Each result is independently truth-tabled.
func TestOriginalSweepABC(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	binary := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../abc"))
	if _, err := os.Stat(binary); err != nil {
		t.Skip("original ABC executable unavailable")
	}
	g := gia.Gia_ManStart(32)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, a, b), c)
	y := gia.Gia_ManAppendAnd(g, a, gia.Gia_ManAppendAnd(g, b, c))
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, y^1)
	data, err := gia.MarshalAiger(g, gia.AigerWriteOptions{})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	in, out := filepath.Join(dir, "in.aig"), filepath.Join(dir, "out.aig")
	if err = os.WriteFile(in, data, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := fmt.Sprintf("&r -s %s; &fraig -W 1 -R 1 -I 10 -D 1 -C 100; &w %s", in, out)
	if log, err := exec.Command(binary, "-c", cmd).CombinedOutput(); err != nil {
		t.Fatalf("ABC: %v %s", err, log)
	}
	data, err = os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	original, err := gia.ParseAiger(data, gia.AigerReadOptions{SkipStrash: true})
	if err != nil {
		t.Fatal(err)
	}
	p := Cec_ParFra_t{nWords: 1, nRounds: 1, nItersMax: 10, nBTLimit: 100, nDepthMax: 1}
	got := Cec_ManSatSweeping(g, &p, 1)
	if p.Error != nil || got == nil {
		t.Fatalf("sweep: %v", p.Error)
	}
	for bits := 0; bits < 8; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		want, e := original.Evaluate(inputs)
		if e != nil {
			t.Fatal(e)
		}
		actual, e := got.Evaluate(inputs)
		if e != nil {
			t.Fatal(e)
		}
		if len(want) != 2 || len(actual) != 2 || actual[0] != want[0] || actual[1] != want[1] {
			t.Fatalf("bits %d ABC %v Go %v", bits, want, actual)
		}
	}
}
