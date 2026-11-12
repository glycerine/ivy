package gia

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// Gia_SortGetTest returns a deterministic integer input for sort benchmarks.
func Gia_SortGetTest(size int32) []int32 {
	if size < 0 {
		panic("gia: negative sort size")
	}
	random := rand.New(rand.NewSource(0))
	result := make([]int32, size)
	for i := range result {
		result[i] = random.Int31()
	}
	return result
}
func Gia_SortVerifySorted(values []int32, size int32) {
	if size < 0 || int(size) > len(values) {
		panic("gia: invalid sort size")
	}
	for i := 1; i < int(size); i++ {
		if values[i-1] > values[i] {
			panic("gia: integers are not sorted")
		}
	}
}
func minisat_sort(values []int32, size int32, less func(int32, int32) bool) {
	if size < 0 || int(size) > len(values) {
		panic("gia: invalid sort size")
	}
	sort.Slice(values[:size], func(i, j int) bool { return less(values[i], values[j]) })
}
func minisat_sort2(values []int32, size int32) {
	minisat_sort(values, size, func(a, b int32) bool { return a < b })
}
func Gia_SortTest() {
	const size = 100000000
	values := Gia_SortGetTest(size)
	started := time.Now()
	minisat_sort2(values, size)
	fmt.Printf("Go sort: %s\n", time.Since(started))
	Gia_SortVerifySorted(values, size)
}

// Gia_SortFloats sorts values in place and applies the same permutation to perm.
// If perm is nil, it starts as the identity permutation.
func Gia_SortFloats(values []float32, perm []int32, size int32) []int32 {
	if size < 0 || int(size) > len(values) {
		panic("gia: invalid sort size")
	}
	if perm == nil {
		perm = make([]int32, size)
		for i := range perm {
			perm[i] = int32(i)
		}
	} else if int(size) > len(perm) {
		panic("gia: short sort permutation")
	}
	sort.Sort(giaFloatPermutation{values[:size], perm[:size]})
	return perm
}

type giaFloatPermutation struct {
	values []float32
	perm   []int32
}

func (p giaFloatPermutation) Len() int           { return len(p.values) }
func (p giaFloatPermutation) Less(i, j int) bool { return p.values[i] < p.values[j] }
func (p giaFloatPermutation) Swap(i, j int) {
	p.values[i], p.values[j] = p.values[j], p.values[i]
	p.perm[i], p.perm[j] = p.perm[j], p.perm[i]
}
func minisat_sort3(values []float32, perm []int32, size int32) { Gia_SortFloats(values, perm, size) }
