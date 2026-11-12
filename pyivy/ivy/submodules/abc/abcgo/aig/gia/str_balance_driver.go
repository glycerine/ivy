package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Str_NtkBalanceMux(dst *Gia_Man_t, ntk *Str_Ntk_t, obj *Str_Obj_t, delays *Vec_Int_t, lutSize, groups, muxes, recursive, optArea, verbose int32) {
	start := obj.id
	for group := int32(0); group < groups; group++ {
		at := start + group*muxes
		result := Str_MuxRestructure(dst, ntk, at, muxes, delays, lutSize, recursive, optArea, verbose)
		if result < 0 {
			for i := int32(0); i < muxes; i++ {
				node := Str_NtkObj(ntk, at+i)
				node.iCopy = Gia_ManHashMuxReal(dst, Str_ObjFaninCopy(ntk, node, 2), Str_ObjFaninCopy(ntk, node, 1), Str_ObjFaninCopy(ntk, node, 0))
				Str_ObjDelay(dst, node.iCopy>>1, lutSize, delays)
			}
		} else {
			Str_NtkObj(ntk, at+muxes-1).iCopy = result
		}
	}
}
func Str_NtkBalance(original *Gia_Man_t, ntk *Str_Ntk_t, lutSize, useMuxes, recursive, optArea, verbose int32) *Gia_Man_t {
	if lutSize <= 0 || lutSize >= 16 || original.pMuxes != nil {
		panic("gia: invalid structured balancing input")
	}
	dst := Gia_ManStart(original.nObjs)
	dst.pName, dst.pSpec = original.pName, original.pSpec
	dst.EnableMuxes()
	vec.Vec_IntFill(&dst.vCopies, int(4*dst.nObjsAlloc), -1)
	delays := vec.Vec_IntStart(int(4 * dst.nObjsAlloc))
	Gia_ManHashStart(dst)
	var timing *tim.Tim_Man_t
	if original.pManTime != nil {
		var ok bool
		timing, ok = original.pManTime.(*tim.Tim_Man_t)
		if !ok {
			panic("gia: invalid timing manager")
		}
		tim.Tim_ManInitPiArrivalAll(timing, 17)
		tim.Tim_ManIncrementTravId(timing)
	}
	var maxDelay int32
	for id := int32(0); id < ntk.nObjs; id++ {
		obj := Str_NtkObj(ntk, id)
		switch obj.Type {
		case STR_CONST0:
			obj.iCopy = 0
			vec.Vec_IntWriteEntry(delays, 0, 17)
		case STR_PI:
			obj.iCopy = Gia_ManAppendCi(dst)
			arrival := int32(17)
			if timing != nil {
				arrival = int32(tim.Tim_ManGetCiArrival(timing, int(Gia_ObjCioId(Gia_ManObj(dst, obj.iCopy>>1)))))
			}
			vec.Vec_IntFillExtra(delays, int((obj.iCopy>>1)+1), 0)
			vec.Vec_IntWriteEntry(delays, int(obj.iCopy>>1), arrival)
		case STR_AND, STR_XOR:
			vec.Vec_IntFillExtra(&dst.vCopies, int(dst.nObjsAlloc*4), -1)
			Str_NtkBalanceMulti(dst, ntk, obj, delays, lutSize)
		case STR_MUX:
			if obj.iTop >= 0 && useMuxes != 0 {
				var groups, muxes int32
				Str_ObjReadGroup(ntk, obj, &groups, &muxes)
				if groups*muxes < 2 {
					panic("gia: invalid structured mux group")
				}
				Str_NtkBalanceMux(dst, ntk, obj, delays, lutSize, groups, muxes, recursive, optArea, verbose)
				id += groups*muxes - 1
			} else {
				obj.iCopy = Gia_ManHashMuxReal(dst, Str_ObjFaninCopy(ntk, obj, 2), Str_ObjFaninCopy(ntk, obj, 1), Str_ObjFaninCopy(ntk, obj, 0))
				Str_ObjDelay(dst, obj.iCopy>>1, lutSize, delays)
			}
		case STR_PO:
			lit := Str_ObjFaninCopy(ntk, obj, 0)
			obj.iCopy = Gia_ManAppendCo(dst, lit)
			arrival := vec.Vec_IntEntry(delays, int(lit>>1))
			maxDelay = max(maxDelay, arrival)
			if timing != nil {
				tim.Tim_ManSetCoArrival(timing, int(Gia_ObjCioId(Gia_ManObj(dst, obj.iCopy>>1))), float32(arrival))
			}
		default:
			panic("gia: unknown structured object")
		}
	}
	if verbose != 0 {
		fmt.Printf("Max delay = %d.  Old objs = %d.  New objs = %d.\n", maxDelay>>4, original.nObjs, dst.nObjs)
	}
	Gia_ManHashStop(dst)
	Gia_ManSetRegNum(dst, Gia_ManRegNum(original))
	result := Gia_ManDupNoMuxes(dst, 0)
	Gia_ManStop(dst)
	return result
}
func Gia_ManLutBalance(p *Gia_Man_t, lutSize, useMuxes, recursive, optArea, verbose int32) *Gia_Man_t {
	start := time.Now()
	work := p
	var oldTiming *tim.Tim_Man_t
	boxed := false
	if p.pManTime != nil {
		manager, ok := p.pManTime.(*tim.Tim_Man_t)
		if !ok {
			panic("gia: invalid timing manager")
		}
		if tim.Tim_ManBoxNum(manager) != 0 && Gia_ManIsNormalized(p) != 0 {
			oldTiming = manager
			p.pManTime = tim.Tim_ManDup(manager, 16)
			work = Gia_ManDupUnnormalize(p)
			if work == nil {
				tim.Tim_ManStop(p.pManTime.(*tim.Tim_Man_t))
				p.pManTime = oldTiming
				return nil
			}
			Gia_ManTransferTiming(work, p)
			boxed = true
		}
	}
	ntk := Str_ManNormalize(work)
	result := Str_NtkBalance(work, ntk, lutSize, useMuxes, recursive, optArea, verbose)
	Gia_ManTransferTiming(result, work)
	if boxed {
		Gia_ManStop(work)
		temp := result
		result = Gia_ManDupNormalize(temp, 0)
		Gia_ManTransferTiming(result, temp)
		Gia_ManStop(temp)
		tim.Tim_ManStop(result.pManTime.(*tim.Tim_Man_t))
		result.pManTime = oldTiming
	}
	if verbose != 0 {
		Str_NtkPs(ntk, start)
	}
	Str_NtkDelete(ntk)
	return result
}
