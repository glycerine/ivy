package pla

/**CFile****************************************************************

  FileName    [plaSimple.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SOP manager.]

  Synopsis    [Scalable SOP transformations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - March 18, 2015.]

  Revision    [$Id: plaSimple.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

import (
	"fmt"
	"os"
	"strings"
)

func Pla_ManExpendDirNum(on []uint64, n, mint int32, vars []int32) int32 {
	count := int32(0)
	for v := int32(0); v < n; v++ {
		if Pla_TtGetBit(on, mint^(1<<uint(v))) != 0 {
			vars[count] = v
			count++
		}
	}
	return count
}
func Pla_PrintBinary(table []uint64, n int32) string {
	var b strings.Builder
	for v := int32(0); v < n; v++ {
		b.WriteByte('0' + byte(Pla_TtGetBit(table, v)))
	}
	b.WriteByte('\n')
	return b.String()
}

// FxMinimize retains the source's two-pass isolated-minterm/edge cover. Its
// previously asserted square case now records the valid four-minterm cube.
func Pla_ManFxMinimize(on []uint64, n int32) []uint64 {
	if n < 0 || n > 30 {
		panic("packed minimizer needs at most 30 variables")
	}
	covered := make([]uint64, (uint64(1)<<uint(n)+63)/64)
	type item struct {
		mint int32
		drop uint32
	}
	var store []item
	size := int32(1) << uint(n)
	dirs := make([]int32, n)
	for mint := int32(0); mint < size; mint++ {
		if Pla_TtGetBit(on, mint) == 0 || Pla_TtGetBit(covered, mint) != 0 {
			continue
		}
		count := Pla_ManExpendDirNum(on, n, mint, dirs)
		switch {
		case count == 0:
			Pla_TtSetBit(covered, mint)
			store = append(store, item{mint, 0})
		case count == 1:
			Pla_TtSetBit(covered, mint)
			Pla_TtSetBit(covered, mint^(1<<uint(dirs[0])))
			store = append(store, item{mint, 1 << uint(dirs[0])})
		case count == 2 && Pla_TtGetBit(on, mint^(1<<uint(dirs[0]))^(1<<uint(dirs[1]))) != 0:
			drop := uint32(1<<uint(dirs[0]) | 1<<uint(dirs[1]))
			for subset := drop; ; subset = (subset - 1) & drop {
				Pla_TtSetBit(covered, mint^int32(subset))
				if subset == 0 {
					break
				}
			}
			store = append(store, item{mint, drop})
		}
	}
	for mint := int32(0); mint < size; mint++ {
		if Pla_TtGetBit(on, mint) == 0 || Pla_TtGetBit(covered, mint) != 0 {
			continue
		}
		count := Pla_ManExpendDirNum(on, n, mint, dirs)
		if count == 0 {
			panic("uncovered isolated minterm")
		}
		chosen := dirs[0]
		for _, v := range dirs[:count] {
			if Pla_TtGetBit(covered, mint^(1<<uint(v))) == 0 {
				chosen = v
				break
			}
		}
		Pla_TtSetBit(covered, mint)
		Pla_TtSetBit(covered, mint^(1<<uint(chosen)))
		store = append(store, item{mint, 1 << uint(chosen)})
	}
	result := make([]uint64, len(store))
	for i, item := range store {
		for v := int32(0); v < n; v++ {
			if item.drop&(1<<uint(v)) != 0 {
				continue
			}
			lit := PLA_LIT_ZERO
			if item.mint&(1<<uint(v)) != 0 {
				lit = PLA_LIT_ONE
			}
			Pla_CubeSetLit(result[i:i+1], v, lit)
		}
	}
	return result
}
func Pla_ManFxPrepare(n int32) *Pla_Man_t {
	table := Pla_ManPrimesTable(n)
	on := make([]uint64, (len(table)+63)/64)
	for i, value := range table {
		if value {
			Pla_TtSetBit(on, int32(i))
		}
	}
	cubes := Pla_ManFxMinimize(on, n)
	p := Pla_ManAlloc(fmt.Sprintf("primes%02d", n), n, 1, int32(len(cubes)))
	for i, c := range cubes {
		p.Cubes[i].In[0] = c
		Pla_CubeSetLit(p.Cubes[i].Out, 0, PLA_LIT_ONE)
	}
	Pla_ManConvertFromBits(p)
	return p
}
func Pla_ManDumpPla(p *Pla_Man_t, name string) error {
	if p.CubeLits == nil {
		Pla_ManConvertFromBits(p)
	}
	q := Pla_ManAlloc(p.Name, p.Ins, 1, int32(len(p.CubeLits)))
	q.CubeLits = p.CubeLits
	for i := range q.Cubes {
		Pla_CubeSetLit(q.Cubes[i].Out, 0, PLA_LIT_ONE)
	}
	Pla_ManConvertToBits(q)
	return Pla_WritePla(q, name)
}
func Pla_ManDumpBlif(p *Pla_Man_t, name string) error {
	if p.CubeLits == nil {
		Pla_ManConvertFromBits(p)
	}
	original := p.Ins
	if len(p.Divs) > 0 {
		original = 0
		for _, d := range p.Divs {
			if d == 0 {
				original++
			}
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, ".model %s\n.inputs", p.Name)
	for v := int32(0); v < original; v++ {
		fmt.Fprintf(&b, " i%d", v)
	}
	b.WriteString("\n.outputs o\n.names")
	for v := int32(0); v < p.Ins; v++ {
		fmt.Fprintf(&b, " i%d", v)
	}
	b.WriteString(" o\n")
	for _, cube := range p.CubeLits {
		line := make([]byte, p.Ins)
		for i := range line {
			line[i] = '-'
		}
		for _, lit := range cube {
			line[lit/2] = '1' - byte(lit&1)
		}
		fmt.Fprintf(&b, "%s 1\n", line)
	}
	for i := int(original); i < len(p.Divs); i++ {
		d := p.Divs[i]
		a, c, e := (d>>2)&1023, (d>>12)&1023, (d>>22)&1023
		fmt.Fprintf(&b, ".names i%d i%d", a/2, c/2)
		if d&3 == 3 {
			fmt.Fprintf(&b, " i%d", e/2)
		}
		fmt.Fprintf(&b, " i%d\n", i)
		switch d & 3 {
		case 1:
			fmt.Fprintf(&b, "%d%d 1\n", 1-a%2, 1-c%2)
		case 2:
			b.WriteString("10 1\n01 1\n")
		case 3:
			fmt.Fprintf(&b, "%d-0 1\n-11 1\n", 1-a%2)
		default:
			return fmt.Errorf("invalid divisor definition at %d", i)
		}
	}
	b.WriteString(".end\n")
	return os.WriteFile(name, []byte(b.String()), 0666)
}
func Pla_ManFxPerformSimple(n int32) int32 {
	p := Pla_ManFxPrepare(n)
	if err := Pla_ManDumpPla(p, fmt.Sprintf("primesmin%02d.pla", n)); err != nil {
		return 0
	}
	return 1
}
