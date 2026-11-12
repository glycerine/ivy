package gia

import (
	"fmt"
	"slices"
	"time"
)

func Gia_ManTerStateHash(state []uint32, words, bins int32) int32 {
	if words < 0 || int(words) > len(state) || bins <= 0 {
		panic("gia: invalid ternary hash parameters")
	}
	var hash uint32
	for i := int32(0); i < words; i++ {
		hash ^= state[i] * giaEraPrimes[i&127]
	}
	return int32(hash % uint32(bins))
}
func Gia_ManTerStateLookup(state []uint32, words int32, bins [][][]uint32, count int32) []uint32 {
	for _, entry := range bins[Gia_ManTerStateHash(state, words, count)] {
		if slices.Equal(entry[:words], state[:words]) {
			return entry
		}
	}
	return nil
}
func Gia_ManTerStateInsert(state []uint32, words int32, bins [][][]uint32, count int32) {
	if Gia_ManTerStateLookup(state, words, bins, count) != nil {
		panic("gia: duplicate ternary state")
	}
	hash := Gia_ManTerStateHash(state, words, count)
	bins[hash] = append(bins[hash], state)
}
func Gia_ManTerStateAlloc(words int32) []uint32 { return make([]uint32, words) }
func Gia_ManTerStateCreate(p *Gia_ManTer_t) []uint32 {
	state := Gia_ManTerStateAlloc(p.nStateWords)
	pis := Gia_ManPiNum(p.pAig)
	for i := pis; i < Gia_ManCiNum(p.pAig); i++ {
		value := Gia_ManTerSimInfoGet(p.pDataSimCis, i)
		Gia_ManTerSimInfoSet(state, i-pis, value)
		if value == GIA_ZER {
			p.pCount0[i-pis]++
		}
		if value == GIA_UND {
			p.pCountX[i-pis]++
		}
	}
	Vec_PtrPush(p.vStates, state)
	return state
}
func Gia_ManTerRetire2(p *Gia_ManTer_t, state []uint32) int32 {
	maxUnknown := int32(-1)
	for i := int32(0); i < Gia_ManRegNum(p.pAig); i++ {
		if Gia_ManTerSimInfoGet(state, i) != GIA_UND && p.pRetired[i] == 0 && maxUnknown < p.pCountX[i] {
			maxUnknown = p.pCountX[i]
		}
	}
	if maxUnknown < 0 {
		panic("gia: no ternary register to retire")
	}
	for i := int32(0); i < Gia_ManRegNum(p.pAig); i++ {
		if Gia_ManTerSimInfoGet(state, i) != GIA_UND && p.pRetired[i] == 0 && p.pCountX[i] == maxUnknown {
			p.pRetired[i] = 1
			Vec_IntPush(p.vRetired, i)
			if maxUnknown == 0 {
				break
			}
		}
	}
	for _, id := range Vec_IntArray(p.vRetired) {
		Gia_ManTerSimInfoSet(p.pDataSimCis, Gia_ManPiNum(p.pAig)+id, GIA_UND)
	}
	return int32(Vec_IntSize(p.vRetired))
}
func Gia_ManTerRetire(p *Gia_ManTer_t, current, previous []uint32) int32 {
	Vec_IntClear(p.vRetired)
	for i := int32(0); i < Gia_ManRegNum(p.pAig); i++ {
		if Gia_ManTerSimInfoGet(current, i) != Gia_ManTerSimInfoGet(previous, i) {
			Vec_IntPush(p.vRetired, i)
		}
	}
	for _, id := range Vec_IntArray(p.vRetired) {
		Gia_ManTerSimInfoSet(p.pDataSimCis, Gia_ManPiNum(p.pAig)+id, GIA_UND)
	}
	return int32(Vec_IntSize(p.vRetired))
}
func Gia_ManTerSimulate(graph *Gia_Man_t, verbose int32) *Gia_ManTer_t {
	if Gia_ManRegNum(graph) <= 0 {
		panic("gia: ternary simulation requires registers")
	}
	start := time.Now()
	p := Gia_ManTerCreate(graph)
	Gia_ManTerSimInfoInit(p)
	state := Gia_ManTerStateCreate(p)
	Gia_ManTerStateInsert(state, p.nStateWords, p.pBins, p.nBins)
	var previous []uint32
	for i := int32(0); ; i++ {
		Gia_ManTerSimulateRound(p)
		Gia_ManTerSimInfoTransfer(p)
		state = Gia_ManTerStateCreate(p)
		if loop := Gia_ManTerStateLookup(state, p.nStateWords, p.pBins, p.nBins); loop != nil {
			graph.nTerStates = int32(Vec_PtrSize(p.vStates))
			graph.nTerLoop = int32(Vec_PtrFind(p.vStates, loop))
			if verbose != 0 {
				fmt.Printf("Ternary simulation saturated after %d iterations. Time: %s\n", i+1, time.Since(start))
			}
			break
		}
		Gia_ManTerStateInsert(state, p.nStateWords, p.pBins, p.nBins)
		if i >= p.nIters && i%10 == 0 {
			Gia_ManTerRetire(p, state, previous)
		}
		previous = state
	}
	return p
}
