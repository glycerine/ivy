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
