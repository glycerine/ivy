package extrab

import (
	"bytes"
	"hash/fnv"
	"math/rand"
	"strings"
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestMiscAllBooleanTransforms(t *testing.T) {
	dd := cudd.NewBDDManager(6)
	for table := uint64(0); table < 256; table++ {
		f := autoTruth(dd, table, 3)
		cube := Extra_bddGetOneCube(dd, f)
		if !cudd.Cudd_bddLeq(dd, cube, f) || (table != 0 && cube == dd.LogicZero) {
			t.Fatal("one cube", table)
		}
		for mask := 0; mask < 8; mask++ {
			vars := dd.One
			for v := 0; v < 3; v++ {
				if mask&(1<<uint(v)) != 0 {
					vars = cudd.Cudd_bddAnd(dd, vars, dd.Vars[v])
				}
			}
			changed := Extra_bddChangePolarity(dd, f, vars)
			for a := 0; a < 8; a++ {
				mapped := a
				for v := 0; v < 3; v++ {
					if mask&(1<<uint(v)) != 0 {
						mapped ^= 1 << uint(2-v)
					}
				}
				values := make([]bool, 6)
				for v := 0; v < 3; v++ {
					values[v] = a&(1<<uint(2-v)) != 0
				}
				got, err := dd.Evaluate(changed, values)
				if err != nil || got != (table>>uint(mapped)&1 != 0) {
					t.Fatal("polarity", table, mask, a)
				}
			}
		}
		moved := Extra_bddMove(dd, f, 2)
		if moved == nil {
			t.Fatal("valid move")
		}
		for a := 0; a < 64; a++ {
			values := make([]bool, 6)
			for v := range values {
				values[v] = a&(1<<uint(v)) != 0
			}
			mapped := 0
			for v := 0; v < 3; v++ {
				if values[v+2] {
					mapped |= 1 << uint(2-v)
				}
			}
			got, err := dd.Evaluate(moved, values)
			if err != nil || got != (table>>uint(mapped)&1 != 0) {
				t.Fatal("move", table, a)
			}
		}
	}
}
func TestMiscCrossManagerFusedAnd(t *testing.T) {
	a, b := cudd.NewBDDManager(3), cudd.NewBDDManager(3)
	cudd.Cudd_ShuffleHeap(a, []int{2, 0, 1})
	cudd.Cudd_ShuffleHeap(b, []int{1, 2, 0})
	left, right := make([]*DdNode, 256), make([]*DdNode, 256)
	for table := 0; table < 256; table++ {
		left[table] = autoTruth(a, uint64(table), 3)
		right[table] = autoTruth(b, uint64(table), 3)
	}
	mapping := []int{2, 0, 1}
	for x, f := range left {
		for y, g := range right {
			want := cudd.Cudd_bddAnd(a, f, Extra_TransferPermute(b, a, g, mapping))
			if Extra_bddAndPermute(a, f, b, g, mapping) != want {
				t.Fatal("cross-manager conjunction", x, y)
			}
		}
	}
	before := len(a.Vars)
	if Extra_bddAndPermute(a, a.One, b, cudd.Cudd_bddAnd(b, b.Vars[0], b.Vars[1]), []int{10, -1, 2}) != nil || len(a.Vars) != before {
		t.Fatal("bad mapping changed destination")
	}
}
func TestMiscSupportsTuplesAndArrays(t *testing.T) {
	dd := cudd.NewBDDManager(4)
	cube := cudd.Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[2])
	if Extra_bddSuppSize(dd, cube) != 2 || !Extra_bddSuppContainVar(dd, cube, dd.Vars[2]) || Extra_bddSuppOverlapping(dd, cube, dd.Vars[1]) || Extra_bddSuppDifferentVars(dd, cube, dd.Vars[1], 9) != 3 {
		t.Fatal("support helpers")
	}
	large, small := dd.One, dd.One
	if !Extra_bddSuppCheckContainment(dd, cube, dd.Vars[0], &large, &small) || large != cube || small != dd.Vars[0] {
		t.Fatal("containment")
	}
	if Extra_bddSuppCheckContainment(dd, cube, cube, &large, &small) || large != cube || small != dd.Vars[0] {
		t.Fatal("equal supports changed outputs")
	}
	for k := 0; k <= 2; k++ {
		f := Extra_bddTuples(dd, k, cube)
		for a := 0; a < 16; a++ {
			values := []bool{a&1 != 0, a&2 != 0, a&4 != 0, a&8 != 0}
			count := 0
			if values[0] {
				count++
			}
			if values[2] {
				count++
			}
			got, err := dd.Evaluate(f, values)
			if err != nil || got != (count == k) {
				t.Fatal("tuples", k, a)
			}
		}
	}
	array := Extra_SupportArray(dd, cube, nil)
	if len(array) != 4 || array[0] != 1 || array[1] != 0 || array[2] != 1 {
		t.Fatal(array)
	}
	out := make([]*DdNode, 1)
	if !Extra_bddPermuteArray(dd, []*DdNode{cube}, out, 1, []int{1, 0, 3, 2}) || out[0].Ref == 0 {
		t.Fatal("referenced array permutation")
	}
	cudd.Cudd_RecursiveDeref(dd, out[0])
	if Extra_bddVarIsInCube(cudd.Cudd_bddAnd(dd, dd.Vars[0], cudd.Cudd_Not(dd.Vars[2])), 2) != 0 || Extra_bddVarIsInCube(cube, 1) != -1 {
		t.Fatal("cube literal query")
	}
}
func TestMiscISOPLimitsAndRepeatedCalls(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	f := cudd.Cudd_bddXor(dd, dd.Vars[0], dd.Vars[1])
	for i := 0; i < 3; i++ {
		guide := []int{9}
		if Extra_bddCountCubes(dd, []*DdNode{f}, 1, 1, 100, guide) != 2 || guide[0] != 1 {
			t.Fatal("ISOP count")
		}
	}
	if dd.ZddSize != 0 {
		t.Fatal("count materialized a ZDD cover")
	}
	if Extra_bddCountCubes(dd, []*DdNode{f}, 1, 1, 1, []int{9}) != -1 || Extra_bddCountCubes(dd, []*DdNode{dd.One}, 1, 1, 0, []int{9}) != -1 {
		t.Fatal("cube limit ignored")
	}
	guide := []int{9, 9}
	if Extra_bddCountCubes(dd, []*DdNode{nil, dd.LogicZero}, 2, -1, 100, guide) != 0 || guide[0] != 9 || guide[1] != 1 {
		t.Fatal("nil output/phase guide")
	}
}
func TestMiscRandomCombinationAndPLA(t *testing.T) {
	dd := cudd.NewBDDManager(0)
	f := extraRandomSet(dd, 3, 8, .5, rand.New(rand.NewSource(17)))
	if cudd.Cudd_zddCount(dd, f) != 8 {
		t.Fatal("random set cardinality")
	}
	if Extra_zddRandomSet(dd, 3, 9, .5) != nil || Extra_zddRandomSet(dd, 3, 2, 0) != nil || Extra_zddRandomSet(dd, 3, 2, 1) != nil {
		t.Fatal("impossible random family accepted")
	}
	coverDD := cudd.NewBDDManager(3)
	fBDD := cudd.Cudd_bddOr(coverDD, coverDD.Vars[0], coverDD.Vars[2])
	cover := Extra_zddPrimes(coverDD, fBDD)
	var output bytes.Buffer
	if err := Extra_zddWritePla(&output, coverDD, cover, 3); err != nil || !strings.Contains(output.String(), ".i 3\n.o 1\n") || !strings.Contains(output.String(), "--1 1\n") || !strings.Contains(output.String(), "1-- 1\n") {
		t.Fatal("PLA", output.String(), err)
	}
	contradiction := cudd.Cudd_zddChange(coverDD, cudd.Cudd_zddChange(coverDD, coverDD.One, 0), 1)
	output.Reset()
	if Extra_zddWritePla(&output, coverDD, contradiction, 3) == nil || output.Len() != 0 {
		t.Fatal("contradictory cover written")
	}
}
func TestOriginalCMiscPLAHash(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	hash := fnv.New64a()
	for table := uint64(0); table < 256; table++ {
		f := autoTruth(dd, table, 3)
		cover := Extra_zddPrimes(dd, f)
		var output bytes.Buffer
		if err := Extra_zddWritePla(&output, dd, cover, 3); err != nil {
			t.Fatal(err)
		}
		hash.Write(output.Bytes())
	}
	if hash.Sum64() != 0x3889814f756ac6b6 {
		t.Fatalf("original C PLA fixture: %016x", hash.Sum64())
	}
}

