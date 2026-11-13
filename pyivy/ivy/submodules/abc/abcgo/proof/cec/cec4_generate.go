package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// cec4GeneratePatterns makes and packs the source heuristic's partial PI
// assignments for pairs that currently share an equivalence class.
func cec4GeneratePatterns(graph *gia.Gia_Man_t, pack *cec4PatternPacker, candidates *cec4Candidates, maxIters int32) (int, bool) {
	if maxIters <= 0 {
		panic("cec4: invalid iteration limit")
	}
	pack.cursor = 0
	clear(pack.reserved.data)
	pattern, visited := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	nPats := 100 * 64 * pack.values.words
	count, tries := 0, 0
	for ; tries < nPats; tries++ {
		candidate := candidates.next()
		if candidate == 0 {
			continue
		}
		repr := gia.Gia_ObjRepr(graph, candidate)
		candidatePhase := gia.Gia_ObjPhase(gia.Gia_ManObj(graph, candidate))
		reprPhase := gia.Gia_ObjPhase(gia.Gia_ManObj(graph, repr))
		found := Cec4_ManGeneratePatternOne(graph, repr, reprPhase, candidate, candidatePhase^1, pattern, visited)
		if found == 0 {
			found = Cec4_ManGeneratePatternOne(graph, repr, reprPhase^1, candidate, candidatePhase, pattern, visited)
		}
		if found == 0 {
			continue
		}
		ret := pack.add(vec.Vec_IntArray(pattern), true)
		// C counts a packed pattern only after checking for the full-pack sentinel.
		if ret == 64*pack.values.words {
			break
		}
		count++
		if count == 8*64*pack.values.words {
			break
		}
	}
	return count, count >= tries/int(maxIters)
}
