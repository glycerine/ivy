package gia

import "fmt"

func Gia_ManPrintPackingStats(p *Gia_Man_t) {
	if p.VPacking == nil {
		return
	}
	data := Vec_IntArray(p.VPacking)
	if len(data) == 0 {
		panic("gia: empty packing")
	}
	var counts [4]int32
	entries := int32(0)
	maximum := 0
	for i := 1; i < len(data); {
		size := int(data[i])
		i++
		if size < 1 || size > 3 || i+size > len(data) {
			panic("gia: invalid packing record")
		}
		counts[size]++
		maximum = max(maximum, size)
		i += size
		entries++
	}
	if entries != data[0] {
		panic("gia: packing count mismatch")
	}
	fmt.Printf("Packing (N=%d)  :  ", maximum)
	var total int32
	for size := 1; size <= maximum; size++ {
		fmt.Printf("%d x LUT = %d   ", size, counts[size])
		total += int32(size) * counts[size]
	}
	fmt.Printf("Total = %d   Total LUT = %d\n", entries, total)
}
func Gia_ManPrintNodeProfile(counts []int32, sizeMax int32) {
	if sizeMax < 0 || int(sizeMax) >= len(counts) {
		panic("gia: invalid LUT profile")
	}
	var total, weighted int32
	for i := int32(0); i <= sizeMax; i++ {
		total += counts[i]
		weighted += i * counts[i]
	}
	fmt.Printf("LUT = %d : ", total)
	for i := int32(2); i <= sizeMax; i++ {
		percent := float64(0)
		if total != 0 {
			percent = 100 * float64(counts[i]) / float64(total)
		}
		fmt.Printf("%d=%d %.1f %%  ", i, counts[i], percent)
	}
	fmt.Printf("Ave = %.2f\n", float64(weighted)/float64(max(total, 1)))
}
func Gia_ManPrintLutStats(p *Gia_Man_t) {
	maximum := Gia_ManLutSizeMax(p)
	if maximum > 32 {
		fmt.Printf("The max LUT size (%d) is too large.\n", maximum)
		return
	}
	if maximum < 0 {
		fmt.Println("LUT = 0 : Ave = 0.00")
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
