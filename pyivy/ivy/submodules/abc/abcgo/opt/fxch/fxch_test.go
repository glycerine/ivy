package fxch

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"testing"
)

func evaluate(c Cubes, output int32, input uint32) (bool, error) {
	state := map[int32]byte{}
	values := map[int32]bool{}
	owners := map[int32]bool{}
	for _, row := range c {
		owners[row[0]] = true
	}
	var eval func(int32) (bool, error)
	eval = func(id int32) (bool, error) {
		if !owners[id] {
			return input&(1<<uint(id)) != 0, nil
		}
		if state[id] == 1 {
			return false, fmt.Errorf("cycle at %d", id)
		}
		if state[id] == 2 {
			return values[id], nil
		}
		state[id] = 1
		value := false
		for _, row := range c {
			if row[0] != id {
				continue
			}
			v := true
			for _, l := range row[1:] {
				x, e := eval(l >> 1)
				if e != nil {
					return false, e
				}
				if l&1 != 0 {
					x = !x
				}
				v = v && x
			}
			value = value || v
		}
		state[id] = 2
		values[id] = value
		return value, nil
	}
	return eval(output)
}
func compare(t *testing.T, a, b Cubes, outputs []int32, n int) {
	t.Helper()
	for bits := 0; bits < 1<<uint(n); bits++ {
		for _, id := range outputs {
			x, e := evaluate(a, id, uint32(bits))
			if e != nil {
				t.Fatal(e)
			}
			y, e := evaluate(b, id, uint32(bits))
			if e != nil {
				t.Fatal(e)
			}
			if x != y {
				t.Fatalf("output%d bits%d got%v want%v\nold%v\nnew%v", id, bits, y, x, a, b)
			}
		}
	}
}
func TestNativeFixtures(t *testing.T) {
	for mode, name := range []string{"single", "double", "mux"} {
		c := Cubes{{8, 0, 2, 4}, {8, 0, 2, 6}, {8, 0, 2, 8}}
		if mode > 0 {
			c = Cubes{{8, 0, 4}, {8, 2, 4}, {9, 0, 6}, {9, 2, 6}}
		}
		if mode == 2 {
			c = Cubes{{8, 0, 2}, {8, 1, 4}, {9, 0, 2, 6}, {9, 1, 4, 6}}
		}
		before := append(Cubes(nil), c...)
		n, e := Extract(&c, 10, 1, io.Discard)
		if e != nil || n != 1 {
			t.Fatal(n, e)
		}
		var b strings.Builder
		for _, row := range c {
			for i, v := range row {
				if i > 0 {
					b.WriteByte(' ')
				}
				fmt.Fprint(&b, v)
			}
			b.WriteByte('\n')
		}
		want, e := os.ReadFile("test_vectors/" + name + ".txt")
		if e != nil {
			t.Fatal(e)
		}
		if b.String() != string(want) {
			t.Fatalf("%s got %swant%s", name, b.String(), want)
		}
		compare(t, before, c, []int32{8, 9}, 5)
	}
}
func TestRandomTruthPreservation(t *testing.T) {
	r := rand.New(rand.NewSource(456))
	for trial := 0; trial < 300; trial++ {
		var c Cubes
		for owner := int32(6); owner < 9; owner++ {
			seen := map[string]bool{}
			for k := 0; k < 8; k++ {
				row := []int32{owner}
				for v := int32(0); v < 5; v++ {
					phase := r.Intn(3)
					if phase != 2 {
						row = append(row, v*2+int32(phase))
					}
				}
				key := divKey(row)
				if !seen[key] {
					c = append(c, row)
					seen[key] = true
				}
			}
		}
		before := append(Cubes(nil), c...)
		n, e := Extract(&c, 10, 20, io.Discard)
		if e != nil || n > 20 {
			t.Fatal(trial, n, e)
		}
		compare(t, before, c, []int32{6, 7, 8}, 5)
	}
}

func TestRejectCyclesWithoutMutation(t *testing.T) {
	c := Cubes{{2, 6}, {3, 4}}
	before := append(Cubes(nil), c...)
	if _, err := Extract(&c, 4, 1, io.Discard); err == nil {
		t.Fatal("accepted combinational cycle")
	}
	if fmt.Sprint(c) != fmt.Sprint(before) {
		t.Fatal("changed rejected input")
	}
}
func TestHashRemovalUsesWholeDescriptor(t *testing.T) {
	p := Fxch_ManAlloc(Cubes{{0, 0, 2, 4}})
	p.outputs = []map[int32]bool{{8: true}}
	table := Fxch_SCHashTableCreate(p, 1)
	Fxch_SCHashTableInsert(table, p.Cubes, 7, 0, 1, 0, 0)
	Fxch_SCHashTableInsert(table, p.Cubes, 7, 0, 2, 0, 0)
	Fxch_SCHashTableRemove(table, p.Cubes, 7, 0, 2, 0, 0)
	if len(table.bins[7]) != 1 || table.bins[7][0].ILit0 != 1 {
		t.Fatal("removed wrong colliding same-cube subcube", table.bins)
	}
}

func TestMultiwordOutputsAndUnlimitedExtraction(t *testing.T) {
	var c Cubes
	var outputs []int32
	for owner := int32(10); owner < 80; owner++ {
		c = append(c, []int32{owner, 0, 2, 4}, []int32{owner, 0, 2, 6}, []int32{owner, 0, 2, 8})
		outputs = append(outputs, owner)
	}
	before := append(Cubes(nil), c...)
	count, e := Extract(&c, 80, 0, io.Discard)
	if e != nil || count < 1 {
		t.Fatal(count, e)
	}
	compare(t, before, c, outputs, 5)
}
func TestConstantsAndTransactionalValidation(t *testing.T) {
	c := Cubes{{6}, {7, 0}, {8, 1}}
	before := append(Cubes(nil), c...)
	if _, e := Extract(&c, 10, 0, io.Discard); e != nil {
		t.Fatal(e)
	}
	compare(t, before, c, []int32{6, 7, 8}, 1)
	bad := []Cubes{{{}}, {{9, 2, 0}}, {{9, 0, 1}}, {{9, 20}}, {{-1, 0}}}
	for _, c := range bad {
		before := fmt.Sprint(c)
		if _, e := Extract(&c, 10, 0, io.Discard); e == nil {
			t.Fatal("accepted", before)
		}
		if fmt.Sprint(c) != before {
			t.Fatal("mutated invalid input")
		}
	}
}
