package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaRecognizeMuxPatterns(t *testing.T) {
	pairs := [][2]int{}
	for a := 0; a < 4; a++ {
		for b := a + 1; b < 4; b++ {
			for phase := 0; phase < 4; phase++ {
				pairs = append(pairs, [2]int{2*a + phase%2, 2*b + phase/2})
			}
		}
	}
	raw, err := os.ReadFile("test_vectors/oracle/gia_mux.json")
	if err != nil {
		t.Fatal(err)
	}
	var decompositionRows [][3]int
	if err = json.Unmarshal(raw, &decompositionRows); err != nil || len(decompositionRows) != 576 {
		t.Fatal("mux decomposition dataset", err)
	}
	refsRaw, err := os.ReadFile("test_vectors/oracle/gia_mux_refs.json")
	if err != nil {
		t.Fatal(err)
	}
	var referenceRows [][]uint64
	if err = json.Unmarshal(refsRaw, &referenceRows); err != nil || len(referenceRows) != 576 {
		t.Fatal(err)
	}
	row := 0
	matches := 0
	for _, left := range pairs {
		for _, right := range pairs {
			g := NewGiaManager()
			inputs := make([]GiaLit, 4)
			for i := range inputs {
				var err error
				inputs[i], err = g.AppendCI()
				if err != nil {
					t.Fatal(err)
				}
			}
			literal := func(code int) GiaLit {
				l := inputs[code/2]
				if code%2 != 0 {
					l = l.Not()
				}
				return l
			}
			a, err := g.AppendAnd(literal(left[0]), literal(left[1]))
			if err != nil {
				t.Fatal(err)
			}
			b, err := g.AppendAnd(literal(right[0]), literal(right[1]))
			if err != nil {
				t.Fatal(err)
			}
			root, err := g.AppendAnd(a.Not(), b.Not())
			if err != nil {
				t.Fatal(err)
			}
			c, yes, no, ok, err := g.RecognizeMux(root.ID())
			if err != nil {
				t.Fatal(err)
			}
			expected := false
			for _, l := range left {
				for _, r := range right {
					if l == r^1 {
						expected = true
					}
				}
			}
			dataset := decompositionRows[row]
			counts, err := g.MuxControlReferenceCounts(0)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(counts, referenceRows[row]) {
				t.Fatal("mux reference dataset", left, right, counts, referenceRows[row])
			}
			x, y, isXor, err := g.RecognizeXor(root.ID())
			if err != nil {
				t.Fatal(err)
			}
			expectedXor := left[0] == right[0]^1 && left[1] == right[1]^1
			if isXor != expectedXor {
				t.Fatal("xor recognition", left, right)
			}
			row++
			if (dataset[0] >= 0) != ok {
				t.Fatal("dataset detection", left, right)
			}
			if ok && (int(c) != dataset[0] || int(yes) != dataset[1] || int(no) != dataset[2]) {
				t.Fatal("dataset decomposition", left, right, c, yes, no, dataset)
			}
			if ok != expected {
				t.Fatal("pattern detection", left, right, ok)
			}
			if !ok {
				continue
			}
			matches++
			for _, l := range []GiaLit{root, c, yes, no} {
				if _, err := g.AppendCO(l); err != nil {
					t.Fatal(err)
				}
			}
			if isXor {
				for _, l := range []GiaLit{x, y} {
					if _, err := g.AppendCO(l); err != nil {
						t.Fatal(err)
					}
				}
			}
			for assignment := 0; assignment < 16; assignment++ {
				bits := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
				values, err := g.Evaluate(bits)
				if err != nil {
					t.Fatal(err)
				}
				if isXor && values[0] != (values[4] != values[5]) {
					t.Fatal("xor decomposition semantics", left, right, assignment)
				}
				mux := values[3]
				if values[1] {
					mux = values[2]
				}
				if values[0] != mux {
					t.Fatal("mux semantics", left, right, assignment, c, yes, no)
				}
			}
		}
	}
	if matches == 0 {
		t.Fatal("no mux cases")
	}
}
func TestGiaRecognizeMuxAndInvalid(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	root, err := g.AppendMuxReal(a.Not(), b.Not(), c)
	if err != nil {
		t.Fatal(err)
	}
	control, yes, no, ok, err := g.RecognizeMux(root.ID())
	if err != nil || !ok {
		t.Fatal(control, yes, no, ok, err)
	}
	for _, l := range []GiaLit{root, control, yes, no} {
		if _, err := g.AppendCO(l); err != nil {
			t.Fatal(err)
		}
	}
	for assignment := 0; assignment < 8; assignment++ {
		values, err := g.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0})
		if err != nil {
			t.Fatal(err)
		}
		mux := values[3]
		if values[1] {
			mux = values[2]
		}
		if values[0] != mux {
			t.Fatal("mux phase semantics")
		}
	}
	if _, _, _, ok, err := g.RecognizeMux(a.ID()); err != nil || ok {
		t.Fatal(ok, err)
	}
	if _, _, _, _, err := g.RecognizeMux(GiaID(g.ObjectCount())); err == nil {
		t.Fatal("invalid ID")
	}
}
