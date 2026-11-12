// Port of ABC FXCH, Bruno Schmitt, UFRGS (2016); original FxchMan.c.
package fxch

import (
	"fmt"
	"io"
	"os"
	"sort"
)

func (p *Fxch_Man_t) writer() io.Writer {
	if p.Log != nil {
		return p.Log
	}
	return os.Stdout
}
func Fxch_ManAlloc(c Cubes) *Fxch_Man_t {
	return &Fxch_Man_t{Cubes: c, divIndex: make(map[string]int32)}
}
func Fxch_ManFree(p *Fxch_Man_t) {
	p.Levels = nil
	p.LitHashKeys = nil
	p.table = nil
	p.divisors = nil
	p.divIndex = nil
	p.outputs = nil
}
func Fxch_ManMapLiteralsIntoCubes(p *Fxch_Man_t, n int32) { p.Levels = make([]int32, n) }
func Fxch_ManGenerateLitHashKeys(p *Fxch_Man_t) {
	p.LitHashKeys = make([]uint32, 2*len(p.Levels))
	p.rngZ, p.rngW = 3716960521, 2174103536
	p.randomKey() // Gia_ManRandom(1) advances while resetting.
	for i := range p.LitHashKeys {
		p.LitHashKeys[i] = p.randomKey()
	}

}
func Fxch_ManSCHashTablesInit(p *Fxch_Man_t) {
	p.table = Fxch_SCHashTableCreate(p, int32(len(p.Cubes)))
}
func Fxch_ManSCHashTablesFree(p *Fxch_Man_t) {
	if p.table != nil {
		Fxch_SCHashTableDelete(p.table)
	}
	p.table = nil
}
func Fxch_ManComputeLevelDiv(p *Fxch_Man_t, d []int32) int32 {
	var level int32
	for _, l := range d {
		if p.Levels[l>>2] > level {
			level = p.Levels[l>>2]
		}
	}
	if level > 800 {
		return 800
	}
	return level
}
func Fxch_ManComputeLevelCube(p *Fxch_Man_t, c []int32) int32 {
	var level int32
	for _, l := range c[1:] {
		if p.Levels[l>>1] > level {
			level = p.Levels[l>>1]
		}
	}
	return level
}
func Fxch_ManComputeLevel(p *Fxch_Man_t) {
	p.Levels = make([]int32, len(p.Levels))
	if len(p.Cubes) == 0 {
		return
	}
	owner := p.Cubes[0][0]
	first := 0
	for i, c := range p.Cubes {
		if owner != c[0] {
			p.Levels[owner] += int32(i - first)
			owner = c[0]
			first = i
		}
		level := Fxch_ManComputeLevelCube(p, c)
		if level > p.Levels[owner] {
			p.Levels[owner] = level
		}
	}
}
func Fxch_ManDivSingleCube(p *Fxch_Man_t, id, add, update int32) int32 {
	c := p.Cubes[id]
	if len(c) < 2 {
		return 0
	}
	n := int32(len(p.outputs[id]))
	if n == 0 {
		n = 1
	}
	for i := 1; i < len(c); i++ {
		for j := i + 1; j < len(c); j++ {
			p.CubeFree = []int32{(c[i] ^ 1) * 2, (c[j]^1)*2 + 1}
			for z := int32(0); z < n; z++ {
				if add != 0 {
					Fxch_DivAdd(p, update, 1, 0)
				} else {
					Fxch_DivRemove(p, update, 1, 0)
				}
			}
		}
	}
	return int32((len(c) - 1) * (len(c) - 2) / 2)
}
func Fxch_ManDivDoubleCube(p *Fxch_Man_t, id, add, update int32) {
	c := p.Cubes[id]
	if len(c) == 0 {
		return
	}
	var hash uint32
	for _, l := range c[1:] {
		hash += p.LitHashKeys[l]
	}
	apply := func(h uint32, a, b int) {
		if add != 0 {
			Fxch_SCHashTableInsert(p.table, p.Cubes, h, uint32(id), uint32(a), uint32(b), byte(update))
		} else {
			Fxch_SCHashTableRemove(p.table, p.Cubes, h, uint32(id), uint32(a), uint32(b), byte(update))
		}
	}
	apply(hash, 0, 0)
	for i := 1; i < len(c); i++ {
		h := hash - p.LitHashKeys[c[i]]
		apply(h, i, 0)
		for j := i + 1; j < len(c); j++ {
			apply(h-p.LitHashKeys[c[j]], i, j)
		}
	}
}
func Fxch_ManDivCreate(p *Fxch_Man_t) {
	p.divisors = nil
	p.divIndex = make(map[string]int32)
	Fxch_ManSCHashTablesInit(p)
	for i, c := range p.Cubes {
		if len(c) == 0 {
			continue
		}
		Fxch_ManDivSingleCube(p, int32(i), 1, 0)
		if c[0] == 0 {
			Fxch_ManDivDoubleCube(p, int32(i), 1, 0)
		}
	}
}
func has(c []int32, l int32) bool {
	for _, v := range c[1:] {
		if v == l {
			return true
		}
	}
	return false
}
func sameSet(a, b map[int32]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}
func cloneSet(a map[int32]bool) map[int32]bool {
	b := make(map[int32]bool, len(a))
	for k := range a {
		b[k] = true
	}
	return b
}
func (p *Fxch_Man_t) addCube(c []int32, outputs map[int32]bool) int {
	p.Cubes = append(p.Cubes, c)
	p.outputs = append(p.outputs, outputs)
	return len(p.Cubes) - 1
}
func Fxch_ManUpdate(p *Fxch_Man_t, id int32) {
	div := p.divisors[id]
	p.Div = copyInts(div.literals)
	p.singles = nil
	if len(p.Div) == 2 {
		a, b := (p.Div[0]>>1)^1, (p.Div[1]>>1)^1
		for i, c := range p.Cubes {
			if len(c) > 0 && has(c, a) && has(c, b) {
				p.singles = append(p.singles, int32(i))
			}
		}
	}
	pairs := append([][2]int32(nil), div.pairs...)
	for i, pair := range pairs {
		if pair[0] > pair[1] {
			pairs[i] = [2]int32{pair[1], pair[0]}
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i][0] != pairs[j][0] {
			return pairs[i][0] < pairs[j][0]
		}
		return pairs[i][1] < pairs[j][1]
	})
	newVar := int32(0)
	if Fxch_DivIsNotConstant1(p.Div) != 0 {
		newVar = int32(len(p.Levels))
		p.LitHashKeys = append(p.LitHashKeys, p.randomKey(), p.randomKey())
		a := []int32{newVar}
		level := int32(0)
		if len(p.Div) == 2 {
			a = append(a, (p.Div[0]>>1)^1, (p.Div[1]>>1)^1)
			sort.Slice(a[1:], func(i, j int) bool { return a[i+1] < a[j+1] })
			level = 1 + Fxch_ManComputeLevelCube(p, a)
			p.addCube(a, nil)
		} else {
			b := []int32{newVar}
			Fxch_DivSepareteCubes(p.Div, &a, &b)
			level = Fxch_ManComputeLevelCube(p, a)
			if x := Fxch_ManComputeLevelCube(p, b); x > level {
				level = x
			}
			level += 2
			p.addCube(a, nil)
			p.addCube(b, nil)
		}
		p.Levels = append(p.Levels, level)
	}
	for _, i := range p.singles {
		c := &p.Cubes[i]
		removeLit(c, (p.Div[0]>>1)^1)
		removeLit(c, (p.Div[1]>>1)^1)
		*c = append(*c, newVar*2)
	}
	var previous [2]int32
	seen := false
	for _, pair := range pairs {
		if seen && pair == previous {
			continue
		}
		previous = pair
		seen = true
		i, j := pair[0], pair[1]
		if len(p.Cubes[i]) == 0 || len(p.Cubes[j]) == 0 {
			continue
		}
		a, b := copyInts(p.Cubes[i]), copyInts(p.Cubes[j])
		compl := int32(0)
		if Fxch_DivRemoveLits(&a, &b, p.Div, &compl) != int32(len(p.Div)) {
			continue
		}
		common := make(map[int32]bool)
		for o := range p.outputs[i] {
			if p.outputs[j][o] {
				common[o] = true
			}
		}
		if len(common) == 0 {
			continue
		}
		if newVar != 0 {
			phase := int32(0)
			if len(p.Div) == 2 || compl != 0 {
				phase = 1
			}
			a = append(a, newVar*2+phase)
		}
		if sameSet(p.outputs[i], p.outputs[j]) {
			p.Cubes[i] = a
			p.Cubes[j] = nil
			p.outputs[j] = nil
		} else {
			for o := range common {
				delete(p.outputs[i], o)
				delete(p.outputs[j], o)
			}
			if len(p.outputs[i]) == 0 {
				p.Cubes[i] = nil
			}
			if len(p.outputs[j]) == 0 {
				p.Cubes[j] = nil
			}
			p.addCube(a, common)
		}
	}
	p.Extracted++
	compactSCC(p)
	Fxch_ManDivCreate(p)
}

