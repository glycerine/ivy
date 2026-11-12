package exor

import (
	"bytes"
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func testManager(t *testing.T, ni, no int) *Manager {
	t.Helper()
	m, err := NewManager(ni, no, Options{})
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func truth(t *testing.T, m *Manager) [][]bool {
	t.Helper()
	out := make([][]bool, 1<<uint(m.Inputs))
	for mask := range out {
		in := make([]bool, m.Inputs)
		for i := range in {
			in[i] = mask&(1<<uint(i)) != 0
		}
		var err error
		out[mask], err = m.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
	}
	return out
}
func enumerateCube(code, n, outputs int) *Cube {
	in := make([]VarValue, n)
	for i := range in {
		in[i] = VarValue(code%3 + 1)
		code /= 3
	}
	out := make([]bool, 2)
	out[0] = outputs&1 != 0
	out[1] = outputs&2 != 0
	return NewCube(in, out)
}
func TestExorLinkExhaustive(t *testing.T) {
	for n := 0; n <= 3; n++ {
		limit := 1
		for i := 0; i < n; i++ {
			limit *= 3
		}
		for a := 0; a < limit; a++ {
			for b := 0; b < limit; b++ {
				for oa := 1; oa < 4; oa++ {
					for ob := 1; ob < 4; ob++ {
						ca, cb := enumerateCube(a, n, oa), enumerateCube(b, n, ob)
						m := testManager(t, n, 2)
						m.Cubes = []*Cube{ca, cb}
						want := truth(t, m)
						d := len(FindDiffVars(ca, cb))
						groups := ExorLink(ca, cb, 4)
						factorial := 1
						for i := 1; i <= d; i++ {
							factorial *= i
						}
						if d == 0 {
							factorial = 0
						}
						if len(groups) != factorial {
							t.Fatalf("distance %d groups %d", d, len(groups))
						}
						for _, group := range groups {
							m.Cubes = group
							if got := truth(t, m); !reflect.DeepEqual(want, got) {
								t.Fatalf("bad link n=%d a=%d b=%d outputs=%d/%d", n, a, b, oa, ob)
							}
						}
					}
				}
			}
		}
	}
}
func TestMinimizeRandomCovers(t *testing.T) {
	r := rand.New(rand.NewSource(191))
	for trial := 0; trial < 80; trial++ {
		m := testManager(t, 4, 2)
		for k := 0; k < 8; k++ {
			m.Add(enumerateCube(r.Intn(81), 4, 1+r.Intn(3)))
		}
		want := truth(t, m)
		cost := m.cost(m.Cubes)
		if err := m.Minimize(); err != nil {
			t.Fatal(err)
		}
		if got := truth(t, m); !reflect.DeepEqual(want, got) {
			t.Fatalf("changed cover trial %d", trial)
		}
		if m.better(cost, m.cost(m.Cubes)) {
			t.Fatal("cover worsened")
		}
		again := cloneCubes(m.Cubes)
		m.Minimize()
		if !reflect.DeepEqual(m.cost(again), m.cost(m.Cubes)) {
			t.Fatal("strict minimization not at fixed point")
		}
	}
}
func TestNormalizeCancellationAndDistanceOne(t *testing.T) {
	m := testManager(t, 2, 2)
	m.Add(NewCube([]VarValue{VAR_NEG, VAR_ABS}, []bool{true, true}))
	m.Add(NewCube([]VarValue{VAR_POS, VAR_ABS}, []bool{true, true}))
	m.Normalize()
	if len(m.Cubes) != 1 || literalCount(m.Cubes[0]) != 0 {
		t.Fatal("complementary pair not merged")
	}
	m.Add(m.Cubes[0])
	m.Normalize()
	if len(m.Cubes) != 0 {
		t.Fatal("duplicate not canceled")
	}
}
func TestGraphRoundTripAllThreeInputFunctions(t *testing.T) {
	for table := 0; table < 256; table++ {
		g := aig.NewGraph()
		in := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
		f := g.Const0()
		for mask := 0; mask < 8; mask++ {
			if table&(1<<uint(mask)) == 0 {
				continue
			}
			term := g.Const1()
			for i, e := range in {
				if mask&(1<<uint(i)) == 0 {
					e = e.Not()
				}
				term = g.AND(term, e)
			}
			f = g.OR(f, term)
		}
		g.AddOutput(f)
		g.AddOutput(f.Not())
		m, err := FromGraph(g, 3)
		if err != nil {
			t.Fatal(err)
		}
		m.Minimize()
		rebuilt, err := m.Graph()
		if err != nil {
			t.Fatal(err)
		}
		for mask := 0; mask < 8; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			got, err := rebuilt.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			want := table&(1<<uint(mask)) != 0
			if got[0] != want || got[1] == want {
				t.Fatal("truth mismatch", table, mask, got)
			}
		}
	}
}
func TestPLAAndRows(t *testing.T) {
	rows := [][]int32{{0, 3, -1}, {2, -2}, {-1}}
	copyRows := make([][]int32, len(rows))
	for i := range rows {
		copyRows[i] = append([]int32(nil), rows[i]...)
	}
	m, err := Exorcism(rows, 2, 2, "")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rows, copyRows) {
		t.Fatal("mutated caller's rows")
	}
	var b bytes.Buffer
	if err = m.WritePLA(&b); err != nil {
		t.Fatal(err)
	}
	round, err := ReadPLA(&b)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(truth(t, m), truth(t, round)) {
		t.Fatal("PLA changed function")
	}
	for _, bad := range []string{".i 1\n.o 1\n.type f\n0 1\n.e\n", ".i 1\n.o 1\n.p 2\n0 1\n.e\n", ".i 1\n.o 1\n2 1\n"} {
		if _, err := ReadPLA(strings.NewReader(bad)); err == nil {
			t.Fatal("accepted malformed PLA")
		}
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, fmt.Errorf("write failed") }
func TestValidationAndIOErrors(t *testing.T) {
	m := testManager(t, 1, 1)
	m.Add(NewCube([]VarValue{VAR_POS}, []bool{true}))
	if err := m.WritePLA(failingWriter{}); err == nil {
		t.Fatal("lost writer error")
	}
	if err := AddCubesToStartingCover(m, [][]int32{{0, 1, -1}}); err == nil {
		t.Fatal("accepted contradictory literals")
	}
	if err := m.Add(&Cube{Inputs: []uint32{0}, Outputs: []uint32{1}}); err == nil {
		t.Fatal("accepted empty value set")
	}
	g, _ := m.Graph()
	if _, err := FromGraph(g, 0); err == nil {
		t.Fatal("ignored table budget")
	}
}
func TestPackedBoundariesAndQueueUndo(t *testing.T) {
	m := testManager(t, 33, 35)
	values := make([]VarValue, 33)
	for i := range values {
		values[i] = VarValue(1 + i%3)
	}
	out := make([]bool, 35)
	out[34] = true
	c := NewCube(values, out)
	for i, want := range values {
		if GetVar(c, i) != want {
			t.Fatal("packed value", i)
		}
	}
	m.Add(c)
	other := c.Clone()
	other.setVar(16, VAR_NEG)
	other.setVar(32, VAR_NEG)
	m.Add(other)
	a, b := IteratorCubePairStart(m, DIST2)
	if a == nil || b == nil {
		t.Fatal("pair not queued")
	}
	MarkSet(m)
	IteratorCubePairNext(m)
	MarkRewind(m)
	if GetQuequeStats(m, DIST2) != 1 {
		t.Fatal("queue stats")
	}
	before := truthSubset(m)
	if m.applyDistance(2) != 0 {
		UndoRecentChanges(m)
		if !reflect.DeepEqual(before, truthSubset(m)) {
			t.Fatal("undo failed")
		}
	}
}
func truthSubset(m *Manager) [][]bool {
	var result [][]bool
	for mask := 0; mask < 16; mask++ {
		in := make([]bool, m.Inputs)
		for i := range in {
			in[i] = mask&(1<<uint(i%4)) != 0
		}
		v, _ := m.Evaluate(in)
		result = append(result, v)
	}
	return result
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) / 2, nil }
func TestReviewShortWritesAndStalePairs(t *testing.T) {
	m := testManager(t, 2, 1)
	m.Add(NewCube([]VarValue{VAR_NEG, VAR_NEG}, []bool{true}))
	m.Add(NewCube([]VarValue{VAR_POS, VAR_POS}, []bool{true}))
	if err := m.WritePLA(shortWriter{}); err == nil {
		t.Error("short successful write was accepted")
	}
	a, b := IteratorCubePairStart(m, DIST2)
	if a == nil || b == nil {
		t.Fatal("missing pair")
	}
	m.Normalize()
	m.pairIndex = 0
	if a, b := IteratorCubePairNext(m); a != nil || b != nil {
		t.Error("stale queue pair survived cover replacement")
	}
	if _, err := ReadPLA(strings.NewReader(".i 2\n.o 1\n1- 1\n-1 1\n.e\n")); err == nil {
		t.Error("implicit SOP was silently interpreted as ESOP")
	}
}
