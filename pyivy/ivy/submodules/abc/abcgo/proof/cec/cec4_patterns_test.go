package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCec4GeneratedPatternsImplyRequestedValues(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b^1)
	y := gia.Gia_ManAppendAnd(g, x, c)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, y)
	for xv := int32(0); xv < 2; xv++ {
		for yv := int32(0); yv < 2; yv++ {
			for trial := 0; trial < 20; trial++ {
				pat, vis := vec.Vec_IntAlloc(3), vec.Vec_IntAlloc(8)
				if Cec4_ManGeneratePatternOne(g, x>>1, xv, y>>1, yv, pat, vis) == 0 {
					continue
				}
				var matched int
				for mask := 0; mask < 8; mask++ {
					values := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
					matches := true
					for _, lit := range vec.Vec_IntArray(pat) {
						id := lit >> 1
						index := int(id - 1)
						if values[index] != (lit&1 != 0) {
							matches = false
							break
						}
					}
					if !matches {
						continue
					}
					matched++
					out, err := g.Evaluate(values)
					if err != nil {
						t.Fatal(err)
					}
					if out[0] != (xv != 0) || out[1] != (yv != 0) {
						t.Fatalf("pattern %v gave %v, want %d,%d", vec.Vec_IntArray(pat), out, xv, yv)
					}
				}
				if matched == 0 {
					t.Fatal("generated inconsistent input pattern")
				}
			}
		}
	}
}

func TestCec4PatternGeneratorHandlesXOR(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendXorReal(g, a, b)
	gia.Gia_ManAppendCo(g, x)
	for av := int32(0); av < 2; av++ {
		for xv := int32(0); xv < 2; xv++ {
			pat, vis := vec.Vec_IntAlloc(2), vec.Vec_IntAlloc(4)
			if Cec4_ManGeneratePatternOne(g, a>>1, av, x>>1, xv, pat, vis) == 0 {
				t.Fatalf("failed for a=%d xor=%d", av, xv)
			}
			values := make([]bool, 2)
			for _, lit := range vec.Vec_IntArray(pat) {
				values[(lit>>1)-1] = lit&1 != 0
			}
			got, err := g.Evaluate(values)
			if err != nil {
				t.Fatal(err)
			}
			if values[0] != (av != 0) || got[0] != (xv != 0) {
				t.Fatalf("a=%d xor=%d pattern %v output %v", av, xv, values, got)
			}
		}
	}
}
