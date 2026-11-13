package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCec2SimulationSourceEntryPoints(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	a := gia.Gia_ManAppendAnd(g, x, y)
	gia.Gia_ManAppendCo(g, a)
	Cec2_ManSimAlloc(g, 2)
	gia.Gia_ManCecSetSimPattern(g, 70)
	Cec2_ObjSimSetInputBit(g, x>>1, 1)
	Cec2_ObjSimSetInputBit(g, y>>1, 1)
	var manager Cec2_Man_t
	Cec2_ManSimulate(g, nil, &manager)
	if Cec2_ObjSim(g, a>>1)[1]&(1<<6) == 0 {
		t.Fatal("AND simulation failed")
	}
	cex := Cec2_ManDeriveCex(g, 0, 70)
	if cex.PData[0] != 3 {
		t.Fatalf("bad witness %x", cex.PData[0])
	}
	if Cec2_ManSimulateCos(g) != 0 || g.SequentialCounterexample() == nil {
		t.Fatal("missing output witness")
	}
	Cec2_ManCreateClasses(g, &manager)
	triples := vec.Vec_IntAlloc(3)
	vec.Vec_IntPush(triples, 0)
	vec.Vec_IntPush(triples, a>>1)
	vec.Vec_IntPush(triples, 70<<1)
	if count := Cec2_ManSimulate(g, triples, &manager); count != 0 {
		t.Fatalf("counterexample failed to replay: %d", count)
	}
	Cec2_ManSaveCis(g)
	if len(g.SimulationInputs()) != 4 {
		t.Fatal("CI pattern history missing")
	}
}

func TestCec2SimulationHelpersMatchSourceBranches(t *testing.T) {
	t.Run("register output copies every simulation word", func(t *testing.T) {
		g := gia.Gia_ManStart(8)
		gia.Gia_ManAppendCi(g) // primary input
		roLit := gia.Gia_ManAppendCi(g)
		gia.Gia_ManAppendCo(g, 0) // primary output
		gia.Gia_ManAppendCo(g, 0) // register input
		gia.Gia_ManSetRegNum(g, 1)
		Cec2_ManSimAlloc(g, 2)
		ro := roLit >> 1
		ri := gia.Gia_ObjRoToRiId(g, ro)
		Cec2_ObjSim(g, ri)[0], Cec2_ObjSim(g, ri)[1] = 0x123456789abcdef0, 0xfedcba9876543210
		Cec2_ObjSim(g, ro)[0], Cec2_ObjSim(g, ro)[1] = 0, 0
		Cec2_ObjSimRo(g, ro)
		if got, want := Cec2_ObjSim(g, ro), Cec2_ObjSim(g, ri); got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("RO simulation = %x, want RI simulation %x", got, want)
		}
	})

	t.Run("random CI setup preserves the reserved zero pattern", func(t *testing.T) {
		g := gia.Gia_ManStart(4)
		x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		Cec2_ManSimAlloc(g, 3)
		gia.Gia_ManCecSetSimPattern(g, 37)
		Cec2_ManSimulateCis(g)
		if gia.Gia_ManCecSimPattern(g) != 0 {
			t.Fatal("SimulateCis did not reset the pattern cursor")
		}
		for _, lit := range []int32{x, y} {
			if got := Cec2_ObjSim(g, lit>>1)[0] & 1; got != 0 {
				t.Fatalf("random CI set reserved pattern bit: %d", got)
			}
		}
	})

	t.Run("hash uses source phase normalization and 32-bit words", func(t *testing.T) {
		sim := []uint64{0x800000007fffffff, 0xfedcba9876543210}
		const tableSize uint32 = 1009
		primes := [...]uint32{1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147}
		words := []uint32{uint32(sim[0]), uint32(sim[0] >> 32), uint32(sim[1]), uint32(sim[1] >> 32)}
		var hash uint32
		phase := words[0]&1 != 0
		for i, word := range words {
			if phase {
				word = ^word
			}
			hash ^= word * primes[i]
		}
		got := Cec2_ManSimHashKey(sim, int32(len(sim)), int32(tableSize))
		if want := int32(hash % tableSize); got != want {
			t.Fatalf("hash=%d, want source hash %d", got, want)
		}
		complement := []uint64{^sim[0], ^sim[1]}
		if gotComplement := Cec2_ManSimHashKey(complement, 2, int32(tableSize)); gotComplement != got {
			t.Fatalf("complement hash=%d, want %d", gotComplement, got)
		}
	})

	t.Run("direct class refinement separates a mismatching signature", func(t *testing.T) {
		g := gia.Gia_ManStart(12)
		x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		a := gia.Gia_ManAppendAnd(g, x, y) >> 1
		b := gia.Gia_ManAppendAnd(g, x, z) >> 1
		c := gia.Gia_ManAppendAnd(g, y, z) >> 1
		g.StartRepresentatives()
		gia.Gia_ObjSetRepr(g, b, a)
		gia.Gia_ObjSetRepr(g, c, a)
		gia.Gia_ObjSetNext(g, a, b)
		gia.Gia_ObjSetNext(g, b, c)
		gia.Gia_ObjSetNext(g, c, -1)
		Cec2_ManSimAlloc(g, 1)
		Cec2_ObjSim(g, a)[0] = 0b0011 // phase bit is zero
		Cec2_ObjSim(g, b)[0] = 0b0011 // remains equivalent to the head
		Cec2_ObjSim(g, c)[0] = 0b0101 // moves to a new class
		Cec2_ManSimClassRefineOne(g, a)
		if gia.Gia_ObjNext(g, a) != b || gia.Gia_ObjNext(g, b) != -1 {
			t.Fatalf("head class links = (%d, %d), want (%d, -1)", gia.Gia_ObjNext(g, a), gia.Gia_ObjNext(g, b), b)
		}
		if gia.Gia_ObjRepr(g, c) != gia.GIA_VOID || gia.Gia_ObjNext(g, c) != -1 {
			t.Fatalf("mismatch remains represented: repr=%d next=%d", gia.Gia_ObjRepr(g, c), gia.Gia_ObjNext(g, c))
		}
	})
}
