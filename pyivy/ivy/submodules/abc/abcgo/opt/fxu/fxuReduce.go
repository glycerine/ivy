// Port of ABC/MVSIS FXU (MVSIS Group, UC Berkeley, 2003).
// Original algorithm and licensing: src/opt/fxu; see repository copyright notice.
package fxu

import "sort"

type cubePairRecord struct {
	a, b     *Fxu_Cube
	distance int
	order    int
}

func Fxu_CountPairDiffs(cover string, diffs []byte) int32 {
	width := 0
	if i := indexSpace(cover); i >= 0 {
		width = i
	}
	parsed, err := parseSOP(cover, width)
	if err != nil {
		return 0
	}
	at := 0
	for i, a := range parsed.patterns {
		for _, b := range parsed.patterns[i+1:] {
			n := 0
			for k := range a {
				if a[k] != b[k] {
					n++
				}
			}
			if at >= len(diffs) {
				panic("FXU pair distance buffer too short")
			}
			if n > 255 {
				n = 255
			}
			diffs[at] = byte(n)
			at++
		}
	}
	return 1
}
func indexSpace(s string) int {
	for i, c := range []byte(s) {
		if c == ' ' {
			return i
		}
		if c == '\n' {
			break
		}
	}
	return -1
}
func allCubePairs(p *Fxu_Matrix) []cubePairRecord {
	owners := make([]int, 0, len(p.covers))
	for owner := range p.covers {
		owners = append(owners, int(owner))
	}
	sort.Ints(owners)
	var pairs []cubePairRecord
	for _, owner := range owners {
		cubes := p.covers[int32(owner)]
		for i, a := range cubes {
			if a.deleted {
				continue
			}
			for _, b := range cubes[i+1:] {
				if b.deleted {
					continue
				}
				first, second, _ := pairParts(a.columns(), b.columns())
				if len(first) == 0 || len(second) == 0 {
					continue
				}
				pairs = append(pairs, cubePairRecord{a: a, b: b, distance: pairPatternDistance(a.columns(), b.columns()), order: len(pairs)})
			}
		}
	}
	return pairs
}
func pairPatternDistance(a, b []int32) int {
	states := make(map[int32]byte)
	for _, v := range a {
		states[v/2] |= 1 << uint(v&1)
	}
	for _, v := range b {
		states[v/2] ^= 1 << uint(2+(v&1))
	}
	n := 0
	for _, s := range states {
		if s&3 != s>>2 {
			n++
		}
	}
	return n
}
func rebuildDoubles(p *Fxu_Matrix, initial bool) {
	p.doubleHeap.clear()
	p.doubleCatalog = make(map[string]*Fxu_Double)
	p.Doubles = nil
	pairs := allCubePairs(p)
	if initial && int32(len(pairs)) > p.params.PairsMax {
		sort.SliceStable(pairs, func(i, j int) bool { return pairs[i].distance < pairs[j].distance })
		pairs = pairs[:p.params.PairsMax]
		sort.Slice(pairs, func(i, j int) bool { return pairs[i].order < pairs[j].order })
	}
	for _, pair := range pairs {
		Fxu_MatrixAddDivisor(p, pair.a, pair.b)
	}
}
func Fxu_PreprocessCubePairs(p *Fxu_Matrix, covers []string, total, maximum int32) int32 {
	if maximum < 0 {
		return 0
	}
	old := p.params.PairsMax
	p.params.PairsMax = maximum
	rebuildDoubles(p, true)
	p.params.PairsMax = old
	return 1
}
