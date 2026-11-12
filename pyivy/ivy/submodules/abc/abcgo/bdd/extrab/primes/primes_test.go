package primes

import (
	"reflect"
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func fromTruth(dd *cudd.DdManager, table uint64, n int) *cudd.DdNode {
	var build func(int, uint64) *cudd.DdNode
	build = func(v int, table uint64) *cudd.DdNode {
		if v == n {
			if table&1 != 0 {
				return dd.One
			}
			return dd.LogicZero
		}
		width := uint(1) << uint(n-v-1)
		mask := (uint64(1) << width) - 1
		return cudd.Cudd_bddIte(dd, dd.Vars[v], build(v+1, table>>width), build(v+1, table&mask))
	}
	return build(0, table)
}
func family(dd *cudd.DdManager, f *cudd.DdNode) map[uint64]bool {
	result := map[uint64]bool{}
	var walk func(*cudd.DdNode, uint64)
	walk = func(n *cudd.DdNode, mask uint64) {
		if n == dd.Zero {
			return
		}
		if n == dd.One {
			result[mask] = true
			return
		}
		walk(n.T, mask|1<<n.Index)
		walk(n.E, mask)
	}
	walk(f, 0)
	return result
}
func implies(table uint64, n int, mask uint64) bool {
	for a := 0; a < 1<<uint(n); a++ {
		matches := true
		for v := 0; v < n; v++ {
			value := a&(1<<uint(n-1-v)) != 0
			if mask&(1<<uint(2*v)) != 0 && !value || mask&(1<<uint(2*v+1)) != 0 && value {
				matches = false
			}
		}
		if matches && table>>uint(a)&1 == 0 {
			return false
		}
	}
	return true
}
func nativePrimes(table uint64, n int) map[uint64]bool {
	result := map[uint64]bool{}
	total := 1
	for v := 0; v < n; v++ {
		total *= 3
	}
	for code := 0; code < total; code++ {
		mask := uint64(0)
		c := code
		for v := 0; v < n; v++ {
			state := c % 3
			c /= 3
			if state != 0 {
				mask |= 1 << uint(2*v+state-1)
			}
		}
		if !implies(table, n, mask) {
			continue
		}
		prime := true
		for v := 0; v < 2*n; v++ {
			if mask&(1<<uint(v)) != 0 && implies(table, n, mask^(1<<uint(v))) {
				prime = false
				break
			}
		}
		if prime {
			result[mask] = true
		}
	}
	return result
}
func TestAllThreeVariablePrimeCovers(t *testing.T) {
	var hash uint64 = 14695981039346656037
	for table := uint64(0); table < 256; table++ {
		dd := cudd.NewBDDManager(3)
		f := fromTruth(dd, table, 3)
		cover := Extra_zddPrimes(dd, f)
		got := family(dd, cover)
		want := nativePrimes(table, 3)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("table %x got%v want%v", table, got, want)
		}
		if cudd.Cudd_MakeBddFromZddCover(dd, cover) != f {
			t.Fatal("cover reconstruction", table)
		}
		for code := 0; code < 27; code++ {
			mask := uint64(0)
			c := code
			for v := 0; v < 3; v++ {
				state := c % 3
				c /= 3
				if state != 0 {
					mask |= 1 << uint(2*v+state-1)
				}
			}
			if got[mask] {
				hash ^= 1
			}
			hash *= 1099511628211
		}
	}
	if hash != 0x5394c07cdac40638 {
		t.Fatalf("original C prime fixture: %016x", hash)
	}
}
func TestPrimeCoverRandomAndIndependentOrders(t *testing.T) {
	seed := uint64(71)
	for sample := 0; sample < 128; sample++ {
		seed ^= seed << 13
		seed ^= seed >> 7
		seed ^= seed << 17
		table := seed & 65535
		dd := cudd.NewBDDManager(4)
		if sample%3 == 1 {
			dd.ZddSize = 8
			dd.InvPermZ = []int{7, 4, 1, 6, 3, 0, 5, 2}
			dd.PermZ = make([]int, 8)
			for level, index := range dd.InvPermZ {
				dd.PermZ[index] = level
			}
			if cudd.InitializeZDDUniverse(dd) != 1 {
				t.Fatal("ZDD universe initialization")
			}
		}
		f := fromTruth(dd, table, 4)
		cudd.Cudd_Ref(f)
		if sample&1 != 0 {
			cudd.Cudd_ShuffleHeap(dd, []int{2, 0, 3, 1})
		}
		if !reflect.DeepEqual(family(dd, Extra_zddPrimes(dd, f)), nativePrimes(table, 4)) {
			t.Fatal("four-variable primes", sample, table)
		}
	}
}
func TestPrimesInvalidOwnershipAndConstants(t *testing.T) {
	dd := cudd.NewBDDManager(1)
	other := cudd.NewBDDManager(1)
	if Extra_zddPrimes(dd, other.Vars[0]) != nil || Extra_zddPrimes(dd, cudd.Cudd_addIthVar(dd, 0)) != nil || Extra_zddPrimes(nil, dd.One) != nil {
		t.Fatal("invalid graph accepted")
	}
	if Extra_zddPrimes(dd, dd.One) != dd.One || Extra_zddPrimes(dd, dd.LogicZero) != dd.Zero {
		t.Fatal("constants")
	}
}
