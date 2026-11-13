package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestDuplicateAbstractionFlopsAndGates(t *testing.T) {
	p := gia.Gia_ManStart(20)
	x, q := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gate := gia.Gia_ManAppendAnd(p, x, q)
	gia.Gia_ManAppendCo(p, gate)
	gia.Gia_ManAppendCo(p, x)
	gia.Gia_ManSetRegNum(p, 1)
	for _, included := range []int32{0, 1} {
		classes := vec.Vec_IntStart(1)
		vec.Vec_IntWriteEntry(classes, 0, included)
		out := Gia_ManDupAbsFlops(p, classes)
		if out == nil || gia.Gia_ManRegNum(out) != included {
			t.Fatalf("included=%d: wrong register count", included)
		}
		graph := gia.Gia_ManToAigSimple(out)
		for bits := 0; bits < 4; bits++ {
			outputs, err := graph.Evaluate([]bool{bits&1 != 0, bits&2 != 0})
			if err != nil || outputs[0] != (bits&3 == 3) {
				t.Fatalf("included=%d inputs=%d outputs=%v err=%v", included, bits, outputs, err)
			}
		}
	}
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(p)))
	vec.Vec_IntWriteEntry(classes, 0, 1)
	vec.Vec_IntWriteEntry(classes, int(q>>1), 1)
	vec.Vec_IntWriteEntry(classes, int(gate>>1), 1)
	out := Gia_ManDupAbsGates(p, classes)
	if out == nil || gia.Gia_ManRegNum(out) != 1 {
		t.Fatal("gate abstraction lost register")
	}
	graph := gia.Gia_ManToAigSimple(out)
	for bits := 0; bits < 4; bits++ {
		outputs, err := graph.Evaluate([]bool{bits&1 != 0, bits&2 != 0})
		if err != nil || outputs[0] != (bits&3 == 3) {
			t.Fatalf("gate inputs=%d outputs=%v err=%v", bits, outputs, err)
		}
	}
	// Excluding a register retains its use as a pseudo-input.
	vec.Vec_IntWriteEntry(classes, int(q>>1), 0)
	out = Gia_ManDupAbsGates(p, classes)
	if out == nil || gia.Gia_ManRegNum(out) != 0 {
		t.Fatal("excluded register retained")
	}
	graph = gia.Gia_ManToAigSimple(out)
	for bits := 0; bits < 4; bits++ {
		outputs, err := graph.Evaluate([]bool{bits&1 != 0, bits&2 != 0})
		if err != nil || outputs[0] != (bits&3 == 3) {
			t.Fatalf("pseudo-inputs=%d outputs=%v err=%v", bits, outputs, err)
		}
	}
}