func TestOriginalCMiscTransformAndCountFixtures(t *testing.T) {
	hash := uint64(14695981039346656037)
	number := func(v uint64) {
		for i := 0; i < 8; i++ {
			hash ^= v >> uint(i*8) & 255
			hash *= 1099511628211
		}
	}
	dd := cudd.NewBDDManager(6)
	for table := uint64(0); table < 256; table++ {
		f := autoTruth(dd, table, 3)
		number(autoTruthBits(t, dd, Extra_bddGetOneCube(dd, f), 6))
		number(autoTruthBits(t, dd, Extra_bddMove(dd, f, 2), 6))
		number(autoTruthBits(t, dd, Extra_bddRemapUp(dd, f), 6))
		for mask := 0; mask < 8; mask++ {
			vars := dd.One
			for v := 0; v < 3; v++ {
				if mask&(1<<uint(v)) != 0 {
					vars = cudd.Cudd_bddAnd(dd, vars, dd.Vars[v])
				}
			}
			number(autoTruthBits(t, dd, Extra_bddChangePolarity(dd, f, vars), 6))
		}
	}
	if hash != 0x869b30057848fe11 {
		t.Fatalf("original C transforms: %016x", hash)
	}
	hash = 14695981039346656037
	for table := uint64(0); table < 256; table++ {
		for mode := -1; mode <= 1; mode++ {
			manager := cudd.NewBDDManager(3)
			f := autoTruth(manager, table, 3)
			guide := []int{9}
			number(uint64(Extra_bddCountCubes(manager, []*DdNode{f}, 1, mode, 100, guide)))
			number(uint64(guide[0]))
		}
	}
	if hash != 0x8e62023c841daec4 {
		t.Fatalf("original C ISOP count: %016x", hash)
	}
}
