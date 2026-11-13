package ssc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

// Sixty-six independent rare conjunctions produce enough SAT witnesses to
// fill one 64-pattern word, exercising the source sweep's in-loop flush and
// subsequent partial-batch flush (sscCore.c:310-384).
func TestSweepFlushesFullWitnessBatch(t *testing.T) {
	g := gia.Gia_ManStart(5000)
	const blocks = 66
	const width = 14
	in := make([]int32, blocks*width)
	for i := range in {
		in[i] = gia.Gia_ManAppendCi(g)
	}
	outputs := make([]int32, blocks)
	for j := 0; j < blocks; j++ {
		x := in[j*width]
		for k := 1; k < width; k++ {
			x = gia.Gia_ManAppendAnd(g, x, in[j*width+k])
		}
		outputs[j] = x
	}
	for _, output := range outputs {
		gia.Gia_ManAppendCo(g, output)
	}
	care := gia.Gia_ManStart(int32(len(in) + 2))
	for range in {
		gia.Gia_ManAppendCi(care)
	}
	gia.Gia_ManAppendCo(care, 0)
	p := DefaultParams()
	p.NWords = 1
	out := Ssc_PerformSweepingInt(g, care, &p)
	if out == nil {
		t.Fatal("nil sweep result")
	}
	if gia.Gia_ManIsNormalized(g) == 0 {
		t.Fatal("source fixture is not normalized")
	}
	if len(out.OutputLiterals()) != blocks {
		t.Fatalf("output count = %d", len(out.OutputLiterals()))
	}
	values := make([]bool, len(in))
	for selected := -1; selected < blocks; selected++ {
		clear(values)
		if selected >= 0 {
			for k := 0; k < width; k++ {
				values[selected*width+k] = true
			}
		}
		got, err := out.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		for i, value := range got {
			if value != (i == selected) {
				t.Fatalf("active block %d output %d = %v", selected, i, value)
			}
		}
	}
}
