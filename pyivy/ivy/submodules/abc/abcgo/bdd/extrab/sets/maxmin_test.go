package sets

import (
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func nativeMinimal(x uint64, n int) uint64 {
	result := x
	for a := 0; a < 1<<uint(n); a++ {
		if x>>uint(a)&1 == 0 {
			continue
		}
		for b := 0; b < 1<<uint(n); b++ {
			if a != b && x>>uint(b)&1 != 0 && b&^a == 0 {
				result &= ^(uint64(1) << uint(a))
				break
			}
		}
	}
	return result
}
func nativePairwise(x, y uint64, n int, intersection bool) uint64 {
	var result uint64
	for a := 0; a < 1<<uint(n); a++ {
		if x>>uint(a)&1 == 0 {
			continue
		}
		for b := 0; b < 1<<uint(n); b++ {
			if y>>uint(b)&1 == 0 {
				continue
			}
			member := a | b
			if intersection {
				member = a & b
			}
			result |= 1 << uint(member)
		}
	}
	return result
}
func TestMaxMinAllThreeElementFamilyPairs(t *testing.T) {
	var hash uint64 = 14695981039346656037
	byteHash := func(b byte) { hash ^= uint64(b); hash *= 1099511628211 }
	ops := []func(*DdManager, *DdNode, *DdNode) *DdNode{Extra_zddDotProduct, Extra_zddCrossProduct, Extra_zddMaxUnion, Extra_zddMinUnion, Extra_zddMaxDotProduct}
	for _, order := range [][]int{{0, 1, 2}, {2, 0, 1}} {
		dd := cudd.NewBDDManager(0)
		dd.ZddSize = 3
		dd.InvPermZ = append([]int(nil), order...)
		dd.PermZ = make([]int, 3)
		for level, index := range order {
			dd.PermZ[index] = level
		}
		if cudd.InitializeZDDUniverse(dd) != 1 {
			t.Fatal("universe")
		}
		families := make([]*DdNode, 256)
		for x := range families {
			families[x] = familyFromBits(dd, uint64(x), 3)
		}
		for x, f := range families {
			minimal := familyBits(dd, Extra_zddMinimal(dd, f))
			if minimal != nativeMinimal(uint64(x), 3) {
				t.Fatal("minimal", x, minimal)
			}
			if order[0] == 0 {
				byteHash(byte(minimal))
			}
			for y, g := range families {
				dot := nativePairwise(uint64(x), uint64(y), 3, false)
				want := []uint64{dot, nativePairwise(uint64(x), uint64(y), 3, true), nativeMaximal(uint64(x|y), 3), nativeMinimal(uint64(x|y), 3), nativeMaximal(dot, 3)}
				for i, op := range ops {
					result := op(dd, f, g)
					if result == nil {
						t.Fatal("nil", x, y, i)
					}
					got := familyBits(dd, result)
					if got != want[i] {
						t.Fatalf("order%v x%x y%x op%d got%x want%x", order, x, y, i, got, want[i])
					}
					if order[0] == 0 {
						byteHash(byte(got))
					}
				}
			}
		}
	}
	if hash != 0xd5ec309d9c0db5a5 {
		t.Fatalf("original MaxMin antichain fixture: %016x", hash)
	}
}
func TestPublicAntichainNormalization(t *testing.T) {
	dd := cudd.NewBDDManager(0)
	a := cudd.Cudd_zddChange(dd, dd.One, 0)
	ab := cudd.Cudd_zddChange(dd, a, 1)
	family := cudd.Cudd_zddUnion(dd, a, ab)
	if Extra_zddMaxUnion(dd, family, dd.Zero) != ab || Extra_zddMaxUnion(dd, family, family) != ab || Extra_zddMaxDotProduct(dd, family, dd.One) != ab {
		t.Fatal("nonmaximal input shortcut")
	}
	if Extra_zddMinUnion(dd, family, dd.Zero) != a || Extra_zddMinUnion(dd, family, family) != a {
		t.Fatal("nonminimal input shortcut")
	}
}
func TestMaxMinFourElementFamilies(t *testing.T) {
	dd := cudd.NewBDDManager(0)
	seed := uint64(27)
	for sample := 0; sample < 1024; sample++ {
		seed ^= seed << 13
		seed ^= seed >> 7
		seed ^= seed << 17
		x, y := seed&65535, seed>>16&65535
		f, g := familyFromBits(dd, x, 4), familyFromBits(dd, y, 4)
		want := nativeMaximal(nativePairwise(x, y, 4, false), 4)
		if familyBits(dd, Extra_zddMaxDotProduct(dd, f, g)) != want || familyBits(dd, Extra_zddMinUnion(dd, f, g)) != nativeMinimal(x|y, 4) {
			t.Fatal("four-element antichain operation", sample)
		}
	}
}
func TestMaxMinInvalidOwnership(t *testing.T) {
	dd := cudd.NewBDDManager(1)
	other := cudd.NewBDDManager(0)
	foreign := cudd.Cudd_zddChange(other, other.One, 0)
	for _, f := range []*DdNode{nil, foreign, dd.Vars[0], dd.LogicZero, cudd.Cudd_addConst(dd, 2)} {
		if Extra_zddMinimal(dd, f) != nil {
			t.Fatal("invalid minimal family")
		}
		for _, op := range []func(*DdManager, *DdNode, *DdNode) *DdNode{Extra_zddMaxUnion, Extra_zddMinUnion, Extra_zddDotProduct, Extra_zddCrossProduct, Extra_zddMaxDotProduct} {
			if op(dd, f, dd.Zero) != nil || op(dd, dd.One, f) != nil {
				t.Fatal("invalid binary family")
			}
		}
	}
}
