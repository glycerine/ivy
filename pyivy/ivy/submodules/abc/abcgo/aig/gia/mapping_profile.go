package gia

import abcutil "github.com/glycerine/abcgo/misc/util"

// Gia_ManPrintPackingStats is giaIf.c's; blocks of more than four LUTs
// (&pack -N 5) are not counted (C overflowed nNumStr[5], CBUGS_FIXED.md #30).
func Gia_ManPrintPackingStats(p *Gia_Man_t) {
	var nNumStr [5]int32
	MaxSize, Count := -1, int32(0)
	if p.VPacking == nil {
		return
	}
	nEntries2 := int32(0)
	for i := 1; i < Vec_IntSize(p.VPacking); i++ {
		Entry := Vec_IntEntry(p.VPacking, i)
		if Entry < 5 {
			nNumStr[Entry]++
		}
		i++
		i += int(Entry)
		i--
		nEntries2++
	}
	if nNumStr[3] > 0 {
		MaxSize = 3
	} else if nNumStr[2] > 0 {
		MaxSize = 2
	} else if nNumStr[1] > 0 {
		MaxSize = 1
	}
	abcutil.Abc_Print(1, "Packing (N=%d)  :  ", MaxSize)
	for i := 1; i <= MaxSize; i++ {
		abcutil.Abc_Print(1, "%d x LUT = %d   ", i, nNumStr[i])
		Count += int32(i) * nNumStr[i]
	}
	abcutil.Abc_Print(1, "Total = %d   ", nEntries2)
	abcutil.Abc_Print(1, "Total LUT = %d", Count)
	abcutil.Abc_Print(1, "\n")
}
func Gia_ManPrintNodeProfile(counts []int32, sizeMax int32) {
	if int(sizeMax) >= len(counts) {
		panic("gia: invalid LUT profile")
	}
	var total, weighted int32
	for i := int32(0); i <= sizeMax; i++ {
		total += counts[i]
		weighted += i * counts[i]
	}
	abcutil.Abc_Print(1, "LUT = %d : ", total)
	for i := int32(2); i <= sizeMax; i++ {
		percent := 100 * float64(counts[i]) / float64(total)
		abcutil.Abc_Print(1, "%d=%d %.1f %%  ", i, counts[i], percent)
	}
	abcutil.Abc_Print(1, "Ave = %.2f\n", float64(weighted)/float64(max(total, 1)))
}
func Gia_ManPrintLutStats(p *Gia_Man_t) {
	maximum := Gia_ManLutSizeMax(p)
	if maximum > 32 {
		abcutil.Abc_Print(1, "The max LUT size (%d) is too large.\n", maximum)
		return
	}
	if maximum < 0 {
		Gia_ManPrintNodeProfile(nil, maximum)
		return
	}
	counts := make([]int32, maximum+1)
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) != 0 {
			counts[Gia_ObjLutSize(p, id)]++
		}
	}
	Gia_ManPrintNodeProfile(counts, maximum)
}
