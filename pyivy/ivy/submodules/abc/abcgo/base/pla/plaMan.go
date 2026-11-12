package pla

/**CFile****************************************************************

  FileName    [plaMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SOP manager.]

  Synopsis    [Scalable SOP transformations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - March 18, 2015.]

  Revision    [$Id: plaMan.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

import (
	"fmt"
	"math/bits"
	"sort"
	"sync"
)

var randomState = struct {
	sync.Mutex
	z, w uint32
}{z: 3716960521, w: 2174103536}

func randomStep(reset bool) uint32 {
	if reset {
		randomState.z = 3716960521
		randomState.w = 2174103536
	}
	randomState.z = 36969*(randomState.z&65535) + (randomState.z >> 16)
	randomState.w = 18000*(randomState.w&65535) + (randomState.w >> 16)
	return randomState.z<<16 + randomState.w
}
func Pla_SeedRandom(seed int32) {
	if seed < 0 {
		panic("negative seed")
	}
	randomState.Lock()
	defer randomState.Unlock()
	randomStep(true)
	for i := int32(0); i < seed; i++ {
		randomStep(false)
	}
}
func Pla_ManPrimesTable(n int32) []bool {
	if n < 1 || n > 30 {
		panic("prime table bit width must be 1..30")
	}
	table := make([]bool, 1<<uint(n))
	for i := 2; i < len(table); i++ {
		table[i] = true
	}
	for v := 2; v < len(table); v++ {
		if table[v] {
			for k := v * 2; k < len(table); k += v {
				table[k] = false
			}
		}
	}
	return table
}
func Pla_GenPrimes(n int32) []int32 {
	var result []int32
	for i, p := range Pla_ManPrimesTable(n) {
		if p {
			result = append(result, int32(i))
		}
	}
	return result
}
func Pla_GenFromMinterms(name string, minterms []int32, n int32) *Pla_Man_t {
	p := Pla_ManAlloc(name, n, 1, int32(len(minterms)))
	for i, mint := range minterms {
		for v := int32(0); v < n; v++ {
			lit := PLA_LIT_ZERO
			if mint&(1<<uint(v)) != 0 {
				lit = PLA_LIT_ONE
			}
			Pla_CubeSetLit(p.Cubes[i].In, v, lit)
		}
		Pla_CubeSetLit(p.Cubes[i].Out, 0, PLA_LIT_ONE)
	}
	return p
}
func Pla_ManPrimesDetector(n int32) *Pla_Man_t {
	return Pla_GenFromMinterms(fmt.Sprintf("primes%02d", n), Pla_GenPrimes(n), n)
}
func genRandomLocked(n, count int32, nonzero bool) []int32 {
	if n < 1 || n > 30 || count < 0 || int64(count) > int64(1)<<uint(n)-int64(boolInt(nonzero)) {
		panic("invalid random PLA dimensions")
	}
	chosen := make(map[int32]bool, count)
	mask := uint32(1<<uint(n)) - 1
	for len(chosen) < int(count) {
		v := int32(randomStep(false) & mask)
		if nonzero && v == 0 {
			continue
		}
		chosen[v] = true
	}
	result := make([]int32, 0, len(chosen))
	for v := range chosen {
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
func Pla_GenRandom(n, count, nonzero int32) []bool {
	randomState.Lock()
	defer randomState.Unlock()
	result := make([]bool, 1<<uint(n))
	for _, v := range genRandomLocked(n, count, nonzero != 0) {
		result[v] = true
	}
	return result
}
func Pla_ManGenerate(ins, outs, cubes, verbose int32) *Pla_Man_t {
	if outs < 1 {
		panic("positive output count required")
	}
	randomState.Lock()
	defer randomState.Unlock()
	minterms := genRandomLocked(ins, cubes, false)
	p := Pla_ManAlloc(fmt.Sprintf("rand_%d_%d_%d", ins, outs, cubes), ins, outs, cubes)
	for i, m := range minterms {
		for v := int32(0); v < ins; v++ {
			lit := PLA_LIT_ZERO
			if m&(1<<uint(v)) != 0 {
				lit = PLA_LIT_ONE
			}
			Pla_CubeSetLit(p.Cubes[i].In, v, lit)
		}
	}
	if outs == 1 {
		for i := range p.Cubes {
			Pla_CubeSetLit(p.Cubes[i].Out, 0, PLA_LIT_ONE)
		}
	} else {
		for i, m := range genRandomLocked(outs, cubes, true) {
			for v := int32(0); v < outs; v++ {
				lit := PLA_LIT_ZERO
				if m&(1<<uint(v)) != 0 {
					lit = PLA_LIT_ONE
				}
				Pla_CubeSetLit(p.Cubes[i].Out, v, lit)
			}
		}
	}
	return p
}
func Pla_ManConvertFromBits(p *Pla_Man_t) {
	p.CubeLits = make([][]int32, len(p.Cubes))
	p.Occurs = make([][]int32, 2*p.Ins)
	for i, c := range p.Cubes {
		for v := int32(0); v < p.Ins; v++ {
			lit := Pla_CubeGetLit(c.In, v)
			if lit == PLA_LIT_DASH {
				continue
			}
			if lit == PLA_LIT_FULL {
				for phase := int32(0); phase < 2; phase++ {
					encoded := v*2 + phase
					p.CubeLits[i] = append(p.CubeLits[i], encoded)
					p.Occurs[encoded] = append(p.Occurs[encoded], int32(i))
				}
				continue
			}
			encoded := v*2 + boolInt(lit == PLA_LIT_ZERO)
			p.CubeLits[i] = append(p.CubeLits[i], encoded)
			p.Occurs[encoded] = append(p.Occurs[encoded], int32(i))
		}
	}
}
func Pla_ManConvertToBits(p *Pla_Man_t) {
	next := make([]Cube, len(p.CubeLits))
	for i, lits := range p.CubeLits {
		next[i] = Cube{In: make([]uint64, words(p.Ins)), Out: make([]uint64, words(p.Outs))}
		if i < len(p.Cubes) {
			copy(next[i].Out, p.Cubes[i].Out)
		}
		for _, lit := range lits {
			if lit < 0 || lit/2 >= p.Ins {
				panic("cube literal outside PLA input width")
			}
			value := PLA_LIT_ONE
			if lit&1 != 0 {
				value = PLA_LIT_ZERO
			}
			Pla_CubeSetLit(next[i].In, lit/2, value)
		}
	}
	p.Cubes = next
}
func Pla_ManDist1Num(p *Pla_Man_t) int32 {
	count := int32(0)
	for i, a := range p.Cubes {
		for _, b := range p.Cubes[i+1:] {
			count += Pla_CubesAreDistance1(a.In, b.In, int32(words(p.Ins)))
		}
	}
	return count
}
func Pla_ManDist1NumTest(p *Pla_Man_t) int32 {
	fmt.Printf("Found %d pairs among %d cubes using cube pair enumeration.\n", Pla_ManDist1Num(p), len(p.Cubes))
	return 1
}
func Pla_GenSorter(n int32) error {
	if n < 1 || n > 30 {
		return fmt.Errorf("sorter width must be 1..30")
	}
	p := Pla_ManAlloc(fmt.Sprintf("sorter%02d", n), n, n, (1<<uint(n))-1)
	for m := int32(1); m < 1<<uint(n); m++ {
		for k := int32(0); k < n; k++ {
			lit := PLA_LIT_ZERO
			if m&(1<<uint(n-1-k)) != 0 {
				lit = PLA_LIT_ONE
			}
			Pla_CubeSetLit(p.Cubes[m-1].In, k, lit)
			out := PLA_LIT_ZERO
			if k < int32(bits.OnesCount32(uint32(m))) {
				out = PLA_LIT_ONE
			}
			Pla_CubeSetLit(p.Cubes[m-1].Out, k, out)
		}
	}
	return Pla_WritePla(p, fmt.Sprintf("sorter%02d.pla", n))
}
