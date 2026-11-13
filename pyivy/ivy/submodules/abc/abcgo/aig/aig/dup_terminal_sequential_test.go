package aig

import "testing"

// Aig_ManDupRepr in aigRepr.c calls Aig_ManSetRegNum after creating
// every terminal, even when no representatives have been merged.
func TestSequentialRepresentativeDupTerminalMetadata(t *testing.T) {
	p := NewGraph()
	x, y, q := p.NewInput(), p.NewInput(), p.NewInput()
	p.AddOutput(p.AND(x, q))
	p.AddOutput(y)
	p.AddOutput(x)
	Aig_ManSetRegNum(p, 1)
	Aig_ManReprStart(p, int32(len(p.nodes)))
	for _, ordered := range []int32{0, 1} {
		copy := Aig_ManDupRepr(p, ordered)
		if copy.RegisterCount() != 1 || copy.TruePiCount() != 2 || copy.TruePoCount() != 2 {
			t.Fatalf("ordered=%d: regs/PIs/POs = %d/%d/%d", ordered, copy.RegisterCount(), copy.TruePiCount(), copy.TruePoCount())
		}
		for i, input := range copy.inputs {
			if got := Aig_ObjCioId(copy, input); got != int32(i) {
				t.Fatalf("ordered=%d: input %d has CioId %d", ordered, i, got)
			}
		}
		for i, output := range copy.outputs {
			if got := Aig_ObjCioId(copy, output); got != int32(i) {
				t.Fatalf("ordered=%d: output %d has CioId %d", ordered, i, got)
			}
		}
	}
}
