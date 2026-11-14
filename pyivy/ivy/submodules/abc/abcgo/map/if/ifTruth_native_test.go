package if_

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau/canon"
	"math/rand"
	"testing"
)

func ifTestTable(n int, r *rand.Rand) []uint64 {
	table := make([]uint64, ifTruthWords(n))
	for i := 0; i < 1<<uint(n); i++ {
		if r.Intn(2) != 0 {
			ifTruthSet(table, i)
		}
	}
	if n < 6 {
		pattern := table[0]
		for shift := 1 << uint(n); shift < 64; shift *= 2 {
			pattern |= pattern << uint(shift)
		}
		table[0] = pattern
	}
	return table
}
func ifTestStore(p *If_Man_t, c *If_Cut_t, table []uint64, canonical bool) {
	table = append([]uint64(nil), table...)
	phase := uint32(table[0] & 1)
	if canonical {
		permutation := make([]byte, c.nLeaves)
		phase = canon.Canonicalize(table, int32(c.nLeaves), permutation)
		leaves := append([]int32(nil), ifLeaves(c)...)
		for i := range ifLeaves(c) {
			c.pLeaves[i] = leaves[permutation[i]]
		}
		c.uMaskFunc = int32(phase & ((1 << c.nLeaves) - 1))
		phase >>= c.nLeaves
	} else if phase != 0 {
		for i := range table {
			table[i] = ^table[i]
		}
	}
	ifTruthIntern(p, c, table, phase, false)
	c.uSign = If_ObjCutSignCompute(c)
}
func ifTestValue(p *If_Man_t, c *If_Cut_t, assignment uint32) bool {
	index := 0
	for i, id := range ifLeaves(c) {
		bit := assignment >> uint(id) & 1
		if c.uMaskFunc >= 0 {
			bit ^= uint32(c.uMaskFunc) >> uint(i) & 1
		}
		index |= int(bit) << uint(i)
	}
	v := ifTruthBit(ifTruthRead(p, c, c.iCutFunc), index)
	return v != ((c.iCutFunc&1)^int32(c.fCompl&1) != 0)
}
func TestTruthNativeComposeAndSupport(t *testing.T) {
	r := rand.New(rand.NewSource(121))
	for trial := 0; trial < 100; trial++ {
		for _, canonical := range []bool{false, true} {
			p := &If_Man_t{pPars: &If_Par_t{FCutMin: 1}}
			a, b := ifTestCut(6, 0, 1, 2), ifTestCut(6, 1, 3)
			ifTestStore(p, a, ifTestTable(3, r), canonical)
			ifTestStore(p, b, ifTestTable(2, r), canonical)
			merged := ifTestCut(6)
			if canonical {
				If_CutMerge(p, a, b, merged)
				If_CutComputeTruthPerm(p, merged, a, b, a.iCutFunc, b.iCutFunc)
			} else {
				If_CutMergeOrdered(p, a, b, merged)
				If_CutComputeTruth(p, merged, a, b, 0, 0)
			}
			for assignment := uint32(0); assignment < 16; assignment++ {
				if ifTestValue(p, merged, assignment) != (ifTestValue(p, a, assignment) && ifTestValue(p, b, assignment)) {
					t.Fatal("composed truth mismatch", trial, canonical, assignment, a.pLeaves, b.pLeaves, merged.pLeaves)
				}
			}
		}
	}
}
func TestTruthNativeComplementarySupportAndCache(t *testing.T) {
	p := &If_Man_t{pPars: &If_Par_t{FCutMin: 1}}
	a, b := ifTestCut(6, 0), ifTestCut(6, 0)
	ifTestStore(p, a, []uint64{0xaaaaaaaaaaaaaaaa}, false)
	ifTestStore(p, b, []uint64{0xaaaaaaaaaaaaaaaa}, false)
	c := ifTestCut(6, 0)
	if If_CutComputeTruth(p, c, a, b, 0, 1) != 1 || c.nLeaves != 0 || ifTestValue(p, c, 0) {
		t.Fatal("x AND !x support not removed")
	}
	p = &If_Man_t{pPars: &If_Par_t{FCutMin: 1}}
	for iteration := 0; iteration < 2; iteration++ {
		offset := int32(iteration * 4)
		a, b := ifTestCut(6, offset, offset+1), ifTestCut(6, offset+2, offset+3)
		ifTestStore(p, a, []uint64{0x8888888888888888}, true)
		ifTestStore(p, b, []uint64{0x6666666666666666}, true)
		c := ifTestCut(6)
		If_CutMerge(p, a, b, c)
		If_CutComputeTruthPerm(p, c, a, b, a.iCutFunc, b.iCutFunc)
		for assignment := uint32(0); assignment < 256; assignment++ {
			if ifTestValue(p, c, assignment) != (ifTestValue(p, a, assignment) && ifTestValue(p, b, assignment)) {
				t.Fatal("cached canonical truth changed")
			}
		}
	}
	if p.nCacheHits != 1 || p.nCacheMisses != 1 {
		t.Fatal("canonical pair cache not used", p.nCacheHits, p.nCacheMisses)
	}
}
func TestTruthNativePinPermutationAndOrbit(t *testing.T) {
	table := []uint64{0x8888888888888888}
	delays := []float32{1, 3}
	leaves := []int32{5, 9}
	If_CutTruthPermute(table, 2, 2, 1, delays, leaves)
	if leaves[0] != 9 || delays[0] != 3 || table[0] != 0x8888888888888888 {
		t.Fatal("pin rotation")
	}
	memory := If_DeriveHashTable6(0, 0x123456789abcdef0)
	if vec.Vec_MemEntryNum(memory) != 2 || *vec.Vec_MemHashLookup(memory, []uint64{0x123456789abcdef0}) < 0 {
		t.Fatal("zero-variable orbit changed high truth bits")
	}
	memory = If_DeriveHashTable6(2, 0x8888888888888888)
	if vec.Vec_MemEntryNum(memory) != 8 {
		t.Fatal("AND NPN orbit size", vec.Vec_MemEntryNum(memory))
	}
}
