package gia

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_RsbEnumerateWindows constructs and optimizes reconvergent windows.
func Gia_RsbEnumerateWindows(p *Gia_Man_t, maxInputs, maxLevels int32) {
	start := time.Now()
	levels := vec.Vec_WecStart(int(Gia_ManLevelNum(p) + 1))
	paths := vec.Vec_IntStart(int(p.nObjs))
	refs := vec.Vec_IntStart(int(p.nObjs))
	Gia_ManStaticFanoutStart(p)
	defer Gia_ManStaticFanoutStop(p)
	var count, totalWin, totalIns, totalOut, gain int32
	for id := int32(0); id < p.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(p, id)) == 0 {
			continue
		}
		var win, ins *Vec_Int_t
		if Gia_RsbWindowCompute(p, id, maxInputs, maxLevels, levels, paths, &win, &ins) == 0 {
			continue
		}
		outs := Gia_RsbFindOutputs(p, win, ins, refs)
		count++
		totalWin += int32(vec.Vec_IntSize(win))
		totalIns += int32(vec.Vec_IntSize(ins))
		totalOut += int32(vec.Vec_IntSize(outs))
		if vec.Vec_IntSize(win) > 1000 {
			fmt.Printf("Obj %d.   Window: Ins = %d. Ands = %d. Outs = %d.\n", id, vec.Vec_IntSize(ins), vec.Vec_IntSize(win)-vec.Vec_IntSize(ins), vec.Vec_IntSize(outs))
		}
		before := Gia_RsbDeriveGiaFromWindows(p, win, ins, outs)
		after := Gia_ManResub2Test(before)
		if Gia_ManVerifyTwoTruths(before, after) == 0 {
			fmt.Printf("Window %d failed verification.\n", id)
		}
		gain += Gia_ManAndNum(before) - Gia_ManAndNum(after)
	}
	denominator := float64(max(1, count))
	fmt.Printf("Computed windows for %d nodes (out of %d). Ave inputs = %.2f. Ave outputs = %.2f. Ave volume = %.2f. Gain = %d. Time = %s\n", count, Gia_ManAndNum(p), float64(totalIns)/denominator, float64(totalOut)/denominator, float64(totalWin)/denominator, gain, time.Since(start))
}

func Gia_RsbTryOneWindow(p *Gia_Man_t) *Gia_Man_t { return Gia_ManResub2Test(p) }
