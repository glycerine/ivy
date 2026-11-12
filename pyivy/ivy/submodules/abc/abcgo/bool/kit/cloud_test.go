package kit

import (
	"math/rand"
	"testing"
)

func cloudFromTruth(dd *CloudManager, truth uint32, nVars int32) *CloudNode {
	var rec func(int32, uint32) *CloudNode
	rec = func(v int32, index uint32) *CloudNode {
		if v == nVars {
			if truth&(1<<index) != 0 {
				return dd.one
			}
			return dd.zero
		}
		low := rec(v+1, index)
		high := rec(v+1, index|1<<uint(v))
		if low == nil || high == nil {
			return nil
		}
		return Cloud_MakeNode(dd, uint32(v), high, low)
	}
	return rec(0, 0)
}
func TestCloudBooleanOperations(t *testing.T) {
	dd := Cloud_Init(4, 14)
	rng := rand.New(rand.NewSource(8))
	for test := 0; test < 500; test++ {
		a, b := rng.Uint32()&0xffff, rng.Uint32()&0xffff
		f, g := cloudFromTruth(dd, a, 4), cloudFromTruth(dd, b, 4)
		if f == nil || g == nil {
			t.Fatal("unexpected node limit")
		}
		ops := []*CloudNode{Cloud_bddAnd(dd, f, g), Cloud_bddOr(dd, f, g), Cloud_bddXor(dd, f, g), Cloud_Not(f)}
		truths := []uint32{a & b, a | b, a ^ b, ^a}
		values := make([]bool, 4)
		for i := 0; i < 16; i++ {
			for v := range values {
				values[v] = i&(1<<uint(v)) != 0
			}
			for k, n := range ops {
				if n == nil || Cloud_Evaluate(dd, n, values) != (truths[k]&(1<<uint(i)) != 0) {
					t.Fatal(test, k, i)
				}
			}
		}
		if cloudFromTruth(dd, a, 4) != f || Cloud_Not(Cloud_Not(f)) != f {
			t.Fatal("noncanonical result")
		}
	}
}
func TestCloudSupportSharingAndCube(t *testing.T) {
	dd := Cloud_Init(4, 10)
	f := Cloud_bddXor(dd, dd.vars[0], dd.vars[2])
	if Cloud_SupportSize(dd, f) != 2 {
		t.Fatal("support")
	}
	support := Cloud_Support(dd, f)
	for i := 0; i < 16; i++ {
		values := []bool{i&1 != 0, i&2 != 0, i&4 != 0, i&8 != 0}
		if Cloud_Evaluate(dd, support, values) != (values[0] && values[2]) {
			t.Fatal("support cube")
		}
	}
	cube := Cloud_GetOneCube(dd, f)
	if cube == nil || Cloud_bddAnd(dd, cube, Cloud_Not(f)) != dd.zero {
		t.Fatal("cube not contained")
	}
	if Cloud_DagCollect(dd, f) != Cloud_DagSize(dd, f) || len(dd.ppNodes) != int(Cloud_DagSize(dd, f)) {
		t.Fatal("DAG collection")
	}
	if Cloud_SharingSize(dd, []*CloudNode{f, Cloud_Not(f)}, 2) != Cloud_DagSize(dd, f) {
		t.Fatal("complement sharing")
	}
}
func TestCloudRestartAndLimit(t *testing.T) {
	dd := Cloud_Init(2, 3)
	limited := false
	for f := uint32(0); f < 16; f++ {
		if cloudFromTruth(dd, f, 2) == nil {
			limited = true
			break
		}
	}
	if !limited {
		t.Fatal("node limit not enforced")
	}
	old := dd.vars[0]
	Cloud_Restart(dd)
	if len(dd.tUnique) != 3 {
		t.Fatal("restart did not reset nodes")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("expired node accepted")
		}
	}()
	Cloud_bddAnd(dd, old, dd.one)
}
