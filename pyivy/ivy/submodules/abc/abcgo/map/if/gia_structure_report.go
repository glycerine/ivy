package if_

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/opt/dau"
)

// Gia_ManTestStruct reports how many mapped LUTs admit the 4-4 structure.
func Gia_ManTestStruct(p *Gia_Man_t) {
	var counts, nondecomposable [8]int32
	for id := int32(1); id < Gia_ManObjNum(p); id++ {
		if gia.Gia_ObjIsLut(p, id) == 0 {
			continue
		}
		leaves, truth, ok := p.LUT(id)
		if !ok || len(leaves) > 7 {
			panic("if: invalid LUT in structure report")
		}
		n := int32(len(leaves))
		counts[max(n, 5)]++
		if n <= 5 {
			continue
		}
		words := make([]uint32, 2*len(truth))
		for i, w := range truth {
			words[2*i], words[2*i+1] = uint32(w), uint32(w>>32)
		}
		if If_CutPerformCheck07(nil, words, 7, n, "") != 0 {
			continue
		}
		nondecomposable[n]++
		if nondecomposable[n] <= 10 {
			dau.Dau_DsdPrintFromTruth(truth, n)
		}
	}
	fmt.Printf("LUT5 = %d    LUT6 = %d  NonDec = %d (%.2f %%)    LUT7 = %d  NonDec = %d (%.2f %%)\n", counts[5], counts[6], nondecomposable[6], 100*float64(nondecomposable[6])/float64(max(counts[6], 1)), counts[7], nondecomposable[7], 100*float64(nondecomposable[7])/float64(max(counts[7], 1)))
}
