package sim

import (
	"github.com/glycerine/abcgo/internal/testoracle"
	"math/bits"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// TestNativeFifoFlipStaysInRow links test_vectors/fifo_oracle.c (which
// includes simSupp.c from ../src) against ABC_LIBABC. Sim_UtilAssignFromFifo
// writes series of nSuppWords words; with 257 inputs (9 words) the eighth
// series has only one of the 64 words left, and flipping bit i >= 32 wrote
// into the next objects' rows (CBUGS_FIXED.md #14).
func TestNativeFifoFlipStaysInRow(t *testing.T) {
	lib := testoracle.NativeLibABCPath()
	if lib == "" {
		t.Skip("set ABC_LIBABC to a native libabc.a built from ../src")
	}
	lib, _ = filepath.Abs(lib)
	src, _ := filepath.Abs("../../../src")
	dir := t.TempDir()
	bin := filepath.Join(dir, "fifo")
	for _, c := range [][]string{
		{"gcc", "-O1", "-DABC_USE_STDINT_H", "-I", src, "-c", "test_vectors/fifo_oracle.c", "-o", filepath.Join(dir, "main.o")},
		{"g++", filepath.Join(dir, "main.o"), lib, "-lreadline", "-lpthread", "-lm", "-ldl", "-o", bin},
	} {
		if out, err := testoracle.Command(t, c[0], c[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c[:3], err, out)
		}
	}
	out, err := testoracle.Command(t, bin, "../../test_vectors/i10.aig").CombinedOutput()
	if err != nil || !strings.Contains(string(out), "corrupted=0") {
		t.Fatalf("native Sim_UtilAssignFromFifo: %v %s", err, out)
	}
}

// The port flips only the bits that fit in the series, as the fixed ABC.
func TestFifoFlipStaysInRow(t *testing.T) {
	g := aig.NewGraph()
	var ins []aig.Edge
	for i := 0; i < 257; i++ {
		ins = append(ins, g.NewInput())
	}
	g.AddOutput(g.AND(ins[0], ins[256]))
	p := Sim_ManStart(g, false)
	for _, row := range p.Sim0 {
		clear(row)
	}
	for k := 0; k < 8; k++ {
		p.Patterns = append(p.Patterns, Sim_ManPatAlloc(p))
	}
	sourceAssignFromFifo(p)
	supp := SIM_NUM_WORDS(len(p.Inputs))
	input := map[int]int{}
	for i, e := range p.Inputs {
		input[e.NodeID()] = i
	}
	for id, row := range p.Sim0 {
		ones := 0
		for _, w := range row {
			ones += bits.OnesCount32(w)
		}
		want := 0
		if i, ok := input[id]; ok {
			for word := 0; word < p.Words; word += supp {
				if word+i/32 < min(word+supp, p.Words) {
					want++
				}
			}
		}
		if ones != want {
			t.Fatalf("object %d has %d ones, want %d", id, ones, want)
		}
	}
}
