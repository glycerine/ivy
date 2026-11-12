package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
)

// Gia_ManDupUnshuffleInputs moves register outputs ahead of box outputs.
func Gia_ManDupUnshuffleInputs(src *Gia_Man_t) *Gia_Man_t {
	if Gia_ManIsNormalized(src) == 0 || Gia_ManIsSeqWithBoxes(src) == 0 {
		panic("gia: expected normalized sequential graph with boxes")
	}
	timing := src.pManTime.(*tim.Tim_Man_t)
	cis, all := int32(tim.Tim_ManPiNum(timing)), int32(tim.Tim_ManCiNum(timing))
	pis := cis - Gia_ManRegNum(src)
	if all != Gia_ManCiNum(src) || pis <= 0 {
		panic("gia: invalid boxed input partition")
	}
	Gia_ManFillValue(src)
	dst := Gia_ManStart(Gia_ManObjNum(src))
	dst.pName, dst.pSpec = src.pName, src.pSpec
	Gia_ManConst0(src).Value = 0
	for _, part := range [][2]int32{{0, pis}, {all - Gia_ManRegNum(src), all}, {pis, all - Gia_ManRegNum(src)}} {
		for i := part[0]; i < part[1]; i++ {
			Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(dst))
		}
	}
	fmt.Println("Warning: Unshuffled CI order to be correct AIG with boxes.")
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			obj.Value = uint32(Gia_ManAppendAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		}
	}
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		obj := Gia_ManCo(src, i)
		obj.Value = uint32(Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj)))
	}
	Gia_ManSetRegNum(dst, Gia_ManRegNum(src))
	dst.nConstrs = src.nConstrs
	if Gia_ManIsNormalized(dst) == 0 {
		panic("gia: unshuffle produced unnormalized graph")
	}
	Gia_ManDupRemapEquiv(dst, src)
	return dst
}
