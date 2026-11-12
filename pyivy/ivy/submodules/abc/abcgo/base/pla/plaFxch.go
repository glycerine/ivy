package pla

/**CFile****************************************************************

  FileName    [plaFxch.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SOP manager.]

  Synopsis    [Scalable SOP transformations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - March 18, 2015.]

  Revision    [$Id: plaFxch.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type fxPair struct{ first, second, skipFirst, skipSecond int32 }
type fxDivisor struct {
	lits   [2]int32
	weight int32
	pairs  []fxPair
}

// Fxch_Man_t implements the experimental source's two-literal fast extraction.
// Candidate costs are recomputed from live cubes after an extraction, replacing
// the fragile packed hash links and incremental priority bookkeeping.
type Fxch_Man_t struct {
	Cubes      [][]int32
	Divs       []uint64
	candidates []fxDivisor
	index      map[[2]int32]int
}

func Fxch_ManStart(cubes, occurs [][]int32) *Fxch_Man_t {
	p := &Fxch_Man_t{Divs: make([]uint64, len(occurs)/2)}
	p.Cubes = make([][]int32, len(cubes))
	for i, c := range cubes {
		p.Cubes[i] = append(make([]int32, 0, len(c)), c...)
	}
	return p
}
func Fxch_ManStop(p *Fxch_Man_t) {
	if p != nil {
		*p = Fxch_Man_t{}
	}
}
func Fxch_DivisorAdd(p *Fxch_Man_t, a, b, weight int32) int32 {
	if a == b {
		panic("duplicate divisor literal")
	}
	if a > b {
		a, b = b, a
	}
	key := [2]int32{a, b}
	if p.index == nil {
		p.candidates = []fxDivisor{{weight: -1}}
		p.index = make(map[[2]int32]int)
	}
	i, ok := p.index[key]
	if !ok {
		i = len(p.candidates)
		p.index[key] = i
		p.candidates = append(p.candidates, fxDivisor{lits: key, weight: -2})
	}
	p.candidates[i].weight += weight
	return int32(i)
}
func Fxch_DivisorRemove(p *Fxch_Man_t, a, b, weight int32) {
	if a > b {
		a, b = b, a
	}
	i, ok := p.index[[2]int32{a, b}]
	if !ok {
		panic("divisor not found")
	}
	p.candidates[i].weight -= weight
}
func Fxch_TabSingleDivisors(p *Fxch_Man_t, cube, add int32) int32 {
	lits := p.Cubes[cube]
	for i, a := range lits {
		for _, b := range lits[i+1:] {
			if add != 0 {
				Fxch_DivisorAdd(p, a, b, 1)
			} else {
				Fxch_DivisorRemove(p, a, b, 1)
			}
		}
	}
	return int32(len(lits) * (len(lits) - 1) / 2)
}
func Fxch_ManCreateDivisors(p *Fxch_Man_t) {
	p.candidates = []fxDivisor{{weight: -1}}
	p.index = make(map[[2]int32]int)
	for i, c := range p.Cubes {
		if c != nil {
			Fxch_TabSingleDivisors(p, int32(i), 1)
		}
	}
	type entry struct{ cube, skip, lit int32 }
	groups := make(map[string][]entry)
	for i, c := range p.Cubes {
		for k, lit := range c {
			key := cubeKey(Vec_IntCopySkip(c, int32(k)))
			list := groups[key]
			paired := -1
			for j, e := range list {
				if e.cube != int32(i) && e.lit != lit {
					paired = j
					break
				}
			}
			if paired >= 0 {
				e := list[paired]
				list = append(list[:paired], list[paired+1:]...)
				groups[key] = list
				id := Fxch_DivisorAdd(p, e.lit^1, lit^1, int32(len(c)))
				p.candidates[id].pairs = append(p.candidates[id].pairs, fxPair{e.cube, int32(i), e.skip, int32(k)})
			} else {
				groups[key] = append(list, entry{int32(i), int32(k), lit})
			}
		}
	}
}
func containsLiteral(c []int32, lit int32) bool {
	for _, v := range c {
		if v == lit {
			return true
		}
	}
	return false
}
func Fxch_ManUpdate(p *Fxch_Man_t, id int32) {
	d := p.candidates[id]
	a, b := d.lits[0], d.lits[1]
	newVar := int32(len(p.Divs))
	p.Divs = append(p.Divs, uint64(uint32(a))|uint64(uint32(b))<<32)
	for i, c := range p.Cubes {
		if containsLiteral(c, a) && containsLiteral(c, b) {
			next := make([]int32, 0, len(c)-1)
			for _, lit := range c {
				if lit != a && lit != b {
					next = append(next, lit)
				}
			}
			p.Cubes[i] = append(next, newVar*2)
		}
	}
	for _, pair := range d.pairs {
		first, second := p.Cubes[pair.first], p.Cubes[pair.second]
		if first == nil || second == nil {
			continue
		}
		skipA, skipB := pair.skipFirst, pair.skipSecond
		if int(skipA) >= len(first) || int(skipB) >= len(second) {
			continue
		}
		x, y := first[skipA], second[skipB]
		if !((x == a^1 && y == b^1) || (x == b^1 && y == a^1)) {
			continue
		}
		restA, restB := Vec_IntCopySkip(first, skipA), Vec_IntCopySkip(second, skipB)
		if !equalInts(restA, restB) {
			continue
		}
		p.Cubes[pair.first] = append(restA, newVar*2+1)
		p.Cubes[pair.second] = nil
	}
	for _, c := range p.Cubes {
		sort.Slice(c, func(i, j int) bool { return c[i] < c[j] })
	}
}
func Fxch_ManFastExtract(p *Fxch_Man_t, verbose, veryVerbose int32) int32 {
	for {
		Fxch_ManCreateDivisors(p)
		best := -1
		for i, d := range p.candidates {
			if d.weight > 0 && (best < 0 || d.weight > p.candidates[best].weight) {
				best = i
			}
		}
		if best < 0 {
			break
		}
		Fxch_ManUpdate(p, int32(best))
	}
	return 1
}
func Fxch_ManWriteBlif(name string, cubes [][]int32, divs []uint64) error {
	original := 0
	for original < len(divs) && divs[original] == 0 {
		original++
	}
	var b strings.Builder
	fmt.Fprintf(&b, ".model %s\n.inputs", name)
	for i := 0; i < original; i++ {
		fmt.Fprintf(&b, " i%d", i)
	}
	b.WriteString("\n.outputs o\n.names")
	for i := range divs {
		fmt.Fprintf(&b, " i%d", i)
	}
	b.WriteString(" o\n")
	for _, cube := range cubes {
		if cube == nil {
			continue
		}
		line := make([]byte, len(divs))
		for i := range line {
			line[i] = '-'
		}
		for _, lit := range cube {
			line[lit/2] = '1' - byte(lit&1)
		}
		fmt.Fprintf(&b, "%s 1\n", line)
	}
	for i := original; i < len(divs); i++ {
		a, c := int32(divs[i]), int32(divs[i]>>32)
		fmt.Fprintf(&b, ".names i%d i%d i%d\n%d%d 1\n", a/2, c/2, i, 1-a%2, 1-c%2)
	}
	b.WriteString(".end\n")
	return os.WriteFile(name, []byte(b.String()), 0666)
}

// The source experimental entry point consumes temporary literal arrays and
// discards the extracted network (the BLIF write is commented out). Preserve
// that public behavior; callers wanting a result use Fxch_ManStart directly.
func Pla_ManPerformFxch(p *Pla_Man_t) int32 {
	Pla_ManConvertFromBits(p)
	fx := Fxch_ManStart(p.CubeLits, p.Occurs)
	p.CubeLits = nil
	p.Occurs = nil
	return Fxch_ManFastExtract(fx, 1, 0)
}
