package gia

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/glycerine/abcgo/misc/tim"
)

// Gia_ManPrintGetMuxFanins identifies the control and two data nodes of a
// structurally represented mux. The returned node IDs do not include phases.
func Gia_ManPrintGetMuxFanins(p *Gia_Man_t, obj *Gia_Obj_t, fanins []int32) {
	var data0, data1 *Gia_Obj_t
	control := Gia_ObjRecognizeMux(obj, &data1, &data0)
	fanins[0] = Gia_ObjId(p, Gia_Regular(control))
	fanins[1] = Gia_ObjId(p, Gia_Regular(data1))
	fanins[2] = Gia_ObjId(p, Gia_Regular(data0))
}

// Gia_ManCountDupLut counts repeated and primary-input mux data fanins.
func Gia_ManCountDupLut(p *Gia_Man_t) int32 {
	var duplicates, inputs, muxes int32
	Gia_ManCleanMark01(p)
	defer Gia_ManCleanMark01(p)
	count := func(fanin *Gia_Obj_t) {
		inputs += Gia_ObjIsCi(fanin)
		duplicates += int32(fanin.fMark0)
		muxes += int32(fanin.fMark1)
		fanin.fMark0 = 1
	}
	for id := int32(1); id < Gia_ManObjNum(p); id++ {
		if Gia_ObjIsLut(p, id) == 0 || Gia_ObjLutIsMux(p, id) == 0 {
			continue
		}
		obj := Gia_ManObj(p, id)
		obj.fMark1 = 1
		switch Gia_ObjLutSize(p, id) {
		case 3:
			var fanins [3]int32
			Gia_ManPrintGetMuxFanins(p, obj, fanins[:])
			count(Gia_ManObj(p, fanins[1]))
			count(Gia_ManObj(p, fanins[2]))
		case 2:
			fanin := Gia_ObjFanin0(obj)
			if fanin.fMark0 != 0 || fanin.fMark1 != 0 {
				fanin = Gia_ObjFanin1(obj)
			}
			count(fanin)
		default:
			panic("gia: mux LUT must have two or three fanins")
		}
	}
	if duplicates+inputs+muxes != 0 {
		fmt.Printf("Dup fanins = %d.  CI fanins = %d.  MUX fanins = %d.  Total = %d.  (%.2f %%)\n", duplicates, inputs, muxes, duplicates+inputs, 100*float64(duplicates+inputs+muxes)/float64(Gia_ManLutNum(p)))
	}
	return duplicates + inputs
}

var giaMappingStatsState struct {
	sync.Mutex
	name  string
	since time.Time
}

// Gia_ManPrintMappingStats reports LUT count, edges and mapped depth. If a
// path is provided, it also appends the source-compatible statistics table.
func Gia_ManPrintMappingStats(p *Gia_Man_t, dumpFile string) {
	if Gia_ManHasMapping(p) == 0 {
		return
	}
	levels := make([]int32, Gia_ManObjNum(p))
	var maxSize, luts, fanins, maxLevel, totalLevel, muxes int32
	for id := int32(1); id < Gia_ManObjNum(p); id++ {
		if Gia_ObjIsLut(p, id) == 0 {
			continue
		}
		obj := Gia_ManObj(p, id)
		if Gia_ObjLutIsMux(p, id) != 0 {
			switch Gia_ObjLutSize(p, id) {
			case 3:
				var fs [3]int32
				Gia_ManPrintGetMuxFanins(p, obj, fs[:])
				levels[id] = max(levels[fs[0]]+1, levels[fs[1]], levels[fs[2]])
			case 2:
				levels[id] = max(levels[Gia_ObjFaninId0(obj, id)], levels[Gia_ObjFaninId1(obj, id)])
			}
			maxLevel = max(maxLevel, levels[id])
			fanins++
			muxes++
			continue
		}
		luts++
		size := Gia_ObjLutSize(p, id)
		fanins += size
		maxSize = max(maxSize, size)
		for _, fanin := range Gia_ObjLutFanins(p, id) {
			levels[id] = max(levels[id], levels[fanin])
		}
		levels[id]++
		maxLevel = max(maxLevel, levels[id])
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		totalLevel += levels[Gia_ObjFaninId0p(p, Gia_ManCo(p, i))]
	}
	var average float64
	if n := Gia_ManCoNum(p); n != 0 {
		average = float64(totalLevel) / float64(n)
	}
	fmt.Printf("Mapping (K=%d)  :  \033[1;33mlut =%7d\033[0m  \033[1;32medge =%8d\033[0m  \033[1;31mlev =%5d\033[0m \033[1;31m(%.2f)\033[0m  ", maxSize, luts, fanins, maxLevel, average)
	if timing, ok := p.PManTime.(*tim.Tim_Man_t); ok && tim.Tim_ManBoxNum(timing) > 0 {
		fmt.Printf("\033[1;31mlevB =%5d\033[0m  ", Gia_ManLutLevelWithBoxes(p))
	}
	fmt.Printf("mem =%5.2f MB\n", 4*float64(Gia_ManObjNum(p)+2*luts+fanins)/(1<<20))
	if muxes != 0 {
		Gia_ManCountDupLut(p)
	}
	if dumpFile == "" {
		return
	}
	giaMappingStatsState.Lock()
	defer giaMappingStatsState.Unlock()
	table, err := os.OpenFile(dumpFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gia: cannot open mapping statistics: %v\n", err)
		return
	}
	defer table.Close()
	now := time.Now()
	if giaMappingStatsState.name != p.pName || giaMappingStatsState.since.IsZero() {
		fmt.Fprintf(table, "\n%s  %d %d %d ", p.pName, Gia_ManAndNum(p), luts, Gia_ManLutLevelWithBoxes(p))
		giaMappingStatsState.name = p.pName
	} else {
		fmt.Fprintf(table, " %d %d %.2f", luts, Gia_ManLutLevelWithBoxes(p), now.Sub(giaMappingStatsState.since).Seconds())
	}
	giaMappingStatsState.since = now
}