// SCC only removes outputs shared with a subsuming cube, retaining unrelated outputs.
func compactSCC(p *Fxch_Man_t) {
	for i, a := range p.Cubes {
		if len(a) == 0 || a[0] != 0 {
			continue
		}
		for j, b := range p.Cubes {
			if i == j || len(b) == 0 || b[0] != 0 || len(b) > len(a) {
				continue
			}
			sub := true
			for _, l := range b[1:] {
				if !has(a, l) {
					sub = false
					break
				}
			}
			if !sub || len(a) == len(b) && i < j {
				continue
			}
			for o := range p.outputs[j] {
				delete(p.outputs[i], o)
			}
			if len(p.outputs[i]) == 0 {
				p.Cubes[i] = nil
				break
			}
		}
	}
}
func Fxch_ManPrintDivs(p *Fxch_Man_t) {
	for i := range p.divisors {
		Fxch_DivPrint(p, int32(i))
	}
}
func Fxch_ManPrintStats(p *Fxch_Man_t) {
	c, l := 0, 0
	for _, a := range p.Cubes {
		if len(a) > 0 {
			c++
			l += len(a) - 1
		}
	}
	fmt.Fprintf(p.writer(), "Cubes =%8d Lits =%8d Divs =%8d Extr =%7d\n", c, l, len(p.divisors), p.Extracted)
}

func (p *Fxch_Man_t) randomKey() uint32 {
	p.rngZ = 36969*(p.rngZ&65535) + (p.rngZ >> 16)
	p.rngW = 18000*(p.rngW&65535) + (p.rngW >> 16)
	return ((p.rngZ << 16) + p.rngW) & 0x3ffffff
}
