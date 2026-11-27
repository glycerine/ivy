package gia

import "fmt"

func Gia_ManCountMuxXor(p *Gia_Man_t, muxes, xors *int32) {
	*muxes, *xors = 0, 0
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsMuxType(obj) == 0 {
			continue
		}
		var a, b *Gia_Obj_t
		if Gia_ObjRecognizeExor(obj, &a, &b) != 0 {
			*xors++
		} else {
			*muxes++
		}
	}
}
func Gia_ManPrintMuxStats(p *Gia_Man_t) {
	var ands, muxes, xors, total int32
	if p.PMuxes != nil {
		ands = Gia_ManAndNotBufNum(p) - Gia_ManXorNum(p) - Gia_ManMuxNum(p)
		xors = Gia_ManXorNum(p)
		muxes = Gia_ManMuxNum(p)
		total = ands + 3*xors + 3*muxes
	} else {
		Gia_ManCountMuxXor(p, &muxes, &xors)
		total = Gia_ManAndNotBufNum(p)
		ands = total - 3*muxes - 3*xors
	}
	denominator := float64(max(total, 1))
	// giaMuxes.c's fields; the line continues in Gia_ManPrintStats
	fmt.Printf("stats:  ")
	fmt.Printf("xor =%8d %6.2f %%   ", xors, 300*float64(xors)/denominator)
	fmt.Printf("mux =%8d %6.2f %%   ", muxes, 300*float64(muxes)/denominator)
	fmt.Printf("and =%8d %6.2f %%   ", ands, 100*float64(ands)/denominator)
	fmt.Printf("obj =%8d  ", Gia_ManAndNotBufNum(p))
}
