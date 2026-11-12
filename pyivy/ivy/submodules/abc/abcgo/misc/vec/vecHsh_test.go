package vec

import (
	"encoding/binary"
	"math/rand/v2"
	"slices"
	"strconv"
	"testing"
)

func TestHshSingleIntegerCollisionAndGrowth(t *testing.T) {
	p := Hsh_Int1ManStart(0)
	defer Hsh_Int1ManStop(p)
	// At least two values share the initial two buckets; force a non-head hit.
	values := []int32{0, -1, 17, -2147483648, 2147483647}
	reference := map[int32]int32{}
	for round := 0; round < 5; round++ {
		for _, value := range values {
			want, exists := reference[value]
			if !exists {
				want = int32(len(reference))
				reference[value] = want
			}
			if got := Hsh_Int1ManAdd(p, value); got != want {
				t.Fatalf("%d: got %d want %d", value, got, want)
			}
		}
	}
	random := rand.New(rand.NewPCG(10, 30))
	for i := 0; i < 3000; i++ {
		value := int32(random.IntN(1500)) - 750
		want, exists := reference[value]
		if !exists {
			want = int32(len(reference))
			reference[value] = want
		}
		if got := Hsh_Int1ManAdd(p, value); got != want {
			t.Fatalf("unstable ID: %d -> %d, want %d", value, got, want)
		}
	}
	for value, want := range reference {
		if got := Hsh_Int1ManLookup(p, value); got != want {
			t.Fatal(value, got, want)
		}
	}
	if Hsh_Int1ManLookup(p, 1234567) != -1 || Hsh_Int1ManEntryNum(p) != len(reference) {
		t.Fatal("lookup/count")
	}
}

func TestHshFixedRecords(t *testing.T) {
	data := Vec_IntAllocArrayCopy([]int32{12, 17, 13, 12, 15, 3, 16, 16, 12, 17, 12, 12}, 12)
	result := Hsh_IntManHashArray(data, 2)
	if !slices.Equal(Vec_IntArray(result), []int32{0, 1, 2, 3, 0, 4}) {
		t.Fatal(Vec_IntArray(result))
	}
	// Stop owns the hash objects, but must not release caller-owned source data.
	if Vec_IntSize(data) != 12 {
		t.Fatal("freed borrowed source")
	}
	manager := Hsh_IntManStart(data, 2, 0)
	for i := 0; i < 6; i++ {
		if got := Hsh_IntManAdd(manager, i); got != Vec_IntEntry(result, i) {
			t.Fatal(i, got)
		}
	}
	if *Hsh_IntManLookup(manager, []int32{12, 17}) != 0 || *Hsh_IntManLookup(manager, []int32{0, 0}) != -1 {
		t.Fatal("fixed lookup")
	}
	if Hsh_IntObj(manager, -1) != nil || Hsh_IntManEntryNum(manager) != 5 {
		t.Fatal("fixed object count")
	}
	if !slices.Equal(Hsh_IntData(manager, 1), []int32{13, 12}) {
		t.Fatal("record view")
	}
	Hsh_IntManStop(manager)
}

func TestHshFixedRecordsGrowth(t *testing.T) {
	data := Vec_IntAlloc(0)
	manager := Hsh_IntManStart(data, 3, 0)
	for i := 0; i < 1000; i++ {
		Vec_IntPush(data, int32(i))
		Vec_IntPush(data, -int32(i))
		Vec_IntPush(data, 17)
		if got := Hsh_IntManAdd(manager, i); got != int32(i) {
			t.Fatal(i, got)
		}
	}
	for i := 999; i >= 0; i-- {
		if got := Hsh_IntManAdd(manager, i); got != int32(i) {
			t.Fatal(i, got)
		}
	}
	if Hsh_IntManEntryNum(manager) != 1000 {
		t.Fatal("duplicates inserted")
	}
}

