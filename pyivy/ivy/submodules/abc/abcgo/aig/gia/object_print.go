package gia

import "fmt"

func Gia_ObjPrint(p *Gia_Man_t, obj *Gia_Obj_t) {
	if obj == nil {
		fmt.Print("Object is NULL.")
		return
	}
	if Gia_IsComplement(obj) != 0 {
		fmt.Print("Compl ")
		obj = Gia_Regular(obj)
	}
	fmt.Printf("Obj %4d : ", obj.id)
	mark := func(c int32) string {
		if c != 0 {
			return "'"
		}
		return " "
	}
	input := func(label string, o *Gia_Obj_t) {
		fmt.Printf("%s( %4d%s )", label, Gia_ObjFaninId0p(p, o), mark(Gia_ObjFaninC0(o)))
	}
	switch {
	case Gia_ObjIsConst0(obj) != 0:
		fmt.Print("constant 0")
	case Gia_ObjIsPi(p, obj) != 0:
		fmt.Print("PI")
	case Gia_ObjIsPo(p, obj) != 0:
		input("PO", obj)
	case Gia_ObjIsCi(obj) != 0:
		input("RO", Gia_ObjRoToRi(p, obj))
	case Gia_ObjIsCo(obj) != 0:
		input("RI", obj)
	case Gia_ObjIsXor(obj) != 0:
		fmt.Printf("XOR( %4d%s, %4d%s )", Gia_ObjFaninId0p(p, obj), mark(Gia_ObjFaninC0(obj)), Gia_ObjFaninId1p(p, obj), mark(Gia_ObjFaninC1(obj)))
	case Gia_ObjIsMuxId(p, obj.id) != 0:
		fmt.Printf("MUX( %4d%s, %4d%s, %4d%s )", Gia_ObjFaninId2p(p, obj), mark(Gia_ObjFaninC2(p, obj)), Gia_ObjFaninId1p(p, obj), mark(Gia_ObjFaninC1(obj)), Gia_ObjFaninId0p(p, obj), mark(Gia_ObjFaninC0(obj)))
	default:
		fmt.Printf("AND( %4d%s, %4d%s )", Gia_ObjFaninId0p(p, obj), mark(Gia_ObjFaninC0(obj)), Gia_ObjFaninId1p(p, obj), mark(Gia_ObjFaninC1(obj)))
	}
	if p.Refs != nil {
		fmt.Printf(" (refs = %3d)", Gia_ObjRefNum(p, obj))
	}
	if obj.fMark0 != 0 {
		fmt.Print(" mark0")
	}
	if obj.fMark1 != 0 {
		fmt.Print(" mark1")
	}
	if Gia_ManHasMapping(p) != 0 && Gia_ObjIsLut(p, obj.id) != 0 {
		fmt.Printf(" Cut = { ")
		for _, id := range Gia_ObjLutFanins(p, obj.id) {
			fmt.Printf("%d ", id)
		}
		fmt.Print("}")
	}
	if Gia_ManHasMapping2(p) != 0 && Gia_ObjIsLut2(p, obj.id) != 0 {
		fmt.Printf(" Cut = { ")
		for _, id := range Vec_IntArray(Gia_ObjLutFanins2(p, obj.id)) {
			fmt.Printf("%d ", id)
		}
		fmt.Print("}")
	}
	fmt.Println()
}
