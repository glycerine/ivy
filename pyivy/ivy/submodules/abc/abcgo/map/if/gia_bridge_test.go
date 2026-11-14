package if_

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestGiaToIfBridge(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a^1, b)
	gia.Gia_ManAppendCo(g, x^1)
	var pars If_Par_t
	Gia_ManSetIfParsDefault(&pars)
	pars.NLutSize = 4
	p := Gia_ManToIf(g, &pars)
	defer If_ManStop(p)
	if If_ManCiNum(p) != 2 || If_ManCoNum(p) != 1 {
		t.Fatal("incorrect interface")
	}
	output := If_ManCo(p, 0)
	node := If_ObjFanin0(output)
	if If_ObjFaninC0(output) != 1 || If_ObjIsAnd(node) == 0 || If_ObjFaninC0(node) != 1 || If_ObjFaninC1(node) != 0 {
		t.Fatal("incorrect edge polarity")
	}
}
