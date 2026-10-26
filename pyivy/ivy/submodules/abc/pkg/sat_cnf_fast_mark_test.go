package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func fastMarkGraph(t testing.TB, ops [][4]uint32, outputs []uint32) *AigManager {
	t.Helper()
	g := NewAigManager()
	pool := []AigLit{0}
	for i := 0; i < 6; i++ {
		v, _ := g.CreateCI()
		pool = append(pool, v)
	}
	for _, op := range ops {
		lits := [3]AigLit{}
		for i := range lits {
			encoded := op[i+1] % uint32(2*len(pool))
			lits[i] = pool[encoded>>1].NotIf(encoded&1 != 0)
		}
		var v AigLit
		var e error
		switch op[0] % 3 {
		case 0:
			v, e = g.And(lits[0], lits[1])
		case 1:
			// The native CNF oracle deliberately sequences these temporaries
			// left-first to pin this fixture's object IDs. AigManager.Xor itself
			// follows aigOper.c's observed native call-expression order.
			var left, right AigLit
			left, e = g.And(lits[0], lits[1].Not())
			if e == nil {
				right, e = g.And(lits[0].Not(), lits[1])
			}
			if e == nil {
				v, e = g.Or(left, right)
			}
		case 2:
			var left, right AigLit
			left, e = g.And(lits[0], lits[1])
			if e == nil {
				right, e = g.And(lits[0].Not(), lits[2])
			}
			if e == nil {
				v, e = g.Or(left, right)
			}
		}
		if e != nil {
			t.Fatal(e)
		}
		pool = append(pool, v)
	}
	for _, out := range outputs {
		out %= uint32(2 * len(pool))
		g.CreateCO(pool[out>>1].NotIf(out&1 != 0))
	}
	return g
}
func checkAigMuxValues(t testing.TB, g *AigManager, root AigID, mux AigMuxInputs) {
	t.Helper()
	for assignment := 0; assignment < 64; assignment++ {
		values := make([]bool, len(g.objects))
		values[0] = true
		for i, id := range g.cis {
			values[id] = assignment>>uint(i)&1 != 0
		}
		value := func(l AigLit) bool { return values[l.ID()] != l.Complemented() }
		for id, obj := range g.objects {
			if obj.Type == AigAND {
				values[id] = value(obj.Fanin0) && value(obj.Fanin1)
			}
		}
		want := value(mux.WhenFalse)
		if values[mux.Control] {
			want = value(mux.WhenTrue)
		}
		if values[root] != want {
			t.Fatal(root, mux, assignment)
		}
	}
}
func TestFastCNFMarkNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_fast_mark.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Ops                             [][4]uint32
		Outputs                         []uint32
		Marked                          []bool
		OutputViolations, AndViolations int
		Muxes                           []struct {
			Root AigID
			AigMuxInputs
		}
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 32 {
		t.Fatal(len(rows))
	}
	muxCount := 0
	for index, row := range rows {
		g := fastMarkGraph(t, row.Ops, row.Outputs)
		before := slices.Clone(g.objects)
		r, e := MarkFastCNF(g, FastCNFConeOptions{})
		if e != nil || !slices.Equal(r.Marked, row.Marked) || r.OutputDriverViolations != row.OutputViolations || r.AndGateViolations != row.AndViolations {
			t.Fatal(index, r, row.Marked, e)
		}
		for _, expected := range row.Muxes {
			mux, ok, e := g.RecognizeMux(expected.Root)
			if e != nil || !ok || mux != expected.AigMuxInputs {
				t.Fatal(index, expected, mux, e)
			}
			checkAigMuxValues(t, g, expected.Root, mux)
			muxCount++
		}
		if !slices.Equal(before, g.objects) {
			t.Fatal("source changed")
		}
	}
	if muxCount == 0 {
		t.Fatal("no mux cases")
	}
}
func TestFastCNFMarkLimits(t *testing.T) {
	g := mappedCNFFixture(t, 0, 0)
	for _, o := range []FastCNFConeOptions{{MaxObjects: 1}, {MaxVisits: 1}, {Stop: func() bool { return true }}} {
		r, e := MarkFastCNF(g, o)
		if r != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(r, e)
		}
	}
	if _, ok, e := g.RecognizeMux(5); e != nil || ok {
		t.Fatal(ok, e)
	}
	if _, _, e := g.RecognizeMux(AigID(len(g.objects))); e == nil {
		t.Fatal("invalid root")
	}
}
func FuzzFastCNFMark(f *testing.F) {
	f.Add([]byte{0, 2, 4, 0, 1, 6, 8, 0, 2, 10, 14, 16})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 96 {
			data = data[:96]
		}
		var ops [][4]uint32
		for i := 0; i+3 < len(data); i += 4 {
			ops = append(ops, [4]uint32{uint32(data[i]), uint32(data[i+1]), uint32(data[i+2]), uint32(data[i+3])})
		}
		g := fastMarkGraph(t, ops, []uint32{uint32(2 * (6 + len(ops))), 3})
		before := slices.Clone(g.objects)
		r, e := MarkFastCNF(g, FastCNFConeOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if !slices.Equal(before, g.objects) {
			t.Fatal("source changed")
		}
		for _, id := range g.cis {
			if !r.Marked[id] {
				t.Fatal("unmarked CI")
			}
		}
		for id, obj := range g.objects {
			if obj.Type != AigAND {
				continue
			}
			mux, ok := aigRecognizeMux(g, AigID(id))
			if ok {
				checkAigMuxValues(t, g, AigID(id), mux)
			}
		}
	})
}
