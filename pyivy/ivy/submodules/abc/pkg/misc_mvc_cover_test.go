package abc

// Tests for src/misc/mvc/mvcApi.c, mvcCover.c, and mvcList.c.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
)

type mvcCoverOracle struct {
	Cases []struct {
		Bits, Words, Unused         int
		Count                       int
		Forward, Reverse, Duplicate string
		CloneCount                  int `json:"clone_count"`
		MaskReuse                   int `json:"mask_reuse"`
		LitsReuse                   int `json:"lits_reuse"`
		AfterDelete                 int `json:"after_delete"`
		Empty, Taut, Buffer         int
		CreatedEmpty                int `json:"created_empty"`
		CreatedTaut                 int `json:"created_taut"`
	} `json:"cases"`
	RawList struct {
		Count  int
		TailOK int `json:"tail_ok"`
	} `json:"raw_list"`
}

func mvcCoverOrder(cubes []*MvcCube) (string, error) {
	var result strings.Builder
	for index, cube := range cubes {
		if index != 0 {
			result.WriteByte(':')
		}
		words, err := cube.Words()
		if err != nil {
			return "", err
		}
		for _, word := range words {
			fmt.Fprintf(&result, "%08x", word)
		}
	}
	return result.String(), nil
}

func mvcSetNativePattern(cube *MvcCube, seed uint32) error {
	words := make([]uint32, cube.cover.WordCount())
	for index := range words {
		words[index] = seed + 0x9e3779b9*uint32(index+1)
	}
	return cube.SetWords(words)
}

func TestMvcCoverNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/mvc_cover.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle mvcCoverOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Cases) != 5 || oracle.RawList.Count != 2 || oracle.RawList.TailOK != 1 {
		t.Fatal(oracle, err)
	}
	manager := NewMvcManager()
	for test, record := range oracle.Cases {
		cover, err := NewMvcCover(manager, record.Bits)
		if err != nil || cover.WordCount() != record.Words || cover.UnusedBits() != record.Unused {
			t.Fatal(test, err)
		}
		a, _ := cover.NewCube()
		b, _ := cover.NewCube()
		if err = mvcSetNativePattern(a, 0x01020304+uint32(test)); err != nil {
			t.Fatal(test, err)
		}
		if err = mvcSetNativePattern(b, 0x10203040+uint32(test)); err != nil {
			t.Fatal(test, err)
		}
		if err = cover.AddTail(a); err != nil {
			t.Fatal(test, err)
		}
		if err = cover.AddHead(b); err != nil {
			t.Fatal(test, err)
		}
		if _, err = cover.AddCloneTail(a); err != nil {
			t.Fatal(test, err)
		}
		if _, err = cover.AddCloneHead(b); err != nil {
			t.Fatal(test, err)
		}
		cubes, err := cover.Cubes()
		forward, orderErr := mvcCoverOrder(cubes)
		if err != nil || orderErr != nil || cover.CubeCount() != record.Count || forward != record.Forward {
			t.Fatal(test, forward, record.Forward, err, orderErr)
		}
		head, _ := cover.Head()
		tail, _ := cover.Tail()
		if head != cubes[0] || tail != cubes[len(cubes)-1] {
			t.Fatal(test, "head/tail")
		}

		array, err := cover.BuildCubeArray()
		if err != nil {
			t.Fatal(test, err)
		}
		slices.Reverse(array)
		if err = cover.ApplyCubeArray(array); err != nil {
			t.Fatal(test, err)
		}
		cubes, _ = cover.Cubes()
		reverse, err := mvcCoverOrder(cubes)
		if err != nil || reverse != record.Reverse {
			t.Fatal(test, reverse, record.Reverse, err)
		}

		clone, err := cover.Clone()
		if err != nil || clone.CubeCount() != record.CloneCount {
			t.Fatal(test, err)
		}
		duplicate, err := cover.Duplicate()
		if err != nil {
			t.Fatal(test, err)
		}
		duplicateCubes, _ := duplicate.Cubes()
		duplicateOrder, err := mvcCoverOrder(duplicateCubes)
		if err != nil || duplicateOrder != record.Duplicate {
			t.Fatal(test, duplicateOrder, record.Duplicate, err)
		}
		for index := range cubes {
			if cubes[index] == duplicateCubes[index] {
				t.Fatal(test, "shallow duplicate", index)
			}
		}

		mask, err := cover.AllocateMask()
		again, secondErr := cover.AllocateMask()
		if err != nil || secondErr != nil || mask != again || record.MaskReuse != 1 {
			t.Fatal(test, err, secondErr)
		}
		lits, err := cover.AllocateLiteralCounts()
		if err != nil || len(lits) != record.Bits {
			t.Fatal(test, len(lits), err)
		}
		if len(lits) > 0 {
			lits[0] = 17
		}
		litsAgain, err := cover.AllocateLiteralCounts()
		if err != nil || len(litsAgain) != len(lits) || len(lits) > 0 && litsAgain[0] != 17 || record.LitsReuse != 1 {
			t.Fatal(test, litsAgain, err)
		}
		if err = cover.DeallocateMask(); err != nil {
			t.Fatal(test, err)
		}
		if err = cover.DeallocateMask(); err != nil {
			t.Fatal(test, err)
		}
		if err = cover.DeallocateLiteralCounts(); err != nil {
			t.Fatal(test, err)
		}
		if err = cover.DeallocateLiteralCounts(); err != nil {
			t.Fatal(test, err)
		}

		cubes, _ = cover.Cubes()
		removed := cubes[1]
		if err = cover.Delete(removed); err != nil || cover.CubeCount() != record.AfterDelete {
			t.Fatal(test, err)
		}
		if err = cover.AddLiteralsOfCube(cubes[0]); err != nil {
			t.Fatal(test, err)
		}
		if err = cover.DeleteLiteralsOfCube(cubes[0]); err != nil {
			t.Fatal(test, err)
		}
		if err = removed.Close(); err != nil {
			t.Fatal(test, err)
		}
		if err = cover.MakeEmpty(); err != nil {
			t.Fatal(test, err)
		}
		emptyValue, _ := cover.IsEmpty()
		if emptyValue != (record.Empty != 0) {
			t.Fatal(test, emptyValue)
		}
		if err = cover.MakeTautology(); err != nil {
			t.Fatal(test, err)
		}
		tautValue, _ := cover.IsTautology()
		if tautValue != (record.Taut != 0) {
			t.Fatal(test, tautValue)
		}
		if record.Bits == 2 {
			if err = cover.MakeEmpty(); err != nil {
				t.Fatal(test, err)
			}
			buffer, _ := cover.NewCube()
			_ = buffer.SetBit(1, true)
			if err = cover.AddTail(buffer); err != nil {
				t.Fatal(test, err)
			}
		}
		bufferValue, _ := cover.IsBinaryBuffer()
		if bufferValue != (record.Buffer != 0) {
			t.Fatal(test, bufferValue)
		}
		emptyCover, err := cover.CreateEmpty()
		if err != nil {
			t.Fatal(test, err)
		}
		tautCover, err := cover.CreateTautology()
		if err != nil {
			t.Fatal(test, err)
		}
		createdEmpty, _ := emptyCover.IsEmpty()
		createdTaut, _ := tautCover.IsTautology()
		if createdEmpty != (record.CreatedEmpty != 0) || createdTaut != (record.CreatedTaut != 0) {
			t.Fatal(test, createdEmpty, createdTaut)
		}
		for _, candidate := range []*MvcCover{tautCover, emptyCover, duplicate, clone, cover} {
			if err = candidate.Close(); err != nil {
				t.Fatal(test, err)
			}
		}
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMvcCoverOwnershipErrors(t *testing.T) {
	manager := NewMvcManager()
	cover, _ := NewMvcCover(manager, 4)
	other, _ := NewMvcCover(manager, 4)
	wide, _ := NewMvcCover(manager, 5)
	cube, _ := cover.NewCube()
	foreign, _ := other.NewCube()
	tooWide, _ := wide.NewCube()
	if err := cover.AddTail(cube); err != nil {
		t.Fatal(err)
	}
	if err := cube.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := cover.AddHead(cube); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := cover.AddTail(foreign); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	mask, err := cover.AllocateMask()
	if err != nil {
		t.Fatal(err)
	}
	if err = cover.AddTail(mask); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = mask.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = cover.DeallocateMask(); err != nil {
		t.Fatal(err)
	}
	if clone, err := cover.AddCloneTail(tooWide); clone != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(clone, err)
	}
	cloned, err := cover.AddCloneTail(foreign)
	if err != nil || cloned.cover != cover {
		t.Fatal(cloned, err)
	}
	original, _ := cover.Cubes()
	if err = cover.ApplyCubeArray([]*MvcCube{cube, cube}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	after, _ := cover.Cubes()
	if !slices.Equal(original, after) {
		t.Fatal("invalid reorder was not atomic")
	}
	if err = cover.ApplyCubeArray(original[:1]); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = cover.Delete(foreign); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = cover.Close(); err != nil {
		t.Fatal(err)
	}
	if err = cover.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = cover.Cubes(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = foreign.Close(); err != nil {
		t.Fatal(err)
	}
	if err = tooWide.Close(); err != nil {
		t.Fatal(err)
	}
	if err = other.Close(); err != nil {
		t.Fatal(err)
	}
	if err = wide.Close(); err != nil {
		t.Fatal(err)
	}
	if err = (*MvcCover)(nil).Close(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
}

func FuzzMvcCoverOrdering(f *testing.F) {
	f.Add(uint16(65), uint8(7), []byte{3, 1, 4, 1, 5, 9})
	f.Add(uint16(0), uint8(1), []byte{})
	f.Fuzz(func(t *testing.T, rawBits uint16, rawCount uint8, data []byte) {
		bits := int(rawBits) % 257
		count := int(rawCount)%24 + 1
		manager := NewMvcManager()
		cover, err := NewMvcCover(manager, bits)
		if err != nil {
			t.Fatal(err)
		}
		for index := 0; index < count; index++ {
			cube, _ := cover.NewCube()
			words := make([]uint32, cover.WordCount())
			for word := range words {
				for offset := 0; offset < 4 && len(data) > 0; offset++ {
					words[word] |= uint32(data[(index*7+word*4+offset)%len(data)]) << uint(offset*8)
				}
			}
			if err = cube.SetWords(words); err != nil {
				t.Fatal(err)
			}
			if err = cover.AddTail(cube); err != nil {
				t.Fatal(err)
			}
		}
		order, err := cover.BuildCubeArray()
		if err != nil {
			t.Fatal(err)
		}
		for index := len(order) - 1; index > 0; index-- {
			choice := 0
			if len(data) > 0 {
				choice = int(data[index%len(data)]) % (index + 1)
			}
			order[index], order[choice] = order[choice], order[index]
		}
		if err = cover.ApplyCubeArray(order); err != nil {
			t.Fatal(err)
		}
		got, _ := cover.Cubes()
		if !slices.Equal(got, order) {
			t.Fatal("order mismatch")
		}
		duplicate, err := cover.Duplicate()
		if err != nil {
			t.Fatal(err)
		}
		copyCubes, _ := duplicate.Cubes()
		for index := range got {
			left, _ := got[index].Words()
			right, _ := copyCubes[index].Words()
			if got[index] == copyCubes[index] || !slices.Equal(left, right) {
				t.Fatal(index)
			}
		}
		if err = cover.MakeTautology(); err != nil {
			t.Fatal(err)
		}
		if tautology, err := cover.IsTautology(); err != nil || !tautology {
			t.Fatal(tautology, err)
		}
		if err = duplicate.Close(); err != nil {
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
