package aig

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestRetimingPackedDeque(t *testing.T) {
	p := Rtm_ManAlloc(nil)
	e := &Rtm_Edg_t{}
	var oracle []Rtm_Init_t
	r := rand.New(rand.NewSource(71))
	for n := 0; n < 10000; n++ {
		op := r.Intn(4)
		if len(oracle) == 0 {
			op = 0
		}
		v := Rtm_Init_t(1 + r.Intn(3))
		switch op {
		case 0:
			Rtm_ObjAddFirst(p, e, v)
			oracle = append([]Rtm_Init_t{v}, oracle...)
		case 1:
			Rtm_ObjAddLast(p, e, v)
			oracle = append(oracle, v)
		case 2:
			if got := Rtm_ObjRemFirst(p, e); got != oracle[0] {
				t.Fatalf("first %d != %d", got, oracle[0])
			}
			oracle = oracle[1:]
		case 3:
			last := len(oracle) - 1
			if got := Rtm_ObjRemLast(p, e); got != oracle[last] {
				t.Fatal("last mismatch")
			}
			oracle = oracle[:last]
		}
		if int(e.nLats) != len(oracle) {
			t.Fatal("length")
		}
		for i, want := range oracle {
			if got := Rtm_ObjGetOne(p, e, int32(i)); got != want {
				t.Fatalf("step %d index %d got %d want %d", n, i, got, want)
			}
		}
	}
}
func TestEdgeRetimingForwardTrace(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	q, r := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(q.Not(), r))
	g.AddOutput(a)
	g.AddOutput(b)
	g.nRegs = 2
	sequence := [][]bool{{false, false}, {true, false}, {true, true}, {false, true}, {false, false}, {true, true}}
	before := retFTrace(t, g, sequence)
	for _, steps := range []int32{0, 1, 3} {
		out := Rtm_ManRetime(g, 1, steps, 0)
		if err := out.Check(); err != nil {
			t.Fatal(err)
		}
		if after := retFTrace(t, out, sequence); !reflect.DeepEqual(before, after) {
			t.Fatalf("steps %d before %v after %v", steps, before, after)
		}
	}
}

func TestRetimeManagerFinalizesTerminalMetadata(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	g.AddOutput(b)
	Aig_ManSetRegNum(g, 0)
	q := Rtm_ManToAig(Rtm_ManFromAig(g))
	if q.TruePiCount() != 2 || q.TruePoCount() != 2 || Aig_ObjCioId(q, q.inputs[1]) != 1 || Aig_ObjCioId(q, q.outputs[1]) != 1 {
		t.Fatal("retiming manager did not finalize terminal metadata")
	}
}

func TestEdgeRetimingFeedbackAndXor(t *testing.T) {
	for _, xor := range []bool{false, true} {
		g := NewGraph()
		a := g.NewInput()
		q, r := g.NewInput(), g.NewInput()
		v := g.AND(q, r.Not())
		if xor {
			v = g.NewXor(q, r.Not())
		}
		g.AddOutput(v)
		g.AddOutput(a)
		g.AddOutput(q.Not())
		g.nRegs = 2
		sequence := [][]bool{{false}, {true}, {false}, {false}, {true}, {true}, {false}}
		want := retFTrace(t, g, sequence)
		for _, direction := range []int32{0, 1} {
			out := Rtm_ManRetime(g, direction, 3, 0)
			if err := out.Check(); err != nil {
				t.Fatal(err)
			}
			if direction == 1 && !reflect.DeepEqual(retFTrace(t, out, sequence), want) {
				t.Fatal("feedback trace changed")
			}
		}
	}
}
func TestRetimingExplicitPackedBoundaries(t *testing.T) {
	p := Rtm_ManAlloc(nil)
	e := &Rtm_Edg_t{}
	for i := 0; i < 100; i++ {
		Rtm_ObjAddLast(p, e, Rtm_Init_t(1+i%3))
	}
	for i := 0; i < 100; i++ {
		if got := Rtm_ObjRemFirst(p, e); got != Rtm_Init_t(1+i%3) {
			t.Fatalf("boundary %d got %d", i, got)
		}
	}
	if e.big != nil || e.LData != 0 {
		t.Fatal("storage not compact after drain")
	}
}
