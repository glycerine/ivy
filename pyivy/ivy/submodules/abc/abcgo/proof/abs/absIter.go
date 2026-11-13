// Go port of absIter.c by Alan Mishchenko, UC Berkeley (2005).
package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bmc"
)

func Gia_ObjIsInGla(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return vec.Vec_IntEntry(gia.Gia_ManGateClasses(p), int(gia.Gia_ObjId(p, obj)))
}
func Gia_ObjAddToGla(p *Gia_Man_t, obj *Gia_Obj_t) {
	vec.Vec_IntWriteEntry(gia.Gia_ManGateClasses(p), int(gia.Gia_ObjId(p, obj)), 1)
}
func Gia_ObjRemFromGla(p *Gia_Man_t, obj *Gia_Obj_t) {
	vec.Vec_IntWriteEntry(gia.Gia_ManGateClasses(p), int(gia.Gia_ObjId(p, obj)), 0)
}

func Gia_IterTryImprove(p *Gia_Man_t, timeout, initialFrame int32) int32 {
	abstraction := Gia_ManDupAbsGates(p, gia.Gia_ManGateClasses(p))
	graph := gia.Gia_ManToAigSimple(abstraction)
	frames := int32(10000000)
	if initialFrame != 0 {
		frames = initialFrame + 1
	}
	var frame int32
	status := bmc.Saig_BmcPerform(graph, 0, frames, 2000, timeout, 0, 0, 0, 0, &frame, 1, 0)
	if status != 0 && status != -1 {
		panic("abs: unexpected bounded-check status")
	}
	aig.Aig_ManStop(graph)
	gia.Gia_ManStop(abstraction)
	return frame
}

// Gia_ManShrinkGla tries removing boundary objects while preserving the
// established bounded-checking depth; accepted removals update the original map.
func Gia_ManShrinkGla(p *Gia_Man_t, nFrameMax, timeout, fUsePdr, fUseSat, fUseBdd, verbose int32) *Gia_Man_t {
	classes := gia.Gia_ManGateClasses(p)
	if gia.Gia_ManPoNum(p) != 1 || classes == nil {
		panic("abs: expected one output and an abstraction map")
	}
	saved := vec.Vec_IntDup(classes)
	initial := nFrameMax - 1
	if nFrameMax == 0 {
		initial = Gia_IterTryImprove(p, 0, 0)
	}
	var tried, removed int32
	started := util.Abc_Clock()
	for {
		changed := false
		for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
			obj := gia.Gia_ManObj(p, id)
			if gia.Gia_ObjMark0(obj) != 0 || Gia_ObjIsInGla(p, obj) == 0 || obj == gia.Gia_ObjFanin0(gia.Gia_ManPo(p, 0)) {
				continue
			}
			if gia.Gia_ObjIsAnd(obj) != 0 && Gia_ObjIsInGla(p, gia.Gia_ObjFanin0(obj)) != 0 && Gia_ObjIsInGla(p, gia.Gia_ObjFanin1(obj)) != 0 {
				continue
			}
			if gia.Gia_ObjIsRo(p, obj) != 0 && Gia_ObjIsInGla(p, gia.Gia_ObjFanin0(gia.Gia_ObjRoToRi(p, obj))) != 0 {
				continue
			}
			mark := util.Abc_Clock()
			fmt.Printf("%5d : Obj =%7d   ", tried, id)
			Gia_ObjRemFromGla(p, obj)
			frame := Gia_IterTryImprove(p, timeout, initial)
			if (nFrameMax != 0 && frame > nFrameMax) || (nFrameMax == 0 && frame > initial) {
				panic("abs: bounded checker exceeded frame limit")
			}
			fmt.Printf("Frame =%6d   ", frame)
			if frame < initial {
				gia.Gia_ObjSetMark0(obj, 1)
				Gia_ObjAddToGla(p, obj)
				fmt.Print("           ")
			} else {
				changed = true
				removed++
				fmt.Print("Removing   ")
				vec.Vec_IntWriteEntry(saved, int(id), 0)
			}
			util.Abc_PrintTime(1, "Time", util.Abc_Clock()-mark)
			tried++
			copy(vec.Vec_IntArray(classes), vec.Vec_IntArray(saved))
		}
		if !changed {
			break
		}
	}
	gia.Gia_ManCleanMark0(p)
	fmt.Printf("Tried = %d.  Removed = %d. (%.2f %%)  ", tried, removed, 100*float64(removed)/float64(vec.Vec_IntCountPositive(classes)))
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-started)
	_ = fUsePdr
	_ = fUseSat
	_ = fUseBdd
	_ = verbose
	return nil
}
