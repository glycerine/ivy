package gia

import (
	"sort"
	"testing"
)

func TestSupergateSimplificationExhaustive(t *testing.T) {
	// Every sorted multiset of up to five literals over two variables, constants
	// included, compared against direct Boolean AND/XOR evaluation.
	for size := 0; size <= 5; size++ {
		total := 1
		for i := 0; i < size; i++ {
			total *= 6
		}
		for code := 0; code < total; code++ {
			literals := make([]int32, size)
			n := code
			for i := range literals {
				literals[i] = int32(n % 6)
				n /= 6
			}
			sort.Slice(literals, func(i, j int) bool { return literals[i] < literals[j] })
			for mode := 0; mode < 2; mode++ {
				v := Vec_IntAlloc(size)
				for _, lit := range literals {
					Vec_IntPush(v, lit)
				}
				if mode == 0 {
					Gia_ManSimplifyAnd(v)
				} else {
					Gia_ManSimplifyXor(v)
				}
				for mask := 0; mask < 4; mask++ {
					eval := func(items []int32) bool {
						value := mode == 0
						for _, lit := range items {
							bit := false
							if lit>>1 != 0 {
								bit = mask&(1<<uint((lit>>1)-1)) != 0
							}
							if lit&1 != 0 {
								bit = !bit
							}
							if mode == 0 {
								value = value && bit
							} else {
								value = value != bit
							}
						}
						return value
					}
					if eval(literals) != eval(Vec_IntArray(v)) {
						t.Fatal("incorrect supergate", mode, literals, Vec_IntArray(v), mask)
					}
				}
			}
		}
	}
}

func TestSupergateCollectionSharedBoundaryAndComplement(t *testing.T) {
	p := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	shared := Gia_ManAppendAnd(p, a, b)
	root := Gia_ManAppendAnd(p, shared, c^1)
	Gia_ManAppendCo(p, root)
	Gia_ManAppendCo(p, shared)
	Gia_ManCreateRefs(p)
	Gia_ManSuperCollect(p, Gia_ManObj(p, root>>1), 1)
	want := []int32{c ^ 1, shared}
	got := Vec_IntArray(p.vSuper)
	if len(got) != len(want) {
		t.Fatal("shared boundary", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatal("shared boundary", got)
		}
	}
	// Non-strict source policy still stops at a two-use node with a one-use
	// fanin. Add another use for each leaf to make the supergate expandable.
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, b)
	p.Refs = nil
	Gia_ManCreateRefs(p)
	Gia_ManSuperCollect(p, Gia_ManObj(p, root>>1), 0)
	got = Vec_IntArray(p.vSuper)
	if len(got) != 3 || got[0] != a || got[1] != b || got[2] != c^1 {
		t.Fatal("expanded supergate", got)
	}
}

func TestXorSupergateCollectionCancelsRepeatedLeaves(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	pair := Gia_ManAppendXorReal(p, a, b)
	root := Gia_ManAppendXorReal(p, pair, a)
	Gia_ManAppendCo(p, root)
	Gia_ManCreateRefs(p)
	Gia_ManSuperCollect(p, Gia_ManObj(p, root>>1), 0)
	if got := Vec_IntArray(p.vSuper); len(got) != 1 || got[0] != b {
		t.Fatal("XOR cancellation", got)
	}
}