func TestHshWordsAndNativePacking(t *testing.T) {
	words := Vec_WrdAllocArrayCopy([]uint64{0xffffffffffffffff, 0x0123456789abcdef, 0, 0xffffffffffffffff, 0x0123456789abcdef, 0}, 6)
	result := Hsh_WrdManHashArray(words, 3)
	if !slices.Equal(Vec_IntArray(result), []int32{0, 0}) {
		t.Fatal(Vec_IntArray(result))
	}
	unique := Vec_WrdAllocArrayCopy([]uint64{0xffffffffffffffff, 0x0123456789abcdef}, 2)
	manager := Hsh_WrdManHashArrayStart(unique, 1)
	if Hsh_IntManEntryNum(manager) != 2 {
		t.Fatal("word start")
	}
	Vec_WrdWriteEntry(unique, 0, 0)
	// ArrayStart copies words into its owned int view.
	if Hsh_IntData(manager, 0)[0] != -1 || Hsh_IntData(manager, 0)[1] != -1 {
		t.Fatal("word source was borrowed")
	}
	var bytes [8]byte
	binary.NativeEndian.PutUint64(bytes[:], Hsh_IntWord(-3, -1))
	if int32(binary.NativeEndian.Uint32(bytes[:4])) != -3 || int32(binary.NativeEndian.Uint32(bytes[4:])) != -1 {
		t.Fatal("C union representation not preserved")
	}
}

func TestHshVariableVectors(t *testing.T) {
	manager := Hsh_VecManStart(0)
	defer Hsh_VecManStop(manager)
	original := Vec_IntAllocArrayCopy([]int32{2, 3, 5}, 3)
	id := Hsh_VecManAdd(manager, original)
	Vec_IntWriteEntry(original, 0, 99)
	if id != 0 || !slices.Equal(Hsh_VecReadArray(manager, id), []int32{2, 3, 5}) {
		t.Fatal("did not copy vector")
	}
	if empty := Hsh_VecManAdd(manager, Vec_IntAlloc(0)); empty != 1 {
		t.Fatal(empty)
	}
	for i := 0; i < 1000; i++ {
		v := Vec_IntStartNatural(i)
		if got := Hsh_VecManAdd(manager, v); i == 0 && got != 1 {
			t.Fatal(got)
		}
	}
	for i := 999; i >= 0; i-- {
		v := Vec_IntStartNatural(i)
		want := int32(i + 1)
		if i == 0 {
			want = 1
		}
		if got := Hsh_VecManAdd(manager, v); got != want {
			t.Fatal(i, got, want)
		}
	}
	if Hsh_VecSize(manager) != 1001 {
		t.Fatal(Hsh_VecSize(manager))
	}
	first := Hsh_VecReadEntry(manager, 0)
	second := Hsh_VecReadEntry1(manager, 1)
	third := Hsh_VecReadEntry2(manager, 2)
	if Vec_IntSize(first) != 3 || Vec_IntSize(second) != 0 || Vec_IntSize(third) != 1 {
		t.Fatal("independent temporary views")
	}
	if Hsh_VecObj(manager, -1) != nil {
		t.Fatal("sentinel object")
	}
}

func TestHshInvalidRecordSizes(t *testing.T) {
	for _, run := range []func(){
		func() { Hsh_IntManHashArray(Vec_IntStart(3), 2) },
		func() { Hsh_IntManHashArray(Vec_IntStart(0), 0) },
		func() { Hsh_WrdManHashArray(Vec_WrdStart(3), 2) },
		func() { Hsh_WrdManHashArrayStart(Vec_WrdStart(2), 1) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("invalid dimensions accepted")
				}
			}()
			run()
		}()
	}
}

func TestHshLargeGoTableSize(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("large Go table size requires 64-bit int")
	}
	largeSize := uint64(1) << 32
	tableSize := int(largeSize) + 1
	data := []int32{-1, 17}
	expected := int(hshJenkinsBytes(data))
	if got := Hsh_IntManHash(data, len(data), tableSize); got != expected {
		t.Fatal(got, expected)
	}
	if got := Hsh_Int1ManHash(-1, tableSize); got != int(hshJenkinsBytes([]int32{-1})) {
		t.Fatal(got)
	}
	v := Vec_IntAllocArrayCopy(data, len(data))
	if got := Hsh_VecManHash(v, tableSize); got != int(uint32(data[0])*hshPrimes[0]+uint32(data[1])*hshPrimes[1]) {
		t.Fatal(got)
	}
	// The byte-weighted alternative also uses the full Go integer table size.
	var bytes [4]byte
	var key uint32
	for i, value := range data {
		binary.NativeEndian.PutUint32(bytes[:], uint32(value))
		for j, b := range bytes {
			key += uint32(b) * hshPrimes[(4*i+j)%len(hshPrimes)]
		}
	}
	if got := Hsh_IntManHash2(data, len(data), tableSize); got != int(key) {
		t.Fatal(got, key)
	}
}
