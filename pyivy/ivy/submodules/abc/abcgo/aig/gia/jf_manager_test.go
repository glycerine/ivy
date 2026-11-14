package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestJfManagerLifecycle(t *testing.T) {
	for _, mode := range []struct{ min, dsd, cnf int32 }{{0, 0, 0}, {1, 0, 0}, {1, 1, 1}} {
		p := Gia_ManStart(8)
		a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		x := Gia_ManAppendAnd(p, a, b)
		Gia_ManAppendCo(p, x)
		pars := &Jf_Par_t{NLutSize: 4, NCutNum: 8, FCutMin: mode.min, FFuncDsd: mode.dsd, FGenCnf: mode.cnf}
		man := Jf_ManAlloc(p, pars)
		if len(p.Refs) != int(p.nObjs) || Jf_ObjRefs(man, a>>1) != 1 || (man.vTtMem != nil) != (mode.min != 0 && mode.dsd == 0) || (man.pDsd != nil) != (mode.dsd != 0) || (man.vCnfs != nil) != (mode.cnf != 0) {
			t.Fatal("manager setup")
		}
		data := []int32{1, Jf_CutSetAll(2, 0, 1), a}
		handle := man.pMem.append(data, len(data))
		vec.Vec_IntWriteEntry(&man.vCuts, int(a>>1), handle)
		if Jf_CutFunc(Jf_ObjCutBest(man, a>>1)) != 2 {
			t.Fatal("stored cut")
		}
		Jf_ManFree(man)
		if p.Refs != nil || man.vTtMem != nil || len(man.pMem.lists) != 0 {
			t.Fatal("manager cleanup")
		}
	}
}
