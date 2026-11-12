package sfm

import (
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
)

func Sfm_LibPrintGate(cell *mio.Mio_Cell2_t, pins []byte, inner *mio.Mio_Cell2_t, innerPins []byte) {
	fmt.Printf(" %-20s(", cell.Name())
	for _, pin := range pins[:cell.Fanins()] {
		if pin == 16 {
			if inner == nil {
				panic("sfm: missing inner gate")
			}
			Sfm_LibPrintGate(inner, innerPins, nil, nil)
		} else {
			fmt.Printf(" %c", 'a'+pin)
		}
	}
	fmt.Print(" )")
}
func Sfm_LibPrintObj(p *Sfm_Lib_t, obj *Sfm_Fun_t) {
	bottom := &p.pCells[obj.pFansB[0]]
	var top *mio.Mio_Cell2_t
	n := bottom.Fanins()
	if obj.pFansT[0] != 0 {
		top = &p.pCells[obj.pFansT[0]]
		n += top.Fanins() - 1
	}
	fmt.Printf("F = %d  A =%6.2f  ", n, float64(obj.Area)/1000)
	if top == nil {
		Sfm_LibPrintGate(bottom, obj.pFansB[1:], nil, nil)
	} else {
		Sfm_LibPrintGate(top, obj.pFansT[1:], bottom, obj.pFansB[1:])
	}
	if p.fDelay != 0 {
		id := Sfm_LibFunId(p, obj)
		for _, v := range p.profiles[id] {
			fmt.Printf("%6.2f ", float64(v)/1000)
		}
	}
}
func Sfm_LibPrint(p *Sfm_Lib_t) {
	for id := 2; id < vec.Vec_MemEntryNum(p.vTtMem); id++ {
		if p.hits[id] == 0 || p.lists[id] < 0 {
			continue
		}
		fmt.Printf("%8d : Num =%5d  Hit =%4d  ", id, p.counts[id], p.hits[id])
		Sfm_LibPrintObj(p, &p.pObjs[p.lists[id]])
		fmt.Print("    ")
		dau.Dau_DsdPrintFromTruth(vec.Vec_MemReadEntry(p.vTtMem, id), p.nVars)
	}
}
func Sfm_LibTest() {
	if abcmain.Abc_FrameReadLibGen() == nil {
		fmt.Println("There is no current library.")
		return
	}
	p := Sfm_LibPrepare(7, 1, 1, 1, 1)
	Sfm_LibPrint(p)
	Sfm_LibStop(p)
}
func Sfm_LibImplementSimple(p *Sfm_Lib_t, truth []uint64, fanins []int32) ([]int32, [][]int32) {
	allZero, allOne, variable, inverse := true, true, true, true
	for _, v := range truth[:p.nWords] {
		allZero = allZero && v == 0
		allOne = allOne && v == ^uint64(0)
		variable = variable && v == 0xaaaaaaaaaaaaaaaa
		inverse = inverse && v == 0x5555555555555555
	}
	index := -1
	switch {
	case allZero:
		index = 0
	case allOne:
		index = 1
	case variable:
		index = 2
	case inverse:
		index = 3
	default:
		panic("sfm: truth is not constant or buffer")
	}
	if index >= len(p.pCells) {
		panic("sfm: missing simple library gates")
	}
	gate := p.pCells[index].Gate()
	if gate == nil || ((allZero || allOne) && len(fanins) != 0) || ((variable || inverse) && len(fanins) != 1) {
		panic("sfm: invalid simple implementation")
	}
	return []int32{mio.Mio_GateReadValue(gate)}, [][]int32{append([]int32(nil), fanins...)}
}
