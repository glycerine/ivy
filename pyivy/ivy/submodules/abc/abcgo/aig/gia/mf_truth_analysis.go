package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
	"sync"
)

var mfTruthStats struct {
	sync.Mutex
	calls  int32
	memory *vec.Vec_Mem_t
}

func Mf_ManTruthCanonicize(t *word, nVars int32) int32 {
	best := *t
	var config int32
	if candidate := ^best; best > candidate {
		best = candidate
		config ^= 1 << nVars
	}
	for i := int32(0); i < nVars; i++ {
		if candidate := util.Abc_Tt6Flip(best, i); best > candidate {
			best = candidate
			config ^= 1 << i
		}
	}
	*t = best
	mfTruthStats.Lock()
	defer mfTruthStats.Unlock()
	if mfTruthStats.memory == nil {
		mfTruthStats.memory = vec.Vec_MemAllocForTT(6, 0)
	}
	vec.Vec_MemHashInsert(mfTruthStats.memory, []uint64{best})
	mfTruthStats.calls++
	return config
}
func Mf_ManTruthQuit() {
	mfTruthStats.Lock()
	defer mfTruthStats.Unlock()
	if mfTruthStats.memory == nil {
		return
	}
	fmt.Printf("TT = %d (%.2f %%)\n", vec.Vec_MemEntryNum(mfTruthStats.memory), 100*float64(vec.Vec_MemEntryNum(mfTruthStats.memory))/float64(mfTruthStats.calls))
	vec.Vec_MemHashFree(mfTruthStats.memory)
	vec.Vec_MemFree(mfTruthStats.memory)
	mfTruthStats.memory = nil
	mfTruthStats.calls = 0
}
func Mf_ManTruthCollect(limit int32) *Vec_Wrd_t {
	perm, comp := extra.Extra_PermSchedule(6), extra.Extra_GreyCodeSchedule(6)
	originals := 0
	seen := make(map[uint64]struct{})
	for _, class := range mfTruthClasses {
		if class.clauses > limit {
			continue
		}
		originals++
		for phase := 0; phase < 2; phase++ {
			current := class.truth
			if phase != 0 {
				current = ^current
			}
			start := current
			for _, swap := range perm {
				old := current
				for _, flip := range comp {
					current = util.Abc_Tt6Flip(current, flip)
					seen[current] = struct{}{}
				}
				if old != current {
					panic("gia: invalid Gray schedule")
				}
				current = util.Abc_Tt6SwapAdjacent(current, swap)
			}
			if start != current {
				panic("gia: invalid permutation schedule")
			}
		}
	}
	canonical := make(map[uint64]struct{}, len(seen))
	for truth := range seen {
		Mf_ManTruthCanonicize(&truth, 6)
		canonical[truth] = struct{}{}
	}
	result := vec.Vec_WrdAlloc(len(canonical))
	for truth := range canonical {
		vec.Vec_WrdPush(result, truth)
	}
	fmt.Printf("Original = %d.  Total = %d.  Unique = %d.  Unique = %d.  \n", originals, originals*2*720*64, len(seen), len(canonical))
	return result
}
func Mf_ManTruthCount() int32 { return int32(vec.Vec_WrdSize(Mf_ManTruthCollect(10))) }
func Mf_ManProfileTruths(p *Mf_Man_t) {
	counts := make([]int32, vec.Vec_IntSize(&p.vCnfSizes))
	for _, obj := range p.pGia.pObjs {
		if Gia_ObjIsAnd(obj) != 0 && Mf_ObjMapRefNum(p, obj.id) != 0 {
			counts[uint32(Mf_CutFunc(Mf_ObjCutBest(p, obj.id)))>>1]++
		}
	}
	counter := 0
	for i, n := range counts {
		if n == 0 {
			continue
		}
		fmt.Printf("%6d : %6d : Occur = %4d  CNF size = %2d  ", counter, i, n, vec.Vec_IntEntry(&p.vCnfSizes, i))
		counter++
		dau.Dau_DsdPrintFromTruth(vec.Vec_MemReadEntry(p.vTtMem, i), p.pPars.NLutSize)
	}
}
