package abc

// Tests for src/misc/mvc/mvcLits.c.

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

type mvcLiteralsOracle struct {
	Cases []struct {
		Bits      int
		Any       int
		AnyMask   int `json:"any_mask"`
		Best      int
		BestMask  int `json:"best_mask"`
		Worst     int
		WorstMask int `json:"worst_mask"`
		First     int
		Total     int
		BestCover int `json:"best_cover"`
	} `json:"cases"`
	OneLiteral []int `json:"one_literal"`
}

func mvcLiteralCover(manager *MvcManager, bitCount int) (*MvcCover, error) {
	cover, err := NewMvcCover(manager, bitCount)
	if err != nil {
		return nil, err
	}
	for cubeIndex := 0; cubeIndex < 6; cubeIndex++ {
		cube, err := cover.NewCube()
		if err != nil {
			return nil, err
		}
		for bit := 0; bit < bitCount; bit++ {
			if (bit*7+cubeIndex*5+bitCount)%13 < 5 {
				_ = cube.SetBit(bit, true)
			}
		}
		if bitCount > 0 && cubeIndex < 3 {
			_ = cube.SetBit(bitCount-1, true)
		}
		if err = cover.AddTail(cube); err != nil {
			return nil, err
		}
	}
	return cover, nil
}

func mvcLiteralMask(cover *MvcCover) (*MvcCube, error) {
	mask, err := cover.NewCube()
	if err != nil {
		return nil, err
	}
	for bit := 0; bit < cover.BitCount(); bit++ {
		if (bit+cover.BitCount())%3 != 0 {
			_ = mask.SetBit(bit, true)
		}
	}
	if cover.BitCount() > 0 {
		_ = mask.SetBit(cover.BitCount()-1, true)
	}
	return mask, nil
}

func TestMvcLiteralsNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/mvc_literals.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle mvcLiteralsOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Cases) != 11 || len(oracle.OneLiteral) != 5 {
		t.Fatal(len(oracle.Cases), oracle.OneLiteral, err)
	}
	manager := NewMvcManager()
	for test, record := range oracle.Cases {
		cover, err := mvcLiteralCover(manager, record.Bits)
		if err != nil {
			t.Fatal(test, err)
		}
		mask, err := mvcLiteralMask(cover)
		if err != nil {
			t.Fatal(test, err)
		}
		any, anyErr := cover.AnyRepeatedLiteral(nil)
		anyMask, anyMaskErr := cover.AnyRepeatedLiteral(mask)
		best, bestErr := cover.BestRepeatedLiteral(nil)
		bestMask, bestMaskErr := cover.BestRepeatedLiteral(mask)
		worst, worstErr := cover.WorstRepeatedLiteral(nil)
		worstMask, worstMaskErr := cover.WorstRepeatedLiteral(mask)
		first, firstErr := cover.FirstCubeFirstLiteral()
		total, totalErr := cover.LiteralCount()
		if anyErr != nil || anyMaskErr != nil || bestErr != nil || bestMaskErr != nil || worstErr != nil || worstMaskErr != nil || firstErr != nil || totalErr != nil {
			t.Fatal(test, anyErr, anyMaskErr, bestErr, bestMaskErr, worstErr, worstMaskErr, firstErr, totalErr)
		}
		got := [...]int{any, anyMask, best, bestMask, worst, worstMask, first, total}
		want := [...]int{record.Any, record.AnyMask, record.Best, record.BestMask, record.Worst, record.WorstMask, record.First, record.Total}
		if got != want {
			t.Fatal(test, got, want)
		}
		frequencies, err := cover.LiteralFrequencies()
		sum := 0
		for _, count := range frequencies {
			sum += count
		}
		if err != nil || len(frequencies) != record.Bits || sum != total {
			t.Fatal(test, len(frequencies), sum, total, err)
		}
		if record.Bits > 0 {
			simple, _ := cover.Clone()
			simpleCube, _ := simple.cloneCube(mask)
			_ = simple.AddTail(simpleCube)
			selected, err := cover.BestLiteralCover(simple)
			if err != nil {
				t.Fatal(test, err)
			}
			selectedBit, bitErr := selected.FirstCubeFirstLiteral()
			selectedTotal, totalErr := selected.LiteralCount()
			if bitErr != nil || totalErr != nil || selectedBit != record.BestCover || selectedTotal != 1 {
				t.Fatal(test, selectedBit, selectedTotal, record.BestCover, bitErr, totalErr)
			}
			if err = selected.Close(); err != nil {
				t.Fatal(test, err)
			}
			if err = simple.Close(); err != nil {
				t.Fatal(test, err)
			}
		} else if record.BestCover != -1 {
			t.Fatal(test, record.BestCover)
		}
		if err = mask.Close(); err != nil {
			t.Fatal(test, err)
		}
		if err = cover.Close(); err != nil {
			t.Fatal(test, err)
		}
	}

	cover, _ := NewMvcCover(manager, 2)
	value, _ := cover.IsOneLiteral()
	if value != (oracle.OneLiteral[0] != 0) {
		t.Fatal(value)
	}
	cube, _ := cover.NewCube()
	_ = cover.AddTail(cube)
	value, _ = cover.IsOneLiteral()
	if value != (oracle.OneLiteral[1] != 0) {
		t.Fatal(value)
	}
	_ = cube.SetBit(0, true)
	value, _ = cover.IsOneLiteral()
	if value != (oracle.OneLiteral[2] != 0) {
		t.Fatal(value)
	}
	_ = cube.SetBit(1, true)
	value, _ = cover.IsOneLiteral()
	if value != (oracle.OneLiteral[3] != 0) {
		t.Fatal(value)
	}
	second, _ := cover.NewCube()
	_ = cover.AddTail(second)
	value, _ = cover.IsOneLiteral()
	if value != (oracle.OneLiteral[4] != 0) {
		t.Fatal(value)
	}
	if err = cover.Close(); err != nil {
		t.Fatal(err)
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMvcLiteralErrors(t *testing.T) {
	manager := NewMvcManager()
	cover, _ := NewMvcCover(manager, 4)
	if first, err := cover.FirstCubeFirstLiteral(); err != nil || first != -1 {
		t.Fatal(first, err)
	}
	wide, _ := NewMvcCover(manager, 5)
	wideMask, _ := wide.NewCube()
	if literal, err := cover.BestRepeatedLiteral(wideMask); literal != -1 || !errors.Is(err, ErrDimension) {
		t.Fatal(literal, err)
	}
	simple, _ := cover.Clone()
	if result, err := cover.BestLiteralCover(simple); result != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
	zero, _ := cover.NewCube()
	_ = simple.AddTail(zero)
	if result, err := cover.BestLiteralCover(simple); result != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
	second, _ := simple.NewCube()
	_ = simple.AddTail(second)
	if result, err := cover.BestLiteralCover(simple); result != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
	if err := wideMask.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := cover.AnyRepeatedLiteral(wideMask); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := simple.Close(); err != nil {
		t.Fatal(err)
	}
	if err := wide.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cover.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := cover.LiteralCount(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err := (*MvcCover)(nil).IsOneLiteral(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func FuzzMvcLiterals(f *testing.F) {
	f.Add(uint16(65), uint8(8), []byte{1, 2, 3, 4}, []byte{0xf0, 0x0f})
	f.Add(uint16(0), uint8(0), []byte{}, []byte{})
	f.Fuzz(func(t *testing.T, rawBits uint16, rawCount uint8, data, maskData []byte) {
		bitCount := int(rawBits) % 257
		cubeCount := int(rawCount) % 24
		manager := NewMvcManager()
		cover, err := NewMvcCover(manager, bitCount)
		if err != nil {
			t.Fatal(err)
		}
		wantFrequencies := make([]int, bitCount)
		wantFirst := -1
		for cubeIndex := 0; cubeIndex < cubeCount; cubeIndex++ {
			cube, _ := cover.NewCube()
			for bit := 0; bit < bitCount; bit++ {
				value := byte(cubeIndex*31 + bit*7)
				if len(data) > 0 {
					value ^= data[(cubeIndex+bit)%len(data)]
				}
				if value&3 == 0 {
					_ = cube.SetBit(bit, true)
					wantFrequencies[bit]++
					if cubeIndex == 0 && wantFirst < 0 {
						wantFirst = bit
					}
				}
			}
			_ = cover.AddTail(cube)
		}
		mask, _ := cover.NewCube()
		selected := make([]bool, bitCount)
		for bit := 0; bit < bitCount; bit++ {
			value := byte(bit*17 + 3)
			if len(maskData) > 0 {
				value ^= maskData[bit%len(maskData)]
			}
			if value&1 != 0 {
				selected[bit] = true
				_ = mask.SetBit(bit, true)
			}
		}
		wantAny, wantBest, wantWorst := -1, -1, -1
		bestCount, worstCount, wantTotal := -1, int(^uint(0)>>1), 0
		for bit, count := range wantFrequencies {
			wantTotal += count
			if !selected[bit] {
				continue
			}
			if count > 1 {
				wantAny = bit
			}
			if count > bestCount {
				wantBest, bestCount = bit, count
			}
			if count >= 2 && count < worstCount {
				wantWorst, worstCount = bit, count
			}
		}
		if bestCount <= 1 {
			wantBest = -1
		}
		gotFrequencies, err := cover.LiteralFrequencies()
		if err != nil || len(gotFrequencies) != len(wantFrequencies) {
			t.Fatal(err)
		}
		for bit := range wantFrequencies {
			if gotFrequencies[bit] != wantFrequencies[bit] {
				t.Fatal(bit, gotFrequencies[bit], wantFrequencies[bit])
			}
		}
		gotAny, _ := cover.AnyRepeatedLiteral(mask)
		gotBest, _ := cover.BestRepeatedLiteral(mask)
		gotWorst, _ := cover.WorstRepeatedLiteral(mask)
		gotFirst, _ := cover.FirstCubeFirstLiteral()
		gotTotal, _ := cover.LiteralCount()
		gotOne, _ := cover.IsOneLiteral()
		wantOne := cubeCount == 1 && wantTotal <= 1
		if gotAny != wantAny || gotBest != wantBest || gotWorst != wantWorst || gotFirst != wantFirst || gotTotal != wantTotal || gotOne != wantOne {
			t.Fatal(gotAny, wantAny, gotBest, wantBest, gotWorst, wantWorst, gotFirst, wantFirst, gotTotal, wantTotal, gotOne, wantOne)
		}
		if wantBest >= 0 {
			simple, _ := cover.Clone()
			simpleMask, _ := simple.cloneCube(mask)
			_ = simple.AddTail(simpleMask)
			bestCover, err := cover.BestLiteralCover(simple)
			if err != nil {
				t.Fatal(err)
			}
			bit, _ := bestCover.FirstCubeFirstLiteral()
			if bit != wantBest {
				t.Fatal(bit, wantBest)
			}
			_ = bestCover.Close()
			_ = simple.Close()
		}
		if err = mask.Close(); err != nil {
			t.Fatal(err)
		}
		if err = cover.Close(); err != nil {
			t.Fatal(err)
		}
		if err = manager.Close(); err != nil {
			t.Fatal(err)
		}
	})
}
