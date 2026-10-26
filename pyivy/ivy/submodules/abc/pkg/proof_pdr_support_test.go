package abc

import (
	"bytes"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestPdrSupportCountsAndColumnMap(t *testing.T) {
	a, _ := NewPdrCube([]PdrLit{2, 7}, nil)
	b, _ := NewPdrCube([]PdrLit{6}, nil)
	inv, _ := NewPdrInvariant([]*PdrCube{a, b, a})
	counts, err := inv.RegisterCounts(5, 0)
	if err != nil || !slices.Equal(counts, []uint64{0, 2, 0, 3, 0}) {
		t.Fatal(counts, err)
	}
	data, columns, err := inv.MarshalSupportPLA(5, 0, 0)
	if err != nil || !slices.Equal(columns, []int{1, 3}) || string(data) != ".i 2\n.o 1\n.p 3\n-1 1\n10 1\n10 1\n.e\n" {
		t.Fatal(string(data), columns, err)
	}
	counts[1] = 0
	columns[0] = 99
	again, err := inv.RegisterCounts(5, 0)
	if err != nil || again[1] != 2 {
		t.Fatal("mutable support alias")
	}
	if _, _, err := inv.MarshalSupportPLA(5, 0, len(data)-1); err == nil {
		t.Fatal("output bound ignored")
	}
	if _, err := inv.RegisterCounts(5, 4); err == nil {
		t.Fatal("register bound ignored")
	}
	if _, err := inv.RegisterCounts(3, 0); err == nil {
		t.Fatal("register reference ignored")
	}
	empty, _ := NewPdrInvariant(nil)
	data, columns, err = empty.MarshalSupportPLA(5, 0, 0)
	if err != nil || len(columns) != 0 || string(data) != ".i 0\n.o 1\n.p 0\n.e\n" {
		t.Fatal(string(data), columns, err)
	}
	cube, _ := NewPdrCube(nil, nil)
	all, _ := NewPdrInvariant([]*PdrCube{cube})
	data, columns, err = all.MarshalSupportPLA(5, 0, 0)
	if err != nil || len(columns) != 0 || string(data) != ".i 0\n.o 1\n.p 1\n 1\n.e\n" {
		t.Fatal(string(data), columns, err)
	}
}

func FuzzPdrSupportPLA(f *testing.F) {
	f.Add([]byte{0, 2, 17, 63})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 16 {
			data = data[:16]
		}
		cubes := make([]*PdrCube, 0, len(data))
		for _, code := range data {
			literals := []PdrLit{}
			for bit := 0; bit < 4; bit++ {
				digit := (code >> (2 * bit)) & 3
				if digit == 1 {
					literals = append(literals, PdrLit(4*bit))
				}
				if digit == 2 {
					literals = append(literals, PdrLit(4*bit+1))
				}
			}
			cube, err := NewPdrCube(literals, nil)
			if err != nil {
				t.Fatal(err)
			}
			cubes = append(cubes, cube)
		}
		inv, _ := NewPdrInvariant(cubes)
		encoded, columns, err := inv.MarshalSupportPLA(8, 8, 4096)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(string(encoded), "\n")
		for i, column := range columns {
			if column%2 != 0 || (i > 0 && column <= columns[i-1]) {
				t.Fatal("column map", columns)
			}
		}
		for state := 0; state < 256; state++ {
			bits := make([]bool, 8)
			for bit := range bits {
				bits[bit] = state&(1<<bit) != 0
			}
			blocked := false
			for _, row := range lines[3 : len(lines)-2] {
				match := true
				if len(row) != len(columns)+2 {
					t.Fatal("row width")
				}
				for col, reg := range columns {
					if row[col] != '-' && (row[col] == '1') != bits[reg] {
						match = false
					}
				}
				blocked = blocked || match
			}
			holds, err := inv.Holds(bits)
			if err != nil || holds == blocked {
				t.Fatal("compressed semantics", err)
			}
		}
	})
}

func TestPdrSupportAgainstNative(t *testing.T) {
	cubes := make([]*PdrCube, 0, 81)
	for code := 0; code < 81; code++ {
		value := code
		literals := []PdrLit{}
		for bit := 0; bit < 4; bit++ {
			digit := value % 3
			value /= 3
			if digit != 0 {
				l := PdrLit(4 * bit)
				if digit == 2 {
					l = l.Not()
				}
				literals = append(literals, l)
			}
		}
		cube, err := NewPdrCube(literals, nil)
		if err != nil {
			t.Fatal(err)
		}
		cubes = append(cubes, cube)
	}
	inv, _ := NewPdrInvariant(cubes)
	data, columns, err := inv.MarshalSupportPLA(8, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	native, err := os.ReadFile("test_vectors/oracle/pdr_support.pla")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, native) || !slices.Equal(columns, []int{0, 2, 4, 6}) {
		t.Fatal("native support mismatch")
	}
}
