package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestArenaViolationHelperTruthTable(t *testing.T) {
	old := aig.Aig_ManStart(20)
	a, b := aig.Aig_ObjCreateCi(old), aig.Aig_ObjCreateCi(old)
	po0, po1 := aig.Aig_ObjCreateCo(old, a), aig.Aig_ObjCreateCo(old, b.Not())
	dst := aig.Aig_ManStart(30)
	ac, bc := aig.Aig_ObjCreateCi(dst), aig.Aig_ObjCreateCi(dst)
	begin, within := aig.Aig_ObjCreateCi(dst), aig.Aig_ObjCreateCi(dst)
	prev0, prev1 := aig.Aig_ObjCreateCi(dst), aig.Aig_ObjCreateCi(dst)
	aig.Aig_ObjSetCopy(old, a, ac)
	aig.Aig_ObjSetCopy(old, b, bc)
	if got := driverToPoNew(old, po0); got != ac {
		t.Fatalf("positive copied driver %v", got)
	}
	if got := driverToPoNew(old, po1); got != bc.Not() {
		t.Fatalf("complemented copied driver %v", got)
	}
	var next, signals []aig.Edge
	violation := createArenaViolation(old, dst, begin, within, [][]int32{{0}, {1}}, []aig.Edge{prev0, prev1}, &next, &signals)
	if len(next) != 2 || len(signals) != 2 {
		t.Fatalf("next=%v signals=%v", next, signals)
	}
	for _, e := range []aig.Edge{signals[0], signals[1], next[0], next[1], violation} {
		aig.Aig_ObjCreateCo(dst, e)
	}
	for bits := 0; bits < 64; bits++ {
		in := make([]bool, 6)
		for i := range in {
			in[i] = bits&(1<<i) != 0
		}
		got, err := dst.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		x, y, start, active, p0, p1 := in[0], !in[1], in[2], in[3], in[4], in[5]
		want := []bool{x, y, (x && start) || p0, (y && start) || p1, active && ((x != p0) || (y != p1))}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("input %v: got %v want %v", in, got, want)
			}
		}
	}
}

func TestSourceArenaHelpersAllowExtraLatchEntries(t *testing.T) {
	old := aig.Aig_ManStart(8)
	x := aig.Aig_ObjCreateCi(old)
	aig.Aig_ObjCreateCo(old, x)
	dst := aig.Aig_ManStart(8)
	copy := aig.Aig_ObjCreateCi(dst)
	extra := aig.Aig_ObjCreateCi(dst)
	aig.Aig_ObjSetCopy(old, x, copy)
	var next, monotone []aig.Edge
	createArenaViolation(old, dst, copy, copy, [][]int32{{0}}, []aig.Edge{copy, extra}, &next, &monotone)
	if len(next) != 1 || len(monotone) != 1 {
		t.Fatalf("next=%v monotone=%v", next, monotone)
	}
	outs := createArenaLi(dst, [][]int32{{0}}, []aig.Edge{copy, extra})
	if len(outs) != 1 {
		t.Fatalf("source emits one CO per barrier, got %d", len(outs))
	}
	if createArenaLO(dst, nil) != nil || createArenaLi(dst, nil, nil) != nil || collectBarrierDisjunctions(old, dst, nil) != nil {
		t.Fatal("source empty helper returns NULL")
	}
}

func TestSourceMonotonicityVerificationTruthTable(t *testing.T) {
	g := aig.NewGraph()
	inputs := make([]aig.Edge, 6)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	result := createAndGateForMonotonicityVerification(g, inputs[:2], inputs[2:4], inputs[4], inputs[5])
	g.AddOutput(result)
	for mask := 0; mask < 64; mask++ {
		v := make([]bool, 6)
		for i := range v {
			v[i] = mask&(1<<i) != 0
		}
		got, err := g.Evaluate(v)
		if err != nil {
			t.Fatal(err)
		}
		guard := v[4] && v[5]
		want := (!guard || !v[2] || v[0]) && (!guard || !v[3] || v[1])
		if got[0] != want {
			t.Fatalf("%v: got %v want %v", v, got[0], want)
		}
	}
}
