package abc

// Go-only byte-layout, semantic-model, malformed-input, and fuzz tests for bblif.c.

import (
	"encoding/hex"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestBblHalfAdderAndText(t *testing.T) {
	manager, err := NewBblHalfAdder(BblLimits{})
	if err != nil {
		t.Fatal(err)
	}
	objects, err := manager.Objects()
	if err != nil || len(objects) != 7 {
		t.Fatal(len(objects), err)
	}
	for index, object := range objects {
		if object.ID != index+1 || object.OriginalID != index+1 {
			t.Fatal(index, object)
		}
	}
	if objects[0].Type != BblObjectCI || objects[3].Type != BblObjectCO || objects[5].Type != BblObjectNode || !slices.Equal(objects[5].Fanins, []int{1, 2, 3}) {
		t.Fatal(objects)
	}
	if objects[5].SOP != "001 1\n010 1\n100 1\n111 1\n" {
		t.Fatal(objects[5].SOP)
	}
	stats, err := manager.Stats()
	if err != nil || stats.Objects != 7 || stats.Nodes != 2 || stats.UniqueFunctions != 2 || stats.NameBytes != 8 {
		t.Fatal(stats, err)
	}
	functionBytes, err := manager.FunctionBytes()
	if err != nil || functionBytes != stats.FunctionBytes {
		t.Fatal(functionBytes, stats, err)
	}
	blif, err := manager.FormatBLIF()
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{".model hadder\n", ".inputs 1\n", ".outputs 4\n", ".names 1 2 3 6\n001 1\n", ".names 6 4\n1 1\n", ".end\n"} {
		if !strings.Contains(blif, fragment) {
			t.Fatalf("missing %q in\n%s", fragment, blif)
		}
	}
	formatted, err := manager.FormatStats()
	if err != nil || !strings.Contains(formatted, "Total objects =       7") || !strings.Contains(formatted, "Unique functions =       2") {
		t.Fatal(formatted, err)
	}
	copyObjects := objects
	copyObjects[5].Fanins[0] = 999
	again, _, err := manager.ObjectByID(6)
	if err != nil || again.Fanins[0] != 1 {
		t.Fatal(again, err)
	}
}

func TestBblExactBinaryLayout(t *testing.T) {
	manager, err := NewBblManager("x", BblLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for _, add := range []struct {
		kind       BblObjectType
		id, fanins int
		sop        string
	}{{BblObjectCI, 10, 0, ""}, {BblObjectNode, 11, 1, "1 1\n"}, {BblObjectCO, 12, 1, ""}} {
		if err = manager.AddObject(add.kind, add.id, add.fanins, add.sop); err != nil {
			t.Fatal(err)
		}
	}
	if err = manager.AddFanin(11, 10); err != nil {
		t.Fatal(err)
	}
	if err = manager.AddFanin(12, 11); err != nil {
		t.Fatal(err)
	}
	data, err := manager.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	wantHex := "04000000780000002c000000" +
		"0a000000ffffffff01000000" +
		"0b00000000000000100000000c000000" +
		"0c000000ffffffff1200000010000000" +
		"0c000000020000003120310a00000000"
	if got := hex.EncodeToString(data); got != wantHex {
		t.Fatalf("binary layout\ngot  %s\nwant %s", got, wantHex)
	}
	parsed, err := ParseBblBinary(data, BblLimits{})
	if err != nil {
		t.Fatal(err)
	}
	parsedObjects, err := parsed.Objects()
	if err != nil || len(parsedObjects) != 3 {
		t.Fatal(parsedObjects, err)
	}
	if parsedObjects[0].ID != 0 || parsedObjects[0].OriginalID != 10 || parsedObjects[1].ID != 1 || parsedObjects[1].OriginalID != 11 || !slices.Equal(parsedObjects[2].Fanins, []int{1}) {
		t.Fatal(parsedObjects)
	}
	second, err := parsed.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	reparsed, err := ParseBblBinary(second, BblLimits{})
	if err != nil {
		t.Fatal(err)
	}
	third, err := reparsed.MarshalBinary()
	if err != nil || !slices.Equal(second, third) {
		t.Fatal(err, hex.EncodeToString(second), hex.EncodeToString(third))
	}
}

func TestBblDedupForwardFaninAndMark(t *testing.T) {
	manager, _ := NewBblManager("dedup", BblLimits{})
	_ = manager.AddObject(BblObjectNode, 10, 1, "1 1\n0 1\n")
	_ = manager.AddObject(BblObjectCI, 20, 0, "")
	_ = manager.AddObject(BblObjectNode, 30, 1, "0 1\n1 1\n")
	_ = manager.AddObject(BblObjectCO, 40, 1, "")
	_ = manager.AddFanin(10, 20) // forward relative pointer
	_ = manager.AddFanin(30, 20)
	_ = manager.AddFanin(40, 30)
	if err := manager.Mark(2); err != nil {
		t.Fatal(err)
	}
	stats, err := manager.Stats()
	if err != nil || stats.UniqueFunctions != 1 {
		t.Fatal(stats, err)
	}
	data, err := manager.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseBblBinary(data, BblLimits{})
	if err != nil {
		t.Fatal(err)
	}
	objects, _ := parsed.Objects()
	if !objects[2].Marked || !slices.Equal(objects[0].Fanins, []int{1}) || objects[0].FunctionHandle != objects[2].FunctionHandle {
		t.Fatal(objects)
	}
}

func TestBblTruthSOPExhaustive(t *testing.T) {
	for variables := 0; variables <= 8; variables++ {
		words := 1
		if variables > 5 {
			words = 1 << uint(variables-5)
		}
		patterns := [][]uint32{make([]uint32, words), make([]uint32, words)}
		for index := range patterns[1] {
			patterns[1][index] = ^uint32(0)
		}
		pattern := make([]uint32, words)
		for bit := 0; bit < words*32; bit++ {
			if (bit*17+variables*13)%7 < 3 {
				pattern[bit>>5] |= uint32(1) << uint(bit&31)
			}
		}
		if variables < 5 {
			period := 1 << uint(variables)
			stretched := uint32(0)
			for bit := 0; bit < 32; bit++ {
				if pattern[0]&(uint32(1)<<uint(bit%period)) != 0 {
					stretched |= uint32(1) << uint(bit)
				}
			}
			pattern[0] = stretched
		}
		patterns = append(patterns, pattern)
		for _, truth := range patterns {
			sop, err := BblTruthToSOP(truth, variables, 0)
			if err != nil {
				t.Fatal(variables, err)
			}
			got, err := BblSOPToTruth(sop, variables, 0)
			if err != nil || !slices.Equal(got, truth) {
				t.Fatal(variables, sop, got, truth, err)
			}
		}
	}
	offset, err := BblSOPToTruth("0- 0\n-1 0\n", 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 32; assignment++ {
		matched := assignment&1 == 0 || assignment&2 != 0
		if got := offset[0]&(uint32(1)<<uint(assignment)) != 0; got != !matched {
			t.Fatal(assignment, got, !matched)
		}
	}
}

func TestBblValidationLimitsAndClose(t *testing.T) {
	manager, _ := NewBblManager("x", BblLimits{MaxObjects: 2, MaxObjectID: 10, MaxFanins: 2, MaxSOPBytes: 16})
	badAdds := []struct {
		kind       BblObjectType
		id, fanins int
		sop        string
	}{{BblObjectCI, 1, 1, ""}, {BblObjectCO, 1, 0, ""}, {BblObjectNode, 1, 1, "x 1\n"}, {BblObjectNone, 1, 0, ""}, {BblObjectCI, -1, 0, ""}, {BblObjectCI, 11, 0, ""}}
	for _, bad := range badAdds {
		if err := manager.AddObject(bad.kind, bad.id, bad.fanins, bad.sop); err == nil {
			t.Fatal("accepted", bad)
		}
	}
	if err := manager.AddObject(BblObjectCI, 1, 0, ""); err != nil {
		t.Fatal(err)
	}
	if err := manager.AddObject(BblObjectCI, 1, 0, ""); err == nil {
		t.Fatal("duplicate ID")
	}
	if err := manager.AddObject(BblObjectNode, 2, 1, "1 1\n"); err != nil {
		t.Fatal(err)
	}
	if err := manager.AddObject(BblObjectCI, 3, 0, ""); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if err := manager.Check(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := manager.AddFanin(2, 1); err != nil {
		t.Fatal(err)
	}
	if err := manager.AddFanin(2, 1); err == nil {
		t.Fatal("fanin overflow")
	}
	if err := manager.Check(); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Object(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := manager.Mark(99); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := manager.MarshalBinary(); err != nil {
		t.Fatal(err)
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Name(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := manager.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	stop := true
	cancelled, _ := NewBblManager("", BblLimits{Stop: func() bool { return stop }})
	if err := cancelled.AddObject(BblObjectCI, 0, 0, ""); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err := NewBblManager("x\x00y", BblLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewBblManager("x", BblLimits{MaxObjects: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	tinyBinary, _ := NewBblManager("x", BblLimits{MaxBinaryBytes: 27})
	if err := tinyBinary.AddObject(BblObjectCI, 0, 0, ""); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	tinyFunctions, _ := NewBblManager("", BblLimits{MaxSOPBytes: 4})
	if err := tinyFunctions.AddObject(BblObjectNode, 0, 1, "1 1\n"); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := BblTruthToSOP(nil, 17, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := BblSOPToTruth("", 1, 0); !errors.Is(err, ErrFormat) {
		t.Fatal(err)
	}
}

func TestBblRejectsMalformedBinary(t *testing.T) {
	manager, _ := NewBblHalfAdder(BblLimits{})
	data, _ := manager.MarshalBinary()
	for length := 0; length < len(data); length++ {
		if _, err := ParseBblBinary(data[:length], BblLimits{}); err == nil {
			t.Fatalf("accepted truncation %d/%d", length, len(data))
		}
	}
	if _, err := ParseBblBinary(append(slices.Clone(data), 0), BblLimits{}); err == nil {
		t.Fatal("accepted trailing byte")
	}
	if _, err := ParseBblBinary(data, BblLimits{MaxBinaryBytes: len(data) - 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	corruptions := [][]byte{}
	badName := slices.Clone(data)
	badName[4] = 'x'
	for index := 5; index < 12; index++ {
		badName[index] = 'y'
	}
	corruptions = append(corruptions, badName)
	badFunction := slices.Clone(data)
	// First object's function handle starts four bytes into object data.
	for index := 20; index < 24; index++ {
		badFunction[index] = 0
	}
	corruptions = append(corruptions, badFunction)
	for index, corrupt := range corruptions {
		if _, err := ParseBblBinary(corrupt, BblLimits{}); err == nil {
			t.Fatal("accepted corruption", index)
		}
	}
}

func FuzzBblTruthRoundTrip(f *testing.F) {
	for _, seed := range [][]byte{{}, {0}, {1, 2, 3}, {255, 0, 170, 85}, {7, 9, 11, 13, 15}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		variables := len(data) % 9
		words := 1
		if variables > 5 {
			words = 1 << uint(variables-5)
		}
		truth := make([]uint32, words)
		for index, value := range data {
			truth[index%words] ^= uint32(value) << uint((index*7)%25)
		}
		if variables < 5 {
			period := 1 << uint(variables)
			mask := uint32(0)
			for bit := 0; bit < 32; bit++ {
				if truth[0]&(uint32(1)<<uint(bit%period)) != 0 {
					mask |= uint32(1) << uint(bit)
				}
			}
			truth[0] = mask
		}
		sop, err := BblTruthToSOP(truth, variables, 1024)
		if err != nil {
			t.Fatal(err)
		}
		got, err := BblSOPToTruth(sop, variables, 1024)
		if err != nil || !slices.Equal(got, truth) {
			t.Fatal(variables, sop, got, truth, err)
		}
	})
}

func FuzzParseBblBinary(f *testing.F) {
	manager, _ := NewBblHalfAdder(BblLimits{})
	valid, _ := manager.MarshalBinary()
	f.Add(valid)
	f.Add([]byte{})
	f.Add([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 4096 {
			data = data[:4096]
		}
		manager, err := ParseBblBinary(data, BblLimits{MaxObjects: 128, MaxObjectID: 1024, MaxFanins: 128, MaxSOPBytes: 4096, MaxBinaryBytes: 4096})
		if err != nil {
			return
		}
		if err = manager.Check(); err != nil {
			t.Fatal(err)
		}
		encoded, err := manager.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		second, err := ParseBblBinary(encoded, BblLimits{MaxObjects: 128, MaxObjectID: 1024, MaxFanins: 128, MaxSOPBytes: 4096, MaxBinaryBytes: 4096})
		if err != nil {
			t.Fatal(err)
		}
		again, err := second.MarshalBinary()
		if err != nil || !slices.Equal(encoded, again) {
			t.Fatal(err)
		}
	})
}
