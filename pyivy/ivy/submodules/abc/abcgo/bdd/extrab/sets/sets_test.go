package sets

import (
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func familyFromBits(dd *DdManager, bits uint64, n int) *DdNode {
	result := dd.Zero
	for set := 0; set < 1<<uint(n); set++ {
		if bits>>uint(set)&1 == 0 {
			continue
		}
		cube := dd.One
		for v := 0; v < n; v++ {
			if set&(1<<uint(v)) != 0 {
				cube = cudd.Cudd_zddChange(dd, cube, v)
			}
		}
		result = cudd.Cudd_zddUnion(dd, result, cube)
	}
	return result
}
func familyBits(dd *DdManager, f *DdNode) uint64 {
	var result uint64
	var visit func(*DdNode, int)
	visit = func(f *DdNode, set int) {
		if f == dd.Zero {
			return
		}
		if f == dd.One {
			result |= 1 << uint(set)
			return
		}
		visit(f.E, set)
		visit(f.T, set|1<<f.Index)
	}
	visit(f, 0)
	return result
}
func nativeFilter(x, y uint64, n int, sup, neg bool) uint64 {
	var result uint64
	for a := 0; a < 1<<uint(n); a++ {
		if x>>uint(a)&1 == 0 {
			continue
		}
		found := false
		for b := 0; b < 1<<uint(n); b++ {
			if y>>uint(b)&1 != 0 && (!sup && a&^b == 0 || sup && b&^a == 0) {
				found = true
				break
			}
		}
		if found != neg {
			result |= 1 << uint(a)
		}
	}
	return result
}
func nativeMaximal(x uint64, n int) uint64 {
	result := x
	for a := 0; a < 1<<uint(n); a++ {
		if x>>uint(a)&1 == 0 {
			continue
		}
		for b := 0; b < 1<<uint(n); b++ {
			if a != b && x>>uint(b)&1 != 0 && a&^b == 0 {
				result &= ^(uint64(1) << uint(a))
				break
			}
		}
	}
	return result
}
func TestAllThreeElementFamilyPairs(t *testing.T) {
	var hash uint64 = 14695981039346656037
	byteHash := func(b byte) { hash ^= uint64(b); hash *= 1099511628211 }
	operations := []func(*DdManager, *DdNode, *DdNode) *DdNode{Extra_zddSubSet, Extra_zddSupSet, Extra_zddNotSubSet, Extra_zddNotSupSet, Extra_zddMaxNotSupSet}
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
			if Extra_zddEmptyBelongs(dd, f) != (x&1 != 0) {
				t.Fatal("empty membership", x)
			}
			one := x != 0 && x&(x-1) == 0
			if Extra_zddIsOneSubset(dd, f) != one {
				t.Fatal("one subset", x)
			}
			maximal := familyBits(dd, Extra_zddMaximal(dd, f))
			if maximal != nativeMaximal(uint64(x), 3) {
				t.Fatal("maximal", x, maximal)
			}
			if order[0] == 0 {
				if x&1 != 0 {
					byteHash(1)
				} else {
					byteHash(0)
				}
				if one {
					byteHash(1)
				} else {
					byteHash(0)
				}
				byteHash(byte(maximal))
			}
			for y, g := range families {
				want := []uint64{nativeFilter(uint64(x), uint64(y), 3, false, false), nativeFilter(uint64(x), uint64(y), 3, true, false), nativeFilter(uint64(x), uint64(y), 3, false, true), nativeFilter(uint64(x), uint64(y), 3, true, true), nativeMaximal(nativeFilter(uint64(x), uint64(y), 3, true, true), 3)}
				for i, op := range operations {
					result := op(dd, f, g)
					if result == nil {
						t.Fatal("nil result", x, y, i)
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
	t.Logf("original set fixture %016x", hash)
}
func TestSetOperationsInvalidFamilyOwnership(t *testing.T) {
	dd := cudd.NewBDDManager(1)
	other := cudd.NewBDDManager(0)
	foreign := cudd.Cudd_zddChange(other, other.One, 0)
	for _, f := range []*DdNode{nil, foreign, dd.LogicZero, dd.Vars[0], cudd.Cudd_addIthVar(dd, 0)} {
		if Extra_zddSubSet(dd, f, dd.One) != nil || Extra_zddSupSet(dd, dd.One, f) != nil || Extra_zddNotSubSet(dd, f, dd.Zero) != nil || Extra_zddNotSupSet(dd, f, dd.Zero) != nil || Extra_zddMaxNotSupSet(dd, f, dd.Zero) != nil || Extra_zddMaximal(dd, f) != nil || Extra_zddEmptyBelongs(dd, f) || Extra_zddIsOneSubset(dd, f) {
			t.Fatal("invalid family accepted")
		}
	}
}
func TestSetOperationsFourElementFamilies(t *testing.T) {
	seed := uint64(0x123456789)
	dd := cudd.NewBDDManager(0)
	for sample := 0; sample < 1024; sample++ {
		seed ^= seed << 13
		seed ^= seed >> 7
		seed ^= seed << 17
		x, y := seed&65535, seed>>16&65535
		f, g := familyFromBits(dd, x, 4), familyFromBits(dd, y, 4)
		got := familyBits(dd, Extra_zddMaxNotSupSet(dd, f, g))
		if got != nativeMaximal(nativeFilter(x, y, 4, true, true), 4) {
			t.Fatal("fused maximal filter", sample)
		}
		notSub := Extra_zddNotSubSet(dd, f, g)
		sub := Extra_zddSubSet(dd, f, g)
		if cudd.Cudd_zddUnion(dd, notSub, sub) != f || cudd.Cudd_zddIntersect(dd, notSub, sub) != dd.Zero {
			t.Fatal("subset partition", sample)
		}
	}
}
