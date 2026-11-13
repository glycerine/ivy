package aig

import "testing"

// In the checked-in C Aig_ManDupPart_rec copies every node using Aig_And,
// including an AIG_OBJ_EXOR created when fCatchExor is enabled.
func TestDupPartNativeXorMatchesCAndCopy(t *testing.T) {
	src := NewGraph()
	x := src.NewInput()
	y := src.NewInput()
	src.AddOutput(src.NewXor(x, y))
	dst := NewGraph()
	var support []int32
	outputs := Aig_ManDupPart(dst, src, []int32{0}, &support, 0)
	dst.AddOutput(outputs[0])
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		got, err := dst.Evaluate(in)
		if err != nil || got[0] != (in[0] && in[1]) {
			t.Fatalf("input=%v output=%v error=%v", in, got, err)
		}
	}
}
