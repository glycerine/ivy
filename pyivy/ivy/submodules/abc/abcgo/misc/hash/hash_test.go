package hash

import (
	"math/rand"
	"testing"
)

func TestIntegerTableCollisionsAndStablePointers(t *testing.T) {
	p := Hash_IntAlloc(1)
	reference := map[int32]int32{}
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 2000; i++ {
		key := int32(rng.Intn(100) - 50)
		value := rng.Int31()
		Hash_IntWriteEntry(p, key, value)
		reference[key] = value
	}
	for key, value := range reference {
		if Hash_IntExists(p, key) != 1 || Hash_IntEntry(p, key, 0) != value {
			t.Fatal(key, value)
		}
	}
	ptr := Hash_IntEntryPtr(p, 100)
	*ptr = 42
	for i := int32(101); i < 200; i++ {
		Hash_IntWriteEntry(p, i, i)
	}
	if *ptr != 42 || Hash_IntEntry(p, 100, 0) != 42 {
		t.Fatal("entry address changed")
	}
	count := 0
	p.Range(func(key int32, value *int32) bool { count++; return true })
	if count != p.Len() {
		t.Fatal(count, p.Len())
	}
	old := p.Len()
	if Hash_IntEntry(p, 999, 0) != 0 || p.Len() != old {
		t.Fatal("lookup created entry")
	}
	Hash_IntEntry(p, 999, 1)
	if p.Len() != old+1 {
		t.Fatal("create failed")
	}
	Hash_IntFree(p)
	if p.Len() != 0 {
		t.Fatal("free failed")
	}
}
func TestFloatAndPointerRemoval(t *testing.T) {
	f := Hash_FltAlloc(1)
	for i := int32(0); i < 3; i++ {
		Hash_FltWriteEntry(f, i, float32(i)+0.5)
	}
	*Hash_FltEntryPtr(f, 1) = 7.25
	Hash_FltRemove(f, 0)
	Hash_FltRemove(f, 2)
	if f.Len() != 1 || Hash_FltEntry(f, 1, 0) != 7.25 {
		t.Fatal(f)
	}
	Hash_FltRemove(f, 1)
	Hash_FltRemove(f, 1)
	if f.Len() != 0 {
		t.Fatal(f)
	}
	p := Hash_PtrAlloc(1)
	a := new(int)
	*a = 9
	Hash_PtrWriteEntry(p, 1, a)
	Hash_PtrWriteEntry(p, 2, "second")
	if Hash_PtrRemove(p, 1) != a || Hash_PtrExists(p, 1) != 0 {
		t.Fatal(p)
	}
	*Hash_PtrEntryPtr(p, 3) = "third"
	if Hash_PtrEntry(p, 3, 0) != "third" || Hash_PtrRemove(p, 88) != nil {
		t.Fatal(p)
	}
	Hash_PtrFree(p)
	Hash_FltFree(f)
}
func TestGenericTablesAndCustomComparison(t *testing.T) {
	p := Hash_GenAlloc(1, nil, nil, 1)
	Hash_GenWriteEntry(p, "first", 12)
	Hash_GenWriteEntry(p, "second", 34)
	entry := Hash_GenEntry(p, "first", 0)
	entry.SetValue(56)
	if entry.Key() != "first" || entry.Value() != 56 || p.Len() != 2 {
		t.Fatal(entry)
	}
	if Hash_GenRemove(p, "first") != 56 || Hash_GenExists(p, "first") != 0 {
		t.Fatal(p)
	}
	seen := 0
	p.Range(func(e *Hash_Gen_Entry_t) bool { seen++; return true })
	if seen != 1 {
		t.Fatal(seen)
	}
	q := Hash_GenAlloc(3, Hash_DefaultHashFuncInt, Hash_DefaultCmpFuncInt, 0)
	Hash_GenWriteEntry(q, int64(-4), "negative")
	if Hash_GenEntry(q, int32(-4), 0).Value() != "negative" {
		t.Fatal(q)
	}
	Hash_GenFree(p)
	Hash_GenFree(q)
}
func TestHashRange(t *testing.T) {
	for bins := int32(1); bins < 15; bins++ {
		for key := int32(-10000); key < 10000; key++ {
			h := Hash_DefaultHashFunc(key, bins)
			if h < 0 || h >= bins {
				t.Fatal(key, bins, h)
			}
		}
	}
}

func TestCStringTerminationAndEarlyIterationStop(t *testing.T) {
	p := Hash_GenAlloc(7, nil, nil, 0)
	Hash_GenWriteEntry(p, "same\x00ignored", 4)
	if Hash_GenEntry(p, "same", 0).Value() != 4 {
		t.Fatal("NUL string equality differs from C")
	}
	q := Hash_IntAlloc(1)
	Hash_IntWriteEntry(q, 1, 1)
	Hash_IntWriteEntry(q, 2, 2)
	visited := 0
	q.Range(func(_ int32, _ *int32) bool { visited++; return false })
	if visited != 1 {
		t.Fatal(visited)
	}
}
