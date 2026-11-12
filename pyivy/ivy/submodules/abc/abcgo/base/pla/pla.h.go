package pla

/**CFile****************************************************************

  FileName    [pla.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SOP manager.]

  Synopsis    [Scalable SOP transformations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - March 18, 2015.]

  Revision    [$Id: pla.h,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

import (
	"fmt"
	"io"
	"math/bits"
	"os"
	"path/filepath"
	"strings"
)

type Pla_File_t uint8

const (
	PLA_FILE_FD Pla_File_t = iota
	PLA_FILE_F
	PLA_FILE_FR
	PLA_FILE_FDR
	PLA_FILE_NONE
)

type Pla_Lit_t uint8

const (
	PLA_LIT_DASH Pla_Lit_t = iota
	PLA_LIT_ZERO
	PLA_LIT_ONE
	PLA_LIT_FULL
)
const MASK55 uint64 = 0x5555555555555555

// Cube stores the original two-bit literal encoding, independently for inputs
// and outputs. This manager is distinct from Espresso's positional-set cubes.
type Cube struct{ In, Out []uint64 }
type Pla_Man_t struct {
	Name, Spec       string
	Type             Pla_File_t
	Ins, Outs        int32
	Cubes            []Cube
	CubeLits, Occurs [][]int32
	Divs             []uint32
	Directives       []string
	Warnings         []string
}

func words(n int32) int { return int((int64(n)*2 + 63) / 64) }
func Pla_ManAlloc(name string, ins, outs, cubes int32) *Pla_Man_t {
	if ins < 0 || outs < 0 || cubes < 0 {
		panic("negative PLA dimensions")
	}
	base := filepath.Base(name)
	p := &Pla_Man_t{Name: strings.TrimSuffix(base, filepath.Ext(base)), Spec: name, Ins: ins, Outs: outs}
	p.Cubes = make([]Cube, cubes)
	for i := range p.Cubes {
		p.Cubes[i] = Cube{make([]uint64, words(ins)), make([]uint64, words(outs))}
	}
	return p
}
func Pla_ManFree(p *Pla_Man_t) {
	if p != nil {
		*p = Pla_Man_t{}
	}
}
func Pla_ManName(p *Pla_Man_t) string            { return p.Name }
func Pla_ManInNum(p *Pla_Man_t) int32            { return p.Ins }
func Pla_ManOutNum(p *Pla_Man_t) int32           { return p.Outs }
func Pla_ManCubeNum(p *Pla_Man_t) int32          { return int32(len(p.Cubes)) }
func Pla_ManDivNum(p *Pla_Man_t) int32           { return int32(len(p.Divs)) }
func Pla_CubeIn(p *Pla_Man_t, i int32) []uint64  { return p.Cubes[i].In }
func Pla_CubeOut(p *Pla_Man_t, i int32) []uint64 { return p.Cubes[i].Out }
func Pla_CubeNum(handle int32) int32             { return handle >> 8 }
func Pla_CubeLit(handle int32) int32             { return handle & 255 }
func Pla_CubeHandle(cube, lit int32) int32 {
	if cube < 0 || cube >= 1<<24 || lit < 0 || lit >= 256 {
		panic("PLA cube handle overflow")
	}
	return cube<<8 | lit
}
func Pla_TtGetBit(p []uint64, i int32) int32       { return int32(p[i>>6] >> uint(i&63) & 1) }
func Pla_TtSetBit(p []uint64, i int32)             { p[i>>6] |= uint64(1) << uint(i&63) }
func Pla_TtXorBit(p []uint64, i int32)             { p[i>>6] ^= uint64(1) << uint(i&63) }
func Pla_CubeGetLit(p []uint64, i int32) Pla_Lit_t { return Pla_Lit_t(p[i>>5] >> uint((i&31)*2) & 3) }

// SetLit retains the source OR-insertion operation; ReplaceLit overwrites.
func Pla_CubeSetLit(p []uint64, i int32, d Pla_Lit_t) {
	if d > 3 {
		panic("invalid PLA literal")
	}
	p[i>>5] |= uint64(d) << uint((i&31)*2)
}
func Pla_CubeReplaceLit(p []uint64, i int32, d Pla_Lit_t) {
	if d > 3 {
		panic("invalid PLA literal")
	}
	shift := uint((i & 31) * 2)
	p[i>>5] = p[i>>5]&^(uint64(3)<<shift) | uint64(d)<<shift
}
func Pla_CubeXorLit(p []uint64, i int32, d Pla_Lit_t) { p[i>>5] ^= uint64(d) << uint((i&31)*2) }
func Pla_OnlyOneOne(x uint64) int32                   { return boolInt(x != 0 && x&(x-1) == 0) }
func Pla_CubesAreDistance1(a, b []uint64, n int32) int32 {
	found := false
	for i := int32(0); i < n; i++ {
		x := a[i] ^ b[i]
		if x == 0 {
			continue
		}
		if found || Pla_OnlyOneOne((x|x>>1)&MASK55) == 0 {
			return 0
		}
		found = true
	}
	return boolInt(found)
}
func Pla_CubesAreConsensus(a, b []uint64, n int32, variable *int32) int32 {
	found := false
	for i := int32(0); i < n; i++ {
		x := a[i] ^ b[i]
		if x == 0 {
			continue
		}
		if found {
			return 0
		}
		test := (x & (x >> 1)) & MASK55
		if Pla_OnlyOneOne(test) == 0 {
			return 0
		}
		found = true
		if variable != nil {
			*variable = i*32 + int32(bits.TrailingZeros64(test)/2)
		}
	}
	return boolInt(found)
}
func Pla_TtCountOnesOne(x uint64) int32 { return int32(bits.OnesCount64(x)) }
func Pla_TtCountOnes(p []uint64, n int32) int32 {
	count := int32(0)
	for _, v := range p[:n] {
		count += Pla_TtCountOnesOne(v)
	}
	return count
}
func Pla_ManLitInNum(p *Pla_Man_t) int32 {
	n := int32(0)
	for _, c := range p.Cubes {
		for i := int32(0); i < p.Ins; i++ {
			if Pla_CubeGetLit(c.In, i) != PLA_LIT_DASH {
				n++
			}
		}
	}
	return n
}
func Pla_ManLitOutNum(p *Pla_Man_t) int32 {
	n := int32(0)
	for _, c := range p.Cubes {
		for i := int32(0); i < p.Outs; i++ {
			if Pla_CubeGetLit(c.Out, i) == PLA_LIT_ONE {
				n++
			}
		}
	}
	return n
}
func (p *Pla_Man_t) PrintStats(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%-16s :  In =%4d  Out =%4d  Cube =%8d  LitIn =%8d  LitOut =%8d  Div =%6d\n", p.Name, p.Ins, p.Outs, len(p.Cubes), Pla_ManLitInNum(p), Pla_ManLitOutNum(p), len(p.Divs))
	return err
}
func Pla_ManPrintStats(p *Pla_Man_t, verbose int32) { p.PrintStats(os.Stdout) }
func boolInt(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

// Evaluate returns explicit 0/1 membership accumulated from matching rows.
// DASH means no explicit membership; FULL means both memberships were present.
func (p *Pla_Man_t) Evaluate(input []bool) []Pla_Lit_t {
	if len(input) != int(p.Ins) {
		panic("PLA input width mismatch")
	}
	out := make([]Pla_Lit_t, p.Outs)
	for _, c := range p.Cubes {
		match := true
		for i, value := range input {
			lit := Pla_CubeGetLit(c.In, int32(i))
			if lit == PLA_LIT_FULL || lit == PLA_LIT_ZERO && value || lit == PLA_LIT_ONE && !value {
				match = false
				break
			}
		}
		if match {
			for i := range out {
				out[i] |= Pla_CubeGetLit(c.Out, int32(i))
			}
		}
	}
	return out
}
