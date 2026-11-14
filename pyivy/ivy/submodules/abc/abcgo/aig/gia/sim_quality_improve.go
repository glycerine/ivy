package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand"
	"time"
)

func Gia_SimQualityPatternsMax(p *Gia_Man_t, pattern *Vec_Int_t, iteration, verbose int32, stats *Vec_Int_t) int32 {
	initial := Gia_SimComputeScore(p, stats, Gia_SimQualityOne(p, pattern, 0))
	best := initial
	var tries []int32
	data := vec.Vec_IntArray(pattern)
	for k := range data {
		data[k] ^= 1
		value := Gia_SimComputeScore(p, stats, Gia_SimQualityOne(p, pattern, 0))
		if best <= value {
			if best < value {
				tries = tries[:0]
			}
			tries = append(tries, int32(k))
			best = value
		}
		data[k] ^= 1
	}
	selected := int32(-1)
	if len(tries) != 0 {
		selected = tries[rand.Intn(len(tries))]
	}
	if verbose != 0 {
		fmt.Printf("Iter %5d : Init = %6.3f  Final = %6.3f  Ratio = %4.2f  Tries = %5d  Choosing %5d\r", iteration, initial, best, best/initial, len(tries), selected)
	}
	return selected
}
func Gia_ManPatDistImprove(p *Gia_Man_t, verbose int32) {
	words := giaPatternWords(p)
	stats := Gia_SimGenerateStats(p)
	patterns := p.VSimsPi
	p.VSimsPi = nil
	defer func() { p.VSimsPi = patterns }()
	util.Abc_Random(1)
	var initialTotal, finalTotal float64
	for n := int32(0); n < 64*words; n++ {
		start := time.Now()
		pattern := Gia_ManPatCollectOne(p, patterns, n, words)
		initial := Gia_SimComputeScore(p, stats, Gia_SimQualityOne(p, pattern, 0))
		initialTotal += initial
		var k int32
		for k = 0; k < 100; k++ {
			index := Gia_SimQualityPatternsMax(p, pattern, k, verbose, stats)
			if index < 0 {
				break
			}
			data := vec.Vec_IntArray(pattern)
			data[index] ^= 1
		}
		final := Gia_SimComputeScore(p, stats, Gia_SimQualityOne(p, pattern, 0))
		finalTotal += final
		if verbose != 0 {
			fmt.Printf("Pat %5d : Tries = %5d  InitValue = %6.3f  FinalValue = %6.3f  Ratio = %4.2f  Time: %s\n", n, k, initial, final, final/initial, time.Since(start))
		}
		Gia_ManPatUpdateOne(p, patterns, n, words, pattern)
	}
	if verbose != 0 {
		fmt.Println()
	}
	fmt.Printf("Improved %d patterns with average init value %.2f and average final value %.2f.\n", 64*words, initialTotal/float64(64*words), finalTotal/float64(64*words))
}
