package util

// Optional allocation-handle recording from utilMem.c. Go owns object storage;
// this recorder tracks numeric allocation IDs for diagnostics and drops records
// when recycled. It does not manufacture addresses or explicitly free Go data.
import (
	"sort"
	"sync"
)

type Vec_Mem_t struct{ pArray []AbcPointerValue }

func Vec_MemAlloc(capacity int32) *Vec_Mem_t {
	if capacity > 0 && capacity < 8 {
		capacity = 8
	}
	return &Vec_Mem_t{pArray: make([]AbcPointerValue, 0, capacity)}
}
func Vec_MemFree(p *Vec_Mem_t) {
	if p != nil {
		p.pArray = nil
	}
}
func Vec_MemGrow(p *Vec_Mem_t, capacity int32) {
	if cap(p.pArray) >= int(capacity) {
		return
	}
	a := make([]AbcPointerValue, len(p.pArray), capacity)
	copy(a, p.pArray)
	p.pArray = a
}
func Vec_MemPush(p *Vec_Mem_t, value AbcPointerValue) {
	if len(p.pArray) == cap(p.pArray) {
		capacity := 2 * cap(p.pArray)
		if capacity < 16 {
			capacity = 16
		}
		Vec_MemGrow(p, int32(capacity))
	}
	p.pArray = append(p.pArray, value)
}
func Vec_MemSort(p *Vec_Mem_t, compare func(AbcPointerValue, AbcPointerValue) int32) {
	sort.Slice(p.pArray, func(i, j int) bool { return compare(p.pArray[i], p.pArray[j]) < 0 })
}
func Util_ComparePointers(a, b AbcPointerValue) int32 {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
func Vec_MemTwoMerge(allocated, freed *Vec_Mem_t) *Vec_Mem_t {
	result := Vec_MemAlloc(int32(len(allocated.pArray)))
	i, j := 0, 0
	for i < len(allocated.pArray) && j < len(freed.pArray) {
		a, b := allocated.pArray[i], freed.pArray[j]
		if a == b {
			i++
			j++
		} else if a < b {
			Vec_MemPush(result, a)
			i++
		} else {
			panic("memory recorder: free without allocation")
		}
	}
	if j != len(freed.pArray) {
		panic("memory recorder: free without allocation")
	}
	for ; i < len(allocated.pArray); i++ {
		Vec_MemPush(result, allocated.pArray[i])
	}
	return result
}

var memoryRecordMu sync.Mutex

func Util_MemRecAlloc(value AbcPointerValue) AbcPointerValue {
	memoryRecordMu.Lock()
	defer memoryRecordMu.Unlock()
	if s_vAllocs != nil {
		Vec_MemPush(s_vAllocs, value)
	}
	return value
}
func Util_MemRecFree(value AbcPointerValue) AbcPointerValue {
	memoryRecordMu.Lock()
	defer memoryRecordMu.Unlock()
	if s_vFrees != nil {
		Vec_MemPush(s_vFrees, value)
	}
	return value
}
func Util_MemRecStart() {
	memoryRecordMu.Lock()
	defer memoryRecordMu.Unlock()
	if s_vAllocs != nil || s_vFrees != nil {
		panic("memory recorder already started")
	}
	s_vAllocs = Vec_MemAlloc(1000)
	s_vFrees = Vec_MemAlloc(1000)
}
func Util_MemRecRecycle() {
	memoryRecordMu.Lock()
	defer memoryRecordMu.Unlock()
	if s_vAllocs == nil || s_vFrees == nil {
		panic("memory recorder not started")
	}
	Vec_MemSort(s_vAllocs, Util_ComparePointers)
	Vec_MemSort(s_vFrees, Util_ComparePointers)
	_ = Vec_MemTwoMerge(s_vAllocs, s_vFrees)
	s_vAllocs.pArray = s_vAllocs.pArray[:0]
	s_vFrees.pArray = s_vFrees.pArray[:0]
}
func Util_MemRecQuit() {
	memoryRecordMu.Lock()
	defer memoryRecordMu.Unlock()
	if s_vAllocs == nil || s_vFrees == nil {
		panic("memory recorder not started")
	}
	Vec_MemFree(s_vAllocs)
	Vec_MemFree(s_vFrees)
	s_vAllocs = nil
	s_vFrees = nil
}
func Util_MemRecIsSet() bool {
	memoryRecordMu.Lock()
	defer memoryRecordMu.Unlock()
	return s_vAllocs != nil && s_vFrees != nil
}
