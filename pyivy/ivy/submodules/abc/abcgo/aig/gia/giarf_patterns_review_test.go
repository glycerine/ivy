package gia

import (
	"slices"
	"testing"
)

func TestGiarfPatternTryAllBitsAndAtomicConflict(t *testing.T) {
	for bit := int32(0); bit < 32; bit++ {
		g := Gia_ManStart(1)
		for i := 0; i < 3; i++ {
			Gia_ManAppendCi(g)
		}
		p := Gia_ManEquivStart(g, 10, 0)
		mask := uint32(1) << uint(bit)
		for i, word := range []uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0} {
			Hcd_ObjSetSim(p, Gia_ManCi(g, int32(i)).id, word)
		}
		before := slices.Clone(p.pSimInfo)
		// The trailing conflicting literal is outside nLits and must be ignored.
		if Gia_GiarfStorePatternTry(p.vSimInfo, p.vSimPres, bit, []int32{0, 3, 1}, 2) != 1 {
			t.Fatalf("bit %d: compatible cube rejected", bit)
		}
		for i := int32(0); i < 3; i++ {
			id := Gia_ManCi(g, i).id
			want, present := before[id], uint32(0)
			if i == 0 {
				want |= mask
				present = mask
			} else if i == 1 {
				want &^= mask
				present = mask
			}
			if got := Hcd_ObjSim(p, id); got != want || Vec_PtrEntry(p.vSimPres, int(i)).([]uint32)[0] != present {
				t.Fatalf("bit %d input %d: data/presence changed incorrectly", bit, i)
			}
		}
		before = slices.Clone(p.pSimInfo)
		// Input 2 is free, but input 0 conflicts. Even the first write must be
		// withheld, and no presence bit may be installed on input 2.
		if Gia_GiarfStorePatternTry(p.vSimInfo, p.vSimPres, bit, []int32{4, 1}, 2) != 0 {
			t.Fatalf("bit %d: conflicting cube accepted", bit)
		}
		if !slices.Equal(before, p.pSimInfo) || Vec_PtrEntry(p.vSimPres, 2).([]uint32)[0] != 0 {
			t.Fatalf("bit %d: rejection modified simulation or presence", bit)
		}
		if Gia_GiarfStorePatternTry(p.vSimInfo, p.vSimPres, bit, []int32{0, 3}, 2) != 1 ||
			Gia_GiarfStorePatternTry(p.vSimInfo, p.vSimPres, bit, nil, 0) != 1 || !slices.Equal(before, p.pSimInfo) {
			t.Fatalf("bit %d: repeated/empty cube was not idempotent", bit)
		}
	}
}

func TestGiarfPatternPackingExhaustsExactly31Slots(t *testing.T) {
	g := Gia_ManStart(1)
	for i := 0; i < 5; i++ {
		Gia_ManAppendCi(g)
	}
	p := Gia_ManEquivStart(g, 10, 0)
	for i := int32(0); i < 5; i++ {
		Hcd_ObjSetSim(p, Gia_ManCi(g, i).id, 1) // reserved phase bit
	}
	cube := Vec_IntAlloc(5)
	// Distinct complete cubes cannot share a slot. There are 32 possible
	// assignments; only 31 non-phase bits are available.
	for assignment := int32(0); assignment < 32; assignment++ {
		Vec_IntClear(cube)
		for input := int32(0); input < 5; input++ {
			Vec_IntPush(cube, 2*input+1-((assignment>>uint(input))&1))
		}
		before := slices.Clone(p.pSimInfo)
		want := int32(1)
		if assignment == 31 {
			want = 0
		}
		if got := Gia_GiarfStorePattern(p.vSimInfo, p.vSimPres, cube); got != want {
			t.Fatalf("assignment %d: accepted=%d, want %d", assignment, got, want)
		}
		if want == 0 && !slices.Equal(before, p.pSimInfo) {
			t.Fatal("exhausted packing modified simulation")
		}
	}
	for input := int32(0); input < 5; input++ {
		want := uint32(1)
		for assignment := uint32(0); assignment < 31; assignment++ {
			want |= ((assignment >> uint(input)) & 1) << (assignment + 1)
		}
		if got := Hcd_ObjSim(p, Gia_ManCi(g, input).id); got != want {
			t.Fatalf("input %d: packed=%08x, want %08x", input, got, want)
		}
		if present := Vec_PtrEntry(p.vSimPres, int(input)).([]uint32)[0]; present != 0xfffffffe {
			t.Fatalf("input %d: presence=%08x", input, present)
		}
	}
}

func TestGiarfInsertPatternUsesCIOrdinals(t *testing.T) {
	g := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCi(g) // ordinal 2 is object 4, not object 3
	Gia_ManAppendCi(g)
	p := Gia_ManEquivStart(g, 10, 0)
	cube := Vec_IntAlloc(2)
	Vec_IntPush(cube, 1) // first CI false
	Vec_IntPush(cube, 4) // third CI true
	for _, bit := range []int32{0, 1, 31} {
		for id := range p.pSimInfo {
			p.pSimInfo[id] = 0xaaaaaaaa
		}
		before := slices.Clone(p.pSimInfo)
		Gia_GiarfInsertPattern(p, cube, bit)
		mask := uint32(1) << uint(bit)
		for id, word := range p.pSimInfo {
			want := before[id]
			if int32(id) == Gia_ManCi(g, 0).id {
				want &^= mask
			} else if int32(id) == Gia_ManCi(g, 2).id {
				want |= mask
			}
			if word != want {
				t.Fatalf("bit %d object %d: got %08x, want %08x", bit, id, word, want)
			}
		}
	}
}
