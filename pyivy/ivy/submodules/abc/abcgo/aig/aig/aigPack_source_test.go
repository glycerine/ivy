package aig

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

func TestPackedRandomWordsMatchSource(t *testing.T) {
	for _, cared := range []bool{false, true} {
		t.Run(fmt.Sprintf("cared=%t", cared), func(t *testing.T) {
			g := NewGraph()
			for i := 0; i < 4; i++ {
				g.NewInput()
			}
			p := Aig_ManPackAlloc(g)
			rng := NewAigRandomSource()
			rng.Uint32() // The reset call consumes the first word.
			want := make([]uint64, 4)
			for i := range want {
				want[i] = (uint64(rng.Uint32())<<32 | uint64(rng.Uint32())) << 1
				if cared {
					p.vPiCare[i] = ^uint64(0)
					p.vPiPats[i] = ^want[i]
				}
			}
			beforeCare := p.CareWords()
			Aig_ManRandom(1)
			defer Aig_ManRandom(1)
			Aig_ManPackSetRandom(p)
			if !slices.Equal(p.InputWords(), want) {
				t.Errorf("random words = %x, want source words %x", p.InputWords(), want)
			}
			if !slices.Equal(p.CareWords(), beforeCare) {
				t.Error("randomization changed care masks")
			}
			if got, next := Aig_ManRandom(0), rng.Uint32(); got != next {
				t.Errorf("next random word = %d, want %d after two draws per CI", got, next)
			}
		})
	}
}

func TestPackedOperationsAgainstCompiledSource(t *testing.T) {
	g := NewGraph()
	inputs := make([]Edge, 7)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	x := g.AND(inputs[0], inputs[1])
	y := g.AND(inputs[0].Not(), inputs[1])
	a := g.AND(inputs[2], inputs[3].Not())
	b := g.AND(inputs[4].Not(), inputs[5].Not())
	c := g.AND(x, y.Not())
	d := g.AND(a.Not(), b.Not())
	g.AddOutput(g.AND(c, d))
	g.AddOutput(x.Not())
	g.AddOutput(g.Const0())
	g.AddOutput(g.Const1())
	p := Aig_ManPackAlloc(g)
	Aig_ManRandom(1)
	defer Aig_ManRandom(1)
	Aig_ManPackSetRandom(p)
	var input, want strings.Builder
	snapshot := func(result int32) {
		Aig_ManPackSimulate(p)
		fmt.Fprintf(&want, "%d %d %d %d:", result, Aig_ManPackCountCares(p), p.nPatTotal, p.nPatSkip)
		for i, word := range p.vPiPats {
			fmt.Fprintf(&want, " %016x/%016x", word, p.vPiCare[i])
		}
		want.WriteString(" | ")
		for _, word := range p.vSigns {
			fmt.Fprintf(&want, "%016x ", word)
		}
		want.WriteString("| ")
		for _, id := range Aig_ManPackConstNodes(p) {
			fmt.Fprintf(&want, "%d ", id)
		}
		want.WriteByte('\n')
	}
	apply := func(action, lane int, literals []int32) {
		fmt.Fprintf(&input, "%d %d %d", action, lane, len(literals))
		for _, lit := range literals {
			fmt.Fprintf(&input, " %d", lit)
		}
		input.WriteByte('\n')
		result := int32(-1)
		switch action {
		case 0:
			result = Aig_ManPackAddPatternTry(p, int32(lane), literals)
		case 1:
			Aig_ManPackAddPattern(p, literals)
		case 2:
			Aig_ManPackSetRandom(p)
		}
		snapshot(result)
	}
	snapshot(-1)
	for assignment := 0; assignment < 128; assignment++ {
		cube := make([]int32, 7)
		for i := range cube {
			cube[i] = int32(2 * i)
			if assignment&(1<<i) == 0 {
				cube[i] |= 1
			}
		}
		apply(1, 0, cube)
	}
	for _, lane := range []int{0, 1, 31, 32, 62, 63} {
		for _, cube := range [][]int32{nil, {0}, {1}, {12}, {13}, {4, 4}} {
			apply(0, lane, cube)
		}
	}
	rng := rand.New(rand.NewPCG(163, 197))
	for n := 0; n < 1000; n++ {
		if n == 250 || n == 750 {
			apply(2, 0, nil)
		}
		var cube []int32
		for variable := 0; variable < 7; variable++ {
			if rng.IntN(3) != 0 {
				continue
			}
			lit := int32(2*variable + rng.IntN(2))
			cube = append(cube, lit)
			if rng.IntN(5) == 0 {
				cube = append(cube, lit) // Consistent duplicates are valid in C.
			}
		}
		apply(rng.IntN(2), rng.IntN(64), cube)
	}
	want.WriteString("care diagnostic:\n")
	want.WriteString(captureAigOutput(t, func() { Aig_ManPackPrintCare(p) }))
	want.WriteString("statistics:\n")
	p.WriteStats(&want)
	want.WriteString(captureAigOutput(t, func() { Aig_ManPackFree(p) }))
	got := runAigSourceFixtureInput(t, "packing_compare.c", input.String(), "aigUtil.c")
	if got != want.String() {
		gotLines, wantLines := strings.Split(got, "\n"), strings.Split(want.String(), "\n")
		for i := 0; i < len(gotLines) && i < len(wantLines); i++ {
			if gotLines[i] != wantLines[i] {
				t.Fatalf("snapshot/line %d: C %q, Go %q", i, gotLines[i], wantLines[i])
			}
		}
		t.Fatalf("C has %d lines, Go has %d", len(gotLines), len(wantLines))
	}
}
