package gia

import (
	"fmt"
	"testing"
)

func quantReviewMux() (*Gia_Man_t, int32) {
	g := Gia_ManStart(128)
	Gia_ManQuantSetSuppStart(g)
	var inputs [3]int32
	for i := range inputs {
		inputs[i] = Gia_ManAppendCi(g)
		Gia_ManQuantSetSuppZero(g)
		Gia_ManQuantSetSuppCi(g, Gia_ManObj(g, inputs[i]>>1))
	}
	Gia_ManHashAlloc(g)
	return g, Gia_ManHashMux(g, inputs[0], inputs[1], inputs[2])
}

func TestQuantSupportRowAppendKeepsNeighbor(t *testing.T) {
	g, _ := quantReviewMux()
	defer Gia_ManStop(g)
	first, second := Gia_ManCi(g, 0).id, Gia_ManCi(g, 1).id
	before := append([]word(nil), Gia_ManQuantInfoId(g, second)...)
	row := Gia_ManQuantInfoId(g, first)
	grown := append(row, word(12345))
	neighbor := Gia_ManQuantInfoId(g, second)
	for i := range before {
		if neighbor[i] != before[i] {
			t.Fatalf("append changed neighboring support word%d from%d to%d", i, before[i], neighbor[i])
		}
	}
	if grown[len(row)] != 12345 {
		t.Fatal("append did not retain value")
	}
	// The original row is still a writable view of precisely its own words.
	row[0] = 9
	if Gia_ManQuantInfoId(g, first)[0] != 9 {
		t.Fatal("row became an unrelated copy")
	}
}

func TestQuantExistEliminatesAllMuxInputs(t *testing.T) {
	for _, mode := range []int{0, 1} {
		for phase := int32(0); phase < 2; phase++ {
			t.Run(fmt.Sprintf("mode%d/phase%d", mode, phase), func(t *testing.T) {
				defer func() {
					if err := recover(); err != nil {
						t.Fatalf("quantifying all MUX inputs panicked: %v", err)
					}
				}()
				g, root := quantReviewMux()
				defer Gia_ManStop(g)
				keepNone := func(any, int32) int32 { return 0 }
				var lit int32
				if mode == 0 {
					lit = Gia_ManQuantExist(g, root^phase, keepNone, nil)
				} else {
					lit = Gia_ManQuantExist2(g, root^phase, keepNone, nil)
				}
				// Both a MUX and its complement are satisfiable, so eliminating every
				// input must leave true, even when true appears before the last variable.
				if lit != 1 {
					t.Fatalf("got literal%d, want constant true", lit)
				}
			})
		}
	}
}
