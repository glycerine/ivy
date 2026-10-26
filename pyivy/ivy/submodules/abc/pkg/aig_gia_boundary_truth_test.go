package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaBoundaryTruth6Native(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_boundary_truth.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Root    GiaLit
		Word    uint64
		Leaves  []GiaID
		Objects [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 96 {
		t.Fatal(len(rows))
	}
	for i, row := range rows {
		m := NewGiaManager()
		for _, obj := range row.Objects[1:] {
			if obj[0] == 1 {
				_, err = m.AppendCI()
			} else {
				_, err = m.AppendAnd(GiaLit(obj[1]), GiaLit(obj[2]))
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		before := slices.Clone(m.objects)
		word, err := m.BoundaryTruth6(row.Root, row.Leaves, GiaCutTruthOptions{})
		if err != nil || word != row.Word {
			t.Fatal(i, word, row.Word, err)
		}
		if !slices.Equal(before, m.objects) {
			t.Fatal("source changed")
		}
	}
}
func TestGiaBoundaryTruth6DuplicatesAndLimits(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	root, _ := m.AppendAnd(a, b)
	leaves := []GiaID{a.ID(), b.ID(), a.ID()}
	word, err := m.BoundaryTruth6(root, leaves, GiaCutTruthOptions{})
	if err != nil || word != 0xc0c0c0c0c0c0c0c0 {
		t.Fatal(word, err)
	}
	first, err := m.CutTruth(root, leaves, GiaCutTruthOptions{})
	if err != nil || first.Table.words[0] != 0x8888888888888888 {
		t.Fatal(first, err)
	}
	for _, leaves := range [][]GiaID{nil, {a.ID()}, {0, b.ID()}, make([]GiaID, 7)} {
		if _, err := m.BoundaryTruth6(root, leaves, GiaCutTruthOptions{}); err == nil {
			t.Fatal("invalid boundary accepted", leaves)
		}
	}
	for _, options := range []GiaCutTruthOptions{{MaxObjects: -1}, {MaxMappingEntries: 1}, {MaxWords: -1}, {Stop: func() bool { return true }}} {
		if _, err := m.BoundaryTruth6(root, leaves, options); err == nil {
			t.Fatal(options)
		}
	}
	var missing *GiaManager
	if _, err := missing.BoundaryTruth6(root, leaves, GiaCutTruthOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	for _, root := range []GiaLit{GiaConst0, GiaConst1, a, a.Not()} {
		if _, err := m.BoundaryTruth6(root, leaves, GiaCutTruthOptions{}); err != nil {
			t.Fatal(root, err)
		}
	}
}
func FuzzGiaBoundaryTruth6(f *testing.F) {
	f.Add(uint8(0), uint8(0))
	f.Add(uint8(255), uint8(17))
	f.Fuzz(func(t *testing.T, flags, slots uint8) {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		x, _ := m.AppendXorReal(a, b.Not())
		y, _ := m.AppendMuxReal(c, x, a.Not())
		z, _ := m.AppendBuffer(y)
		root := z.NotIf(flags&1 != 0)
		if flags&2 != 0 {
			root, _ = m.AppendCO(root)
		}
		leaves := []GiaID{a.ID(), b.ID(), c.ID()}
		pool := []GiaID{a.ID(), b.ID(), c.ID(), x.ID(), y.ID(), z.ID()}
		for i := 0; i < int(flags%4); i++ {
			leaves = append(leaves, pool[(int(slots)+i)%len(pool)])
		}
		if flags&4 != 0 {
			slices.Reverse(leaves)
		}
		before := slices.Clone(m.objects)
		word, err := m.BoundaryTruth6(root, leaves, GiaCutTruthOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 64; assignment++ {
			values := make([]bool, m.ObjectCount())
			read := func(lit GiaLit) bool { return values[lit.ID()] != lit.Complemented() }
			for id := 1; id < m.ObjectCount(); id++ {
				slot := -1
				for i, leaf := range leaves {
					if int(leaf) == id {
						slot = i
					}
				}
				if slot >= 0 {
					values[id] = assignment&(1<<uint(slot)) != 0
					continue
				}
				obj := m.objects[id]
				if obj.IsCI() {
					continue
				}
				first, err := obj.Fanin0(GiaID(id))
				if err != nil {
					t.Fatal(err)
				}
				value := read(first)
				if obj.IsAnd() && !obj.IsBuffer() {
					second, _ := obj.Fanin1(GiaID(id))
					if control, mux := m.MuxControl(GiaID(id)); mux {
						if read(control) {
							value = read(second)
						}
					} else if obj.IsXor() {
						value = value != read(second)
					} else {
						value = value && read(second)
					}
				}
				values[id] = value
			}
			if (word>>uint(assignment)&1 != 0) != read(root) {
				t.Fatal("free-boundary assignment", assignment, leaves, word)
			}
		}
		if !slices.Equal(before, m.objects) {
			t.Fatal("source changed")
		}
	})
}
