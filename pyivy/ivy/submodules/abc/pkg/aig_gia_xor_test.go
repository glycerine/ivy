package abc

import "testing"

func TestGiaNativeXorRecognitionAndMuxReferences(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	x, err := g.AppendXorReal(a.Not(), b)
	if err != nil {
		t.Fatal(err)
	}
	first, second, ok, err := g.RecognizeXor(x.ID())
	if err != nil || !ok {
		t.Fatal(ok, err)
	}
	mux, err := g.AppendMuxReal(c, a, b)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range []GiaLit{GiaLit(x.ID()) << 1, first, second, mux} {
		if _, err := g.AppendCO(l); err != nil {
			t.Fatal(err)
		}
	}
	for assignment := 0; assignment < 8; assignment++ {
		values, err := g.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0})
		if err != nil {
			t.Fatal(err)
		}
		if values[0] != (values[1] != values[2]) {
			t.Fatal("native xor phase")
		}
	}
	refs, err := g.MuxControlReferenceCounts(g.ObjectCount())
	if err != nil {
		t.Fatal(err)
	}
	for i, value := range refs {
		want := uint64(0)
		if GiaID(i) == c.ID() {
			want = 1
		}
		if value != want {
			t.Fatal("native mux control refs", refs)
		}
	}
	if _, err := g.MuxControlReferenceCounts(g.ObjectCount() - 1); err == nil {
		t.Fatal("object cap")
	}
	if _, err := g.MuxControlReferenceCounts(-1); err == nil {
		t.Fatal("negative cap")
	}
	if _, _, _, err := g.RecognizeXor(GiaID(g.ObjectCount())); err == nil {
		t.Fatal("invalid ID")
	}
	if _, _, ok, err := g.RecognizeXor(mux.ID()); err != nil || ok {
		t.Fatal("mux classified as xor")
	}
}
