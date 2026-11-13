package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestXorImplicationAndCount(t *testing.T) {
	g := aig.NewGraph()
	li, lo, cand := g.NewInput(), g.NewInput(), g.NewInput()
	p := Ssw_SmlStart(g, 1, 1, 1)
	Ssw_SmlSimInfo(p, li)[1] = 3
	Ssw_SmlSimInfo(p, lo)[1] = 1
	Ssw_SmlSimInfo(p, cand)[1] = 1
	if Ssw_SmlCheckXorImplication(p, li, lo, cand) != 1 || Ssw_SmlCountXorImplication(p, li, lo, cand) != 1 {
		t.Fatal("wrong implication result")
	}
	Ssw_SmlSimInfo(p, cand)[1] = 3
	if Ssw_SmlCheckXorImplication(p, li, lo, cand) != 0 || Ssw_SmlCountXorImplication(p, li, lo, cand) != 1 {
		t.Fatal("failed implication missed")
	}
	if Ssw_SmlCountXorImplication(p, li, lo, cand.Not()) != 30 {
		t.Fatal("complemented candidate count wrong")
	}
}
