package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManComputeCarryOuts(p *Gia_Man_t) *Vec_Int_t {
	tm := giaDfsTiming(p)
	if tm == nil {
		panic("gia: missing timing manager")
	}
	result := vec.Vec_IntAlloc(tim.Tim_ManBoxNum(tm))
	for i := 0; i < tim.Tim_ManBoxNum(tm); i++ {
		last := tim.Tim_ManBoxInputLast(tm, i)
		obj := Gia_ObjFanin0(Gia_ManCo(p, int32(last)))
		if Gia_ObjIsCi(obj) == 0 {
			continue
		}
		box := tim.Tim_ManBoxForCi(tm, int(Gia_ObjCioId(obj)))
		if box >= 0 && int(Gia_ObjCioId(obj)) == tim.Tim_ManBoxOutputLast(tm, box) {
			vec.Vec_IntPush(result, obj.id)
		}
	}
	return result
}
func Gia_ManCheckIntegrityWithBoxes(p *Gia_Man_t) {
	if p.PManTime == nil {
		return
	}
	p.Refs = nil
	Gia_ManCreateRefs(p)
	var regs, carry int32
	for i := Gia_ManPoNum(p) - Gia_ManRegBoxNum(p); i < Gia_ManPoNum(p); i++ {
		obj := Gia_ObjFanin0(Gia_ManPo(p, i))
		if Gia_ObjIsCi(obj) == 0 {
			panic("gia: invalid complex flop driver")
		}
		if Gia_ObjRefNum(p, obj) > 1 {
			regs++
		}
	}
	for _, id := range vec.Vec_IntArray(Gia_ManComputeCarryOuts(p)) {
		if Gia_ObjRefNumId(p, id) > 1 {
			carry++
		}
	}
	if regs != 0 || carry != 0 {
		fmt.Printf("Warning: AIG with boxes has internal fanout in %d complex flops and %d carries.\n", regs, carry)
	}
	p.Refs = nil
}
