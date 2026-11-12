package gia

import (
	"github.com/glycerine/abcgo/map/mio"
	"testing"
)

func TestNfEnumeratesCuts(t *testing.T) {
	const genlib = `GATE zero 0 O=CONST0;
GATE one 0 O=CONST1;
GATE buf 1 O=a; PIN * NONINV 1 999 1 0 1 0
GATE inv 1 O=!a; PIN * INV 1 999 1 0 1 0
GATE and2 2 O=a*b; PIN * NONINV 1 999 1 0 1 0
.end
`
	lib, err := mio.ReadLibrary("nf", genlib, nil)
	if err != nil {
		t.Fatal(err)
	}
	library := nfPrepareLibrary(mio.Mio_CollectRootsNew2(lib, 6, nil, 0), false, true, false)
	g := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, y)
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	p := newNfCutManager(g, &pars, library)
	p.computeCuts()
	if p.sets[x>>1] == 0 || p.sets[y>>1] == 0 {
		t.Fatal("missing cut set")
	}
	var cuts [NF_CUT_MAX]Nf_Cut_t
	n := p.prepare(cuts[:], y>>1, false)
	found := false
	for _, cut := range cuts[:n] {
		if cut.nLeaves == 2 && cut.pLeaves[0] == x>>1 && cut.pLeaves[1] == c>>1 {
			found = true
		}
	}
	// Leaves are sorted by object ID, including the intermediate AND.
	if !found {
		for _, cut := range cuts[:n] {
			if cut.nLeaves == 2 && cut.pLeaves[0] == c>>1 && cut.pLeaves[1] == x>>1 {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("missing two-input cut: %+v", cuts[:n])
	}
	if p.delays[y>>1] == 0 || p.counts[4] == 0 {
		t.Fatal("missing cut delay or library matches")
	}
}
