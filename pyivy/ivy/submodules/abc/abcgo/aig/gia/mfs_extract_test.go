package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGiaMfsExtractSimple(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	pars.FCoarsen = 0
	mapped := Mf_ManPerformMapping(g, &pars)
	d := giaMfsExtract(mapped)
	if d.pis != 2 || d.pos != 1 || len(d.fanins) != 5 || len(d.fanins[3]) != 2 || d.truths[3][0]&15 != 8 || len(d.fanins[4]) != 1 {
		t.Fatalf("extracted %+v", d)
	}
}

func TestGiaMfsExtractWhiteBoxInput(t *testing.T) {
	g := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	boxOut := Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, a)
	Gia_ManAppendCo(g, b)
	Gia_ManAppendCo(g, boxOut)
	tm := tim.Tim_ManStart(3, 3)
	tim.Tim_ManCreateBox(tm, 0, 2, 2, 1, -1, 0)
	g.pManTime = tm
	extra := Gia_ManStart(5)
	Gia_ManAppendCi(extra)
	e1 := Gia_ManAppendCi(extra)
	Gia_ManAppendCo(extra, e1)
	g.pAigExtra = extra
	g.vMapping = vec.Vec_IntStart(int(g.nObjs))
	data := giaMfsExtract(g)
	node := Gia_ObjCopyArray(g, boxOut>>1)
	if !data.fixed[node] || len(data.fanins[node]) != 1 || data.truths[node][0] != 0xaaaaaaaaaaaaaaaa {
		t.Fatalf("box fanins %v truth %x", data.fanins[node], data.truths[node])
	}
	result := giaMfsInsert(g, data, true)
	if Gia_ManCiNum(result) != 3 || Gia_ManCoNum(result) != 3 || Gia_ManBoxNum(result) != 1 {
		t.Fatalf("roundtrip CI %d CO %d boxes %d", Gia_ManCiNum(result), Gia_ManCoNum(result), Gia_ManBoxNum(result))
	}
}

func TestGiaMfsExtractBlackBox(t *testing.T) {
	g := Gia_ManStart(8)
	a := Gia_ManAppendCi(g)
	boxOut := Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, a)
	Gia_ManAppendCo(g, boxOut)
	tm := tim.Tim_ManStart(2, 2)
	tim.Tim_ManCreateBox(tm, 0, 1, 1, 1, -1, 1)
	g.pManTime = tm
	extra := Gia_ManStart(4)
	Gia_ManAppendCi(extra)
	Gia_ManAppendCo(extra, 0)
	g.pAigExtra = extra
	g.vMapping = vec.Vec_IntStart(int(g.nObjs))
	data := giaMfsExtract(g)
	node := Gia_ObjCopyArray(g, boxOut>>1)
	if data.pis != 2 || data.pos != 2 || len(data.fanins[node]) != 1 || data.fanins[node][0] != 0 {
		t.Fatalf("black box extraction %+v", data)
	}
	result := giaMfsInsert(g, data, true)
	if Gia_ManCiNum(result) != 2 || Gia_ManCoNum(result) != 2 || Gia_ManBoxNum(result) != 1 {
		t.Fatalf("roundtrip CI %d CO %d boxes %d", Gia_ManCiNum(result), Gia_ManCoNum(result), Gia_ManBoxNum(result))
	}
}
