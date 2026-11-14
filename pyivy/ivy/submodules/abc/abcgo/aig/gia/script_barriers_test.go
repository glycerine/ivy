package gia

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/dch"
	"testing"
)

func TestScriptBarrierRoundTrip(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	z := Gia_ManAppendBuf(p, x)
	Gia_ManAppendCo(p, z)
	split := Gia_ManDupFromBarBufs(p)
	if Gia_ManCiNum(split) != 3 || Gia_ManCoNum(split) != 2 {
		t.Fatal("barrier not split")
	}
	split.pSibls = make([]int32, split.nObjs+1)
	restored := Gia_ManDupToBarBufs(split, 1)
	if Gia_ManBufNum(restored) != 1 || Gia_ManCiNum(restored) != 2 || Gia_ManCoNum(restored) != 1 {
		t.Fatal("barrier not restored")
	}
	for _, in := range [][]bool{{false, false}, {false, true}, {true, false}, {true, true}} {
		aout, e := p.Evaluate(in)
		if e != nil {
			t.Fatal(e)
		}
		bout, e := restored.Evaluate(in)
		if e != nil || aout[0] != bout[0] {
			t.Fatalf("input %v outputs %v %v", in, aout, bout)
		}
	}
}

func TestScriptSynthesisTruth(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	for i, run := range []func(*Gia_Man_t, int32, int32) *Gia_Man_t{Gia_ManAigSyn3, Gia_ManAigSyn4} {
		q := run(p, 0, 0)
		if q == nil || Gia_ManVerifyTwoTruths(p, q) != 1 {
			t.Fatalf("synthesis %d changed truth", i)
		}
	}
}

func TestSyn2SynthesisTruth(t *testing.T) {
	for _, old := range []int32{0, 1} {
		for _, delay := range []int32{0, 1} {
			p := Gia_ManStart(10)
			a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
			x := Gia_ManAppendAnd(p, a, b^1)
			Gia_ManAppendCo(p, Gia_ManAppendAnd(p, x, c))
			calls := 0
			balance := func(g *Gia_Man_t, size, cuts, relax, verbose int32) *Gia_Man_t {
				calls++
				if size != 6 || cuts != 4 || relax != 0 || verbose != 0 {
					t.Fatal("incorrect DSD balancing parameters")
				}
				return Gia_ManDup(g)
			}
			q := giaAigSyn2(p, old, 0, 0, 0, delay, 0, 0, balance)
			if q == nil || Gia_ManVerifyTwoTruths(p, q) != 1 || calls != int(delay) {
				t.Fatalf("old=%d delay=%d calls=%d: incorrect result", old, delay, calls)
			}
		}
	}
}

func TestScriptPortOrdering(t *testing.T) {
	g := Gia_ManStart(8)
	a := Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, a)
	b := Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, b)
	p := aig.NewGraph()
	x, y := p.NewInput(), p.NewInput()
	outX, outY := p.AddOutput(x), p.AddOutput(y)
	order := Gia_ManOrderPios(p, g)
	want := []aig.Edge{x, outX, y, outY}
	for i, edge := range want {
		if vec.Vec_PtrEntry(order, i) != edge {
			t.Fatalf("port %d: %v want %v", i, vec.Vec_PtrEntry(order, i), edge)
		}
	}
}

func TestScriptChoiceConstruction(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	out := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, out)
	pars := dch.DefaultParams()
	pars.Synthesis = false
	q := Gia_ManAigSynch2Choices(p, Gia_ManDup(p), Gia_ManDup(p), &pars)
	if q == nil || Gia_ManVerifyTwoTruths(p, q) != 1 {
		t.Fatal("choice construction changed function")
	}
}

func TestSynch2ChoiceSynthesis(t *testing.T) {
	for _, light := range []bool{false, true} {
		p := Gia_ManStart(12)
		p.pName, p.PSpec = "choice-test", "example"
		a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		x := Gia_ManAppendAnd(p, a, b^1)
		Gia_ManAppendCo(p, Gia_ManAppendAnd(p, x, c))
		pars := dch.DefaultParams()
		pars.LightSynth, pars.Synthesis = light, false
		calls := 0
		balance := func(g *Gia_Man_t, size, cuts, relax, verbose int32) *Gia_Man_t {
			calls++
			if size != 6 || cuts != 8 || relax != 0 || verbose != 0 {
				t.Fatal("incorrect DSD balancing parameters")
			}
			return Gia_ManDup(g)
		}
		q := giaAigSynch2(p, &pars, 10, balance)
		wantCalls := 1
		if light {
			wantCalls = 0
		}
		if q == nil || Gia_ManVerifyTwoTruths(p, q) != 1 || calls != wantCalls || q.pName != p.pName || q.PSpec != p.PSpec {
			t.Fatalf("light=%v calls=%d: synthesis changed the graph or metadata", light, calls)
		}
	}
}
