package abci

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

func TestSesStoreHashAndTruthBucketsPreserveSourceOrder(t *testing.T) {
	store := Ses_StoreAlloc(17, 0, 0)
	defer Ses_StoreClean(store)

	var first, second [4]word
	first[0], second[0] = 1, 2
	profiles := [][8]int32{{1, 2}, {1, 2}, {3, 2}}
	if got := Ses_StoreAddEntry(store, &first[0], 2, &profiles[0][0], "", 0); got != 1 {
		t.Fatalf("first insert = %d, want 1", got)
	}
	if got := Ses_StoreAddEntry(store, &first[0], 2, &profiles[1][0], "", 0); got != 0 {
		t.Fatalf("duplicate timing insert = %d, want 0", got)
	}
	if got := Ses_StoreAddEntry(store, &first[0], 2, &profiles[2][0], "", 1); got != 1 {
		t.Fatalf("new timing insert = %d, want 1", got)
	}
	if store.nEntriesCount != 2 || store.nUnsynthesizedImp != 2 || store.nUnsynthesizedRL != 1 {
		t.Fatalf("C insertion statistics = entries %d, unsynth-imp %d, unsynth-RL %d", store.nEntriesCount, store.nUnsynthesizedImp, store.nUnsynthesizedRL)
	}

	// Find a second truth with a matching bucket to check head insertion and
	// full truth comparison after a hash collision.
	firstKey := Ses_StoreTableHash(&first[0], 2)
	for candidate := word(3); ; candidate++ {
		second[0] = candidate
		if candidate != first[0] && Ses_StoreTableHash(&second[0], 2) == firstKey {
			break
		}
	}
	if got := Ses_StoreAddEntry(store, &second[0], 2, &profiles[0][0], "", 0); got != 1 {
		t.Fatalf("colliding truth insert = %d, want 1", got)
	}
	bucket := store.pEntries[firstKey]
	if bucket == nil || bucket.pTruth[0] != second[0] || bucket.next == nil || bucket.next.pTruth[0] != first[0] {
		t.Fatal("truth collision chain order/content differs from source head insertion")
	}
	var solution string
	if Ses_StoreGetEntrySimple(store, &first[0], 2, &profiles[0][0], &solution) != 1 || solution != "" {
		t.Fatal("simple lookup did not report the unsynthesized timing entry")
	}
}

func TestSesStoreDatabaseMatchesSourceRecordFixture(t *testing.T) {
	store := Ses_StoreAlloc(17, 0, 0)
	defer Ses_StoreClean(store)
	truth := [4]word{0x0123456789abcdef}
	arrival := [8]int32{1, 2}
	if Ses_StoreAddEntry(store, &truth[0], 2, &arrival[0], "", 0) != 1 {
		t.Fatal("failed to add unsynthesized fixture entry")
	}
	path := t.TempDir() + "/ses.bin"
	Ses_StoreWrite(store, path, 0, 0, 1, 0)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("test_vectors/ses_store_unsynth_v1.bin")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("serialized record differs from C format: got %x, want %x", got, want)
	}

	loaded := Ses_StoreAlloc(17, 0, 0)
	defer Ses_StoreClean(loaded)
	Ses_StoreRead(loaded, path, 0, 0, 1, 0)
	if loaded.nEntriesCount != 1 || loaded.nUnsynthesizedImp != 1 {
		t.Fatalf("loaded counts = entries %d, unsynth-imp %d, want 1/1", loaded.nEntriesCount, loaded.nUnsynthesizedImp)
	}
	var solution string
	if Ses_StoreGetEntrySimple(loaded, &truth[0], 2, &arrival[0], &solution) != 1 || solution != "" {
		t.Fatal("loaded database entry was not found")
	}
}

func TestSesStoreNormalizesAndLooksUpSynthesizedSolutions(t *testing.T) {
	store := Ses_StoreAlloc(17, 0, 0)
	defer Ses_StoreClean(store)
	truth := [4]word{0x8}
	arrival := [8]int32{2, 5}
	// Header: two vars, one function, one gate. The gate reads input 0 and 1.
	solutionBytes := []byte{2, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0}
	solution := string(solutionBytes)
	if Ses_StoreAddEntry(store, &truth[0], 2, &arrival[0], solution, 0) != 1 {
		t.Fatal("failed to add synthesized solution")
	}
	if arrival[0] != 0 || arrival[1] != 0 {
		t.Fatalf("normalized arrival profile = %v, want [0 0]", arrival[:2])
	}
	query := [8]int32{5, 8}
	var found string
	if Ses_StoreGetEntry(store, &truth[0], 2, &query[0], &found) != 1 || found != solution {
		t.Fatal("network-normalized lookup did not return the stored solution")
	}
}

func TestSesStoreWriteSynthesizedRecordAndCategoryFilter(t *testing.T) {
	store := Ses_StoreAlloc(17, 0, 0)
	defer Ses_StoreClean(store)
	truth := [4]word{0x0123456789abcdef, 0x1020304050607080}
	arrival := [8]int32{2, 5}
	// The source serializer writes this 11-byte network: 3-byte header,
	// one four-byte gate, two trailer bytes, then two output literals.
	network := string([]byte{2, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0})
	if Ses_StoreAddEntry(store, &truth[0], 2, &arrival[0], network, 0) != 1 {
		t.Fatal("failed to add synthesized entry")
	}

	path := t.TempDir() + "/synth.bin"
	Ses_StoreWrite(store, path, 1, 0, 0, 0)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var want bytes.Buffer
	put64 := func(v uint64) { _ = binary.Write(&want, binary.NativeEndian, v) }
	put32 := func(v int32) { _ = binary.Write(&want, binary.NativeEndian, v) }
	put64(1)
	put64(uint64(truth[0]))
	put64(uint64(truth[1]))
	put64(0)
	put64(0)
	put32(2)
	put32(0)
	put32(3)
	for range 6 {
		put32(0)
	}
	put32(0)
	want.WriteString(network)
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("synthesized database record differs from C field order: got %x, want %x", got, want.Bytes())
	}

	filtered := Ses_StoreAlloc(17, 0, 0)
	defer Ses_StoreClean(filtered)
	Ses_StoreRead(filtered, path, 0, 0, 1, 1)
	if filtered.nEntriesCount != 0 {
		t.Fatalf("synthesized record passed unsynthesized-only read filter: %d entries", filtered.nEntriesCount)
	}

	// The C routine's category selection emits a zero-count database when
	// this sole synthesized-improvement record is excluded.
	Ses_StoreWrite(store, path, 0, 0, 1, 1)
	got, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var zeroCount [8]byte
	if !bytes.Equal(got, zeroCount[:]) {
		t.Fatalf("excluded category output = %x, want eight zero count bytes", got)
	}
}
