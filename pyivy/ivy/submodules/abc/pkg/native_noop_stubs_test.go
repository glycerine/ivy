package abc

// Tests for observable constant/no-op native compatibility stubs.

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

type noopStubBatchOracle struct {
	Cases, Nulls, DistanceSum, CbaNulls, IiffMatchSum, IiffNulls int
}

func TestNativeNoopStubs(t *testing.T) {
	LegalizePlacement()
	AcbDataReadTest()
	WlcJSONTest()
	for _, test := range []struct {
		files                           [3]string
		timeout, words, beam, low, high int
		flags                           [5]bool
	}{
		{},
		{files: [3]string{"a", "b", "c"}, timeout: -1, words: 17, beam: 99, low: -8, high: 42, flags: [5]bool{true, false, true, true, false}},
	} {
		if got := GiaSim4Try(test.files[0], test.files[1], test.files[2], test.timeout, test.words, test.beam, test.low, test.high, test.flags[0], test.flags[1], test.flags[2], test.flags[3], test.flags[4]); got != 0 {
			t.Fatal(got)
		}
	}
}

func TestNativeNoopStubBatch(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/noop_stub_batch.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle noopStubBatchOracle
	if err = json.Unmarshal(data, &oracle); err != nil || oracle.Cases != 16 {
		t.Fatal(err, oracle)
	}
	SimInit(0)
	SimInit(^uintptr(0))
	SimEnd(1)
	GiaDatFree(0)
	GiaDatFree(^uintptr(0))
	manager := NewGiaManager()
	manager.Name, manager.Spec = "unchanged", "stub"
	ExtraCommandCfs(manager, -1, 17, 23, 31, true, false, true, 0.375, true)
	if manager.Name != "unchanged" || manager.Spec != "stub" || manager.ObjectCount() != 1 {
		t.Fatal(manager.Name, manager.Spec, manager.ObjectCount())
	}
	nulls, distanceSum := 0, 0
	cbaNulls, iiffMatchSum, iiffNulls := 0, 0, 0
	for seed := 0; seed < oracle.Cases; seed++ {
		if generated := ExtraCommandGen(seed-8, seed+1, seed*3, seed%5, seed*17, seed-2, seed&3, seed&1 != 0, seed&2 != 0); generated != nil {
			t.Fatal(seed, generated)
		} else {
			nulls++
		}
		if result := PlaManDist1Merge(uintptr(seed*19 + 1)); result != 0 {
			t.Fatal(seed, result)
		} else {
			distanceSum += result
		}
		MpmManTest()
		TestInit(uintptr(seed*31 + 1))
		TestEnd(uintptr(seed*31 + 1))
		files2 := [2]string{"", "right"}
		files3 := [3]string{"left", "", "right"}
		AcbNtkRunGen(files2, seed&1 != 0)
		AbcNtkEco(files3)
		if CbaManReadCba(files3[seed%3]) == 0 {
			cbaNulls++
		}
		CbaManWriteCba(files3[(seed+1)%3], uintptr(seed*29+1))
		truthMemories := [3]uintptr{uintptr(seed + 1), uintptr(seed*2 + 1), uintptr(seed*3 + 1)}
		matches2 := [3]uintptr{uintptr(seed*5 + 1), uintptr(seed*7 + 1), uintptr(seed*11 + 1)}
		iiffMatchSum += GiaManDeriveMatches(uintptr(seed+1), uintptr(seed*13+1), uintptr(seed*17+1), uintptr(seed*19+1), truthMemories, matches2)
		if GiaManIiffTest(files3[0], manager, seed-8, seed+1, seed&1 != 0, seed&2 != 0, seed&4 != 0, seed&8 != 0) == nil {
			iiffNulls++
		}
	}
	if nulls != oracle.Nulls || distanceSum != oracle.DistanceSum || cbaNulls != oracle.CbaNulls || iiffMatchSum != oracle.IiffMatchSum || iiffNulls != oracle.IiffNulls {
		t.Fatal(nulls, distanceSum, cbaNulls, iiffMatchSum, iiffNulls, oracle)
	}
}

func FuzzNativeNoopStubBatch(f *testing.F) {
	f.Add(int32(-1), int32(17), uint32(0x3e800000), uint8(7))
	f.Fuzz(func(t *testing.T, first, second int32, probabilityBits uint32, flags uint8) {
		probability := float32(probabilityBits) / float32(^uint32(0))
		manager := NewGiaManager()
		ExtraCommandCfs(manager, int(first), int(second), int(first^second), int(first+second), flags&1 != 0, flags&2 != 0, flags&4 != 0, probability, flags&8 != 0)
		if manager.ObjectCount() != 1 {
			t.Fatal(manager.ObjectCount())
		}
		if ExtraCommandGen(int(first), int(second), int(first>>1), int(second>>1), int(first^second), int(first+second), int(flags), flags&16 != 0, flags&32 != 0) != nil {
			t.Fatal("non-nil stub result")
		}
		if PlaManDist1Merge(uintptr(probabilityBits)) != 0 {
			t.Fatal("nonzero stub result")
		}
		SimInit(uintptr(probabilityBits))
		SimEnd(uintptr(probabilityBits))
		GiaDatFree(uintptr(probabilityBits))
		MpmManTest()
		TestInit(uintptr(probabilityBits))
		TestEnd(uintptr(probabilityBits))
		files2 := [2]string{"left", "right"}
		files3 := [3]string{"left", "middle", "right"}
		AcbNtkRunGen(files2, flags&64 != 0)
		AbcNtkEco(files3)
		if CbaManReadCba(files3[int(flags)%3]) != 0 {
			t.Fatal("nonzero CBA stub result")
		}
		CbaManWriteCba(files3[int(flags>>2)%3], uintptr(probabilityBits))
		opaque := [3]uintptr{uintptr(uint32(first)), uintptr(uint32(second)), uintptr(probabilityBits)}
		if GiaManDeriveMatches(uintptr(uint32(first)), uintptr(uint32(second)), uintptr(probabilityBits), uintptr(flags), opaque, opaque) != 0 {
			t.Fatal("nonzero IIFF match result")
		}
		if GiaManIiffTest(files3[0], manager, int(first), int(second), flags&1 != 0, flags&2 != 0, flags&4 != 0, flags&8 != 0) != nil {
			t.Fatal("non-nil IIFF stub result")
		}
	})
}

