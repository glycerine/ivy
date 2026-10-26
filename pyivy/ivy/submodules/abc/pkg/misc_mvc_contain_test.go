package abc

// Tests for src/misc/mvc/mvcContain.c and mvcSort.c.

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

type mvcContainOracle struct {
	Cases []struct {
		Bits    int
		Orders  []string
		Changed int
		Contain string
	} `json:"cases"`
}

func mvcFillSortCube(cube *MvcCube, id int) error {
	for bit := 0; bit < cube.cover.BitCount(); bit++ {
		include := bit < 4 && id>>uint(bit)&1 != 0
		if bit >= 4 {
			include = (bit*13+id*7+id*bit)%11 < 5
		}
		if include {
			if err := cube.SetBit(bit, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func mvcFillContainCube(cube *MvcCube, id int) error {
	for bit := 0; bit < cube.cover.BitCount(); bit++ {
		include := false
		switch id {
		case 0, 1:
			include = bit%4 < 2
		case 2:
			include = bit%4 < 2 && bit != 0
		case 3:
			include = bit%4 >= 2
		case 4:
			include = bit%4 >= 2 && bit != 2
		case 5:
			include = bit%2 == 0
		case 6:
			include = bit%2 == 0 && bit != 0
		case 7:
			include = false
		case 8, 9:
			include = bit%5 < 2
		case 10:
			include = bit%2 == 1
		default:
			include = (bit*7+3)%11 < 4
		}
		if include {
			if err := cube.SetBit(bit, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func mvcPatternCover(manager *MvcManager, bitCount int, contain bool) (*MvcCover, []*MvcCube, error) {
	cover, err := NewMvcCover(manager, bitCount)
	if err != nil {
		return nil, nil, err
	}
	original := make([]*MvcCube, 12)
	for id := range original {
		original[id], err = cover.NewCube()
		if err != nil {
			return nil, nil, err
		}
		if contain {
			err = mvcFillContainCube(original[id], id)
		} else {
			err = mvcFillSortCube(original[id], id)
		}
		if err != nil {
			return nil, nil, err
		}
		if err = cover.AddTail(original[id]); err != nil {
			return nil, nil, err
		}
	}
	return cover, original, nil
}

func mvcIdentityOrder(cover *MvcCover, original []*MvcCube) (string, error) {
	cubes, err := cover.Cubes()
	if err != nil {
		return "", err
	}
	result := make([]byte, len(cubes))
	for position, cube := range cubes {
		id := slices.Index(original, cube)
		if id < 0 || id >= 12 {
			return "", ErrInvariant
		}
		if id < 10 {
			result[position] = byte('0' + id)
		} else {
			result[position] = byte('A' + id - 10)
		}
	}
	return string(result), nil
}

func mvcFillSortMask(mask *MvcCube) error {
	for bit := 0; bit < mask.cover.BitCount(); bit++ {
		if (bit+mask.cover.BitCount())%3 == 1 {
			if err := mask.SetBit(bit, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func TestMvcSortAndContainNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/mvc_contain.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle mvcContainOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Cases) != 12 {
		t.Fatal(len(oracle.Cases), err)
	}
	manager := NewMvcManager()
	for test, record := range oracle.Cases {
		if len(record.Orders) != 5 {
			t.Fatal(test, record.Orders)
		}
		for order := MvcOrderInteger; order <= MvcOrderOutsideAndUnderMask; order++ {
			cover, original, err := mvcPatternCover(manager, record.Bits, false)
			if err != nil {
				t.Fatal(test, order, err)
			}
			mask, _ := cover.NewCube()
			if err = mvcFillSortMask(mask); err != nil {
				t.Fatal(test, order, err)
			}
			if err = cover.Sort(order, mask); err != nil {
				t.Fatal(test, order, err)
			}
			got, err := mvcIdentityOrder(cover, original)
			if err != nil || got != record.Orders[int(order)] {
				t.Fatal(test, order, got, record.Orders[int(order)], err)
			}
			if err = mask.Close(); err != nil {
				t.Fatal(test, order, err)
			}
			if err = cover.Close(); err != nil {
				t.Fatal(test, order, err)
			}
		}

		cover, original, err := mvcPatternCover(manager, record.Bits, true)
		if err != nil {
			t.Fatal(test, err)
		}
		changed, err := cover.RemoveContained()
		got, orderErr := mvcIdentityOrder(cover, original)
		if err != nil || orderErr != nil || changed != (record.Changed != 0) || got != record.Contain {
			t.Fatal(test, changed, got, record.Contain, err, orderErr)
		}
		live, _ := cover.Cubes()
		for _, cube := range original {
			isLive := slices.Contains(live, cube)
			if isLive == cube.closed {
				t.Fatal(test, "removed cube lifecycle", cube)
			}
		}
		changedAgain, err := cover.RemoveContained()
		if err != nil || changedAgain {
			t.Fatal(test, changedAgain, err)
		}
		if err = cover.Close(); err != nil {
			t.Fatal(test, err)
		}
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMvcSortErrorsAndSmallContainment(t *testing.T) {
	manager := NewMvcManager()
	cover, _ := NewMvcCover(manager, 5)
	cube, _ := cover.NewCube()
	_ = cover.AddTail(cube)
	original, _ := cover.Cubes()
	if err := cover.Sort(MvcCubeOrder(99), nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := cover.Sort(MvcOrderUnderMask, nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	after, _ := cover.Cubes()
	if !slices.Equal(original, after) {
		t.Fatal("failed sort changed cover")
	}
	wide, _ := NewMvcCover(manager, 6)
	wideMask, _ := wide.NewCube()
	if err := cover.Sort(MvcOrderOutsideMask, wideMask); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if changed, err := cover.RemoveContained(); err != nil || changed {
		t.Fatal(changed, err)
	}
	empty, _ := NewMvcCover(manager, 0)
	for range 5 {
		item, _ := empty.NewCube()
		_ = empty.AddTail(item)
	}
	if changed, err := empty.RemoveContained(); err != nil || !changed || empty.CubeCount() != 1 {
		t.Fatal(changed, empty.CubeCount(), err)
	}
	if err := empty.Close(); err != nil {
		t.Fatal(err)
	}
	if err := wideMask.Close(); err != nil {
		t.Fatal(err)
	}
	if err := wide.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cover.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := cover.RemoveContained(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func FuzzMvcContain(f *testing.F) {
	f.Add(uint16(65), uint8(12), []byte{1, 2, 3, 4, 5})
	f.Add(uint16(0), uint8(4), []byte{})
	f.Fuzz(func(t *testing.T, rawBits uint16, rawCount uint8, data []byte) {
		bitCount := int(rawBits) % 257
		count := int(rawCount)%24 + 1
		manager := NewMvcManager()
		cover, err := NewMvcCover(manager, bitCount)
		if err != nil {
			t.Fatal(err)
		}
		originalWords := make([][]uint32, count)
		for index := 0; index < count; index++ {
			cube, _ := cover.NewCube()
			words := make([]uint32, cover.WordCount())
			for word := range words {
				for offset := 0; offset < 4; offset++ {
					value := byte(index*31 + word*13 + offset*7)
					if len(data) > 0 {
						value ^= data[(index+word*4+offset)%len(data)]
					}
					words[word] |= uint32(value) << uint(offset*8)
				}
			}
			if index > 0 && len(data) > 0 && data[index%len(data)]&3 == 0 {
				words = slices.Clone(originalWords[index-1])
			}
			if err = cube.SetWords(words); err != nil {
				t.Fatal(err)
			}
			originalWords[index], _ = cube.Words()
			if err = cover.AddTail(cube); err != nil {
				t.Fatal(err)
			}
		}
		changed, err := cover.RemoveContained()
		if err != nil {
			t.Fatal(err)
		}
		survivors, _ := cover.Cubes()
		for left := range survivors {
			leftWords, _ := survivors[left].Words()
			foundOriginal := false
			for _, words := range originalWords {
				foundOriginal = foundOriginal || slices.Equal(leftWords, words)
			}
			if !foundOriginal {
				t.Fatal("invented survivor")
			}
			for right := range survivors {
				if left == right {
					continue
				}
				rightWords, _ := survivors[right].Words()
				if slices.Equal(leftWords, rightWords) || mvcWordsSubset(leftWords, rightWords) {
					t.Fatal("non-maximal survivors", left, right)
				}
			}
		}
		for _, original := range originalWords {
			covered := false
			for _, survivor := range survivors {
				words, _ := survivor.Words()
				if mvcWordsSubset(original, words) {
					covered = true
					break
				}
			}
			if !covered {
				t.Fatal("lost maximal cover")
			}
		}
		if changed != (len(survivors) != count) {
			t.Fatal(changed, len(survivors), count)
		}
		if err = cover.Close(); err != nil {
			t.Fatal(err)
		}
		if err = manager.Close(); err != nil {
			t.Fatal(err)
		}
	})
}

func mvcWordsSubset(first, second []uint32) bool {
	for index := range first {
		if first[index]&^second[index] != 0 {
			return false
		}
	}
	return true
}
