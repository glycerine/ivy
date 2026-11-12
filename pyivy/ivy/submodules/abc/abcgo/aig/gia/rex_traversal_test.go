package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRexEpsilonTraversal(t *testing.T) {
	s := Gia_ManRexPreprocess("(a|b)*c")
	transitions := rexTransitions(s)
	visited := vec.Vec_IntStartFull(len(s) + 1)
	reached := vec.Vec_IntAlloc(4)
	Gia_ManCollectReached(s, transitions, 0, reached, visited, 1)
	symbols := map[byte]bool{}
	for _, i := range vec.Vec_IntArray(reached) {
		symbols[rexSymbol(s, i)] = true
	}
	if !symbols['a'] || !symbols['b'] || !symbols['c'] || len(symbols) != 3 {
		t.Fatalf("initial symbols = %v", symbols)
	}
	var c int32
	for i := range s {
		if s[i] == 'c' {
			c = int32(i)
			break
		}
	}
	Gia_ManCollectReached(s, transitions, c, reached, visited, 2)
	if got := vec.Vec_IntArray(reached); len(got) != 1 || got[0] != int32(len(s)) {
		t.Fatalf("terminal states = %v", got)
	}
}
