package extra

import (
	"math/bits"
	"slices"
	"testing"
)

func TestTruthHeaderBitCountsAndPredicates(t *testing.T) {
	for _, tc := range []struct{ bits, words int32 }{{0, 0}, {1, 1}, {31, 1}, {32, 1}, {33, 2}, {1<<31 - 1, 1 << 26}} {
		if Extra_BitWordNum(tc.bits) != tc.words {
			t.Fatal(tc)
		}
	}
	table := make([]uint32, 2)
	Extra_TruthSetBit(table, 31)
	Extra_TruthSetBit(table, 32)
	if Extra_TruthHasBit(table, 31) != 1 || Extra_TruthCountOnes(table, 6) != 2 {
		t.Fatal(table)
	}
	Extra_TruthXorBit(table, 31)
	if Extra_TruthHasBit(table, 31) != 0 {
		t.Fatal(table)
	}
	for n := int32(0); n <= 10; n++ {
		words := make([]uint32, truthWords(n))
		if Extra_TruthIsConst0(words, n) != 1 {
			t.Fatal(n)
		}
		Extra_TruthFill(words, n)
		if Extra_TruthIsConst1(words, n) != 1 || Extra_TruthCountOnes(words, n) != int32(len(words))*32 {
			t.Fatal(n)
		}
		Extra_TruthClear(words, n)
		if Extra_TruthIsConst0(words, n) != 1 {
			t.Fatal(n)
		}
	}
	if Extra_WordCountOnes(0x3456789a) != int32(bits.OnesCount32(0x3456789a)) {
		t.Fatal("popcount")
	}
}
func TestTruthHeaderOperatorsAndAliases(t *testing.T) {
	a, b := []uint32{0x01234567, 0x89abcdef}, []uint32{0xf0f0f0f0, 0xff00ff00}
	for _, tc := range []struct {
		run  func([]uint32, []uint32, []uint32, int32)
		kind byte
	}{{Extra_TruthAnd, '&'}, {Extra_TruthOr, '|'}, {Extra_TruthSharp, '-'}, {Extra_TruthNand, '!'}} {
		want := make([]uint32, 2)
		for i, x := range a {
			switch tc.kind {
			case '&':
				want[i] = x & b[i]
			case '|':
				want[i] = x | b[i]
			case '-':
				want[i] = x &^ b[i]
			case '!':
				want[i] = ^(x & b[i])
			}
		}
		got := slices.Clone(a)
		tc.run(got, got, b, 6)
		if !slices.Equal(got, want) {
			t.Fatal(tc.kind, got, want)
		}
		data := append(slices.Clone(a), uint32(0))
		tc.run(data[1:], data[:2], b, 6)
		if !slices.Equal(data[1:], want) {
			t.Fatal("partial alias", tc.kind)
		}
	}
	for c0 := int32(0); c0 < 2; c0++ {
		for c1 := int32(0); c1 < 2; c1++ {
			got := make([]uint32, 2)
			Extra_TruthAndPhase(got, a, b, 6, c0, c1)
			for i, x := range a {
				y := b[i]
				if c0 != 0 {
					x = ^x
				}
				if c1 != 0 {
					y = ^y
				}
				if got[i] != x&y {
					t.Fatal(c0, c1)
				}
			}
		}
	}
	out := make([]uint32, 2)
	Extra_TruthNot(out, a, 6)
	for i, v := range a {
		if out[i] != ^v {
			t.Fatal("not")
		}
	}
	Extra_TruthCopy(out, a, 6)
	if Extra_TruthIsEqual(out, a, 6) != 1 || Extra_TruthIsImply(out, a, 6) != 1 {
		t.Fatal("copy/equality")
	}
	out[0] |= 0x80000000
	if Extra_TruthIsImply(out, a, 6) != 0 {
		t.Fatal("implication")
	}
}
