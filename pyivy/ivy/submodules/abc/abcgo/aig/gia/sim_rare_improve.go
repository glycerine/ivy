package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManPatRareImprove(p *Gia_Man_t, limit, verbose int32) {
	start := util.Abc_Clock()
	words := giaPatternWords(p)
	patterns := p.VSimsPi
	p.VSimsPi = nil
	defer func() { p.VSimsPi = patterns }()
	initial := Gia_ManPatGetTotalQuo(p, limit, patterns, words)
	var rares, changes int
	for n := int32(0); n < 64*words; n++ {
		frameStart := util.Abc_Clock()
		rare := Gia_SimCollectRare(p, patterns, limit)
		pattern := Gia_ManPatCollectOne(p, patterns, n, words)
		impact := Gia_SimQualityImpact(p, pattern, rare)
		best := Gia_SimCollectBest(impact)
		if verbose != 0 {
			count := 0
			var change float32
			if best != nil {
				count = vec.Vec_IntSize(best)
				change = vec.Vec_FltEntry(impact, int(vec.Vec_IntEntry(best, 0)))
			}
			fmt.Printf("Pat %5d : Rare = %4d  Cands = %3d  Value = %8.3f  Change = %8.3f  ", n, vec.Vec_IntSize(rare)/2, count, Gia_ManPatGetOneQuo(p, limit, patterns, words, n), change)
			util.Abc_PrintTime(1, "Time", util.Abc_Clock()-frameStart)
		}
		if best != nil {
			id := vec.Vec_IntEntry(best, int(util.C_Rand())%vec.Vec_IntSize(best)) // rand() % Vec_IntSize(vBest)
			data := vec.Vec_WrdArray(patterns)
			giaSimFlip(data[id*words:(id+1)*words], n)
			changes++
		}
		rares = vec.Vec_IntSize(rare) / 2
	}
	if verbose != 0 {
		fmt.Println()
	}
	final := Gia_ManPatGetTotalQuo(p, limit, patterns, words)
	fmt.Printf("Improved %d out of %d patterns using %d rare nodes: %.2f -> %.2f.  ", changes, 64*words, rares, initial, final)
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-start)
}
