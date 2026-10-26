package abc

import (
	"encoding/binary"
	"slices"
	"testing"
)

func TestPdrABCVectorFormatAndBounds(t *testing.T) {
	vector := []int32{3, 2, 0, 3, 1, 4, 0, 3}
	inv, registers, err := ReadPdrABCVector(vector, 0)
	if err != nil || registers != 3 {
		t.Fatal(registers, err)
	}
	result, err := inv.ABCVector(registers, len(vector))
	if err != nil || !slices.Equal(result, vector) {
		t.Fatal(result, err)
	}
	vector[2] = 5
	result, err = inv.ABCVector(3, 0)
	if err != nil || result[2] != 0 {
		t.Fatal("input aliases invariant")
	}
	for _, bad := range [][]int32{nil, {}, {0}, {-1, 0}, {0, -1}, {1, 2, 0, 1}, {1, -1, 0}, {1, 1, -1, 1}, {1, 1, 2, 1}, {0, 0, 0}, {1, 2, 0, 1, 1}, {2, 0, 0}} {
		if _, _, err := ReadPdrABCVector(bad, 0); err == nil {
			t.Fatal("malformed vector accepted", bad)
		}
	}
	if _, _, err := ReadPdrABCVector(result, len(result)-1); err == nil {
		t.Fatal("input cap ignored")
	}
	if _, err := inv.ABCVector(3, len(result)-1); err == nil {
		t.Fatal("output cap ignored")
	}
	for _, data := range [][]int32{{0, 0}, {1, 0, 0}} {
		got, n, err := ReadPdrABCVector(data, 0)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := got.ABCVector(n, 0)
		if err != nil || !slices.Equal(data, encoded) {
			t.Fatal(encoded, err)
		}
	}
}
func FuzzPdrABCVector(f *testing.F) {
	f.Add([]byte{0, 0, 0, 0, 0, 0, 0, 0})
	for _, seed := range [][]int32{{1, 0, 0}, {2, 1, 0, 1, 3, 2}, {1, 2, 0, 3, 2}} {
		raw := make([]byte, len(seed)*4)
		for i, v := range seed {
			binary.LittleEndian.PutUint32(raw[i*4:], uint32(v))
		}
		f.Add(raw)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1024 {
			data = data[:1024]
		}
		vector := make([]int32, len(data)/4)
		for i := range vector {
			vector[i] = int32(binary.LittleEndian.Uint32(data[4*i:]))
		}
		inv, n, err := ReadPdrABCVector(vector, 256)
		if err != nil {
			return
		}
		encoded, err := inv.ABCVector(n, 256)
		if err != nil {
			t.Fatal(err)
		}
		again, m, err := ReadPdrABCVector(encoded, 256)
		if err != nil || m != n {
			t.Fatal(err)
		}
		twice, err := again.ABCVector(m, 256)
		if err != nil || !slices.Equal(encoded, twice) {
			t.Fatal("noncanonical roundtrip", err)
		}
	})
}
