package gia

import "fmt"

const GIA_ZER = 1
const GIA_ONE = 2
const GIA_UND = 3

func Gia_XsimNotCond(value, compl int32) int32 {
	if value == GIA_UND {
		return GIA_UND
	}
	if value == GIA_ZER+compl {
		return GIA_ZER
	}
	return GIA_ONE
}
func Gia_XsimAndCond(a, ca, b, cb int32) int32 {
	if a == GIA_ZER+ca || b == GIA_ZER+cb {
		return GIA_ZER
	}
	if a == GIA_UND || b == GIA_UND {
		return GIA_UND
	}
	return GIA_ONE
}
func Gia_ObjTerSimSetC(obj *Gia_Obj_t)       { obj.fMark0 = 0; obj.fMark1 = 0 }
func Gia_ObjTerSimSet0(obj *Gia_Obj_t)       { obj.fMark0 = 1; obj.fMark1 = 0 }
func Gia_ObjTerSimSet1(obj *Gia_Obj_t)       { obj.fMark0 = 0; obj.fMark1 = 1 }
func Gia_ObjTerSimSetX(obj *Gia_Obj_t)       { obj.fMark0 = 1; obj.fMark1 = 1 }
func Gia_ObjTerSimGetC(obj *Gia_Obj_t) int32 { return giaBool(obj.fMark0 == 0 && obj.fMark1 == 0) }
func Gia_ObjTerSimGet0(obj *Gia_Obj_t) int32 { return giaBool(obj.fMark0 != 0 && obj.fMark1 == 0) }
func Gia_ObjTerSimGet1(obj *Gia_Obj_t) int32 { return giaBool(obj.fMark0 == 0 && obj.fMark1 != 0) }
func Gia_ObjTerSimGetX(obj *Gia_Obj_t) int32 { return giaBool(obj.fMark0 != 0 && obj.fMark1 != 0) }
func giaTerFaninValue(obj *Gia_Obj_t, compl int32, one bool) int32 {
	if one != (compl != 0) {
		return Gia_ObjTerSimGet1(obj)
	}
	return Gia_ObjTerSimGet0(obj)
}
func Gia_ObjTerSimGet0Fanin0(obj *Gia_Obj_t) int32 {
	return giaTerFaninValue(Gia_ObjFanin0(obj), Gia_ObjFaninC0(obj), false)
}
func Gia_ObjTerSimGet1Fanin0(obj *Gia_Obj_t) int32 {
	return giaTerFaninValue(Gia_ObjFanin0(obj), Gia_ObjFaninC0(obj), true)
}
func Gia_ObjTerSimGet0Fanin1(obj *Gia_Obj_t) int32 {
	return giaTerFaninValue(Gia_ObjFanin1(obj), Gia_ObjFaninC1(obj), false)
}
func Gia_ObjTerSimGet1Fanin1(obj *Gia_Obj_t) int32 {
	return giaTerFaninValue(Gia_ObjFanin1(obj), Gia_ObjFaninC1(obj), true)
}
func Gia_ObjTerSimAnd(obj *Gia_Obj_t) {
	if Gia_ObjIsAnd(obj) == 0 || Gia_ObjTerSimGetC(Gia_ObjFanin0(obj)) != 0 || Gia_ObjTerSimGetC(Gia_ObjFanin1(obj)) != 0 {
		panic("gia: ternary AND requires initialized fanins")
	}
	if Gia_ObjTerSimGet0Fanin0(obj) != 0 || Gia_ObjTerSimGet0Fanin1(obj) != 0 {
		Gia_ObjTerSimSet0(obj)
	} else if Gia_ObjTerSimGet1Fanin0(obj) != 0 && Gia_ObjTerSimGet1Fanin1(obj) != 0 {
		Gia_ObjTerSimSet1(obj)
	} else {
		Gia_ObjTerSimSetX(obj)
	}
}
func Gia_ObjTerSimCo(obj *Gia_Obj_t) {
	if Gia_ObjIsCo(obj) == 0 || Gia_ObjTerSimGetC(Gia_ObjFanin0(obj)) != 0 {
		panic("gia: ternary CO requires initialized fanin")
	}
	if Gia_ObjTerSimGet0Fanin0(obj) != 0 {
		Gia_ObjTerSimSet0(obj)
	} else if Gia_ObjTerSimGet1Fanin0(obj) != 0 {
		Gia_ObjTerSimSet1(obj)
	} else {
		Gia_ObjTerSimSetX(obj)
	}
}
func Gia_ObjTerSimRo(p *Gia_Man_t, obj *Gia_Obj_t) {
	ri := Gia_ObjRoToRi(p, obj)
	if Gia_ObjTerSimGetC(ri) != 0 {
		panic("gia: ternary register input is uninitialized")
	}
	obj.fMark0 = ri.fMark0
	obj.fMark1 = ri.fMark1
}
func Gia_ObjTerSimPrint(obj *Gia_Obj_t) {
	if Gia_ObjTerSimGet0(obj) != 0 {
		fmt.Print("0")
	} else if Gia_ObjTerSimGet1(obj) != 0 {
		fmt.Print("1")
	} else if Gia_ObjTerSimGetX(obj) != 0 {
		fmt.Print("X")
	}
}
