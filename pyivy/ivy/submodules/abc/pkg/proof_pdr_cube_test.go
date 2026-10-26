package abc

import (
	"encoding/json"
	"math"
	"os"
	"slices"
	"testing"
)

func pdrTestCube(t testing.TB, state, inputs []PdrLit) *PdrCube {
	t.Helper()
	c, err := NewPdrCube(state, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Check(); err != nil {
		t.Fatal(err)
	}
	return c
}
func pdrPartialCube(t testing.TB, code, variables int) *PdrCube {
	t.Helper()
	var state []PdrLit
	for i := 0; i < variables; i++ {
		choice := code % 3
		code /= 3
		if choice > 0 {
			state = append(state, PdrLit(2*i+choice-1))
		}
	}
	return pdrTestCube(t, state, nil)
}
func TestPdrCubeExhaustiveSets(t *testing.T) {
	cubes := make([]*PdrCube, 81)
	for i := range cubes {
		cubes[i] = pdrPartialCube(t, i, 4)
	}
	for _, a := range cubes {
		for _, b := range cubes {
			expected := true
			for _, l := range b.state {
				if !slices.Contains(a.state, l) {
					expected = false
				}
			}
			if a.Contains(b) != expected {
				t.Fatal("literal containment", a.state, b.state)
			}
			intersection, ok, err := a.Intersection(b, nil)
			if err != nil || !ok {
				t.Fatal(err)
			}
			for _, l := range intersection.state {
				if !slices.Contains(a.state, l) || !slices.Contains(b.state, l) {
					t.Fatal("intersection")
				}
			}
			for _, l := range a.state {
				if slices.Contains(b.state, l) && !slices.Contains(intersection.state, l) {
					t.Fatal("missing common literal")
				}
			}
			for assignment := 0; assignment < 16; assignment++ {
				values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
				am, _ := a.Matches(values)
				bm, _ := b.Matches(values)
				if expected && am && !bm {
					t.Fatal("containment direction")
				}
			}
		}
	}
	for _, cube := range cubes {
		for remove := -1; remove < len(cube.state); remove++ {
			got, err := cube.IntersectsZero(remove)
			if err != nil {
				t.Fatal(err)
			}
			want := true
			for i, l := range cube.state {
				if i != remove && l&1 == 0 {
					want = false
				}
			}
			if got != want {
				t.Fatal("zero-init intersection")
			}
		}
		for initialCode := 0; initialCode < 81; initialCode++ {
			initial := make([]LatchInit, 4)
			code := initialCode
			for i := range initial {
				initial[i] = LatchInit(code % 3)
				code /= 3
			}
			expected := false
			for assignment := 0; assignment < 16; assignment++ {
				values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
				matches := true
				for i, v := range initial {
					if v != LatchInitUnknown && values[i] != (v == LatchInitOne) {
						matches = false
					}
				}
				for _, l := range cube.state {
					if values[l.Index()] == l.Negative() {
						matches = false
					}
				}
				expected = expected || matches
			}
			actual, err := cube.IntersectsInitial(initial)
			if err != nil || actual != expected {
				t.Fatal("general initialization intersection", cube.state, initial, actual, expected, err)
			}
		}
	}
}
func TestPdrCubeOwnershipAndCanonicalization(t *testing.T) {
	state, inputs := []PdrLit{4, 1, 4}, []PdrLit{3, 0, 3}
	cube := pdrTestCube(t, state, inputs)
	state[0] = InvalidPdrLit
	inputs[0] = InvalidPdrLit
	if !slices.Equal(cube.StateLiterals(), []PdrLit{1, 4}) || !slices.Equal(cube.InputLiterals(), []PdrLit{3, 0}) || cube.LiteralCount() != 4 {
		t.Fatal("cube normalization")
	}
	returned := cube.StateLiterals()
	returned[0] = InvalidPdrLit
	returned = cube.InputLiterals()
	returned[0] = InvalidPdrLit
	clone := cube.Clone()
	clone.state[0] = InvalidPdrLit
	if cube.Check() != nil {
		t.Fatal("clone aliases")
	}
	without, err := cube.Without(0)
	if err != nil || !slices.Equal(without.state, []PdrLit{4}) || !slices.Equal(without.inputs, cube.inputs) {
		t.Fatal(without, err)
	}
	subset, err := cube.Subset([]PdrLit{4, 4})
	if err != nil || !slices.Equal(subset.state, []PdrLit{4}) {
		t.Fatal(subset, err)
	}
	if _, err := cube.Subset([]PdrLit{2}); err == nil {
		t.Fatal("subset introduces literal")
	}
	text, err := cube.StateString(4)
	if err != nil || text != "0-1-" {
		t.Fatal(text, err)
	}
	if _, err := cube.StateString(1); err == nil {
		t.Fatal("state display bounds")
	}
	// Different literals can have the same modulo-63 signature.
	a, b := pdrTestCube(t, []PdrLit{0}, nil), pdrTestCube(t, []PdrLit{63}, nil)
	if a.Signature() != b.Signature() || a.Contains(b) || b.Contains(a) {
		t.Fatal("signature collision mistaken for containment")
	}
	for _, bad := range [][]PdrLit{{0, 1}, {InvalidPdrLit}, {PdrLit(math.MaxInt32) + 1}} {
		if _, err := NewPdrCube(bad, nil); err == nil {
			t.Fatal("invalid state cube", bad)
		}
		if _, err := NewPdrCube(nil, bad); err == nil {
			t.Fatal("invalid PI cube", bad)
		}
	}
	for _, remove := range []int{-2, 2} {
		if _, err := cube.Without(remove); err == nil {
			t.Fatal(remove)
		}
		if _, err := cube.IntersectsZero(remove); err == nil {
			t.Fatal(remove)
		}
	}
	if _, err := cube.IntersectsInitial([]LatchInit{0}); err == nil {
		t.Fatal("initial bounds")
	}
	if _, err := cube.IntersectsInitial([]LatchInit{0, 0, LatchInit(4)}); err == nil {
		t.Fatal("initial enum")
	}
	if _, err := cube.Matches(nil); err == nil {
		t.Fatal("state bounds")
	}
	if _, err := MakePdrLit(MaxLiteralID+1, false); err == nil {
		t.Fatal("literal range")
	}
	l, err := MakePdrLit(MaxLiteralID, true)
	if err != nil || !l.Valid() || l.Index() != MaxLiteralID || l.Not().Not() != l {
		t.Fatal(l, err)
	}
	empty := pdrTestCube(t, nil, nil)
	if zero, err := empty.IntersectsZero(-1); err != nil || !zero {
		t.Fatal("empty cube initialization")
	}
	if !cube.Contains(empty) || empty.Contains(cube) {
		t.Fatal("empty set containment")
	}
	corrupt := cube.Clone()
	corrupt.signature ^= 1
	if corrupt.Check() == nil {
		t.Fatal("signature invariant")
	}
}
func TestPdrProtectedIntersectionTails(t *testing.T) {
	a, b := pdrTestCube(t, []PdrLit{2}, []PdrLit{0}), pdrTestCube(t, []PdrLit{2, 4}, nil)
	if result, ok, err := a.Intersection(b, []PdrLit{4}); err != nil || ok || result != nil {
		t.Fatal("protected tail dropped", result, ok, err)
	}
	if result, ok, err := b.Intersection(a, []PdrLit{4}); err != nil || ok || result != nil {
		t.Fatal("protected opposite tail dropped", result, ok, err)
	}
	result, ok, err := a.Intersection(b, []PdrLit{2, 100})
	if err != nil || !ok || !slices.Equal(result.state, []PdrLit{2}) || len(result.inputs) != 0 {
		t.Fatal(result, ok, err)
	}
	if _, _, err := a.Intersection(b, []PdrLit{InvalidPdrLit}); err == nil {
		t.Fatal("invalid protected literal")
	}
}
func TestPdrNativeUtilities(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/pdr_util.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cubes []struct {
			State, Inputs []PdrLit
			Signature     uint64
			Zero          bool
			Removed       struct {
				State, Inputs []PdrLit
				Signature     uint64
				Zero          bool
			}
			Subset struct {
				State, Inputs []PdrLit
				Signature     uint64
			}
		}
		Contains      [][]bool
		Compare       [][]int
		Queue         []struct{ Frame, Priority int }
		PopOrder      []int `json:"pop_order"`
		ProtectedTail struct {
			Left, Right, Protected []PdrLit
			CReturnedCube          bool `json:"c_returned_cube"`
		} `json:"protected_tail_observed"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cubes) != 32 || len(fixture.Queue) != 64 {
		t.Fatal("native PDR fixture size")
	}
	cubes := make([]*PdrCube, len(fixture.Cubes))
	for i, row := range fixture.Cubes {
		c := pdrTestCube(t, row.State, row.Inputs)
		cubes[i] = c
		if c.Signature() != row.Signature {
			t.Fatal("native signature", i)
		}
		zero, err := c.IntersectsZero(-1)
		if err != nil || zero != row.Zero {
			t.Fatal("native init test", i)
		}
		removed, err := c.Without(0)
		if err != nil || !slices.Equal(removed.state, row.Removed.State) || !slices.Equal(removed.inputs, row.Removed.Inputs) || removed.Signature() != row.Removed.Signature {
			t.Fatal("native removal", i, err)
		}
		zero, err = c.IntersectsZero(0)
		if err != nil || zero != row.Removed.Zero {
			t.Fatal("native remove/init test", i)
		}
		subset, err := c.Subset(row.Subset.State)
		if err != nil || subset.Signature() != row.Subset.Signature || !slices.Equal(subset.inputs, row.Subset.Inputs) {
			t.Fatal("native subset", i, err)
		}
	}
	for i, a := range cubes {
		for j, b := range cubes {
			if a.Contains(b) != fixture.Contains[i][j] || a.Compare(b) != fixture.Compare[i][j] {
				t.Fatal("native cube relation", i, j)
			}
		}
	}
	var q PdrQueue
	for i, row := range fixture.Queue {
		cube := pdrTestCube(t, []PdrLit{PdrLit(2 * i)}, nil)
		o, err := NewPdrObligation(row.Frame, row.Priority, cube, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := q.Push(o); err != nil {
			t.Fatal(err)
		}
	}
	for _, expected := range fixture.PopOrder {
		o := q.Pop()
		if o == nil || int(o.state.state[0].Index()) != expected {
			t.Fatal("native queue order", expected)
		}
	}
	if q.Len() != 0 || q.MaximumLen() != 64 || q.PushCount() != 64 {
		t.Fatal("native queue statistics")
	}
	observed := fixture.ProtectedTail
	if !observed.CReturnedCube {
		t.Fatal("missing native tail observation")
	}
	a, b := pdrTestCube(t, observed.Left, nil), pdrTestCube(t, observed.Right, nil)
	if _, ok, err := a.Intersection(b, observed.Protected); err != nil || ok {
		t.Fatal("protected tail semantics", err)
	}
}
func FuzzPdrCubes(f *testing.F) {
	for _, seed := range [][]byte{{}, {0, 1, 2, 3, 4, 5}, {255, 254, 253, 252}, {63, 0, 1, 64, 127}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 128 {
			return
		}
		var aa, bb []PdrLit
		for index, value := range data {
			switch value % 3 {
			case 1:
				aa = append(aa, PdrLit(2*index)+PdrLit((value>>2)&1))
			case 2:
				bb = append(bb, PdrLit(2*index)+PdrLit((value>>2)&1))
			}
			if value&32 != 0 {
				aa = append(aa, PdrLit(2*index)+PdrLit((value>>2)&1))
				bb = append(bb, PdrLit(2*index)+PdrLit((value>>2)&1))
			}
		}
		a, b := pdrTestCube(t, aa, nil), pdrTestCube(t, bb, nil)
		expected := true
		for _, l := range b.state {
			if !slices.Contains(a.state, l) {
				expected = false
			}
		}
		if a.Contains(b) != expected {
			t.Fatal("cube containment")
		}
		common, ok, err := a.Intersection(b, nil)
		if err != nil || !ok {
			t.Fatal(err)
		}
		for _, l := range aa {
			if slices.Contains(bb, l) != slices.Contains(common.state, l) {
				t.Fatal("cube intersection")
			}
		}
		for i := range a.state {
			without, err := a.Without(i)
			if err != nil || without.Check() != nil || !a.Contains(without) || without.StateCount() != a.StateCount()-1 {
				t.Fatal("cube reduction", err)
			}
		}
		if a.Compare(b) != -b.Compare(a) {
			t.Fatal("comparison antisymmetry")
		}
	})
}
