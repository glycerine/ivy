package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand"
	"time"
)

func Gia_ManPatRareImprove(p *Gia_Man_t, limit, verbose int32) {
	start := time.Now()
	words := giaPatternWords(p)
	patterns := p.vSimsPi
	p.vSimsPi = nil
	defer func() { p.vSimsPi = patterns }()
	initial := Gia_ManPatGetTotalQuo(p, limit, patterns, words)
	var rares, changes int
	for n := int32(0); n < 64*words; n++ {
		frameStart := time.Now()
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
			fmt.Printf("Pat %5d : Rare = %4d  Cands = %3d  Value = %8.3f  Change = %8.3f  Time: %s\n", n, vec.Vec_IntSize(rare)/2, count, Gia_ManPatGetOneQuo(p, limit, patterns, words, n), change, time.Since(frameStart))
		}
		if best != nil {
			id := vec.Vec_IntEntry(best, rand.Intn(vec.Vec_IntSize(best)))
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
	fmt.Printf("Improved %d out of %d patterns using %d rare nodes: %.2f -> %.2f.  Time: %s\n", changes, 64*words, rares, initial, final, time.Since(start))
}
