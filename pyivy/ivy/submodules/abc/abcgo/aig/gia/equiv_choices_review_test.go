package gia

import "testing"

// Compare every combinational input, including register outputs, independently
// of the choice structure. Register inputs are evaluated as ordinary outputs.
func checkChoiceOutputs(t *testing.T, p, q *Gia_Man_t, inputs int) {
	t.Helper()
	if Gia_ManRegNum(p) != Gia_ManRegNum(q) {
		t.Fatalf("registers: source %d result %d", Gia_ManRegNum(p), Gia_ManRegNum(q))
	}
	for bits := 0; bits < 1<<inputs; bits++ {
		in := make([]bool, inputs)
		for i := range in {
			in[i] = bits&(1<<i) != 0
		}
		want, err := p.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		got, err := q.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		if len(want) != len(got) {
			t.Fatalf("bits %d: outputs %v != %v", bits, got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("bits %d: outputs %v != %v", bits, got, want)
			}
		}
	}
}

func TestEquivChoicesComplementedPhase(t *testing.T) {
	p := Gia_ManStart(32)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	or := func(x, y int32) int32 { return Gia_ManAppendAnd(p, x^1, y^1) ^ 1 }
	x := Gia_ManAppendAnd(p, or(a, b), or(a, c))
	y := Gia_ManAppendAnd(p, a^1, or(b^1, c^1))
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, y)
	Gia_ManAppendCo(p, y^1)
	p.StartRepresentatives()
	Gia_ObjSetRepr(p, y>>1, x>>1)
	p.pNexts = Gia_ManDeriveNexts(p)
	q := Gia_ManEquivToChoices(p, 1)
	checkChoiceOutputs(t, p, q, 3)
	if Gia_ManCountChoices(q) != 1 {
		t.Fatalf("expected complemented choice, got %d", Gia_ManCountChoices(q))
	}
	Gia_ManSetPhase(q)
	for head, next := range q.pNexts {
		if next == 0 {
			continue
		}
		for bits := 0; bits < 8; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			values, err := giaEvaluateNodes(q, in)
			if err != nil {
				t.Fatal(err)
			}
			phaseDiff := q.pObjs[head].fPhase != q.pObjs[next].fPhase
			if (values[head] != values[next]) != phaseDiff {
				t.Fatalf("choice members differ beyond phase for input %d", bits)
			}
		}
	}
}

func TestEquivChoicesDependentAndShared(t *testing.T) {
	p := Gia_ManStart(24)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	or := Gia_ManAppendAnd(p, a^1, b^1) ^ 1
	y := Gia_ManAppendAnd(p, x, or) // y = x but depends on x.
	Gia_ManAppendCo(p, y)
	Gia_ManAppendCo(p, x)
	p.StartRepresentatives()
	Gia_ObjSetRepr(p, y>>1, x>>1)
	p.pNexts = Gia_ManDeriveNexts(p)
	q := Gia_ManEquivToChoices(p, 1)
	checkChoiceOutputs(t, p, q, 2)
	if Gia_ManCountChoices(q) != 0 {
		t.Fatalf("dependent alternative must not be choice: %d", Gia_ManCountChoices(q))
	}
}

func TestEquivChoicesConstantsAndRegisters(t *testing.T) {
	p := Gia_ManStart(24)
	a, ro := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	out := Gia_ManAppendAnd(p, a, ro)
	zero := Gia_ManAppendAnd(p, out, a^1)
	one := zero ^ 1
	Gia_ManAppendCo(p, zero)
	Gia_ManAppendCo(p, one)
	Gia_ManAppendCo(p, out)
	Gia_ManAppendCo(p, a^1)
	Gia_ManSetRegNum(p, 1)
	p.StartRepresentatives()
	Gia_ObjSetRepr(p, zero>>1, 0)
	p.pNexts = Gia_ManDeriveNexts(p)
	q := Gia_ManEquivToChoices(p, 1)
	checkChoiceOutputs(t, p, q, 2)
}

func TestEquivChoicesSnapshots(t *testing.T) {
	p := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, x^1)
	Gia_ManAppendCo(p, x^1)
	p.StartRepresentatives()
	q := Gia_ManEquivToChoices(p, 2)
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		got, err := q.Evaluate(in)
		if err != nil || len(got) != 2 || got[0] != (bits == 3) || got[1] != (bits != 3) {
			t.Fatalf("bits=%d outputs=%v error=%v", bits, got, err)
		}
	}
}
