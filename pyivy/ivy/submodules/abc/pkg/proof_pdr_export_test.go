package abc

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestPdrPLAOrderingPolarityAndBounds(t *testing.T) {
	a, _ := NewPdrCube([]PdrLit{0, 3}, nil)
	b, _ := NewPdrCube([]PdrLit{4}, nil)
	inv, _ := NewPdrInvariant([]*PdrCube{a, b})
	data, err := inv.MarshalPLA(3, 0)
	if err != nil || string(data) != ".i 3\n.o 1\n.p 2\n--1 1\n10- 1\n.e\n" {
		t.Fatal(string(data), err)
	}
	if inv.clauses[0].Compare(a) != 0 {
		t.Fatal("export reordered invariant")
	}
	for state := 0; state < 8; state++ {
		holds, err := inv.Holds([]bool{state&1 != 0, state&2 != 0, state&4 != 0})
		if err != nil {
			t.Fatal(err)
		}
		blocked := false
		for _, row := range []string{"--1", "10-"} {
			match := true
			for bit, column := range row {
				if column != '-' && (column == '1') != (state&(1<<bit) != 0) {
					match = false
				}
			}
			blocked = blocked || match
		}
		if holds == blocked {
			t.Fatal("PLA polarity", state)
		}
	}
	if _, err = inv.MarshalPLA(3, len(data)); err != nil {
		t.Fatal("exact size", err)
	}
	if _, err = inv.MarshalPLA(3, len(data)-1); err == nil {
		t.Fatal("output cap ignored")
	}
	for _, n := range []int{-1, 0, 1, 2} {
		if _, err = inv.MarshalPLA(n, 0); err == nil {
			t.Fatal("invalid width", n)
		}
	}
	if _, err = inv.MarshalPLA(3, -1); err == nil {
		t.Fatal("negative cap")
	}
	empty, _ := NewPdrInvariant(nil)
	data, err = empty.MarshalPLA(0, 0)
	if err != nil || string(data) != ".i 0\n.o 1\n.p 0\n.e\n" {
		t.Fatal(string(data), err)
	}
	cube, _ := NewPdrCube(nil, nil)
	falseInv, _ := NewPdrInvariant([]*PdrCube{cube})
	data, err = falseInv.MarshalPLA(0, 0)
	if err != nil || string(data) != ".i 0\n.o 1\n.p 1\n 1\n.e\n" {
		t.Fatal(string(data), err)
	}
}
func FuzzPdrPLAExport(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64 {
			data = data[:64]
		}
		cubes := make([]*PdrCube, 0, len(data))
		for _, value := range data {
			state := []PdrLit{}
			for bit := 0; bit < 4; bit++ {
				v := (int(value) >> (2 * bit)) & 3
				if v < 2 {
					state = append(state, PdrLit(2*bit+v))
				}
			}
			c, err := NewPdrCube(state, nil)
			if err != nil {
				t.Fatal(err)
			}
			cubes = append(cubes, c)
		}
		inv, err := NewPdrInvariant(cubes)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := inv.MarshalPLA(4, 4096)
		if err != nil {
			t.Fatal(err)
		}
		again, err := inv.MarshalPLA(4, 4096)
		if err != nil || !bytes.Equal(encoded, again) {
			t.Fatal("nondeterministic export")
		}
		lines := strings.Split(string(encoded), "\n")
		if lines[0] != ".i 4" || lines[1] != ".o 1" || lines[2] != ".p "+strconv.Itoa(len(cubes)) || lines[len(lines)-2] != ".e" {
			t.Fatal("header", lines)
		}
		for state := 0; state < 16; state++ {
			blocked := false
			for _, row := range lines[3 : len(lines)-2] {
				if len(row) != 6 || row[4:] != " 1" {
					t.Fatal("malformed row", row)
				}
				match := true
				for bit := 0; bit < 4; bit++ {
					if row[bit] != '-' && (row[bit] == '1') != (state&(1<<bit) != 0) {
						match = false
					}
				}
				blocked = blocked || match
			}
			holds, err := inv.Holds([]bool{state&1 != 0, state&2 != 0, state&4 != 0, state&8 != 0})
			if err != nil || holds == blocked {
				t.Fatal("export semantics", state, err)
			}
		}
	})
}

func TestPdrPLAAgainstNative(t *testing.T) {
	cubes := make([]*PdrCube, 0, 81)
	for code := 0; code < 81; code++ {
		value := code
		state := []PdrLit{}
		for bit := 0; bit < 4; bit++ {
			digit := value % 3
			value /= 3
			if digit != 0 {
				l := PdrLit(2 * bit)
				if digit == 2 {
					l = l.Not()
				}
				state = append(state, l)
			}
		}
		cube, err := NewPdrCube(state, nil)
		if err != nil {
			t.Fatal(err)
		}
		cubes = append(cubes, cube)
	}
	inv, err := NewPdrInvariant(cubes)
	if err != nil {
		t.Fatal(err)
	}
	data, err := inv.MarshalPLA(4, 0)
	if err != nil {
		t.Fatal(err)
	}
	native, err := os.ReadFile("test_vectors/oracle/pdr_export.pla")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, native) {
		t.Fatal("native PLA ordering or polarity mismatch")
	}
}
