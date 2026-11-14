package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
)

func Gia_ManLevelWithBoxes_rec(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	stack := []giaDfsFrame{{id: Gia_ObjId(p, obj)}}
	for len(stack) > 0 {
		frame := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		node := Gia_ManObj(p, frame.id)
		if frame.expanded {
			Gia_ObjSetAndLevel(p, node)
			continue
		}
		if Gia_ObjIsTravIdCurrentId(p, frame.id) != 0 {
			continue
		}
		Gia_ObjSetTravIdCurrentId(p, frame.id)
		if Gia_ObjIsCi(node) != 0 {
			return 1
		}
		if Gia_ObjIsAnd(node) == 0 {
			panic("gia: box timing requires internal AND nodes")
		}
		// The source visits choices before fanins without using their failure result.
		if sibling := Gia_ObjSibl(p, node.id); sibling != 0 {
			Gia_ManLevelWithBoxes_rec(p, Gia_ManObj(p, sibling))
		}
		stack = append(stack, giaDfsFrame{id: frame.id, expanded: true}, giaDfsFrame{id: Gia_ObjFaninId1(node, node.id)}, giaDfsFrame{id: Gia_ObjFaninId0(node, node.id)})
	}
	return 0
}
func Gia_ManLevelWithBoxes(p *Gia_Man_t) int32 {
	if p.PManTime == nil {
		return Gia_ManLevelNum(p)
	}
	timing, ok := p.PManTime.(*tim.Tim_Man_t)
	if !ok {
		panic("gia: incompatible timing manager")
	}
	if p.nRegs != 0 || Gia_ManBufNum(p) != 0 {
		panic("gia: box timing requires combinational logic without buffers")
	}
	delay := p.nAnd2Delay
	if delay == 0 {
		delay = 1
	}
	Gia_ManCleanLevels(p, p.nObjs)
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrentId(p, 0)
	for i := 0; i < tim.Tim_ManPiNum(timing); i++ {
		obj := Gia_ManCi(p, int32(i))
		Gia_ObjSetLevel(p, obj, int32(tim.Tim_ManGetCiArrival(timing, i)/float32(delay)))
		Gia_ObjSetTravIdCurrent(p, obj)
	}
	currentCI, currentCO := tim.Tim_ManPiNum(timing), 0
	for box := 0; box < tim.Tim_ManBoxNum(timing); box++ {
		inputs, outputs := tim.Tim_ManBoxInputNum(timing, box), tim.Tim_ManBoxOutputNum(timing, box)
		table := tim.Tim_ManBoxDelayTable(timing, box)
		if int(table[1]) != inputs {
			panic("gia: box delay table input count mismatch")
		}
		for i := 0; i < inputs; i++ {
			obj := Gia_ManCo(p, int32(currentCO+i))
			if Gia_ManLevelWithBoxes_rec(p, Gia_ObjFanin0(obj)) != 0 {
				fmt.Println("Boxes are not in a topological order. Switching to level computation without boxes.")
				return Gia_ManLevelNum(p)
			}
			Gia_ObjSetCoLevel(p, obj)
		}
		for i := 0; i < outputs; i++ {
			obj := Gia_ManCi(p, int32(currentCI+i))
			Gia_ObjSetTravIdCurrent(p, obj)
			var level int32
			for j := 0; j < inputs; j++ {
				arc := int32(table[3+i*inputs+j])
				if arc != -giaInfinity {
					level = max(level, Gia_ObjLevel(p, Gia_ManCo(p, int32(currentCO+j)))+arc/delay)
				}
			}
			Gia_ObjSetLevel(p, obj, level)
		}
		currentCO += inputs
		currentCI += outputs
	}
	p.nLevels = 0
	for i := tim.Tim_ManCoNum(timing) - tim.Tim_ManPoNum(timing); i < tim.Tim_ManCoNum(timing); i++ {
		obj := Gia_ManCo(p, int32(i))
		Gia_ManLevelWithBoxes_rec(p, Gia_ObjFanin0(obj))
		Gia_ObjSetCoLevel(p, obj)
		p.nLevels = max(p.nLevels, Gia_ObjLevel(p, obj))
	}
	currentCO += tim.Tim_ManPoNum(timing)
	if currentCI != int(Gia_ManCiNum(p)) || currentCO != int(Gia_ManCoNum(p)) {
		panic("gia: timing terminal count mismatch")
	}
	return p.nLevels
}

// BoxInputObjects returns independent CO object IDs driving the box owning this
// CI object, or nil for a primary input/a manager without timing boxes.
func (p *Gia_Man_t) BoxInputObjects(ciID int32) []int32 {
	if p.PManTime == nil {
		return nil
	}
	timing, ok := p.PManTime.(*tim.Tim_Man_t)
	if !ok {
		panic("gia: incompatible timing manager")
	}
	obj := Gia_ManObj(p, ciID)
	if Gia_ObjIsCi(obj) == 0 {
		panic("gia: box lookup requires a CI object")
	}
	box := tim.Tim_ManBoxForCi(timing, int(Gia_ObjCioId(obj)))
	if box < 0 {
		return nil
	}
	first, count := tim.Tim_ManBoxInputFirst(timing, box), tim.Tim_ManBoxInputNum(timing, box)
	result := make([]int32, count)
	for i := range result {
		result[i] = Gia_ObjId(p, Gia_ManCo(p, int32(first+i)))
	}
	return result
}

// BoxLevelInputObjects excludes disabled (-ABC_INFINITY) arcs for this box
// output, while BoxInputObjects retains every structural box input.
func (p *Gia_Man_t) BoxLevelInputObjects(ciID int32) []int32 {
	inputs := p.BoxInputObjects(ciID)
	if inputs == nil {
		return nil
	}
	timing := p.PManTime.(*tim.Tim_Man_t)
	ci := int(Gia_ObjCioId(Gia_ManObj(p, ciID)))
	box := tim.Tim_ManBoxForCi(timing, ci)
	output := ci - tim.Tim_ManBoxOutputFirst(timing, box)
	table := tim.Tim_ManBoxDelayTable(timing, box)
	if table == nil {
		panic("gia: box level lookup requires a delay table")
	}
	result := make([]int32, 0, len(inputs))
	for i, id := range inputs {
		if int32(table[3+output*len(inputs)+i]) != -giaInfinity {
			result = append(result, id)
		}
	}
	return result
}
