package abc

// Tests for src/misc/mvc/mvcPrint.c.

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func mvcPrintCover(manager *MvcManager, bits int) (*MvcCover, error) {
	cover, err := NewMvcCover(manager, bits)
	if err != nil {
		return nil, err
	}
	for cubeIndex := 0; cubeIndex < 3; cubeIndex++ {
		cube, cubeErr := cover.NewCube()
		if cubeErr != nil {
			return nil, cubeErr
		}
		for variable := 0; variable < bits/2; variable++ {
			value := (variable+cubeIndex)%3 + 1
			bit := variable * 2
			_ = cube.SetBit(bit, value&1 != 0)
			_ = cube.SetBit(bit+1, value&2 != 0)
		}
		if bits&1 != 0 {
			_ = cube.SetBit(bits-1, cubeIndex&1 == 0)
		}
		if cubeIndex == 2 && bits >= 6 {
			_ = cube.SetBit(2, false)
			_ = cube.SetBit(3, false)
		}
		if cubeErr = cover.AddTail(cube); cubeErr != nil {
			return nil, cubeErr
		}
	}
	counts, err := cover.AllocateLiteralCounts()
	if err != nil {
		return nil, err
	}
	for index := range counts {
		counts[index] = index*3 - bits
	}
	return cover, nil
}

func TestMvcPrintNative(t *testing.T) {
	want, err := os.ReadFile("test_vectors/oracle/mvc_print.sha256")
	if err != nil {
		t.Fatal(err)
	}
	manager := NewMvcManager()
	var got strings.Builder
	for _, bits := range []int{0, 1, 2, 5, 31, 32, 33, 64, 65, 127} {
		cover, makeErr := mvcPrintCover(manager, bits)
		if makeErr != nil {
			t.Fatal(bits, makeErr)
		}
		got.WriteString("CASE ")
		got.WriteString(strconv.Itoa(bits))
		got.WriteString(" PLAIN\n")
		plain, plainErr := cover.PrintText()
		if plainErr != nil {
			t.Fatal(bits, plainErr)
		}
		got.WriteString(plain)
		got.WriteString("CASE ")
		got.WriteString(strconv.Itoa(bits))
		got.WriteString(" BINARY\n")
		binary, binaryErr := cover.PrintBinaryText()
		if binaryErr != nil {
			t.Fatal(bits, binaryErr)
		}
		got.WriteString(binary)
		if err = cover.Close(); err != nil {
			t.Fatal(bits, err)
		}
	}
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(got.String())))
	if digest != strings.TrimSpace(string(want)) {
		t.Fatalf("native print mismatch: got %s want %s\n%s", digest, strings.TrimSpace(string(want)), got.String())
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMvcPrintErrors(t *testing.T) {
	manager := NewMvcManager()
	cover, _ := NewMvcCover(manager, 2)
	cube, _ := cover.NewCube()
	if err := cube.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := cube.BitText(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err := cube.BinaryText(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := cover.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := cover.PrintText(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func FuzzMvcPrint(f *testing.F) {
	f.Add(uint16(65), []byte{1, 2, 3, 4})
	f.Add(uint16(0), []byte{})
	f.Fuzz(func(t *testing.T, rawBits uint16, data []byte) {
		bits := int(rawBits) % 514
		words := (bits + 31) / 32
		unused := words*32 - bits
		values := mvcWordsFromData(words, unused, 0, data, 0x53)
		manager := NewMvcManager()
		cover, err := mvcCoverFromWords(manager, bits, [][]uint32{values})
		if err != nil {
			t.Fatal(err)
		}
		cube := cover.cubes[0]
		plain, err := cube.BitText()
		if err != nil || len(plain) != bits {
			t.Fatal(len(plain), bits, err)
		}
		for bit := 0; bit < bits; bit++ {
			want := byte('0')
			if values[bit>>5]&(uint32(1)<<uint(bit&31)) != 0 {
				want = '1'
			}
			if plain[bit] != want {
				t.Fatal(bit, plain[bit], want)
			}
		}
		binary, err := cube.BinaryText()
		if err != nil {
			t.Fatal(err)
		}
		var model strings.Builder
		for variable := 0; variable < bits/2; variable++ {
			bit := variable * 2
			value := (values[bit>>5] >> uint(bit&31)) & 3
			if value == 0 {
				break
			}
			model.WriteByte([]byte{'?', '0', '1', '-'}[value])
		}
		if binary != model.String() {
			t.Fatal(binary, model.String())
		}
		if err = cover.Close(); err != nil {
			t.Fatal(err)
		}
		if err = manager.Close(); err != nil {
			t.Fatal(err)
		}
	})
}
