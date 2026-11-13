package cec

// Backward greedy justification from cecSimBack.c. This is a diagnostic
// simulation heuristic, not a complete SAT solver: it never backtracks when
// neither zero-valued AND input has been chosen yet.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

func Cec_ManCheckSat_rec(p *Gia_Man_t, o *Gia_Obj_t, value int32) int32 {
	if gia.Gia_ObjIsTravIdCurrent(p, o) != 0 {
		if value == gia.Gia_ObjMark0(o) {
			return 1
		}
		return 0
	}
	gia.Gia_ObjSetTravIdCurrent(p, o)
	gia.Gia_ObjSetMark0(o, value)
	if gia.Gia_ObjIsAnd(o) == 0 {
		return 1
	}
	cecRequireAndGraph(p, o)
	a, b := gia.Gia_ObjFanin0(o), gia.Gia_ObjFanin1(o)
	ca, cb := gia.Gia_ObjFaninC0(o), gia.Gia_ObjFaninC1(o)
	if value != 0 {
		if Cec_ManCheckSat_rec(p, a, ca^1) == 0 {
			return 0
		}
		return Cec_ManCheckSat_rec(p, b, cb^1)
	}
	if gia.Gia_ObjIsTravIdCurrent(p, a) != 0 {
		if ca == gia.Gia_ObjMark0(a) {
			return 1
		}
		return Cec_ManCheckSat_rec(p, b, cb)
	}
	if gia.Gia_ObjIsTravIdCurrent(p, b) != 0 {
		if cb == gia.Gia_ObjMark0(b) {
			return 1
		}
		return Cec_ManCheckSat_rec(p, a, ca)
	}
	return Cec_ManCheckSat_rec(p, a, ca)
}
func Cec_ManCheckSat2_rec(p *Gia_Man_t, o *Gia_Obj_t, value uint64, values *Vec_Wrd_t) uint64 {
	if gia.Gia_ObjIsTravIdCurrent(p, o) != 0 {
		if value == uint64(gia.Gia_ObjMark0(o)) {
			return 1
		}
		return 0
	}
	gia.Gia_ObjSetTravIdCurrent(p, o)
	gia.Gia_ObjSetMark0(o, int32(value))
	if gia.Gia_ObjIsAnd(o) == 0 {
		return 1
	}
	cecRequireAndGraph(p, o)
	a, b := gia.Gia_ObjFanin0(o), gia.Gia_ObjFanin1(o)
	ca, cb := gia.Gia_ObjFaninC0(o), gia.Gia_ObjFaninC1(o)
	if value != 0 {
		if Cec_ManCheckSat2_rec(p, a, uint64(ca^1), values) == 0 {
			return 0
		}
		return Cec_ManCheckSat2_rec(p, b, uint64(cb^1), values)
	}
	if gia.Gia_ObjIsTravIdCurrent(p, a) != 0 {
		if ca == gia.Gia_ObjMark0(a) {
			return 1
		}
		return Cec_ManCheckSat2_rec(p, b, uint64(cb), values)
	}
	if gia.Gia_ObjIsTravIdCurrent(p, b) != 0 {
		if cb == gia.Gia_ObjMark0(b) {
			return 1
		}
		return Cec_ManCheckSat2_rec(p, a, uint64(ca), values)
	}
	return Cec_ManCheckSat2_rec(p, a, uint64(ca), values)
}

// The original wrapper is unfinished and always returns zero without
// modifying the graph or the caller's value vector.
func Cec_ManCheckSat2(p *Gia_Man_t, o *Gia_Obj_t, value int32, values *Vec_Wrd_t) uint64 {
	return 0
}
func Cec_ObjSatVerify(p *Gia_Man_t, root *Gia_Obj_t, value int32) int32 {
	fmt.Printf("Obj = %4d Value = %d ", gia.Gia_ObjId(p, root), value)
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		o := gia.Gia_ManCi(p, i)
		v := int32(3)
		if gia.Gia_ObjIsTravIdCurrent(p, o) == 0 {
			fmt.Print("x")
		} else if gia.Gia_ObjMark0(o) != 0 {
			v = 2
			fmt.Print("1")
		} else {
			v = 1
			fmt.Print("0")
		}
		gia.Gia_ObjSetMark0(o, v&1)
		gia.Gia_ObjSetMark1(o, v>>1)
	}
	zero := gia.Gia_ManConst0(p)
	gia.Gia_ObjSetMark0(zero, 1)
	gia.Gia_ObjSetMark1(zero, 0)
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) == 0 {
			continue
		}
		cecRequireAndGraph(p, o)
		a, b := gia.Gia_ObjFanin0(o), gia.Gia_ObjFanin1(o)
		va := gia.Gia_ObjMark1(a)<<1 | gia.Gia_ObjMark0(a)
		vb := gia.Gia_ObjMark1(b)<<1 | gia.Gia_ObjMark0(b)
		v := cecXAnd(cecXNot(va, gia.Gia_ObjFaninC0(o)), cecXNot(vb, gia.Gia_ObjFaninC1(o)))
		gia.Gia_ObjSetMark0(o, v&1)
		gia.Gia_ObjSetMark1(o, v>>1)
	}
	want := int32(1)
	if value != 0 {
		want = 2
	}
	got := gia.Gia_ObjMark1(root)<<1 | gia.Gia_ObjMark0(root)
	if got == want {
		fmt.Println(" Verification successful.")
		return 1
	}
	fmt.Println(" Verification failed.")
	return 0
}
func cecSimSetPhases(p *Gia_Man_t) { gia.Gia_ManSetPhase(p) }
func Cec_ManSimBack(p *Gia_Man_t) {
	start := util.Abc_Clock()
	cecSimSetPhases(p)
	count, total := 0, 0
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) == 0 {
			continue
		}
		total++
		gia.Gia_ManIncrementTravId(p)
		if Cec_ManCheckSat_rec(p, o, gia.Gia_ObjPhase(o)^1) != 0 {
			count++
		}
	}
	fmt.Printf("Obj = %6d. SAT = %6d. Time = %.2f sec\n", total, count, float64(util.Abc_Clock()-start)/1e9)
}
