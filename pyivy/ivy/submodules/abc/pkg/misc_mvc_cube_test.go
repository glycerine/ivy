package abc

// Tests for src/misc/mvc/mvcCover.c:Mvc_CoverAlloc and mvcCube.c.

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

type mvcCubeOracle struct {
	Cases []struct {
		Bits, Words, Unused int
		Before, After       string
	} `json:"cases"`
}

func mvcOracleWords(t testing.TB, encoded string, count int) []uint32 {
	t.Helper()
	data, err := hex.DecodeString(encoded)
	if err != nil || len(data) != count*4 {
		t.Fatal(len(data), count, err)
	}
	words := make([]uint32, count)
	for index := range words {
		words[index] = binary.BigEndian.Uint32(data[index*4:])
	}
	return words
}

func TestMvcCubeNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/mvc_cube.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle mvcCubeOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Cases) != 10 {
		t.Fatal(len(oracle.Cases), err)
	}
	manager := NewMvcManager()
	for index, record := range oracle.Cases {
		cover, err := NewMvcCover(manager, record.Bits)
		if err != nil || cover.BitCount() != record.Bits || cover.WordCount() != record.Words || cover.UnusedBits() != record.Unused {
			t.Fatal(index, cover, err)
		}
		cube, err := cover.NewCube()
		if err != nil {
			t.Fatal(index, err)
		}
		before := mvcOracleWords(t, record.Before, record.Words)
		if err = cube.SetWords(before); err != nil {
			t.Fatal(index, err)
		}
		copyCube, err := cube.Clone()
		if err != nil {
			t.Fatal(index, err)
		}
		if err = copyCube.RemoveDontCares(); err != nil {
			t.Fatal(index, err)
		}
		got, err := copyCube.Words()
		want := mvcOracleWords(t, record.After, record.Words)
		if err != nil || !slices.Equal(got, want) {
			t.Fatal(index, got, want, err)
		}
		if record.Bits > 0 {
			original, _ := copyCube.Bit(0)
			cube.SetBit(0, !original)
			retained, _ := copyCube.Bit(0)
			if retained != original {
				t.Fatal("clone aliases", index)
			}
		}
		if err = copyCube.Close(); err != nil {
			t.Fatal(index, err)
		}
		if err = cube.Close(); err != nil {
			t.Fatal(index, err)
		}
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMvcCubeErrors(t *testing.T) {
	manager := NewMvcManager()
	if cover, err := NewMvcCover(nil, 1); cover != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(cover, err)
	}
	if cover, err := NewMvcCover(manager, -1); cover != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(cover, err)
	}
	cover, err := NewMvcCover(manager, 33)
	if err != nil {
		t.Fatal(err)
	}
	cube, err := cover.NewCube()
	if err != nil {
		t.Fatal(err)
	}
	if err = cube.SetWords([]uint32{1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	for _, bit := range []int{-1, 33} {
		if _, err = cube.Bit(bit); !errors.Is(err, ErrDimension) {
			t.Fatal(bit, err)
		}
		if err = cube.SetBit(bit, true); !errors.Is(err, ErrDimension) {
			t.Fatal(bit, err)
		}
	}
	if err = cube.Close(); err != nil {
		t.Fatal(err)
	}
	if err = cube.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if words, err := cube.Words(); words != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(words, err)
	}
	if err = (*MvcCube)(nil).Close(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
	if cube, err := cover.NewCube(); cube != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(cube, err)
	}
}

func FuzzMvcCubeRemoveDontCares(f *testing.F) {
	f.Add(uint16(65), []byte{0xff, 0x00, 0x5a})
	f.Add(uint16(1), []byte{1})
	f.Fuzz(func(t *testing.T, rawBits uint16, data []byte) {
		bits := int(rawBits) % 513
		manager := NewMvcManager()
		cover, err := NewMvcCover(manager, bits)
		if err != nil {
			t.Fatal(err)
		}
		cube, err := cover.NewCube()
		if err != nil {
			t.Fatal(err)
		}
		words := make([]uint32, cover.WordCount())
		for index := range words {
			for offset := 0; offset < 4; offset++ {
				if len(data) > 0 {
					words[index] |= uint32(data[(index*4+offset)%len(data)]) << uint(offset*8)
				}
			}
		}
		if err = cube.SetWords(words); err != nil {
			t.Fatal(err)
		}
		original, _ := cube.Words()
		if err = cube.RemoveDontCares(); err != nil {
			t.Fatal(err)
		}
		for bit := 0; bit < bits; bit += 2 {
			first := original[bit>>5]&(uint32(1)<<uint(bit&31)) != 0
			second := false
			if bit+1 < bits {
				second = original[(bit+1)>>5]&(uint32(1)<<uint((bit+1)&31)) != 0
			}
			for offset, want := range []bool{first && !second, second && !first} {
				if bit+offset >= bits {
					continue
				}
				got, err := cube.Bit(bit + offset)
				if err != nil || got != want {
					t.Fatal(bit+offset, got, want, err)
				}
			}
		}
		if err = cube.Close(); err != nil {
			t.Fatal(err)
		}
		if err = manager.Close(); err != nil {
			t.Fatal(err)
		}
	})
}
