package sfm

import (
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
)

type Sfm_LibMatch struct {
	Bottom, Top         *mio.Mio_Gate_t
	BottomPins, TopPins []byte
}

func Sfm_LibFindAreaMatch(p *Sfm_Lib_t, truth []uint64) (int32, int32) {
	id := *vec.Vec_MemHashLookup(p.vTtMem, truth)
	if id < 0 || int(id) >= len(p.lists) || p.lists[id] < 0 {
		return -1, -1
	}
	obj := p.lists[id]
	return p.pObjs[obj].Area, obj
}
func Sfm_LibFindDelayMatches(p *Sfm_Lib_t, truth []uint64) []Sfm_LibMatch {
	id := *vec.Vec_MemHashLookup(p.vTtMem, truth)
	if id < 0 || int(id) >= len(p.lists) {
		return nil
	}
	p.hits[id]++
	var result []Sfm_LibMatch
	for i := p.lists[id]; i >= 0; i = p.pObjs[i].Next {
		obj := &p.pObjs[i]
		bottom := &p.pCells[obj.pFansB[0]]
		match := Sfm_LibMatch{Bottom: bottom.Gate(), BottomPins: append([]byte(nil), obj.pFansB[1:1+bottom.Fanins()]...)}
		if obj.pFansT[0] != 0 {
			top := &p.pCells[obj.pFansT[0]]
			match.Top = top.Gate()
			match.TopPins = append([]byte(nil), obj.pFansT[1:1+top.Fanins()]...)
		}
		result = append(result, match)
	}
	return result
}
func Sfm_LibImplementGatesDelay(fanins []int32, bottom, top *mio.Mio_Gate_t, bottomPins, topPins []byte) ([]int32, [][]int32) {
	gates := []int32{mio.Mio_GateReadValue(bottom)}
	levels := [][]int32{make([]int32, len(bottomPins))}
	for i, pin := range bottomPins {
		levels[0][i] = fanins[pin]
	}
	if top == nil {
		return gates, levels
	}
	gates = append(gates, mio.Mio_GateReadValue(top))
	levels = append(levels, make([]int32, len(topPins)))
	for i, pin := range topPins {
		if pin == 16 {
			levels[1][i] = 0
		} else {
			levels[1][i] = fanins[pin]
		}
	}
	return gates, levels
}
func Sfm_LibImplementGatesArea(p *Sfm_Lib_t, fanins []int32, id int32) ([]int32, [][]int32) {
	obj := &p.pObjs[id]
	bottom := &p.pCells[obj.pFansB[0]]
	var top *mio.Mio_Cell2_t
	if obj.pFansT[0] != 0 {
		top = &p.pCells[obj.pFansT[0]]
	}
	bp := obj.pFansB[1 : 1+bottom.Fanins()]
	var gate *mio.Mio_Gate_t
	var tp []byte
	if top != nil {
		gate = top.Gate()
		tp = obj.pFansT[1 : 1+top.Fanins()]
	}
	return Sfm_LibImplementGatesDelay(fanins, bottom.Gate(), gate, bp, tp)
}
