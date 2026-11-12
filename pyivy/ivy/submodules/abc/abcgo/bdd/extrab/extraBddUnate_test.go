package extrab

import (
	"reflect"
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func unateTruth(dd *DdManager, table uint64, n int) *DdNode {
	var build func(int, uint64) *DdNode
	build = func(level int, table uint64) *DdNode {
		if level == n {
			if table&1 != 0 {
				return dd.One
			}
			return dd.LogicZero
		}
		width := uint(1) << uint(n-level-1)
		mask := (uint64(1) << width) - 1
		return cudd.Cudd_bddIte(dd, dd.Vars[level], build(level+1, table>>width), build(level+1, table&mask))
	}
	return build(0, table)
}
func TestUnatenessExhaustive(t *testing.T) {
	for table := uint64(0); table < 256; table++ {
		dd := cudd.NewBDDManager(3)
		f := unateTruth(dd, table, 3)
		slow, fast := Extra_UnateComputeSlow(dd, f), Extra_UnateComputeFast(dd, f)
		if !reflect.DeepEqual(slow, fast) {
			t.Fatalf("table %x slow=%+v fast=%+v", table, slow, fast)
		}
		for v := 0; v < 3; v++ {
			positive, negative := true, true
			for a := 0; a < 8; a++ {
				mask := 1 << uint(2-v)
				if a&mask != 0 {
					continue
				}
				lo := table >> uint(a) & 1
				hi := table >> uint(a|mask) & 1
				if lo > hi {
					positive = false
				}
				if hi > lo {
					negative = false
				}
			}
			want := 0
			if positive {
				want = 1
			} else if negative {
				want = -1
			}
			if Extra_bddCheckUnateNaive(dd, f, v) != want {
				t.Fatal("native polarity", table, v)
			}
		}
	}
}
func TestUnatenessSkippedVariablesAndReordering(t *testing.T) {
	dd := cudd.NewBDDManager(5)
	f := cudd.Cudd_bddAnd(dd, dd.Vars[4], cudd.Cudd_Not(dd.Vars[1]))
	cudd.Cudd_Ref(f)
	cudd.Cudd_ShuffleHeap(dd, []int{3, 4, 2, 1, 0})
	info := Extra_UnateComputeFast(dd, f)
	if info == nil || info.NVars != 2 || info.NUnate != 2 || info.Vars[0].IVar != 4 || !info.Vars[0].Pos || info.Vars[1].IVar != 1 || !info.Vars[1].Neg {
		t.Fatalf("reordered support: %+v", info)
	}
	if !reflect.DeepEqual(info, Extra_UnateComputeSlow(dd, f)) {
		t.Fatal("order mismatch")
	}
}
func TestUnatenessExtraSupportSingletonConvention(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	vars := cudd.Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[2])
	family := Extra_zddGetSingletonsBoth(dd, vars)
	if family == nil {
		t.Fatal("singletons")
	}
	remaining := family
	for _, index := range []int{0, 1, 4, 5} {
		single := unateSingleton(dd, index)
		if cudd.Cudd_zddIntersect(dd, family, single) != single {
			t.Fatal("missing singleton", index)
		}
		remaining = cudd.Cudd_zddDiff(dd, remaining, single)
	}
	if remaining != dd.One {
		t.Fatal("lost original empty-set base case")
	}
	if Extra_UnateInfoCreateFromZdd(dd, family, vars) != nil {
		t.Fatal("accepted malformed singleton-only representation")
	}
	if Extra_zddUnateInfoCompute(dd, dd.Vars[1], vars) != nil {
		t.Fatal("missing support accepted")
	}
	if Extra_zddGetSingletonsBoth(dd, cudd.Cudd_Not(dd.Vars[0])) != nil {
		t.Fatal("negative support cube accepted")
	}
}

func TestOriginalCUnatenessFixture(t *testing.T) {
	var hash uint64 = 14695981039346656037
	number := func(v uint64) {
		for i := 0; i < 8; i++ {
			hash ^= v >> uint(i*8) & 255
			hash *= 1099511628211
		}
	}
	for table := uint64(0); table < 256; table++ {
		dd := cudd.NewBDDManager(3)
		f := unateTruth(dd, table, 3)
		fast, slow := Extra_UnateComputeFast(dd, f), Extra_UnateComputeSlow(dd, f)
		for i := 0; i < 3; i++ {
			number(uint64(Extra_bddCheckUnateNaive(dd, f, i)))
		}
		number(uint64(fast.NVars))
		number(uint64(fast.NVarsMax))
		number(uint64(fast.NUnate))
		for _, v := range fast.Vars {
			number(uint64(v.IVar))
			if v.Pos {
				number(1)
			} else {
				number(0)
			}
			if v.Neg {
				number(1)
			} else {
				number(0)
			}
		}
		number(uint64(slow.NUnate))
	}
	if hash != 0x94e5759806dcde0c {
		t.Fatalf("original C unateness checksum: %016x", hash)
	}
}
