package if_

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func ifTestCut(limit uint32, leaves ...int32) *If_Cut_t {
	c := &If_Cut_t{nLimit: limit, nLeaves: uint32(len(leaves)), pLeaves: append([]int32(nil), leaves...), uMaskFunc: 0}
	c.uSign = If_ObjCutSignCompute(c)
	return c
}
func TestCutNativeMergeOracle(t *testing.T) {
	r := rand.New(rand.NewSource(332))
	for trial := 0; trial < 2000; trial++ {
		limit := 1 + r.Intn(16)
		setA, setB := map[int32]bool{}, map[int32]bool{}
		for i := 0; i < r.Intn(limit+1); i++ {
			setA[int32(r.Intn(32))] = true
		}
		for i := 0; i < r.Intn(limit+1); i++ {
			setB[int32(r.Intn(32))] = true
		}
		var a, b, want []int32
		for leaf := range setA {
			a = append(a, leaf)
		}
		for leaf := range setB {
			b = append(b, leaf)
		}
		sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
		sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })
		union := map[int32]bool{}
		for _, leaf := range append(append([]int32(nil), a...), b...) {
			union[leaf] = true
		}
		for leaf := range union {
			want = append(want, leaf)
		}
		sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })
		p := &If_Man_t{}
		ca, cb, out := ifTestCut(uint32(limit), a...), ifTestCut(uint32(limit), b...), ifTestCut(uint32(limit), 99)
		got := If_CutMergeOrdered_(p, ca, cb, out)
		if (got != 0) != (len(want) <= limit) {
			t.Fatal("wrong feasibility")
		}
		if got != 0 {
			if !reflect.DeepEqual(ifLeaves(out), want) && !(len(want) == 0 && out.nLeaves == 0) {
				t.Fatal("union mismatch", a, b, want, out.pLeaves)
			}
			for i, id := range a {
				if out.pLeaves[p.pPerm[0][i]] != id {
					t.Fatal("first permutation")
				}
			}
			for i, id := range b {
				if out.pLeaves[p.pPerm[1][i]] != id {
					t.Fatal("second permutation")
				}
			}
		} else if !reflect.DeepEqual(out.pLeaves, []int32{99}) {
			t.Fatal("failed merge changed destination")
		}
	}
}
func TestCutNativeFilterAndRanking(t *testing.T) {
	p := &If_Man_t{pPars: &If_Par_t{}, fEpsilon: 0.001}
	a, b, c := ifTestCut(6, 1, 2, 3), ifTestCut(6, 7, 8), ifTestCut(6, 1, 2)
	set := &If_Set_t{nCuts: 2, nCutsMax: 3, ppCuts: []*If_Cut_t{a, b, c, nil}}
	if If_CutFilter(set, c, 0) != 0 || set.nCuts != 1 || set.ppCuts[0] != b || set.ppCuts[1] != c {
		t.Fatal("dominance filtering")
	}
	a.Delay, a.Area, b.Delay, b.Area = 1, 10, 2, 1
	for mode, want := range []int32{-1, 1, -1} {
		p.SortMode = int32(mode)
		if If_ManSortCompare(p, a, b) != want {
			t.Fatal("sort mode", mode)
		}
	}
	set = &If_Set_t{nCuts: 1, nCutsMax: 2, ppCuts: []*If_Cut_t{b, a, nil}}
	p.SortMode = 0
	If_CutSort(p, set, a)
	if set.nCuts != 2 || set.ppCuts[0] != a {
		t.Fatal("priority insertion")
	}
}
func TestCutNativeReferenceAccountingRollback(t *testing.T) {
	p := &If_Man_t{pPars: &If_Par_t{}, fEpsilon: 0.001, vObjs: vec.Vec_PtrAlloc(4)}
	objects := []*If_Obj_t{{Type: IF_CONST1, Id: 0}, {Type: IF_CI, Id: 1}, {Type: IF_CI, Id: 2}, {Type: IF_AND, Id: 3}}
	for _, o := range objects {
		vec.Vec_PtrPush(p.vObjs, o)
	}
	objects[3].CutBest = *ifTestCut(6, 1, 2)
	c := ifTestCut(6, 2, 3)
	if got := If_CutAreaDerefed(p, c); got != 2 {
		t.Fatal("cone area", got)
	}
	for _, o := range objects {
		if o.nRefs != 0 {
			t.Fatal("evaluation mutated refs")
		}
	}
	If_CutAreaRef(p, c)
	if objects[1].nRefs != 1 || objects[2].nRefs != 2 || objects[3].nRefs != 1 {
		t.Fatal("wrong shared references")
	}
	If_CutAreaDeref(p, c)
	objects[2].nRefs = 1
	func() {
		defer func() {
			if recover() == nil {
				t.Error("underflow accepted")
			}
		}()
		If_CutAreaDeref(p, c)
	}()
	if objects[2].nRefs != 1 || objects[3].nRefs != 0 {
		t.Fatal("failed traversal leaked ref changes")
	}
}
func TestCutNativeFilterRequiresCandidateSlot(t *testing.T) {
	c := ifTestCut(6, 1)
	set := &If_Set_t{nCuts: 1, nCutsMax: 2, ppCuts: []*If_Cut_t{ifTestCut(6, 1, 2), ifTestCut(6, 3)}}
	defer func() {
		if recover() == nil {
			t.Error("filter accepted candidate outside its scratch slot")
		}
	}()
	If_CutFilter(set, c, 0)
}
