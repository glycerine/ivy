package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestMirrorDerivationNativeMuxAndBuffer(t *testing.T) {
	g := gia.Gia_ManStart(16)
	g.EnableMuxes()
	control, a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	mux := gia.Gia_ManAppendMuxReal(g, control, a, b)
	buffer := gia.Gia_ManAppendBuf(g, mux^1)
	gia.Gia_ManAppendCo(g, buffer)
	mirrors := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(g)))
	vec.Vec_IntWriteEntry(mirrors, int(control>>1), a^1)
	result := Acec_ManDerive(g, mirrors)
	for mask := 0; mask < 8; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		transformed := []bool{!input[1], input[1], input[2]}
		want, e1 := g.Evaluate(transformed)
		got, e2 := result.Evaluate(input)
		if e1 != nil || e2 != nil || len(got) != 1 || got[0] != want[0] {
			t.Fatalf("input=%v got=%v want=%v errors %v/%v", input, got, want, e1, e2)
		}
	}
}
func TestMirrorDerivationSimplifiesCoincidentXor(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendXorReal(g, a, b)
	gia.Gia_ManAppendCo(g, x)
	mirrors := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(g)))
	vec.Vec_IntWriteEntry(mirrors, int(b>>1), a)
	result := Acec_ManDerive(g, mirrors)
	for mask := 0; mask < 4; mask++ {
		out, err := result.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
		if err != nil || len(out) != 1 || out[0] {
			t.Fatalf("mask=%d out=%v err=%v", mask, out, err)
		}
	}
}
