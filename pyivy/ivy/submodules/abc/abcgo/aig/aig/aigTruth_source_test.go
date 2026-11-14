package aig

import (
	"fmt"
	"strings"
	"testing"
)

func TestCutTruthWordsMatchOriginalC(t *testing.T) {
	var want strings.Builder
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	av, bv := make([]uint32, 17), make([]uint32, 17)
	for i := range av {
		av[i] = uint32(0x9e3779b9)*uint32(i+1) ^ 0x13579bdf
		bv[i] = uint32(0x7f4a7c15)*uint32(i+3) + 0x2468ace0
	}
	children := map[int][]uint32{a.NodeID(): av, b.NodeID(): bv}
	for kind := 0; kind < 2; kind++ {
		signs := 4
		if kind != 0 {
			signs = 1
		}
		for sign := 0; sign < signs; sign++ {
			left, right := dupPhase(a, sign&1 != 0), dupPhase(b, sign&2 != 0)
			root := g.AND(left, right)
			if kind != 0 {
				root = g.NewXor(left, right)
			}
			for _, words := range []int{1, 2, 4, 17} {
				result := Aig_ManCutTruthOne(g, root, make([]uint32, words), children)
				fmt.Fprintf(&want, "kind %d signs %d words %d:", kind, sign, words)
				for _, word := range result {
					fmt.Fprintf(&want, " %08x", word)
				}
				fmt.Fprintln(&want)
			}
		}
	}
	got := runAigSourceFixture(t, "truth_words_compare.c")
	if got != want.String() {
		t.Fatalf("Go truth words:\n%sC truth words:\n%s", want.String(), got)
	}
}
