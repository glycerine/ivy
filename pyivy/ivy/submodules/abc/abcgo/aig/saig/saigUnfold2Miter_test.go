package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestCreateIndMiter2(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	c := vec.Vec_VecAlloc(1)
	vec.Vec_VecPush(c, 0, lo)
	vec.Vec_VecPush(c, 0, pi.Not())
	q := Saig_ManCreateIndMiter2(g, c)
	if q.InputCount() != 4 || q.OutputCount() != 6 {
		t.Fatalf("inputs=%d outputs=%d", q.InputCount(), q.OutputCount())
	}
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 16; bits++ {
		in := make([]bool, 4)
		for i := range in {
			in[i] = bits&(1<<i) != 0
		}
		out, err := q.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		v0, v1, v2 := in[3], in[0], in[1]
		u0, u1, u2 := !in[0], !in[1], !in[2]
		want := []bool{v0 && !v1, v0 && v1, v0 && v1 && !v2, u0 && !u1, u0 && u1, u0 && u1 && !u2}
		for i := range want {
			if out[i] != want[i] {
				t.Fatalf("inputs %v output %v want %v", in, out, want)
			}
		}
	}
}