func TestSFMMitPlaceholders(t *testing.T) {
	if manager := SFMMitStart(1, 2, 3, 4, -5); manager != nil {
		t.Fatal(manager)
	}
	var manager *SFMMit
	SFMMitStop(manager)
	if SFMMitNetworkDelay(manager) != 0 || SFMMitNetworkMinimumSlack(manager) != 0 || SFMMitObjectDelay(manager, -7) != 0 {
		t.Fatal("nonzero timing placeholder")
	}
	SFMMitTransferLoad(manager, 11, 13)
	SFMMitTimingGrow(manager)
	timing := []int{5, -1, 5, 9}
	candidates := []int{8, 3, 8, -2}
	mapping := []int{7, 0, -1}
	fanins := []int{2, 4, 6}
	mffc := []int{1, 3, 5}
	gate1Fanins := []byte{3, 1, 4}
	gate2Fanins := []byte{1, 5, 9}
	timingBefore, candidatesBefore := slices.Clone(timing), slices.Clone(candidates)
	mappingBefore, faninsBefore, mffcBefore := slices.Clone(mapping), slices.Clone(fanins), slices.Clone(mffc)
	gate1Before, gate2Before := slices.Clone(gate1Fanins), slices.Clone(gate2Fanins)
	SFMMitUpdateLoad(manager, timing, true)
	SFMMitUpdateLoad(manager, timing, false)
	SFMMitUpdateTiming(manager, timing)
	if SFMMitSortByArrival(manager, candidates, -1) != 0 ||
		SFMMitPriorityNodes(manager, candidates, 17) != 0 ||
		SFMMitNodeIsNoncritical(manager, 19, 23) != 0 ||
		SFMMitEvaluateRemapping(manager, mffc, 29, fanins, mapping, 31, gate1Fanins, 37, gate2Fanins) != 0 {
		t.Fatal("nonzero decision placeholder")
	}
	if !slices.Equal(timing, timingBefore) || !slices.Equal(candidates, candidatesBefore) ||
		!slices.Equal(mapping, mappingBefore) || !slices.Equal(fanins, faninsBefore) || !slices.Equal(mffc, mffcBefore) ||
		!slices.Equal(gate1Fanins, gate1Before) || !slices.Equal(gate2Fanins, gate2Before) {
		t.Fatal("placeholder mutated caller storage")
	}
}

func FuzzSFMMitPlaceholders(f *testing.F) {
	f.Add(int32(-1), int32(17), uint64(0x123456789abcdef0), []byte{1, 2, 3, 4})
	f.Fuzz(func(t *testing.T, first, second int32, opaque uint64, bytes []byte) {
		if len(bytes) > 64 {
			bytes = bytes[:64]
		}
		values := make([]int, len(bytes))
		for index, value := range bytes {
			values[index] = int(value) - 128
		}
		valuesBefore, bytesBefore := slices.Clone(values), slices.Clone(bytes)
		manager := SFMMitStart(uintptr(opaque), uintptr(uint32(first)), uintptr(uint32(second)), uintptr(opaque>>1), int(first))
		if manager != nil || SFMMitNetworkDelay(manager) != 0 || SFMMitNetworkMinimumSlack(manager) != 0 ||
			SFMMitObjectDelay(manager, int(second)) != 0 || SFMMitSortByArrival(manager, values, int(first)) != 0 ||
			SFMMitPriorityNodes(manager, values, int(second)) != 0 || SFMMitNodeIsNoncritical(manager, uintptr(opaque), uintptr(opaque>>1)) != 0 ||
			SFMMitEvaluateRemapping(manager, values, uintptr(opaque), values, values, uintptr(opaque), bytes, uintptr(opaque>>1), bytes) != 0 {
			t.Fatal("placeholder result")
		}
		SFMMitTransferLoad(manager, uintptr(opaque), uintptr(opaque>>1))
		SFMMitTimingGrow(manager)
		SFMMitUpdateLoad(manager, values, first&1 != 0)
		SFMMitUpdateTiming(manager, values)
		SFMMitStop(manager)
		if !slices.Equal(values, valuesBefore) || !slices.Equal(bytes, bytesBefore) {
			t.Fatal("placeholder mutation")
		}
	})
}
