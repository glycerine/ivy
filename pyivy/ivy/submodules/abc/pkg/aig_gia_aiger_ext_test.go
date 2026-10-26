package abc

// Go-only byte-model, round-trip, malformed-input, and fuzz checks for src/aig/gia/giaAigerExt.c.

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func TestGiaAigerEquivalenceExactBytesAndRoundTrip(t *testing.T) {
	records := []GiaEquivalenceRecord{
		{Node: 1, Representative: 0},
		{Node: 3, Representative: 0, Proved: true},
		{Node: 4, Representative: 2},
		{Node: 7, Representative: 2, Proved: true},
	}
	classes, err := NewGiaEquivalenceClasses(8, records, GiaAigerExtensionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := EncodeGiaAigerEquivalenceClasses(classes, GiaAigerExtensionOptions{})
	want := []byte{0, 0, 0, 10, 1, 4, 10, 5, 8, 14}
	if err != nil || !slices.Equal(encoded, want) {
		t.Fatal(encoded, want, err)
	}
	decoded, err := DecodeGiaAigerEquivalenceClasses(encoded, 8, GiaAigerExtensionOptions{})
	if err != nil || decoded.ObjectCount() != 8 || !slices.Equal(decoded.Records(), records) {
		t.Fatal(decoded, err)
	}
	copyRecords := decoded.Records()
	copyRecords[0].Node = 7
	if decoded.Records()[0].Node != 1 {
		t.Fatal("records alias storage")
	}
	empty, _ := NewGiaEquivalenceClasses(1, nil, GiaAigerExtensionOptions{})
	encoded, err = EncodeGiaAigerEquivalenceClasses(empty, GiaAigerExtensionOptions{})
	decoded, decodeErr := DecodeGiaAigerEquivalenceClasses(encoded, 1, GiaAigerExtensionOptions{})
	if err != nil || decodeErr != nil || len(decoded.Records()) != 0 || !slices.Equal(encoded, []byte{0, 0, 0, 5, 1}) {
		t.Fatal(encoded, decoded, err, decodeErr)
	}
}

func giaAigerExtensionMapping(t testing.TB) *GiaLUTMapping {
	t.Helper()
	mapping, err := NewGiaLUTMapping(8, []GiaLUTRecord{
		{Root: 4, Fanins: []GiaID{1, 3}, MuxID: 4},
		{Root: 7, Fanins: []GiaID{2, 4}, MuxID: 7},
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	return mapping
}

func equalGiaAigerMappingRecords(left, right *GiaLUTMapping) bool {
	if left == nil || right == nil || left.objects != right.objects || !slices.Equal(left.Roots(), right.Roots()) {
		return false
	}
	for _, root := range left.Roots() {
		a, _ := left.Record(root)
		b, _ := right.Record(root)
		if a.Root != b.Root || !slices.Equal(a.Fanins, b.Fanins) {
			return false
		}
	}
	return true
}

func TestGiaAigerMappingCodecsExactBytes(t *testing.T) {
	mapping := giaAigerExtensionMapping(t)
	compact, err := EncodeGiaAigerMapping(mapping, GiaAigerExtensionOptions{})
	wantCompact := []byte{0, 0, 0, 12, 2, 3, 5, 3, 2, 4, 5, 7}
	if err != nil || !slices.Equal(compact, wantCompact) {
		t.Fatal(compact, wantCompact, err)
	}
	decoded, err := DecodeGiaAigerMapping(compact, 8, GiaAigerExtensionOptions{})
	if err != nil || !equalGiaAigerMappingRecords(mapping, decoded) {
		t.Fatal(decoded, err)
	}

	simple, err := EncodeGiaAigerMappingSimple(mapping, GiaAigerExtensionOptions{})
	if err != nil || len(simple) != 4*len(mapping.Packed()) {
		t.Fatal(len(simple), err)
	}
	for index, value := range mapping.Packed() {
		if int(int32(binary.LittleEndian.Uint32(simple[4*index:]))) != value {
			t.Fatal(index, value)
		}
	}
	simpleDecoded, err := DecodeGiaAigerMappingSimple(simple, 8, GiaAigerExtensionOptions{})
	if err != nil || !slices.Equal(simpleDecoded.Packed(), mapping.Packed()) {
		t.Fatal(simpleDecoded, err)
	}

	documented, err := EncodeGiaAigerMappingDocumented(mapping, GiaAigerExtensionOptions{})
	wantWords := []int{2, 2, 4, 2, 1, 3, 7, 2, 2, 4}
	if err != nil || len(documented) != 4*len(wantWords) {
		t.Fatal(len(documented), err)
	}
	for index, value := range wantWords {
		if int(binary.BigEndian.Uint32(documented[4*index:])) != value {
			t.Fatal(index, value, documented)
		}
	}
	docDecoded, err := DecodeGiaAigerMappingDocumented(documented, 8, GiaAigerExtensionOptions{})
	if err != nil || !equalGiaAigerMappingRecords(mapping, docDecoded) {
		t.Fatal(docDecoded, err)
	}
}

func TestGiaAigerPackingSignedBigEndian(t *testing.T) {
	values := []int{2, 3, 7, 4, 1, -9, 0, 2147483647, -2147483648}
	encoded, err := EncodeGiaAigerPacking(values, GiaAigerExtensionOptions{})
	if err != nil || len(encoded) != 4*len(values) {
		t.Fatal(encoded, err)
	}
	for index, value := range values {
		if int(int32(binary.BigEndian.Uint32(encoded[4*index:]))) != value {
			t.Fatal(index, value)
		}
	}
	decoded, err := DecodeGiaAigerPacking(encoded, GiaAigerExtensionOptions{})
	if err != nil || !slices.Equal(decoded, values) {
		t.Fatal(decoded, err)
	}
	decoded[0] = 99
	again, _ := DecodeGiaAigerPacking(encoded, GiaAigerExtensionOptions{})
	if again[0] != values[0] {
		t.Fatal("decoded packing aliases input")
	}
}

func TestGiaAigerExtensionMalformedLimitsAndCancellation(t *testing.T) {
	if _, err := NewGiaEquivalenceClasses(0, nil, GiaAigerExtensionOptions{}); err == nil {
		t.Fatal("accepted zero objects")
	}
	for _, records := range [][]GiaEquivalenceRecord{
		{{Node: 2, Representative: 2}},
		{{Node: 3, Representative: 1}, {Node: 3, Representative: 2}},
		{{Node: 8, Representative: 0}},
	} {
		if _, err := NewGiaEquivalenceClasses(8, records, GiaAigerExtensionOptions{}); err == nil {
			t.Fatal(records)
		}
	}
	mapping := giaAigerExtensionMapping(t)
	compact, _ := EncodeGiaAigerMapping(mapping, GiaAigerExtensionOptions{})
	equivalence, _ := NewGiaEquivalenceClasses(8, []GiaEquivalenceRecord{{Node: 2, Representative: 1}}, GiaAigerExtensionOptions{})
	equivalenceBytes, _ := EncodeGiaAigerEquivalenceClasses(equivalence, GiaAigerExtensionOptions{})
	malformedBlocks := [][]byte{
		nil, {0, 0, 0, 3}, {0, 0, 0, 6, 0x80, 0x00}, {0, 0, 0, 9, 0xff, 0xff, 0xff, 0xff, 0x10},
	}
	for _, data := range malformedBlocks {
		if _, err := DecodeGiaAigerEquivalenceClasses(data, 8, GiaAigerExtensionOptions{}); err == nil {
			t.Fatal(data)
		}
	}
	badLength := slices.Clone(compact)
	badLength[3]--
	if _, err := DecodeGiaAigerMapping(badLength, 8, GiaAigerExtensionOptions{}); err == nil {
		t.Fatal("accepted mismatched length")
	}
	underflow := []byte{0, 0, 0, 7, 0, 2, 0}
	if _, err := DecodeGiaAigerMapping(underflow, 8, GiaAigerExtensionOptions{}); err == nil {
		t.Fatal("accepted delta underflow")
	}
	if _, err := DecodeGiaAigerMappingSimple([]byte{1, 2, 3}, 8, GiaAigerExtensionOptions{}); err == nil {
		t.Fatal("accepted partial simple word")
	}
	documented, _ := EncodeGiaAigerMappingDocumented(mapping, GiaAigerExtensionOptions{})
	brokenDocumented := slices.Clone(documented)
	binary.BigEndian.PutUint32(brokenDocumented[4:], 3)
	if _, err := DecodeGiaAigerMappingDocumented(brokenDocumented, 8, GiaAigerExtensionOptions{}); err == nil {
		t.Fatal("accepted incorrect documented maximum")
	}
	if _, err := DecodeGiaAigerPacking([]byte{1}, GiaAigerExtensionOptions{}); err == nil {
		t.Fatal("accepted partial packing word")
	}
	for _, options := range []GiaAigerExtensionOptions{{MaxObjects: -1}, {MaxEntries: -1}, {MaxBytes: -1}, {MaxObjects: 1}, {MaxEntries: 1}, {MaxBytes: 4}} {
		if _, err := EncodeGiaAigerMapping(mapping, options); err == nil {
			t.Fatal(options)
		}
	}
	for threshold := 1; threshold <= 8; threshold++ {
		calls := 0
		_, err := DecodeGiaAigerEquivalenceClasses(equivalenceBytes, 8, GiaAigerExtensionOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if err != nil && !errors.Is(err, ErrCancelled) {
			t.Fatal(threshold, calls, err)
		}
		calls = 0
		_, err = DecodeGiaAigerMapping(compact, 8, GiaAigerExtensionOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if err != nil && !errors.Is(err, ErrCancelled) {
			t.Fatal(threshold, calls, err)
		}
	}
}

func FuzzGiaAigerEquivalenceCodec(f *testing.F) {
	f.Add(uint64(1))
	f.Add(uint64(0x9e3779b97f4a7c15))
	f.Fuzz(func(t *testing.T, seed uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		objects := 1 + random.Intn(40)
		records := make([]GiaEquivalenceRecord, 0)
		for node := 1; node < objects; node++ {
			if random.Intn(3) == 0 {
				records = append(records, GiaEquivalenceRecord{Node: GiaID(node), Representative: GiaID(random.Intn(node)), Proved: random.Intn(2) != 0})
			}
		}
		classes, err := NewGiaEquivalenceClasses(objects, records, GiaAigerExtensionOptions{MaxObjects: 128, MaxEntries: 256, MaxBytes: 4096})
		if err != nil {
			t.Fatal(seed, err)
		}
		data, err := EncodeGiaAigerEquivalenceClasses(classes, GiaAigerExtensionOptions{MaxObjects: 128, MaxEntries: 256, MaxBytes: 4096})
		if err != nil {
			t.Fatal(seed, err)
		}
		decoded, err := DecodeGiaAigerEquivalenceClasses(data, objects, GiaAigerExtensionOptions{MaxObjects: 128, MaxEntries: 256, MaxBytes: 4096})
		if err != nil || !slices.Equal(decoded.Records(), classes.Records()) {
			t.Fatal(seed, decoded, err)
		}
		if len(data) > 5 {
			mutated := slices.Clone(data)
			mutated[random.Intn(len(mutated))] ^= byte(1 + random.Intn(255))
			_, _ = DecodeGiaAigerEquivalenceClasses(mutated, objects, GiaAigerExtensionOptions{MaxObjects: 128, MaxEntries: 256, MaxBytes: 4096})
		}
	})
}

func FuzzGiaAigerMappingCodecs(f *testing.F) {
	f.Add(uint64(1))
	f.Add(uint64(0xd1b54a32d192ed03))
	f.Fuzz(func(t *testing.T, seed uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		objects := 2 + random.Intn(30)
		records := make([]GiaLUTRecord, 0)
		for root := 1; root < objects; root++ {
			if random.Intn(3) != 0 {
				continue
			}
			count := random.Intn(min(root, 5) + 1)
			fanins := make([]GiaID, count)
			for index := range fanins {
				fanins[index] = GiaID(random.Intn(root))
			}
			records = append(records, GiaLUTRecord{Root: GiaID(root), Fanins: fanins, MuxID: root})
		}
		mapping, err := NewGiaLUTMapping(objects, records, 512)
		if err != nil {
			t.Fatal(seed, err)
		}
		options := GiaAigerExtensionOptions{MaxObjects: 128, MaxEntries: 512, MaxBytes: 8192}
		compact, err := EncodeGiaAigerMapping(mapping, options)
		if err != nil {
			t.Fatal(seed, err)
		}
		fromCompact, err := DecodeGiaAigerMapping(compact, objects, options)
		if err != nil || !equalGiaAigerMappingRecords(mapping, fromCompact) {
			t.Fatal(seed, fromCompact, err)
		}
		simple, err := EncodeGiaAigerMappingSimple(mapping, options)
		if err != nil {
			t.Fatal(seed, err)
		}
		fromSimple, err := DecodeGiaAigerMappingSimple(simple, objects, options)
		if err != nil || !slices.Equal(mapping.Packed(), fromSimple.Packed()) {
			t.Fatal(seed, fromSimple, err)
		}
		documented, err := EncodeGiaAigerMappingDocumented(mapping, options)
		if err != nil {
			t.Fatal(seed, err)
		}
		fromDocumented, err := DecodeGiaAigerMappingDocumented(documented, objects, options)
		if err != nil || !equalGiaAigerMappingRecords(mapping, fromDocumented) {
			t.Fatal(seed, fromDocumented, err)
		}
		if len(compact) > 4 {
			mutated := slices.Clone(compact)
			mutated[4+random.Intn(len(mutated)-4)] ^= byte(1 + random.Intn(255))
			_, _ = DecodeGiaAigerMapping(mutated, objects, options)
		}
	})
}
