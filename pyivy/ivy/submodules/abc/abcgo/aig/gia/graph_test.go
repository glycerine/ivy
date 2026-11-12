package gia

import "testing"

func TestGraphTruthTables(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.And(a, b))
	g.AddOutput(g.Or(a, b))
	g.AddOutput(g.Xor(a, b))
	g.AddOutput(g.Mux(c, a, b))
	for mask := 0; mask < 8; mask++ {
		av, bv, cv := mask&1 != 0, mask&2 != 0, mask&4 != 0
		out := g.Evaluate([]bool{av, bv, cv})
		mux := bv
		if cv {
			mux = av
		}
		want := []bool{av && bv, av || bv, av != bv, mux}
		for i := range want {
			if out[i] != want[i] {
				t.Fatalf("mask %d output %d: got %v want %v", mask, i, out[i], want[i])
			}
		}
	}
}
func TestGraphHashingAndReductions(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	for _, test := range []struct{ a, b, want Literal }{{a, 0, 0}, {a, 1, a}, {a, a, a}, {a, a.Not(), 0}} {
		if got := g.And(test.a, test.b); got != test.want {
			t.Fatal(test, got)
		}
	}
	x := g.And(a, b)
	n := g.ObjectCount()
	if g.And(b, a) != x || g.ObjectCount() != n {
		t.Fatal("structural hashing failed")
	}
	g.AddOutput(g.And(x, a.Not()))
	if g.Levels()[0] != 2 {
		t.Fatal(g.Levels())
	}
	if a.Object() != 1 || a.Not().Object() != 1 || !a.Not().Complemented() || a.Not().Not() != a {
		t.Fatal("literal encoding")
	}
}
func TestGraphValidation(t *testing.T) {
	for _, fn := range []func(){func() { NewGraph().AddOutput(2) }, func() { NewGraph().Evaluate([]bool{true}) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("expected panic")
				}
			}()
			fn()
		}()
	}
}

func TestOriginalObjectRoundTrip(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.Xor(a, b))
	g.AddOutput(g.And(a, b).Not())
	objects := g.Objects()
	if Gia_ObjIsConst0(objects[0]) != 1 || Gia_ObjIsCi(objects[1]) != 1 || Gia_ObjIsCo(objects[len(objects)-1]) != 1 {
		t.Fatal("original object classification")
	}
	copy, err := GraphFromObjects(objects)
	if err != nil {
		t.Fatal(err)
	}
	for mask := 0; mask < 4; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0}
		want, got := g.Evaluate(inputs), copy.Evaluate(inputs)
		for i := range want {
			if got[i] != want[i] {
				t.Fatal(mask, got, want)
			}
		}
	}
	objects[3].iDiff0 = 999
	if _, err := GraphFromObjects(objects); err == nil {
		t.Fatal("invalid backward fanin accepted")
	}
}

func TestOriginalFaninAccessors(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	and := g.And(a, b.Not())
	g.AddOutput(and.Not())
	objects := g.Objects()
	p := objects[and.Object()]
	if Gia_ObjFanin0(p) != objects[a.Object()] || Gia_ObjFanin1(p) != objects[b.Object()] || Gia_ObjFaninC1(p) != 1 {
		t.Fatal("fanin accessors")
	}
	if Gia_ObjFaninLit0(p, int32(and.Object())) != int32(a) || Gia_ObjFaninLit1(p, int32(and.Object())) != int32(b.Not()) {
		t.Fatal("fanin literals")
	}
	co := objects[len(objects)-1]
	if Gia_ObjCioId(co) != 0 || Gia_ObjCioId(objects[2]) != 1 {
		t.Fatal("CIO order")
	}
	Gia_ObjFlipFaninC0(co)
	if Gia_ObjFaninC0(co) != 0 {
		t.Fatal("complement flip")
	}
	Gia_ObjSetValue(p, -1)
	if Gia_ObjValue(p) != -1 {
		t.Fatal("uint32/int32 value roundtrip")
	}
}

func TestOriginalComplementHandles(t *testing.T) {
	p := &Gia_Obj_t{Value: 42}
	inverse := Gia_Not(p)
	if Gia_Not(p) != inverse || Gia_Not(inverse) != p || Gia_Regular(inverse) != p || Gia_IsComplement(inverse) != 1 || Gia_IsComplement(p) != 0 {
		t.Fatal("complement pairing")
	}
	if Gia_NotCond(p, 0) != p || Gia_NotCond(p, 1) != inverse || Gia_Regular(inverse).Value != 42 {
		t.Fatal("conditional complement")
	}
}

func TestOriginalComplementHandlesConcurrent(t *testing.T) {
	object := &Gia_Obj_t{}
	results := make(chan *Gia_Obj_t, 64)
	for i := 0; i < 64; i++ {
		go func() { results <- Gia_Not(object) }()
	}
	expected := Gia_Not(object)
	for i := 0; i < 64; i++ {
		if got := <-results; got != expected || Gia_Regular(got) != object {
			t.Fatal("unstable concurrent edge handle")
		}
	}
}

func TestOriginalObjectImportRejectsMalformedTerminals(t *testing.T) {
	constant := &Gia_Obj_t{iDiff0: objectNoFanin, iDiff1: objectNoFanin, fTerm: 1}
	if _, err := GraphFromObjects([]*Gia_Obj_t{constant}); err == nil {
		t.Fatal("terminal accepted as constant")
	}
	g := NewGraph()
	a := g.NewInput()
	g.AddOutput(a)
	objects := g.Objects()
	objects = append(objects, &Gia_Obj_t{iDiff0: 1, iDiff1: 2})
	if _, err := GraphFromObjects(objects); err == nil {
		t.Fatal("CO accepted as an AND fanin")
	}
}
