// Port of ABC FXCH, Bruno Schmitt, UFRGS (2016); original src/opt/fxch.
package fxch

import (
	"io"
	"math/bits"
)

// Cubes contain an output node ID followed by sorted ABC literals (2*id+phase).
type Cubes [][]int32
type Fxch_SubCube_t struct {
	Id, ICube    uint32
	ILit0, ILit1 uint32
}
type Fxch_SCHashTable_Entry_t struct{ Data []Fxch_SubCube_t }
type Fxch_SCHashTable_t struct {
	man  *Fxch_Man_t
	bins map[uint32][]Fxch_SubCube_t
}
type divisor struct {
	pairs    [][2]int32
	literals []int32
	weight   float32
}
type Fxch_Man_t struct {
	rngZ, rngW  uint32
	Cubes       Cubes
	Levels      []int32
	LitHashKeys []uint32
	CubeFree    []int32
	Div         []int32
	Log         io.Writer
	outputs     []map[int32]bool
	table       *Fxch_SCHashTable_t
	divisors    []divisor
	divIndex    map[string]int32
	singles     []int32
	Extracted   int32
}

func Fxch_CountOnes(v uint32) int32                  { return int32(bits.OnesCount32(v)) }
func Fxch_ManGetCube(p *Fxch_Man_t, c int32) []int32 { return p.Cubes[c] }
func Fxch_ManGetLit(p *Fxch_Man_t, c, l int32) int32 { return p.Cubes[c][l] }
func copyInts(a []int32) []int32                     { return append([]int32(nil), a...) }
func boolInt(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
