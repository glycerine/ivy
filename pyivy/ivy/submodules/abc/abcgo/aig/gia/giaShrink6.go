package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/bool/rsb"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

// Gia_ManMapShrink6 resynthesizes mapped LUTs using divisors and bidecomposition.
func Gia_ManMapShrink6(g *Gia_Man_t, maxFanout, keepLevel, verbose int32) *Gia_Man_t {
	_ = keepLevel
	if Gia_ManHasMapping(g) == 0 {
		panic("gia: LUT shrinking requires mapping")
	}
	started := time.Now()
	p := Shr_ManAlloc(g)
	Gia_ManFillValue(g)
	Gia_ManConst0(g).Value = 0
	patterns := [6]word{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}
	var resubs, decomps int32
	for id := int32(1); id < g.nObjs; id++ {
		obj := Gia_ManObj(g, id)
		switch {
		case Gia_ObjIsCi(obj) != 0:
			lit := Gia_ManAppendCi(p.pNew)
			obj.Value = uint32(lit)
			if g.vLevels != nil {
				Gia_ObjSetLevelId(p.pNew, lit>>1, Gia_ObjLevelId(g, id))
			}
		case Gia_ObjIsCo(obj) != 0:
			obj.Value = uint32(Gia_ManAppendCo(p.pNew, Gia_ObjFanin0Copy(obj)))
		case Gia_ObjIsLut(g, id) != 0:
			vec.Vec_IntClear(p.vLeaves)
			for _, fan := range Gia_ObjLutFanins(g, id) {
				vec.Vec_IntPush(p.vLeaves, fan)
			}
			if vec.Vec_IntSize(p.vLeaves) > 6 {
				panic("gia: LUT exceeds six inputs")
			}
			truth := Shr_ManComputeTruth6(g, obj, p.vLeaves, p.vTruths)
			if truth == 0 {
				obj.Value = 0
			} else if truth == ^word(0) {
				obj.Value = 1
			} else {
				for i, fan := range vec.Vec_IntArray(p.vLeaves) {
					if truth == patterns[i] {
						obj.Value = Gia_ManObj(g, fan).Value
						break
					}
					if ^truth == patterns[i] {
						obj.Value = Gia_ManObj(g, fan).Value ^ 1
						break
					}
				}
			}
			if obj.Value != ^uint32(0) {
				continue
			}
			for i, fan := range vec.Vec_IntArray(p.vLeaves) {
				value := Gia_ManObj(g, fan).Value
				if value == ^uint32(0) {
					panic("gia: unmapped shrink LUT fanin")
				}
				if value&1 != 0 {
					shift := uint(1) << uint(i)
					truth = ((truth & patterns[i]) >> shift) | ((truth &^ patterns[i]) << shift)
				}
				vec.Vec_IntWriteEntry(p.vLeaves, i, int32(value>>1))
			}
			Shr_ManCollectDivisors(p, p.vLeaves, p.nDivMax, maxFanout)
			Shr_ManComputeTruths(p.pNew, int32(vec.Vec_IntSize(p.vLeaves)), p.vDivs, p.vDivTruths, p.vTruths)
			var off, on word
			if rsb.Rsb_ManPerformResub6(p.pManRsb, int32(vec.Vec_IntSize(p.vLeaves)), truth, vec.Vec_WrdArray(p.vDivTruths), &off, &on, 0) != 0 {
				vec.Vec_IntClear(p.vDivResub)
				for _, index := range rsb.Rsb_ManGetFanins(p.pManRsb) {
					vec.Vec_IntPush(p.vDivResub, vec.Vec_IntEntry(p.vDivs, int(index)))
				}
				obj.Value = uint32(Shr_ObjPerformBidec(p, p.pManDec, p.pNew, p.vDivResub, on, off|on))
				resubs++
			} else {
				obj.Value = uint32(Shr_ObjPerformBidec(p, p.pManDec, p.pNew, p.vLeaves, truth, ^word(0)))
				decomps++
			}
		}
	}
	if verbose != 0 {
		fmt.Printf("Performed %d resubs and %d decomps. Gain in AIG nodes = %d. Runtime %.2f sec\n", resubs, decomps, g.nObjs-p.pNew.nObjs, time.Since(started).Seconds())
	}
	return Shr_ManFree(p)
}
