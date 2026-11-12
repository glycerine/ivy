package vec

import (
	"math/rand"
	"testing"
)

func TestHashIntPairsPayloadAndCollision(t *testing.T) {
	p := Hash_IntManStart(0)
	defer Hash_IntManStop(p)
	if Hash_IntObj(p, 0) != nil || Hash_IntManEntryNum(p) != 0 {
		t.Fatal("sentinel")
	}
	for i := int32(0); i < 20; i++ {
		a := i * int32(Vec_IntSize(p.vTable))
		if id := Hash_Int2ManInsert(p, a, 0, i+10); id != i+1 {
			t.Fatal(id, i)
		}
	}
	if Hash_IntManBucketLengths(p)[0] != 20 {
		t.Fatal(Hash_IntManBucketLengths(p))
	}
	id := Hash_Int2ManInsert(p, 0, 0, 999)
	if id != 1 || Hash_IntObjData2(p, id) != 10 {
		t.Fatal("duplicate changed payload")
	}
	if Hash_Int2ObjInc(p, id) != 10 || Hash_IntObjData2(p, id) != 11 || Hash_Int2ObjDec(p, id) != 10 {
		t.Fatal("counter semantics")
	}
	Hash_Int2ObjSetData2(p, id, -19)
	if Hash_IntObjData2(p, id) != -19 {
		t.Fatal("payload write")
	}
	if *Hash_Int2ManLookup(p, 42, 43) != 0 {
		t.Fatal("missing key found")
	}
	if Hash_IntManEntryNum(p) != 20 {
		t.Fatal("duplicate created object")
	}
}

func TestHashIntPairResizeStableObjects(t *testing.T) {
	p := Hash_IntManStart(0)
	defer Hash_IntManStop(p)
	id := Hash_Int2ManInsert(p, -1, -2, -3)
	saved := Hash_IntObj(p, id)
	const count = 20000
	for i := int32(0); i < count; i++ {
		got := Hash_Int2ManInsert(p, i*12345, -i, i)
		if got != i+2 {
			t.Fatal(got, i)
		}
	}
	if Hash_IntObj(p, id) != saved {
		t.Fatal("object pointer changed on resize")
	}
	if Vec_IntSize(p.vTable) < count {
		t.Fatal("table did not resize")
	}
	for i := int32(0); i < count; i++ {
		got := *Hash_Int2ManLookup(p, i*12345, -i)
		if got != i+2 || Hash_IntObjData2(p, got) != i {
			t.Fatal(got, i)
		}
	}
	if *Hash_Int2ManLookup(p, -1, -2) != id {
		t.Fatal("initial key lost")
	}
	total := 0
	for _, n := range Hash_IntManBucketLengths(p) {
		total += n
	}
	if total != count+1 {
		t.Fatal(total)
	}
}

func TestHashIntTriples(t *testing.T) {
	p := Hash_IntManStart(1)
	defer Hash_IntManStop(p)
	for i := int32(0); i < 1000; i++ {
		id := Hsh_Int3ManInsert(p, -i, 7, i%13)
		if id != i+1 {
			t.Fatal(id)
		}
	}
	a := Hsh_Int3ManInsert(p, 0, 7, 14)
	if a != 1001 {
		t.Fatal(a)
	}
	if Hsh_Int3ManInsert(p, 0, 7, 14) != a || *Hsh_Int3ManLookup(p, 0, 7, 15) != 0 {
		t.Fatal("third key ignored")
	}
	for i := int32(0); i < 1000; i++ {
		id := *Hsh_Int3ManLookup(p, -i, 7, i%13)
		if id != i+1 || Hash_IntObjData0(p, id) != -i || Hash_IntObjData1(p, id) != 7 || Hash_IntObjData2(p, id) != i%13 {
			t.Fatal(i, id)
		}
	}
}

func TestHashIntUnsignedArithmetic(t *testing.T) {
	random := rand.New(rand.NewSource(1))
	for i := 0; i < 10000; i++ {
		a, b, c := int32(random.Uint32()), int32(random.Uint32()), int32(random.Uint32())
		size := 101 + i%100
		pair := uint32(uint64(uint32(a))*4177 + uint64(uint32(b))*7873)
		triple := uint32(uint64(uint32(a))*4177 + uint64(uint32(b))*7873 + uint64(uint32(c))*1699)
		if Hash_Int2ManHash(a, b, size) != int(pair%uint32(size)) || Hsh_Int3ManHash(a, b, c, size) != int(triple%uint32(size)) {
			t.Fatal(a, b, c)
		}
	}
}

func TestHashIntReferenceLifecycle(t *testing.T) {
	p := Hash_IntManStart(5)
	q := Hash_IntManRef(p)
	if p != q || p.nRefs != 2 {
		t.Fatal("reference")
	}
	Hash_IntManDeref(p)
	if p.nRefs != 1 {
		t.Fatal("premature release")
	}
	Hash_Int2ManInsert(q, 1, 2, 3)
	Hash_IntManDeref(q)
	if q.nRefs != 0 || q.vTable != nil || q.objects != nil {
		t.Fatal("not released")
	}
	Hash_IntManDeref(nil)
	Hash_IntManStop(nil)
	Hash_IntManHashArrayTest()
}
