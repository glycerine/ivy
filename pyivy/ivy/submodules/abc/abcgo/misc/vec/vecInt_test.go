package vec

import (
	"reflect"
	"testing"
)

func TestIntegerVectorsAndRecords(t *testing.T) {
	p := Vec_IntAlloc(0)
	for i := int32(0); i < 100; i++ {
		Vec_IntPush(p, i)
	}
	if Vec_IntSum(p) != 4950 || Vec_IntCountPositive(p) != 99 {
		t.Fatal("aggregate")
	}
	q := Vec_IntStartFull(3)
	if Vec_IntEntry(q, 2) != -1 {
		t.Fatal("full")
	}
	Vec_IntSetEntry(q, 7, 8)
	if q.nSize != 8 || Vec_IntEntry(q, 5) != 0 {
		t.Fatal("sparse set")
	}
	a := Vec_IntAllocArrayCopy([]int32{1, 2, 2, 3}, 4)
	if Vec_IntUniqify(a) != 1 || !reflect.DeepEqual(Vec_IntArray(a), []int32{1, 2, 3}) {
		t.Fatal("unique")
	}
	b := Vec_IntAllocArrayCopy([]int32{2, 4}, 2)
	merged := Vec_IntTwoMerge(a, b)
	if !reflect.DeepEqual(Vec_IntArray(merged), []int32{1, 2, 3, 4}) {
		t.Fatal("merge")
	}
	common := Vec_IntAlloc(0)
	Vec_IntTwoRemoveCommon(a, b, common)
	if !reflect.DeepEqual(Vec_IntArray(common), []int32{2}) || !reflect.DeepEqual(Vec_IntArray(a), []int32{1, 3}) || !reflect.DeepEqual(Vec_IntArray(b), []int32{4}) {
		t.Fatal("split")
	}
	records := Vec_IntAllocArrayCopy([]int32{3, 4, 1, 2, 3, 4}, 6)
	var mapping *Vec_Int_t
	if Vec_IntUniqueCount(records, 2, &mapping) != 2 || !reflect.DeepEqual(Vec_IntArray(mapping), []int32{0, 1, 0}) {
		t.Fatal("record classes")
	}
	unique := Vec_IntUniqifyHash(records, 2)
	if !reflect.DeepEqual(Vec_IntArray(unique), []int32{3, 4, 1, 2}) {
		t.Fatal("record order")
	}
	Vec_IntSortMulti(records, 2, 0)
	if records.pArray[0] != 1 || records.pArray[1] != 2 {
		t.Fatal("record sorting")
	}
}
func TestIntegerFetchAndOverflow(t *testing.T) {
	p := Vec_IntAllocExact(2)
	words := Vec_IntFetch(p, 2)
	words[0] = 0xffffffff
	if Vec_IntEntry(p, 0) != -1 {
		t.Fatal("unsigned alias")
	}
	if Vec_IntFetch(p, 1) != nil || p.nSize != 3 {
		t.Fatal("failed reservation semantics")
	}
	q := Vec_IntStart(1)
	Vec_IntWriteEntry(q, 0, 2147483647)
	if Vec_IntAddToEntry(q, 0, 1) != -2147483648 {
		t.Fatal("32-bit overflow")
	}
}
