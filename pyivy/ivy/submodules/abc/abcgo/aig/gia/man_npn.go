package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/bool/kit"
	dar "github.com/glycerine/abcgo/opt/dar/core"
)

func Gia_ManPrintNpnClasses(p *Gia_Man_t) {
	if Gia_ManHasMapping(p) == 0 || Gia_ManLutSizeMax(p) > 4 {
		panic("gia: NPN statistics require LUT4 mapping")
	}
	names := kit.Kit_DsdNpn4ClassNames()
	var counts [222]int32
	classes := make([]int, p.nObjs)
	Gia_ManCleanTruth(p)
	for id := int32(0); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) == 0 || Gia_ObjLutSize(p, id) > 4 {
			continue
		}
		leaves := Gia_ObjLutFanins(p, id)
		truth := giaNewTruthEval(p, 4, leaves).node(id)[0]
		class := dar.Dar_LibReturnClass(uint32(truth))
		if class < 0 || class >= len(counts) {
			panic("gia: invalid NPN class")
		}
		counts[class]++
		classes[id] = class
	}
	p.vTruths = nil
	var total int32
	for _, n := range counts {
		total += n
	}
	fmt.Printf("NPN CLASS STATISTICS (for %d LUT4 present in the current mapping):\n", total)
	var shown, other int32
	for i, n := range counts {
		if n == 0 {
			continue
		}
		shown++
		other += n
		fmt.Printf("%3d: Class %3d :  Count = %6d   (%7.2f %%)   %s\n", shown, i, n, 100*float64(n)/float64(total+1), names[i])
	}
	other = total - other
	fmt.Printf("Other     :  Count = %6d   (%7.2f %%)\n", other, 100*float64(other)/float64(total+1))
	p.Refs = nil
	Gia_ManSetRefsMapped(p)
	var all, mffc int32
	for id := int32(0); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) == 0 || classes[id] != 109 {
			continue
		}
		var n, nUnique int32
		for _, fanin := range Gia_ObjLutFanins(p, id) {
			if classes[fanin] == 109 {
				n++
				if Gia_ObjRefNumId(p, fanin) == 1 {
					nUnique++
				}
			}
		}
		if n > 1 {
			all++
		}
		if nUnique > 1 {
			mffc++
		}
	}
	p.Refs = nil
	fmt.Printf("Approximate number of 4:1 MUX structures: All = %6d  (%7.2f %%)  MFFC = %6d  (%7.2f %%)\n", all, 100*float64(all)/float64(total+1), mffc, 100*float64(mffc)/float64(total+1))
}
