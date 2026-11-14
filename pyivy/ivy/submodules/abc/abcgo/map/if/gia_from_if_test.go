package if_

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
	"testing"
)

func TestGiaFromIfDsd(t *testing.T) {
	p, cut := ifDelayTestManager(t)
	defer If_ManStop(p)
	p.pPars.PLutStruct = ""
	p.pPars.FDsdBalance = 1
	p.pIfDsdMan = If_DsdManAlloc(3, 4)
	defer If_DsdManFree(p.pIfDsdMan, 0)
	truth := dau.Dau_DsdToTruth("<abc>", 3)
	pins := make([]byte, 3)
	lit := If_DsdManCompute(p.pIfDsdMan, truth, 3, pins, "")
	p.vTtDsds[3] = vec.Vec_IntAllocArrayCopy([]int32{0, lit}, 2)
	p.vTtPerms[3] = vec.Vec_StrAlloc(12)
	vec.Vec_StrFill(p.vTtPerms[3], 12, 0)
	for i, pin := range pins {
		vec.Vec_StrWriteEntry(p.vTtPerms[3], 6+i, pin)
	}
	a, b := If_ManCi(p, 0), If_ManCi(p, 1)
	root := If_ManCreateAnd(p, a, b)
	root.CutBest = *cut
	If_ManCreateCo(p, If_Not(root))
	result := Gia_ManFromIfAig(p)
	defer Gia_ManStop(result)
	if Gia_ManCiNum(result) != 3 || Gia_ManCoNum(result) != 1 {
		t.Fatal("interface was not preserved")
	}
	for mask := 0; mask < 8; mask++ {
		vals := make([]bool, Gia_ManObjNum(result))
		for i := int32(0); i < Gia_ManCiNum(result); i++ {
			vals[gia.Gia_ObjId(result, Gia_ManCi(result, i))] = mask>>uint(i)&1 != 0
		}
		for id := int32(1); id < Gia_ManObjNum(result); id++ {
			node := Gia_ManObj(result, id)
			if gia.Gia_ObjIsAnd(node) == 0 {
				continue
			}
			x := vals[gia.Gia_ObjFaninId0(node, id)] != (gia.Gia_ObjFaninC0(node) != 0)
			y := vals[gia.Gia_ObjFaninId1(node, id)] != (gia.Gia_ObjFaninC1(node) != 0)
			vals[id] = x && y
		}
		co := Gia_ManCo(result, 0)
		got := vals[gia.Gia_ObjFaninId0p(result, co)] != (gia.Gia_ObjFaninC0(co) != 0)
		want := truth[0]>>uint(mask)&1 == 0
		if got != want {
			t.Fatalf("assignment %d: got %v, want %v", mask, got, want)
		}
	}
}

func TestGiaFromIfRecordedLibraryCallback(t *testing.T) {
	pars := &If_Par_t_{NLutSize: 4, NCutsMax: 4, FUserRecLib: 1}
	calls := 0
	pars.SetRecordedGiaBuilder(func(dst *Gia_Man_t, _ *If_Man_t, cut *If_Cut_t, leaves *Vec_Int_t, hash int32) int32 {
		calls++
		if cut.nLeaves != 2 || Vec_IntSize(leaves) != 2 || hash != 0 {
			t.Fatal("incorrect recorded-library arguments")
		}
		return gia.Gia_ManAppendAnd(dst, Vec_IntEntry(leaves, 0), Vec_IntEntry(leaves, 1))
	})
	p := If_ManStart(pars)
	defer If_ManStop(p)
	a, b := If_ManCreateCi(p), If_ManCreateCi(p)
	root := If_ManCreateAnd(p, a, b)
	root.CutBest = If_Cut_t{nLeaves: 2, pLeaves: []int32{a.Id, b.Id}}
	If_ManCreateCo(p, root)
	result := Gia_ManFromIfAig(p)
	defer Gia_ManStop(result)
	if calls != 1 || gia.Gia_ManAndNum(result) != 1 {
		t.Fatalf("calls %d, AND nodes %d", calls, gia.Gia_ManAndNum(result))
	}
}
