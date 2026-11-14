package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
)

func Gia_ManLutLevelWithBoxes_rec(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return 0
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsCi(obj) != 0 {
		return 1
	}
	if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsLut(p, obj.id) == 0 {
		panic("gia: expected mapped LUT")
	}
	var level int32
	for _, fanin := range Gia_ObjLutFanins(p, obj.id) {
		if Gia_ManLutLevelWithBoxes_rec(p, Gia_ManObj(p, fanin)) != 0 {
			return 1
		}
		level = max(level, Gia_ObjLevelId(p, fanin))
	}
	Gia_ObjSetLevelId(p, obj.id, level+1)
	return 0
}
func Gia_ManLutLevelWithBoxes(p *Gia_Man_t) int32 {
	if p.nRegs != 0 {
		panic("gia: boxed LUT timing requires combinational graph")
	}
	if p.PManTime == nil {
		return Gia_ManLutLevel(p, nil)
	}
	timing, ok := p.PManTime.(*tim.Tim_Man_t)
	if !ok {
		panic("gia: invalid timing manager")
	}
	Gia_ManCleanLevels(p, p.nObjs)
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrent(p, Gia_ManConst0(p))
	for i := 0; i < tim.Tim_ManPiNum(timing); i++ {
		Gia_ObjSetTravIdCurrent(p, Gia_ManCi(p, int32(i)))
	}
	ci, co := tim.Tim_ManPiNum(timing), 0
	for box := 0; box < tim.Tim_ManBoxNum(timing); box++ {
		inputs, outputs := tim.Tim_ManBoxInputNum(timing, box), tim.Tim_ManBoxOutputNum(timing, box)
		table := tim.Tim_ManBoxDelayTable(timing, box)
		if len(table) < 3+inputs*outputs || int(table[1]) != inputs {
			panic("gia: invalid box delay table")
		}
		for i := 0; i < inputs; i++ {
			obj := Gia_ManCo(p, int32(co+i))
			if Gia_ManLutLevelWithBoxes_rec(p, Gia_ObjFanin0(obj)) != 0 {
				fmt.Println("Boxes are not in a topological order. Switching to level computation without boxes.")
				return Gia_ManLevelNum(p)
			}
			Gia_ObjSetCoLevel(p, obj)
		}
		for i := 0; i < outputs; i++ {
			obj := Gia_ManCi(p, int32(ci+i))
			Gia_ObjSetTravIdCurrent(p, obj)
			var level int32
			for j := 0; j < inputs; j++ {
				if int32(table[3+i*inputs+j]) != -giaInfinity {
					level = max(level, Gia_ObjLevel(p, Gia_ManCo(p, int32(co+j)))+1)
				}
			}
			Gia_ObjSetLevel(p, obj, level)
		}
		ci += outputs
		co += inputs
	}
	p.nLevels = 0
	for i := tim.Tim_ManCoNum(timing) - tim.Tim_ManPoNum(timing); i < tim.Tim_ManCoNum(timing); i++ {
		obj := Gia_ManCo(p, int32(i))
		Gia_ManLutLevelWithBoxes_rec(p, Gia_ObjFanin0(obj))
		Gia_ObjSetCoLevel(p, obj)
		p.nLevels = max(p.nLevels, Gia_ObjLevel(p, obj))
	}
	co += tim.Tim_ManPoNum(timing)
	if ci != int(Gia_ManCiNum(p)) || co != int(Gia_ManCoNum(p)) {
		panic("gia: timing terminal count mismatch")
	}
	return p.nLevels
}
func Gia_ManLutParams(p *Gia_Man_t, luts, edges, levels *int32) {
	*luts = Gia_ManLutNum(p)
	*edges = Gia_ManLutFaninCount(p)
	if timing, ok := p.PManTime.(*tim.Tim_Man_t); ok && tim.Tim_ManBoxNum(timing) > 0 {
		*levels = Gia_ManLutLevelWithBoxes(p)
	} else {
		*levels = Gia_ManLutLevel(p, nil)
	}
}
