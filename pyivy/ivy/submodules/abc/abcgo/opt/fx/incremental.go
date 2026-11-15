package fx

import (
	"encoding/binary"
	"fmt"
	"slices"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// extraction is the mutable matrix/queue in src/base/abci/abcFx.c. Cube
// indices stay stable until extraction finishes; empty rows are tombstones.
// The map only assigns insertion-order divisor IDs and is never iterated.
type extraction struct {
	cubes                                     [][]int32
	lits                                      [][]int
	first                                     []int
	levels                                    []int32
	divisors                                  [][]int32
	ids                                       map[string]int
	weights                                   []float32
	queue                                     *vec.Vec_Que_t
	marked                                    []bool
	contained                                 []int
	options                                   Options
	literalCount, selected, complementedMuxes int
	started                                   int64
}

func (p *extraction) level(cube []int32) (level int32) {
	for _, lit := range cube[1:] {
		level = max(level, p.levels[lit>>1])
	}
	return
}
func (p *extraction) divisorID(div []int32) int {
	key := make([]byte, 4*len(div))
	for i, lit := range div {
		binary.LittleEndian.PutUint32(key[4*i:], uint32(lit))
	}
	if id, ok := p.ids[string(key)]; ok {
		return id
	}
	id := len(p.divisors)
	p.ids[string(key)] = id
	p.divisors = append(p.divisors, slices.Clone(div))
	return id
}
func (p *extraction) weight(div []int32, single, remove, update bool, gain int) {
	id := p.divisorID(div)
	if id == len(p.weights) {
		if remove {
			panic("fx: removing unknown divisor")
		}
		var level int32
		for _, lit := range div {
			level = max(level, p.levels[lit>>2])
		}
		level = min(level, 800)
		initial := -float64(len(div)) + 0.9 - 0.0009*float64(level)
		if single {
			initial = -2 + 0.9 - 0.001*float64(level)
		}
		p.weights = append(p.weights, float32(initial))
	}
	if remove {
		gain = -gain
	}
	p.weights[id] += float32(gain)
	if update {
		if vec.Vec_QueIsMember(p.queue, id) != 0 {
			vec.Vec_QueUpdate(p.queue, id)
		} else if !remove {
			vec.Vec_QuePush(p.queue, id)
		}
	}
}
func (p *extraction) singles(index int, remove, update bool) {
	cube := p.cubes[index]
	for a := 1; a < len(cube); a++ {
		for b := a + 1; b < len(cube); b++ {
			p.weight([]int32{(cube[a] ^ 1) << 1, (cube[b]^1)<<1 | 1}, true, remove, update, 1)
		}
	}
}
func (p *extraction) doubles(first, pivot int, remove, update bool) error {
	for i := first; i < len(p.cubes); i++ {
		cube := p.cubes[i]
		if len(cube) == 0 || i == pivot {
			continue
		}
		if p.marked[i] && p.marked[pivot] && i > pivot {
			continue
		}
		if cube[0] != p.cubes[pivot][0] {
			break
		}
		div, base, err := CubeFreeDivisor(cube, p.cubes[pivot])
		if err != nil {
			return err
		}
		if base == -1 {
			if !remove {
				if len(cube) > len(p.cubes[pivot]) {
					p.contained = append(p.contained, i)
				} else {
					p.contained = append(p.contained, pivot)
				}
			}
			continue
		}
		if len(div) == 4 {
			phase := NormalizeDivisor(div)
			if phase == 1 {
				p.complementedMuxes++
			}
			if phase < 0 && p.options.CanonicalOnly {
				continue
			}
		}
		if p.options.LiteralLimit != 0 && p.options.LiteralLimit < int32(len(div)) {
			continue
		}
		if p.options.CanonicalOnly && len(div) == 3 {
			continue
		}
		p.weight(div, false, remove, update, base+len(div)-1)
	}
	return nil
}
func (p *extraction) compress(lit int32) {
	row := p.lits[lit][:0]
	for _, index := range p.lits[lit] {
		if len(p.cubes[index]) != 0 {
			row = append(row, index)
		}
	}
	p.lits[lit] = row
}

// removeCommon is Vec_IntTwoRemoveCommon: the shared entries leave both
// incidence lists and become the ordered single-cube occurrence list.
func removeCommon(a, b []int) ([]int, []int, []int) {
	var common []int
	i, j, na, nb := 0, 0, 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			common = append(common, a[i])
			i++
			j++
		} else if a[i] < b[j] {
			a[na] = a[i]
			na++
			i++
		} else {
			b[nb] = b[j]
			nb++
			j++
		}
	}
	na += copy(a[na:], a[i:])
	nb += copy(b[nb:], b[j:])
	return a[:na], b[:nb], common
}
func removeIndices(a, b []int) []int {
	i, j, n := 0, 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			i++
			j++
		} else if a[i] < b[j] {
			a[n] = a[i]
			n++
			i++
		} else {
			j++
		}
	}
	n += copy(a[n:], a[i:])
	return a[:n]
}
func (p *extraction) update(id int) error {
	p.selected++
	div := slices.Clone(p.divisors[id])
	a, b := DivisorPivots(div)
	var singles []int
	if len(div) == 2 {
		p.compress(a ^ 1)
		p.compress(b ^ 1)
		p.lits[a^1], p.lits[b^1], singles = removeCommon(p.lits[a^1], p.lits[b^1])
	}
	p.compress(a)
	p.compress(b)
	pairs, err := FindCommonPairs(p.cubes, p.lits[a], p.lits[b], div)
	if err != nil {
		return err
	}
	doubles := make([]int, 0, 2*len(pairs))
	for _, pair := range pairs {
		doubles = append(doubles, pair.First, pair.Second)
	}
	// Subtract old contributions before changing the matrix. Marking ensures
	// a pair of affected rows is counted only once, in the source's order.
	for _, group := range [][]int{singles, doubles} {
		for _, i := range group {
			p.singles(i, true, true)
		}
	}
	for _, group := range [][]int{singles, doubles} {
		for _, i := range group {
			p.marked[i] = true
		}
	}
	for _, group := range [][]int{singles, doubles} {
		for _, i := range group {
			if err := p.doubles(p.first[p.cubes[i][0]], i, true, true); err != nil {
				return err
			}
		}
	}
	for _, group := range [][]int{singles, doubles} {
		for _, i := range group {
			p.marked[i] = false
		}
	}
	newVar := int32(0)
	if a>>1 != b>>1 || len(div) != 2 {
		newVar = int32(len(p.first))
		p.first = append(p.first, len(p.cubes))
		var level int32
		if len(div) == 2 {
			cube := []int32{newVar, a ^ 1, b ^ 1}
			level = 1 + p.level(cube)
			p.cubes = append(p.cubes, cube)
		} else {
			left, right := AddDivisorLiterals([]int32{newVar}, []int32{newVar}, div)
			level = 2 + max(p.level(left), p.level(right))
			p.cubes = append(p.cubes, left, right)
		}
		p.levels = append(p.levels, level)
		for len(p.marked) < len(p.cubes) {
			p.marked = append(p.marked, false)
		}
		p.literalCount += len(div)
		p.lits = append(p.lits, nil, nil)
		for _, i := range singles {
			cube, n := RemoveDivisorLiterals(p.cubes[i], div, false)
			if n != 2 {
				return fmt.Errorf("fx: invalid single-cube occurrence")
			}
			p.literalCount--
			p.cubes[i] = append(cube, newVar<<1)
			p.lits[newVar<<1] = append(p.lits[newVar<<1], i)
		}
	}
	doubles = doubles[:0]
	complements := 0
	for _, pair := range pairs {
		var n, m int
		p.cubes[pair.First], n = RemoveDivisorLiterals(p.cubes[pair.First], div, pair.Complement)
		p.cubes[pair.Second], m = RemoveDivisorLiterals(p.cubes[pair.Second], div, pair.Complement)
		if n+m != len(div) && n+m != len(div)+1 {
			return fmt.Errorf("fx: invalid paired-cube occurrence")
		}
		if pair.Complement {
			complements++
		}
		if newVar > 0 {
			lit := newVar << 1
			if len(div) == 2 || pair.Complement {
				lit |= 1
			}
			p.cubes[pair.First] = append(p.cubes[pair.First], lit)
			p.lits[lit] = append(p.lits[lit], pair.First)
		}
		p.literalCount -= len(div) + len(p.cubes[pair.Second]) - 2
		doubles = append(doubles, pair.First)
		p.cubes[pair.Second] = nil
	}
	slices.Sort(doubles)
	for _, group := range [][]int{singles, doubles} {
		for _, i := range group {
			p.singles(i, false, true)
		}
	}
	for _, group := range [][]int{singles, doubles} {
		for _, i := range group {
			p.marked[i] = true
		}
	}
	for _, group := range [][]int{singles, doubles} {
		for _, i := range group {
			if err := p.doubles(p.first[p.cubes[i][0]], i, false, true); err != nil {
				return err
			}
		}
	}
	for _, group := range [][]int{singles, doubles} {
		for _, i := range group {
			p.marked[i] = false
		}
	}
	if len(p.contained) > 0 {
		slices.Sort(p.contained)
		p.contained = slices.Compact(p.contained)
		for _, i := range p.contained {
			// A paired rewrite may already have removed a recorded row.
			if len(p.cubes[i]) == 0 {
				continue
			}
			if err := p.doubles(p.first[p.cubes[i][0]], i, true, true); err != nil {
				return err
			}
			p.cubes[i] = nil
		}
		p.contained = p.contained[:0]
	}
	if len(div) > 2 {
		for i := len(p.cubes) - 2; i < len(p.cubes); i++ {
			p.singles(i, false, true)
			for _, lit := range p.cubes[i][1:] {
				p.lits[lit] = append(p.lits[lit], i)
			}
		}
	}
	for i, encoded := range div {
		lit := encoded >> 1
		p.lits[lit] = removeIndices(p.lits[lit], doubles)
		if complements > 0 && i > 1 || len(div) == 2 {
			p.lits[lit^1] = removeIndices(p.lits[lit^1], doubles)
		}
	}
	return nil
}

