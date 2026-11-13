package cec

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestCec3GraphSimulationAgainstEvaluation(t *testing.T) {
	for flags := int32(0); flags < 8; flags++ {
		graph := gia.Gia_ManStart(8)
		a, b := gia.Gia_ManAppendCi(graph), gia.Gia_ManAppendCi(graph)
		and := gia.Gia_ManAppendAnd(graph, a^(flags&1), b^((flags>>1)&1))
		gia.Gia_ManAppendCo(graph, and^((flags>>2)&1))
		out := gia.Gia_ObjId(graph, gia.Gia_ManCo(graph, 0))
		sim := newCec3Simulation(graph.ObjectCount(), 2)
		for bit := 0; bit < 128; bit++ {
			sim.setInputBit(a>>1, bit, bit&1 != 0)
			sim.setInputBit(b>>1, bit, bit&2 != 0)
		}
		sim.simulateAnd(graph, and>>1)
		sim.simulateOutput(graph, out)
		for bit := 0; bit < 128; bit++ {
			outputs, err := graph.Evaluate([]bool{bit&1 != 0, bit&2 != 0})
			if err != nil {
				t.Fatal(err)
			}
			got := sim.object(out)[bit>>6]&(uint64(1)<<uint(bit&63)) != 0
			if got != outputs[0] {
				t.Fatalf("flags=%d bit=%d got=%v want=%v", flags, bit, got, outputs[0])
			}
		}
	}
}

func TestCec3RefinesPhaseNormalizedClasses(t *testing.T) {
	graph := gia.Gia_ManStart(8)
	for i := 0; i < 5; i++ {
		gia.Gia_ManAppendCi(graph)
	}
	graph.StartRepresentatives()
	for id := int32(2); id <= 5; id++ {
		gia.Gia_ObjSetRepr(graph, id, 1)
	}
	for id := int32(1); id < 5; id++ {
		gia.Gia_ObjSetNext(graph, id, id+1)
	}
	gia.Gia_ObjSetNext(graph, 5, -1)
	sim := newCec3Simulation(graph.ObjectCount(), 1)
	x, y := uint64(0xaaaaaaaaaaaaaaaa), uint64(0xcccccccccccccccc)
	for id, value := range []uint64{0, x, ^x, y, x, ^y} {
		sim.object(int32(id))[0] = value
	}
	sim.refineClass(graph, 1)
	if gia.Gia_ObjNext(graph, 1) != 2 || gia.Gia_ObjNext(graph, 2) != 4 || gia.Gia_ObjNext(graph, 4) != -1 ||
		gia.Gia_ObjRepr(graph, 3) != gia.GIA_VOID || gia.Gia_ObjNext(graph, 3) != 5 || gia.Gia_ObjRepr(graph, 5) != 3 || gia.Gia_ObjNext(graph, 5) != -1 {
		t.Fatal("simulation refinement did not preserve expected classes")
	}
}

func TestCec3CreateClasses(t *testing.T) {
	graph := gia.Gia_ManStart(10)
	for i := 0; i < 6; i++ {
		gia.Gia_ManAppendCi(graph)
	}
	gia.Gia_ManAppendCo(graph, 2)
	sim := newCec3Simulation(graph.ObjectCount(), 1)
	x, y := uint64(0xaaaaaaaaaaaaaaaa), uint64(0xcccccccccccccccc)
	for id, value := range []uint64{0, x, ^x, y, x, ^y, ^uint64(0), x} {
		sim.object(int32(id))[0] = value
	}
	sim.createClasses(graph)
	for id, want := range map[int32]int32{1: gia.GIA_VOID, 2: 1, 3: gia.GIA_VOID, 4: 1, 5: 3, 6: 0, 7: gia.GIA_VOID} {
		if got := gia.Gia_ObjRepr(graph, id); got != want {
			t.Errorf("representative of %d = %d, want %d", id, got, want)
		}
	}
}