func (p *extraction) printDivisor(id int) {
	div := p.divisors[id]
	fmt.Printf("%4d : Div %7d : Weight %12.5f  ", p.selected, id, p.weights[id])
	for _, lit := range div {
		if lit&1 == 0 {
			fmt.Printf("%d(1)", lit>>1)
		}
	}
	fmt.Print(" + ")
	for _, lit := range div {
		if lit&1 != 0 {
			fmt.Printf("%d(2)", lit>>1)
		}
	}
	for i := len(div) + 3; i < 16; i++ {
		fmt.Print(" ")
	}
	fmt.Printf("Lits =%7d  Divs =%8d  ", p.literalCount, len(p.divisors))
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-p.started)
}
func (p *extraction) printStats(elapsed int64) {
	cubes, lits := 0, 0
	for _, row := range p.cubes {
		if len(row) > 0 {
			cubes++
		}
	}
	for _, row := range p.lits {
		if len(row) > 0 {
			lits++
		}
	}
	fmt.Printf("Cubes =%8d  Lits  =%8d  Divs  =%8d  Divs+ =%8d  Compl =%8d  Extr  =%7d  ", cubes, lits, len(p.divisors), vec.Vec_QueSize(p.queue), p.complementedMuxes, p.selected)
	util.Abc_PrintTime(1, "Time", elapsed)
}
